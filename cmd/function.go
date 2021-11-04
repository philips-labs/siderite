//go:build !windows
// +build !windows

package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/iron-io/iron_go3/worker"
	chclient "github.com/jpillora/chisel/client"
	"github.com/jpillora/chisel/share/cos"
	"github.com/philips-labs/siderite/logger"
	"github.com/philips-labs/siderite/models"
	"github.com/spf13/cobra"
)

type request struct {
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
	Callback string            `json:"callback"`
	Path     string            `json:"path"`
}

func NewFunctionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "function",
		Short: "Run in function mode",
		Long:  `Runs siderite in hsdp_function support mode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			worker.ParseFlags()
			var p models.Payload
			err := worker.PayloadFromJSON(&p)
			if err != nil {
				fmt.Printf("Failed to read payload from JSON: %v\n", err)
				return err
			}

			if len(p.Version) < 1 || p.Version != "1" {
				fmt.Printf("[siderite] unsupported or unknown payload version: %v\n", p.Version)
			}

			taskID := os.Getenv("TASK_ID") // Get our task ID
			if taskID == "" {
				taskID = "local"
			}

			_, deferFunc, err := logger.SetupHSDPLogging(p, taskID)
			if err == nil {
				defer deferFunc()
			}
			_, _ = fmt.Fprintf(os.Stdout, "[siderite] function version %s start\n", GitCommit)

			// Start
			c := make(chan int)
			go func() {
				_ = task(false, c)(cmd, args)
			}()
			// Wait for the function to become available
			pid := <-c
			_, _ = fmt.Fprintf(os.Stdout, "[siderite] waiting for application to become available on 127.0.0.1:8080\n")
			waitOk, err := waitForPort(30*time.Second, "127.0.0.1:8080")
			fmt.Printf("waitOk = %v, err = %v\n", waitOk, err)
			fmt.Printf("Mode = '%s'\n", p.Mode)
			fmt.Printf("PID = %d\n", pid)
			if p.Mode == "async" {
				// Retrieve the original request data
				client := resty.New()
				resp, err := client.R().
					SetHeader("Authorization", fmt.Sprintf("Token %s", p.Token)).
					Get(fmt.Sprintf("https://%s/payload/%s", p.Upstream, taskID))
				if err != nil {
					fmt.Printf("Error retrieving payload. Need to recover somehow...\n")
					_ = kill(pid, syscall.SIGTERM)
					return err
				}
				var originalRequest request
				err = json.Unmarshal(resp.Body(), &originalRequest)
				if err != nil {
					fmt.Printf("Error decoding. Need to recover somehow...\n")
					_ = kill(pid, syscall.SIGTERM)
					return err
				}
				// Replay the request to the function
				resp, err = client.R().SetHeaders(originalRequest.Headers).
					SetBody(originalRequest.Body).
					Post("http://127.0.0.1:8080" + originalRequest.Path)
				if err != nil {
					fmt.Printf("Error performing request. Need to recover somehow...\n")
					_ = kill(pid, syscall.SIGTERM)
					return err
				}
				// Callback with results
				fmt.Printf("[siderite] posting result to callback URL: %s\n", originalRequest.Callback)
				resp, err = client.R().SetBody(resp.Body()).Post(originalRequest.Callback)
				if err != nil {
					fmt.Printf("[siderite] callback error: %v\n", err)
				}
				if resp != nil {
					fmt.Printf("[siderite] callback statusCode: %d\n", resp.StatusCode())
				}
				fmt.Printf("[siderite] sending SIGTERM to myself\n")
				_ = kill(pid, syscall.SIGTERM)
				return err
			}

			// Build chisel connect args
			server := fmt.Sprintf("https://%s:4443", p.Upstream)
			remote := "R:8081:127.0.0.1:8080"
			fmt.Printf("[siderite] setting up reverse tunnel: %s %s\n", server, remote)
			chiselArgs := []string{
				server,
				remote,
			}
			auth := fmt.Sprintf("chisel:%s", p.Token)
			client, err := chiselClient(chiselArgs, auth)
			if err != nil {
				fmt.Printf("[siderite] error creating chisel client: %v\n", err)
				return err
			}
			go cos.GoStats()
			ctx := cos.InterruptContext()
			if err := client.Start(ctx); err != nil {
				fmt.Printf("[siderite] error starting chisel client: %v\n", err)
				return err
			}
			fmt.Printf("[siderite] chisel client running. waiting...\n")
			if err := client.Wait(); err != nil {
				log.Fatal(err)
			}
			return nil
		},
	}
}

func init() {
	rootCmd.AddCommand(NewFunctionCmd())
}

func waitForPort(timeout time.Duration, host string) (bool, error) {
	if timeout == 0 {
		timeout = time.Duration(1) * time.Minute
	}
	until := time.Now().Add(timeout)
	for {
		var conn net.Conn
		conn, _ = net.DialTimeout("tcp", host, timeout)
		if conn != nil {
			err := conn.Close()
			return true, err
		}
		time.Sleep(100 * time.Millisecond)
		if time.Now().After(until) {
			return false, fmt.Errorf("timed out waiting for %s", host)
		}
	}
}
func chiselClient(args []string, auth string) (*chclient.Client, error) {
	config := chclient.Config{
		Headers: http.Header{},
		Auth:    auth,
	}
	if len(args) < 2 {
		return nil, fmt.Errorf("a server and least one remote is required")
	}
	config.Server = args[0]
	config.Remotes = args[1:]
	//default auth
	if config.Auth == "" {
		config.Auth = os.Getenv("AUTH")
	}
	//ready
	c, err := chclient.NewClient(&config)
	if err != nil {
		return nil, err
	}
	c.Debug = true
	return c, nil
}

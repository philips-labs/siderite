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
	"github.com/philips-labs/siderite"
	"github.com/spf13/cobra"
)

type request struct {
	Headers  map[string]string `json:"headers"`
	Body     string            `json:"body"`
	Callback string            `json:"callback"`
	Path     string            `json:"path"`
}

// functionCmd represents the function command
var functionCmd = &cobra.Command{
	Use:   "function",
	Short: "Run in function mode",
	Long:  `Runs siderite in hsdp_function support mode`,
	Run: func(cmd *cobra.Command, args []string) {
		worker.ParseFlags()
		p := &siderite.Payload{}
		err := worker.PayloadFromJSON(p)
		if err != nil {
			fmt.Printf("Failed to read payload from JSON: %v", err)
			return
		}

		if len(p.Version) < 1 || p.Version != "1" {
			fmt.Println("[siderite] unsupported or unknown payload version", p.Version)
		}
		// Start
		c := make(chan int)
		go runner(false, c)(cmd, args)
		// Wait for the function to become available
		pid := <-c
		_, _ = waitForPort(30*time.Second, "127.0.0.1:8080")

		fmt.Printf("Mode = '%s'\n", p.Mode)
		fmt.Printf("PID = %d\n", pid)
		if p.Mode == "async" {
			// Retrieve the original request data
			taskID := os.Getenv("TASK_ID") // Get our task ID
			client := resty.New()
			resp, err := client.R().
				SetHeader("Authorization", fmt.Sprintf("Token %s", p.Token)).
				Get(fmt.Sprintf("https://%s/payload/%s", p.Upstream, taskID))
			if err != nil {
				fmt.Printf("Error retrieving payload. Need to recover somehow...\n")
				_ = syscall.Kill(pid, syscall.SIGTERM)
				return
			}
			var originalRequest request
			err = json.Unmarshal(resp.Body(), &originalRequest)
			if err != nil {
				fmt.Printf("Error decoding. Need to recover somehow...\n")
				_ = syscall.Kill(pid, syscall.SIGTERM)
				return
			}
			// Replay the request to the function
			resp, err = client.R().SetHeaders(originalRequest.Headers).
				SetBody(originalRequest.Body).
				Post("http://127.0.0.1:8080" + originalRequest.Path)
			if err != nil {
				fmt.Printf("Error performing request. Need to recover somehow...\n")
				_ = syscall.Kill(pid, syscall.SIGTERM)
				return
			}
			// Callback with results
			_, _ = client.R().SetBody(resp.Body()).Post(originalRequest.Callback)
			_ = syscall.Kill(pid, syscall.SIGTERM)
			return
		}

		// Build chisel connect args
		chiselArgs := []string{
			fmt.Sprintf("https://%s:4443", p.Upstream),
			fmt.Sprintf("R:8081:127.0.0.1:8080"),
		}
		auth := fmt.Sprintf("chisel:%s", p.Token)
		chiselClient(chiselArgs, auth)
	},
}

func init() {
	rootCmd.AddCommand(functionCmd)
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
func chiselClient(args []string, auth string) {
	config := chclient.Config{
		Headers: http.Header{},
		Auth:    auth,
	}
	if len(args) < 2 {
		log.Fatalf("A server and least one remote is required")
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
		log.Fatal(err)
	}
	/*
		c.Debug = *verbose
		if *pid {
			//generatePidFile()
		}
	*/
	go cos.GoStats()
	ctx := cos.InterruptContext()
	if err := c.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := c.Wait(); err != nil {
		log.Fatal(err)
	}
}

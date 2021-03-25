package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/iron-io/iron_go3/worker"
	chclient "github.com/jpillora/chisel/client"
	"github.com/jpillora/chisel/share/cos"
	"github.com/spf13/cobra"
)

// functionCmd represents the function command
var functionCmd = &cobra.Command{
	Use:   "function",
	Short: "Run in function mode",
	Long:  `Runs siderite in hsdp_function support mode`,
	Run: func(cmd *cobra.Command, args []string) {
		worker.ParseFlags()
		p := &Payload{}
		err := worker.PayloadFromJSON(p)
		if err != nil {
			fmt.Printf("Failed to read payload from JSON: %v", err)
			return
		}

		if len(p.Version) < 1 || p.Version != "1" {
			fmt.Println("[siderite] unsupported or unknown payload version", p.Version)
		}
		// Start
		go runner(false)(cmd, args)
		// Wait for the function to become available
		_, _ = waitForPort(30*time.Second, "127.0.0.1:8080")

		if p.Mode == "async" {
			// TODO: call the gateway and pick up the payload and reploy it here
			http.Post("http://127.0.0.1:8080/", "application/json", ioutil.NopCloser(strings.NewReader("")))
			// Exit job
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
	//flags := flag.NewFlagSet("client", flag.ContinueOnError)
	config := chclient.Config{
		Headers: http.Header{},
		Auth:    auth,
	}
	/*
		flags.StringVar(&config.Fingerprint, "fingerprint", "", "")
		flags.StringVar(&config.Auth, "auth", "", "")
		flags.DurationVar(&config.KeepAlive, "keepalive", 25*time.Second, "")
		flags.IntVar(&config.MaxRetryCount, "max-retry-count", -1, "")
		flags.DurationVar(&config.MaxRetryInterval, "max-retry-interval", 0, "")
		flags.StringVar(&config.Proxy, "proxy", "", "")
		flags.StringVar(&config.TLS.CA, "tls-ca", "", "")
		flags.BoolVar(&config.TLS.SkipVerify, "tls-skip-verify", false, "")
		flags.StringVar(&config.TLS.Cert, "tls-cert", "", "")
		flags.StringVar(&config.TLS.Key, "tls-key", "", "")
		hostname := flags.String("hostname", "", "")
		pid := flags.Bool("pid", false, "")
		verbose := flags.Bool("v", false, "")
		flags.Parse(args)
		//pull out options, put back remaining args
		args = flags.Args()
	*/
	if len(args) < 2 {
		log.Fatalf("A server and least one remote is required")
	}
	config.Server = args[0]
	config.Remotes = args[1:]
	//default auth
	if config.Auth == "" {
		config.Auth = os.Getenv("AUTH")
	}
	//move hostname onto headers
	/*
		if *hostname != "" {
			config.Headers.Set("Host", *hostname)
		}
	*/
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

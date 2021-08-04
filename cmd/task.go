package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/iron-io/iron_go3/worker"
	"github.com/philips-labs/siderite/logger"
	"github.com/philips-labs/siderite/models"
	"github.com/philips-software/go-hsdp-api/logging"
	"github.com/spf13/cobra"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:     "task",
	Aliases: []string{"runner"},
	Short:   "runs command described in payload",
	Long: `runs the command provided in the payload file.

this mode should be used inside an IronIO docker task. siderite
will block until the command exits.`,
	Run: task(true, nil),
}

func init() {
	rootCmd.AddCommand(taskCmd)
}

func task(parseFlags bool, c chan int) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		taskID := os.Getenv("TASK_ID") // Get our task ID
		if taskID == "" {
			taskID = "local"
		}
		done := make(chan bool)
		old := os.Stdout // keep backup of the real stdout
		r, w, err := os.Pipe()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "Error setting up pipe: %v\n", err)
			return
		}
		os.Stdout = w

		err = logger.ToHSDP(r, logging.Resource{
			ApplicationInstance: uuid.New().String(),
			EventID:             "1",
			ApplicationName:     "hsdp_function",
			ApplicationVersion:  "1.0.0",
			Component:           "siderite",
			Category:            "TaskLog",
			Severity:            "info",
			OriginatingUser:     "siderite",
			ServerName:          "iron.io",
			ServiceName:         taskID,
		}, done)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[siderite] not logging to HSDP: %v\n", err)
			os.Stdout = old
		} else {
			_, _ = fmt.Fprintf(os.Stdout, "[siderite] logging stdout to HSDP logging\n")
			defer func() {
				os.Stdout = old
				fmt.Printf("flushing logs\n")
				time.Sleep(3 * time.Second)
				done <- true
			}()
		}

		fmt.Fprintf(os.Stdout, "[siderite] task version %s start\n", GitCommit)

		if parseFlags {
			worker.ParseFlags()
		}
		p := &models.Payload{}
		err = worker.PayloadFromJSON(p)
		if err != nil {
			fmt.Printf("Failed to read payload from JSON: %v\n", err)
			return
		}

		if len(p.Version) < 1 || p.Version != "1" {
			fmt.Printf("[siderite] unsupported or unknown payload version: %s\n", p.Version)
		}
		if len(p.Cmd) < 1 {
			fmt.Printf("[siderite] missing command\n")
			return
		}

		fmt.Printf("executing: %s %v\n", p.Cmd[0], p.Cmd[1:])
		command := exec.Command(p.Cmd[0], p.Cmd[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Env = os.Environ()

		for k, v := range p.Env {
			command.Env = append(command.Env, k+"="+v)
		}
		_ = command.Start()
		if c != nil {
			c <- command.Process.Pid // Send to parent
		}
		err = command.Wait()
		fmt.Printf("result: %v\n", err)
		fmt.Printf("[siderite] version %s exit\n", GitCommit)
	}
}

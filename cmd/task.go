package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/iron-io/iron_go3/worker"
	"github.com/philips-labs/siderite/logger"
	"github.com/philips-labs/siderite/models"
	"github.com/spf13/cobra"
)

func NewTaskCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "task",
		Aliases: []string{"runner"},
		Short:   "runs command described in payload",
		Long: `runs the command provided in the payload file.

this mode should be used inside an IronIO docker task. siderite
will block until the command exits.`,
		RunE: task(true, nil),
	}
}

func kill(pid int, sig os.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Signal(sig)
}

func init() {
	rootCmd.AddCommand(NewTaskCmd())
}

func task(parseFlags bool, c chan int) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var p models.Payload
		if parseFlags {
			worker.ParseFlags()
		}
		err := worker.PayloadFromJSON(&p)
		if err != nil {
			fmt.Printf("[siderite] failed to read payload from JSON: %v\n", err)
			return err
		}

		taskID := os.Getenv("TASK_ID") // Get our task ID
		if taskID == "" {
			taskID = "local"
		}
		done, deferFunc, err := logger.Setup(p, taskID)
		if err == nil {
			defer deferFunc()
		}
		if err != nil {
			fmt.Printf("[siderite] logger disabled: %v\n", err)
		}

		_, _ = fmt.Fprintf(os.Stderr, "[siderite] task version %s start\n", GitCommit)

		if len(p.Version) < 1 || p.Version != "1" {
			fmt.Printf("[siderite] unsupported or unknown payload version: %s\n", p.Version)
		}
		if len(p.Cmd) < 1 {
			fmt.Printf("[siderite] missing command\n")
			return fmt.Errorf("missing command")
		}

		fmt.Printf("[siderite] executing: %s %v\n", p.Cmd[0], p.Cmd[1:])
		command := exec.Command(p.Cmd[0], p.Cmd[1:]...)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Env = os.Environ()

		for k, v := range p.Env {
			command.Env = append(command.Env, k+"="+v)
		}
		err = command.Start()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[siderite] error starting command: %v\n", err)
		}
		if c != nil {
			c <- command.Process.Pid // Send to parent
		}
		err = command.Wait()
		_, _ = fmt.Fprintf(os.Stdout, "[siderite] command result: %v\n", err)
		_, _ = fmt.Fprintf(os.Stderr, "[siderite] version %s exit\n", GitCommit)
		if done != nil {
			done <- true
		}
		return err
	}
}

package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/iron-io/iron_go3/worker"
	"github.com/spf13/cobra"
)

// runnerCmd represents the runner command
var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Runs command described in payload",
	Long: `Runs the command provided in the payload file.

This mode should be used inside an IronIO docker task. siderite
will block until the command exits.`,
	Run: run,
}

func init() {
	rootCmd.AddCommand(runnerCmd)
}

// Payload describes the JSON payload file
type Payload struct {
	Env map[string]string `json:"env,omitempty"`
	Cmd []string          `json:"cmd,omitempty"`
}

func run(cmd *cobra.Command, args []string) {
	worker.ParseFlags()
	p := &Payload{}
	worker.PayloadFromJSON(p)

	if len(p.Cmd) < 1 {
		fmt.Println("Missing command")
		os.Exit(1)
	}

	command := exec.Command(p.Cmd[0], p.Cmd[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Env = os.Environ()

	for k, v := range p.Env {
		command.Env = append(command.Env, k+"="+v)
	}
	command.Run()
	fmt.Printf("siderite %s exit\n", GitCommit)
}

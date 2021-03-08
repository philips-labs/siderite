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
	Short: "runs command described in payload",
	Long: `runs the command provided in the payload file.

this mode should be used inside an IronIO docker task. siderite
will block until the command exits.`,
	Run: run,
}

func init() {
	rootCmd.AddCommand(runnerCmd)
}

// Payload describes the JSON payload file
type Payload struct {
	Version string            `json:"version"`
	Env     map[string]string `json:"env,omitempty"`
	Cmd     []string          `json:"cmd,omitempty"`
}

func run(cmd *cobra.Command, args []string) {
	fmt.Printf("[siderite] version %s start\n", GitCommit)

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
	if len(p.Cmd) < 1 {
		fmt.Println("[siderite] missing command")
		os.Exit(1)
	}

	fmt.Printf("executing: %s %v\n", p.Cmd[0], p.Cmd[1:])
	command := exec.Command(p.Cmd[0], p.Cmd[1:]...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Env = os.Environ()

	for k, v := range p.Env {
		command.Env = append(command.Env, k+"="+v)
	}
	err = command.Run()
	fmt.Printf("result: %v\n", err)
	fmt.Printf("[siderite] version %s exit\n", GitCommit)
}

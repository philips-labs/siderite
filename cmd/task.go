package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/iron-io/iron_go3/worker"
	"github.com/philips-labs/siderite/models"
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
		fmt.Printf("[siderite] version %s start\n", GitCommit)

		if parseFlags {
			worker.ParseFlags()
		}
		p := &models.Payload{}
		err := worker.PayloadFromJSON(p)
		if err != nil {
			fmt.Printf("Failed to read payload from JSON: %v\n", err)
			fmt.Printf("Environment:\n%v\n", os.Environ())
			cmd := exec.Command("mount")
			var out bytes.Buffer
			cmd.Stdout = &out
			err = cmd.Run()
			if err != nil {
				fmt.Printf("[siderite] error running: %v\n", err)
			}
			fmt.Printf("Mount:\n%s\n", out.String())
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
		_ = command.Start()
		if c != nil {
			c <- command.Process.Pid // Send to parent
		}
		err = command.Wait()
		fmt.Printf("result: %v\n", err)
		fmt.Printf("[siderite] version %s exit\n", GitCommit)
	}
}

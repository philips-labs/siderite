/*
Copyright © 2020 Andy Lo-A-Foe <andy.lo-a-foe@philips.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runnerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runnerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Payload describes the JSON payload file
type Payload struct {
	Env map[string]string `json:"env"`
	Cmd []string          `json:"cmd"`
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
	fmt.Println("Siderite exit")
}

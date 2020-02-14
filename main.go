package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/iron-io/iron_go3/worker"
)

type File struct {
	Name string
	Body string
}

type Payload struct {
	Env   map[string]string `json:"env"`
	Cmd   []string          `json:"cmd"`
	Files []File            `json:"files"`
}

func main() {
	worker.ParseFlags()
	p := &Payload{}
	worker.PayloadFromJSON(p)

	if len(p.Cmd) < 1 {
		fmt.Println("Missing command")
		os.Exit(1)
	}

	cmd := exec.Command(p.Cmd[0], p.Cmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	for k, v := range p.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	cmd.Run()
}

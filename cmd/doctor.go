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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Checks system configuration",
	Long: `Check wether your system is configure so it can interact
	with the HSDP iron`,
	Run: doctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

// Config describes IronIO config
type Config struct {
	ClusterInfo []struct {
		ClusterID   string `json:"cluster_id"`
		ClusterName string `json:"cluster_name"`
		Pubkey      string `json:"pubkey"`
		UserID      string `json:"user_id"`
	} `json:"cluster_info"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Project   string `json:"project"`
	ProjectID string `json:"project_id"`
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
}

type proc func() error

var yellow = color.New(color.FgYellow).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

var pass = green("[✓]")
var warn = yellow("[!]")
var problem = red("[✗]")

func testIronCLI() error {
	path, err := exec.LookPath("iron")

	if err != nil {
		fmt.Println(problem, "iron CLI not found. Install it: https://github.com/iron-io/ironcli")
		return err
	}
	out, err := exec.Command(path, "-version").Output()
	if err != nil {
		fmt.Println(problem, "iron CLI failed to run:", err.Error())
		return err
	}
	version := strings.TrimSpace(string(out))
	if version != "0.1.6" {
		fmt.Printf("%s iron CLI version 0.1.6 not detected (version %s)", warn, version)
		return err
	}

	fmt.Printf("%s iron CLI installed (version %s)\n", pass, version)
	return err
}

func testCF() error {
	path, err := exec.LookPath("cf")

	if err != nil {
		fmt.Println(problem, "cf CLI not found. Install it: https://docs.cloudfoundry.org/cf-cli/install-go-cli.html")
		return err
	}
	out, err := exec.Command(path, "version").Output()
	if err != nil {
		fmt.Println(problem, "cf CLI failed to run:", err.Error())
		return err
	}
	version := strings.TrimSpace(string(out))
	fmt.Printf("%s cf CLI installed (%s)\n", pass, version)
	return nil
}

func testConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s user home directory issue: %s\n", problem, err.Error())
		return err
	}
	configFile := filepath.Join(home, ".iron.json")
	configJSON, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("%s config file issue: %s\n", problem, err.Error())
		return err
	}
	var jsonConfig Config
	err = json.Unmarshal(configJSON, &jsonConfig)
	if err != nil {
		fmt.Printf("%s error parsing config: %s\n", problem, err.Error())
		return err
	}
	fmt.Printf("%s iron configuration file (%s)\n", pass, configFile)
	return nil
}

func doctor(cmd *cobra.Command, args []string) {
	var errors bool

	e := []proc{
		testIronCLI,
		testConfig,
		testCF,
	}

	for _, p := range e {
		if p() != nil {
			errors = true
		}
	}
	if errors {
		fmt.Println("some errors were detected")
	}
}

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// env2jsonCmd represents the env2json command
var env2jsonCmd = &cobra.Command{
	Use:   "env2json",
	Short: "Converts env output to JSON payload",
	Long: `You can pipe the output of the env command to this command 
which will output a JSON structure with proper escaping`,
	Run: env2JSON,
}

func init() {
	rootCmd.AddCommand(env2jsonCmd)
}

var envParse = regexp.MustCompile(`^(.*?)=(.*)$`)

func env2JSON(cmd *cobra.Command, args []string) {
	var payload Payload

	envInput, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println(err)
		return
	}
	payload.Env = make(map[string]string)

	for _, line := range strings.Split(strings.TrimSuffix(string(envInput), "\n"), "\n") {
		parsed := envParse.FindStringSubmatch(line)
		if len(parsed) < 3 {
			fmt.Println("Skipping line")
		} else {
			payload.Env[parsed[1]] = parsed[2]
		}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(`{"env":[]}`)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	out.WriteTo(os.Stdout)
}

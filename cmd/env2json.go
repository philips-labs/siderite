package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
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
	env2jsonCmd.Flags().StringP("include", "i", "", "comma separated list of variables to include")
	env2jsonCmd.Flags().StringP("exclude", "x", "", "comma separated list of variables to exclude")
	env2jsonCmd.Flags().StringSliceP("env", "e", []string{}, "add environment variable")
	env2jsonCmd.Flags().StringSliceP("cmd", "c", []string{}, "command to include")
	env2jsonCmd.Flags().BoolP("nostdin", "n", false, "skip reading from stdin")
}

var envParse = regexp.MustCompile(`^(.*?)=(.*)$`)

func contains(s []string, searchterm string) bool {
	i := sort.SearchStrings(s, searchterm)
	return i < len(s) && s[i] == searchterm
}

func env2JSON(cmd *cobra.Command, args []string) {
	var payload Payload

	includeList, _ := cmd.Flags().GetString("include")
	excludeList, _ := cmd.Flags().GetString("exclude")
	include := strings.Split(includeList, ",")
	exclude := strings.Split(excludeList, ",")
	sort.Strings(include)
	sort.Strings(exclude)

	if len(include) > 0 && include[0] != "" && len(exclude) > 0 && exclude[0] != "" {
		fmt.Fprintf(os.Stderr, "can't use include and exclude simultaneously\n")
		return
	}
	envInput := []byte("")
	var err error

	nostdin, _ := cmd.Flags().GetBool("nostdin")

	if !nostdin {
		envInput, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	payload.Env = make(map[string]string)

	for _, line := range strings.Split(strings.TrimSuffix(string(envInput), "\n"), "\n") {
		parsed := envParse.FindStringSubmatch(line)
		if len(parsed) == 3 {
			key := parsed[1]
			value := parsed[2]
			if contains(exclude, key) {
				continue
			}
			if len(include) > 0 && include[0] != "" && !contains(include, key) {
				continue
			}
			payload.Env[key] = value
		}
	}
	// Extra environment
	extraVars, _ := cmd.Flags().GetStringSlice("env")
	for _, e := range extraVars {
		parsed := envParse.FindStringSubmatch(e)
		if len(parsed) == 3 {
			key := parsed[1]
			value := parsed[2]
			payload.Env[key] = value
		}
	}
	// Command
	cmdVars, _ := cmd.Flags().GetStringSlice("cmd")
	for _, c := range cmdVars {
		payload.Cmd = append(payload.Cmd, c)
	}
	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	var out bytes.Buffer
	json.Indent(&out, b, "", "  ")
	out.WriteTo(os.Stdout)
}

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"siderite/iron"

	"github.com/spf13/cobra"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "encrypts input with the cluster public key",
	Long:  `encrypts input (stdin or file) using the cluster public key`,
	Run:   encrypt,
}

func init() {
	rootCmd.AddCommand(encryptCmd)

	encryptCmd.Flags().StringP("keyfile", "k", "", "public key. Looks in ~/.iron.json otherwise")
	encryptCmd.Flags().StringP("infile", "i", "", "input file. Default is standard input")
}

func encrypt(cmd *cobra.Command, args []string) {
	var key string
	var err error

	// Key
	keyFile, _ := cmd.Flags().GetString("pubkey")
	if keyFile != "" {
		keyBytes, err := ioutil.ReadFile(keyFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		key = string(keyBytes)
	} else { // Read from ~/.iron.json
		config, err := iron.LoadConfig()
		if err != nil {
			fmt.Println(err)
			return
		}
		if key = config.ClusterInfo[0].Pubkey; key == "" {
			fmt.Println("missing public key in configuration")
			return
		}
	}

	// Input
	inFile, _ := cmd.Flags().GetString("infile")
	var input *os.File
	if inFile == "" {
		input = os.Stdin
	} else {
		input, err = os.Open(inFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer input.Close()
	}
	payload, err := ioutil.ReadAll(input)
	if err != nil {
		fmt.Println(err)
		return
	}
	ciphertext, err := iron.EncryptPayload([]byte(key), payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintf(os.Stdout, ciphertext+"\n")
}

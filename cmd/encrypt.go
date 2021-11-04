package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	config "github.com/philips-labs/siderite/iron"
	"github.com/philips-software/go-hsdp-api/iron"

	"github.com/spf13/cobra"
)

func NewEncryptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "encrypt",
		Short: "encrypts input with the cluster public key",
		Long:  `encrypts input (stdin or file) using the cluster public key`,
		RunE:  encrypt,
	}
}

func init() {
	cmd := NewEncryptCmd()
	rootCmd.AddCommand(cmd)

	cmd.Flags().StringP("keyfile", "k", "", "public key. Looks in ~/.iron.json otherwise")
	cmd.Flags().StringP("infile", "i", "", "input file. Default is standard input")
}

func encrypt(cmd *cobra.Command, _ []string) error {
	var key string
	var err error

	// Key
	keyFile, _ := cmd.Flags().GetString("pubkey")
	if keyFile != "" {
		keyBytes, err := ioutil.ReadFile(keyFile)
		if err != nil {
			fmt.Println(err)
			return err
		}
		key = string(keyBytes)
	} else { // Read from ~/.iron.json
		cfg, err := config.Load()
		if err != nil {
			fmt.Println(err)
			return err
		}
		if len(cfg.ClusterInfo) == 0 {
			fmt.Println("missing cluster_info in configuration")
			return err
		}
		if key = cfg.ClusterInfo[0].Pubkey; key == "" {
			fmt.Println("missing public key in configuration")
			return err
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
			return err
		}
		defer func() {
			_ = input.Close()
		}()
	}
	payload, err := ioutil.ReadAll(input)
	if err != nil {
		fmt.Println(err)
		return err
	}
	ciphertext, err := iron.EncryptPayload([]byte(key), payload)
	if err != nil {
		fmt.Println(err)
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, ciphertext+"\n")
	return nil
}

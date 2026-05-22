package cmd

import (
	"browser-bridge/cli/client"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "browser-bridge",
	Short: "Browser automation bridge for AI agents",
	Long:  "CLI tool to control Chrome browser via browser-bridge native-host",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("format", "json", "Output format: json or table")
}

func getClient() (*client.Client, error) {
	return client.NewClient()
}

func outputResult(data interface{}, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	out, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(out))
}

func outputResponse(resp *client.APIResponse, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	if !resp.Success {
		fmt.Fprintln(os.Stderr, "Error:", resp.Error)
		os.Exit(1)
	}
	fmt.Println(string(resp.Data))
}

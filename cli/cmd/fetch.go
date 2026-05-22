package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch web content",
}

var fetchContentCmd = &cobra.Command{
	Use:   "content <url>",
	Short: "Fetch URL content as markdown (opens tab, extracts content, closes tab)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		timeout, _ := cmd.Flags().GetInt("timeout")
		output, _ := cmd.Flags().GetString("output")

		body := map[string]interface{}{
			"url": args[0],
		}
		if timeout > 0 {
			body["timeout"] = timeout
		}

		resp, err := c.Post("/fetch/content", body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintln(os.Stderr, "Error:", resp.Error)
			os.Exit(1)
		}

		if output != "" {
			var result struct {
				URL     string `json:"url"`
				Content string `json:"content"`
				Format  string `json:"format"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			if err := os.WriteFile(output, []byte(result.Content), 0644); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Content saved to %s\n", output)
		} else {
			fmt.Println(string(resp.Data))
		}
	},
}

func init() {
	fetchContentCmd.Flags().Int("timeout", 15000, "Page load timeout in milliseconds")
	fetchContentCmd.Flags().StringP("output", "o", "", "Save markdown content to file")
	fetchCmd.AddCommand(fetchContentCmd)
	rootCmd.AddCommand(fetchCmd)
}

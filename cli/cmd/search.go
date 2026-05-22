package cmd

import (
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search the web",
}

var searchBaiduCmd = &cobra.Command{
	Use:   "baidu <query>",
	Short: "Search using Baidu",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		outputResponse(c.Post("/search", map[string]interface{}{
			"engine": "baidu",
			"query":  args[0],
		}))
	},
}

var searchBingCmd = &cobra.Command{
	Use:   "bing <query>",
	Short: "Search using Bing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		outputResponse(c.Post("/search", map[string]interface{}{
			"engine": "bing",
			"query":  args[0],
		}))
	},
}

func init() {
	searchCmd.AddCommand(searchBaiduCmd)
	searchCmd.AddCommand(searchBingCmd)
	rootCmd.AddCommand(searchCmd)
}

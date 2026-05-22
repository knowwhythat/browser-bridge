package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cookieCmd = &cobra.Command{
	Use:   "cookie",
	Short: "Cookie operations",
}

var cookieGetCmd = &cobra.Command{
	Use:   "get [url]",
	Short: "Get cookies",
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		path := "/cookies"
		if len(args) > 0 {
			path += "?url=" + args[0]
		}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			if len(args) > 0 {
				path += "&name=" + name
			} else {
				path += "?name=" + name
			}
		}
		outputResponse(c.Get(path))
	},
}

var cookieSetCmd = &cobra.Command{
	Use:   "set <url> <name> <value>",
	Short: "Set a cookie",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		body := map[string]interface{}{
			"url":   args[0],
			"name":  args[1],
			"value": args[2],
		}
		outputResponse(c.Post("/cookies", body))
	},
}

var cookieDeleteCmd = &cobra.Command{
	Use:   "delete <url> <name>",
	Short: "Delete a cookie",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		path := fmt.Sprintf("/cookies?url=%s&name=%s", args[0], args[1])
		outputResponse(c.Delete(path))
	},
}

func init() {
	cookieGetCmd.Flags().String("name", "", "Cookie name to filter")
	cookieCmd.AddCommand(cookieGetCmd)
	cookieCmd.AddCommand(cookieSetCmd)
	cookieCmd.AddCommand(cookieDeleteCmd)
	rootCmd.AddCommand(cookieCmd)
}

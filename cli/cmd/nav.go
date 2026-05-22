package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var navCmd = &cobra.Command{
	Use:   "nav",
	Short: "Navigation operations",
}

var navGotoCmd = &cobra.Command{
	Use:   "goto <tabId> <url>",
	Short: "Navigate to URL",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		url := args[1]
		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/navigate", tabID), map[string]interface{}{
			"url": url,
		}))
	},
}

var navBackCmd = &cobra.Command{
	Use:   "back <tabId>",
	Short: "Go back",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		outputResponse(c.Get(fmt.Sprintf("/tabs/%d/back", tabID)))
	},
}

var navForwardCmd = &cobra.Command{
	Use:   "forward <tabId>",
	Short: "Go forward",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		outputResponse(c.Get(fmt.Sprintf("/tabs/%d/forward", tabID)))
	},
}

var navReloadCmd = &cobra.Command{
	Use:   "reload <tabId>",
	Short: "Reload the page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		outputResponse(c.Get(fmt.Sprintf("/tabs/%d/reload", tabID)))
	},
}

func init() {
	navCmd.AddCommand(navGotoCmd)
	navCmd.AddCommand(navBackCmd)
	navCmd.AddCommand(navForwardCmd)
	navCmd.AddCommand(navReloadCmd)
	rootCmd.AddCommand(navCmd)
}

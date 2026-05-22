package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var tabCmd = &cobra.Command{
	Use:   "tab",
	Short: "Manage browser tabs",
}

var tabListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all open tabs",
	Aliases: []string{"ls"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getClient()
		outputResponse(c.Get("/tabs"))
		_ = err
	},
}

var tabGetCmd = &cobra.Command{
	Use:   "get <tabId>",
	Short: "Get tab info by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		outputResponse(c.Get(fmt.Sprintf("/tabs/%d", tabID)))
	},
}

var tabCreateCmd = &cobra.Command{
	Use:   "create [url]",
	Short: "Create a new tab",
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		body := map[string]interface{}{}
		if len(args) > 0 {
			body["url"] = args[0]
		}
		active, _ := cmd.Flags().GetBool("active")
		body["active"] = active
		outputResponse(c.Post("/tabs", body))
	},
}

var tabCloseCmd = &cobra.Command{
	Use:   "close <tabId>",
	Short: "Close a tab",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		outputResponse(c.Get(fmt.Sprintf("/tabs/%d/close", tabID)))
	},
}

var tabActivateCmd = &cobra.Command{
	Use:   "activate <tabId>",
	Short: "Activate a tab",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		outputResponse(c.Get(fmt.Sprintf("/tabs/%d/activate", tabID)))
	},
}

func init() {
	tabCreateCmd.Flags().Bool("active", true, "Make the tab active")
	tabCmd.AddCommand(tabListCmd)
	tabCmd.AddCommand(tabGetCmd)
	tabCmd.AddCommand(tabCreateCmd)
	tabCmd.AddCommand(tabCloseCmd)
	tabCmd.AddCommand(tabActivateCmd)
	rootCmd.AddCommand(tabCmd)
}

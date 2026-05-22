package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var pageCmd = &cobra.Command{
	Use:   "page",
	Short: "Page operations",
}

var pageContentCmd = &cobra.Command{
	Use:   "content <tabId>",
	Short: "Get page content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		format, _ := cmd.Flags().GetString("format")
		path := fmt.Sprintf("/tabs/%d/content", tabID)
		if format != "" {
			path += "?format=" + format
		}
		outputResponse(c.Get(path))
	},
}

var pageSnapshotCmd = &cobra.Command{
	Use:   "snapshot <tabId>",
	Short: "Get page snapshot with refs for interactive elements",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		refOnly, _ := cmd.Flags().GetBool("ref-only")
		path := fmt.Sprintf("/tabs/%d/snapshot", tabID)
		if refOnly {
			path += "?refOnly=true"
		}
		outputResponse(c.Get(path))
	},
}

var pageScreenshotCmd = &cobra.Command{
	Use:   "screenshot <tabId>",
	Short: "Take a screenshot of the page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		body := map[string]interface{}{}

		if ref, _ := cmd.Flags().GetString("ref"); ref != "" {
			body["ref"] = ref
		}
		if selector, _ := cmd.Flags().GetString("selector"); selector != "" {
			body["selector"] = selector
		}
		if fullPage, _ := cmd.Flags().GetBool("full-page"); fullPage {
			body["fullPage"] = true
		}
		if format, _ := cmd.Flags().GetString("format"); format != "" {
			body["format"] = format
		}

		resp, err := c.Post(fmt.Sprintf("/tabs/%d/screenshot", tabID), body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		if !resp.Success {
			fmt.Fprintln(os.Stderr, "Error:", resp.Error)
			os.Exit(1)
		}

		output, _ := cmd.Flags().GetString("output")
		if output != "" {
			// Parse response to extract base64 data
			var result struct {
				Data string `json:"data"`
				Format string `json:"format"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				fmt.Fprintln(os.Stderr, "Error: failed to parse screenshot response:", err)
				os.Exit(1)
			}
			imgData, err := base64.StdEncoding.DecodeString(result.Data)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error: failed to decode base64:", err)
				os.Exit(1)
			}
			if err := os.WriteFile(output, imgData, 0644); err != nil {
				fmt.Fprintln(os.Stderr, "Error: failed to write file:", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Screenshot saved to %s\n", output)
		} else {
			fmt.Println(string(resp.Data))
		}
	},
}

var pageExecuteCmd = &cobra.Command{
	Use:   "execute <tabId> <script>",
	Short: "Execute JavaScript in the page",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		script := args[1]

		// 支持从文件加载脚本
		if file, _ := cmd.Flags().GetString("file"); file != "" {
			outputResponse(c.Post(fmt.Sprintf("/tabs/%d/execute", tabID), map[string]interface{}{
				"script": script,
				"file":   file,
			}))
			return
		}

		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/execute", tabID), map[string]interface{}{
			"script": script,
		}))
	},
}

var pageClickCmd = &cobra.Command{
	Use:   "click <tabId>",
	Short: "Click an element (use --ref or --selector)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		body := map[string]interface{}{}
		if ref, _ := cmd.Flags().GetString("ref"); ref != "" {
			body["ref"] = ref
		}
		if selector, _ := cmd.Flags().GetString("selector"); selector != "" {
			body["selector"] = selector
		}
		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/click", tabID), body))
	},
}

var pageTypeCmd = &cobra.Command{
	Use:   "type <tabId> <text>",
	Short: "Type text into an element (use --ref or --selector)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		text := args[1]
		body := map[string]interface{}{"text": text}
		if ref, _ := cmd.Flags().GetString("ref"); ref != "" {
			body["ref"] = ref
		}
		if selector, _ := cmd.Flags().GetString("selector"); selector != "" {
			body["selector"] = selector
		}
		if clear, _ := cmd.Flags().GetBool("clear"); clear {
			body["clear"] = true
		}
		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/type", tabID), body))
	},
}

var pageSelectCmd = &cobra.Command{
	Use:   "select <tabId> <value>",
	Short: "Select an option in a dropdown (use --ref or --selector)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		value := args[1]
		body := map[string]interface{}{"value": value}
		if ref, _ := cmd.Flags().GetString("ref"); ref != "" {
			body["ref"] = ref
		}
		if selector, _ := cmd.Flags().GetString("selector"); selector != "" {
			body["selector"] = selector
		}
		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/select", tabID), body))
	},
}

var pageScrollCmd = &cobra.Command{
	Use:   "scroll <tabId>",
	Short: "Scroll the page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		body := map[string]interface{}{}
		if ref, _ := cmd.Flags().GetString("ref"); ref != "" {
			body["ref"] = ref
		}
		if selector, _ := cmd.Flags().GetString("selector"); selector != "" {
			body["selector"] = selector
		}
		if x, _ := cmd.Flags().GetInt("x"); x != 0 {
			body["x"] = x
		}
		if y, _ := cmd.Flags().GetInt("y"); y != 0 {
			body["y"] = y
		}
		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/scroll", tabID), body))
	},
}

var pageQueryCmd = &cobra.Command{
	Use:   "query <tabId> <selector>",
	Short: "Query DOM elements",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		selector := args[1]
		body := map[string]interface{}{"selector": selector}
		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/query", tabID), body))
	},
}

var pageWaitCmd = &cobra.Command{
	Use:   "wait <tabId> <selector>",
	Short: "Wait for an element to appear",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c, _ := getClient()
		tabID, _ := strconv.Atoi(args[0])
		selector := args[1]
		body := map[string]interface{}{"selector": selector}
		if timeout, _ := cmd.Flags().GetInt("timeout"); timeout > 0 {
			body["timeout"] = timeout
		}
		outputResponse(c.Post(fmt.Sprintf("/tabs/%d/wait", tabID), body))
	},
}

func init() {
	// snapshot
	pageSnapshotCmd.Flags().Bool("ref-only", false, "Only return elements with refs")

	// screenshot
	pageScreenshotCmd.Flags().String("ref", "", "Element ref to screenshot")
	pageScreenshotCmd.Flags().String("selector", "", "CSS selector to screenshot")
	pageScreenshotCmd.Flags().Bool("full-page", false, "Capture full page")
	pageScreenshotCmd.Flags().String("format", "png", "Image format: png or jpeg")
	pageScreenshotCmd.Flags().StringP("output", "o", "", "Save screenshot to file path")

	// content
	pageContentCmd.Flags().String("format", "html", "Content format: html or markdown")

	// execute
	pageExecuteCmd.Flags().String("file", "", "Load script from file")

	// click
	pageClickCmd.Flags().String("ref", "", "Element ref to click")
	pageClickCmd.Flags().String("selector", "", "CSS selector to click")

	// type
	pageTypeCmd.Flags().String("ref", "", "Element ref to type into")
	pageTypeCmd.Flags().String("selector", "", "CSS selector to type into")
	pageTypeCmd.Flags().Bool("clear", false, "Clear existing text before typing")

	// select
	pageSelectCmd.Flags().String("ref", "", "Element ref of dropdown")
	pageSelectCmd.Flags().String("selector", "", "CSS selector of dropdown")

	// scroll
	pageScrollCmd.Flags().String("ref", "", "Element ref to scroll to")
	pageScrollCmd.Flags().String("selector", "", "CSS selector to scroll to")
	pageScrollCmd.Flags().Int("x", 0, "Horizontal scroll offset")
	pageScrollCmd.Flags().Int("y", 0, "Vertical scroll offset")

	// wait
	pageWaitCmd.Flags().Int("timeout", 5000, "Timeout in milliseconds")

	pageCmd.AddCommand(pageContentCmd)
	pageCmd.AddCommand(pageSnapshotCmd)
	pageCmd.AddCommand(pageScreenshotCmd)
	pageCmd.AddCommand(pageExecuteCmd)
	pageCmd.AddCommand(pageClickCmd)
	pageCmd.AddCommand(pageTypeCmd)
	pageCmd.AddCommand(pageSelectCmd)
	pageCmd.AddCommand(pageScrollCmd)
	pageCmd.AddCommand(pageQueryCmd)
	pageCmd.AddCommand(pageWaitCmd)
	rootCmd.AddCommand(pageCmd)
}

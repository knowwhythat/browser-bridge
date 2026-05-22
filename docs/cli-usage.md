# Browser Bridge CLI

CLI tool to control Chrome browser via browser-bridge native-host.

## Prerequisites

- native-host is running (started automatically by Chrome when extension connects)
- Port file `nativehost_port` exists next to `native-host.exe` (or set `BROWSER_BRIDGE_PORT` env var)

## Global Flags

| Flag       | Default | Description                  |
| ---------- | ------- | ---------------------------- |
| `--format` | `json`  | Output format: json or table |

---

## Tab Management

### List all open tabs

```bash
browser-bridge tab list
browser-bridge tab ls
```

### Get tab info by ID

```bash
browser-bridge tab get <tabId>
```

### Create a new tab

```bash
browser-bridge tab create [url]
browser-bridge tab create https://www.baidu.com
browser-bridge tab create                    # open blank tab
```

| Flag       | Default | Description         |
| ---------- | ------- | ------------------- |
| `--active` | `true`  | Make the tab active |

### Close a tab

```bash
browser-bridge tab close <tabId>
```

### Activate a tab

```bash
browser-bridge tab activate <tabId>
```

---

## Page Operations

### Get page content

```bash
browser-bridge page content <tabId>
browser-bridge page content <tabId> --format markdown
```

| Flag       | Default | Description                                                        |
| ---------- | ------- | ------------------------------------------------------------------ |
| `--format` | `html`  | Content format: `html` (raw HTML) or `markdown` (cleaned markdown) |

When `--format markdown` is used, the extension strips navigation, ads, scripts, and other non-content elements, then converts the remaining content to markdown.

### Get page snapshot with refs

Returns an accessibility-tree-based snapshot of the page. Interactive elements are assigned unique `ref` identifiers that can be used with `click`, `type`, `select`, `scroll` commands.

```bash
browser-bridge page snapshot <tabId>
browser-bridge page snapshot <tabId> --ref-only
```

| Flag         | Default | Description                    |
| ------------ | ------- | ------------------------------ |
| `--ref-only` | `false` | Only return elements with refs |

### Take a screenshot

```bash
browser-bridge page screenshot <tabId>
browser-bridge page screenshot <tabId> --full-page
browser-bridge page screenshot <tabId> --ref btn1
browser-bridge page screenshot <tabId> --selector "#header"
browser-bridge page screenshot <tabId> --format jpeg
browser-bridge page screenshot <tabId> -o screenshot.png
browser-bridge page screenshot <tabId> -o shot.jpg --format jpeg
```

| Flag           | Default | Description                              |
| -------------- | ------- | ---------------------------------------- |
| `--ref`        |         | Element ref to screenshot                |
| `--selector`   |         | CSS selector to screenshot               |
| `--full-page`  | `false` | Capture full page                        |
| `--format`     | `png`   | Image format: png or jpeg                |
| `-o, --output` |         | Save screenshot to file (base64 decoded) |

### Execute JavaScript

```bash
browser-bridge page execute <tabId> "document.title"
browser-bridge page execute <tabId> "1+1"
```

| Flag     | Default | Description           |
| -------- | ------- | --------------------- |
| `--file` |         | Load script from file |

### Click an element

```bash
browser-bridge page click <tabId> --ref btn1
browser-bridge page click <tabId> --selector "#submit-btn"
```

| Flag         | Default | Description           |
| ------------ | ------- | --------------------- |
| `--ref`      |         | Element ref to click  |
| `--selector` |         | CSS selector to click |

### Type text into an element

```bash
browser-bridge page type <tabId> "hello world" --ref input1
browser-bridge page type <tabId> "search query" --selector "#search" --clear
```

| Flag         | Default | Description                       |
| ------------ | ------- | --------------------------------- |
| `--ref`      |         | Element ref to type into          |
| `--selector` |         | CSS selector to type into         |
| `--clear`    | `false` | Clear existing text before typing |

### Select an option in a dropdown

```bash
browser-bridge page select <tabId> "option_value" --ref sel1
browser-bridge page select <tabId> "US" --selector "#country"
```

| Flag         | Default | Description              |
| ------------ | ------- | ------------------------ |
| `--ref`      |         | Element ref of dropdown  |
| `--selector` |         | CSS selector of dropdown |

### Scroll the page

```bash
browser-bridge page scroll <tabId> --y 500
browser-bridge page scroll <tabId> --ref section2
browser-bridge page scroll <tabId> --selector "#footer" --x 0 --y 100
```

| Flag         | Default | Description               |
| ------------ | ------- | ------------------------- |
| `--ref`      |         | Element ref to scroll to  |
| `--selector` |         | CSS selector to scroll to |
| `--x`        | `0`     | Horizontal scroll offset  |
| `--y`        | `0`     | Vertical scroll offset    |

### Query DOM elements

```bash
browser-bridge page query <tabId> "a[href]"
browser-bridge page query <tabId> ".product-item"
```

### Wait for an element to appear

```bash
browser-bridge page wait <tabId> "#results"
browser-bridge page wait <tabId> ".loaded" --timeout 10000
```

| Flag        | Default | Description             |
| ----------- | ------- | ----------------------- |
| `--timeout` | `5000`  | Timeout in milliseconds |

---

## Navigation

### Navigate to URL

```bash
browser-bridge nav goto <tabId> <url>
browser-bridge nav goto 1 https://www.google.com
```

### Go back

```bash
browser-bridge nav back <tabId>
```

### Go forward

```bash
browser-bridge nav forward <tabId>
```

### Reload the page

```bash
browser-bridge nav reload <tabId>
```

---

## Search

Search returns a list of results with title, url, and snippet. The search tab is automatically closed after results are extracted.

### Search using Baidu

```bash
browser-bridge search baidu "golang tutorial"
```

Response example:

```json
{
  "engine": "baidu",
  "query": "golang tutorial",
  "results": [
    {
      "title": "Go 语言教程 | 菜鸟教程",
      "url": "https://www.runoob.com/go/go-tutorial.html",
      "snippet": "Go 是一个开源的编程语言..."
    }
  ]
}
```

### Search using Bing

```bash
browser-bridge search bing "rust vs go"
```

Response example:

```json
{
  "engine": "bing",
  "query": "rust vs go",
  "results": [
    {
      "title": "Rust vs Go | Comparison",
      "url": "https://example.com/rust-vs-go",
      "snippet": "A detailed comparison..."
    }
  ]
}
```

---

## Cookie Operations

### Get cookies

```bash
browser-bridge cookie get
browser-bridge cookie get https://www.example.com
browser-bridge cookie get https://www.example.com --name session_id
```

| Flag     | Default | Description           |
| -------- | ------- | --------------------- |
| `--name` |         | Cookie name to filter |

### Set a cookie

```bash
browser-bridge cookie set <url> <name> <value>
browser-bridge cookie set https://www.example.com token abc123
```

### Delete a cookie

```bash
browser-bridge cookie delete <url> <name>
browser-bridge cookie delete https://www.example.com token
```

---

## Typical Workflow

```bash
# 1. List tabs to find the target tabId
browser-bridge tab list

# 2. Get a snapshot to understand the page structure
browser-bridge page snapshot 1

# 3. Interact with elements using refs from the snapshot
browser-bridge page click 1 --ref input1
browser-bridge page type 1 "browser automation" --ref input1

# 4. Take a screenshot to verify the result
browser-bridge page screenshot 1

# 5. Navigate to another page
browser-bridge nav goto 1 https://www.bing.com

# 6. Search via shortcut
browser-bridge search bing "browser automation"
```

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working in this repository.

## Project Overview

**Browser Bridge** is a browser automation tool for AI agents. It consists of three components that together allow programmatic control of Chrome:

1. **Chrome Extension** (`extension/`) — Manifest V3 service worker that uses Chrome APIs (tabs, scripting, cookies, nativeMessaging) to execute browser actions
2. **Native Host** (`native-host/`) — Go binary that bridges the extension to a local HTTP API via Chrome's Native Messaging protocol (stdin/stdout with 4-byte length-prefixed JSON)
3. **CLI** (`cli/`) — Go Cobra CLI tool that talks HTTP to the native host, providing commands like `tab list`, `page snapshot`, `page click`, etc.

### Communication Flow

```
CLI (HTTP) → Native Host (HTTP server on localhost) ↔ Extension (Native Messaging stdin/stdout) → Chrome Tabs
```

The native host auto-picks a port in 3000–4000, writes it to `~/.browser-bridge/nativehost_port`, and the CLI reads that file (or `BROWSER_BRIDGE_PORT` env var) to discover it.

## Tech Stack

- **Go 1.23.2** — native-host and CLI (two separate Go modules)
- **TypeScript** — Chrome extension (ES2020, bundled with esbuild)
- **Cobra** — CLI framework
- **Chrome APIs** — tabs, scripting, cookies, nativeMessaging

## Build Commands

### Build everything (from install script)
```bash
cd native-host && go build -o native-host.exe .
cd cli && go build -o browser-bridge.exe .
cd extension && npm run build
```

### Extension development
```bash
cd extension
npm run build      # production build (minified)
npm run dev        # watch mode with sourcemaps
```

The extension build uses esbuild to bundle `src/background.ts` → `dist/background.js`, then copies `manifest.json` and `icons/` into `dist/`. Load `extension/dist` as an unpacked extension in Chrome.

### Go modules
There are **two independent Go modules**: `browser-bridge/cli` and `browser-bridge/native-host`. Run `go build` from within each directory.

## Architecture Details

### Message Protocol
All messages between native host and extension use a JSON format with a 4-byte little-endian length prefix on the wire:
```json
{ "id": "uuid", "action": "page.snapshot", "params": { "tabId": 1 } }
```
Responses: `{ "id": "uuid", "success": true, "data": {...} }` or `{ "success": false, "error": "..." }`

### Extension (`extension/src/`)
- **`background.ts`** — Service worker entry point. Connects to native host via `chrome.runtime.connectNative()`, dispatches actions to handler modules using a switch statement
- **`types.ts`** — All TypeScript interfaces for the message protocol, action types, and result types
- **`actions/page.ts`** — Page operations: content extraction (HTML or markdown), screenshots (with element clipping via OffscreenCanvas), script execution, click/type/select/scroll/query/wait, and fetch (open URL → extract markdown → close tab)
- **`actions/snapshot.ts`** — Accessibility-tree-based page snapshot. Walks the DOM, assigns `ref` identifiers (e.g. `r1`, `r2`) to interactive elements, maps refs to CSS selectors. Supports `refOnly` mode for compact output
- **`actions/tab.ts`** — Tab CRUD: list, get, create, close, activate
- **`actions/search.ts`** — Search via Baidu or Bing: opens a temporary tab, waits for load, extracts results, closes tab

### Native Host (`native-host/`)
- **`main.go`** — Starts the Native Messaging bridge (stdin pump), finds an available port, writes port file, starts HTTP server
- **`native/bridge.go`** — Manages bidirectional Native Messaging: `PumpStdin()` reads length-prefixed JSON from stdin, `Send()` writes to stdout, `SendAndWait()` correlates requests/responses via UUID-matched channels with 30s timeout
- **`handler/handler.go`** — HTTP handler that translates REST API calls into Native Messaging messages. All handlers follow the same pattern: build params → `sendAndWait(action, params)` → return JSON response. Routes registered at `/api/v1/*`
- **`model/types.go`** — Shared request/response structs

### CLI (`cli/`)
- **`main.go`** — Calls `cmd.Execute()`
- **`cmd/`** — Cobra command definitions: `root.go` (base), `tab.go`, `page.go`, `nav.go`, `cookie.go`, `search.go`, `fetch.go`
- **`client/client.go`** — HTTP client that reads the port file to discover the native host, provides `Get/Post/Delete/Put` methods

### Key Design Patterns
- **Ref system**: The snapshot assigns refs like `r1`, `r2` to interactive elements. These are mapped to CSS selectors stored in memory (per tabId). Commands like `click`, `type`, `select` accept `--ref r1` which gets resolved to a CSS selector before injection into the page
- **Content extraction**: Both `page.content` and `fetch.content` strip non-content elements (nav, footer, ads, scripts) and convert HTML to markdown
- **Search**: Opens a background tab, waits for load, scrapes results, closes the tab — all transparent to the user

## Installation

Run `install/install.bat` which builds all three components, then registers the native messaging host in the Windows registry (`HKCU\Software\Google\Chrome\NativeMessagingHosts\com.browser.bridge`). After that, load `extension/dist` as an unpacked Chrome extension.

## API Routes (Native Host)

All routes are under `/api/v1/`:
- `GET /health` — Health check
- `GET/POST /tabs` — List/create tabs
- `GET /tabs/{tabId}` — Get tab info
| `GET /tabs/{tabId}/close` — Close tab
| `GET /tabs/{tabId}/activate` — Activate tab
| `GET /tabs/{tabId}/snapshot` — Page snapshot (`?refOnly=true`)
| `GET /tabs/{tabId}/content` — Page content (`?format=html|markdown`)
| `POST /tabs/{tabId}/screenshot` — Screenshot
| `POST /tabs/{tabId}/click|type|select|scroll|query|wait|execute` — Page interactions
| `POST /tabs/{tabId}/navigate` — Navigate to URL
| `GET /tabs/{tabId}/back|forward|reload` — Navigation
| `GET/POST/DELETE /cookies` — Cookie operations
| `POST /search` — Search (`{"engine":"baidu|bing","query":"..."}`)
| `POST /fetch/content` — Fetch URL as markdown

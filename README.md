# Browser Bridge

Browser Bridge 是一个面向 AI Agent 的浏览器自动化工具，通过三层架构实现对 Chrome 浏览器的程序化控制：

```
┌──────────────────┐     HTTP     ┌──────────────┐  Native Msg  ┌──────────────────┐
│ browser-bridge-cli│ ──────────▶ │  native-host  │ ──────────▶ │ Chrome Extension  │
│   (Go CLI 工具)   │ ◀────────── │  (Go HTTP 服务)│ ◀────────── │   (浏览器插件)     │
└──────────────────┘     HTTP     └──────────────┘  Native Msg  └──────────────────┘
```

## 架构概览

| 组件 | 目录 | 语言 | 职责 |
|------|------|------|------|
| **Chrome Extension** | `extension/` | TypeScript | 通过 Chrome API 操作浏览器（标签页、脚本执行、截图、Cookie 等） |
| **Native Host** | `native-host/` | Go | 桥接 Extension 和 CLI：通过 Native Messaging 与 Extension 通信，通过 HTTP 与 CLI 通信 |
| **CLI** | `cli/` | Go | 命令行工具，提供用户/API 友好的浏览器操作接口 |

### 通信协议

- **CLI → Native Host**: HTTP REST API（`http://127.0.0.1:{port}/api/v1`）
- **Native Host ↔ Extension**: Chrome Native Messaging（4 字节 little-endian 长度前缀 + JSON）
- **请求-响应匹配**: UUID + pending channel map，30 秒超时

## 功能特性

- **Tab 管理** — 列出、创建、关闭、激活标签页
- **页面快照** — 基于可访问性树的页面结构快照，为可交互元素自动分配 `ref` 标识符
- **页面内容** — 获取原始 HTML 或清理后的 Markdown
- **页面截图** — 支持全页截图、元素截图（通过 ref 或 selector）
- **页面交互** — 点击、输入、选择下拉框、滚动、查询 DOM、等待元素
- **脚本执行** — 在页面上下文中执行 JavaScript
- **导航** — 前进、后退、刷新、等待页面加载
- **搜索** — 百度搜索、Bing 搜索（自动提取结果并关闭临时标签页）
- **Cookie 操作** — 获取、设置、删除 Cookie
- **内容抓取** — 打开 URL → 提取 Markdown 内容 → 自动关闭标签页

## 快速开始

### 前置条件

- Go 1.23+
- Node.js（用于构建 Chrome Extension）
- Chrome 浏览器（开发者模式）

### 安装

运行安装脚本自动完成编译、注册和配置：

```bash
install/install.bat
```

安装完成后，按提示在 Chrome 中加载 `extension/dist` 目录作为 unpacked 扩展。

### 手动构建

**构建 Native Host:**
```bash
cd native-host
go build -o native-host.exe .
```

**构建 CLI:**
```bash
cd cli
go build -o browser-bridge.exe .
```

**构建 Chrome Extension:**
```bash
cd extension
npm install
npm run build      # 生产构建
npm run dev        # 开发模式（watch + sourcemap）
```

### Native Messaging Host 注册

在 Windows 注册表中写入：
```
HKCU\Software\Google\Chrome\NativeMessagingHosts\com.browser.bridge
```

值为 `install/com.browser.bridge.json` 的路径。该 JSON 文件指定了 native-host 可执行文件路径和允许的 Extension origin。

## CLI 用法

### Tab 管理

```bash
browser-bridge tab list                              # 列出所有标签页
browser-bridge tab get <tabId>                       # 获取标签页信息
browser-bridge tab create [url]                      # 创建新标签页
browser-bridge tab close <tabId>                     # 关闭标签页
browser-bridge tab activate <tabId>                  # 激活标签页
```

### 页面操作

```bash
# 快照（可访问性树，含 ref）
browser-bridge page snapshot <tabId>                 # 完整快照
browser-bridge page snapshot <tabId> --ref-only      # 仅可交互元素

# 内容提取
browser-bridge page content <tabId>                  # 获取 HTML
browser-bridge page content <tabId> --format markdown # 获取 Markdown

# 截图
browser-bridge page screenshot <tabId>               # 截图（base64 JSON）
browser-bridge page screenshot <tabId> -o shot.png   # 保存到文件
browser-bridge page screenshot <tabId> --full-page   # 全页截图
browser-bridge page screenshot <tabId> --ref r1      # 截取指定元素

# 页面交互（支持 --ref 或 --selector）
browser-bridge page click <tabId> --ref r1           # 点击元素
browser-bridge page type <tabId> "文本" --ref r2     # 输入文本
browser-bridge page select <tabId> "值" --ref r3     # 选择下拉框
browser-bridge page scroll <tabId> --y 500           # 滚动页面
browser-bridge page query <tabId> "a[href]"          # 查询 DOM
browser-bridge page wait <tabId> "#results"          # 等待元素

# 脚本执行
browser-bridge page execute <tabId> "document.title" # 执行 JS
```

### 导航

```bash
browser-bridge nav goto <tabId> <url>                # 导航到 URL
browser-bridge nav back <tabId>                      # 后退
browser-bridge nav forward <tabId>                   # 前进
browser-bridge nav reload <tabId>                    # 刷新
```

### 搜索

```bash
browser-bridge search baidu "查询内容"               # 百度搜索
browser-bridge search bing "查询内容"                # Bing 搜索
```

### Cookie 操作

```bash
browser-bridge cookie get [url]                      # 获取 Cookie
browser-bridge cookie set <url> <name> <value>      # 设置 Cookie
browser-bridge cookie delete <url> <name>            # 删除 Cookie
```

### 内容抓取

```bash
browser-bridge fetch content <url>                   # 抓取 URL 内容为 Markdown
browser-bridge fetch content <url> -o output.md      # 保存到文件
```

### 典型工作流

```bash
# 1. 列出标签页，找到目标 tabId
browser-bridge tab list

# 2. 获取快照，了解页面结构
browser-bridge page snapshot 1

# 3. 使用 ref 与页面元素交互
browser-bridge page click 1 --ref r1
browser-bridge page type 1 "browser automation" --ref r2

# 4. 截图验证结果
browser-bridge page screenshot 1 -o result.png

# 5. 导航到新页面
browser-bridge nav goto 1 https://www.bing.com

# 6. 搜索
browser-bridge search bing "browser automation"
```

## HTTP API

Native Host 在 `http://127.0.0.1:{port}` 上提供 REST API（端口自动分配，写入 `~/.browser-bridge/nativehost_port`）：

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/health` | 健康检查 |
| `GET` | `/api/v1/tabs` | 列出所有标签页 |
| `POST` | `/api/v1/tabs` | 创建标签页 |
| `GET` | `/api/v1/tabs/{tabId}` | 获取标签页信息 |
| `GET` | `/api/v1/tabs/{tabId}/close` | 关闭标签页 |
| `GET` | `/api/v1/tabs/{tabId}/activate` | 激活标签页 |
| `GET` | `/api/v1/tabs/{tabId}/snapshot` | 页面快照（`?refOnly=true`） |
| `GET` | `/api/v1/tabs/{tabId}/content` | 页面内容（`?format=html\|markdown`） |
| `POST` | `/api/v1/tabs/{tabId}/screenshot` | 页面截图 |
| `POST` | `/api/v1/tabs/{tabId}/click` | 点击元素 |
| `POST` | `/api/v1/tabs/{tabId}/type` | 输入文本 |
| `POST` | `/api/v1/tabs/{tabId}/select` | 选择下拉框 |
| `POST` | `/api/v1/tabs/{tabId}/scroll` | 滚动页面 |
| `POST` | `/api/v1/tabs/{tabId}/query` | 查询 DOM |
| `POST` | `/api/v1/tabs/{tabId}/wait` | 等待元素 |
| `POST` | `/api/v1/tabs/{tabId}/execute` | 执行 JS |
| `POST` | `/api/v1/tabs/{tabId}/navigate` | 导航到 URL |
| `GET` | `/api/v1/tabs/{tabId}/back\|forward\|reload` | 导航操作 |
| `GET/POST/DELETE` | `/api/v1/cookies` | Cookie 操作 |
| `POST` | `/api/v1/search` | 搜索 |
| `POST` | `/api/v1/fetch/content` | 抓取 URL 内容 |

所有 API 响应格式统一为：
```json
{ "success": true, "data": { ... } }
// 或
{ "success": false, "error": "错误信息" }
```

## 项目结构

```
browser-bridge/
├── extension/                  # Chrome 浏览器插件（TypeScript）
│   ├── src/
│   │   ├── background.ts       # Service Worker，Native Messaging 消息分发
│   │   ├── types.ts            # 消息协议与类型定义
│   │   └── actions/
│   │       ├── page.ts         # 页面操作（内容、截图、交互、脚本执行）
│   │       ├── snapshot.ts     # 页面快照与 ref 管理
│   │       ├── tab.ts          # Tab 管理
│   │       └── search.ts       # 搜索功能
│   ├── dist/                   # 编译输出
│   ├── manifest.json           # Manifest V3 配置
│   └── package.json
├── native-host/                # Native Host 程序（Go）
│   ├── main.go                 # 入口：启动 HTTP 服务 + Native Messaging
│   ├── handler/handler.go      # HTTP 请求处理器，转发到 Extension
│   ├── native/bridge.go        # Native Messaging 双向通信（长度前缀协议）
│   └── model/types.go          # 数据模型
├── cli/                        # CLI 工具（Go + Cobra）
│   ├── main.go                 # CLI 入口
│   ├── cmd/                    # 命令定义（tab/page/nav/cookie/search/fetch）
│   └── client/client.go        # HTTP 客户端，自动发现 native-host 端口
├── install/                    # 安装脚本与配置
│   ├── install.bat             # Windows 安装脚本
│   └── com.browser.bridge.json # Native Messaging Host 注册配置
└── docs/
    ├── cli-usage.md            # CLI 详细使用文档
    └── test-report.md          # 测试报告
```

## 核心设计

### Ref 系统

页面快照为可交互元素（按钮、链接、输入框等）分配唯一 `ref` 标识符（如 `r1`、`r2`），并映射到对应的 CSS Selector。Agent 通过 `--ref r1` 操作元素，无需关心底层选择器：

```
page.snapshot → 返回含 ref 的可访问性树
     ↓
page.click --ref r1 → 自动解析为 CSS Selector → 注入页面执行 click()
```

### 内容提取

`page content --format markdown` 和 `fetch content` 会：
1. 克隆页面 DOM，移除导航栏、广告、脚本等非内容元素
2. 将清理后的 HTML 转换为 Markdown
3. `fetch content` 会在后台标签页中完成提取后自动关闭

### 端口发现

Native Host 启动时在 3000–4000 范围内自动选择可用端口，写入 `~/.browser-bridge/nativehost_port`。CLI 读取该文件（或 `BROWSER_BRIDGE_PORT` 环境变量）来发现服务地址。

## 技术栈

| 组件 | 技术 |
|------|------|
| Extension | TypeScript, esbuild, Chrome Extension API (Manifest V3) |
| Native Host | Go 1.23, net/http, encoding/binary |
| CLI | Go 1.23, cobra, net/http |
| 通信协议 | Chrome Native Messaging (4-byte LE length prefix + JSON) |

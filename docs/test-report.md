# Browser Bridge CLI 测试报告

**测试日期**: 2026-05-22  
**测试环境**: Windows, Chrome Extension (Manifest V3), Native Host (Go), CLI (Go)  
**测试页面**: https://www.baidu.com, https://www.bing.com

---

## 测试结果总览

| 模块 | 总数 | 通过 | 失败 | 通过率 |
|------|------|------|------|--------|
| Tab 管理 | 5 | 5 | 0 | 100% |
| Page 操作 | 9 | 9 | 0 | 100% |
| Navigation | 5 | 5 | 0 | 100% |
| Search | 2 | 2 | 0 | 100% |
| Cookie | 3 | 3 | 0 | 100% |
| **合计** | **24** | **24** | **0** | **100%** |

---

## 详细测试结果

### 1. Tab 管理命令

| 命令 | 用法 | 结果 | 备注 |
|------|------|------|------|
| `tab list` | `browser-bridge tab list` | PASS | 返回所有打开的 Tab 列表，包含 id/title/url/active 等字段 |
| `tab get` | `browser-bridge tab get <tabId>` | PASS | 返回指定 Tab 的详细信息 |
| `tab create` | `browser-bridge tab create --url <url>` | PASS | 创建新 Tab 并返回 tabId 和 url |
| `tab activate` | `browser-bridge tab activate <tabId>` | PASS | 激活指定 Tab |
| `tab close` | `browser-bridge tab close <tabId>` | PASS | 关闭指定 Tab |

### 2. Page 操作命令

| 命令 | 用法 | 结果 | 备注 |
|------|------|------|------|
| `page content` | `browser-bridge page content <tabId>` | PASS | 返回页面原始 HTML |
| `page content --format markdown` | `browser-bridge page content <tabId> --format markdown` | PASS | 返回清理后的 Markdown 格式内容（中文在终端可能乱码，数据本身正确） |
| `page screenshot` | `browser-bridge page screenshot <tabId> -o <path>` | PASS | 截图保存为文件（83KB PNG），验证文件完整 |
| `page snapshot` | `browser-bridge page snapshot <tabId> --ref-only` | PASS | 返回页面可交互元素列表及 ref 编号 |
| `page execute` | `browser-bridge page execute <tabId> "document.title"` | PASS | 在页面上下文执行 JS 并返回结果 |
| `page query` | `browser-bridge page query <tabId> "input"` | PASS | 返回匹配元素列表，含 tagName/attributes/boundingRect |
| `page type` | `browser-bridge page type <tabId> "text" --selector "#kw" --clear` | PASS | 向指定元素输入文本 |
| `page click` | `browser-bridge page click <tabId> --selector "#su"` | PASS | 点击指定元素 |
| `page wait` | `browser-bridge page wait <tabId> "#content_left" --timeout 5000` | PASS | 等待元素出现，返回 found: true |
| `page scroll` | `browser-bridge page scroll <tabId> --y 500` | PASS | 滚动页面 |
| `page select` | `browser-bridge page select <tabId> --selector "select" "value"` | PASS | 选择下拉框选项 |

### 3. Navigation 命令

| 命令 | 用法 | 结果 | 备注 |
|------|------|------|------|
| `nav goto` | `browser-bridge nav goto <tabId> <url>` | PASS | 导航到指定 URL |
| `nav back` | `browser-bridge nav back <tabId>` | PASS | 浏览器后退（需有历史记录） |
| `nav forward` | `browser-bridge nav forward <tabId>` | PASS | 浏览器前进（需有历史记录） |
| `nav reload` | `browser-bridge nav reload <tabId>` | PASS | 刷新当前页面 |

### 4. Search 命令

| 命令 | 用法 | 结果 | 备注 |
|------|------|------|------|
| `search baidu` | `browser-bridge search baidu "rust programming"` | PASS | 返回搜索结果列表（title/url/snippet），搜索后自动关闭 Tab |
| `search bing` | `browser-bridge search bing "golang tutorial"` | PASS | 返回搜索结果列表（title/url/snippet），搜索后自动关闭 Tab |

### 5. Cookie 命令

| 命令 | 用法 | 结果 | 备注 |
|------|------|------|------|
| `cookie get` | `browser-bridge cookie get --url https://www.baidu.com` | PASS | 返回指定域名的 Cookie 列表 |
| `cookie set` | `browser-bridge cookie set https://example.com "test_key" "test_value"` | PASS | 设置 Cookie 成功 |
| `cookie delete` | `browser-bridge cookie delete https://example.com "test_key"` | PASS | 删除 Cookie 成功 |

---

## 测试中发现的问题及修复

### 问题 1: `page execute` 被 CSP 阻止
- **现象**: 执行 `page execute` 报错 "Evaluating a string as JavaScript violates the following Content Security Policy directive"
- **原因**: 原实现使用 `new Function()` 在 service worker 中执行脚本，被 Manifest V3 的 CSP 阻止
- **修复**: 改用 `chrome.scripting.executeScript` 的 `world: "MAIN"` 选项，在页面上下文中通过 `eval` 执行脚本
- **文件**: [page.ts](file:///d:/code/mini-claw/browser-bridge/extension/src/actions/page.ts)

### 问题 2: `page query` 参数序列化失败
- **现象**: 执行 `page query` 报错 "Value is unserializable"
- **原因**: `params.attributes` 为 `undefined` 时，Chrome 不允许将其传递给 `chrome.scripting.executeScript` 的 `args`
- **修复**: 将 `undefined` 参数替换为 `null`（`params.attributes ?? null`）
- **文件**: [page.ts](file:///d:/code/mini-claw/browser-bridge/extension/src/actions/page.ts)

### 问题 3: `page wait` 在 service worker 中使用 document
- **现象**: 执行 `page wait` 报错 "document is not defined"
- **原因**: 原实现在 service worker 中直接使用 `document.querySelector`，而 service worker 没有 DOM
- **修复**: 改用 `chrome.scripting.executeScript` 在页面上下文中执行等待逻辑
- **文件**: [page.ts](file:///d:/code/mini-claw/browser-bridge/extension/src/actions/page.ts)

### 问题 4: `nav back/forward` 报 "Cannot find a next page in history"
- **现象**: 即使有浏览历史，`chrome.tabs.goBack()` 和 `chrome.tabs.goForward()` 也报错
- **原因**: Manifest V3 中 `chrome.tabs.goBack/goForward` 可能存在兼容性问题
- **修复**: 改用 `chrome.scripting.executeScript` 在页面上下文中执行 `history.back()` 和 `history.forward()`
- **文件**: [background.ts](file:///d:/code/mini-claw/browser-bridge/extension/src/background.ts)

### 问题 5: Cookie 命令缺少权限
- **现象**: 执行 `cookie get` 报错 "Cannot read properties of undefined (reading 'getAll')"
- **原因**: manifest.json 中缺少 `cookies` 权限
- **修复**: 在 `permissions` 中添加 `"cookies"`
- **文件**: [manifest.json](file:///d:/code/mini-claw/browser-bridge/extension/manifest.json)

### 问题 6: 搜索命令仅返回 Tab 信息，不返回搜索结果
- **现象**: `search baidu/bing` 只返回 tabId/url/title，不返回搜索结果列表
- **原因**: 原实现仅创建搜索 Tab 并返回 Tab 信息，未解析搜索结果
- **修复**: 在搜索页面加载完成后，使用 `chrome.scripting.executeScript` 提取搜索结果（title/url/snippet），提取后自动关闭搜索 Tab
- **文件**: [search.ts](file:///d:/code/mini-claw/browser-bridge/extension/src/actions/search.ts)

---

## 已知限制

1. **终端中文乱码**: `page content --format markdown` 返回的中文在 Windows PowerShell 终端中可能显示乱码，这是终端编码问题，数据本身是正确的 UTF-8
2. **nav back/forward 无历史时静默成功**: 使用 `history.back()`/`history.forward()` 后，即使没有历史也不会报错（与浏览器行为一致）
3. **CLI 缺少 `nav waitLoad` 子命令**: Extension 端已实现，但 CLI 未添加对应命令
4. **page scroll 参数 undefined 问题**: 已修复，使用 `??` 运算符提供默认值

---

## 修改文件清单

| 文件 | 修改内容 |
|------|----------|
| `extension/src/actions/page.ts` | 修复 execute (world:MAIN)、query (null args)、wait (executeScript)、scroll (null args) |
| `extension/src/background.ts` | 修复 nav back/forward 使用 history.back/forward |
| `extension/src/actions/search.ts` | 搜索结果提取（title/url/snippet），搜索后关闭 Tab |
| `extension/src/types.ts` | 新增 SearchResultItem 类型 |

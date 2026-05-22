# Browser Bridge - 项目原型设计

## 1. 项目概述

Browser Bridge 是一个提供给 Agent 进行浏览器自动化的工具，采用三层架构：

```
┌──────────────────┐     HTTP      ┌──────────────┐   Native Msg   ┌──────────────────┐
│ browser-bridge-cli│ ──────────▶  │  native-host  │ ────────────▶  │ Chrome Extension  │
│   (Go CLI工具)    │ ◀────────── │  (Go HTTP服务) │ ◀──────────── │  (浏览器插件)      │
└──────────────────┘     HTTP      └──────────────┘   Native Msg   └──────────────────┘
```

**通信链路**：

- CLI → Native Host：HTTP REST API
- Native Host → Chrome Extension：Chrome Native Messaging（stdin/stdout 二进制协议）
- Chrome Extension → Native Host：Chrome Native Messaging（stdin/stdout 二进制协议）
- Native Host → CLI：HTTP Response

---

## 2. 整体目录结构

```
browser-bridge/
├── .plan/                    # 项目设计文档
├── extension/                # Chrome 浏览器插件（TypeScript）
│   ├── src/                  # TypeScript 源码
│   │   ├── background.ts     # Service Worker，处理 Native Messaging 通信
│   │   ├── content.ts        # Content Script，注入页面执行 JS
│   │   ├── popup.ts          # 弹窗逻辑
│   │   ├── types.ts          # 类型定义（消息协议、Tab、Page 等）
│   │   ├── actions/          # Action 处理器
│   │   │   ├── tab.ts        # Tab 相关操作
│   │   │   ├── page.ts       # 页面内容/脚本执行/截图
│   │   │   ├── snapshot.ts   # 页面快照与 ref 管理
│   │   │   └── search.ts     # 搜索功能
│   ├── dist/                 # 编译输出（加载到 Chrome 的目录）
│   ├── manifest.json         # 插件配置（含 nativeMessaging 权限）
│   ├── popup.html            # 插件弹窗 UI（可选，调试用）
│   ├── icons/                # 插件图标
│   ├── tsconfig.json         # TypeScript 配置
│   └── package.json          # 依赖管理
├── native-host/              # Native Host 程序（Go）
│   ├── main.go               # 入口，启动 HTTP 服务
│   ├── handler/              # HTTP 请求处理器
│   │   ├── tab.go            # Tab 相关接口
│   │   ├── page.go           # 页面内容/执行脚本接口
│   │   └── search.go         # 搜索相关接口
│   ├── native/               # Native Messaging 通信层
│   │   ├── protocol.go       # 消息编解码（4字节长度前缀）
│   │   └── bridge.go         # 与 Extension 的双向消息转发
│   ├── server/               # HTTP 服务配置
│   │   └── server.go
│   └── model/                # 数据模型
│       └── types.go
├── cli/                      # browser-bridge-cli 程序（Go）
│   ├── main.go               # CLI 入口
│   ├── cmd/                  # cobra 命令定义
│   │   ├── root.go
│   │   ├── tab.go            # tab list 命令
│   │   ├── page.go           # page content / execute 命令
│   │   └── search.go         # search 命令
│   └── client/               # HTTP 客户端
│       └── client.go
└── install/                  # 安装脚本与配置
    ├── install.bat           # Windows 安装脚本
    ├── com.browser.bridge.json  # Native Messaging Host 注册配置
    └── registry.go           # 注册表写入逻辑（可选）
```

---

## 3. Chrome 浏览器插件设计（TypeScript）

### 3.1 技术栈

| 工具          | 用途                          |
| ------------- | ----------------------------- |
| TypeScript    | 主开发语言                    |
| @chrome/types | Chrome Extension API 类型定义 |
| esbuild       | 构建打包（TS → JS，极速构建） |

### 3.2 manifest.json

```json
{
  "manifest_version": 3,
  "name": "Browser Bridge",
  "version": "0.1.0",
  "description": "Browser automation bridge for AI agents",
  "permissions": ["nativeMessaging", "tabs", "activeTab", "scripting"],
  "background": {
    "service_worker": "dist/background.js"
  },
  "content_scripts": [],
  "icons": {
    "16": "icons/icon16.png",
    "48": "icons/icon48.png",
    "128": "icons/icon128.png"
  }
}
```

### 3.3 类型定义（src/types.ts）

```typescript
// 消息协议类型
export interface RequestMessage {
  id: string;
  action: ActionType;
  params: Record<string, unknown>;
}

export interface ResponseMessage {
  id: string;
  success: boolean;
  data?: unknown;
  error?: string;
}

export type ActionType =
  // Tab 管理
  | "tab.list"
  | "tab.get"
  | "tab.create"
  | "tab.close"
  | "tab.activate"
  // 页面操作
  | "page.snapshot"
  | "page.content"
  | "page.screenshot"
  | "page.execute"
  | "page.click"
  | "page.type"
  | "page.select"
  | "page.scroll"
  | "page.query"
  | "page.wait"
  // 导航
  | "nav.goto"
  | "nav.back"
  | "nav.forward"
  | "nav.reload"
  | "nav.waitLoad"
  // Cookie
  | "cookie.get"
  | "cookie.set"
  | "cookie.delete"
  // 搜索
  | "search.baidu"
  | "search.bing";

// Action 参数类型
export interface TabGetParams {
  tabId: number;
}

export interface TabCreateParams {
  url?: string;
  active?: boolean;
}

export interface TabCloseParams {
  tabId: number;
}

export interface TabActivateParams {
  tabId: number;
}

export interface PageContentParams {
  tabId: number;
}

export interface PageSnapshotParams {
  tabId: number;
  refOnly?: boolean; // 仅返回带 ref 的元素，精简输出
}

export interface PageScreenshotParams {
  tabId: number;
  format?: "png" | "jpeg";
  quality?: number;
  fullPage?: boolean;
  ref?: string; // 通过 ref 截取指定元素
  selector?: string;
}

export interface PageExecuteParams {
  tabId: number;
  script: string;
}

export interface PageClickParams {
  tabId: number;
  ref?: string; // 优先使用 ref
  selector?: string;
}

export interface PageTypeParams {
  tabId: number;
  ref?: string;
  selector?: string;
  text: string;
  clear?: boolean; // 输入前清空已有内容
}

export interface PageSelectParams {
  tabId: number;
  ref?: string;
  selector?: string;
  value: string;
}

export interface PageScrollParams {
  tabId: number;
  x?: number;
  y?: number;
  ref?: string;
  selector?: string;
}

export interface PageQueryParams {
  tabId: number;
  selector: string;
  attributes?: string[];
}

export interface PageWaitParams {
  tabId: number;
  selector: string;
  timeout?: number;
}

export interface NavGotoParams {
  tabId: number;
  url: string;
}

export interface NavWaitLoadParams {
  tabId: number;
  timeout?: number;
}

export interface CookieGetParams {
  url?: string;
  name?: string;
  domain?: string;
}

export interface CookieSetParams {
  url: string;
  name: string;
  value: string;
  domain?: string;
  path?: string;
  secure?: boolean;
  httpOnly?: boolean;
  expirationDate?: number;
}

export interface CookieDeleteParams {
  url: string;
  name: string;
}

export interface SearchParams {
  query: string;
}

// Tab 信息类型
export interface TabInfo {
  id: number;
  url: string;
  title: string;
  favIconUrl?: string;
  active: boolean;
  windowId: number;
}

// 搜索结果类型
export interface SearchResult {
  tabId: number;
  url: string;
  title: string;
}

// 截图结果类型
export interface ScreenshotResult {
  tabId: number;
  format: string;
  data: string; // base64 编码
}

// 页面快照结果类型
export interface SnapshotResult {
  tabId: number;
  url: string;
  title: string;
  snapshot: SnapshotNode[];
}

// 快照节点 - 可访问性树结构，每个可交互元素分配唯一 ref
export interface SnapshotNode {
  ref?: string; // 可交互元素的唯一引用 ID（如 "r1", "r2"），Agent 通过 ref 操作元素
  tagName: string; // 标签名
  role?: string; // ARIA role（如 "button", "link", "textbox"）
  name?: string; // 可访问性名称（按钮文字、链接文字等）
  attributes?: Record<string, string>; // 关键属性（href, src, placeholder 等）
  children?: SnapshotNode[];
  textContent?: string; // 文本内容（叶子节点）
}

// DOM 元素查询结果
export interface ElementInfo {
  tagName: string;
  textContent: string | null;
  attributes: Record<string, string>;
  innerHTML: string;
  boundingRect: {
    x: number;
    y: number;
    width: number;
    height: number;
  };
}
```

### 3.4 background.ts - 核心消息处理

**职责**：

1. 建立 Native Messaging 连接
2. 监听来自 Native Host 的消息，解析指令并执行
3. 将执行结果回传给 Native Host

**消息协议（Extension ↔ Native Host）**：

```json
{
  "id": "unique-request-id",
  "action": "action-name",
  "params": { ... }
}
```

**响应格式**：

```json
{
  "id": "unique-request-id",
  "success": true,
  "data": { ... },
  "error": ""
}
```

**支持的 action**：

### Tab 管理

| Action         | 说明               | 参数                                 |
| -------------- | ------------------ | ------------------------------------ |
| `tab.list`     | 获取所有打开的 Tab | 无                                   |
| `tab.get`      | 获取指定 Tab 信息  | `{ tabId: number }`                  |
| `tab.create`   | 创建新 Tab         | `{ url?: string, active?: boolean }` |
| `tab.close`    | 关闭 Tab           | `{ tabId: number }`                  |
| `tab.activate` | 激活 Tab           | `{ tabId: number }`                  |

### 页面操作

| Action            | 说明                   | 参数                                                                                |
| ----------------- | ---------------------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| `page.snapshot`   | 获取页面快照（含 ref） | `{ tabId: number, refOnly?: boolean }`                                              |
| `page.content`    | 获取页面 HTML 内容     | `{ tabId: number }`                                                                 |
| `page.screenshot` | 页面截图               | `{ tabId: number, format?: "png"                                                    | "jpeg", quality?: number, fullPage?: boolean, ref?: string }` |
| `page.execute`    | 在页面执行 JS 脚本     | `{ tabId: number, script: string }`                                                 |
| `page.click`      | 点击元素（支持 ref）   | `{ tabId: number, ref?: string, selector?: string }`                                |
| `page.type`       | 输入文本（支持 ref）   | `{ tabId: number, ref?: string, selector?: string, text: string, clear?: boolean }` |
| `page.select`     | 选择下拉框（支持 ref） | `{ tabId: number, ref?: string, selector?: string, value: string }`                 |
| `page.scroll`     | 滚动页面               | `{ tabId: number, x?: number, y?: number, ref?: string, selector?: string }`        |
| `page.query`      | 查询 DOM 元素信息      | `{ tabId: number, selector: string, attributes?: string[] }`                        |
| `page.wait`       | 等待元素出现           | `{ tabId: number, selector: string, timeout?: number }`                             |

### 导航

| Action         | 说明             | 参数                                  |
| -------------- | ---------------- | ------------------------------------- |
| `nav.goto`     | 导航到 URL       | `{ tabId: number, url: string }`      |
| `nav.back`     | 后退             | `{ tabId: number }`                   |
| `nav.forward`  | 前进             | `{ tabId: number }`                   |
| `nav.reload`   | 刷新页面         | `{ tabId: number }`                   |
| `nav.waitLoad` | 等待页面加载完成 | `{ tabId: number, timeout?: number }` |

### Cookie 与存储

| Action          | 说明        | 参数                                                |
| --------------- | ----------- | --------------------------------------------------- |
| `cookie.get`    | 获取 Cookie | `{ url?: string, name?: string, domain?: string }`  |
| `cookie.set`    | 设置 Cookie | `{ url: string, name: string, value: string, ... }` |
| `cookie.delete` | 删除 Cookie | `{ url: string, name: string }`                     |

### 搜索

| Action         | 说明           | 参数                |
| -------------- | -------------- | ------------------- |
| `search.baidu` | 通过百度搜索   | `{ query: string }` |
| `search.bing`  | 通过 Bing 搜索 | `{ query: string }` |

**核心逻辑**：

```typescript
// src/background.ts
import { RequestMessage, ResponseMessage, ActionType } from "./types";
import { handleTabList, handleTabGet } from "./actions/tab";
import {
  handlePageContent,
  handlePageExecute,
  handlePageScreenshot,
  handlePageClick,
  handlePageType,
  handlePageSelect,
  handlePageScroll,
} from "./actions/page";
import { handlePageSnapshot, resolveRef } from "./actions/snapshot";
import { handleSearch } from "./actions/search";

const port = chrome.runtime.connectNative("com.browser.bridge");

port.onMessage.addListener(async (msg: RequestMessage) => {
  const { id, action, params } = msg;
  try {
    let result: unknown;
    switch (action as ActionType) {
      case "tab.list":
        result = await handleTabList();
        break;
      case "tab.get":
        result = await handleTabGet(params as { tabId: number });
        break;
      case "page.snapshot":
        result = await handlePageSnapshot(
          params as { tabId: number; refOnly?: boolean },
        );
        break;
      case "page.content":
        result = await handlePageContent(params as { tabId: number });
        break;
      case "page.screenshot":
        result = await handlePageScreenshot(
          params as {
            tabId: number;
            format?: string;
            quality?: number;
            fullPage?: boolean;
            ref?: string;
            selector?: string;
          },
        );
        break;
      case "page.execute":
        result = await handlePageExecute(
          params as { tabId: number; script: string },
        );
        break;
      case "page.click": {
        const clickParams = params as {
          tabId: number;
          ref?: string;
          selector?: string;
        };
        const selector = clickParams.ref
          ? await resolveRef(clickParams.tabId, clickParams.ref)
          : clickParams.selector!;
        result = await handlePageClick({ tabId: clickParams.tabId, selector });
        break;
      }
      case "page.type": {
        const typeParams = params as {
          tabId: number;
          ref?: string;
          selector?: string;
          text: string;
          clear?: boolean;
        };
        const selector = typeParams.ref
          ? await resolveRef(typeParams.tabId, typeParams.ref)
          : typeParams.selector!;
        result = await handlePageType({
          tabId: typeParams.tabId,
          selector,
          text: typeParams.text,
          clear: typeParams.clear,
        });
        break;
      }
      case "page.select": {
        const selectParams = params as {
          tabId: number;
          ref?: string;
          selector?: string;
          value: string;
        };
        const selector = selectParams.ref
          ? await resolveRef(selectParams.tabId, selectParams.ref)
          : selectParams.selector!;
        result = await handlePageSelect({
          tabId: selectParams.tabId,
          selector,
          value: selectParams.value,
        });
        break;
      }
      case "page.scroll": {
        const scrollParams = params as {
          tabId: number;
          ref?: string;
          selector?: string;
          x?: number;
          y?: number;
        };
        const selector = scrollParams.ref
          ? await resolveRef(scrollParams.tabId, scrollParams.ref)
          : scrollParams.selector;
        result = await handlePageScroll({
          tabId: scrollParams.tabId,
          selector,
          x: scrollParams.x,
          y: scrollParams.y,
        });
        break;
      }
      case "search.baidu":
        result = await handleSearch(
          "https://www.baidu.com/s?wd=",
          params as { query: string },
        );
        break;
      case "search.bing":
        result = await handleSearch(
          "https://www.bing.com/search?q=",
          params as { query: string },
        );
        break;
      default:
        throw new Error(`Unknown action: ${action}`);
    }
    const response: ResponseMessage = { id, success: true, data: result };
    port.postMessage(response);
  } catch (e) {
    const response: ResponseMessage = {
      id,
      success: false,
      error: (e as Error).message,
    };
    port.postMessage(response);
  }
});

port.onDisconnect.addListener(() => {
  console.error("Native host disconnected:", chrome.runtime.lastError);
});
```

### 3.5 Action 处理器

**src/actions/tab.ts**：

```typescript
import { TabInfo } from "../types";

export async function handleTabList(): Promise<TabInfo[]> {
  const tabs = await chrome.tabs.query({});
  return tabs.map((tab) => ({
    id: tab.id!,
    url: tab.url ?? "",
    title: tab.title ?? "",
    favIconUrl: tab.favIconUrl,
    active: tab.active,
    windowId: tab.windowId,
  }));
}

export async function handleTabGet(params: {
  tabId: number;
}): Promise<TabInfo> {
  const tab = await chrome.tabs.get(params.tabId);
  return {
    id: tab.id!,
    url: tab.url ?? "",
    title: tab.title ?? "",
    favIconUrl: tab.favIconUrl,
    active: tab.active,
    windowId: tab.windowId,
  };
}
```

**src/actions/snapshot.ts**：

```typescript
import { SnapshotResult, SnapshotNode } from "../types";

// ref → CSS selector 映射表（按 tabId 隔离）
const refMaps = new Map<number, Map<string, string>>();

// 可交互元素标签
const INTERACTIVE_TAGS = new Set([
  "a",
  "button",
  "input",
  "select",
  "textarea",
  "summary",
  "details",
  "option",
  "optgroup",
]);

// 可交互 ARIA roles
const INTERACTIVE_ROLES = new Set([
  "button",
  "link",
  "textbox",
  "checkbox",
  "radio",
  "combobox",
  "listbox",
  "menuitem",
  "tab",
  "slider",
  "spinbutton",
  "switch",
  "searchbox",
]);

// 获取页面快照
export async function handlePageSnapshot(params: {
  tabId: number;
  refOnly?: boolean;
}): Promise<SnapshotResult> {
  const tab = await chrome.tabs.get(params.tabId);

  // 在页面中执行快照采集脚本
  const results = await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: collectAccessibilityTree,
  });

  const rawTree = results[0].result as SnapshotNode[];
  const refMap = new Map<string, string>();
  let refCounter = 0;

  // 为可交互元素分配 ref 并建立映射
  const processedTree = assignRefs(rawTree, refMap, () => `r${++refCounter}`);

  // 更新 ref 映射表
  refMaps.set(params.tabId, refMap);

  // refOnly 模式：只保留带 ref 的节点
  const snapshot = params.refOnly
    ? filterRefOnly(processedTree)
    : processedTree;

  return {
    tabId: params.tabId,
    url: tab.url ?? "",
    title: tab.title ?? "",
    snapshot,
  };
}

// 通过 ref 解析为 CSS selector
export async function resolveRef(tabId: number, ref: string): Promise<string> {
  const map = refMaps.get(tabId);
  if (!map || !map.has(ref)) {
    throw new Error(
      `ref "${ref}" not found, please call page.snapshot first to refresh refs`,
    );
  }
  return map.get(ref)!;
}

// 在页面中采集可访问性树（注入到页面执行）
function collectAccessibilityTree(): SnapshotNode[] {
  function walk(el: Element): SnapshotNode {
    const node: SnapshotNode = {
      tagName: el.tagName.toLowerCase(),
    };

    // 获取 ARIA role
    const role = el.getAttribute("role") || getImplicitRole(el);
    if (role) node.role = role;

    // 获取可访问性名称
    const name =
      el.getAttribute("aria-label") ||
      el.getAttribute("alt") ||
      el.getAttribute("title") ||
      el.getAttribute("placeholder") ||
      el.textContent?.trim().slice(0, 100) ||
      undefined;
    if (name) node.name = name;

    // 收集关键属性
    const attrs: Record<string, string> = {};
    for (const attr of [
      "href",
      "src",
      "type",
      "value",
      "placeholder",
      "disabled",
      "checked",
    ]) {
      if (el.hasAttribute(attr)) attrs[attr] = el.getAttribute(attr)!;
    }
    if (el.id) attrs.id = el.id;
    if (Object.keys(attrs).length > 0) node.attributes = attrs;

    // 生成唯一 CSS selector 用于 ref 映射
    const selector = generateSelector(el);

    // 递归子节点
    const children: SnapshotNode[] = [];
    for (const child of el.children) {
      // 跳过 script, style, noscript 等不可见元素
      const tag = child.tagName.toLowerCase();
      if (["script", "style", "noscript", "svg", "path"].includes(tag))
        continue;
      children.push(walk(child));
    }
    if (children.length > 0) node.children = children;

    // 叶子文本节点
    if (children.length === 0 && el.childNodes.length > 0) {
      const text = el.textContent?.trim();
      if (text) node.textContent = text.slice(0, 200);
    }

    // 存储 selector 供 ref 分配使用（临时字段）
    (node as any)._selector = selector;

    return node;
  }

  function generateSelector(el: Element): string {
    if (el.id) return `#${CSS.escape(el.id)}`;
    const parent = el.parentElement;
    if (!parent) return el.tagName.toLowerCase();
    const siblings = Array.from(parent.children).filter(
      (c) => c.tagName === el.tagName,
    );
    if (siblings.length === 1)
      return `${generateSelector(parent)} > ${el.tagName.toLowerCase()}`;
    const index = siblings.indexOf(el) + 1;
    return `${generateSelector(parent)} > ${el.tagName.toLowerCase()}:nth-of-type(${index})`;
  }

  function getImplicitRole(el: Element): string | undefined {
    const tag = el.tagName.toLowerCase();
    const roleMap: Record<string, string> = {
      a: "link",
      button: "button",
      input: "textbox",
      select: "combobox",
      textarea: "textbox",
      summary: "button",
      details: "group",
    };
    return roleMap[tag];
  }

  return Array.from(document.body.children).map(walk);
}

// 为可交互元素分配 ref
function assignRefs(
  nodes: SnapshotNode[],
  refMap: Map<string, string>,
  nextRef: () => string,
): SnapshotNode[] {
  return nodes.map((node) => {
    const isInteractive =
      INTERACTIVE_TAGS.has(node.tagName) ||
      (node.role && INTERACTIVE_ROLES.has(node.role));

    const newNode: SnapshotNode = { ...node };
    if (isInteractive) {
      const ref = nextRef();
      newNode.ref = ref;
      refMap.set(ref, (node as any)._selector);
    }
    // 清理临时字段
    delete (newNode as any)._selector;

    if (node.children) {
      newNode.children = assignRefs(node.children, refMap, nextRef);
    }
    return newNode;
  });
}

// 仅保留带 ref 的节点（精简模式）
function filterRefOnly(nodes: SnapshotNode[]): SnapshotNode[] {
  const result: SnapshotNode[] = [];
  for (const node of nodes) {
    if (node.ref) {
      const filtered: SnapshotNode = { ...node };
      if (node.children) {
        filtered.children = filterRefOnly(node.children);
        if (filtered.children!.length === 0) delete filtered.children;
      }
      result.push(filtered);
    } else if (node.children) {
      result.push(...filterRefOnly(node.children));
    }
  }
  return result;
}
```

**src/actions/page.ts**：

```typescript
// 获取页面内容
export async function handlePageContent(params: {
  tabId: number;
}): Promise<{ content: string; tabId: number; url: string }> {
  const tab = await chrome.tabs.get(params.tabId);
  const results = await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: () => document.documentElement.outerHTML,
  });
  return {
    content: results[0].result as string,
    tabId: params.tabId,
    url: tab.url ?? "",
  };
}

// 页面截图
export async function handlePageScreenshot(params: {
  tabId: number;
  format?: string;
  quality?: number;
  fullPage?: boolean;
  ref?: string;
  selector?: string;
}): Promise<{ tabId: number; format: string; data: string }> {
  const format = (params.format as "png" | "jpeg") || "png";

  // 如果指定了 ref/selector，先滚动到元素再截图
  if (params.selector) {
    await chrome.scripting.executeScript({
      target: { tabId: params.tabId },
      func: (sel: string) => {
        const el = document.querySelector(sel);
        el?.scrollIntoView({ block: "center" });
      },
      args: [params.selector],
    });
  }

  const dataUri = await chrome.tabs.captureVisibleTab(
    (await chrome.tabs.get(params.tabId)).windowId,
    { format, quality: params.quality },
  );

  // data:image/png;base64,... → 去掉前缀只保留 base64
  const base64 = dataUri.split(",")[1];

  return { tabId: params.tabId, format, data: base64 };
}

// 执行脚本
export async function handlePageExecute(params: {
  tabId: number;
  script: string;
}): Promise<{ result: unknown; tabId: number }> {
  const results = await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: new Function(params.script) as () => unknown,
  });
  return {
    result: results[0].result,
    tabId: params.tabId,
  };
}

// 点击元素
export async function handlePageClick(params: {
  tabId: number;
  selector: string;
}): Promise<{ tabId: number; selector: string }> {
  await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: (sel: string) => {
      const el = document.querySelector(sel) as HTMLElement;
      if (!el) throw new Error(`Element not found: ${sel}`);
      el.click();
    },
    args: [params.selector],
  });
  return { tabId: params.tabId, selector: params.selector };
}

// 输入文本
export async function handlePageType(params: {
  tabId: number;
  selector: string;
  text: string;
  clear?: boolean;
}): Promise<{ tabId: number; selector: string; text: string }> {
  await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: (sel: string, text: string, clear: boolean) => {
      const el = document.querySelector(sel) as HTMLInputElement;
      if (!el) throw new Error(`Element not found: ${sel}`);
      if (clear) {
        el.value = "";
        el.focus();
      }
      el.value += text;
      el.dispatchEvent(new Event("input", { bubbles: true }));
      el.dispatchEvent(new Event("change", { bubbles: true }));
    },
    args: [params.selector, params.text, params.clear ?? false],
  });
  return { tabId: params.tabId, selector: params.selector, text: params.text };
}

// 选择下拉框
export async function handlePageSelect(params: {
  tabId: number;
  selector: string;
  value: string;
}): Promise<{ tabId: number; selector: string; value: string }> {
  await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: (sel: string, value: string) => {
      const el = document.querySelector(sel) as HTMLSelectElement;
      if (!el) throw new Error(`Element not found: ${sel}`);
      el.value = value;
      el.dispatchEvent(new Event("change", { bubbles: true }));
    },
    args: [params.selector, params.value],
  });
  return {
    tabId: params.tabId,
    selector: params.selector,
    value: params.value,
  };
}

// 滚动页面
export async function handlePageScroll(params: {
  tabId: number;
  selector?: string;
  x?: number;
  y?: number;
}): Promise<{ tabId: number }> {
  await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: (
      sel: string | undefined,
      x: number | undefined,
      y: number | undefined,
    ) => {
      if (sel) {
        document.querySelector(sel)?.scrollIntoView({ block: "center" });
      } else {
        window.scrollBy(x ?? 0, y ?? 0);
      }
    },
    args: [params.selector, params.x, params.y],
  });
  return { tabId: params.tabId };
}
```

**src/actions/search.ts**：

```typescript
import { SearchResult } from "../types";

export async function handleSearch(
  baseUrl: string,
  params: { query: string },
): Promise<SearchResult> {
  const tab = await chrome.tabs.create({
    url: baseUrl + encodeURIComponent(params.query),
  });
  // 等待页面加载完成
  await new Promise<void>((resolve) => {
    const listener = (tabId: number, info: chrome.tabs.TabChangeInfo) => {
      if (tabId === tab.id && info.status === "complete") {
        chrome.tabs.onUpdated.removeListener(listener);
        resolve();
      }
    };
    chrome.tabs.onUpdated.addListener(listener);
  });
  // 刷新 tab 信息获取最新 title
  const updatedTab = await chrome.tabs.get(tab.id!);
  return {
    tabId: tab.id!,
    url: updatedTab.url ?? "",
    title: updatedTab.title ?? "",
  };
}
```

### 3.6 构建配置

**tsconfig.json**：

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ES2020",
    "lib": ["ES2020", "DOM"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "moduleResolution": "node",
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

**package.json**：

```json
{
  "name": "browser-bridge-extension",
  "version": "0.1.0",
  "scripts": {
    "build": "esbuild src/background.ts --bundle --outfile=dist/background.js --target=chrome100",
    "dev": "esbuild src/background.ts --bundle --outfile=dist/background.js --target=chrome100 --watch --sourcemap"
  },
  "devDependencies": {
    "typescript": "^5.0.0",
    "@chrome/types": "^0.1.0",
    "esbuild": "^0.20.0"
  }
}
```

---

## 4. Native Host 程序设计

### 4.1 架构

Native Host 同时承担两个角色：

1. **Chrome Native Messaging 端**：通过 stdin/stdout 与 Extension 双向通信
2. **HTTP 服务端**：接收 CLI 的请求并转发给 Extension

```
                    ┌─────────────────────────────────┐
  CLI ──HTTP──▶    │         Native Host              │
                    │                                  │
                    │  HTTP Server ──▶ Message Router  │──stdout──▶ Extension
                    │                                  │◀──stdin─── Extension
                    │  Response Map (request-id based) │
                    └─────────────────────────────────┘
```

### 4.2 HTTP API 设计

**基础路径**: `http://127.0.0.1:{port}/api/v1`

#### 4.2.1 获取 Tab 列表

```
GET /api/v1/tabs
```

**响应**：

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "url": "https://www.baidu.com",
      "title": "百度一下",
      "favIconUrl": "https://www.baidu.com/favicon.ico",
      "active": true,
      "windowId": 1
    }
  ]
}
```

#### 4.2.2 获取指定 Tab

```
GET /api/v1/tabs/:tabId
```

#### 4.2.3 创建 Tab

```
POST /api/v1/tabs
Content-Type: application/json

{ "url": "https://baidu.com", "active": true }
```

#### 4.2.4 关闭 Tab

```
DELETE /api/v1/tabs/:tabId
```

#### 4.2.5 激活 Tab

```
PUT /api/v1/tabs/:tabId/activate
```

#### 4.2.6 获取页面快照

```
GET /api/v1/tabs/:tabId/snapshot?refOnly=false
```

**响应**：

```json
{
  "success": true,
  "data": {
    "tabId": 1,
    "url": "https://www.baidu.com",
    "title": "百度一下",
    "snapshot": [
      {
        "tagName": "div",
        "role": "main",
        "children": [
          {
            "ref": "r1",
            "tagName": "input",
            "role": "textbox",
            "name": "搜索",
            "attributes": { "placeholder": "输入关键词" }
          },
          {
            "ref": "r2",
            "tagName": "button",
            "role": "button",
            "name": "百度一下"
          },
          {
            "ref": "r3",
            "tagName": "a",
            "role": "link",
            "name": "新闻",
            "attributes": { "href": "http://news.baidu.com" }
          }
        ]
      }
    ]
  }
}
```

**说明**：

- `ref` 仅分配给可交互元素（按钮、链接、输入框、下拉框等），Agent 通过 ref 操作元素
- `refOnly=true` 时仅返回带 ref 的元素，精简输出，适合 Agent 快速定位可操作元素
- 快照基于可访问性树（Accessibility Tree），比 HTML 更精简，更适合 Agent 理解

#### 4.2.7 获取页面内容

```
GET /api/v1/tabs/:tabId/content
```

**响应**：

```json
{
  "success": true,
  "data": {
    "content": "<html>...</html>",
    "tabId": 1,
    "url": "https://www.baidu.com"
  }
}
```

#### 4.2.8 页面截图

```
POST /api/v1/tabs/:tabId/screenshot
Content-Type: application/json

{
  "format": "png",
  "quality": 90,
  "fullPage": false,
  "ref": "r3",
  "selector": "#logo"
}
```

**说明**：`ref` 和 `selector` 二选一，`ref` 优先

**响应**：

```json
{
  "success": true,
  "data": {
    "tabId": 1,
    "format": "png",
    "data": "iVBORw0KGgo..."
  }
}
```

#### 4.2.9 执行 JS 脚本

```
POST /api/v1/tabs/:tabId/execute
Content-Type: application/json

{ "script": "document.title" }
```

**响应**：

```json
{
  "success": true,
  "data": {
    "result": "百度一下",
    "tabId": 1
  }
}
```

#### 4.2.10 点击元素

```
POST /api/v1/tabs/:tabId/click
Content-Type: application/json

{ "ref": "r2" }
```

**说明**：`ref` 和 `selector` 二选一，`ref` 优先

#### 4.2.11 输入文本

```
POST /api/v1/tabs/:tabId/type
Content-Type: application/json

{ "ref": "r1", "text": "golang", "clear": true }
```

**说明**：`ref` 和 `selector` 二选一，`ref` 优先；`clear=true` 先清空已有内容

#### 4.2.12 选择下拉框

```
POST /api/v1/tabs/:tabId/select
Content-Type: application/json

{ "ref": "r5", "value": "option1" }
```

#### 4.2.13 滚动页面

```
POST /api/v1/tabs/:tabId/scroll
Content-Type: application/json

{ "ref": "r3" }
```

**说明**：传入 `ref` 滚动到指定元素可见

#### 4.2.14 查询 DOM 元素

```
POST /api/v1/tabs/:tabId/query
Content-Type: application/json

{ "selector": ".result", "attributes": ["href", "class"] }
```

**响应**：

```json
{
  "success": true,
  "data": {
    "elements": [
      {
        "tagName": "div",
        "textContent": "搜索结果1",
        "attributes": { "href": "/link1", "class": "result" },
        "innerHTML": "...",
        "boundingRect": { "x": 0, "y": 100, "width": 800, "height": 60 }
      }
    ]
  }
}
```

#### 4.2.15 等待元素出现

```
POST /api/v1/tabs/:tabId/wait
Content-Type: application/json

{ "selector": ".result", "timeout": 5000 }
```

#### 4.2.16 导航

```
POST /api/v1/tabs/:tabId/navigate
Content-Type: application/json

{ "url": "https://github.com" }
```

#### 4.2.17 后退/前进/刷新

```
POST /api/v1/tabs/:tabId/back
POST /api/v1/tabs/:tabId/forward
POST /api/v1/tabs/:tabId/reload
```

#### 4.2.18 等待页面加载

```
POST /api/v1/tabs/:tabId/wait-load
Content-Type: application/json

{ "timeout": 10000 }
```

#### 4.2.19 Cookie 操作

```
GET    /api/v1/cookies?url=https://baidu.com
POST   /api/v1/cookies  { "url": "...", "name": "token", "value": "abc" }
DELETE /api/v1/cookies?url=https://baidu.com&name=token
```

#### 4.2.20 搜索

```
POST /api/v1/search
Content-Type: application/json

{ "engine": "baidu", "query": "golang tutorial" }
```

**响应**：

```json
{
  "success": true,
  "data": {
    "tabId": 42,
    "url": "https://www.baidu.com/s?wd=golang+tutorial",
    "title": "golang tutorial_百度搜索"
  }
}
```

### 4.3 消息路由与请求-响应匹配

Native Host 需要处理 HTTP 请求与 Native Messaging 响应的异步匹配：

```go
// 请求-响应映射
type PendingRequests struct {
    mu       sync.Mutex
    requests map[string]chan *Response // key: request-id
}

// HTTP handler 生成唯一 ID，发送到 Extension，然后等待响应
func (h *Handler) handleRequest(w http.ResponseWriter, r *http.Request) {
    id := uuid.New().String()
    respCh := make(chan *Response, 1)
    pending.Add(id, respCh)
    defer pending.Remove(id)

    // 发送消息到 Extension
    native.Send(id, action, params)

    // 等待响应（带超时）
    select {
    case resp := <-respCh:
        json.NewEncoder(w).Encode(resp)
    case <-time.After(30 * time.Second):
        json.NewEncoder(w).Encode(ErrorResponse("timeout"))
    }
}
```

### 4.4 Native Messaging 协议

Chrome Native Messaging 使用 4 字节 little-endian 长度前缀：

```
[4 bytes: message length (uint32 LE)] [message bytes (JSON)]
```

---

## 5. browser-bridge-cli 设计

### 5.1 命令结构

```
browser-bridge [command]
```

| 命令                                         | 说明               | 示例                                                    |
| -------------------------------------------- | ------------------ | ------------------------------------------------------- |
| `tab list`                                   | 列出所有 Tab       | `browser-bridge tab list`                               |
| `tab get <id>`                               | 获取指定 Tab 信息  | `browser-bridge tab get 1`                              |
| `tab create [url]`                           | 创建新 Tab         | `browser-bridge tab create https://baidu.com`           |
| `tab close <id>`                             | 关闭 Tab           | `browser-bridge tab close 1`                            |
| `tab activate <id>`                          | 激活 Tab           | `browser-bridge tab activate 1`                         |
| `page content <tabId>`                       | 获取页面内容       | `browser-bridge page content 1`                         |
| `page snapshot <tabId>`                      | 获取页面快照       | `browser-bridge page snapshot 1`                        |
| `page snapshot <tabId> --ref-only`           | 仅获取可交互元素   | `browser-bridge page snapshot 1 --ref-only`             |
| `page screenshot <tabId>`                    | 页面截图           | `browser-bridge page screenshot 1`                      |
| `page screenshot <tabId> --full-page`        | 全页截图           | `browser-bridge page screenshot 1 --full-page`          |
| `page screenshot <tabId> --ref`              | 元素截图（by ref） | `browser-bridge page screenshot 1 --ref r3`             |
| `page screenshot <tabId> --selector`         | 元素截图           | `browser-bridge page screenshot 1 --selector "#logo"`   |
| `page execute <tabId> <script>`              | 执行 JS 脚本       | `browser-bridge page execute 1 "document.title"`        |
| `page execute <tabId> -f <file>`             | 从文件加载脚本执行 | `browser-bridge page execute 1 -f script.js`            |
| `page click <tabId> --ref <ref>`             | 点击元素（by ref） | `browser-bridge page click 1 --ref r2`                  |
| `page click <tabId> --selector <sel>`        | 点击元素           | `browser-bridge page click 1 --selector "#search-btn"`  |
| `page type <tabId> --ref <ref> <text>`       | 输入文本（by ref） | `browser-bridge page type 1 --ref r1 "golang"`          |
| `page type <tabId> --selector <sel> <text>`  | 输入文本           | `browser-bridge page type 1 --selector "#kw" "golang"`  |
| `page select <tabId> --ref <ref> <val>`      | 选择下拉框(by ref) | `browser-bridge page select 1 --ref r5 "option1"`       |
| `page select <tabId> --selector <sel> <val>` | 选择下拉框         | `browser-bridge page select 1 --selector "#sel" "opt1"` |
| `page scroll <tabId> [x] [y]`                | 滚动页面           | `browser-bridge page scroll 1 0 500`                    |
| `page scroll <tabId> --ref <ref>`            | 滚动到元素(by ref) | `browser-bridge page scroll 1 --ref r3`                 |
| `page query <tabId> <selector>`              | 查询 DOM 元素      | `browser-bridge page query 1 ".result"`                 |
| `page wait <tabId> <selector>`               | 等待元素出现       | `browser-bridge page wait 1 ".result" --timeout 5000`   |
| `nav goto <tabId> <url>`                     | 导航到 URL         | `browser-bridge nav goto 1 https://github.com`          |
| `nav back <tabId>`                           | 后退               | `browser-bridge nav back 1`                             |
| `nav forward <tabId>`                        | 前进               | `browser-bridge nav forward 1`                          |
| `nav reload <tabId>`                         | 刷新页面           | `browser-bridge nav reload 1`                           |
| `cookie get [url]`                           | 获取 Cookie        | `browser-bridge cookie get https://baidu.com`           |
| `cookie set <url> <name> <value>`            | 设置 Cookie        | `browser-bridge cookie set https://baidu.com token abc` |
| `cookie delete <url> <name>`                 | 删除 Cookie        | `browser-bridge cookie delete https://baidu.com token`  |
| `search baidu <query>`                       | 百度搜索           | `browser-bridge search baidu "golang"`                  |
| `search bing <query>`                        | Bing 搜索          | `browser-bridge search bing "golang"`                   |

### 5.2 CLI 发现 Native Host 端口

CLI 通过读取 native-host 同目录下的 `nativehost_port` 文件获取 HTTP 服务端口：

```go
func getNativeHostPort() (int, error) {
    // 1. 查找 native-host 可执行文件同目录
    // 2. 读取 nativehost_port 文件
    // 3. 返回端口号
}
```

### 5.3 输出格式

默认 JSON 输出，支持 `--format table` 切换为表格：

```bash
# JSON 格式（默认）
$ browser-bridge tab list
{"success":true,"data":[{"id":1,"title":"百度一下",...}]}

# 表格格式
$ browser-bridge tab list --format table
ID    URL                          TITLE          ACTIVE
1     https://www.baidu.com        百度一下        true
2     https://github.com           GitHub          false
```

---

## 6. 安装与注册

### 6.1 Native Messaging Host 注册

Windows 下需要在注册表写入：

```
HKCU\Software\Google\Chrome\NativeMessagingHosts\com.browser.bridge
```

值为 `com.browser.bridge.json` 文件的路径。

### 6.2 com.browser.bridge.json

```json
{
  "name": "com.browser.bridge",
  "description": "Browser Bridge Native Host",
  "path": "C:\\path\\to\\native-host.exe",
  "type": "stdio",
  "allowed_origins": ["chrome-extension://EXTENSION_ID/"]
}
```

### 6.3 安装流程

1. 编译 native-host 和 cli
2. 将二进制文件放到目标目录
3. 运行 `install.bat`，写入注册表和生成 JSON 配置
4. 在 Chrome 加载 Extension（开发者模式）
5. 将 Extension ID 填入 `com.browser.bridge.json`

---

## 7. 数据流完整示例

### 示例：获取所有 Tab

```
1. 用户执行: browser-bridge tab list
2. CLI 读取 nativehost_port → 获取端口 3001
3. CLI 发送: GET http://127.0.0.1:3001/api/v1/tabs
4. Native Host 收到 HTTP 请求
5. Native Host 生成 request-id: "abc-123"
6. Native Host 将 {id:"abc-123", action:"tab.list", params:{}} 写入 stdout（4字节长度前缀）
7. Extension background.js 收到消息
8. Extension 执行 chrome.tabs.query({})
9. Extension 将结果 {id:"abc-123", success:true, data:[...]} 通过 Native Messaging 回传
10. Native Host 从 stdin 读取响应
11. Native Host 根据 id 匹配到等待的 HTTP 请求
12. Native Host 将结果作为 HTTP Response 返回给 CLI
13. CLI 格式化输出结果
```

---

## 8. 开发优先级

| 阶段 | 内容                                               | 优先级 |
| ---- | -------------------------------------------------- | ------ |
| P0   | Native Host HTTP 服务 + Native Messaging 双向通信  | 高     |
| P0   | Extension background.ts 消息处理                   | 高     |
| P0   | tab.list / tab.get / tab.create / tab.close 功能   | 高     |
| P0   | page.snapshot 快照 + ref 机制                      | 高     |
| P1   | page.content 获取页面内容                          | 中     |
| P1   | page.execute 执行脚本                              | 中     |
| P1   | page.screenshot 截图功能                           | 中     |
| P1   | page.click / page.type（支持 ref）页面交互         | 中     |
| P1   | nav.goto / nav.back / nav.forward 导航             | 中     |
| P1   | search.baidu / search.bing 搜索功能                | 中     |
| P2   | page.scroll / page.select / page.query / page.wait | 低     |
| P2   | cookie.get / cookie.set / cookie.delete            | 低     |
| P2   | CLI 表格输出                                       | 低     |
| P2   | 安装脚本自动化                                     | 低     |
| P2   | 错误处理与重连机制                                 | 低     |

---

## 9. 关键技术决策

| 决策                          | 选择                           | 原因                     |
| ----------------------------- | ------------------------------ | ------------------------ |
| Extension 开发语言            | TypeScript                     | 类型安全，提升可维护性   |
| Extension 构建                | esbuild                        | 极速构建，零配置打包     |
| Extension 与 Native Host 通信 | Chrome Native Messaging        | 官方标准，安全可靠       |
| CLI 与 Native Host 通信       | HTTP REST                      | 简单易用，跨语言兼容     |
| 请求-响应匹配                 | UUID + pending map             | 异步消息需要关联         |
| 脚本执行                      | chrome.scripting.executeScript | Manifest V3 推荐方式     |
| CLI 框架                      | cobra                          | Go 生态最成熟的 CLI 框架 |
| HTTP 路由                     | net/http + gorilla/mux         | 轻量，与现有代码一致     |

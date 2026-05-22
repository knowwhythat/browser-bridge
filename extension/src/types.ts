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
  | "search.bing"
  // Fetch
  | "fetch.content";

// Tab 参数类型
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

// 页面操作参数类型
export interface PageSnapshotParams {
  tabId: number;
  refOnly?: boolean;
}

export interface PageContentParams {
  tabId: number;
}

export interface PageScreenshotParams {
  tabId: number;
  format?: "png" | "jpeg";
  quality?: number;
  fullPage?: boolean;
  ref?: string;
  selector?: string;
}

export interface PageExecuteParams {
  tabId: number;
  script: string;
}

export interface PageClickParams {
  tabId: number;
  ref?: string;
  selector?: string;
}

export interface PageTypeParams {
  tabId: number;
  ref?: string;
  selector?: string;
  text: string;
  clear?: boolean;
}

export interface PageSelectParams {
  tabId: number;
  ref?: string;
  selector?: string;
  value: string;
}

export interface PageScrollParams {
  tabId: number;
  ref?: string;
  selector?: string;
  x?: number;
  y?: number;
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

// 导航参数类型
export interface NavGotoParams {
  tabId: number;
  url: string;
}

export interface NavWaitLoadParams {
  tabId: number;
  timeout?: number;
}

// Cookie 参数类型
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

// 搜索参数类型
export interface SearchParams {
  query: string;
}

// 结果类型
export interface TabInfo {
  id: number;
  url: string;
  title: string;
  favIconUrl?: string;
  active: boolean;
  windowId: number;
}

export interface SearchResultItem {
  title: string;
  url: string;
  snippet: string;
}

export interface SearchResult {
  query: string;
  engine: string;
  tabId: number;
  results: SearchResultItem[];
}

export interface ScreenshotResult {
  tabId: number;
  format: string;
  data: string;
}

export interface SnapshotResult {
  tabId: number;
  url: string;
  title: string;
  snapshot: SnapshotNode[];
}

export interface SnapshotNode {
  ref?: string;
  tagName: string;
  role?: string;
  name?: string;
  attributes?: Record<string, string>;
  children?: SnapshotNode[];
  textContent?: string;
}

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

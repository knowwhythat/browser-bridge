import { RequestMessage, ResponseMessage, ActionType } from "./types";
import { handleTabList, handleTabGet, handleTabCreate, handleTabClose, handleTabActivate } from "./actions/tab";
import {
  handlePageContent,
  handlePageExecute,
  handlePageScreenshot,
  handlePageClick,
  handlePageType,
  handlePageSelect,
  handlePageScroll,
  handlePageQuery,
  handlePageWait,
  handleFetchContent,
} from "./actions/page";
import { handlePageSnapshot, resolveRef } from "./actions/snapshot";
import { handleSearch } from "./actions/search";

const port = chrome.runtime.connectNative("com.browser.bridge");

port.onMessage.addListener(async (msg: RequestMessage) => {
  const { id, action, params } = msg;
  try {
    let result: unknown;
    switch (action as ActionType) {
      // Tab 管理
      case "tab.list":
        result = await handleTabList();
        break;
      case "tab.get":
        result = await handleTabGet(params as { tabId: number });
        break;
      case "tab.create":
        result = await handleTabCreate(params as { url?: string; active?: boolean });
        break;
      case "tab.close":
        result = await handleTabClose(params as { tabId: number });
        break;
      case "tab.activate":
        result = await handleTabActivate(params as { tabId: number });
        break;

      // 页面快照
      case "page.snapshot":
        result = await handlePageSnapshot(params as { tabId: number; refOnly?: boolean });
        break;

      // 页面内容
      case "page.content":
        result = await handlePageContent(params as { tabId: number });
        break;

      // 页面截图
      case "page.screenshot": {
        const p = params as { tabId: number; format?: string; quality?: number; fullPage?: boolean; ref?: string; selector?: string };
        if (p.ref) {
          p.selector = await resolveRef(p.tabId, p.ref);
        }
        result = await handlePageScreenshot(p as Parameters<typeof handlePageScreenshot>[0]);
        break;
      }

      // 执行脚本
      case "page.execute":
        result = await handlePageExecute(params as { tabId: number; script: string });
        break;

      // 点击元素
      case "page.click": {
        const p = params as { tabId: number; ref?: string; selector?: string };
        const selector = p.ref ? await resolveRef(p.tabId, p.ref) : p.selector!;
        result = await handlePageClick({ tabId: p.tabId, selector });
        break;
      }

      // 输入文本
      case "page.type": {
        const p = params as { tabId: number; ref?: string; selector?: string; text: string; clear?: boolean };
        const selector = p.ref ? await resolveRef(p.tabId, p.ref) : p.selector!;
        result = await handlePageType({ tabId: p.tabId, selector, text: p.text, clear: p.clear });
        break;
      }

      // 选择下拉框
      case "page.select": {
        const p = params as { tabId: number; ref?: string; selector?: string; value: string };
        const selector = p.ref ? await resolveRef(p.tabId, p.ref) : p.selector!;
        result = await handlePageSelect({ tabId: p.tabId, selector, value: p.value });
        break;
      }

      // 滚动页面
      case "page.scroll": {
        const p = params as { tabId: number; ref?: string; selector?: string; x?: number; y?: number };
        const selector = p.ref ? await resolveRef(p.tabId, p.ref) : p.selector;
        result = await handlePageScroll({ tabId: p.tabId, selector, x: p.x, y: p.y });
        break;
      }

      // 查询 DOM 元素
      case "page.query":
        result = await handlePageQuery(params as { tabId: number; selector: string; attributes?: string[] });
        break;

      // 等待元素出现
      case "page.wait":
        result = await handlePageWait(params as { tabId: number; selector: string; timeout?: number });
        break;

      // 导航
      case "nav.goto": {
        const p = params as { tabId: number; url: string };
        await chrome.tabs.update(p.tabId, { url: p.url });
        result = { tabId: p.tabId, url: p.url };
        break;
      }
      case "nav.back": {
        const tabId = (params as { tabId: number }).tabId;
        await chrome.scripting.executeScript({
          target: { tabId },
          func: () => { history.back(); },
        });
        result = { tabId };
        break;
      }
      case "nav.forward": {
        const tabId = (params as { tabId: number }).tabId;
        await chrome.scripting.executeScript({
          target: { tabId },
          func: () => { history.forward(); },
        });
        result = { tabId };
        break;
      }
      case "nav.reload": {
        await chrome.tabs.reload((params as { tabId: number }).tabId);
        result = { tabId: (params as { tabId: number }).tabId };
        break;
      }
      case "nav.waitLoad": {
        const p = params as { tabId: number; timeout?: number };
        const timeout = p.timeout ?? 10000;
        await new Promise<void>((resolve, reject) => {
          const timer = setTimeout(() => {
            chrome.tabs.onUpdated.removeListener(listener);
            reject(new Error("wait for page load timeout"));
          }, timeout);
          const listener = (tabId: number, info: chrome.tabs.TabChangeInfo) => {
            if (tabId === p.tabId && info.status === "complete") {
              clearTimeout(timer);
              chrome.tabs.onUpdated.removeListener(listener);
              resolve();
            }
          };
          chrome.tabs.onUpdated.addListener(listener);
        });
        result = { tabId: p.tabId };
        break;
      }

      // Cookie
      case "cookie.get": {
        const p = params as { url?: string; name?: string; domain?: string };
        const cookies = await chrome.cookies.getAll({ url: p.url, name: p.name, domain: p.domain });
        result = cookies;
        break;
      }
      case "cookie.set": {
        const p = params as { url: string; name: string; value: string; domain?: string; path?: string; secure?: boolean; httpOnly?: boolean; expirationDate?: number };
        const cookie = await chrome.cookies.set(p as chrome.cookies.SetDetails);
        result = cookie;
        break;
      }
      case "cookie.delete": {
        const p = params as { url: string; name: string };
        await chrome.cookies.remove({ url: p.url, name: p.name });
        result = { url: p.url, name: p.name, deleted: true };
        break;
      }

      // 搜索
      case "search.baidu":
        result = await handleSearch("https://www.baidu.com/s?wd=", params as { query: string });
        break;
      case "search.bing":
        result = await handleSearch("https://www.bing.com/search?q=", params as { query: string });
        break;

      // Fetch content
      case "fetch.content":
        result = await handleFetchContent(params as { url: string; timeout?: number });
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
  console.error("Native host disconnected:", chrome.runtime.lastError?.message);
});

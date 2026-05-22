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

export async function handleTabGet(params: { tabId: number }): Promise<TabInfo> {
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

export async function handleTabCreate(params: { url?: string; active?: boolean }): Promise<TabInfo> {
  const tab = await chrome.tabs.create({
    url: params.url,
    active: params.active ?? true,
  });
  return {
    id: tab.id!,
    url: tab.url ?? "",
    title: tab.title ?? "",
    favIconUrl: tab.favIconUrl,
    active: tab.active,
    windowId: tab.windowId,
  };
}

export async function handleTabClose(params: { tabId: number }): Promise<{ tabId: number }> {
  await chrome.tabs.remove(params.tabId);
  return { tabId: params.tabId };
}

export async function handleTabActivate(params: { tabId: number }): Promise<{ tabId: number }> {
  const tab = await chrome.tabs.get(params.tabId);
  await chrome.tabs.update(params.tabId, { active: true });
  await chrome.windows.update(tab.windowId, { focused: true });
  return { tabId: params.tabId };
}

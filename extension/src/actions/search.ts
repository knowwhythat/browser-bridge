interface ExtractedResult {
  title: string;
  url: string;
  snippet: string;
}

export async function handleSearch(
  baseUrl: string,
  params: { query: string }
): Promise<{
  query: string;
  engine: string;
  tabId: number;
  results: ExtractedResult[];
}> {
  const engine = baseUrl.includes("baidu.com") ? "baidu" : "bing";
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
    // 超时 15 秒
    setTimeout(() => {
      chrome.tabs.onUpdated.removeListener(listener);
      resolve();
    }, 15000);
  });

  // 提取搜索结果
  const results = await chrome.scripting.executeScript({
    target: { tabId: tab.id! },
    world: "MAIN",
    func: (eng: string): ExtractedResult[] => {
      const items: ExtractedResult[] = [];

      if (eng === "baidu") {
        // 百度搜索结果
        const resultElements = document.querySelectorAll(".result.c-container, .c-container");
        resultElements.forEach((el) => {
          const titleEl = el.querySelector("h3 a, .t a");

          if (titleEl) {
            const title = titleEl.textContent?.trim() ?? "";
            const url = (titleEl as HTMLAnchorElement).href ?? "";
            // 提取摘要：从 innerText 中去掉 title 行，取剩余文本
            const fullText = el.innerText?.trim() ?? "";
            const snippet = fullText.replace(title, "").trim().split("\n").filter((s: string) => s.trim()).slice(0, 2).join(" ");
            if (title && url) {
              items.push({ title, url, snippet });
            }
          }
        });
      } else if (eng === "bing") {
        // Bing 搜索结果
        const resultElements = document.querySelectorAll("#b_results > li.b_algo");
        resultElements.forEach((el) => {
          const titleEl = el.querySelector("h2 a");
          const snippetEl = el.querySelector(".b_caption p, .b_lineclamp2");

          if (titleEl) {
            const title = titleEl.textContent?.trim() ?? "";
            const url = (titleEl as HTMLAnchorElement).href ?? "";
            const snippet = snippetEl?.textContent?.trim() ?? "";
            if (title && url) {
              items.push({ title, url, snippet });
            }
          }
        });
      }

      return items;
    },
    args: [engine],
  });

  const searchResults = results[0]?.result as ExtractedResult[] ?? [];

  // 关闭搜索 Tab
  await chrome.tabs.remove(tab.id!);

  return {
    query: params.query,
    engine,
    tabId: tab.id!,
    results: searchResults,
  };
}

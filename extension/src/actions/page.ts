// 获取页面内容
export async function handlePageContent(params: {
  tabId: number;
  format?: string; // "html" | "markdown"
}): Promise<{ content: string; tabId: number; url: string; format: string }> {
  const tab = await chrome.tabs.get(params.tabId);
  const fmt = params.format || "html";

  if (fmt === "markdown") {
    const results = await chrome.scripting.executeScript({
      target: { tabId: params.tabId },
      func: () => {
        // Clone body to avoid modifying the actual page
        const clone = document.body.cloneNode(true) as HTMLElement;

        // Remove unwanted elements
        const removeSelectors = [
          "script",
          "style",
          "noscript",
          "svg",
          "canvas",
          "video",
          "audio",
          "iframe",
          "object",
          "embed",
          "form",
          "button",
          "input",
          "select",
          "textarea",
          "nav",
          "footer",
          "header[role='banner']",
          "[role='navigation']",
          "[role='complementary']",
          "[role='contentinfo']",
          ".ad",
          ".ads",
          ".advertisement",
          ".sidebar",
          ".nav",
          ".menu",
          ".footer",
          ".header",
          ".comment",
          ".comments",
          ".social",
          ".share",
          ".popup",
          ".modal",
          ".cookie-banner",
          ".cookie-notice",
        ];
        for (const sel of removeSelectors) {
          clone.querySelectorAll(sel).forEach((el) => el.remove());
        }

        // Convert HTML to markdown-like text
        function htmlToMarkdown(el: Element, indent = 0): string {
          const tag = el.tagName.toLowerCase();
          const lines: string[] = [];

          for (const child of Array.from(el.childNodes)) {
            if (child.nodeType === Node.TEXT_NODE) {
              const text = child.textContent?.trim();
              if (text) lines.push(text);
            } else if (child.nodeType === Node.ELEMENT_NODE) {
              const childEl = child as Element;
              const childTag = childEl.tagName.toLowerCase();

              // Skip hidden elements
              const style = (childEl as HTMLElement).style;
              if (style?.display === "none" || style?.visibility === "hidden")
                continue;

              switch (childTag) {
                case "h1":
                  lines.push(`# ${childEl.textContent?.trim()}`);
                  break;
                case "h2":
                  lines.push(`## ${childEl.textContent?.trim()}`);
                  break;
                case "h3":
                  lines.push(`### ${childEl.textContent?.trim()}`);
                  break;
                case "h4":
                  lines.push(`#### ${childEl.textContent?.trim()}`);
                  break;
                case "h5":
                  lines.push(`##### ${childEl.textContent?.trim()}`);
                  break;
                case "h6":
                  lines.push(`###### ${childEl.textContent?.trim()}`);
                  break;
                case "p":
                  lines.push(childEl.textContent?.trim() || "");
                  break;
                case "a": {
                  const href = childEl.getAttribute("href") || "";
                  const text = childEl.textContent?.trim() || "";
                  lines.push(`[${text}](${href})`);
                  break;
                }
                case "img": {
                  const src = childEl.getAttribute("src") || "";
                  const alt = childEl.getAttribute("alt") || "";
                  lines.push(`![${alt}](${src})`);
                  break;
                }
                case "strong":
                case "b":
                  lines.push(`**${childEl.textContent?.trim()}**`);
                  break;
                case "em":
                case "i":
                  lines.push(`*${childEl.textContent?.trim()}*`);
                  break;
                case "code":
                  lines.push(`\`${childEl.textContent?.trim()}\``);
                  break;
                case "pre":
                  lines.push(
                    "```\n" + (childEl.textContent?.trim() || "") + "\n```",
                  );
                  break;
                case "ul":
                case "ol": {
                  const items = childEl.querySelectorAll(":scope > li");
                  items.forEach((li, i) => {
                    const prefix = childTag === "ol" ? `${i + 1}. ` : "- ";
                    lines.push(prefix + li.textContent?.trim());
                  });
                  break;
                }
                case "br":
                  lines.push("\n");
                  break;
                case "hr":
                  lines.push("---");
                  break;
                case "table": {
                  const rows = childEl.querySelectorAll("tr");
                  rows.forEach((row, rowIdx) => {
                    const cells = Array.from(
                      row.querySelectorAll("th, td"),
                    ).map((c) => c.textContent?.trim() || "");
                    lines.push("| " + cells.join(" | ") + " |");
                    if (rowIdx === 0) {
                      lines.push(
                        "| " + cells.map(() => "---").join(" | ") + " |",
                      );
                    }
                  });
                  break;
                }
                case "blockquote":
                  childEl.textContent
                    ?.trim()
                    .split("\n")
                    .forEach((line) => {
                      lines.push(`> ${line}`);
                    });
                  break;
                default:
                  lines.push(htmlToMarkdown(childEl, indent));
              }
            }
          }

          // Add spacing for block elements
          const blockTags = [
            "h1",
            "h2",
            "h3",
            "h4",
            "h5",
            "h6",
            "p",
            "ul",
            "ol",
            "pre",
            "table",
            "blockquote",
            "hr",
            "div",
            "section",
            "article",
          ];
          if (blockTags.includes(tag) && indent === 0) {
            return lines.filter((l) => l !== "").join("\n\n");
          }
          return lines.join(" ");
        }

        const md = htmlToMarkdown(clone);
        // Clean up excessive blank lines
        return md.replace(/\n{3,}/g, "\n\n").trim();
      },
    });
    return {
      content: results[0].result as string,
      tabId: params.tabId,
      url: tab.url ?? "",
      format: "markdown",
    };
  }

  // Default: return raw HTML
  const results = await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: () => document.documentElement.outerHTML,
  });
  return {
    content: results[0].result as string,
    tabId: params.tabId,
    url: tab.url ?? "",
    format: "html",
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
  clipElement?: boolean;
}): Promise<{ tabId: number; format: string; data: string }> {
  const format = (params.format as "png" | "jpeg") || "png";

  // 如果指定了 selector + clipElement，只截取元素区域
  if (params.selector && params.clipElement) {
    // 先滚动到元素
    await chrome.scripting.executeScript({
      target: { tabId: params.tabId },
      func: (sel: string) => {
        const el = document.querySelector(sel);
        el?.scrollIntoView({ block: "start" });
      },
      args: [params.selector],
    });

    // 获取元素位置
    const rectResults = await chrome.scripting.executeScript({
      target: { tabId: params.tabId },
      func: (sel: string) => {
        const el = document.querySelector(sel);
        if (!el) return null;
        const rect = el.getBoundingClientRect();
        return { x: rect.x, y: rect.y, width: rect.width, height: rect.height };
      },
      args: [params.selector],
    });

    const rect = rectResults[0].result as {
      x: number;
      y: number;
      width: number;
      height: number;
    } | null;
    if (rect) {
      const dataUri = await chrome.tabs.captureVisibleTab(
        (await chrome.tabs.get(params.tabId)).windowId,
        {
          format,
          quality: params.quality,
        },
      );
      // Crop the screenshot to the element rect using an offscreen canvas
      const cropResults = await chrome.scripting.executeScript({
        target: { tabId: params.tabId },
        func: (
          uri: string,
          cropRect: { x: number; y: number; width: number; height: number },
          fmt: string,
        ) => {
          return new Promise<string>((resolve) => {
            const img = new Image();
            img.onload = () => {
              const canvas = document.createElement("canvas");
              canvas.width = cropRect.width;
              canvas.height = cropRect.height;
              const ctx = canvas.getContext("2d")!;
              ctx.drawImage(
                img,
                cropRect.x,
                cropRect.y,
                cropRect.width,
                cropRect.height,
                0,
                0,
                cropRect.width,
                cropRect.height,
              );
              const mimeType = fmt === "jpeg" ? "image/jpeg" : "image/png";
              const cropped = canvas.toDataURL(mimeType);
              resolve(cropped.split(",")[1]);
            };
            img.src = uri;
          });
        },
        args: [
          dataUri,
          {
            x: Math.round(rect.x),
            y: Math.round(rect.y),
            width: Math.round(rect.width),
            height: Math.round(rect.height),
          },
          format,
        ],
      });
      const base64 = cropResults[0].result as string;
      return { tabId: params.tabId, format, data: base64 };
    }
  }

  // 如果指定了 selector，先滚动到元素再截图
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
    world: "MAIN",
    func: (code: string) => {
      try {
        const evalFn = eval;
        return { success: true, value: evalFn(code) };
      } catch (e) {
        return { success: false, error: (e as Error).message };
      }
    },
    args: [params.script],
  });
  const res = results[0].result as {
    success: boolean;
    value: unknown;
    error?: string;
  };
  if (!res.success) {
    throw new Error(res.error ?? "script execution failed");
  }
  return {
    result: res.value,
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
      const el = document.querySelector(sel) as HTMLElement | null;
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
      const el = document.querySelector(sel) as HTMLInputElement | null;
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
      const el = document.querySelector(sel) as HTMLSelectElement | null;
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
    args: [params.selector ?? undefined, params.x ?? 0, params.y ?? 0],
  });
  return { tabId: params.tabId };
}

// 查询 DOM 元素
export async function handlePageQuery(params: {
  tabId: number;
  selector: string;
  attributes?: string[];
}): Promise<{ elements: unknown[] }> {
  const results = await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    func: (sel: string, attrs: string[] | undefined) => {
      const elements = document.querySelectorAll(sel);
      return Array.from(elements).map((el) => {
        const info: Record<string, unknown> = {
          tagName: el.tagName.toLowerCase(),
          textContent: el.textContent?.trim() ?? undefined,
          innerHTML: el.innerHTML,
          boundingRect: el.getBoundingClientRect().toJSON(),
        };
        const attrMap: Record<string, string> = {};
        if (attrs && attrs.length > 0) {
          for (const attr of attrs) {
            if (el.hasAttribute(attr)) attrMap[attr] = el.getAttribute(attr)!;
          }
        } else {
          for (const attr of Array.from(el.attributes)) {
            attrMap[attr.name] = attr.value;
          }
        }
        info.attributes = attrMap;
        return info;
      });
    },
    args: [params.selector, params.attributes ?? undefined],
  });
  return { elements: results[0].result as unknown[] };
}

// 等待元素出现
export async function handlePageWait(params: {
  tabId: number;
  selector: string;
  timeout?: number;
}): Promise<{ found: boolean; selector: string }> {
  const timeout = params.timeout ?? 5000;
  const results = await chrome.scripting.executeScript({
    target: { tabId: params.tabId },
    world: "MAIN",
    func: (sel: string, ms: number) => {
      return new Promise<boolean>((resolve) => {
        if (document.querySelector(sel)) {
          resolve(true);
          return;
        }
        const observer = new MutationObserver(() => {
          if (document.querySelector(sel)) {
            observer.disconnect();
            resolve(true);
          }
        });
        observer.observe(document.body, { childList: true, subtree: true });
        setTimeout(() => {
          observer.disconnect();
          resolve(!!document.querySelector(sel));
        }, ms);
      });
    },
    args: [params.selector, timeout],
  });
  return { found: results[0].result as boolean, selector: params.selector };
}

// Fetch content: open URL, wait for load, get markdown content, close tab
export async function handleFetchContent(params: {
  url: string;
  timeout?: number;
}): Promise<{ url: string; content: string; format: string }> {
  const tab = await chrome.tabs.create({ url: params.url, active: false });
  const timeout = params.timeout ?? 15000;

  // Wait for page load
  await new Promise<void>((resolve) => {
    const listener = (tabId: number, info: chrome.tabs.TabChangeInfo) => {
      if (tabId === tab.id && info.status === "complete") {
        chrome.tabs.onUpdated.removeListener(listener);
        resolve();
      }
    };
    chrome.tabs.onUpdated.addListener(listener);
    setTimeout(() => {
      chrome.tabs.onUpdated.removeListener(listener);
      resolve();
    }, timeout);
  });

  // Get markdown content
  const results = await chrome.scripting.executeScript({
    target: { tabId: tab.id! },
    func: () => {
      const clone = document.body.cloneNode(true) as HTMLElement;

      const removeSelectors = [
        "script",
        "style",
        "noscript",
        "svg",
        "canvas",
        "video",
        "audio",
        "iframe",
        "object",
        "embed",
        "form",
        "button",
        "input",
        "select",
        "textarea",
        "nav",
        "footer",
        "header[role='banner']",
        "[role='navigation']",
        "[role='complementary']",
        "[role='contentinfo']",
        ".ad",
        ".ads",
        ".advertisement",
        ".sidebar",
        ".nav",
        ".menu",
        ".footer",
        ".header",
        ".comment",
        ".comments",
        ".social",
        ".share",
        ".popup",
        ".modal",
        ".cookie-banner",
        ".cookie-notice",
      ];
      for (const sel of removeSelectors) {
        clone.querySelectorAll(sel).forEach((el) => el.remove());
      }

      function htmlToMarkdown(el: Element, indent = 0): string {
        const parts: string[] = [];
        for (const node of Array.from(el.childNodes)) {
          if (node.nodeType === Node.TEXT_NODE) {
            const text = node.textContent?.replace(/\s+/g, " ").trim();
            if (text) parts.push(text);
          } else if (node.nodeType === Node.ELEMENT_NODE) {
            const child = node as Element;
            const tag = child.tagName.toLowerCase();
            switch (tag) {
              case "h1":
                parts.push(`\n\n# ${child.textContent?.trim()}\n\n`);
                break;
              case "h2":
                parts.push(`\n\n## ${child.textContent?.trim()}\n\n`);
                break;
              case "h3":
                parts.push(`\n\n### ${child.textContent?.trim()}\n\n`);
                break;
              case "h4":
                parts.push(`\n\n#### ${child.textContent?.trim()}\n\n`);
                break;
              case "h5":
                parts.push(`\n\n##### ${child.textContent?.trim()}\n\n`);
                break;
              case "h6":
                parts.push(`\n\n###### ${child.textContent?.trim()}\n\n`);
                break;
              case "p":
                parts.push(`\n\n${htmlToMarkdown(child)}\n\n`);
                break;
              case "a": {
                const href = child.getAttribute("href") ?? "";
                const text = child.textContent?.trim() ?? "";
                parts.push(href ? `[${text}](${href})` : text);
                break;
              }
              case "img": {
                const src = child.getAttribute("src") ?? "";
                const alt = child.getAttribute("alt") ?? "";
                if (src) parts.push(`![${alt}](${src})`);
                break;
              }
              case "strong":
              case "b":
                parts.push(`**${child.textContent?.trim()}**`);
                break;
              case "em":
              case "i":
                parts.push(`*${child.textContent?.trim()}*`);
                break;
              case "code":
                parts.push(`\`${child.textContent?.trim()}\``);
                break;
              case "pre":
                parts.push(`\n\n\`\`\`\n${child.textContent}\n\`\`\`\n\n`);
                break;
              case "ul":
              case "ol": {
                const items = Array.from(child.children);
                items.forEach((li, i) => {
                  const prefix = tag === "ol" ? `${i + 1}. ` : "- ";
                  parts.push(
                    `\n${"  ".repeat(indent)}${prefix}${htmlToMarkdown(li, indent + 1).trim()}`,
                  );
                });
                parts.push("\n");
                break;
              }
              case "blockquote":
                parts.push(
                  `\n> ${htmlToMarkdown(child).trim().replace(/\n/g, "\n> ")}\n\n`,
                );
                break;
              case "br":
                parts.push("\n");
                break;
              case "hr":
                parts.push("\n\n---\n\n");
                break;
              case "table": {
                const rows = Array.from(child.querySelectorAll("tr"));
                rows.forEach((row, ri) => {
                  const cells = Array.from(row.querySelectorAll("th, td"));
                  parts.push(
                    `| ${cells.map((c) => c.textContent?.trim()).join(" | ")} |\n`,
                  );
                  if (ri === 0) {
                    parts.push(`| ${cells.map(() => "---").join(" | ")} |\n`);
                  }
                });
                parts.push("\n");
                break;
              }
              default:
                parts.push(htmlToMarkdown(child, indent));
                break;
            }
          }
        }
        return parts.join("");
      }

      const md = htmlToMarkdown(clone);
      return md.replace(/\n{3,}/g, "\n\n").trim();
    },
  });

  const content = (results[0]?.result as string) ?? "";

  // Close the tab
  await chrome.tabs.remove(tab.id!);

  return { url: params.url, content, format: "markdown" };
}

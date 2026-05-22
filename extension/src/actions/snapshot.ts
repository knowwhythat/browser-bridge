import { SnapshotResult, SnapshotNode } from "../types";

// ref → CSS selector 映射表（按 tabId 隔离）
const refMaps = new Map<number, Map<string, string>>();

// 可交互元素标签
const INTERACTIVE_TAGS = new Set([
  "a", "button", "input", "select", "textarea",
  "summary", "details", "option", "optgroup",
]);

// 可交互 ARIA roles
const INTERACTIVE_ROLES = new Set([
  "button", "link", "textbox", "checkbox", "radio",
  "combobox", "listbox", "menuitem", "tab",
  "slider", "spinbutton", "switch", "searchbox",
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
  const snapshot = params.refOnly ? filterRefOnly(processedTree) : processedTree;

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
    throw new Error(`ref "${ref}" not found, please call page.snapshot first to refresh refs`);
  }
  return map.get(ref)!;
}

// 在页面中采集可访问性树（注入到页面执行）
function collectAccessibilityTree(): SnapshotNode[] {
  function walk(el: Element): SnapshotNode & { _selector?: string } {
    const node: SnapshotNode & { _selector?: string } = {
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
    const children: (SnapshotNode & { _selector?: string })[] = [];
    for (const child of el.children) {
      const tag = child.tagName.toLowerCase();
      if (["script", "style", "noscript", "svg", "path"].includes(tag)) continue;
      children.push(walk(child));
    }
    if (children.length > 0) node.children = children;

    // 叶子文本节点
    if (children.length === 0 && el.childNodes.length > 0) {
      const text = el.textContent?.trim();
      if (text) node.textContent = text.slice(0, 200);
    }

    // 存储 selector 供 ref 分配使用（临时字段）
    node._selector = selector;

    return node;
  }

  function generateSelector(el: Element): string {
    if (el.id) return `#${CSS.escape(el.id)}`;
    const parent = el.parentElement;
    if (!parent) return el.tagName.toLowerCase();
    const siblings = Array.from(parent.children).filter(
      (c) => c.tagName === el.tagName
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
  nodes: (SnapshotNode & { _selector?: string })[],
  refMap: Map<string, string>,
  nextRef: () => string
): SnapshotNode[] {
  return nodes.map((node) => {
    const isInteractive =
      INTERACTIVE_TAGS.has(node.tagName) ||
      (!!node.role && INTERACTIVE_ROLES.has(node.role));

    const newNode: SnapshotNode = {
      tagName: node.tagName,
      ...(node.ref && { ref: node.ref }),
      ...(node.role && { role: node.role }),
      ...(node.name && { name: node.name }),
      ...(node.attributes && { attributes: node.attributes }),
      ...(node.textContent && { textContent: node.textContent }),
    };

    if (isInteractive && node._selector) {
      const ref = nextRef();
      newNode.ref = ref;
      refMap.set(ref, node._selector);
    }

    if (node.children) {
      newNode.children = assignRefs(
        node.children as (SnapshotNode & { _selector?: string })[],
        refMap,
        nextRef
      );
    }

    return newNode;
  });
}

// 仅保留带 ref 的节点（精简模式）
function filterRefOnly(nodes: SnapshotNode[]): SnapshotNode[] {
  const result: SnapshotNode[] = [];
  for (const node of nodes) {
    if (node.ref) {
      const filtered: SnapshotNode = {
        ref: node.ref,
        tagName: node.tagName,
        ...(node.role && { role: node.role }),
        ...(node.name && { name: node.name }),
        ...(node.attributes && { attributes: node.attributes }),
      };
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

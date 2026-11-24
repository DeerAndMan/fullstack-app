// ==UserScript==
// @name         订阅数据
// @namespace    http://tampermonkey.net/
// @version      2026-11-24.1
// @description  雪球的一些功能
// @author       You
// @match        https://xueqiu.com/*
// @match        https://www.xueqiu.com/*
// @icon         https://www.google.com/s2/favicons?sz=64&domain=undefined.localhost
// @grant        GM_xmlhttpRequest
// @connect      api.fuguilu.dpdns.org
// @require      https://lf26-cdn-tos.bytecdntp.com/cdn/expire-1-M/jquery/3.2.1/jquery.min.js
// @require      https://cdn.jsdelivr.net/npm/tesseract.js@5/dist/tesseract.min.js
// @require      https://cdn.jsdelivr.net/npm/interactjs/dist/interact.min.js
// @require      https://unpkg.com/@alenaksu/json-viewer@2.1.2/dist/json-viewer.bundle.js
// ==/UserScript==

/**
 * 重要事项记录：
 *
 * 1. 获取订阅的用户是否有数据更新
 * /statuses/home_timeline/unread.json?since_id=x&source=x&usergroup_id=x&_=x
 * 返回的数据格式 {"count":1} 如果count大于0，则有数据更新
 *
 * /v4/statuses/home_timeline.json?source=user&since_id=369516330&_=1767776389628 该接口为更新当前订阅的数据，
 *  返回的数据格式为 {}
 *  对应的点击className 为 home-timeline__unread 对应的元素
 */

const BOX_STORAGE_KEY = "xueqiu_drag_box_config";
const CONSOLE_BOX_STORAGE_KEY = "xueqiu_console_box_config";
const TIMELINE_UPLOAD_URL = "https://api.fuguilu.dpdns.org/api/v1/theme-contents/batch";
const _originalFetch = window.fetch.bind(window); // 保存原始 fetch，避免被拦截器捕获

const injectTailwind = () => {
  if (document.getElementById("tailwind-cdn")) return;
  const script = document.createElement("script");
  script.id = "tailwind-cdn";
  script.src = "https://cdn.tailwindcss.com";
  document.head.appendChild(script);
};

// 拖动框
const saveBoxState = (target) => {
  const config = {
    x: parseFloat(target.getAttribute("data-x")) || 0,
    y: parseFloat(target.getAttribute("data-y")) || 0,
    width: target.style.width,
    height: target.style.height,
    isCollapsed: target.getAttribute("data-collapsed") === "true",
  };
  localStorage.setItem(BOX_STORAGE_KEY, JSON.stringify(config));
};
const handleDragMove = (event) => {
  const target = event.target;
  const x = (parseFloat(target.getAttribute("data-x")) || 0) + event.dx;
  const y = (parseFloat(target.getAttribute("data-y")) || 0) + event.dy;

  target.style.transform = `translate(${x}px, ${y}px)`;
  target.setAttribute("data-x", x);
  target.setAttribute("data-y", y);
  saveBoxState(target);
};
const handleResizeMove = (event) => {
  const target = event.target;
  let x = parseFloat(target.getAttribute("data-x")) || 0;
  let y = parseFloat(target.getAttribute("data-y")) || 0;

  target.style.width = `${event.rect.width}px`;
  target.style.height = `${event.rect.height}px`;

  x += event.deltaRect.left;
  y += event.deltaRect.top;

  target.style.transform = `translate(${x}px, ${y}px)`;
  target.setAttribute("data-x", x);
  target.setAttribute("data-y", y);
  saveBoxState(target);
};

// 初始化拖动和缩放
const initInteract = (selector) => {
  if (typeof interact === "undefined") return;

  interact(selector)
    .draggable({
      inertia: true,
      modifiers: [
        interact.modifiers.restrictRect({
          restriction: "parent",
          endOnly: true,
        }),
      ],
      autoScroll: true,
      listeners: { move: handleDragMove },
    })
    .resizable({
      edges: { left: true, right: true, bottom: true, top: true },
      listeners: {
        move: handleResizeMove,
      },
      modifiers: [
        interact.modifiers.restrictEdges({ outer: "parent" }),
        interact.modifiers.restrictSize({ min: { width: 100, height: 100 } }),
      ],
      inertia: true,
    });
};

// 捕获全部的接口请求
const createTooltip = () => {
  let tooltip = document.getElementById("xq-tooltip");
  if (tooltip) return tooltip;

  tooltip = document.createElement("div");
  tooltip.id = "xq-tooltip";
  tooltip.className =
    "fixed hidden z-[2147483647] bg-gray-900 text-gray-100 p-4 rounded-lg shadow-2xl text-[12px] font-mono max-w-[800px] max-h-[500px] overflow-auto pointer-events-none border border-gray-700 whitespace-pre-wrap break-all leading-relaxed";
  document.body.appendChild(tooltip);
  return tooltip;
};

/**
 * 等待元素出现并触发点击
 * @param {string} selector 选择器
 * @param {number} timeout 超时时间 (ms)
 */
const waitForElementAndClick = (selector, timeout = 10000) => {
  const startTime = Date.now();

  const check = () => {
    // 如果自动化被禁用，立即停止轮询
    if (!window._xq_automation_enabled) {
      console.log(`[自动化] 监听已停止，取消等待元素 ${selector}`);
      return;
    }

    const element = document.querySelector(selector);
    if (element) {
      console.log(`[自动化] 找到元素 ${selector}，触发点击`);
      element.click();
      return;
    }

    if (Date.now() - startTime < timeout) {
      setTimeout(check, 500); // 每 500ms 检查一次
    } else {
      console.log(`[自动化] 等待元素 ${selector} 超时 (10s)`);
    }
  };

  check();
};

let isAutomationTriggered = false;
let unreadTriggerCount = 0; // 未读数触发累计计数，满 10 次自动刷新页面
// 将状态挂载到全局，确保拦截器能正确访问
window._xq_automation_enabled = true;

// 每次自动刷新只提交一批，订阅筛选由服务端处理。
const sendHomeTimelineToBackend = (homeTimeline) => {
  if (!Array.isArray(homeTimeline) || homeTimeline.length === 0) return;
  if (typeof GM_xmlhttpRequest !== "function") {
    console.error("[自动化] 无法上传，请检查油猴跨域请求授权");
    return;
  }

  GM_xmlhttpRequest({
    method: "POST",
    url: TIMELINE_UPLOAD_URL,
    anonymous: true,
    headers: { "Content-Type": "application/json", Accept: "application/json" },
    data: JSON.stringify({ home_timeline: homeTimeline }),
    responseType: "json",
    timeout: 15000,
    onload({ status, response }) {
      const count = response?.data?.count;
      if (status !== 200 || response?.code !== 0 || !Number.isInteger(count) || count < 0) {
        console.warn("[自动化] 订阅数据上传失败", status, response?.code);
        return;
      }
      console.log(`[自动化] 已提交 ${homeTimeline.length} 条，服务端匹配订阅 ${count} 条`);
    },
    onerror: () => console.warn("[自动化] 订阅数据上传失败，请检查网络"),
    ontimeout: () => console.warn("[自动化] 订阅数据上传超时"),
    onabort: () => console.warn("[自动化] 订阅数据上传已取消"),
  });
};

const handleSideEffectAutomation = (url, responseData) => {
  if (!window._xq_automation_enabled) return;

  // 1. 拦截未读数接口
  if (url.includes("/statuses/home_timeline/unread.json")) {
    try {
      const data = JSON.parse(responseData);
      if (data && typeof data.count === "number" && data.count >= 5) {
        unreadTriggerCount++;
        console.log(
          `[自动化] 检测到未读数 ${data.count} >= 5，累计触发 ${unreadTriggerCount}/10 次，开始轮询等待 DOM 节点...`,
        );
        if (unreadTriggerCount >= 10) {
          console.log("[自动化] 累计触发达到 10 次，自动刷新页面...");
          unreadTriggerCount = 0;
          location.reload();
          return;
        }
        isAutomationTriggered = true;
        waitForElementAndClick(".home-timeline__unread");
      }
    } catch (e) {
      // 解析失败忽略
    }
    return;
  }

  // 2. 拦截主时间线数据接口，仅在自动触发点击后才将数据写入数据库
  if (url.includes("/v4/statuses/home_timeline.json")) {
    try {
      const data = JSON.parse(responseData);
      console.log("[自动化] 监听到 home_timeline 数据返回:", data);

      // 只有由 waitForElementAndClick 自动触发的请求才保存到后端
      if (isAutomationTriggered) {
        isAutomationTriggered = false;
        const homeTimeline = data.home_timeline || data.list || [];
        if (homeTimeline.length > 0) {
          console.log(
            `[自动化] 自动触发的请求，开始保存 ${homeTimeline.length} 条数据到数据库...`,
          );
          sendHomeTimelineToBackend(homeTimeline);
        }
      }
    } catch (e) {
      // 解析失败忽略
    }
    return;
  }
};

const addLog = (url, method, status, type, responseData) => {
  if (!window._xq_automation_enabled) return;

  const container = document.getElementById("request-logs");
  if (!container) return;

  // 如果是第一条日志，清空等待提示
  const placeholder = container.querySelector(".placeholder-text");
  if (placeholder) {
    container.innerHTML = "";
  }

  const entry = document.createElement("div");
  entry.className =
    "border-b border-gray-100 py-1 px-2 break-all flex items-start gap-2 hover:bg-gray-50 transition-colors text-[11px]";

  const statusColor =
    status >= 200 && status < 300 ? "text-green-600" : "text-red-600";

  // 提取路径和简短 URL
  let displayUrl = url;
  try {
    const urlObj = new URL(url, window.location.origin);
    displayUrl = urlObj.pathname + urlObj.search;
  } catch (e) {
    displayUrl = url.substring(0, 50);
  }

  // 响应数据预览（取前100字符）
  const dataPreview = (responseData || "")
    .substring(0, 100)
    .replace(/\s+/g, " ");

  entry.innerHTML = `
    <span class="font-bold text-purple-600 shrink-0 text-[9px] px-1 bg-purple-50 rounded uppercase mt-0.5">${type}</span>
    <span class="font-bold text-blue-600 shrink-0 text-[9px] px-1 bg-blue-50 rounded uppercase mt-0.5">${method}</span>
    <span class="${statusColor} shrink-0 font-bold w-8 text-right">${status}</span>
    <span class="url-text text-gray-700 font-medium truncate flex-1 cursor-copy hover:text-blue-600 transition-colors" title="点击复制完整 URL">${displayUrl}</span>
    <span class="data-preview text-gray-400 w-64 shrink-0 truncate ml-2 border-l pl-2 cursor-copy hover:text-blue-600 transition-colors text-[10px] italic">${dataPreview}</span>
  `;

  const urlSpan = entry.querySelector(".url-text");
  const previewSpan = entry.querySelector(".data-preview");

  // 复制功能逻辑
  const copyToClipboard = (text, element) => {
    navigator.clipboard.writeText(text).then(() => {
      const originalColor = element.style.color;
      element.style.color = "#10b981"; // 绿色反馈
      setTimeout(() => {
        element.style.color = originalColor;
      }, 500);
    });
  };

  urlSpan.addEventListener("click", (e) => {
    e.stopPropagation();
    copyToClipboard(url, urlSpan);
  });

  previewSpan.addEventListener("click", (e) => {
    e.stopPropagation();
    copyToClipboard(responseData, previewSpan);
  });

  const handleMouseEnter = () => {
    const tooltip = createTooltip();
    let content = responseData;
    try {
      const json = JSON.parse(responseData);
      content = JSON.stringify(json, null, 2);
    } catch (err) {
      content = responseData || "[无响应数据]";
    }
    tooltip.textContent = content;
    tooltip.classList.remove("hidden");
  };

  const handleMouseMove = (e) => {
    const tooltip = document.getElementById("xq-tooltip");
    if (!tooltip) return;

    const offset = 15;
    let left = e.clientX + offset;
    let top = e.clientY + offset;

    if (left + tooltip.offsetWidth > window.innerWidth) {
      left = e.clientX - tooltip.offsetWidth - offset;
    }
    if (top + tooltip.offsetHeight > window.innerHeight) {
      top = e.clientY - tooltip.offsetHeight - offset;
    }

    tooltip.style.left = `${left}px`;
    tooltip.style.top = `${top}px`;
  };

  const handleMouseLeave = () => {
    const tooltip = document.getElementById("xq-tooltip");
    if (tooltip) tooltip.classList.add("hidden");
  };

  previewSpan.addEventListener("mouseenter", handleMouseEnter);
  previewSpan.addEventListener("mousemove", handleMouseMove);
  previewSpan.addEventListener("mouseleave", handleMouseLeave);

  container.prepend(entry);

  // 执行雪球自动化逻辑
  handleSideEffectAutomation(url, responseData);

  // 限制日志数量，防止内存占用过高
  if (container.children.length > 100) {
    container.removeChild(container.lastChild);
  }
};

/**
 * 拦截请求
 */
const interceptRequests = () => {
  // 拦截 Fetch
  const originalFetch = window.fetch;
  window.fetch = async (...args) => {
    const url = typeof args[0] === "string" ? args[0] : args[0].url;
    const method = (args[1] && args[1].method) || "GET";
    try {
      const response = await originalFetch(...args);
      if (window._xq_automation_enabled) {
        const clone = response.clone();
        const responseData = await clone.text();
        addLog(url, method, response.status, "FETCH", responseData);
      }
      return response;
    } catch (error) {
      if (window._xq_automation_enabled) {
        addLog(url, method, "ERR", "FETCH", error.message);
      }
      throw error;
    }
  };

  // 拦截 XHR
  const originalOpen = XMLHttpRequest.prototype.open;
  XMLHttpRequest.prototype.open = function (method, url) {
    this._url = url;
    this._method = method;
    return originalOpen.apply(this, arguments);
  };

  const originalSend = XMLHttpRequest.prototype.send;
  XMLHttpRequest.prototype.send = function () {
    this.addEventListener("load", function () {
      if (!window._xq_automation_enabled) return;
      addLog(this._url, this._method, this.status, "XHR", this.responseText);
    });
    this.addEventListener("error", function () {
      if (!window._xq_automation_enabled) return;
      addLog(this._url, this._method, "ERR", "XHR", "[请求失败]");
    });
    return originalSend.apply(this, arguments);
  };
};

// 初始化拖动框
const dragBox = () => {
  // 如果已存在，先移除旧的，确保 UI 更新
  const existingBox = document.getElementById("draggable-box");
  if (existingBox) existingBox.remove();

  injectTailwind();

  const savedConfig = JSON.parse(localStorage.getItem(BOX_STORAGE_KEY) || "{}");

  const box = document.createElement("div");
  box.id = "draggable-box";
  // 使用 z-[2147483647] 确保层级绝对最高
  box.className =
    "fixed top-20 left-20 bg-white border border-gray-300 shadow-2xl z-[2147483647] overflow-hidden flex flex-col cursor-move rounded-lg touch-none";

  // 设置初始大小和位置
  box.style.width = savedConfig.width || "400px";
  box.style.height = savedConfig.height || "300px";

  const initialX = savedConfig.x || 0;
  const initialY = savedConfig.y || 0;
  box.style.transform = `translate(${initialX}px, ${initialY}px)`;
  box.setAttribute("data-x", initialX);
  box.setAttribute("data-y", initialY);
  box.setAttribute("data-collapsed", savedConfig.isCollapsed || false);

  box.setAttribute("aria-label", "网络请求监控面板");
  box.setAttribute("tabindex", "0");

  const isInitialCollapsed = savedConfig.isCollapsed || false;

  // 内部 HTML 结构：头部 + 滚动内容区
  box.innerHTML = `
    <div class="p-2 border-b bg-gray-50 select-none text-gray-500 font-bold flex justify-between items-center shrink-0">
      <div class="flex items-center gap-2">
        <span id="status-indicator" class="w-2 h-2 rounded-full ${
          window._xq_automation_enabled
            ? "bg-green-500 animate-pulse"
            : "bg-red-500"
        }"></span>
        <span>监控控制台</span>
      </div>
      <div class="flex items-center gap-1">
        <button id="toggle-automation" class="text-[10px] px-2 py-1 ${
          window._xq_automation_enabled
            ? "bg-red-100 hover:bg-red-200 text-red-700"
            : "bg-green-100 hover:bg-green-200 text-green-700"
        } rounded transition-colors font-medium cursor-pointer">
          ${window._xq_automation_enabled ? "停止" : "开始"}
        </button>
        <button id="manual-trigger" class="text-[10px] px-2 py-1 bg-yellow-100 hover:bg-yellow-200 text-yellow-700 rounded transition-colors font-medium cursor-pointer">
          手动触发
        </button>
        <button id="clear-logs" class="text-[10px] px-2 py-1 bg-gray-200 hover:bg-gray-300 rounded transition-colors text-gray-700 font-medium cursor-pointer">
          清空
        </button>
        <button id="toggle-collapse" class="text-[10px] px-2 py-1 ${
          isInitialCollapsed
            ? "bg-green-100 hover:bg-green-200 text-green-700"
            : "bg-blue-100 hover:bg-blue-200 text-blue-700"
        } rounded transition-colors font-medium cursor-pointer">
          ${isInitialCollapsed ? "展开" : "收起"}
        </button>
      </div>
    </div>
    <div id="box-content" class="flex flex-col flex-1 overflow-hidden ${
      isInitialCollapsed ? "hidden" : ""
    }">
      <div id="request-logs" class="flex-1 overflow-y-auto p-2 text-xs font-mono bg-white scrollbar-thin scrollbar-thumb-gray-300">
        <div class="placeholder-text text-gray-400 text-center py-4 italic">等待请求中...</div>
      </div>
    </div>
  `;

  if (isInitialCollapsed) {
    box.style.height = "auto";
  }

  // 确保直接挂载在 body 最外层
  document.body.appendChild(box);

  // 1. 开始/停止按钮逻辑
  const toggleBtn = box.querySelector("#toggle-automation");
  const indicator = box.querySelector("#status-indicator");
  toggleBtn.addEventListener("click", () => {
    window._xq_automation_enabled = !window._xq_automation_enabled;
    if (!window._xq_automation_enabled) {
      isAutomationTriggered = false; // 停止时重置触发状态
    }
    toggleBtn.textContent = window._xq_automation_enabled ? "停止" : "开始";
    toggleBtn.className = `text-[10px] px-2 py-1 ${
      window._xq_automation_enabled
        ? "bg-red-100 hover:bg-red-200 text-red-700"
        : "bg-green-100 hover:bg-green-200 text-green-700"
    } rounded transition-colors font-medium cursor-pointer`;
    indicator.className = `w-2 h-2 rounded-full ${
      window._xq_automation_enabled
        ? "bg-green-500 animate-pulse"
        : "bg-red-500"
    }`;
    console.log(
      `[控制台] 自动化监听已${window._xq_automation_enabled ? "开启" : "关闭"}`,
    );
  });

  // 2. 收起/展开逻辑
  const collapseBtn = box.querySelector("#toggle-collapse");
  const content = box.querySelector("#box-content");
  let isCollapsed = isInitialCollapsed;
  const originalHeight = savedConfig.height || "300px";

  collapseBtn.addEventListener("click", () => {
    isCollapsed = !isCollapsed;
    box.setAttribute("data-collapsed", isCollapsed);
    if (isCollapsed) {
      content.classList.add("hidden");
      box.style.height = "auto";
      collapseBtn.textContent = "展开";
      collapseBtn.className =
        "text-[10px] px-2 py-1 bg-green-100 hover:bg-green-200 text-green-700 rounded transition-colors font-medium cursor-pointer";
    } else {
      content.classList.remove("hidden");
      box.style.height = originalHeight;
      collapseBtn.textContent = "收起";
      collapseBtn.className =
        "text-[10px] px-2 py-1 bg-blue-100 hover:bg-blue-200 text-blue-700 rounded transition-colors font-medium cursor-pointer";
    }
    saveBoxState(box);
  });

  // 3. 清空日志按钮逻辑
  const handleClearLogs = () => {
    const logs = document.getElementById("request-logs");
    if (logs)
      logs.innerHTML =
        '<div class="placeholder-text text-gray-400 text-center py-4 italic">等待请求中...</div>';
  };
  box.querySelector("#clear-logs").addEventListener("click", handleClearLogs);

  // 4. 手动触发按钮逻辑 — 模拟未读数触发的完整流程
  box.querySelector("#manual-trigger").addEventListener("click", () => {
    unreadTriggerCount++;
    console.log(
      `[自动化] 手动触发，累计触发 ${unreadTriggerCount}/10 次，开始轮询等待 DOM 节点...`,
    );
    if (unreadTriggerCount >= 10) {
      console.log("[自动化] 累计触发达到 10 次，自动刷新页面...");
      unreadTriggerCount = 0;
      location.reload();
      return;
    }
    isAutomationTriggered = true;
    waitForElementAndClick(".home-timeline__unread");
  });

  initInteract("#draggable-box");
};

// 保存日志面板状态
const saveConsoleBoxState = (target) => {
  const config = {
    x: parseFloat(target.getAttribute("data-x")) || 0,
    y: parseFloat(target.getAttribute("data-y")) || 0,
    width: target.style.width,
    height: target.style.height,
    isCollapsed: target.getAttribute("data-collapsed") === "true",
  };
  localStorage.setItem(CONSOLE_BOX_STORAGE_KEY, JSON.stringify(config));
};

// 日志面板 - 用于显示 console.log 内容
const consoleBox = () => {
  const existingBox = document.getElementById("console-box");
  if (existingBox) existingBox.remove();

  const savedConfig = JSON.parse(
    localStorage.getItem(CONSOLE_BOX_STORAGE_KEY) || "{}",
  );

  const box = document.createElement("div");
  box.id = "console-box";

  const isInitialCollapsed = savedConfig.isCollapsed || false;

  // 收起时为小圆球样式，展开时为正常面板
  if (isInitialCollapsed) {
    box.className =
      "fixed z-[2147483646] bg-gray-800 rounded-full shadow-2xl cursor-pointer flex items-center justify-center border-2 border-yellow-400/60 touch-none";
    box.title = "展开 Console 日志";
    box.innerHTML = `<span class="text-yellow-400 text-lg select-none">&#9654;</span>`;
    box.style.width = "46px";
    box.style.height = "46px";
    box.style.right = "20px";
    box.style.bottom = "20px";
    box.style.top = "auto";
    box.style.left = "auto";
    box.style.transform = "none";
    // hover 放大效果
    box.addEventListener(
      "mouseenter",
      () => (box.style.transform = "scale(1.15)"),
    );
    box.addEventListener("mouseleave", () => (box.style.transform = "none"));
  } else {
    box.className =
      "fixed top-20 bg-white border border-gray-300 shadow-2xl z-[2147483646] overflow-hidden flex flex-col cursor-move rounded-lg touch-none";
    box.style.width = savedConfig.width || "450px";
    box.style.height = savedConfig.height || "350px";
    box.innerHTML = `
      <div class="p-2 border-b bg-gray-800 select-none text-gray-300 font-bold flex justify-between items-center shrink-0">
        <div class="flex items-center gap-2">
          <span class="text-yellow-400 text-[10px]">&#9654;</span>
          <span class="text-[12px]">Console 日志</span>
          <span id="console-count" class="text-[9px] bg-gray-700 text-gray-400 px-1.5 rounded-full">0</span>
        </div>
        <div class="flex items-center gap-1">
          <button id="console-clear" class="text-[10px] px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors font-medium cursor-pointer">
            清空
          </button>
          <button id="console-collapse" class="text-[10px] px-2 py-1 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded transition-colors font-medium cursor-pointer">
            收起
          </button>
        </div>
      </div>
      <div id="console-content" class="flex-1 overflow-hidden flex flex-col">
        <div id="console-logs" class="flex-1 overflow-y-auto bg-gray-900 text-[12px] font-mono">
          <div class="console-placeholder text-gray-500 text-center py-6 italic">暂无日志...</div>
        </div>
      </div>
    `;
  }

  if (!isInitialCollapsed) {
    const initialX = savedConfig.x || 420;
    const initialY = savedConfig.y || 0;
    box.style.transform = `translate(${initialX}px, ${initialY}px)`;
    box.setAttribute("data-x", initialX);
    box.setAttribute("data-y", initialY);
  }
  box.setAttribute("data-collapsed", isInitialCollapsed);

  document.body.appendChild(box);

  // 注入 json-viewer 暗色主题样式
  if (!document.getElementById("json-viewer-style")) {
    const style = document.createElement("style");
    style.id = "json-viewer-style";
    style.textContent = `
      #console-logs json-viewer {
        --font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
        --font-size: 12px;
        --string-color: #86efac;
        --number-color: #c4b5fd;
        --boolean-color: #f9a8d4;
        --null-color: #fca5a5;
        --property-color: #93c5fd;
        --preview-color: #9ca3af;
        --background-color: transparent;
        --toggler-color: #fbbf24;
        display: inline-block;
        max-width: 100%;
      }
    `;
    document.head.appendChild(style);
  }

  let logCount = 0;

  // 添加日志条目到面板
  window._addConsoleEntry = (level, args) => {
    const container = document.getElementById("console-logs");
    if (!container) return;

    const placeholder = container.querySelector(".console-placeholder");
    if (placeholder) placeholder.remove();

    logCount++;
    const countEl = document.getElementById("console-count");
    if (countEl) countEl.textContent = logCount;

    const entry = document.createElement("div");
    const borderColor =
      level === "error"
        ? "border-red-800 bg-red-950/30"
        : level === "warn"
          ? "border-yellow-800 bg-yellow-950/30"
          : "border-gray-800 bg-transparent";
    entry.className = `px-3 py-1.5 border-b ${borderColor} hover:bg-gray-800/50 transition-colors`;

    // 时间标签
    const time = document.createElement("span");
    const now = new Date();
    time.className = "text-gray-600 text-[10px] mr-2 select-none";
    time.textContent = `${now.getHours().toString().padStart(2, "0")}:${now.getMinutes().toString().padStart(2, "0")}:${now.getSeconds().toString().padStart(2, "0")}`;
    entry.appendChild(time);

    // 级别标签
    const levelTag = document.createElement("span");
    const levelColors = {
      log: "text-gray-400",
      warn: "text-yellow-400",
      error: "text-red-400",
      info: "text-blue-400",
    };
    levelTag.className = `text-[9px] font-bold mr-2 ${levelColors[level] || "text-gray-400"}`;
    levelTag.textContent = level.toUpperCase();
    entry.appendChild(levelTag);

    // 渲染每个参数
    for (const arg of args) {
      if (
        typeof arg === "object" &&
        arg !== null &&
        customElements.get("json-viewer")
      ) {
        const viewer = document.createElement("json-viewer");
        viewer.data = arg;
        entry.appendChild(viewer);
      } else {
        const text = document.createElement("span");
        text.className = "text-gray-200 break-all";
        text.textContent = typeof arg === "string" ? arg : String(arg);
        entry.appendChild(text);
        entry.appendChild(document.createTextNode(" "));
      }
    }

    container.appendChild(entry);
    container.scrollTop = container.scrollHeight;

    // 限制日志数量
    while (container.children.length > 200) {
      container.removeChild(container.firstChild);
    }
  };

  // 收起/展开切换 — 通过重建 consoleBox 实现小圆球和面板的切换
  const toggleCollapse = (collapse) => {
    const currentConfig = JSON.parse(
      localStorage.getItem(CONSOLE_BOX_STORAGE_KEY) || "{}",
    );
    // 展开→收起时保存当前尺寸和位置
    if (collapse) {
      currentConfig.width = box.style.width;
      currentConfig.height = box.style.height;
    }
    currentConfig.x = parseFloat(box.getAttribute("data-x")) || 0;
    currentConfig.y = parseFloat(box.getAttribute("data-y")) || 0;
    currentConfig.isCollapsed = collapse;
    localStorage.setItem(
      CONSOLE_BOX_STORAGE_KEY,
      JSON.stringify(currentConfig),
    );
    consoleBox(); // 重建
  };

  if (isInitialCollapsed) {
    // 小圆球模式：点击展开
    box.addEventListener("click", () => toggleCollapse(false));
  } else {
    // 面板模式：收起按钮 + 清空按钮
    box
      .querySelector("#console-collapse")
      .addEventListener("click", () => toggleCollapse(true));

    box.querySelector("#console-clear").addEventListener("click", () => {
      const logs = document.getElementById("console-logs");
      if (logs) {
        logs.innerHTML =
          '<div class="console-placeholder text-gray-500 text-center py-6 italic">暂无日志...</div>';
        logCount = 0;
        const countEl = document.getElementById("console-count");
        if (countEl) countEl.textContent = "0";
      }
    });
  }

  // 展开状态才启用拖动和缩放（小圆球固定在右下角，不需要）
  if (typeof interact !== "undefined" && !isInitialCollapsed) {
    interact("#console-box").draggable({
      inertia: true,
      modifiers: [
        interact.modifiers.restrictRect({
          restriction: "parent",
          endOnly: true,
        }),
      ],
      autoScroll: true,
      listeners: {
        move(event) {
          const target = event.target;
          const x = (parseFloat(target.getAttribute("data-x")) || 0) + event.dx;
          const y = (parseFloat(target.getAttribute("data-y")) || 0) + event.dy;
          target.style.transform = `translate(${x}px, ${y}px)`;
          target.setAttribute("data-x", x);
          target.setAttribute("data-y", y);
          saveConsoleBoxState(target);
        },
      },
    });

    interact("#console-box").resizable({
      edges: { left: true, right: true, bottom: true, top: true },
      listeners: {
        move(event) {
          const target = event.target;
          let x = parseFloat(target.getAttribute("data-x")) || 0;
          let y = parseFloat(target.getAttribute("data-y")) || 0;
          target.style.width = `${event.rect.width}px`;
          target.style.height = `${event.rect.height}px`;
          x += event.deltaRect.left;
          y += event.deltaRect.top;
          target.style.transform = `translate(${x}px, ${y}px)`;
          target.setAttribute("data-x", x);
          target.setAttribute("data-y", y);
          saveConsoleBoxState(target);
        },
      },
      modifiers: [
        interact.modifiers.restrictEdges({ outer: "parent" }),
        interact.modifiers.restrictSize({ min: { width: 150, height: 100 } }),
      ],
      inertia: true,
    });
  }
};

// 拦截 console 方法，将日志同步到日志面板（过滤掉脚本自身的日志）
const interceptConsole = () => {
  const originalLog = console.log;
  const originalWarn = console.warn;
  const originalError = console.error;
  const originalInfo = console.info;

  const isInternalLog = (args) =>
    args.length > 0 &&
    typeof args[0] === "string" &&
    (args[0].startsWith("[自动化]") || args[0].startsWith("[控制台]"));

  console.log = function (...args) {
    originalLog.apply(console, args);
    if (window._addConsoleEntry && !isInternalLog(args))
      window._addConsoleEntry("log", args);
  };
  console.warn = function (...args) {
    originalWarn.apply(console, args);
    if (window._addConsoleEntry && !isInternalLog(args))
      window._addConsoleEntry("warn", args);
  };
  console.error = function (...args) {
    originalError.apply(console, args);
    if (window._addConsoleEntry && !isInternalLog(args))
      window._addConsoleEntry("error", args);
  };
  console.info = function (...args) {
    originalInfo.apply(console, args);
    if (window._addConsoleEntry && !isInternalLog(args))
      window._addConsoleEntry("info", args);
  };
};

(function () {
  "use strict";

  interceptConsole();
  interceptRequests();
  dragBox();
  consoleBox();
})();

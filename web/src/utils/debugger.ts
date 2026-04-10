/* eslint-disable */
// @ts-nocheck

// ==UserScript==
// @name         高级Debugger防护系统
// @namespace    http://tampermonkey.net/
// @version      1.1
// @description  全方位拦截网站反调试机制，支持白名单和日志记录
// @author       doubao
// @match        *://*/*
// @grant        document-start
// ==/UserScript==

(function () {
  "use strict";

  // 配置选项
  const config = {
    enableLogging: true, // 是否启用日志记录
    blockAllDebuggers: true, // 是否拦截所有debugger(包括自己的)
    useWhiteList: false, // 是否使用白名单模式
    whiteList: ["example.com"], // 白名单域名
  };

  // 日志系统
  function log(message, type = "info") {
    if (!config.enableLogging) return;
    const prefix = `[Debugger防护] [${type.toUpperCase()}]`;
    switch (type) {
      case "error":
        console.error(prefix, message);
        break;
      case "warn":
        console.warn(prefix, message);
        break;
      default:
        console.log(prefix, message);
    }
  }

  // 域名检查
  function isDomainWhitelisted() {
    if (!config.useWhiteList) return false;
    return config.whiteList.some(domain => window.location.hostname.includes(domain));
  }

  if (isDomainWhitelisted()) {
    log("当前域名在白名单中，防护系统未启用");
    return;
  }

  // 防护状态指示器
  function createStatusIndicator() {
    const indicator = document.createElement("div");
    indicator.id = "debugger-protector-status";
    indicator.style.cssText =
      "position:fixed;top:10px;right:10px;background:rgba(0,0,0,0.7);color:white;padding:5px 10px;border-radius:5px;font-size:12px;z-index:9999;";
    indicator.innerHTML = '<i class="fas fa-shield-alt"></i> Debugger防护已启动';
    document.body.appendChild(indicator);
  }

  // 创建调试检测陷阱
  function createDebugTrap() {
    let trapTriggered = false;
    setInterval(() => {
      const start = performance.now();
      debugger;
      const end = performance.now();

      // 如果debugger被触发，时间差会显著增加
      if (end - start > 100) {
        if (!trapTriggered) {
          log("检测到调试器活动！", "warn");
          trapTriggered = true;
        }
      } else {
        trapTriggered = false;
      }
    }, 5000);
  }

  // Hook构造函数
  Function.prototype._constructor = Function.prototype.constructor;
  Function.prototype.constructor = function () {
    const argsStr = arguments.toString();
    if (argsStr.includes("debugger")) {
      log(`拦截到构造函数注入debugger: ${argsStr.substring(0, 100)}...`);
      return null;
    }
    return Function.prototype._constructor.apply(this, arguments);
  };

  // Hook Function
  const F_ = Function;
  Function = function () {
    const args = Array.from(arguments);
    const content = args.join(" ");
    if (content.includes("debugger")) {
      log(`拦截到Function注入debugger: ${content.substring(0, 100)}...`);
      return () => {};
    }
    return F_.apply(this, arguments);
  };
  Function.prototype = F_.prototype;

  // Hook eval
  const eval_ = eval;
  unsafeWindow.eval = function (a) {
    const str = a.toString();
    if (str.includes("debugger")) {
      log(`拦截到eval注入debugger: ${str.substring(0, 100)}...`);
      return "";
    }
    return eval_.call(unsafeWindow, a);
  };

  // Hook Function.prototype.toString
  const originalToString = Function.prototype.toString;
  Function.prototype.toString = function () {
    if (this.name === "eval" || this.name === "Function") {
      return originalToString.call(this);
    }
    return originalToString.apply(this, arguments);
  };

  // Hook定时器
  const hookTimer = (originalTimer, timerName) => {
    return function (callback, delay, ...args) {
      let cbStr = callback.toString();
      if (typeof callback === "string") {
        cbStr = callback;
      }
      if (cbStr.includes("debugger")) {
        log(`拦截到${timerName}注入debugger: ${cbStr.substring(0, 100)}...`);
        return null;
      }
      return originalTimer.apply(this, [callback, delay, ...args]);
    };
  };

  window.setInterval = hookTimer(window.setInterval, "setInterval");
  window.setTimeout = hookTimer(window.setTimeout, "setTimeout");
  window.requestAnimationFrame = hookTimer(window.requestAnimationFrame, "requestAnimationFrame");

  // Hook DOM操作
  const oldAppendChild = Node.prototype.appendChild;
  Node.prototype.appendChild = function () {
    if (arguments[0] && arguments[0].innerHTML) {
      const html = arguments[0].innerHTML;
      if (html.includes("debugger")) {
        const newHtml = html.replace(/debugger/g, "// debugger已被拦截");
        arguments[0].innerHTML = newHtml;
        log(`拦截到DOM注入debugger，已清理内容`);
      }
    }
    return oldAppendChild.apply(this, arguments);
  };

  // Hook属性设置
  const oldSetAttribute = Element.prototype.setAttribute;
  Element.prototype.setAttribute = function (name, value) {
    if (typeof value === "string" && value.includes("debugger")) {
      const newValue = value.replace(/debugger/g, "// debugger已被拦截");
      log(`拦截到属性注入debugger: ${name}`);
      return oldSetAttribute.call(this, name, newValue);
    }
    return oldSetAttribute.apply(this, arguments);
  };

  // Hook事件监听
  const oldAddEventListener = EventTarget.prototype.addEventListener;
  EventTarget.prototype.addEventListener = function (type, listener, options) {
    if (typeof listener === "function") {
      const listenerStr = listener.toString();
      if (listenerStr.includes("debugger")) {
        log(`拦截到事件监听器注入debugger: ${type}`);
        return;
      }
    }
    return oldAddEventListener.apply(this, arguments);
  };

  // Hook Web Workers
  if (window.Worker) {
    const OriginalWorker = window.Worker;
    window.Worker = function (url, options) {
      if (typeof url === "string") {
        fetch(url)
          .then(response => response.text())
          .then(text => {
            if (text.includes("debugger")) {
              log(`拦截到Worker注入debugger: ${url}`);
            }
          })
          .catch(err => log(`检查Worker失败: ${err.message}`, "error"));
      }
      return new OriginalWorker(url, options);
    };
  }

  // 启动防护
  createStatusIndicator();
  if (config.blockAllDebuggers) {
    createDebugTrap();
  }
  log("高级Debugger防护系统已启动");
})();

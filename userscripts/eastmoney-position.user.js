// ==UserScript==
// @name         自动刷新 东方
// @namespace    http://tampermonkey.net/
// @version      0.2
// @description  自动刷新页面
// @author       You
// @match        https://jywg.18.cn/Search/Position
// @icon         data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw==
// @grant        GM_xmlhttpRequest
// @connect      api.fuguilu.dpdns.org
// @require      https://cdn.bootcdn.net/ajax/libs/jquery/3.6.0/jquery.min.js
/* globals jQuery, $, waitForKeyElements */

// ==/UserScript==

!(function () {
  "use strict";

  const tableThead = document.querySelector(".listtable > table > thead");
  const tableTbody = document.querySelector(".listtable > table > tbody");
  // thead 插入当日盈亏 比率
  tableThead.children[0].innerHTML += "<td>盈亏</td><td>盈亏比率</td>";

  const Base__Name = "tampermonkey_";
  const wrapper_Dom_div = `<div id="${Base__Name}wrapper"></div>`;
  $("body").append(wrapper_Dom_div);
  $("#tampermonkey_wrapper").css({
    position: "fixed",
    zIndex: 2021,
    width: "200px",
    height: "100px",
    top: 0,
    left: "30%",
    backgroundColor: "#e1eff3",
    boxShadow: "0 2px 12px 0 rgba(0, 0, 0, 0.1)",
  });
  /*  */
  const timeList = [
    1, 3, 5, 30, 60, 180, 300, 600, 900, 1200, 1800, 3600, 7200,
  ];
  const start_radio = 6;
  let formRadioList = ""; // 添加的dom 元素字符串
  timeList.forEach((item, i) => {
    formRadioList += `
      <label>
        <input class="radioTamper"
          type="radio"
          name="radioTamper"
          ${i == start_radio ? "checked" : ""}
          class="${Base__Name}control-radio"
          value=${item}
        >${item < 60 ? item + "秒" : item / 60 + "分"}
      </label>
    `;
  });
  $("#tampermonkey_wrapper").append(
    `<div id="${Base__Name}">${formRadioList}</div>`,
  );

  $("#tampermonkey_wrapper").append(`
    <div id="${Base__Name}butBox">
      <button class="${Base__Name}start">开始</button>
      <button class="${Base__Name}stop">暂停</button>
    </div>
  `);
  /** 计时器 */
  const startTimer = new Date().getTime();
  /** 刷新时间 */
  let countdownTimer = 2 * 60 * 60 * 1000;

  /** 定时器 */
  let interVal = 0;

  /** 接口启动，启动定时器 */
  function startGetData() {
    // /* 获取单选框的值 */
    // const radio_val = $("input[name='radioTamper']:checked").val();
    // clearInterval(interVal);

    // interVal = setInterval(() => {}, 1000 * radio_val || 1000);
    // console.log("定时器的状态 已开启", interVal)

    // let $favicon = document.querySelector('link[type="image/x-icon"]');
    // $favicon.href = "https://www.baidu.com/favicon.ico";
    sideEffect();

    document.title = "定时器已经开启";
    try {
      // 自动刷新函数
      init();
      // 获取全部的数据
      utools.loadAjax(
        "/Com/queryAssetAndPositionV1",
        { moneyType: "RMB" },
        function (result) {
          if (result && result.Status == "0") {
            const data = result.Data;
            // console.log('全部的数据', data, data[0]?.positions)
            document.title = "开启  " + data[0]?.Dryk;
            if (new Date().getHours() >= 15) {
              document.title = "超出  " + data[0]?.Dryk;
              clearInterval(interVal);
              window.sessionStorage.setItem("tmk_start", 0);
            }
            setTimeout(() => {
              data[0]?.positions.forEach((item, i) => {
                // console.log(item, i)
                const style =
                  item.Dryk >= 0
                    ? "color: red;font-weight: bold;"
                    : "color: green;font-weight: bold;";
                tableTbody.children[i].innerHTML += `
                    <td><span style="${style}">${item.Dryk}</span></td>
                    <td><span style="${style}">${
                      (item.Drykbl * 100).toFixed(5) + "%"
                    }</span></td>
                  `;
              });
            }, 100);

            if (Array.isArray(data) && data.length > 0) {
              GM_xmlhttpRequest({
                method: "POST",
                url: "https://api.fuguilu.dpdns.org/api/v1/energy/asset",
                anonymous: true,
                headers: { "Content-Type": "application/json; charset=utf-8" },
                data: JSON.stringify(data),
                responseType: "json",
                timeout: 15000,
                onload(response) {
                  if (response.status !== 200 || response.response?.code !== 0) {
                    console.warn("持仓上传失败", response.status, response.response?.code);
                  }
                },
                onerror() { console.warn("持仓上传失败，请检查网络连接"); },
                ontimeout() { console.warn("持仓上传超时"); },
              });
            }
          }
        },
      );
    } catch (error) {
      // localStorage.removeItem('DFCF___wealth')
    }
    // console.log( interVal, 'interVal' )
  }

  // 配置目标时间段（09:10-11:50、12:50-15:10）
  const targetRanges = [
    [9, 10, 11, 50], // 第一时段
    [12, 50, 15, 20], // 第二时段
  ];
  // 检查当前时间是否在任意目标时段内
  function isInTargetRange() {
    const now = new Date();
    const currentHour = now.getHours();
    const currentMinute = now.getMinutes();
    const currentTotalMin = currentHour * 60 + currentMinute;

    return targetRanges.some((range) => {
      const [startH, startM, endH, endM] = range;
      const startTotalMin = startH * 60 + startM;
      const endTotalMin = endH * 60 + endM;
      return currentTotalMin >= startTotalMin && currentTotalMin <= endTotalMin;
    });
  }

  /** 启动定时器 */
  function startInter() {
    /* 获取单选框的值 */
    const radio_val = $("input[name='radioTamper']:checked").val();
    clearInterval(interVal);
    interVal = setInterval(
      () => {
        if (isInTargetRange()) {
          console.log("现在的时间：", new Date().toLocaleString());
          startGetData();
          getEnergy();
        }
      },
      1000 * radio_val || 1000,
    );
  }
  /** 关闭接口，清理定时器 */
  function clearInter() {
    document.title = "定时器已关闭";
    clearInterval(interVal);
    window.sessionStorage.clear();
    // localStorage.removeItem('DFCF___wealth')
    // console.log("定时器已关闭")
  }

  $("." + Base__Name + "start")
    .off("click")
    .on("click", () => {
      /* 获取单选框的值 */
      const radio_val = $("input[name='radioTamper']:checked").val();
      // console.log("radio_val", radio_val);
      clearInterval(interVal);
      window.sessionStorage.clear();
      window.sessionStorage.setItem("tmk_start", radio_val);
      startGetData();
      startInter();
    });

  $("." + Base__Name + "stop")
    .off("click")
    .on("click", clearInter);

  /* 拖拽 */
  let scroll_scrollLeft = 0,
    scroll_scrollTop = 0;
  let mouse_pageX = 0,
    mouse_pageY = 0;
  let wrapper__pageX = 0,
    wrapper__pageY = 0;
  let wrapper__offsetLeft = 0,
    wrapper__offsetTop = 0;
  $("#" + Base__Name + "wrapper")
    .off("mousedown")
    .on("mousedown", (e) => {
      wrapper__pageX = e.clientX;
      wrapper__pageY = e.clientY;
      wrapper__offsetLeft = $("#" + Base__Name + "wrapper").offset().left;
      wrapper__offsetTop = $("#" + Base__Name + "wrapper").offset().top;

      scroll_scrollLeft = document.documentElement.scrollLeft;
      scroll_scrollTop = document.documentElement.scrollTop;

      mouse_pageX = e.clientX - wrapper__offsetLeft;
      mouse_pageY = e.clientY - wrapper__offsetTop;

      // console.log( '点击处于X位置', e.clientX, mouse_pageX, '盒子边框 Top 的距离', wrapper__offsetLeft, )
      // console.log( '点击处于Y位置', e.clientY, mouse_pageY, '盒子边框 Top 的距离', wrapper__offsetTop )
      wrapperDivMove();
    });

  function wrapperDivMove() {
    $("#" + Base__Name + "wrapper")
      .off("mousemove")
      .on("mousemove", (e) => {
        // let left = e.clientX -  mouse_pageX
        // let top = e.clientY -  mouse_pageY

        // console.log( '滚动条', scroll_scrollLeft, scroll_scrollTop)

        let left =
          e.clientX -
          (wrapper__pageX - wrapper__offsetLeft) -
          scroll_scrollLeft;
        let top =
          e.clientY - (wrapper__pageY - wrapper__offsetTop) - scroll_scrollTop;
        // console.log( left, top )
        $("#" + Base__Name + "wrapper").css({
          top,
          left,
        });
      });
  }

  $(document)
    .off("mouseup")
    .on("mouseup", () => {
      $("#" + Base__Name + "wrapper").off("mousemove");
    });

  // 一些副作用操作
  function sideEffect() {
    let $favicon = document.querySelector('link[type="image/x-icon"]');
    $favicon.href = "https://xqdoc.imedao.com/17aebcfb84a145d33fc18679.ico"; // "https://www.baidu.com/favicon.ico";
    // Object.prototype.toString.call(imgList)

    $("img").each(function () {
      $(this).attr("src", "");
    });
  }

  const locationStart = window.sessionStorage.getItem("tmk_start");

  // 定时刷新
  const timerReload = () => {
    const locationStart = window.sessionStorage.getItem("tmk_start");
    // console.log("locationStart", locationStart);
    if (!locationStart) return;
    window.sessionStorage.setItem("tmk_start", locationStart);
    const endTime = new Date().getTime();
    // console.log("startTimer", startTimer, "endTime", endTime);
    // console.log("间隔时间", endTime - startTimer, countdownTimer);

    setTimeout(() => {
      $("input[name='radioTamper'][value='" + locationStart + "']").prop(
        "checked",
        true,
      );
      startGetData();
      startInter();
    }, 100);

    if (endTime - startTimer > countdownTimer) {
      window.location.reload();
    }
  };

  if (locationStart) {
    timerReload();
  }

  getEnergy();
})();

function getValidateKey() {
  const validateValue = $("#em_validatekey").val();
  return validateValue;
}

async function getEnergy() {
  const validateValue = getValidateKey();
  if (!validateValue) return;

  const response = await $.ajax({
    type: "POST",
    url: "/Com/queryAssetAndPositionV1" + "?validatekey=" + validateValue,
    // headers: {
    //   gw_reqtimestamp: new Date().getTime(),
    // },
  });

  const jsonResponse = JSON.parse(response);
  if (jsonResponse.Status == "0") {
    const data = jsonResponse.Data;
    console.log("data", data);
  }
}

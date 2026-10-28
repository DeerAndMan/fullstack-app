// @vitest-environment jsdom
import { describe, expect, it } from "vitest";
import { renderToStaticMarkup } from "react-dom/server";

import PostContent, { sanitizePostContent } from "@/sections/subscribe/detail/post-content";

function parse(html: string) {
  const container = document.createElement("div");
  container.innerHTML = html;
  return container;
}

describe("订阅帖子正文", () => {
  it("显示超过 80 字的完整正文，保留换行、实体和配图", () => {
    const text = "完整正文".repeat(30);
    const content = parse(renderToStaticMarkup(
      <PostContent text={`${text}<br/>最后一段 &amp; 附图<img class="ke_img" src="https://xqimg.imedao.com/example.jpeg!custom.jpg" />`} />
    ));

    expect(content.textContent).toBe(`${text}最后一段 & 附图`);
    expect(content.querySelector("br")).not.toBeNull();
    expect(content.querySelector("img")?.getAttribute("src")).toBe(
      "https://xqimg.imedao.com/example.jpeg!custom.jpg"
    );
  });

  it("保留表情和大小，将协议相对图片地址转换为 HTTPS", () => {
    const content = parse(sanitizePostContent(
      '怎么又砸了<img src="//assets.imedao.com/ugc/images/face/emoji_77_vision.png?v=1" title="[流鼻血]" alt="[流鼻血]" height="24" />'
    ));
    const img = content.querySelector("img");

    expect(img?.getAttribute("src")).toBe("https://assets.imedao.com/ugc/images/face/emoji_77_vision.png?v=1");
    expect(img?.getAttribute("alt")).toBe("[流鼻血]");
    expect(img?.getAttribute("height")).toBe("24");
    expect(img?.getAttribute("referrerpolicy")).toBe("no-referrer");
  });

  it("过滤脚本、事件、危险链接和会隐藏正文的样式", () => {
    const content = parse(sanitizePostContent(
      '<script>alert(1)</script><iframe src="https://example.com"></iframe>' +
      '<div style="display:none" class="hidden" onclick="alert(1)">保留正文</div>' +
      '<img src="https://example.com/image.jpg" onerror="alert(1)">' +
      '<a href="javascript:alert(1)">危险链接</a>'
    ));

    expect(content.querySelector("script, iframe, [style], [class], [onclick], [onerror]")).toBeNull();
    expect(content.querySelector("a")?.getAttribute("href")).toBeNull();
    expect(content.textContent).toContain("保留正文");
  });

  it("保留正常链接并在新标签页打开", () => {
    const link = parse(sanitizePostContent('<a href="https://xueqiu.com/example">原文</a>')).querySelector("a");

    expect(link?.getAttribute("href")).toBe("https://xueqiu.com/example");
    expect(link?.getAttribute("target")).toBe("_blank");
    expect(link?.getAttribute("rel")).toBe("noopener noreferrer");
  });

  it.each([undefined, null, ""])("空正文 %s 显示占位符", text => {
    expect(renderToStaticMarkup(<PostContent text={text} />)).toBe("<span>-</span>");
  });
});

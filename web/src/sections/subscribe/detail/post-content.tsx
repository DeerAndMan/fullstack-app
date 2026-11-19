import { useMemo, useState, type SyntheticEvent } from "react";
import { Image } from "antd";
import DOMPurify from "dompurify";

import "./post-content.css";

export function sanitizePostContent(text: string): string {
  const content = DOMPurify.sanitize(text, {
    ALLOWED_TAGS: [
      "p", "div", "span", "br", "img", "a", "b", "strong", "i", "em",
      "u", "s", "blockquote", "ul", "ol", "li", "pre", "code",
    ],
    ALLOWED_ATTR: ["src", "alt", "title", "href", "height"],
    ALLOW_DATA_ATTR: false,
    ALLOW_ARIA_ATTR: false,
    RETURN_DOM_FRAGMENT: true,
  });

  content.querySelectorAll("img").forEach(img => {
    // 雪球的表情图使用 // 开头的地址，统一通过 HTTPS 加载。
    const src = img.getAttribute("src");
    if (src?.startsWith("//")) img.setAttribute("src", `https:${src}`);
    img.setAttribute("referrerpolicy", "no-referrer");
    img.setAttribute("loading", "lazy");
    if (src && img.getAttribute("height") !== "24") {
      img.setAttribute("role", "button");
      img.setAttribute("tabindex", "0");
      img.setAttribute("aria-label", "点击查看大图");
      img.setAttribute("title", "点击查看大图");
    }
  });
  content.querySelectorAll("a").forEach(link => {
    link.setAttribute("target", "_blank");
    link.setAttribute("rel", "noopener noreferrer");
  });

  const container = document.createElement("div");
  container.append(content);
  return container.innerHTML;
}

export default function PostContent({ text }: { text?: string | null }) {
  const html = useMemo(() => sanitizePostContent(text || ""), [text]);
  const [previewSrc, setPreviewSrc] = useState<string | null>(null);

  function openPreview(event: SyntheticEvent<HTMLDivElement>) {
    const target = event.target;
    if (!(target instanceof HTMLImageElement) || target.getAttribute("height") === "24") return;
    const src = target.getAttribute("src");
    if (!src) return;
    event.preventDefault();
    setPreviewSrc(src);
  }

  return html ? (
    <>
      <div
        className="subscribe-post-content"
        dangerouslySetInnerHTML={{ __html: html }}
        onClick={openPreview}
        onKeyDown={event => {
          if (event.key === "Enter" || event.key === " ") openPreview(event);
        }}
      />
      {previewSrc && (
        <Image
          src={previewSrc}
          referrerPolicy="no-referrer"
          wrapperStyle={{ display: "none" }}
          preview={{
            visible: true,
            onVisibleChange: visible => {
              if (!visible) setPreviewSrc(null);
            },
          }}
        />
      )}
    </>
  ) : (
    <span>-</span>
  );
}

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import express from "express";
import cookieParser from "cookie-parser";
import type { Express } from "express";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export async function setupProd(app: Express) {
  app.use(cookieParser());

  // 静态资源
  app.use(express.static(path.resolve(__dirname, "../dist/client"), { index: false }));

  // 预加载构建产物
  // @ts-expect-error 构建产物无类型声明，下方已手动断言类型
  const ssrMod = await import("../dist/server/ssr-entry.js");
  const render = ssrMod.render as (url: string, context?: { token?: string }) => { html: string };
  const ssrPaths = (ssrMod.ssrPaths ?? []) as string[];
  const ssrRoutes = (ssrMod.ssrRoutes ?? []) as { path: string; title: string }[];
  const spaRoutes = (ssrMod.spaRoutes ?? []) as { path: string; title: string }[];
  const template = fs.readFileSync(path.resolve(__dirname, "../dist/client/index.html"), "utf-8");

  app.use("/{*path}", async (req, res, next) => {
    const url = req.originalUrl;

    try {
      const isSSR = ssrPaths.includes(url.split("?")[0]);

      if (isSSR) {
        // 从 Cookie 读取 token 并传入 SSR 上下文
        const token = req.cookies?.auth_token;
        const { html: appHtml } = render(url, { token });
        const html = template.replace('<div id="root"></div>', `<div id="root">${appHtml}</div>`);
        res.status(200).set({ "Content-Type": "text/html" }).end(html);
      } else {
        res.status(200).set({ "Content-Type": "text/html" }).end(template);
      }
    } catch (e) {
      next(e);
    }
  });

  return { ssrRoutes, spaRoutes };
}

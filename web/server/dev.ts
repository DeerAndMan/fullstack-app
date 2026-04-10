import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import cookieParser from "cookie-parser";

import type { Express } from "express";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export async function setupDev(app: Express) {
  const { createServer: createViteServer } = await import("vite");

  const vite = await createViteServer({
    server: { middlewareMode: true },
    appType: "custom",
  });

  app.use(cookieParser());
  app.use(vite.middlewares);

  app.use("/{*path}", async (req, res, next) => {
    const url = req.originalUrl;

    try {
      const { ssrPaths } = await vite.ssrLoadModule("/src/router/section/ssr-routes.ts");
      const isSSR = (ssrPaths as string[]).includes(url.split("?")[0]);

      let template = fs.readFileSync(path.resolve(__dirname, "../index.html"), "utf-8");
      template = await vite.transformIndexHtml(url, template);

      if (isSSR) {
        const { render } = await vite.ssrLoadModule("/src/ssr-entry.tsx");
        // 从 Cookie 读取 token 并传入 SSR 上下文
        const token = req.cookies?.auth_token;
        const { html: appHtml } = render(url, { token });
        const html = template.replace('<div id="root"></div>', `<div id="root">${appHtml}</div>`);
        res.status(200).set({ "Content-Type": "text/html" }).end(html);
      } else {
        res.status(200).set({ "Content-Type": "text/html" }).end(template);
      }
    } catch (e) {
      vite.ssrFixStacktrace(e as Error);
      next(e);
    }
  });

  // 加载路由配置供启动时打印
  const { ssrRoutes } = await vite.ssrLoadModule("/src/router/section/ssr-routes.ts");
  const { spaRoutes } = await vite.ssrLoadModule("/src/router/section/spa-routes.ts");

  return {
    vite,
    ssrRoutes: ssrRoutes as { path: string; title: string }[],
    spaRoutes: spaRoutes as { path: string; title: string }[],
  };
}

import express from "express";

const isProduction = process.env.NODE_ENV === "production";
const port = Number(process.env.PORT) || 7878;

interface RouteInfo {
  ssrRoutes: { path: string; title: string }[];
  spaRoutes: { path: string; title: string }[];
}

async function start() {
  const app = express();

  let routes: RouteInfo;

  if (isProduction) {
    const { setupProd } = await import("./server/prod.js");
    routes = await setupProd(app);
  } else {
    const { setupDev } = await import("./server/dev.js");
    routes = await setupDev(app);
  }

  app.listen(port, () => {
    const mode = isProduction ? "production" : "development";
    console.log(`\nSSR server running at http://localhost:${port} (${mode})\n`);

    console.log("SSR 页面（服务端渲染）：");
    console.table(
      routes.ssrRoutes.map((r) => ({
        路径: r.path,
        标题: r.title,
        地址: `http://localhost:${port}${r.path}`,
      })),
    );

    console.log("SPA 页面（客户端渲染）：");
    console.table(
      routes.spaRoutes.map((r) => ({
        路径: r.path,
        标题: r.title,
        地址: `http://localhost:${port}${r.path}`,
      })),
    );
  });
}

start();

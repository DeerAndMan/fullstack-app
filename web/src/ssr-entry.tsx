import { renderToString } from "react-dom/server";
import { StaticRouter } from "react-router";
import { Routes, Route } from "react-router-dom";
import { ConfigProvider } from "antd";
import locale from "antd/locale/zh_CN";
import { ssrRoutes, ssrPaths } from "./router/section/ssr-routes";
import { spaRoutes } from "./router/section/spa-routes";
import { SSRContext, type SSRContextValue } from "./ssr-context";

// 导出路由数据供 server 使用
export { ssrRoutes, ssrPaths, spaRoutes };

export function render(url: string, context?: SSRContextValue) {
  const html = renderToString(
    <SSRContext.Provider value={context}>
      <ConfigProvider locale={locale}>
        <StaticRouter location={url}>
          <Routes>
            {ssrRoutes.map(({ path, component: Comp }) => (
              <Route key={path} path={path} element={<Comp />} />
            ))}
          </Routes>
        </StaticRouter>
      </ConfigProvider>
    </SSRContext.Provider>
  );
  return { html };
}

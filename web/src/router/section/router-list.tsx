import { Suspense, lazy, createElement } from "react";
import { Outlet } from "react-router-dom";

import Layout from "@/layouts";
import { ssrRoutes } from "./ssr-routes";

import type { RouteObject } from "react-router-dom";

export const Login = lazy(() => import("@pages/login"));
export const User = lazy(() => import("@pages/user"));
export const Operation = lazy(() => import("@pages/user/operation"));
export const HomePage = lazy(() => import("@pages/home"));
export const TradePage = lazy(() => import("@pages/trade"));
export const WsPage = lazy(() => import("@/pages/ws/ws-page"));
export const SubscribePage = lazy(() => import("@/pages/subscribe/home"));
export const SubscribeListPage = lazy(() => import("@/pages/subscribe/list"));
export const SubscribeDetailPage = lazy(() => import("@/pages/subscribe/detail"));

export const MenuPage = lazy(() => import("@/pages/role/Menu"));
export const RoleListPage = lazy(() => import("@/pages/role/list"));

const routerList: RouteObject[] = [
  {
    element: (
      <Layout>
        <Suspense fallback={<span className="loading loading-infinity loading-xl" />}>
          <Outlet />
        </Suspense>
      </Layout>
    ),

    children: [
      { path: "/", element: <HomePage /> },
      { path: "/login", element: <Login /> },
      {
        path: "/user",
        element: <Outlet />,
        children: [
          { index: true, element: <User /> },
          { path: "operation", element: <Operation /> },
        ],
      },
      { path: "/data", element: <TradePage /> },
      { path: "/ws", element: <WsPage /> },
      // SSR 页面（从 ssr-routes.ts 自动生成，支持任意路径）
      ...ssrRoutes.map(({ path, component }) => ({
        path,
        element: createElement(component),
      })),
      {
        path: "/subscribe",
        element: <Outlet />,
        children: [
          { index: true, element: <SubscribePage /> },
          { path: "list", element: <SubscribeListPage /> },
          { path: "detail/:id/:userId", element: <SubscribeDetailPage /> },
        ],
      },
      {
        path: "/role",
        element: <Outlet />,
        children: [
          {
            path: "menu",
            children: [
              { index: true, element: <MenuPage /> },
              {
                path: ":id",
                children: [{ path: "add", element: <>添加菜单</> }],
              },
            ],
          },
          {
            path: "list",
            element: <RoleListPage />,
          },
        ],
      },
    ],
  },
];

// eslint-disable-next-line react-refresh/only-export-components
export default routerList;

import { Suspense } from "react";
import { useRoutes, Outlet, Navigate } from "react-router-dom";

import routerList from "./section/router-list";

import type { RouteObject } from "react-router-dom";

const layoutRouter: RouteObject[] = [
  ...routerList,
  {
    element: (
      <div>
        <Suspense fallback={<span className="loading loading-infinity loading-xl" />}>
          <Outlet />
        </Suspense>
      </div>
    ),
    errorElement: <h2>404 Not Found</h2>,
    children: [{ path: "404", element: <h2>404 Not Found</h2> }],
  },
  { path: "*", element: <Navigate to="/404" replace /> },
];

/** 路由组件 */
export function Router() {
  return useRoutes(layoutRouter);
}

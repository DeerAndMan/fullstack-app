import { createContext, useContext } from "react";

/**
 * SSR 上下文：服务端渲染时由 Express 注入，客户端渲染时为 undefined
 */
export interface SSRContextValue {
  /** 用户认证 token，从请求 Cookie 中读取 */
  token?: string;
}

export const SSRContext = createContext<SSRContextValue | undefined>(undefined);

/**
 * 在 SSR 组件中获取服务端上下文
 *
 * @example
 * ```tsx
 * const ctx = useSSRContext();
 * // 服务端：ctx?.token 有值（从 Cookie 读取）
 * // 客户端：ctx 为 undefined（走正常 Redux token）
 * ```
 */
export function useSSRContext(): SSRContextValue | undefined {
  return useContext(SSRContext);
}

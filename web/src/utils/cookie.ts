const COOKIE_KEY = "auth_token";
const MAX_AGE = 7 * 24 * 3600; // 7 天

/**
 * 登录成功后将 token 同步写入 Cookie
 * 服务端渲染时 Express 通过 cookie-parser 读取该 Cookie
 */
export function setAuthCookie(token: string) {
  document.cookie = `${COOKIE_KEY}=${encodeURIComponent(token)}; path=/; max-age=${MAX_AGE}; SameSite=Lax`;
}

/**
 * 退出登录时清除 Cookie
 */
export function clearAuthCookie() {
  document.cookie = `${COOKIE_KEY}=; path=/; max-age=0`;
}

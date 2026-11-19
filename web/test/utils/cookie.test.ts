import { beforeEach, describe, expect, it } from "vitest";

import { clearAuthCookie, setAuthCookie } from "@/utils/cookie";

describe("auth cookie utils", () => {
  beforeEach(() => {
    Object.defineProperty(globalThis, "document", {
      configurable: true,
      value: { cookie: "" },
      writable: true,
    });
  });

  it("写入经过 URI 编码且有效期为 7 天的认证 Cookie", () => {
    setAuthCookie("token with/特殊字符");

    expect(document.cookie).toBe(
      "auth_token=token%20with%2F%E7%89%B9%E6%AE%8A%E5%AD%97%E7%AC%A6; path=/; max-age=604800; SameSite=Lax",
    );
  });

  it("通过 max-age=0 清除认证 Cookie", () => {
    clearAuthCookie();

    expect(document.cookie).toBe("auth_token=; path=/; max-age=0");
  });
});

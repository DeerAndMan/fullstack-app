import { beforeEach, describe, expect, it, vi } from "vitest";

const storageState = vi.hoisted(() => {
  const values = new Map<string, string>();
  const storage: Storage = {
    get length() {
      return values.size;
    },
    clear: () => values.clear(),
    getItem: key => values.get(key) ?? null,
    key: index => Array.from(values.keys())[index] ?? null,
    removeItem: key => values.delete(key),
    setItem: (key, value) => values.set(key, value),
  };
  Object.defineProperty(globalThis, "localStorage", {
    configurable: true,
    value: storage,
  });
  return { storage };
});

const apiMocks = vi.hoisted(() => ({
  logout: vi.fn(),
  getRoleList: vi.fn(),
  clearAuthCookie: vi.fn(),
}));

vi.mock("../../src/api/user/index.ts", () => ({ logout: apiMocks.logout }));
vi.mock("../../src/api/enum/index.ts", () => ({ getRoleList: apiMocks.getRoleList }));
vi.mock("../../src/utils/cookie.ts", () => ({
  clearAuthCookie: apiMocks.clearAuthCookie,
  setAuthCookie: vi.fn(),
}));

import { useAuthStore } from "@/stores/auth";
import { useEnumStore } from "@/stores/enum";
import { useGlobalStore } from "@/stores/global";
import { RoleKey } from "@/types/enum";

import type { Account, Role } from "@/types/user";
import type { MenuItemType } from "@/types/menu-router";

const account: Account = {
  age: 30,
  created_at: "2026-01-01",
  description: "test user",
  email: "alice@example.com",
  id: 1,
  name: "Alice",
  password: "",
  updated_at: "2026-01-02",
  addTime: "2026-01-01",
  avatar: "old.png",
  roleKey: RoleKey.ADMIN,
  role: null,
};

const role: Role = {
  role_id: 2,
  role_name: "管理员",
  role_key: "admin",
  sort: 1,
  role_status: 1,
  create_by: "system",
  create_time: "2026-01-01",
  update_by: "system",
  update_time: "2026-01-01",
  del_flag: 0,
};

const menuRole = {
  id: "1",
  name: "首页",
  link_url: "/",
  menu_code: "home",
  parent_id: "0",
  node_type: 1,
  icon_url: "",
  level: 1,
  path: "",
  is_delete: 0,
} satisfies MenuItemType;

describe("zustand stores", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    storageState.storage.clear();
    useAuthStore.setState({
      token: "",
      user: null,
      role: null,
      menuRoles: null,
      loading: false,
    });
    useEnumStore.setState({ roleList: [] });
    useGlobalStore.setState({ messageApi: null });
  });

  describe("auth store", () => {
    it("设置 token、用户、角色和菜单权限", () => {
      const state = useAuthStore.getState();

      state.setToken("access-token");
      state.setUser(account);
      state.setRole(role);
      state.setMenuRoles([menuRole]);

      expect(useAuthStore.getState()).toMatchObject({
        token: "access-token",
        user: account,
        role,
        menuRoles: [menuRole],
      });
    });

    it("只在用户存在时更新头像", () => {
      useAuthStore.getState().setUserAvatar("ignored.png");
      expect(useAuthStore.getState().user).toBeNull();

      useAuthStore.getState().setUser(account);
      useAuthStore.getState().setUserAvatar("new.png");

      expect(useAuthStore.getState().user).toEqual({ ...account, avatar: "new.png" });
    });

    it("clearToken 清理认证 Cookie 和全部认证状态", () => {
      useAuthStore.setState({
        token: "token",
        user: account,
        role,
        menuRoles: [menuRole],
      });

      useAuthStore.getState().clearToken();

      expect(apiMocks.clearAuthCookie).toHaveBeenCalledOnce();
      expect(useAuthStore.getState()).toMatchObject({
        token: "",
        user: null,
        role: null,
        menuRoles: null,
      });
    });

    it("退出成功后清空认证状态并恢复 loading", async () => {
      apiMocks.logout.mockResolvedValue({ code: 0, data: null, message: "ok" });
      useAuthStore.setState({ token: "token", user: account, role, menuRoles: [menuRole] });

      await expect(useAuthStore.getState().storeLogout()).resolves.toBe(true);

      expect(apiMocks.clearAuthCookie).toHaveBeenCalledOnce();
      expect(useAuthStore.getState()).toMatchObject({
        token: "",
        user: null,
        role: null,
        menuRoles: null,
        loading: false,
      });
    });

    it("退出接口返回业务失败时保留认证状态", async () => {
      apiMocks.logout.mockResolvedValue({ code: 1, message: "failed" });
      useAuthStore.setState({ token: "token", user: account });

      await expect(useAuthStore.getState().storeLogout()).resolves.toBe(false);

      expect(apiMocks.clearAuthCookie).not.toHaveBeenCalled();
      expect(useAuthStore.getState()).toMatchObject({
        token: "token",
        user: account,
        loading: false,
      });
    });

    it("退出请求异常时返回 false 并恢复 loading", async () => {
      apiMocks.logout.mockRejectedValue(new Error("network error"));

      await expect(useAuthStore.getState().storeLogout()).resolves.toBe(false);
      expect(useAuthStore.getState().loading).toBe(false);
    });
  });

  describe("enum store", () => {
    it("设置和获取角色枚举", () => {
      const roles = [{ id: 1, name: "管理员", role_key: RoleKey.ADMIN }];

      useEnumStore.getState().setRoleList(roles);

      expect(useEnumStore.getState().roleList).toEqual(roles);
    });

    it("成功获取角色列表", async () => {
      const roles = [{ id: 1, name: "管理员", role_key: RoleKey.ADMIN }];
      apiMocks.getRoleList.mockResolvedValue({ code: 0, data: roles });

      await useEnumStore.getState().fetchRoles();

      expect(useEnumStore.getState().roleList).toEqual(roles);
    });

    it.each([
      ["业务失败", { code: 1, data: [{ id: 1 }] }],
      ["请求异常", new Error("network error")],
    ])("%s 时清空角色列表", async (_name, result) => {
      useEnumStore.setState({
        roleList: [{ id: 1, name: "旧角色", role_key: RoleKey.USER }],
      });
      if (result instanceof Error) {
        apiMocks.getRoleList.mockRejectedValue(result);
      } else {
        apiMocks.getRoleList.mockResolvedValue(result);
      }

      await useEnumStore.getState().fetchRoles();

      expect(useEnumStore.getState().roleList).toEqual([]);
    });
  });

  it("global store 保存 Ant Design message 实例", () => {
    const messageApi = { success: vi.fn(), error: vi.fn() } as never;

    useGlobalStore.getState().setMessageApi(messageApi);

    expect(useGlobalStore.getState().messageApi).toBe(messageApi);
  });
});

import { beforeEach, describe, expect, it, vi } from "vitest";

const requestMock = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}));

const requestSchemaMock = vi.hoisted(() => vi.fn());

vi.mock("../../src/api/index.ts", () => ({
  request: requestMock,
  RequestSchema: requestSchemaMock,
}));

vi.mock("../../src/api/request.ts", () => ({
  default: requestMock,
}));

import { login as authLogin, refreshToken, register } from "@/api/auth";
import { getRoleList } from "@/api/enum/enum.service";
import {
  addMenuRouter,
  addRoleRouting,
  getMenuRouterList,
  getRoleRoutingByRoleId,
} from "@/api/menu/menu.service";
import { getSummary, getTrade } from "@/api/trade/trade.service";
import {
  addUser,
  deleteUser,
  getAllUser,
  getProfile,
  login,
  logout,
  updateUser,
  updateUserRole,
  uploadAvatar,
} from "@/api/user/user.service";
import { apiPathsV1 } from "@/api/paths/v1";

describe("API services", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("auth API 使用预期路径和参数", () => {
    const credentials = { username: "alice", password: "secret" };

    authLogin(credentials);
    register({ ...credentials, nickname: "Alice" });
    refreshToken({ refresh_token: "refresh-token" });

    expect(requestMock.post).toHaveBeenNthCalledWith(1, apiPathsV1.auth.login, credentials);
    expect(requestMock.post).toHaveBeenNthCalledWith(2, apiPathsV1.auth.register, {
      ...credentials,
      nickname: "Alice",
    });
    expect(requestMock.post).toHaveBeenNthCalledWith(3, apiPathsV1.auth.refreshToken, {
      refresh_token: "refresh-token",
    });
  });

  it("user service 正确拼接资源路径并转发配置", () => {
    const loginConfig = { noToken: true } as never;
    const addParams = {
      age: 20,
      description: "desc",
      email: "a@example.com",
      name: "Alice",
      password: "secret",
    };

    login({ name: "Alice", password: "secret" }, loginConfig);
    logout();
    getAllUser();
    getProfile();
    deleteUser(7);
    addUser(addParams);
    updateUser({ ...addParams, id: 7 });
    updateUserRole({ user_id: 7, role_id: 2 });

    expect(requestMock.post).toHaveBeenNthCalledWith(
      1,
      apiPathsV1.auth.login,
      { name: "Alice", password: "secret" },
      { noToken: true },
    );
    expect(requestMock.post).toHaveBeenNthCalledWith(2, apiPathsV1.auth.logout);
    expect(requestMock.get).toHaveBeenNthCalledWith(1, apiPathsV1.user.list);
    expect(requestMock.get).toHaveBeenNthCalledWith(2, apiPathsV1.user.profile);
    expect(requestMock.delete).toHaveBeenCalledWith(`${apiPathsV1.user.detail}/7`);
    expect(requestMock.post).toHaveBeenCalledWith(apiPathsV1.user.list, addParams);
    expect(requestMock.put).toHaveBeenCalledWith(`${apiPathsV1.user.detail}/7`, {
      ...addParams,
      id: 7,
    });
    expect(requestMock.put).toHaveBeenCalledWith(`${apiPathsV1.user.detail}/7`, {
      user_id: 7,
      role_id: 2,
    });
  });

  it("头像上传使用 multipart/form-data", () => {
    const formData = new FormData();
    formData.append("avatar", new Blob(["avatar"]), "avatar.txt");

    uploadAvatar(formData);

    expect(requestMock.post).toHaveBeenCalledWith(
      apiPathsV1.upload,
      formData,
      expect.objectContaining({
        headers: { "Content-Type": "multipart/form-data" },
      }),
    );
  });

  it("menu 和 enum service 使用统一路径表", () => {
    const menuParams = [{ key: "menu" }] as never;
    const roleRouting = { roleId: 2, menuIds: ["1", "2"] };

    getRoleList();
    getMenuRouterList();
    addMenuRouter(menuParams);
    addRoleRouting(roleRouting);
    getRoleRoutingByRoleId(2);

    expect(requestMock.get).toHaveBeenCalledWith(apiPathsV1.role.all);
    expect(requestMock.get).toHaveBeenCalledWith(apiPathsV1.menu.list);
    expect(requestMock.post).toHaveBeenCalledWith(apiPathsV1.menu.add, menuParams);
    expect(requestMock.post).toHaveBeenCalledWith(apiPathsV1.menu.roleBinding, roleRouting);
    expect(requestMock.get).toHaveBeenCalledWith(`${apiPathsV1.menu.roleListByRoleId}/2`);
  });

  it("trade service 分别调用 Schema 请求和普通请求", () => {
    const params = { startTime: "2026-08-01", endTime: "2026-08-27" };

    getTrade(params);
    getSummary(params);

    expect(requestSchemaMock).toHaveBeenCalledWith(
      expect.objectContaining({
        method: "POST",
        url: apiPathsV1.trade.index,
        data: params,
      }),
    );
    expect(requestMock.post).toHaveBeenCalledWith(apiPathsV1.trade.summary, params);
  });
});

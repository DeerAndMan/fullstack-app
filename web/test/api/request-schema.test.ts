import { beforeEach, describe, expect, it, vi } from "vitest";
import { z } from "zod";

const requestMock = vi.hoisted(() => {
  const request = vi.fn();
  return {
    request,
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  };
});

vi.mock("../../src/api/request.ts", () => {
  Object.assign(requestMock.request, {
    get: requestMock.get,
    post: requestMock.post,
    put: requestMock.put,
    delete: requestMock.delete,
  });
  return { default: requestMock.request };
});

import {
  BaseResponseSchema,
  RequestDelete,
  RequestGet,
  RequestPost,
  RequestPut,
  RequestSchema,
} from "@/api/request-schema";

describe("request-schema", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("基础响应 Schema 校验 code 和 message", () => {
    expect(BaseResponseSchema.parse({ code: 0, message: "ok" })).toEqual({
      code: 0,
      message: "ok",
    });
    expect(() => BaseResponseSchema.parse({ code: "0", message: "ok" })).toThrow();
  });

  it("RequestSchema 合并基础响应结构并返回 data/message", async () => {
    requestMock.request.mockResolvedValue({
      code: 0,
      message: "success",
      data: { id: 1, name: "Alice" },
    });

    const result = await RequestSchema({
      method: "GET",
      url: "/users/1",
      schema: z.object({ id: z.number(), name: z.string() }),
    } as never);

    expect(result).toEqual({
      data: { id: 1, name: "Alice" },
      message: "success",
    });
    expect(requestMock.request).toHaveBeenCalledOnce();

    const forwardedConfig = requestMock.request.mock.calls[0][0];
    expect(forwardedConfig).not.toHaveProperty("schemaType");
    expect(
      forwardedConfig.schema.safeParse({
        code: 0,
        message: "success",
        data: { id: 1, name: "Alice" },
      }).success,
    ).toBe(true);
  });

  it("支持 table 类型的分页响应结构", async () => {
    const response = {
      code: 0,
      message: "ok",
      data: {
        pageNumber: 1,
        pageSize: 20,
        totalPage: 2,
        totalCount: 21,
        list: [{ id: 1 }, { id: 2 }],
      },
    };
    requestMock.get.mockResolvedValue(response);

    await expect(
      RequestGet("/items", {
        schema: z.object({ id: z.number() }),
        schemaType: "table",
      } as never),
    ).resolves.toEqual({ data: response.data, message: "ok" });

    const config = requestMock.get.mock.calls[0][1];
    expect(config).not.toHaveProperty("schemaType");
    expect(config.schema.safeParse(response).success).toBe(true);
  });

  it("响应业务 code 非 0 时抛出包含服务端 message 的错误", async () => {
    requestMock.request.mockResolvedValue({ code: 1001, message: "业务失败", data: null });

    await expect(
      RequestSchema({
        method: "GET",
        url: "/failed",
        schema: z.null(),
      } as never),
    ).rejects.toThrow("业务失败");
  });

  it("响应不符合 Schema 时输出警告并抛出 ZodError", async () => {
    const warn = vi.spyOn(console, "warn").mockImplementation(() => undefined);
    requestMock.request.mockResolvedValue({
      code: 0,
      message: "ok",
      data: { id: "not-a-number" },
    });

    await expect(
      RequestSchema({
        method: "GET",
        url: "/invalid",
        schema: z.object({ id: z.number() }),
      } as never),
    ).rejects.toBeInstanceOf(z.ZodError);

    expect(warn).toHaveBeenCalledWith(expect.stringContaining("响应数据验证失败"));
    expect(warn).toHaveBeenCalledWith(expect.stringContaining("data.id"));
    warn.mockRestore();
  });

  it("GET/POST/PUT/DELETE 包装器转发参数", async () => {
    const response = { code: 0, message: "ok", data: { id: 1 } };
    requestMock.get.mockResolvedValue(response);
    requestMock.post.mockResolvedValue(response);
    requestMock.put.mockResolvedValue(response);
    requestMock.delete.mockResolvedValue(response);
    const config = { schema: z.object({ id: z.number() }), noToken: true } as never;

    await RequestGet("/resource", config);
    await RequestPost("/resource", { name: "new" }, config);
    await RequestPut("/resource/1", { name: "updated" }, config);
    await RequestDelete("/resource/1", config);

    expect(requestMock.get).toHaveBeenCalledWith(
      "/resource",
      expect.objectContaining({ noToken: true, schema: expect.anything() }),
    );
    expect(requestMock.post).toHaveBeenCalledWith(
      "/resource",
      { name: "new" },
      expect.objectContaining({ noToken: true, schema: expect.anything() }),
    );
    expect(requestMock.put).toHaveBeenCalledWith(
      "/resource/1",
      { name: "updated" },
      expect.objectContaining({ noToken: true, schema: expect.anything() }),
    );
    expect(requestMock.delete).toHaveBeenCalledWith(
      "/resource/1",
      expect.objectContaining({ noToken: true, schema: expect.anything() }),
    );
  });
});

import { beforeEach, describe, expect, it, vi } from "vitest";

const apiMock = vi.hoisted(() => ({ baseApi: "https://api.example.com" }));

vi.mock("../../src/api/index.ts", () => ({
  get BASE_API() {
    return apiMock.baseApi;
  },
}));

class FakeWebSocket {
  static instances: FakeWebSocket[] = [];

  url: string;
  onopen: ((event?: Event) => void) | null = null;
  onerror: ((event?: Event) => void) | null = null;
  onclose: ((event?: CloseEvent) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;

  constructor(url: string) {
    this.url = url;
    FakeWebSocket.instances.push(this);
  }
}

describe("wsInit", () => {
  beforeEach(() => {
    FakeWebSocket.instances = [];
    apiMock.baseApi = "https://api.example.com";
    vi.stubGlobal("WebSocket", FakeWebSocket);
  });

  it("将 HTTPS API 地址转换为 WSS 地址并注册事件处理器", async () => {
    const { wsInit } = await import("@/utils/websocket");
    const client = new wsInit("/socket");
    const socket = FakeWebSocket.instances[0];

    expect(socket.url).toBe("wss://api.example.com/socket");
    expect(socket.onopen).toBe(client.onopen);
    expect(socket.onerror).toBe(client.onerror);
    expect(socket.onclose).toBe(client.onclose);
    expect(socket.onmessage).toBe(client.onmessage);
  });

  it("解析消息 JSON 并交给回调", async () => {
    const onMessage = vi.fn();
    const { wsInit } = await import("@/utils/websocket");
    const client = new wsInit<{ id: number }>("/events", { onMessage });

    client.onmessage({ data: '{"id":42}' } as MessageEvent);

    expect(onMessage).toHaveBeenCalledWith({ id: 42 });
  });

  it("支持 HTTP 和空 API 地址", async () => {
    const { wsInit } = await import("@/utils/websocket");

    apiMock.baseApi = "http://localhost:8080";
    const httpClient = new wsInit("/ws");
    expect(httpClient.wsUrl).toBe("ws://localhost:8080");

    apiMock.baseApi = "";
    const emptyClient = new wsInit("/ws");
    expect(emptyClient.wsUrl).toBe("");
    expect(FakeWebSocket.instances.at(-1)?.url).toBe("/ws");
  });
});

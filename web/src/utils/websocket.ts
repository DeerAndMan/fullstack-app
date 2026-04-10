import { BASE_API } from "@/api";

type WsType<T> = Partial<WebSocket> & {
  path: string;
  wsUrl: string;
  wsFunc?: WsFunc<T>;
};

export type WsFunc<T> = {
  onMessage: (data: MessageEvent["data"] & T) => void;
};

// 定义类型
export class wsInit<T> implements WsType<T> {
  path: string;
  ws: WebSocket = {} as WebSocket;
  wsFunc?: WsFunc<T>;

  constructor(path: string, wsFunc?: WsFunc<T>) {
    this.path = path;
    this.wsFunc = wsFunc;
    this.init();
  }

  init() {
    this.ws = new WebSocket(this.wsUrl + this.path);
    this.ws.onopen = this.onopen;
    this.ws.onerror = this.onerror;
    this.ws.onclose = this.onclose;
    this.ws.onmessage = this.onmessage;
  }

  onopen = () => {
    console.log("连接成功 ！！！");
  };

  onerror = () => {
    console.log("连接失败 ！！！");
  };

  onclose = () => {
    console.log("连接关闭 ！！！");
  };

  onmessage = (d: MessageEvent) => {
    this.wsFunc?.onMessage(JSON.parse(d.data));
  };

  get wsUrl() {
    if (!BASE_API) return "";
    return BASE_API.replace("http://", "ws://").replace("https://", "wss://");
  }
}

import { useEffect, useMemo, useState } from "react";
import dayjs from "dayjs";

import { apiPathsV2 } from "@/api/paths/v2";
import { wsInit } from "@/utils/websocket";

import type { WsFunc } from "@/utils/websocket";

type DType = {
  content: string;
  timestamp: Date;
  type: string;
};

/**
 * ws-page 页面 长连接测试
 * @returns {React.FunctionComponent}
 */
export const WsPage = () => {
  const [wsData, setWsData] = useState<DType[]>([]);

  const onMessage: WsFunc<DType>["onMessage"] = data => {
    setWsData(pre => [data, ...pre]);
  };

  const ws = useMemo(() => new wsInit<DType[]>(apiPathsV2.ws.conversations, { onMessage }).ws, []);

  useEffect(() => {
    return () => {
      ws.close();
    };
  }, []);

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">WebSocket 测试</h1>
      <div className="space-y-2">
        {wsData.length === 0 ? (
          <div className="text-gray-400">等待消息...</div>
        ) : (
          wsData.map((item, index) => (
            <div key={index} className="border border-gray-200 dark:border-gray-700 rounded p-3">
              <div className="text-sm text-gray-500 mb-1">{dayjs(item.timestamp).format("YYYY-MM-DD HH:mm:ss")}</div>
              <div className="font-medium">{item.content}</div>
              {item.type && <div className="text-xs text-gray-400 mt-1">类型: {item.type}</div>}
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default WsPage;

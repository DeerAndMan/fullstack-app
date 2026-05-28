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
  const [wsData, setWsData] = useState<DType | null>(null);

  const onMessage: WsFunc<DType>["onMessage"] = data => {
    setWsData(data);
  };

  const ws = useMemo(() => new wsInit<DType>(apiPathsV2.ws.conversations, { onMessage }).ws, []);

  useEffect(() => {
    return () => {
      ws.close();
    };
  }, []);

  return (
    <div>
      WsPage
      <div>{wsData?.content}</div>
      <div>{dayjs(wsData?.timestamp).format("YYYY-MM-DD HH:mm:ss")}</div>
    </div>
  );
};

export default WsPage;

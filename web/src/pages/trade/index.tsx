import { Card, Tabs } from "antd";

import RawDataTab from "./raw-data-tab";
import StatsTab from "./stats-tab";

/**
 * 交易数据页面：原始推送数据 + 周度/月度/半年度/年度汇总统计
 */
export default function Trade() {
  return (
    <Card>
      <Tabs
        defaultActiveKey="raw"
        items={[
          { key: "raw", label: "原始数据", children: <RawDataTab /> },
          { key: "stats", label: "统计数据", children: <StatsTab /> },
        ]}
      />
    </Card>
  );
}

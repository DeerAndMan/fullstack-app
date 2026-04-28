import { useCallback, useEffect, useState } from "react";
import { Button, Space, Switch, DatePicker, Card } from "antd";

import dayjs from "dayjs";
import { DualAxesChart, LineChart } from "@/components";
import { tradeListQuery } from "@/api/trade";

import type { RangePickerProps } from "antd/es/date-picker";
import type { EnergyItem } from "./type";
import { useGlobalStore } from "@/stores/global";

interface LineItem {
  time: string;
  value: number;
  proportion: number;
}
interface LineTypeItem extends Omit<LineItem, "proportion"> {
  type: string;
  total: EnergyItem;
}

const { RangePicker } = DatePicker;

export default function Trade() {
  const messageApi = useGlobalStore(s => s.messageApi);

  const [lineData, setLineData] = useState<LineItem[]>([]);
  const [lineTypeData, setLineTypeData] = useState<LineTypeItem[]>([]);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [closeNum, setCloseNum] = useState(false);

  const { queryList, stateOperations, operations } = tradeListQuery();

  // 梳理数据
  const combingData = useCallback(() => {
    if (!queryList.data || queryList.data.data.length === 0) return;
    // console.log("queryList", queryList.data.data);
    const list: LineItem[] = [];
    const listType: LineTypeItem[] = [];

    queryList.data.data.forEach(l => {
      const time = dayjs(l.date).format("YYYY-MM-DD HH:mm:ss");
      list.push({ time, value: Number(l.dryk), proportion: Number(l.drhz) * 100 });

      l.positions.forEach(p => {
        listType.push({ time, type: p.Zqmc, value: Number(p.Dryk), total: p });
      });
    });

    setLineData(list);
    setLineTypeData(listType);
  }, [queryList.data]);

  const pickerChange = (val: RangePickerProps["value"]) => {
    if (val instanceof Array) {
      stateOperations.setDateTime([dayjs(val[0]), dayjs(val[1])]);
    }
  };

  const toggleAutoRefresh = (checked: boolean) => {
    setAutoRefresh(checked);
  };

  useEffect(() => {
    combingData();
  }, [combingData]);

  useEffect(() => {
    let newTimer: ReturnType<typeof setInterval> | null = null;
    if (autoRefresh) {
      newTimer = setInterval(
        () => {
          const now = dayjs();
          const minutes = now.hour() * 60 + now.minute();

          // 早上 9:15 到 11:50 (555 - 710 分钟)
          const isMorning = minutes >= 555 && minutes <= 710;
          // 下午 12:40 到 15:20 (760 - 920 分钟)
          const isAfternoon = minutes >= 760 && minutes <= 920;

          if (isMorning || isAfternoon) {
            operations.refresh();
          } else {
            messageApi?.warning("自动刷新时间在 早上9:15到11:50 及 下午12:40到15:20 之间");
            console.error("自动刷新已关闭!!!");
            setAutoRefresh(false);
          }
        },
        1000 * 60 * 3
      );
    } else {
      messageApi?.warning("自动刷新已关闭");
      if (newTimer) clearInterval(newTimer);
    }

    return () => {
      if (newTimer) {
        clearInterval(newTimer);
      }
    };
  }, [messageApi, autoRefresh, operations]);

  return (
    <Card>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">数据{lineData[lineData.length - 1]?.value}</h2>
        <Space size="middle">
          <RangePicker
            maxDate={dayjs()}
            defaultValue={[stateOperations.dateTime[0], stateOperations.dateTime[1]]}
            format={"YYYY-MM-DD"}
            onChange={pickerChange}
          />
          <Button type="primary" onClick={operations.refresh} loading={queryList.isLoading}>
            立即刷新
          </Button>

          <Switch
            checked={autoRefresh}
            onChange={toggleAutoRefresh}
            checkedChildren="刷新开启"
            unCheckedChildren="刷新关闭"
          />

          <Switch
            checked={!closeNum}
            onChange={() => setCloseNum(!closeNum)}
            checkedChildren="显示"
            unCheckedChildren="隐藏"
          />
        </Space>
      </div>

      <DualAxesChart data={lineData} closeNum={closeNum} />
      <LineChart data={lineTypeData} />
    </Card>
  );
}

import { DualAxes } from "@ant-design/charts";
import { useMemo } from "react";

interface DualAxesDataItem {
  time: string;
  value: number;
  proportion: number;
}

export interface DualAxesProps {
  data: DualAxesDataItem[];
  closeNum: boolean;
}

export default function DualAxesChart(props: DualAxesProps) {
  const { data, closeNum } = props;

  const isLast: DualAxesDataItem = useMemo(() => {
    return data[data.length - 1];
  }, [data]);

  const valueMixMax = useMemo(() => {
    if (!data.length || data.length <= 2) return { minTime: "", maxTime: "" };
    const list = data.map(item => item.value).filter(l => l);

    const min = Math.min(...list);
    const max = Math.max(...list);

    const minTime = data.find(item => item.value === min)?.time ?? "";
    const maxTime = data.findLast(item => item.value === max)?.time ?? "";

    return { minTime, maxTime };
  }, [data]);

  const proportionMinMax = useMemo(() => {
    if (!data.length || data.length <= 2) return { minTime: "", maxTime: "" };
    const list = data.map(item => item.proportion).filter(l => l);

    const min = Math.min(...list);
    const max = Math.max(...list);

    const minTime = data.find(item => item.proportion === min)?.time ?? "";
    const maxTime = data.find(item => item.proportion === max)?.time ?? "";

    return { minTime, maxTime };
  }, [data]);

  const config = {
    data: data,
    xField: "time",
    legend: true,
    slider: { x: {} },
    children: [
      {
        type: "line",
        yField: "value",
        style: { stroke: "#5B8FF9", lineWidth: 2 },
        label: {
          text: (datum: DualAxesDataItem) => {
            if (closeNum) return "";
            if (valueMixMax.minTime === datum.time || valueMixMax.maxTime === datum.time) {
              return datum.value;
            }
            if (isLast.time !== datum.time) return "";

            return datum.value;
          },
          style: { dy: 15, textAlign: "middle" },
        },
        colorField: "type",
        axis: {
          y: {
            position: "left",
            title: "汇总",
            style: { titleFill: "#5B8FF9" },
          },
        },
      },
      {
        type: "line",
        yField: "proportion",
        style: { stroke: "#E74C3C", lineWidth: 2, lineDash: [5, 5] },
        itemStyle: { fill: "#E74C3C" },
        label: {
          text: (datum: DualAxesDataItem) => {
            if (
              proportionMinMax.minTime === datum.time ||
              proportionMinMax.maxTime === datum.time
            ) {
              return `${datum.proportion.toFixed(2)}%`;
            }
            if (isLast.time !== datum.time) return "";

            return `${datum.proportion.toFixed(2)}%`;
          },
          style: { dy: -15, textAlign: "middle" },
        },
        axis: {
          y: {
            position: "right",
            title: "比例",
            style: { titleFill: "#E74C3C" },
          },
        },
      },
    ],
  };

  return <DualAxes {...config} />;
}

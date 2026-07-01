import { Line } from "@ant-design/charts";

import type { EnergyItem } from "@/types/schema";
import type { LineConfig, Tooltip } from "@ant-design/charts";
import { formatPercent } from "@/utils/number";

interface LineDataItem {
  time: string;
  value: number;
  type: string;
  total: EnergyItem;
}

type ToolTipItem = { title: string; items: { name: string; color: string; value: string }[] };

export interface LineProps {
  data: LineDataItem[];
}

// const getCssVar = (name: string) => {
//   return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
// };
// const colorPrimary = getCssVar("--app-color-primary");

export default function LineChart(props: LineProps) {
  const { data } = props;

  const config: LineConfig = {
    data,
    xField: "time",
    yField: "value",
    colorField: "type",
    point: { shapeField: "triangle", sizeField: 4 },
    slider: { x: {} },
    legend: {
      color: {
        // itemMarker: "plus",
        // itemMarkerStroke: "text-primary",
        // itemMarkerFill: "#bf3232",
        // itemMarkerFillOpacity: 0.9,
        // itemMarkerShadowColor: "#d3d3d3",
      },
    },
    style: {
      lineWidth: 2,
      lineDash: (items: LineDataItem[]) => {
        const { value } = items[items.length - 1];
        return value < 0 ? [2, 4] : [0, 0];
      },
    },
    interaction: {
      tooltip: {
        marker: false,
        // shared: true,
        // mount: "body",
        render: (_e: HTMLElement, { title, items }: ToolTipItem) => {
          const filterMap = new Map<string, EnergyItem>();
          data.forEach(d => {
            if (d.time === title) {
              filterMap.set(title + " " + d.type, d.total);
            }
          });

          return (
            <div key={title}>
              <h4>{title}</h4>
              {items.map((item, index) => {
                const { name, value, color } = item;
                const nowData = filterMap.get(title + " " + name);
                return (
                  <div key={index}>
                    <div
                      style={{ margin: 0, display: "flex", justifyContent: "space-around", gap: 6 }}
                    >
                      <div>
                        <span
                          style={{
                            display: "inline-block",
                            width: 6,
                            height: 6,
                            borderRadius: "50%",
                            backgroundColor: color,
                            marginRight: 6,
                          }}
                        ></span>
                        <span>{name}</span>
                        {/* <span>{}</span> */}
                      </div>
                      <b>{value}</b>

                      <b>
                        【{nowData?.Zxjg}*{nowData?.Zqsl}={nowData?.Zxsz}】
                      </b>
                      <b>{formatPercent(nowData?.Drykbl)}</b>
                      <b>{nowData?.Gfssmmce || 0}</b>
                    </div>
                  </div>
                );
              })}
            </div>
          );
        },
      } as Tooltip,
    },
  };

  return (
    <div>
      <Line {...config} />
    </div>
  );
}

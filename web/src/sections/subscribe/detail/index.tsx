import { useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Button, Card, Descriptions, Table, Tag, Tooltip } from "antd";
import { ArrowLeftOutlined } from "@ant-design/icons";
import dayjs from "dayjs";

import { useQuerySubscribeDetail } from "@/hooks/query/subscribe";
import type { ThemeContentItem } from "@/types/xq/subscribe/home";

function stripHtml(html: string): string {
  const div = document.createElement("div");
  div.innerHTML = html;
  return div.textContent || div.innerText || "";
}

const KEYWORD_LABEL_MAP: Record<string, string> = {
  ip_location: "IP属地",
  post_source: "来源",
  stockList: "关联股票",
  aiStockList: "AI股票",
  stockCorrelation: "股票关联",
  post_position: "发布位置",
};

const KEYWORD_DISPLAY_KEYS = ["ip_location", "post_source", "stockList", "aiStockList"];

/** 去掉股票代码后缀的数字，如 SZ300750_11 -> SZ300750 */
function cleanStockCode(code: string): string {
  return code.replace(/_\d+$/, "");
}

function formatStockValue(val: string): string {
  const codes = val
    .split(",")
    .map(s => cleanStockCode(s.trim()))
    .filter(Boolean);
  if (codes.length <= 3) return codes.join(", ");
  return `${codes.slice(0, 3).join(", ")} 等${codes.length}只`;
}

function parseKeywords(raw: string): { label: string; value: string; full?: string }[] {
  if (!raw) return [];
  try {
    const obj = JSON.parse(raw);
    if (typeof obj === "object" && obj !== null && !Array.isArray(obj)) {
      return KEYWORD_DISPLAY_KEYS.filter(
        k => obj[k] !== undefined && obj[k] !== null && obj[k] !== ""
      ).map(k => {
        const rawVal = String(obj[k]);
        const isStock = k === "stockList" || k === "aiStockList";
        const displayVal = isStock ? formatStockValue(rawVal) : rawVal;
        const fullVal = isStock
          ? rawVal
              .split(",")
              .map(s => cleanStockCode(s.trim()))
              .join(", ")
          : undefined;
        return {
          label: KEYWORD_LABEL_MAP[k] || k,
          value: displayVal,
          full: fullVal !== displayVal ? fullVal : undefined,
        };
      });
    }
  } catch {
    // not JSON
  }
  return raw
    .split(",")
    .map(s => s.trim())
    .filter(Boolean)
    .map(s => ({ label: "", value: s }));
}

export default function SubscribeDetail() {
  const { id, userId } = useParams<{ id: string; userId: string }>();
  const navigate = useNavigate();
  const { queryDetail, queryTable, pageNumber, setPageNumber, pageSize, setPageSize } =
    useQuerySubscribeDetail(id || "", userId || "");

  const subscription = queryDetail.data?.data;
  const table = queryTable.data?.data;

  const columns = useMemo(
    () => [
      {
        title: "昵称",
        dataIndex: "screen_name",
        key: "screen_name",
        width: 140,
      },
      {
        title: "内容摘要",
        dataIndex: "text",
        key: "text",
        ellipsis: { showTitle: false },
        render: (text: string) => {
          const plain = text ? stripHtml(text) : "";
          const summary = plain.length > 80 ? `${plain.slice(0, 80)}...` : plain;
          return (
            <Tooltip title={plain} placement="topLeft">
              <span>{summary || "-"}</span>
            </Tooltip>
          );
        },
      },
      {
        title: "创建时间",
        dataIndex: "created_at",
        key: "created_at",
        width: 170,
        render: (text: string) => {
          if (!text) return "-";
          const d = dayjs(text);
          return d.isValid() ? d.format("YYYY-MM-DD HH:mm:ss") : text;
        },
      },
      {
        title: "关键词",
        dataIndex: "meta_keywords",
        key: "meta_keywords",
        width: 200,
        render: (keywords: string) => {
          const tags = parseKeywords(keywords);
          if (!tags.length) return "-";
          return (
            <div className="flex flex-wrap gap-1">
              {tags.map((item, i) => {
                const text = item.label ? `${item.label}: ${item.value}` : item.value;
                return item.full ? (
                  <Tooltip key={i} title={`${item.label}: ${item.full}`}>
                    <Tag
                      color="blue"
                      style={{
                        maxWidth: 200,
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {text}
                    </Tag>
                  </Tooltip>
                ) : (
                  <Tag key={i} color="blue">
                    {text}
                  </Tag>
                );
              })}
            </div>
          );
        },
      },
    ],
    []
  );

  return (
    <div className="flex flex-col gap-4">
      <div>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/subscribe")}>
          返回
        </Button>
      </div>

      <Card title="订阅信息" loading={queryDetail.isLoading} size="small">
        {subscription && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="ID">{subscription.id}</Descriptions.Item>
            <Descriptions.Item label="用户ID">{subscription.user_id}</Descriptions.Item>
            <Descriptions.Item label="曾用名" span={2}>
              {subscription.description || "-"}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={subscription.enabled ? "green" : "red"}>
                {subscription.enabled ? "启用" : "禁用"}
              </Tag>
            </Descriptions.Item>
          </Descriptions>
        )}
      </Card>

      <Card title={`帖子列表（共 ${table?.totalCount ?? 0} 条）`} size="small">
        <Table
          size="small"
          bordered
          loading={queryTable.isLoading}
          dataSource={table?.list || []}
          columns={columns}
          rowKey={(record: ThemeContentItem) => `${record.id}-${record.user_id}`}
          scroll={{ x: 830 }}
          pagination={{
            current: pageNumber,
            total: table?.totalCount || 0,
            pageSize,
            showTotal: total => `共 ${total} 条`,
            showSizeChanger: true,
            onChange: (p, ps) => {
              setPageNumber(p);
              setPageSize(ps);
            },
          }}
        />
      </Card>
    </div>
  );
}

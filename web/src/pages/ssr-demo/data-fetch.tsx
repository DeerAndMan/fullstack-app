import { useState } from "react";
import { Button, Card, List, Tag, Typography, Divider, Alert, Descriptions } from "antd";
import {
  ApiOutlined,
  CheckCircleOutlined,
  SyncOutlined,
  SafetyCertificateOutlined,
} from "@ant-design/icons";
import { useSSRContext } from "@/ssr-context";

const { Title, Paragraph, Text } = Typography;

interface FetchResult {
  key: string;
  label: string;
  status: "idle" | "loading" | "success" | "error";
  data?: string;
}

const serverRenderTime = new Date().toISOString();
const isServer = typeof window === "undefined";

export const SSRDataFetch = () => {
  const ssrContext = useSSRContext();

  const [results, setResults] = useState<FetchResult[]>([
    { key: "time", label: "服务端时间", status: "idle" },
    { key: "random", label: "随机数据", status: "idle" },
    { key: "env", label: "环境变量", status: "idle" },
  ]);

  const simulateFetch = (key: string) => {
    setResults(prev =>
      prev.map(r => (r.key === key ? { ...r, status: "loading" as const } : r))
    );

    setTimeout(() => {
      const dataMap: Record<string, string> = {
        time: new Date().toLocaleString("zh-CN"),
        random: `随机值：${Math.random().toFixed(6)}`,
        env: `当前模式：${import.meta.env.MODE}`,
      };

      setResults(prev =>
        prev.map(r =>
          r.key === key ? { ...r, status: "success" as const, data: dataMap[key] } : r
        )
      );
    }, 800);
  };

  const fetchAll = () => {
    results.forEach(r => simulateFetch(r.key));
  };

  const statusTag = (status: FetchResult["status"]) => {
    switch (status) {
      case "idle":
        return <Tag>待请求</Tag>;
      case "loading":
        return <Tag icon={<SyncOutlined spin />} color="processing">请求中</Tag>;
      case "success":
        return <Tag icon={<CheckCircleOutlined />} color="success">成功</Tag>;
      case "error":
        return <Tag color="error">失败</Tag>;
    }
  };

  // 判断 token 状态
  const hasToken = isServer ? !!ssrContext?.token : true; // 客户端由 Redux 管理
  const tokenPreview = ssrContext?.token
    ? `${ssrContext.token.slice(0, 20)}...${ssrContext.token.slice(-10)}`
    : undefined;

  return (
    <div style={{ maxWidth: 720, margin: "40px auto", padding: "0 16px" }}>
      <Title level={2}>
        <ApiOutlined /> SSR 数据获取测试
      </Title>

      <Paragraph>
        此页面演示 SSR 场景下通过 Cookie 传递 Token 的认证模式。
        登录后刷新此页面，服务端能从 Cookie 中读取 Token 用于请求需要认证的 API。
      </Paragraph>

      {/* Token 状态展示 */}
      <Card
        title={
          <span>
            <SafetyCertificateOutlined /> SSR 认证状态
          </span>
        }
        style={{ marginBottom: 24 }}
      >
        <Descriptions column={1} size="small">
          <Descriptions.Item label="渲染环境">
            <Tag color={isServer ? "blue" : "green"}>{isServer ? "服务端 (SSR)" : "客户端 (Hydrated)"}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Cookie Token">
            {ssrContext?.token ? (
              <Text code>{tokenPreview}</Text>
            ) : (
              <Text type="secondary">未检测到（未登录或客户端渲染）</Text>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="认证可用">
            <Tag color={hasToken ? "success" : "warning"}>
              {hasToken ? "可以发起认证请求" : "需要先登录"}
            </Tag>
          </Descriptions.Item>
        </Descriptions>

        {!ssrContext?.token && !isServer && (
          <Alert
            type="info"
            showIcon
            style={{ marginTop: 12 }}
            message="客户端渲染模式"
            description="当前为客户端渲染（hydration 后），Token 由 Redux 状态管理。刷新页面可查看服务端渲染时的 Token 传递效果。"
          />
        )}
      </Card>

      {/* 数据请求演示 */}
      <Card style={{ marginBottom: 24 }}>
        <Text type="secondary">服务端渲染时间：{serverRenderTime}</Text>
        <Divider />
        <Button type="primary" icon={<ApiOutlined />} onClick={fetchAll}>
          全部请求
        </Button>
      </Card>

      <List
        dataSource={results}
        renderItem={item => (
          <List.Item
            actions={[
              statusTag(item.status),
              <Button
                key="fetch"
                size="small"
                onClick={() => simulateFetch(item.key)}
                loading={item.status === "loading"}
              >
                请求
              </Button>,
            ]}
          >
            <List.Item.Meta
              title={item.label}
              description={item.data ?? "暂无数据"}
            />
          </List.Item>
        )}
      />

      {/* 使用说明 */}
      <Card title="Cookie Token 工作流程" style={{ marginTop: 24 }}>
        <Paragraph>
          <ol>
            <li>用户登录 → Token 存入 Redux + 写入 Cookie（<Text code>auth_token</Text>）</li>
            <li>浏览器请求 SSR 页面 → Cookie 自动携带</li>
            <li>Express 通过 <Text code>cookie-parser</Text> 读取 → 传入 <Text code>render(url, {"{ token }"})</Text></li>
            <li>SSR 组件通过 <Text code>useSSRContext()</Text> 获取 Token → 用于服务端 API 请求</li>
            <li>Token 刷新 / 退出登录 → 同步更新 / 清除 Cookie</li>
          </ol>
        </Paragraph>
      </Card>
    </div>
  );
};

export default SSRDataFetch;

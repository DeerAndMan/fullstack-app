import { useState } from "react";
import { Button, Card, Tag, Space, Typography, Statistic, Row, Col } from "antd";
import {
  CheckOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  RocketOutlined,
} from "@ant-design/icons";

const { Title, Paragraph } = Typography;

const serverTime = new Date().toISOString();

export const SSRPerformance = () => {
  const [counter, setCounter] = useState(0);
  const [hydrated, setHydrated] = useState(false);

  const handleClick = () => {
    setCounter(c => c + 1);
    if (!hydrated) setHydrated(true);
  };

  const startMark = performance.now();

  return (
    <div style={{ maxWidth: 720, margin: "40px auto", padding: "0 16px" }}>
      <Title level={2}>
        <ThunderboltOutlined /> SSR 性能测试
      </Title>

      <Paragraph>
        此页面用于测试 SSR 渲染性能指标。服务端渲染时会记录渲染耗时，客户端 hydration
        后可交互。
      </Paragraph>

      <Row gutter={16} style={{ marginBottom: 24 }}>
        <Col span={8}>
          <Card>
            <Statistic
              title="服务端渲染时间"
              value={serverTime}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ fontSize: 14 }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="页面加载标记"
              value={Math.round(startMark)}
              suffix="ms"
              prefix={<RocketOutlined />}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic
              title="交互次数"
              value={counter}
              prefix={<ThunderboltOutlined />}
              valueStyle={{ color: counter > 0 ? "#3f8600" : undefined }}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Hydration 交互测试">
        <Space direction="vertical" size="middle">
          <Paragraph>
            通过连续点击按钮测试 hydration 后的客户端响应能力。
          </Paragraph>
          <Space>
            <Button type="primary" size="large" onClick={handleClick}>
              点击测试：{counter}
            </Button>
            {hydrated && (
              <Tag icon={<CheckOutlined />} color="success">
                Hydration 已生效
              </Tag>
            )}
          </Space>
        </Space>
      </Card>
    </div>
  );
};

export default SSRPerformance;

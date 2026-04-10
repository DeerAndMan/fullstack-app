import { useState } from "react";
import { Button, Card, Descriptions, Tag, Space, Typography } from "antd";
import { CheckOutlined, CloudServerOutlined } from "@ant-design/icons";

const { Title, Paragraph } = Typography;

const serverTime = new Date().toISOString();

export const SSRDemo = () => {
  const [count, setCount] = useState(0);
  const [hydrated, setHydrated] = useState(false);

  const handleClick = () => {
    setCount(c => c + 1);
    if (!hydrated) setHydrated(true);
  };

  return (
    <div style={{ maxWidth: 720, margin: "40px auto", padding: "0 16px" }}>
      <Title level={2}>
        <CloudServerOutlined /> Vite 8 SSR Demo
      </Title>

      <Paragraph>
        此页面由服务端渲染（SSR）生成。查看页面源代码可以看到 HTML 中包含完整的渲染内容， 而非空的{" "}
        <Tag>&lt;div id=`root`&gt;</Tag>。
      </Paragraph>

      <Card title="服务端渲染信息" style={{ marginBottom: 24 }}>
        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="渲染时间">{serverTime}</Descriptions.Item>
          <Descriptions.Item label="渲染方式">
            <Tag color="green">Server-Side Rendering</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="React 版本">{__REACT_VERSION__}</Descriptions.Item>
          <Descriptions.Item label="运行环境">
            {typeof window === "undefined" ? (
              <Tag color="blue">Node.js (Server)</Tag>
            ) : (
              <Tag color="orange">Browser (Client)</Tag>
            )}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="客户端交互测试（Hydration）">
        <Space direction="vertical" size="middle">
          <Paragraph>
            点击按钮测试客户端 hydration 是否生效。如果计数器能正常递增，说明 hydration 成功。
          </Paragraph>
          <Space>
            <Button type="primary" onClick={handleClick}>
              点击计数：{count}
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

declare const __REACT_VERSION__: string;

export default SSRDemo;

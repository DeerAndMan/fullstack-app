import { theme } from "antd";

// 亮色主题配置 - 米色系
export const lightTheme = {
  token: {
    colorPrimary: "#00b96b",
    colorBgBase: "#f5f2eb",
    colorBgContainer: "#faf7f0",
    colorBgElevated: "#ffffff",
    colorBgLayout: "#f0ece3",
  },
  algorithm: theme.defaultAlgorithm,
};

// 暗色主题配置 - 暗黑系
export const darkTheme = {
  token: {
    colorPrimary: "#00b96b",
    colorBgBase: "#000000",
    colorBgContainer: "#141414",
    colorBgElevated: "#1f1f1f",
    colorBgLayout: "#000000",
    colorText: "#e0e0e0",
    colorBgMask: "rgba(0, 0, 0, 0.75)",
  },
  algorithm: theme.darkAlgorithm,
};

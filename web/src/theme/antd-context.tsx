import React, { createContext, useState, useContext, useEffect } from "react";
import { ConfigProvider, theme } from "antd";
import zhCN from "antd/locale/zh_CN";
import "dayjs/locale/zh-cn";

import { lightTheme, darkTheme } from "./antd-theme";

type ThemeType = "light" | "dark";

interface ThemeContextType {
  theme: ThemeType;
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextType>({
  theme: "light",
  toggleTheme: () => {},
});

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [theme, setTheme] = useState<ThemeType>(
    () => (localStorage.getItem("theme") as ThemeType) || "light"
  );

  const toggleTheme = () => {
    const newTheme = theme === "light" ? "dark" : "light";
    setTheme(newTheme);
    localStorage.setItem("theme", newTheme);
  };

  useEffect(() => {
    document.body.className = theme;
  }, [theme]);

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      <ConfigProvider locale={zhCN} theme={theme === "light" ? lightTheme : darkTheme}>
        {children}

        {/* 用于在子组件中使用主题 */}
        <ContextHolder />
      </ConfigProvider>
    </ThemeContext.Provider>
  );
};

// eslint-disable-next-line react-refresh/only-export-components
export const useTheme = () => useContext(ThemeContext);

export const ContextHolder = () => {
  const { token } = theme.useToken();

  useEffect(() => {
    const root = document.documentElement;
    root.style.setProperty("--app-color-primary", token.colorPrimary);
    root.style.setProperty("--app-color-bg-base", token.colorBgBase);
    root.style.setProperty("--app-color-bg-container", token.colorBgContainer);
    root.style.setProperty("--app-color-bg-elevated", token.colorBgElevated);
    root.style.setProperty("--app-color-bg-layout", token.colorBgLayout);
  }, [token]);

  return null;
};

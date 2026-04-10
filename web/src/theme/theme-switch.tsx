import { Switch } from "antd";
import { useTheme } from "./antd-context";

export default function ThemeSwitch() {
  const { theme, toggleTheme } = useTheme();

  return (
    <div>
      <Switch
        checked={theme === "dark"}
        onChange={toggleTheme}
        checkedChildren="🌙"
        unCheckedChildren="☀️"
      />
    </div>
  );
}

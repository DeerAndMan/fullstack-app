import { useCallback, useEffect, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { Avatar, Button, Dropdown, Menu, Space, theme } from "antd";
import { UserOutlined } from "@ant-design/icons";

import { navRouter } from "@/router";
import ThemeSwitch from "@/theme/theme-switch";
import { useAuthStore } from "@/stores/auth";
import { useGlobalStore } from "@/stores/global";
import { base64ToImg } from "@/utils/img";
import { clearAuthCookie } from "@/utils/cookie";
import { RoleKey } from "@/types/enum";

import type { MenuProps } from "antd";
import type { NavRouter } from "@/router";

type MenuItem = Required<MenuProps>["items"][number];

/**
 * Nav 组件
 * @returns {React.FunctionComponent}
 */
export const Nav = () => {
  const { token } = theme.useToken();
  const navigate = useNavigate();
  const location = useLocation();
  const messageApi = useGlobalStore(s => s.messageApi);
  const { role, menuRoles, user } = useAuthStore();
  const storeLogout = useAuthStore(s => s.storeLogout);

  const [menuItems, setMenuItems] = useState<MenuItem[]>([]);

  const handleLogout = async () => {
    try {
      const success = await storeLogout();
      if (success) {
        clearAuthCookie();
        navigate("/login");
      } else {
        messageApi?.error("退出登录失败");
      }
    } catch {
      messageApi?.error("退出登录失败");
    }
  };

  const transformNode = useCallback(
    (routeData: NavRouter[], showKeys?: Record<string, boolean>) => {
      const transform = (node: NavRouter): MenuItem => {
        if (showKeys && !showKeys[node.path]) return null;
        const menuItem: MenuItem = {
          key: node.path === "" ? `menu-${node.name}` : node.path,
          label: node.path !== "" ? <Link to={node.path}>{node.name}</Link> : node.name,
        };

        if (node.children && node.children.length > 0) {
          return { ...menuItem, children: node.children.map(child => transform(child)) };
        }

        return menuItem;
      };

      return routeData.map(node => transform(node)).filter(m => m);
    },
    []
  );

  // 将路由转换为菜单项
  useEffect(() => {
    // 超级管理员显示所有菜单
    if (role?.role_key === RoleKey.SUPER_ADMIN) {
      setMenuItems(transformNode(navRouter));
      return;
    }

    const showNavKeys: Record<string, boolean> = {};
    menuRoles?.forEach(m => (showNavKeys[m.link_url] = true));
    setMenuItems(transformNode(navRouter, showNavKeys));
  }, [transformNode, menuRoles, role?.role_key]);

  return (
    <div
      className="rounded-md flex items-center justify-between sticky top-0 z-50 shadow-sm"
      style={{
        backgroundColor: token.colorBgBase,
        borderBottom: `1px solid ${token.colorBorderSecondary}`,
      }}
    >
      <Menu
        mode="horizontal"
        selectedKeys={[location.pathname]}
        items={menuItems}
        style={{
          flex: 1,
          minWidth: 50,
          border: "none",
          backgroundColor: "transparent",
          color: token.colorText,
        }}
      />

      <Space className="pr-4 flex direction-row items-center">
        <ThemeSwitch />
        <Link to="/user/operation">
          <UserOutlined style={{ color: "#eb2f96", fontSize: "18px" }} />
        </Link>

        <Dropdown
          trigger={["click", "hover"]}
          menu={{
            items: [
              { key: 1, label: <Link to="/user">详情</Link> },
              {
                key: 2,
                label: (
                  <Button size="small" type="text" onClick={handleLogout}>
                    登出
                  </Button>
                ),
              },
            ],
          }}
          overlayClassName="rounded-md w-[120px] p-2"
        >
          <Avatar
            src={<img src={user?.avatar ? base64ToImg(user.avatar) : ""} />}
            className="cursor-pointer"
          >
            无
          </Avatar>
        </Dropdown>
      </Space>
    </div>
  );
};

export default Nav;

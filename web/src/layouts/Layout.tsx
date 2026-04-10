import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { message } from "antd";

import Nav from "./Nav";
import { useAuthStore } from "@/stores/auth";
import { useEnumStore } from "@/stores/enum";
import { useGlobalStore } from "@/stores/global";

export type Props = React.PropsWithChildren;

/**
 * Layout 组件
 * @returns {React.FunctionComponent}
 */
export const Layout = (props: Props) => {
  const location = useLocation();
  const navigate = useNavigate();
  const stateToken = useAuthStore(s => s.token);

  const [messageApi, contextHolder] = message.useMessage();

  const { children } = props;

  const hideNavList = ["/login"];

  useEffect(() => {
    useGlobalStore.getState().setMessageApi(messageApi);
  }, [messageApi]);

  useEffect(() => {
    if (!stateToken) return;
    useEnumStore.getState().fetchRoles();
  }, [stateToken]);

  useEffect(() => {
    if (!stateToken) {
      navigate("/login", { replace: true });
      return;
    }
  }, [stateToken, navigate]);

  return (
    <div className="min-h-screen">
      {contextHolder}

      {hideNavList.includes(location.pathname) ? null : <Nav />}

      {location.pathname === "/login" ? (
        <div className="h-screen">{children}</div>
      ) : (
        <div className="p-4">{children}</div>
      )}
    </div>
  );
};

export default Layout;

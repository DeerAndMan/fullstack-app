import { useNavigate } from "react-router-dom";
import { Button } from "antd";

import { useAuthStore } from "@/stores/auth";
import { useGlobalStore } from "@/stores/global";
import { tradeListQuery } from "@/api/trade";
import { toPercent } from "@/utils/number";

export const Home = () => {
  const navigate = useNavigate();
  const storeLogout = useAuthStore(s => s.storeLogout);
  const messageApi = useGlobalStore(s => s.messageApi);

  const { queryList } = tradeListQuery();
  const last = queryList.data?.data[queryList.data?.data.length - 1];

  const handleLogout = async () => {
    try {
      const success = await storeLogout();
      if (success) {
        navigate("/login");
      } else {
        messageApi?.error("退出登录失败");
      }
    } catch {
      messageApi?.error("退出登录失败");
    }
  };

  return (
    <div className="relative">
      <div className="text-sm font-bold min-h-192 text-primary">
        总结：{toPercent(last?.drhz)}
        &nbsp; &nbsp;
        {last?.dryk}
      </div>
      <Button block className="btn btn-active btn-sm btn-secondary " onClick={handleLogout}>
        登出
      </Button>
    </div>
  );
};

export default Home;

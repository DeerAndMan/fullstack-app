import { theme } from "antd";

import { tradeListQuery } from "@/api/trade";
import { toPercent } from "@/utils/number";
import { withTokenColors } from "../hotspots/panels";
import marketStyles from "../hotspots/hotspots.module.css";
import { Headlines } from "../hotspots/headlines";
import styles from "./home.module.css";

export const Home = () => {
  const { token } = theme.useToken();
  const { queryList } = tradeListQuery();
  const last = queryList.data?.data[queryList.data?.data.length - 1];

  return (
    <main className={`${marketStyles.page} ${styles.home}`} style={withTokenColors(token)}>
      <div className={styles.summary}>
        <div className="text-sm font-bold text-primary">
          总结：{toPercent(last?.drhz)}
          &nbsp; &nbsp;
          {last?.dryk}
        </div>
      </div>
      <Headlines />
    </main>
  );
};

export default Home;

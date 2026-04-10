import { Suspense } from "react";
import { BrowserRouter } from "react-router-dom";
import { ConfigProvider } from "antd";
import dayjs from "dayjs";

import locale from "antd/locale/zh_CN";
import "dayjs/locale/zh-cn";
dayjs.locale("zh-cn");
import { Router } from "@/router/main";
import QueryProvider from "./components/query/QueryProvider";

const App = () => {
  return (
    <QueryProvider>
      <ConfigProvider locale={locale}>
        <BrowserRouter>
          <Suspense fallback={<span className="loading loading-infinity loading-xl" />}>
            <Router />
          </Suspense>
        </BrowserRouter>
      </ConfigProvider>
    </QueryProvider>
  );
};

export default App;

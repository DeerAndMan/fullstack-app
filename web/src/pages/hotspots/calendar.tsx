import { CalendarOutlined, CloseCircleFilled } from "@ant-design/icons";
import dayjs from "dayjs";
import zhCN from "antd/es/date-picker/locale/zh_CN";
import "dayjs/locale/zh-cn";

const dateFormat = "YYYY年M月D日";

export const hotspotDatePicker = {
  suffixIcon: <CalendarOutlined aria-label="日历" />,
  allowClear: { clearIcon: <CloseCircleFilled aria-label="清除日期" /> },
  format: ["YYYY-MM-DD", dateFormat],
  locale: {
    ...zhCN,
    lang: {
      ...zhCN.lang,
      shortWeekDays: ["周日", "周一", "周二", "周三", "周四", "周五", "周六"],
      monthFormat: "M月",
      fieldDateFormat: dateFormat,
      fieldMonthFormat: "YYYY年M月",
      previousYear: "上一年",
      nextYear: "下一年",
    },
  },
};

export function displayMarketDate(date?: string, includeYear = true) {
  return date ? dayjs(date).format(includeYear ? dateFormat : "M月D日") : "—";
}

import { useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs from "dayjs";

import { getTrade } from "./trade.service";

const dateFormat = "YYYY-MM-DD";

const queryKey = ["trade", "router"];

export const tradeListQuery = (enabled = true) => {
  const enabledRef = useRef(true);
  const queryClient = useQueryClient();

  const [dateTime, setDateTime] = useState([dayjs().startOf("day"), dayjs().endOf("day")]);

  const queryList = useQuery({
    queryKey: [...queryKey, "list"],
    queryFn: () =>
      getTrade({
        startTime: dateTime[0].format(dateFormat),
        endTime: dateTime[1].format(dateFormat),
      }),
    enabled: enabledRef.current && enabled,
  });

  const refresh = () => {
    queryClient.invalidateQueries({ queryKey });
  };

  const stateOperations = { enabledRef, dateTime, setDateTime };
  const operations = { refresh };

  return { queryList, stateOperations, operations };
};

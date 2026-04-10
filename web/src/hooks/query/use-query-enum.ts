import { useQuery } from "@tanstack/react-query";

import { getRoleList } from "@/api/enum";

export const roleQuery = (enabled = true) =>
  useQuery({
    queryKey: ["enum", "role"],
    queryFn: () => getRoleList().then(res => (res.code === 0 ? res.data : [])),
    enabled,
  });

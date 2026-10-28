import type { SubscribeItem } from "@/types/xq/subscribe/home";

export type SubscriptionStatus = "all" | "enabled" | "disabled";
export type SubscriptionSort = "priority" | "idAsc" | "idDesc" | "name";

export function selectSubscriptions(
  items: SubscribeItem[],
  search: string,
  status: SubscriptionStatus,
  starredOnly: boolean,
  sort: SubscriptionSort
) {
  const keyword = search.trim().toLocaleLowerCase();
  return items.filter(item => {
    if (starredOnly && !item.is_starred) return false;
    if (status !== "all" && item.enabled !== (status === "enabled")) return false;
    return !keyword || [item.id, item.user_id, item.description || ""]
      .some(value => value.toLocaleLowerCase().includes(keyword));
  }).sort((a, b) => {
    // ID 为 64 位整数字符串，不能转成 Number 比较。
    const byId = a.id.localeCompare(b.id, "en", { numeric: true }) ||
      a.user_id.localeCompare(b.user_id, "en", { numeric: true });
    switch (sort) {
      case "idAsc": return byId;
      case "idDesc": return -byId;
      case "name": return (a.description || "").localeCompare(b.description || "", "zh-CN") || byId;
      default: return Number(b.is_starred) - Number(a.is_starred) || b.sort_order - a.sort_order || byId;
    }
  });
}

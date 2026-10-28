import { describe, expect, it } from "vitest";
import { selectSubscriptions } from "@/sections/subscribe/home/list-utils";
import { SubscribeItemSchema, type SubscribeItem } from "@/types/xq/subscribe/home";

const items: SubscribeItem[] = [
  { id: "373676297131397121", user_id: "6571686717", description: "Alpha|曾用名", enabled: true, is_starred: true, sort_order: 10 },
  { id: "373676297131397120", user_id: "123", description: "Beta", enabled: false, is_starred: true, sort_order: 10 },
  { id: "2", user_id: "456", description: null, enabled: true, is_starred: false, sort_order: 999999 },
  { id: "10", user_id: "789", description: "Gamma", enabled: true, is_starred: true, sort_order: 20 },
];

describe("订阅列表筛选和排序", () => {
  it("星标优先，同组按排序值倒序，同值按精确 ID 稳定排序", () => {
    const selected = selectSubscriptions(items, "", "all", false, "priority");
    expect(selected.map(item => item.id)).toEqual(["10", "373676297131397120", "373676297131397121", "2"]);
    expect(items[0].id).toBe("373676297131397121");
  });

  it("搜索支持 ID、用户ID、曾用名及忽略大小写与首尾空格", () => {
    for (const search of ["  ALPHA ", "曾用名", "6571686717", "373676297131397121"]) {
      expect(selectSubscriptions(items, search, "all", false, "priority").map(item => item.id)).toEqual([items[0].id]);
    }
    expect(selectSubscriptions(items, "不存在", "all", false, "priority")).toEqual([]);
  });

  it("组合筛选先处理全部记录，再交给表格分页", () => {
    expect(selectSubscriptions(items, "", "enabled", true, "priority").map(item => item.id)).toEqual(["10", items[0].id]);
    expect(selectSubscriptions(items, "", "disabled", false, "priority").map(item => item.id)).toEqual([items[1].id]);
    expect(selectSubscriptions(items, "Beta", "enabled", true, "priority")).toEqual([]);
  });

  it("支持 ID 升降序且不会丢失大整数精度", () => {
    expect(selectSubscriptions(items, "", "all", false, "idAsc").map(item => item.id)).toEqual(["2", "10", items[1].id, items[0].id]);
    expect(selectSubscriptions(items, "", "all", false, "idDesc").map(item => item.id)).toEqual([items[0].id, items[1].id, "10", "2"]);
  });

  it("取消星标和排序归零后按新值重新排列", () => {
    const updated = items.map(item => ({ ...item, is_starred: false, sort_order: 0 }));
    expect(selectSubscriptions(updated, "", "all", false, "priority").map(item => item.id)).toEqual(["2", "10", items[1].id, items[0].id]);
    expect(selectSubscriptions(updated, "", "all", true, "priority")).toEqual([]);
  });

  it("兼容旧接口记录，并保留服务器返回的星标和排序", () => {
    expect(SubscribeItemSchema.parse({ id: "1", user_id: "2", enabled: true })).toMatchObject({ is_starred: false, sort_order: 0 });
    expect(SubscribeItemSchema.parse(items[0])).toEqual(items[0]);
  });
});

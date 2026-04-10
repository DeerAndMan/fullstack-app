export const apiControl = {
  user: {
    login: "/admin/user/login",
    admin: "/admin/user",
  },
  enum: {
    role: "/enum/role",
  },
  trade: {
    root: "/trade",
    summary: "/trade/summary",
  },
  menu: {
    list: "/menu-router/list",
    // action: "/menu-router",
    add: "/menu-router/add",
    roleRouting: "/menu-router/role-routing",
    roleListByRoleId: "/menu-router/role-routing/",
  },
  xq: {
    subscribe: {
      list: "/xq/subscription/all",
      add: "/xq/subscription",
      formerName: "/xq/subscription/former-name",
      delete: "/xq/subscription/delete",
      detail: "/xq/subscription/detail",
      detailTable: "/xq/subscription/detail-table",
      search: "/xq/subscription/search",
      export: "/xq/subscription/export",
      import: "/xq/subscription/import",
      subscribe: "/xq/subscription/subscribe",
      toggle: "/xq/subscription/toggle",
    },
  },
};

export default apiControl;

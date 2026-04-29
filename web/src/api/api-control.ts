export const apiControl = {
  auth: {
    login: "/api/v1/auth/login",
    register: "/api/v1/auth/register",
    refreshToken: "/api/v1/auth/refresh-token",
    logout: "/api/v1/auth/logout",
  },
  user: {
    list: "/api/v1/users",
    profile: "/api/v1/users/profile",
    detail: "/api/v1/users",
  },
  role: {
    list: "/api/v1/roles",
    all: "/api/v1/roles/all",
    detail: "/api/v1/roles",
  },
  upload: "/api/v1/upload",
  trade: {
    index: "/api/v1/trade/index",
    summary: "/api/v1/trade/summary",
  },
  energy: {
    insert: "/api/v1/energy/insert",
  },
  jyData: {
    latest: "/api/v1/jydata/latest",
    list: "/api/v1/jydata/list",
  },
  sse: {
    chatMessages: "/api/v1/sse/chat-messages",
  },
  ai: {
    conversations: "/api/v1/ai/conversations",
  },
  menu: {
    list: "/menu-router/list",
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

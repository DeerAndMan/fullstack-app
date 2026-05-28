export const apiPathsV1 = {
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
    list: "/api/v1/menus",
    add: "/api/v1/menus",
    roleBinding: "/api/v1/menus/role-binding",
    roleListByRoleId: "/api/v1/menus/role-binding",
  },
  xq: {
    subscribe: {
      list: "/api/v1/subscriptions",
      add: "/api/v1/subscriptions",
      formerName: "/api/v1/subscriptions/description",
      delete: "/api/v1/subscriptions",
      detail: "/api/v1/subscriptions/detail",
      detailTable: "/api/v1/subscriptions/detail-table",
      toggle: "/api/v1/subscriptions/toggle",
      exists: "/api/v1/subscriptions/exists",
      user: "/api/v1/subscriptions/user",
    },
    themeContent: {
      list: "/api/v1/theme-contents",
      add: "/api/v1/theme-contents",
      batch: "/api/v1/theme-contents/batch",
      detail: "/api/v1/theme-contents",
      user: "/api/v1/theme-contents/user",
      exists: "/api/v1/theme-contents/exists",
      search: "/api/v1/theme-contents/search",
      timeline: "/api/v1/theme-contents/timeline",
    },
  },
  enum: {
    roles: "/api/v1/enums/roles",
  },
};

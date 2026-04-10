const ROUTER_PATH = {
  login: "/login",
  home: "/",
  data: "/data",
  ws: "/ws",
  subscribe: {
    home: "/subscribe",
    list: "/subscribe/list",
    detail: "/subscribe/detail/:id/:userId",
  },
  auth: {
    root: "/auth/role",
  },
  role: {
    list: "/role/list",
    menu: "/role/menu",
  },
  user: {
    root: "/user",
    operation: "/user/operation",
  },
};

export default ROUTER_PATH;

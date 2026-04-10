# Fullstack App

Go + React 企业级全栈 Monorepo。

## 技术栈

### 后端 (server/)

| 类别 | 技术 |
|------|------|
| Web 框架 | Hertz (ByteDance) |
| ORM | GORM 2 |
| 数据库 | MySQL 8.0+ |
| 缓存 | Redis 7+ |
| 认证 | JWT (golang-jwt/jwt v5) |
| 权限 | Casbin (RBAC) |
| 配置 | Viper |
| 日志 | Slog (标准库) |
| 热重载 | Air |

### 前端 (web/)

| 类别 | 技术 |
|------|------|
| 构建 | Vite 8 (SWC) |
| 框架 | React 19 + TypeScript 5.8 |
| UI | Ant Design 5 + Tailwind CSS 4 |
| 状态管理 | Zustand 5 (客户端) + TanStack Query 5 (服务端) |
| 路由 | React Router 7 |
| 图表 | @ant-design/charts + @antv/g2 |
| 校验 | Zod 4 |
| SSR | Express + react-router StaticRouter |

## 项目结构

```
fullstack-app/
├── server/                          # Go 后端
│   ├── cmd/server/main.go           # 程序入口
│   ├── config/                      # 配置文件 (YAML)
│   ├── internal/
│   │   ├── config/                  # Viper 配置加载
│   │   ├── database/                # MySQL + Redis 连接
│   │   ├── handler/                 # HTTP 处理器
│   │   ├── middleware/              # 中间件 (CORS/JWT/Casbin/日志/恢复/链路ID)
│   │   ├── model/                   # 数据模型
│   │   ├── repository/              # 数据访问层
│   │   ├── router/                  # 路由注册
│   │   └── service/                 # 业务逻辑层
│   └── pkg/                         # 公共工具包
│       ├── errcode/                 # 错误码
│       ├── jwt/                     # JWT 签发/解析
│       ├── response/                # 统一响应
│       └── upload/                  # 文件上传
├── web/                             # React 前端
│   ├── src/
│   │   ├── api/                     # Axios 封装 + Zod 验证
│   │   ├── components/              # 通用组件 (chart/form/query/toast)
│   │   ├── hooks/                   # 自定义 Hooks + React Query Hooks
│   │   ├── layouts/                 # 布局 + 导航 (登录守卫)
│   │   ├── pages/                   # 路由页面
│   │   ├── router/                  # 路由配置 (懒加载)
│   │   ├── sections/                # 页面子模块
│   │   ├── stores/                  # Zustand 状态 (auth/enum/global)
│   │   ├── theme/                   # 主题配置 (亮色/暗色)
│   │   ├── types/                   # TypeScript 类型 + Zod Schema
│   │   └── utils/                   # 工具函数
│   └── vite.config.ts
├── deploy/                          # 部署配置
│   ├── docker/                      # Dockerfile + Nginx
│   └── docker-compose.prod.yml      # 生产环境编排
├── .github/workflows/ci.yml         # CI/CD
├── docker-compose.yml               # 本地开发 (MySQL + Redis)
└── Makefile                         # 顶层命令
```

## 快速开始

### 前置条件

- Go 1.23+
- Node.js 22+ & pnpm
- Docker & Docker Compose
- [Air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`)

### 1. 克隆 & 配置

```bash
cd fullstack-app

# 复制后端配置
cp server/config/config.example.yaml server/config/config.yaml
```

### 2. 一键启动

```bash
make dev
```

等效于：启动 MySQL + Redis → 并行启动前端 (Vite) + 后端 (Air 热重载)

### 3. 分步启动

```bash
# 启动基础设施
docker compose up -d mysql redis

# 启动后端 (终端 1)
cd server && air

# 启动前端 (终端 2)
cd web && pnpm dev
```

### 4. 访问

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:5173 |
| 后端 API | http://localhost:8080 |
| MySQL | 127.0.0.1:3306 (app / app123) |
| Redis | 127.0.0.1:6379 |

## API 接口

所有 API 基础路径：`/api/v1`

### 认证 (公开)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /auth/register | 注册 |
| POST | /auth/login | 登录 |
| POST | /auth/refresh-token | 刷新 Token |

### 用户管理 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /users | 用户列表 (分页) |
| POST | /users | 创建用户 |
| GET | /users/profile | 当前用户信息 |
| GET | /users/:id | 用户详情 |
| PUT | /users/:id | 更新用户 |
| DELETE | /users/:id | 删除用户 |

### 角色管理 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /roles | 角色列表 (分页) |
| GET | /roles/all | 全部角色 |
| POST | /roles | 创建角色 |
| GET | /roles/:id | 角色详情 |
| PUT | /roles/:id | 更新角色 |
| DELETE | /roles/:id | 删除角色 |

### 文件上传 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /upload | 上传文件 |

### 统一响应格式

```json
{
  "code": 0,
  "data": {},
  "message": "success"
}
```

| Code | 含义 |
|------|------|
| 0 | 成功 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 500 | 内部错误 |

## 后端架构

```
请求 → Middleware (链路ID/日志/CORS/恢复) → JWT 认证 → Casbin 权限
  → Handler (参数绑定/校验) → Service (业务逻辑) → Repository (数据访问) → MySQL/Redis
```

依赖注入采用手动构造器方式：

```go
repo := repository.NewUserRepository(db)
svc  := service.NewAuthService(repo)
h    := handler.NewAuthHandler(svc)
```

## 常用命令

```bash
make dev              # 启动开发环境 (MySQL + Redis + 前后端)
make build            # 构建前后端
make test             # 运行测试
make lint             # 代码检查
make clean            # 清理构建产物
make deploy           # Docker 生产部署
```

## 部署

```bash
# 生产环境一键部署
make deploy

# 或手动
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

包含：前端 (Nginx) + 后端 (scratch 镜像, ~10MB) + MySQL + Redis

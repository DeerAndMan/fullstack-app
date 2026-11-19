# Redis 登录功能说明

## 功能概述

后端服务已启用 Redis，并在用户登录时自动将登录信息写入 Redis。

## 修改内容

### 1. main.go
- Redis 连接失败改为致命错误（之前是警告跳过）
- 将 Redis 实例传递给服务层

### 2. wire.go
- 添加 `redis.Client` 导入
- `initHandlers` 函数接收 `rdb *redis.Client` 参数
- `AuthService` 初始化时传入 Redis 客户端

### 3. auth.go (service)
- `AuthService` 结构体新增 `rdb *redis.Client` 字段
- `NewAuthService` 构造函数接收 Redis 客户端
- `Login` 方法在登录成功后调用 `saveLoginInfoToRedis`
- 新增 `saveLoginInfoToRedis` 方法，将登录信息保存到 Redis

## Redis 数据结构

### Key 格式
```
user:login:{userID}
```

### Value 格式（JSON）
```json
{
  "user_id": 1,
  "username": "admin",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "login_time": "2026-06-09 18:51:23"
}
```

### 过期时间
- 7 天（168 小时）

## 测试步骤

### 1. 确保 MySQL 和 Redis 运行
```bash
docker-compose up -d
```

### 2. 启动后端服务
```bash
make dev
# 或
./bin/server
```

### 3. 执行登录请求
```bash
curl -X POST http://localhost:6767/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"name":"admin","password":"admin123"}'
```

### 4. 查看 Redis 中的登录信息
```bash
# 方式 1：使用 redis-cli
redis-cli

# 查看所有登录 key
KEYS user:login:*

# 查看具体用户的登录信息（假设用户 ID 为 1）
GET user:login:1

# 查看过期时间（秒）
TTL user:login:1

# 方式 2：使用 RedisInsight 或其他 GUI 工具
```

### 5. 预期结果

#### 后端日志
```
2026/06/09 18:51:23 INFO 用户登录信息已保存到 Redis
    userID: 1
    username: admin
    key: user:login:1
```

#### Redis 查询结果
```bash
127.0.0.1:6379> GET user:login:1
"{\"user_id\":1,\"username\":\"admin\",\"access_token\":\"eyJhbGci...\",\"login_time\":\"2026-06-09 18:51:23\"}"

127.0.0.1:6379> TTL user:login:1
(integer) 604800  # 7 天 = 604800 秒
```

## 注意事项

1. **非阻塞设计**：Redis 写入失败不会阻塞登录流程，只会记录错误日志
2. **自动过期**：登录信息会在 7 天后自动过期
3. **覆盖策略**：同一用户重复登录会覆盖之前的登录信息
4. **Logout 未清理**：当前 `Logout` 方法仍为空实现，可以后续扩展为清理 Redis 中的登录信息

## 后续扩展方向

### 1. Logout 清理 Redis
```go
func (s *AuthService) Logout(userID uint) error {
    ctx := context.Background()
    key := fmt.Sprintf("user:login:%d", userID)
    return s.rdb.Del(ctx, key).Err()
}
```

### 2. Token 黑名单机制
- 登出时将 access token 加入黑名单
- JWT 中间件检查 token 是否在黑名单中

### 3. 在线用户统计
```bash
# 获取所有在线用户数
KEYS user:login:* | wc -l
```

### 4. 单点登录（踢出旧设备）
- 登录时检查是否已存在登录信息
- 如存在，将旧 token 加入黑名单

### 5. 登录历史记录
```
user:login:history:{userID} -> List
```

## 故障排查

### Redis 连接失败
```
FATAL connect redis failed error="connect redis: dial tcp 127.0.0.1:6379: connect: connection refused"
```
**解决方案**：
1. 检查 Redis 是否运行：`redis-cli ping`
2. 检查配置文件：`config/config.yaml` 中的 Redis 配置
3. 启动 Redis：`docker-compose up -d redis`

### 登录成功但 Redis 无数据
**排查步骤**：
1. 查看后端日志，是否有 "保存登录信息到 Redis 失败" 错误
2. 检查 Redis 连接：`redis-cli ping`
3. 手动测试写入：`redis-cli SET test:key "test value"`
4. 检查 Redis 配置权限

### Key 找不到
```bash
127.0.0.1:6379> GET user:login:1
(nil)
```
**可能原因**：
1. 数据已过期（7 天后自动删除）
2. 用户 ID 不匹配
3. 登录请求实际失败了
4. Redis 数据被手动清理

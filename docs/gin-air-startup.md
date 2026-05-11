# GIN + Air 启动执行路径

本文说明在 `gin/` 目录使用 `air` 启动服务时的执行路径，以及关键配置与入口位置。

## 1) Air 构建与运行入口

`air` 读取 `gin/.air.toml`，按如下流程构建并运行：

1. 使用 `go build -o ./tmp/main{{.GOEXE}} .` 编译 `gin/` 目录下的主程序
2. 运行编译后的二进制，并传入参数：`server --env development`

对应配置：

```toml
[build]
  cmd = "go build -o ./tmp/main{{.GOEXE}} ."
  bin = "./tmp/main{{.GOEXE}}"
  args_bin = ["server", "--env", "development"]
```

## 2) 主入口

主入口在 `gin/main.go`：

```
main.go
  -> cmd.Execute()
```

`cmd.Execute()` 来自 `gin/cmd/root.go`，负责启动 Cobra 命令树。

## 3) Cobra 初始化与配置读取

执行 `cmd.Execute()` 后：

1. `cobra.OnInitialize(initConfig)` 被触发
2. `initConfig()` 使用 `--env` 选择配置文件名：
   - `config.<env>.yaml`
3. 配置文件搜索路径：
   - `./config`
   - `./gin/config`
   - `.`
   - `$HOME/.gin-server`

## 4) server 子命令执行路径

`air` 传入 `server`，触发 `gin/cmd/server.go` 中的 `serverCmd`：

```
serverCmd (Run: runServer)
  -> initialize.InitConfig(env)
  -> utils.GetLogger()
  -> gin.SetMode(...)
  -> gin.Default()
  -> CORS 中间件配置
  -> Session 中间件配置
  -> routers.InitRouters(r)
  -> api/router.InitRouter(r)
  -> r.Run(":<port>")
```

关键点：

- `--env development` 决定加载 `config.development.yaml`
- `--port`/`--mode` 参数可覆盖配置文件
- 路由初始化来自 `gin/routers/` 与 `gin/api/router/`

## 5) 简版执行链路（从 air 开始）

```
air
  -> go build ./gin
  -> ./tmp/main server --env development
     -> gin/main.go (cmd.Execute)
       -> cobra initConfig()
       -> serverCmd.runServer()
         -> initialize.InitConfig(env)
         -> gin.SetMode / gin.Default
         -> middlewares + routes
         -> r.Run(":<port>")
```

.PHONY: dev free-port build build-linux test lint swagger tidy

# 后端监听端口。可通过 `make dev PORT=6768` 临时覆盖。
PORT ?= 6767

# 根据操作系统选择 air 配置文件
# Windows 需要 .exe 后缀（cmd 才识别为可执行），Unix(Mac/Linux/Debian) 不需要
ifeq ($(OS),Windows_NT)
    AIR_CONFIG := .air.windows.toml
else
    AIR_CONFIG := .air.toml
endif

dev:
	@$(MAKE) free-port PORT=$(PORT)
	air -c $(AIR_CONFIG)

# Air 只负责管理它自己启动的子进程；如果上一次异常退出，残留进程可能仍占用端口。
# 这里仅在启动 Air 前清理一次，避免把 `lsof` 写进 [build].pre_cmd 后每次热重载都误杀进程。
free-port:
	@if command -v lsof >/dev/null 2>&1; then \
		pids=$$(lsof -tiTCP:$(PORT) -sTCP:LISTEN 2>/dev/null || true); \
		if [ -n "$$pids" ]; then \
			echo "检测到端口 $(PORT) 被占用，停止进程: $$pids"; \
			kill $$pids 2>/dev/null || true; \
			for i in 1 2 3 4 5; do \
				lsof -tiTCP:$(PORT) -sTCP:LISTEN >/dev/null 2>&1 || break; \
				sleep 1; \
			done; \
			if lsof -tiTCP:$(PORT) -sTCP:LISTEN >/dev/null 2>&1; then \
				echo "错误：端口 $(PORT) 仍被占用，请手动检查：lsof -nP -iTCP:$(PORT) -sTCP:LISTEN"; \
				exit 1; \
			fi; \
		fi; \
	else \
		echo "警告：未找到 lsof，跳过端口 $(PORT) 检查"; \
	fi

# 本地构建（当前平台）
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

# 交叉编译 Debian/Linux x86_64 发布二进制
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/server-linux ./cmd/server

test:
	go test ./... -v -cover

lint:
	golangci-lint run ./...

swagger:
	swag init -g cmd/server/main.go -o docs

tidy:
	go mod tidy

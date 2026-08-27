.PHONY: dev dev-web dev-server dev-server-remote build build-web build-server build-linux build-server-linux test lint clean

# ── 远程数据库配置 ────────────────────────────────────────
# 远程连接信息从 server/.env.remote.local 读取（该文件被 .gitignore 忽略，不进版本库）
# 首次使用：cp server/.env.remote.local.example server/.env.remote.local，再填入真实地址/密码
# 用法：make dev-server-remote
REMOTE_ENV_FILE = server/.env.remote.local

# ── Development ──────────────────────────────────────────
dev:
	@if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then \
		if docker compose up -d mysql redis; then \
			echo "MySQL & Redis started"; \
		else \
			echo "警告：Docker Compose 启动 MySQL/Redis 失败，继续启动前后端"; \
		fi; \
	else \
		echo "未检测到正在运行的 Docker，跳过启动 MySQL/Redis（请确保数据库已可用，如连远程库）"; \
	fi
	$(MAKE) -j2 dev-web dev-server

dev-web:
	cd web && pnpm dev

dev-server:
	cd server && $(MAKE) dev

# 连接远程数据库启动后端（不依赖本地 docker MySQL）
# 通过 set -a 把 env 文件里的变量全部导出为环境变量，由 Viper 的 APP_ 前缀覆盖 config.yaml
dev-server-remote:
	@test -f $(REMOTE_ENV_FILE) || { echo "缺少 $(REMOTE_ENV_FILE)，请先复制 $(REMOTE_ENV_FILE).example 并填写"; exit 1; }
	cd server && set -a && . ./.env.remote.local && set +a && $(MAKE) dev

# ── Build ────────────────────────────────────────────────
build: build-web build-server

build-web:
	cd web && pnpm build

build-server:
	cd server && CGO_ENABLED=0 go build -o bin/server ./cmd/server

# 交叉编译 Linux x86_64 发布二进制（产物：server/bin/server-linux）
build-linux:
	cd server && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/server-linux ./cmd/server

# 在服务器（Linux）本机编译并部署到站点目录 /www/wwwroot/dragot
# 用法：在服务器上 git pull 后执行 make build-server-linux
# 说明：
#   - 服务器本身是 Linux，无需交叉编译，直接 go build 即可
#   - 产物为 /www/wwwroot/dragot/server-linux
#   - 同时把 config/config.yaml 一并复制过去（若目标已存在则不覆盖，避免覆盖线上配置）
DEPLOY_DIR ?= /www/wwwroot/dragot
build-server-linux:
	mkdir -p $(DEPLOY_DIR)/config
	cd server && CGO_ENABLED=0 go build -ldflags="-s -w" -o $(DEPLOY_DIR)/server-linux ./cmd/server
	@if [ ! -f $(DEPLOY_DIR)/config/config.yaml ]; then \
		cp server/config/config.yaml $(DEPLOY_DIR)/config/config.yaml && \
		echo "已复制 config.yaml 到 $(DEPLOY_DIR)/config/"; \
	else \
		echo "$(DEPLOY_DIR)/config/config.yaml 已存在，跳过复制（如需更新请手动覆盖）"; \
	fi
	@echo "构建完成：$(DEPLOY_DIR)/server-linux"
	@echo "重启服务请执行：pkill -f server-linux; cd $(DEPLOY_DIR) && sudo -u www nohup ./server-linux -config config/config.yaml > server.log 2>&1 &"

# ── Test ─────────────────────────────────────────────────
test:
	cd web && pnpm test
	cd server && go test ./...

# ── Lint ─────────────────────────────────────────────────
lint:
	cd web && pnpm lint
	cd server && golangci-lint run ./...

# ── Deploy ───────────────────────────────────────────────
deploy:
	docker compose -f deploy/docker-compose.prod.yml up -d --build

# ── Tools ────────────────────────────────────────────────
swagger:
	cd server && swag init -g cmd/server/main.go -o docs

migrate-up:
	cd server && migrate -path migrations -database "$$DB_URL" up

migrate-down:
	cd server && migrate -path migrations -database "$$DB_URL" down 1

# ── Clean ────────────────────────────────────────────────
clean:
	rm -rf web/dist server/bin server/tmp

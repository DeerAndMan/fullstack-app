.PHONY: dev build build-linux test lint swagger tidy

# 根据操作系统选择 air 配置文件
# Windows 需要 .exe 后缀（cmd 才识别为可执行），Unix(Mac/Linux/Debian) 不需要
ifeq ($(OS),Windows_NT)
    AIR_CONFIG := .air.windows.toml
else
    AIR_CONFIG := .air.toml
endif

dev:
	air -c $(AIR_CONFIG)

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

.PHONY: dev build test lint swagger

dev:
	air

build:
	CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

test:
	go test ./... -v -cover

lint:
	golangci-lint run ./...

swagger:
	swag init -g cmd/server/main.go -o docs

tidy:
	go mod tidy

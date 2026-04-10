.PHONY: dev dev-web dev-server build build-web build-server test lint clean

# ── Development ──────────────────────────────────────────
dev:
	docker compose up -d mysql redis
	@echo "MySQL & Redis started"
	$(MAKE) -j2 dev-web dev-server

dev-web:
	cd web && pnpm dev

dev-server:
	cd server && air

# ── Build ────────────────────────────────────────────────
build: build-web build-server

build-web:
	cd web && pnpm build

build-server:
	cd server && CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

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

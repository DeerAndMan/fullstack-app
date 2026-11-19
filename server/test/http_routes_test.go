package test_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"fullstack-app/server/internal/handler/v2"
	"fullstack-app/server/internal/middleware"
	"fullstack-app/server/internal/router"
	routerv1 "fullstack-app/server/internal/router/v1"
	routerv2 "fullstack-app/server/internal/router/v2"
	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/alicebob/miniredis/v2"
	"github.com/cloudwego/hertz/pkg/app"
	hertzserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
)

func newJWTManager(t *testing.T) *jwtpkg.Manager {
	t.Helper()

	miniRedis := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: miniRedis.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return jwtpkg.NewManager(rdb, 24, 168)
}

func performRequest(t *testing.T, h *hertzserver.Hertz, method, path string, headers map[string]string) *app.RequestContext {
	t.Helper()

	c := h.NewContext()
	c.Request.SetMethod(method)
	c.Request.SetRequestURI(path)
	c.Request.Header.SetHost("example.test")
	for key, value := range headers {
		c.Request.Header.Set(key, value)
	}

	h.ServeHTTP(context.Background(), c)
	return c
}

func responseBody(t *testing.T, c *app.RequestContext) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(c.Response.Body(), &body); err != nil {
		t.Fatalf("decode response body %q: %v", c.Response.Body(), err)
	}
	return body
}

func responseCode(t *testing.T, c *app.RequestContext) int {
	t.Helper()

	body := responseBody(t, c)
	code, ok := body["code"].(float64)
	if !ok {
		t.Fatalf("response body has no numeric code: %#v", body)
	}
	return int(code)
}

func TestHealthRoute(t *testing.T) {
	h := hertzserver.New()
	jwtManager := newJWTManager(t)

	router.Setup(
		h,
		jwtManager,
		nil,
		&routerv1.Handlers{},
		&routerv2.Handlers{Test: v2.NewTestHandler()},
	)

	c := performRequest(t, h, http.MethodGet, "/health", nil)
	if got, want := c.Response.StatusCode(), http.StatusOK; got != want {
		t.Fatalf("GET /health status = %d, want %d", got, want)
	}

	body := responseBody(t, c)
	if got, want := body["status"], "ok"; got != want {
		t.Fatalf("GET /health status field = %#v, want %#v", got, want)
	}
}

func TestV2PingRoute(t *testing.T) {
	h := hertzserver.New()
	jwtManager := newJWTManager(t)

	router.Setup(
		h,
		jwtManager,
		nil,
		&routerv1.Handlers{},
		&routerv2.Handlers{Test: v2.NewTestHandler()},
	)

	c := performRequest(t, h, http.MethodGet, "/api/v2/test/ping", nil)
	if got, want := c.Response.StatusCode(), http.StatusOK; got != want {
		t.Fatalf("GET /api/v2/test/ping status = %d, want %d", got, want)
	}

	body := responseBody(t, c)
	if got, want := body["code"], float64(0); got != want {
		t.Fatalf("GET /api/v2/test/ping code = %#v, want %#v", got, want)
	}
	if got, want := body["data"], "pong from v2"; got != want {
		t.Fatalf("GET /api/v2/test/ping data = %#v, want %#v", got, want)
	}
}

func TestV1ProtectedRouteIsRegisteredAndRequiresAuthentication(t *testing.T) {
	h := hertzserver.New()
	jwtManager := newJWTManager(t)

	router.Setup(
		h,
		jwtManager,
		nil,
		&routerv1.Handlers{},
		&routerv2.Handlers{Test: v2.NewTestHandler()},
	)

	// /api/v1/users 是受保护路由。401 而不是 404，说明 v1 路由已注册且挂载了 JWTAuth。
	c := performRequest(t, h, http.MethodGet, "/api/v1/users", nil)
	if got, want := c.Response.StatusCode(), http.StatusUnauthorized; got != want {
		t.Fatalf("GET /api/v1/users without token status = %d, want %d", got, want)
	}
	if got, want := responseCode(t, c), 401; got != want {
		t.Fatalf("GET /api/v1/users without token code = %d, want %d", got, want)
	}
}

func TestJWTAuth(t *testing.T) {
	jwtManager := newJWTManager(t)
	h := hertzserver.New()
	h.GET("/protected", middleware.JWTAuth(jwtManager), func(_ context.Context, c *app.RequestContext) {
		c.JSON(http.StatusOK, map[string]any{
			"user_id":  middleware.GetUserID(c),
			"login_at": middleware.GetLoginAt(c),
		})
	})

	tests := []struct {
		name          string
		authorization string
		status        int
		code          int
	}{
		{
			name:   "missing token",
			status: http.StatusUnauthorized,
			code:   401,
		},
		{
			name:          "invalid authorization scheme",
			authorization: "Basic abc",
			status:        http.StatusUnauthorized,
			code:          401,
		},
		{
			name:          "invalid bearer token",
			authorization: "Bearer token-that-does-not-exist",
			status:        http.StatusUnauthorized,
			code:          10003,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{}
			if tt.authorization != "" {
				headers[middleware.HeaderAuthorization] = tt.authorization
			}

			c := performRequest(t, h, http.MethodGet, "/protected", headers)
			if got := c.Response.StatusCode(); got != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", got, tt.status, c.Response.Body())
			}
			if got := responseCode(t, c); got != tt.code {
				t.Fatalf("code = %d, want %d; body = %s", got, tt.code, c.Response.Body())
			}
		})
	}

	loginAt := int64(1700000000)
	pair, err := jwtManager.GenerateTokenPair(42, loginAt)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	c := performRequest(t, h, http.MethodGet, "/protected", map[string]string{
		middleware.HeaderAuthorization: middleware.BearerPrefix + pair.AccessToken,
	})
	if got, want := c.Response.StatusCode(), http.StatusOK; got != want {
		t.Fatalf("valid token status = %d, want %d; body = %s", got, want, c.Response.Body())
	}

	body := responseBody(t, c)
	if got, want := body["user_id"], float64(42); got != want {
		t.Fatalf("valid token user_id = %#v, want %#v", got, want)
	}
	if got, want := body["login_at"], float64(loginAt); got != want {
		t.Fatalf("valid token login_at = %#v, want %#v", got, want)
	}
}

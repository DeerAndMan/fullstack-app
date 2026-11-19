package test

import (
	"context"
	"encoding/json"
	"testing"

	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

func callResponse(t *testing.T, fn func(context.Context, *app.RequestContext)) (int, response.Body) {
	t.Helper()
	ctx := context.Background()
	c := app.NewContext(0)
	fn(ctx, c)

	resp := c.GetResponse()
	var body response.Body
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		t.Fatalf("decode response body %q: %v", resp.Body(), err)
	}
	return resp.StatusCode(), body
}

func TestResponseOK(t *testing.T) {
	status, body := callResponse(t, func(ctx context.Context, c *app.RequestContext) {
		response.OK(ctx, c, map[string]any{"name": "alice", "count": 2})
	})

	if status != 200 {
		t.Errorf("status = %d, want 200", status)
	}
	if body.Code != 0 || body.Message != "success" {
		t.Errorf("unexpected body metadata: %+v", body)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("Data type = %T, want map[string]any", body.Data)
	}
	if data["name"] != "alice" || data["count"] != float64(2) {
		t.Errorf("Data = %#v, want name=alice and count=2", data)
	}
}

func TestResponseOKWithMessage(t *testing.T) {
	status, body := callResponse(t, func(ctx context.Context, c *app.RequestContext) {
		response.OKWithMessage(ctx, c, "已创建")
	})

	if status != 200 || body.Code != 0 || body.Message != "已创建" {
		t.Errorf("unexpected response: status=%d body=%+v", status, body)
	}
	if body.Data != nil {
		t.Errorf("Data = %#v, want nil", body.Data)
	}
}

func TestResponseFailAndFailWithMessage(t *testing.T) {
	status, body := callResponse(t, func(ctx context.Context, c *app.RequestContext) {
		response.Fail(ctx, c, errcode.ErrNotFound)
	})
	if status != 404 || body.Code != errcode.ErrNotFound.Code || body.Message != errcode.ErrNotFound.Message || body.Data != nil {
		t.Errorf("unexpected Fail response: status=%d body=%+v", status, body)
	}

	status, body = callResponse(t, func(ctx context.Context, c *app.RequestContext) {
		response.FailWithMessage(ctx, c, errcode.ErrUnauthorized, "令牌已失效")
	})
	if status != 401 || body.Code != errcode.ErrUnauthorized.Code || body.Message != "令牌已失效" || body.Data != nil {
		t.Errorf("unexpected FailWithMessage response: status=%d body=%+v", status, body)
	}
}

func TestResponseOKWithPage(t *testing.T) {
	status, body := callResponse(t, func(ctx context.Context, c *app.RequestContext) {
		response.OKWithPage(ctx, c, []string{"a", "b"}, 8, 2, 2)
	})

	if status != 200 || body.Code != 0 || body.Message != "success" {
		t.Errorf("unexpected response metadata: status=%d body=%+v", status, body)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("Data type = %T, want map[string]any", body.Data)
	}
	if data["total"] != float64(8) || data["page"] != float64(2) || data["pageSize"] != float64(2) {
		t.Errorf("unexpected page metadata: %#v", data)
	}
	list, ok := data["list"].([]any)
	if !ok || len(list) != 2 || list[0] != "a" || list[1] != "b" {
		t.Errorf("list = %#v, want [a b]", data["list"])
	}
}

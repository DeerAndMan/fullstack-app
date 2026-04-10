package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

func RequestID() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		reqID := string(c.GetHeader(RequestIDHeader))
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Header(RequestIDHeader, reqID)
		c.Set("request_id", reqID)
		c.Next(ctx)
	}
}

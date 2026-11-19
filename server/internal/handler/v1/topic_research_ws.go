package v1

import (
	"context"
	"time"

	jwtpkg "fullstack-app/server/pkg/jwt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

// Browsers authenticate in the first frame so access tokens stay out of URLs and access logs.
func (h *TopicResearchHandler) Stream(manager *jwtpkg.Manager) app.HandlerFunc {
	upgrader := websocket.HertzUpgrader{CheckOrigin: func(*app.RequestContext) bool { return true }}
	return func(_ context.Context, c *app.RequestContext) {
		_ = upgrader.Upgrade(c, func(conn *websocket.Conn) {
			defer conn.Close()
			conn.SetReadLimit(8192)
			_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			var request struct {
				Token string
				RunID uint
			}
			if conn.ReadJSON(&request) != nil {
				return
			}
			claims, err := manager.ParseAccessToken(request.Token)
			if err != nil || request.RunID == 0 {
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			_ = conn.SetReadDeadline(time.Time{})
			go func() { _, _, _ = conn.ReadMessage(); cancel() }()
			ticker := time.NewTicker(300 * time.Millisecond)
			defer ticker.Stop()
			last := ""
			draft := ""
			for {
				// Read the persisted snapshot on subscribe and when the worker has finished.
				phase := h.research.Progress(request.RunID)
				if last == "" || phase == "" {
					run, err := h.research.Run(ctx, claims.UserID, request.RunID)
					if err != nil {
						return
					}
					if run.Status != "running" {
						_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
						_ = conn.WriteJSON(map[string]any{"type": "result", "run": run})
						return
					}
				}
				if next := h.research.Draft(request.RunID); next != "" && next != draft {
					draft = next
					_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
					if conn.WriteJSON(map[string]string{"type": "reply", "text": draft}) != nil {
						return
					}
				}
				if phase == "" {
					phase = "正在准备查询"
				}
				if last != phase {
					_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
					if conn.WriteJSON(map[string]string{"type": "progress", "message": phase}) != nil {
						return
					}
					last = phase
				}
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		})
	}
}

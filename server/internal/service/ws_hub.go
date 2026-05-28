package service

import (
	"encoding/json"
	"sync"

	"github.com/hertz-contrib/websocket"
)

// wsClient 单个连接的封装。
// 每个连接附带一把写锁，避免多 goroutine 并发写同一个 conn 导致协议错乱。
type wsClient struct {
	conn   *websocket.Conn
	writeM sync.Mutex
}

func (c *wsClient) writeJSON(data []byte) error {
	c.writeM.Lock()
	defer c.writeM.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// WsHub 维护所有活跃 WebSocket 连接，支持主动广播 / 单播。
// key 可以是 connID、userID 或业务自定义标识。
type WsHub struct {
	mu      sync.RWMutex
	clients map[string]*wsClient
}

// NewWsHub 创建一个空的 Hub。
func NewWsHub() *WsHub {
	return &WsHub{clients: make(map[string]*wsClient)}
}

// Register 注册一条新连接，已存在同名 key 会替换并关闭旧连接。
func (h *WsHub) Register(id string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.clients[id]; ok {
		_ = old.conn.Close()
	}
	h.clients[id] = &wsClient{conn: conn}
}

// Unregister 移除连接并关闭。
func (h *WsHub) Unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[id]; ok {
		_ = c.conn.Close()
		delete(h.clients, id)
	}
}

// Broadcast 主动广播给所有在线连接。
// 写失败的连接会被立即清理，避免连接泄漏。
func (h *WsHub) Broadcast(payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// 先在读锁下复制快照，避免持锁期间执行 IO
	h.mu.RLock()
	snapshot := make(map[string]*wsClient, len(h.clients))
	for id, c := range h.clients {
		snapshot[id] = c
	}
	h.mu.RUnlock()

	var stale []string
	for id, c := range snapshot {
		if err := c.writeJSON(data); err != nil {
			stale = append(stale, id)
		}
	}

	if len(stale) > 0 {
		h.mu.Lock()
		for _, id := range stale {
			if c, ok := h.clients[id]; ok {
				_ = c.conn.Close()
				delete(h.clients, id)
			}
		}
		h.mu.Unlock()
	}
}

// SendTo 主动推送给指定连接，连接不存在时返回 nil（静默忽略）。
func (h *WsHub) SendTo(id string, payload any) error {
	h.mu.RLock()
	c, ok := h.clients[id]
	h.mu.RUnlock()
	if !ok {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.writeJSON(data)
}

// Count 返回当前在线连接数，用于观测。
func (h *WsHub) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

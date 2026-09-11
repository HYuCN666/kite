package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub 管理所有 WebSocket 连接并广播消息。
type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

// NewHub 创建 Hub。
func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]bool)}
}

// Register 登记一个新连接。
func (h *Hub) Register(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

// Unregister 移除连接。
func (h *Hub) Unregister(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// Broadcast 向所有客户端广播消息。
func (h *Hub) Broadcast(msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		_ = c.WriteMessage(websocket.TextMessage, data)
	}
}

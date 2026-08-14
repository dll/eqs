package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/middleware"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ==================== 站内实时推送（SSE） ====================
// 说明：不依赖第三方推送服务。H5/Admin 使用 EventSource 订阅 /api/v1/notify/stream，
// 小程序/App 无 EventSource，由客户端轮询 unread-count 兜底（见 client realtime util）。
// 业务通知统一经 CreateNotification → publishNotification 推送到在线用户。

// notifyHub 用户级通知广播器
type notifyHub struct {
	mu    sync.RWMutex
	conns map[uint]map[chan string]struct{}
}

var hub = &notifyHub{conns: make(map[uint]map[chan string]struct{})}

// subscribe 注册连接，返回接收通道与取消函数
func (h *notifyHub) subscribe(userID uint) (<-chan string, func()) {
	ch := make(chan string, 16)
	h.mu.Lock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[chan string]struct{})
	}
	h.conns[userID][ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		// 仅从 hub 移除，不 close(ch)：publish 的 select-default 与移除互斥，
		// 避免"发送到已关闭通道"panic；通道由 GC 回收。
		h.mu.Lock()
		if m, ok := h.conns[userID]; ok {
			delete(m, ch)
			if len(m) == 0 {
				delete(h.conns, userID)
			}
		}
		h.mu.Unlock()
	}
}

// publish 向用户全部在线连接推送事件（非阻塞，连接慢则丢弃该条）
func (h *notifyHub) publish(userID uint, event string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.conns[userID] {
		select {
		case ch <- event:
		default:
			// 通道已满（消费端过慢），跳过该条避免阻塞业务
		}
	}
}

// NotifyStream 通知实时流（SSE）
// GET /api/v1/notify/stream?token=<JWT>（EventSource 无法携带请求头，令牌走查询参数）
func NotifyStream(c *gin.Context) {
	cfg := config.Get()
	token := c.Query("token")
	if token == "" {
		badRequest(c, "缺少 token")
		return
	}
	userID, _, err := middleware.ValidateToken(cfg, token)
	if err != nil {
		unauthorized(c, "token 无效")
		return
	}

	// 初始推送一次未读数（含连接确认）
	unread := notificationUnread(userID)
	initData, _ := json.Marshal(gin.H{"type": "connected", "unread": unread})
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 关闭 nginx 缓冲，保证实时性
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		// 不支持流式写（非 HTTP/1.1），直接返回
		return
	}
	fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", initData)
	flusher.Flush()

	ch, cancel := hub.subscribe(userID)
	defer cancel()

	// 心跳：每 25s 发送注释行，保持代理连接不超时
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	done := c.Request.Context().Done()
	for {
		select {
		case <-done:
			return
		case evt := <-ch:
			// 客户端断开后通道已关闭，读零值直接退出
			if evt == "" {
				return
			}
			fmt.Fprintf(c.Writer, "event: notification\ndata: %s\n\n", evt)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(c.Writer, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}

// notificationUnread 查询用户未读通知数
func notificationUnread(userID uint) int64 {
	var n int64
	model.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, 0).Count(&n)
	return n
}

// publishNotification 业务通知创建后推送实时事件（与 CreateNotification 联动）
func publishNotification(userID uint, title, content, ntype string) {
	if userID == 0 {
		return
	}
	unread := notificationUnread(userID)
	data, err := json.Marshal(gin.H{
		"type":    "notification",
		"title":   title,
		"content": content,
		"ntype":   ntype,
		"unread":  unread,
	})
	if err != nil {
		return
	}
	hub.publish(userID, string(data))
}

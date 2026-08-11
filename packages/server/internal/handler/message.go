package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

type SendMessageRequest struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	OrderID    uint   `json:"order_id"`
	Content    string `json:"content" binding:"required"`
}

// SendMessage 发送消息
// POST /api/v1/message/send
func SendMessage(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	msg := model.Message{
		SenderID:   userID,
		ReceiverID: req.ReceiverID,
		OrderID:    req.OrderID,
		Content:    req.Content,
	}
	if err := model.DB.Create(&msg).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": msg})
}

// ListMessages 消息列表（会话内，双向）
// GET /api/v1/message/list?other_id=&order_id=
func ListMessages(c *gin.Context) {
	userID := c.GetUint("user_id")

	query := model.DB.Model(&model.Message{}).
		Where("sender_id = ? OR receiver_id = ?", userID, userID)

	if otherID := c.Query("other_id"); otherID != "" {
		if oid, err := parseUint(otherID); err == nil {
			query = query.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
				userID, oid, oid, userID)
		}
	}
	if orderID := c.Query("order_id"); orderID != "" {
		if oid, err := parseUint(orderID); err == nil {
			query = query.Where("order_id = ?", oid)
		}
	}

	var messages []model.Message
	query.Order("created_at DESC").Limit(100).Find(&messages)
	ok(c, gin.H{"messages": messages})
}

// MarkMessageRead 标记消息已读
// PUT /api/v1/message/read/:id
func MarkMessageRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	msgID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "消息ID无效")
		return
	}

	var msg model.Message
	if err := model.DB.First(&msg, msgID).Error; err != nil {
		notFound(c, "消息不存在")
		return
	}
	if msg.ReceiverID != userID {
		forbidden(c, "无权操作该消息")
		return
	}
	model.DB.Model(&msg).Update("is_read", 1)
	ok(c, gin.H{"message": "已读"})
}

// GetMessage 消息详情（仅收发双方）
func GetMessage(c *gin.Context) {
	userID := c.GetUint("user_id")
	msgID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "消息ID无效")
		return
	}

	var msg model.Message
	if err := model.DB.First(&msg, msgID).Error; err != nil {
		notFound(c, "消息不存在")
		return
	}
	if msg.SenderID != userID && msg.ReceiverID != userID && !isAdmin(c) {
		forbidden(c, "无权查看该消息")
		return
	}
	ok(c, gin.H{"message": msg})
}

// DeleteMessage 删除消息（仅发送方可删除自己的消息；接收方删除标记隐藏）
func DeleteMessage(c *gin.Context) {
	userID := c.GetUint("user_id")
	msgID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "消息ID无效")
		return
	}

	var msg model.Message
	if err := model.DB.First(&msg, msgID).Error; err != nil {
		notFound(c, "消息不存在")
		return
	}
	if msg.SenderID != userID && !isAdmin(c) {
		forbidden(c, "仅发送方可删除消息")
		return
	}
	if err := model.DB.Delete(&msg).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"message": "消息已删除"})
}

// ListNotifications 通知列表
// GET /api/v1/notification/list
func ListNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")

	var notifications []model.Notification
	model.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(50).Find(&notifications)
	ok(c, gin.H{"notifications": notifications})
}

// NotificationUnreadCount 通知未读数（消息触达增强：供前端角标轮询）
// GET /api/v1/notification/unread-count
func NotificationUnreadCount(c *gin.Context) {
	userID := c.GetUint("user_id")

	var unread int64
	model.DB.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, 0).Count(&unread)
	ok(c, gin.H{"unread": unread})
}

// MarkNotificationRead 标记通知已读
// PUT /api/v1/notification/read/:id
func MarkNotificationRead(c *gin.Context) {
	userID := c.GetUint("user_id")

	notifID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "通知ID无效")
		return
	}

	var notif model.Notification
	if err := model.DB.First(&notif, notifID).Error; err != nil {
		notFound(c, "通知不存在")
		return
	}
	if notif.UserID != userID {
		forbidden(c, "无权操作该通知")
		return
	}
	model.DB.Model(&notif).Update("is_read", 1)
	ok(c, gin.H{"message": "已读"})
}

// CreateNotification 内部工具：创建通知（供业务 handler 调用）
func CreateNotification(userID uint, title, content, ntype string) {
	if userID == 0 {
		return
	}
	notif := model.Notification{
		UserID:  userID,
		Title:   title,
		Content: content,
		Type:    ntype,
	}
	model.DB.Create(&notif)
}

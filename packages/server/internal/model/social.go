package model

import "time"

// Review 评价
type Review struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	OrderID    uint      `json:"order_id" gorm:"index"`
	ReviewerID uint      `json:"reviewer_id"`
	RevieweeID uint      `json:"reviewee_id"`
	Rating     int       `json:"rating"` // 1-5
	Content    string    `json:"content" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
}

// Message 消息
type Message struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SenderID   uint      `json:"sender_id"`
	ReceiverID uint      `json:"receiver_id" gorm:"index"`
	OrderID    uint      `json:"order_id"`
	Content    string    `json:"content" gorm:"type:text"`
	IsRead     int       `json:"is_read" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
}

// Notification 通知
type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Title     string    `json:"title" gorm:"size:200"`
	Content   string    `json:"content" gorm:"type:text"`
	Type      string    `json:"type" gorm:"size:50"` // order/system/payment
	IsRead    int       `json:"is_read" gorm:"default:0"`
	CreatedAt time.Time `json:"created_at"`
}
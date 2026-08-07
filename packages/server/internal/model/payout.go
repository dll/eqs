package model

import "time"

type Payout struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index"`
	OrderID   uint      `json:"order_id" gorm:"index"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type" gorm:"size:20"` // recharge, withdraw, payment, refund
	Status    int       `json:"status" gorm:"default:0"` // 0:处理中 1:成功 2:失败
	CreatedAt time.Time `json:"created_at"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
	Order     Order     `json:"order" gorm:"foreignKey:OrderID"`
}

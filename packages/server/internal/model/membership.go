package model

import "time"

// 会员等级（代码内定义权益，避免配置漂移）
// free: 免费版；silver: 高级会员；gold: 企业会员
const (
	MemberLevelFree   = "free"
	MemberLevelSilver = "silver"
	MemberLevelGold   = "gold"
)

// MembershipOrder 会员开通/续费记录
type MembershipOrder struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Level     string    `json:"level" gorm:"size:20"`
	Months    int       `json:"months"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status" gorm:"size:20;default:paid"` // paid（模拟支付即生效）/ pending
	CreatedAt time.Time `json:"created_at"`
}

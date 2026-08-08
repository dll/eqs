package model

import "time"

// User 用户表
// user_type: 1-甲方 2-服务方 3-管理员 4-评审专家
type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Phone       string    `json:"phone" gorm:"uniqueIndex;size:20"`
	WxOpenID    *string   `json:"wx_openid" gorm:"uniqueIndex;size:100"`
	WxUnionID   *string   `json:"wx_unionid" gorm:"size:100"`
	UserType    int       `json:"user_type" gorm:"default:1"`
	CompanyName string    `json:"company_name" gorm:"size:100"`
	CreditScore float64   `json:"credit_score" gorm:"default:100"`
	Status      int       `json:"status" gorm:"default:1"` // 0:禁用 1:正常
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
package model

import "time"

type User struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Phone       string    `json:"phone" gorm:"uniqueIndex;size:20"`
	UserType    int       `json:"user_type"` // 1:甲方 2:服务方 3:管理员
	CompanyName string    `json:"company_name" gorm:"size:100"`
	CreditScore float64   `json:"credit_score" gorm:"default:100"`
	Status      int       `json:"status" gorm:"default:1"` // 0:禁用 1:正常
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

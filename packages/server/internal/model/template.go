package model

import "time"

// DeliveryTemplate 标准交付模板
// status: draft/active/retired
type DeliveryTemplate struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	ServiceType string     `json:"service_type" gorm:"size:50"`
	Name        string     `json:"name" gorm:"size:100"`
	Version     string     `json:"version" gorm:"size:20"`
	FileID      uint       `json:"file_id"`
	Checklist   string     `json:"checklist" gorm:"type:text"`
	Status      string     `json:"status" gorm:"size:20;default:draft"`
	EffectiveAt *time.Time `json:"effective_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ContractTemplate 合同模板
type ContractTemplate struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ServiceType string    `json:"service_type" gorm:"size:50"`
	Name        string    `json:"name" gorm:"size:100"`
	Version     string    `json:"version" gorm:"size:20"`
	Content     string    `json:"content" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"index"`
	Action     string    `json:"action" gorm:"size:50"`
	TargetType string    `json:"target_type" gorm:"size:50"`
	TargetID   uint      `json:"target_id"`
	Detail     string    `json:"detail" gorm:"type:text"`
	IP         string    `json:"ip" gorm:"size:50"`
	CreatedAt  time.Time `json:"created_at"`
}
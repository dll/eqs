package model

import "time"

// Contract 合同表
// status: draft/signing/signed/voided
type Contract struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	OrderID          uint       `json:"order_id" gorm:"uniqueIndex"`
	TemplateID       uint       `json:"template_id"`
	TemplateVersion  string     `json:"template_version" gorm:"size:20"`
	ContractNo       string     `json:"contract_no" gorm:"uniqueIndex;size:50"`
	ContractFileID   uint       `json:"contract_file_id"`
	SignProvider     string     `json:"sign_provider" gorm:"size:50;default:mock"`
	SignFlowID       string     `json:"sign_flow_id" gorm:"uniqueIndex;size:100"`
	Status           string     `json:"status" gorm:"size:20;default:draft"`
	SignedAt         *time.Time `json:"signed_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
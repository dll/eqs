package model

import "time"

// Bid 报价表
// status: submitted/selected/rejected/withdrawn
type Bid struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	ProjectID      uint      `json:"project_id" gorm:"index"`
	SupplierID     uint      `json:"supplier_id" gorm:"index"`
	Amount         float64   `json:"amount"`
	ServiceDays    int       `json:"service_days"`
	ProposalFileID uint      `json:"proposal_file_id"`
	Status         string    `json:"status" gorm:"size:20;default:submitted"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Supplier       User      `json:"supplier" gorm:"foreignKey:SupplierID"`
}
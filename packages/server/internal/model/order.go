package model

import "time"

type Order struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	ProjectID  uint       `json:"project_id" gorm:"index"`
	SupplierID uint       `json:"supplier_id" gorm:"index"`
	Amount     float64    `json:"amount"`
	Status     int        `json:"status" gorm:"default:0"` // 0:待签约 1:进行中 2:待验收 3:已完成 4:纠纷中
	SignedAt   *time.Time `json:"signed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	Project    Project    `json:"project" gorm:"foreignKey:ProjectID"`
	Supplier   User       `json:"supplier" gorm:"foreignKey:SupplierID"`
}

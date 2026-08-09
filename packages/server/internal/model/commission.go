package model

import "time"

// CommissionRecord 平台佣金记录（SRC-BIZ-01：按合同额收取 5%-10% 佣金，比例可配置）
type CommissionRecord struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	OrderID     uint       `json:"order_id" gorm:"index"`
	SupplierID  uint       `json:"supplier_id" gorm:"index"`
	ProjectID   uint       `json:"project_id"`
	Amount      float64    `json:"amount"`   // 合同金额
	Rate        float64    `json:"rate"`     // 佣金比例（%）
	Commission  float64    `json:"commission"` // 佣金金额
	Status      string     `json:"status" gorm:"size:20;default:pending"` // pending/collected
	CollectedAt *time.Time `json:"collected_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

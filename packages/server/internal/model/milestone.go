package model

import "time"

// PaymentMilestone 付款节点
// status: pending/submitted/accepted/rejected/disputed/settled
type PaymentMilestone struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	OrderID         uint       `json:"order_id" gorm:"index"`
	Name            string     `json:"name" gorm:"size:100"`
	Sequence        int        `json:"sequence"`
	Ratio           float64    `json:"ratio"`
	Amount          float64    `json:"amount"`
	AcceptanceDueAt *time.Time `json:"acceptance_due_at"`
	Status          string     `json:"status" gorm:"size:20;default:pending"`
	AcceptedBy      uint       `json:"accepted_by"`
	AcceptedAt      *time.Time `json:"accepted_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
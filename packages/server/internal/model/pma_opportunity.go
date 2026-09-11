package model

import "time"

// PMAOpportunity is an internal, offline-captured opportunity record.
// It is deliberately separate from ArchivedProject and the Project/Bid/Order
// transaction domain; this MVP only supports capture and triage.
type PMAOpportunity struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Title      string    `json:"title" gorm:"size:200;not null"`
	Source     string    `json:"source" gorm:"size:30;not null;default:offline"`
	Status     string    `json:"status" gorm:"size:20;not null;default:new"`
	Summary    string    `json:"summary" gorm:"type:text"`
	OwnerOrg   string    `json:"owner_org" gorm:"size:150"`
	OperatorID uint      `json:"operator_id" gorm:"index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Operator   User      `json:"operator" gorm:"foreignKey:OperatorID"`
}

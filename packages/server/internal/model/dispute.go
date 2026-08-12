package model

import "time"

// Dispute 争议案件
// status: evidence/review/mediation/reconsideration/closed
type Dispute struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	OrderID           uint       `json:"order_id" gorm:"index"`
	MilestoneID       uint       `json:"milestone_id"`
	InitiatorID       uint       `json:"initiator_id"`
	Reason            string     `json:"reason" gorm:"type:text"`
	Claim             string     `json:"claim" gorm:"type:text"`
	Status            string     `json:"status" gorm:"size:30;default:evidence"`
	ExpertResult      string     `json:"expert_result" gorm:"type:text"`
	ResolutionType    string     `json:"resolution_type" gorm:"size:30"` // settlement/agreement/award/judgment
	ResolutionFileID  uint       `json:"resolution_file_id"`
	CreatedAt         time.Time  `json:"created_at"`
	ClosedAt          *time.Time `json:"closed_at"`
}

// DisputeEvidence 争议证据
type DisputeEvidence struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DisputeID  uint      `json:"dispute_id" gorm:"index"`
	UserID     uint      `json:"user_id"`
	FileID     uint      `json:"file_id"`
	SHA256     string    `json:"sha256" gorm:"size:64"`
	Content    string    `json:"content" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
}

// DisputeExpertAssignment 争议专家指派
type DisputeExpertAssignment struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	DisputeID       uint       `json:"dispute_id" gorm:"index"`
	ExpertUserID    uint       `json:"expert_user_id"`
	ConflictDeclared int       `json:"conflict_declared" gorm:"default:0"`
	RecusalStatus   string     `json:"recusal_status" gorm:"size:20;default:not_required"`
	Opinion         string     `json:"opinion" gorm:"type:text"`
	Vote            string     `json:"vote" gorm:"size:20"` // support_client/support_supplier/partial
	SubmittedAt     *time.Time `json:"submitted_at"`
	CreatedAt       time.Time  `json:"created_at"`
}
package model

import "time"

// Order 订单表
// status: 0-待签约 1-进行中 2-待验收 3-已完成 4-纠纷中 5-已下架 6-已取消
type Order struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	ProjectID     uint       `json:"project_id" gorm:"index"`
	SupplierID    uint       `json:"supplier_id" gorm:"index"`
	SelectedBidID uint       `json:"selected_bid_id"`
	Amount        float64    `json:"amount"`
	Status        int        `json:"status" gorm:"default:0"`
	// V9：订单扩展字段
	Remark      string     `json:"remark" gorm:"size:500"`     // 甲方备注
	ExpectDays  int        `json:"expect_days"`                 // 期望工期（天）
	CancelReason string    `json:"cancel_reason" gorm:"size:300"` // 取消原因
	CancelledAt *time.Time `json:"cancelled_at"`                 // 取消时间
	SignedAt    *time.Time `json:"signed_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Project     Project    `json:"project" gorm:"foreignKey:ProjectID"`
	Supplier    User       `json:"supplier" gorm:"foreignKey:SupplierID"`
}
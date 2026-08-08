package model

import "time"

// PaymentTransaction 支付与结算流水
// type: payment/settlement/refund/freeze/unfreeze
// status: 0-处理中 1-成功 2-失败
type PaymentTransaction struct {
	ID                   uint      `json:"id" gorm:"primaryKey"`
	UserID               uint      `json:"user_id" gorm:"index"`
	OrderID              uint      `json:"order_id"`
	MilestoneID          uint      `json:"milestone_id"`
	Amount               float64   `json:"amount"`
	Type                 string    `json:"type" gorm:"size:20"`
	Channel              string    `json:"channel" gorm:"size:20;default:mock"` // wechat/bank/other_licensed_provider
	ExternalTransactionID string   `json:"external_transaction_id" gorm:"uniqueIndex;size:100"`
	Status               int       `json:"status" gorm:"default:0"`
	CreatedAt            time.Time `json:"created_at"`
}
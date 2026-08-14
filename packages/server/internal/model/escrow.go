package model

import "time"

// EscrowLedger 资金托管台账（资金托管/结算/争议冻结-释放的可对账记录）
// 语义：甲方付款进入平台托管 → 节点验收结算释放（release）→ 争议冻结（freeze）→ 结案释放/退款（release/refund）。
// 真实支付通道接入前，Mock 通道下的托管状态同样以台账为准，保证资金可对账、可审计。
// type: freeze（冻结）/ release（释放给服务方）/ refund（退款给甲方）
type EscrowLedger struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OrderID     uint      `json:"order_id" gorm:"index"`
	MilestoneID uint      `json:"milestone_id"`
	DisputeID   uint      `json:"dispute_id"`
	ActorUserID uint      `json:"actor_user_id"` // 触发方（平台/甲方/服务方）
	Type        string    `json:"type" gorm:"size:20"` // freeze/release/refund
	Amount      float64   `json:"amount"`
	Note        string    `json:"note" gorm:"size:300"`
	CreatedAt   time.Time `json:"created_at"`
}

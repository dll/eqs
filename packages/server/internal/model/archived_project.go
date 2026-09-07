package model

import "time"

// ArchivedProject 存量项目归档（已有/线下承接项目直接收录 EQS，不走“发需求→招标→竞标”主流程）
//
// 定位与红线（见 docs/L0-平台底座基线/既有交易域保护清单）：
//   - 平行于竞标/订单交易状态机，作「登记/补录/只读展示/公司看板聚合」载体；
//   - 不进 Bid/Order 交易主链路、不触发招标、不牵动资金/托管/里程碑结算；
//   - 仅做加法：不改动 Project/Bid/Order 既有模型、落库值与状态语义。
//
// status_phase: 存量项目所处阶段
//
//	preparing    = 筹备/立项前期
//	in_progress  = 在建/执行中
//	delivered    = 已交付
//	settled      = 已结算
//	liquidation  = 清收/清算中
//
// source: 收录来源（manual 普通单条补录 / batch 批量导入 / migrate 从外系统迁移建档）
type ArchivedProject struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title" gorm:"size:200" binding:"required"`
	Description string `json:"description" gorm:"type:text"`
	Address     string `json:"address" gorm:"size:300"`
	ServiceType string `json:"service_type" gorm:"size:50"` // cost/supervision/geotech/design 等（复用既有枚举）
	ProjectType string `json:"project_type" gorm:"size:50"`
	// OwnerOrg 甲方/所属组织或来源单位名称
	OwnerOrg string  `json:"owner_org" gorm:"size:150"`
	Amount   float64 `json:"amount"` // 存档金额（只读归档，不计入支付/托管/结算）
	// OperatorID 补录人（平台/OPC 侧操作账号）
	OperatorID  uint       `json:"operator_id" gorm:"index"`
	StatusPhase string     `json:"status_phase" gorm:"size:20;default:preparing"`
	Source      string     `json:"source" gorm:"size:20;default:manual"`
	StartDate   *time.Time `json:"start_date"`
	PlannedEnd  *time.Time `json:"planned_end"`
	ActualEnd   *time.Time `json:"actual_end"`
	// ExtRef 可选：存量项目后续若转正式 Project/Order，记录关联 ID（默认不自动转主流程）
	ExtProjectID uint `json:"ext_project_id" gorm:"default:0"`
	// Remark 运营备注（补录时完善项目信息用）
	Remark string `json:"remark" gorm:"type:text"`
	// DeletedAt 软删除（NULL=正常；非空=已逻辑删除，便于审计撤回）
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	Operator  User       `json:"operator" gorm:"foreignKey:OperatorID"`
}

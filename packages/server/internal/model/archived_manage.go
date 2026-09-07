package model

import "time"

// ===========================================================================
// 公司/OPC 视角的「在管项目」推进配套（挂载在 ArchivedProject 上）
//
// 背景与红线：
//   - CEO 第一单要求：OPC/公司侧可获得「阶段 → 里程碑/进度/风险/验收/决策待办」的管理闭环视图。
//   - 存量项目归档 ArchivedProject 走「不走招标」轨道，与交易域 Project/Order 平行、不混不碰；
//     因此其"里程碑/进度/风险/待办"作为**关联在归档项目下**的推进记录，绝不触碰交易状态机/PaymentMilestone。
//   - 只做加法：新增表/字段不改既有 28 模型成员语义，不动 bid/order/payment/escrow/dispute/AI。
// ===========================================================================

// ArchivedMilestone 归档项目里程碑（阶段点推进）：status todo=待/进行中,done=完成
type ArchivedMilestone struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	ArchivedProjectID uint       `json:"archived_project_id" gorm:"index"`
	Name              string     `json:"name" gorm:"size:150" binding:"required"`
	SortOrder         int        `json:"sort_order" gorm:"default:0"`
	DueAt             *time.Time `json:"due_at"`
	Status            string     `json:"status" gorm:"size:20;default:todo"` // todo / done
	Notes             string     `json:"notes" gorm:"type:text"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// ArchivedRisk 归档项目风险条目：level low/medium/high；status open/mitigated/closed
type ArchivedRisk struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	ArchivedProjectID uint      `json:"archived_project_id" gorm:"index"`
	Level             string    `json:"level" gorm:"size:16;default:medium"`
	Title             string    `json:"title" gorm:"size:200" binding:"required"`
	Status            string    `json:"status" gorm:"size:16;default:open"` // open / mitigated / closed
	Mitigation        string    `json:"mitigation" gorm:"type:text"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ArchivedTodo 归档项目决策/验收待办：kind decision/acceptance/review/other
type ArchivedTodo struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	ArchivedProjectID uint      `json:"archived_project_id" gorm:"index"`
	Title             string    `json:"title" gorm:"size:200" binding:"required"`
	Kind              string    `json:"kind" gorm:"size:20;default:other"`
	Status            string    `json:"status" gorm:"size:16;default:open"` // open / done
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

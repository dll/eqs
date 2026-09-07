package model

import "time"

// ===========================================================================
// 组织层 + 公司侧身份（真人 09-07 放行新增表；红线不动）
//
// 用途：让「甲方/服务方可按所属组织隔离」「OPC/公司监管只读+审批可独立授权」
//      「公司侧 AI agent 用可识别服务身份而非共享人类账号」成为可能。
// 设计原则：**不改变**既有 User.user_type / RABC 判定 / AdminGroup 语义，
//     只做加法：新增组织模型与成员/角色映射，鉴权在既有 owner 判定基础上"并上"组织/角色面。
// ===========================================================================

// Organization 组织（公司/单位）。甲方可一个组织多个账号，服务方可一个单位多账号。
type Organization struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:150;uniqueIndex"`
	Code      string    `json:"code" gorm:"size:60;index"`
	OwnerID   uint      `json:"owner_id"`                          // 创建人(组织主/管理员，任意已登录用户可建自己的组织)
	Type      string    `json:"type" gorm:"size:20;default:other"` // client甲方/supplier服务方/internal内部/other
	Status    int       `json:"status" gorm:"default:1"`           // 1正常 0停用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrgMember 组织成员：role owner/admin/member
type OrgMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	OrgID     uint      `json:"org_id" gorm:"uniqueIndex:idx_org_user"`
	UserID    uint      `json:"user_id" gorm:"uniqueIndex:idx_org_user"`
	Role      string    `json:"role" gorm:"size:20;default:member"` // owner/admin/member
	Status    int       `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AgentPrincipal 公司侧 AI agent 服务身份（区别于人类登录账号）
// 说明：公司要求 agent 不共享人类 admin；agent 用此表登记，映射唯一调用主与固定作用域。
// scope JSON: {"read_domains":["project","archive"],"write_domains":[],"roles":["company_ops"]}
type AgentPrincipal struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	AgentKey  string    `json:"agent_key" gorm:"size:64;uniqueIndex"` // 例 leader-eqs / ccit-pma
	Display   string    `json:"display" gorm:"size:120"`
	OrgID     uint      `json:"org_id" gorm:"default:0"` // 若属某组织可填
	ScopeJSON string    `json:"scope_json" gorm:"type:text"`
	Status    int       `json:"status" gorm:"default:1"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

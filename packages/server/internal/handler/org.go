package handler

import (
	"strings"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ===========================================================================
// 组织层（Organization / OrgMember / AgentPrincipal）
// 真人 09-07 放行新增表；红线不动交易域。
// 新增鉴权面 isOpcScope / isOrgManager（供后续 & 存量聚合使用），不改既有 user_type 判定。
// ===========================================================================

// isOpcScope 是否为公司内部监管/审批位（platform 内部组织成员 or uType3 平台运营）
// OPC 独立角色 = 一个 type=internal 的组织内的 owner/admin（公司侧）；非个人 admin 自助可得。
func isOpcScope(c *gin.Context) bool {
	if isAdmin(c) {
		return true
	}
	userID := c.GetUint("user_id")
	var cnt int64
	model.DB.Model(&model.OrgMember{}).
		Joins("JOIN organizations ON organizations.id = org_members.org_id").
		Where("org_members.user_id = ? AND org_members.role IN ? AND organizations.type = ?", userID, []string{"owner", "admin"}, "internal").
		Count(&cnt)
	return cnt > 0
}

// --- Organization 自助（非平台运营也可建自己的组织，用于一个单位多账号归属） ---

func OrgCreate(c *gin.Context) {
	userID := c.GetUint("user_id")
	var in struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code"`
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	name := in.Name
	typ := in.Type
	if typ != "client" && typ != "supplier" && typ != "internal" {
		typ = "other"
	}
	org := model.Organization{Name: name, Code: in.Code, OwnerID: userID, Type: typ, Status: 1}
	if err := model.DB.Create(&org).Error; err != nil {
		if isDuplicateError(err) {
			badRequest(c, "组织名已存在")
			return
		}
		serverError(c, err)
		return
	}
	// 创建者默认 owner
	om := model.OrgMember{OrgID: org.ID, UserID: userID, Role: "owner", Status: 1}
	model.DB.Create(&om)
	WriteAudit(c, "org.create", "organization", org.ID, gin.H{"name": in.Name})
	ok(c, gin.H{"organization": org})
}

func OrgJoin(c *gin.Context) {
	userID := c.GetUint("user_id")
	var in struct {
		OrgID uint   `json:"org_id" binding:"required"`
		Role  string `json:"role"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	role := in.Role
	if role != "member" && role != "admin" && role != "owner" {
		role = "member"
	}
	if role != "member" {
		// 提升为 admin/owner 需原 owner/admin 操作；本自助路径仅允许普通加入
		role = "member"
	}
	var om model.OrgMember
	if err := model.DB.Where("org_id = ? AND user_id = ?", in.OrgID, userID).First(&om).Error; err == nil {
		ok(c, gin.H{"org_member": om, "message": "已是成员"})
		return
	}
	om = model.OrgMember{OrgID: in.OrgID, UserID: userID, Role: role, Status: 1}
	if err := model.DB.Create(&om).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "org.join", "organization", in.OrgID, gin.H{})
	ok(c, gin.H{"org_member": om})
}

func OrgMyList(c *gin.Context) {
	userID := c.GetUint("user_id")
	var orgs []model.Organization
	model.DB.Joins("JOIN org_members ON org_members.org_id = organizations.id").
		Where("org_members.user_id = ?", userID).Find(&orgs)
	if orgs == nil {
		orgs = []model.Organization{}
	}
	ok(c, gin.H{"organizations": orgs})
}

func OrgMembers(c *gin.Context) {
	userID := c.GetUint("user_id")
	var isMem int64
	model.DB.Model(&model.OrgMember{}).Where("org_id = ? AND user_id = ?", c.Param("oid"), userID).Count(&isMem)
	if isMem == 0 && !isAdmin(c) {
		forbidden(c, "非本组织成员")
		return
	}
	var list []model.OrgMember
	if err := model.DB.Where("org_id = ?", c.Param("oid")).Find(&list).Error; err != nil {
		serverError(c, err)
		return
	}
	ids := make([]uint, 0, len(list))
	for _, m := range list {
		ids = append(ids, m.UserID)
	}
	var users []model.User
	if len(ids) > 0 {
		model.DB.Where("id IN ?", ids).Find(&users)
	}
	byID := map[uint]model.User{}
	for _, u := range users {
		byID[u.ID] = u
	}
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		u := byID[m.UserID]
		out = append(out, gin.H{"id": m.ID, "user_id": m.UserID, "role": m.Role, "phone": model.MaskPhone(u.Phone)})
	}
	ok(c, gin.H{"members": out})
}

// --- AgentPrincipal 平台运营登记（admin） ---
func AdminListAgentPrincipals(c *gin.Context) {
	var list []model.AgentPrincipal
	model.DB.Order("id ASC").Find(&list)
	if list == nil {
		list = []model.AgentPrincipal{}
	}
	ok(c, gin.H{"agent_principals": list})
}

func AdminRegisterAgentPrincipal(c *gin.Context) {
	var in struct {
		AgentKey  string `json:"agent_key" binding:"required"`
		Display   string `json:"display"`
		OrgID     uint   `json:"org_id"`
		ScopeJSON string `json:"scope_json"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	ap := model.AgentPrincipal{AgentKey: in.AgentKey, Display: in.Display, OrgID: in.OrgID, ScopeJSON: in.ScopeJSON, Status: 1}
	if err := model.DB.Create(&ap).Error; err != nil {
		if isDuplicateError(err) {
			badRequest(c, "该 agent_key 已登记")
			return
		}
		serverError(c, err)
		return
	}
	WriteAudit(c, "agent.register", "agent_principal", ap.ID, gin.H{"agent_key": in.AgentKey})
	ok(c, gin.H{"agent_principal": ap})
}

// isDuplicateError 识别唯一键冲突（驱动差异容忍）
func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "UNIQUE") || strings.Contains(s, "Duplicate") || strings.Contains(s, "duplicate") || strings.Contains(s, "already exists") || strings.Contains(s, "Error 1062")
}

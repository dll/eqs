package handler

import (
	"net/http"
	"strings"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

var pmaOpportunityStatuses = map[string]bool{"new": true, "triaged": true, "closed": true}

var pmaOpportunityStatusTransitions = map[string]map[string]bool{
	"new":     {"triaged": true, "closed": true},
	"triaged": {"closed": true},
}

// requirePMAOpportunityAccess limits this internal MVP to active internal members.
// Platform administrators may inspect system data, but do not implicitly become PMA operators.
func requirePMAOpportunityAccess(c *gin.Context) bool {
	if pmaInternalMember(c, false) {
		return true
	}
	forbidden(c, "无PMA商机权限")
	return false
}

// CreatePMAOpportunity records an offline opportunity only; it never creates a project or order.
// POST /api/v1/pma/opportunities
func CreatePMAOpportunity(c *gin.Context) {
	if !requirePMAOpportunityAccess(c) {
		return
	}
	var in struct {
		Title    string `json:"title" binding:"required"`
		Source   string `json:"source"`
		Status   string `json:"status"`
		Summary  string `json:"summary"`
		OwnerOrg string `json:"owner_org"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || in.Title == "" {
		badRequest(c, "商机标题不能为空")
		return
	}
	if in.Source == "" {
		in.Source = "offline"
	}
	if in.Status == "" {
		in.Status = "new"
	}
	if !pmaOpportunityStatuses[in.Status] {
		badRequest(c, "商机状态无效")
		return
	}
	op := model.PMAOpportunity{Title: in.Title, Source: in.Source, Status: in.Status, Summary: in.Summary, OwnerOrg: in.OwnerOrg, OperatorID: c.GetUint("user_id")}
	if err := model.DB.Create(&op).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "pma.opportunity.create", "pma_opportunity", op.ID, gin.H{"title": op.Title, "source": op.Source, "status": op.Status})
	ok(c, gin.H{"opportunity": op, "message": "线下商机已录入"})
}

// ListPMAOpportunities returns a filtered internal opportunity list.
// GET /api/v1/pma/opportunities?status=&source=&q=&page=&size=
func ListPMAOpportunities(c *gin.Context) {
	if !requirePMAOpportunityAccess(c) {
		return
	}
	page, size := parsePage(c)
	db := model.DB.Model(&model.PMAOpportunity{})
	if v := c.Query("status"); v != "" {
		db = db.Where("status = ?", v)
	}
	if v := c.Query("source"); v != "" {
		db = db.Where("source = ?", v)
	}
	if v := c.Query("q"); v != "" {
		like := "%" + v + "%"
		db = db.Where("title LIKE ? OR owner_org LIKE ? OR summary LIKE ?", like, like, like)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		serverError(c, err)
		return
	}
	var list []model.PMAOpportunity
	if err := db.Preload("Operator").Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		serverError(c, err)
		return
	}
	if list == nil {
		list = []model.PMAOpportunity{}
	}
	ok(c, gin.H{"opportunities": list, "count": total, "page": page, "size": size})
}

// UpdatePMAOpportunityStatus applies the approved new->triaged->closed flow and records a complete audit entry.
// PUT /api/v1/pma/opportunities/:id/status
func UpdatePMAOpportunityStatus(c *gin.Context) {
	if !requirePMAOpportunityAccess(c) {
		return
	}
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "商机ID无效")
		return
	}
	var in struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || !pmaOpportunityStatuses[in.Status] || strings.TrimSpace(in.Reason) == "" {
		badRequest(c, "商机状态和变更原因不能为空，且状态必须有效")
		return
	}
	var op model.PMAOpportunity
	if err := model.DB.First(&op, id).Error; err != nil {
		notFound(c, "商机不存在")
		return
	}
	old := op.Status
	if old == in.Status || !pmaOpportunityStatusTransitions[old][in.Status] {
		fail(c, http.StatusConflict, "invalid_status_transition", "不允许该商机状态转换")
		return
	}
	result := model.DB.Model(&model.PMAOpportunity{}).
		Where("id = ? AND status = ?", op.ID, old).
		Update("status", in.Status)
	if result.Error != nil {
		serverError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		fail(c, http.StatusConflict, "status_conflict", "商机状态已被其他请求变更")
		return
	}
	op.Status = in.Status
	WriteAudit(c, "pma.opportunity.status", "pma_opportunity", op.ID, gin.H{
		"from": old, "to": in.Status, "reason": strings.TrimSpace(in.Reason),
		"operator_principal": gin.H{"type": "user", "user_id": c.GetUint("user_id")},
	})
	ok(c, gin.H{"opportunity": op})
}

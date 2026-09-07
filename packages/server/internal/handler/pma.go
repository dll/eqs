package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// requirePMARead 允许平台管理员或启用中的 internal 组织成员查看项目管理视图。
func requirePMARead(c *gin.Context) bool {
	if isAdmin(c) || pmaInternalMember(c, false) {
		return true
	}
	forbidden(c, "无PMA项目查看权限")
	return false
}

func requirePMADecision(c *gin.Context) bool {
	if isAdmin(c) || pmaInternalMember(c, true) {
		return true
	}
	forbidden(c, "仅OPC或平台管理员可处理项目决策")
	return false
}

func pmaInternalMember(c *gin.Context, ownersOnly bool) bool {
	uid := c.GetUint("user_id")
	q := model.DB.Model(&model.OrgMember{}).
		Joins("JOIN organizations ON organizations.id = org_members.org_id").
		Where("org_members.user_id = ? AND org_members.status = 1 AND organizations.type = ? AND organizations.status = 1", uid, "internal")
	if ownersOnly {
		q = q.Where("org_members.role IN ?", []string{"owner", "admin"})
	}
	var n int64
	q.Count(&n)
	return n > 0
}

// PMAListProjects 返回现有存量项目，不创建机会、不生成提案、不改变交易状态。
func PMAListProjects(c *gin.Context) {
	if !requirePMARead(c) {
		return
	}
	var projects []model.ArchivedProject
	q := model.DB.Where("deleted_at IS NULL")
	if phase := c.Query("phase"); phase != "" {
		q = q.Where("status_phase = ?", phase)
	}
	if err := q.Order("id DESC").Find(&projects).Error; err != nil {
		serverError(c, err)
		return
	}
	if projects == nil {
		projects = []model.ArchivedProject{}
	}
	ok(c, gin.H{"projects": projects})
}

// PMAProjectOverview 聚合项目、里程碑、风险与决策待办。
func PMAProjectOverview(c *gin.Context) {
	if !requirePMARead(c) {
		return
	}
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	ap, found := getArchivedProjectOr404(c, id)
	if !found {
		return
	}
	ok(c, ArchivedProjectOverviewFields(ap))
}

// PMADecideTodo 仅允许 OPC 内部组织 owner/admin 关闭或重新打开决策待办。
func PMADecideTodo(c *gin.Context) {
	if !requirePMADecision(c) {
		return
	}
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	tid, err := parseUint(c.Param("tid"))
	if err != nil {
		badRequest(c, "待办ID无效")
		return
	}
	var todo model.ArchivedTodo
	if err := model.DB.Where("id = ? AND archived_project_id = ?", tid, id).First(&todo).Error; err != nil {
		notFound(c, "待办不存在")
		return
	}
	var in struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || (in.Status != "open" && in.Status != "done") {
		badRequest(c, "决策状态无效")
		return
	}
	if err := model.DB.Model(&todo).Update("status", in.Status).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "pma.todo.decide", "archived_project", id, gin.H{"todo_id": tid, "status": in.Status})
	ok(c, gin.H{"todo": todo, "message": "决策待办已更新"})
}

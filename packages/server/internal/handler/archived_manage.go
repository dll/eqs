package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ===========================================================================
// 存量/在管项目推进闭环（ArchivedProject 下的 里程碑/风险/待办）
// 真人 2026-09-07 放行新增表；红线：不改交易 Project/Order 状态机、不动资金/dispute/AI，
// 纯加法，做公司/OPC 侧「在管项目」的里程碑/进度/风险/决策待办闭环。
// 端点均挂 admin(平台/公司运营) 组，写动作 WriteAudit。
// ===========================================================================

func getArchivedProjectOr404(c *gin.Context, id uint) (*model.ArchivedProject, bool) {
	var ap model.ArchivedProject
	if err := model.DB.Where("deleted_at IS NULL").First(&ap, id).Error; err != nil {
		notFound(c, "存量项目不存在")
		return nil, false
	}
	return &ap, true
}

// ArchivedProjectOverviewFields 项目卡：项目 + 里程碑 + 风险 + 待办（只读聚合）
func ArchivedProjectOverviewFields(ap *model.ArchivedProject) gin.H {
	var ms []model.ArchivedMilestone
	model.DB.Where("archived_project_id = ?", ap.ID).Order("sort_order ASC, id ASC").Find(&ms)
	var risks []model.ArchivedRisk
	model.DB.Where("archived_project_id = ?", ap.ID).Order("id ASC").Find(&risks)
	var todos []model.ArchivedTodo
	model.DB.Where("archived_project_id = ?", ap.ID).Order("id ASC").Find(&todos)

	msDone := 0
	for _, m := range ms {
		if m.Status == "done" {
			msDone++
		}
	}
	risksOpen := 0
	for _, r := range risks {
		if r.Status != "closed" {
			risksOpen++
		}
	}
	todosOpen := 0
	for _, t := range todos {
		if t.Status != "done" {
			todosOpen++
		}
	}
	if ms == nil {
		ms = []model.ArchivedMilestone{}
	}
	if risks == nil {
		risks = []model.ArchivedRisk{}
	}
	if todos == nil {
		todos = []model.ArchivedTodo{}
	}

	progress := 0
	if len(ms) > 0 {
		progress = msDone * 100 / len(ms)
	}
	return gin.H{
		"archived_project": ap,
		"milestones":       ms,
		"risks":            risks,
		"todos":            todos,
		"summary": gin.H{
			"milestone_total": len(ms), "milestone_done": msDone,
			"risk_open": risksOpen, "risk_total": len(risks),
			"todo_open": todosOpen, "todo_total": len(todos),
			"progress_percent": progress,
		},
	}
}

// ArchiveGetOverview 项目卡概览
// GET /api/v1/admin/archive/projects/:id/overview
func ArchiveGetOverview(c *gin.Context) {
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

// ---------------------------------------------------------------------------
// 里程碑
// ---------------------------------------------------------------------------
func ArchiveAddMilestone(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	if _, ok := getArchivedProjectOr404(c, id); !ok {
		return
	}
	var in struct {
		Name      string     `json:"name" binding:"required"`
		SortOrder int        `json:"sort_order"`
		DueAt     *time.Time `json:"due_at"`
		Notes     string     `json:"notes"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	m := model.ArchivedMilestone{ArchivedProjectID: id, Name: in.Name, SortOrder: in.SortOrder, DueAt: in.DueAt, Notes: in.Notes, Status: "todo"}
	if err := model.DB.Create(&m).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.milestone.create", "archived_project", id, gin.H{"milestone_id": m.ID, "name": m.Name})
	ok(c, gin.H{"milestone": m})
}

func ArchiveUpdateMilestone(c *gin.Context) {
	id, _ := parseUint(c.Param("id"))
	mid, err := parseUint(c.Param("mid"))
	if err != nil {
		badRequest(c, "里程碑ID无效")
		return
	}
	var m model.ArchivedMilestone
	if err := model.DB.Where("id = ? AND archived_project_id = ?", mid, id).First(&m).Error; err != nil {
		notFound(c, "里程碑不存在")
		return
	}
	var in struct {
		Name      string     `json:"name"`
		Status    string     `json:"status"`
		DueAt     *time.Time `json:"due_at"`
		SortOrder *int       `json:"sort_order"`
		Notes     string     `json:"notes"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	upd := map[string]interface{}{"updated_at": time.Now()}
	if in.Name != "" {
		upd["name"] = in.Name
	}
	if in.Status == "todo" || in.Status == "done" {
		upd["status"] = in.Status
	}
	if in.SortOrder != nil {
		upd["sort_order"] = *in.SortOrder
	}
	if in.DueAt != nil {
		upd["due_at"] = in.DueAt
	}
	if in.Notes != "" {
		upd["notes"] = in.Notes
	}
	if err := model.DB.Model(&m).Updates(upd).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.milestone.update", "archived_project", id, gin.H{"milestone_id": mid, "status": in.Status})
	ok(c, gin.H{"message": "已更新"})
}

func ArchiveDeleteMilestone(c *gin.Context) {
	id, _ := parseUint(c.Param("id"))
	mid, err := parseUint(c.Param("mid"))
	if err != nil {
		badRequest(c, "里程碑ID无效")
		return
	}
	if err := model.DB.Where("id = ? AND archived_project_id = ?", mid, id).Delete(&model.ArchivedMilestone{}).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.milestone.delete", "archived_project", id, gin.H{"milestone_id": mid})
	ok(c, gin.H{"message": "已删除"})
}

// ---------------------------------------------------------------------------
// 风险
// ---------------------------------------------------------------------------
func ArchiveAddRisk(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	if _, ok := getArchivedProjectOr404(c, id); !ok {
		return
	}
	var in struct {
		Level      string `json:"level"`
		Title      string `json:"title" binding:"required"`
		Mitigation string `json:"mitigation"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	lv := in.Level
	if lv != "high" && lv != "medium" {
		lv = "low"
	}
	r := model.ArchivedRisk{ArchivedProjectID: id, Level: lv, Title: in.Title, Mitigation: in.Mitigation, Status: "open"}
	if err := model.DB.Create(&r).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.risk.create", "archived_project", id, gin.H{"risk_id": r.ID, "level": lv})
	ok(c, gin.H{"risk": r})
}

func ArchiveUpdateRisk(c *gin.Context) {
	id, _ := parseUint(c.Param("id"))
	rid, err := parseUint(c.Param("rid"))
	if err != nil {
		badRequest(c, "风险ID无效")
		return
	}
	var r model.ArchivedRisk
	if err := model.DB.Where("id = ? AND archived_project_id = ?", rid, id).First(&r).Error; err != nil {
		notFound(c, "风险不存在")
		return
	}
	var in struct {
		Level      string `json:"level"`
		Title      string `json:"title"`
		Status     string `json:"status"`
		Mitigation string `json:"mitigation"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	upd := map[string]interface{}{"updated_at": time.Now()}
	for _, lv := range []string{"low", "medium", "high"} {
		if in.Level == lv {
			upd["level"] = lv
		}
	}
	for _, st := range []string{"open", "mitigated", "closed"} {
		if in.Status == st {
			upd["status"] = st
		}
	}
	if in.Title != "" {
		upd["title"] = in.Title
	}
	if in.Mitigation != "" {
		upd["mitigation"] = in.Mitigation
	}
	if err := model.DB.Model(&r).Updates(upd).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.risk.update", "archived_project", id, gin.H{"risk_id": rid})
	ok(c, gin.H{"message": "已更新"})
}

func ArchiveDeleteRisk(c *gin.Context) {
	id, _ := parseUint(c.Param("id"))
	rid, err := parseUint(c.Param("rid"))
	if err != nil {
		badRequest(c, "风险ID无效")
		return
	}
	if err := model.DB.Where("id = ? AND archived_project_id = ?", rid, id).Delete(&model.ArchivedRisk{}).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.risk.delete", "archived_project", id, gin.H{"risk_id": rid})
	ok(c, gin.H{"message": "已删除"})
}

// ---------------------------------------------------------------------------
// 决策/验收待办
// ---------------------------------------------------------------------------
func ArchiveAddTodo(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	if _, ok := getArchivedProjectOr404(c, id); !ok {
		return
	}
	var in struct {
		Title string `json:"title" binding:"required"`
		Kind  string `json:"kind"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	kind := in.Kind
	if kind != "decision" && kind != "acceptance" && kind != "review" {
		kind = "other"
	}
	t := model.ArchivedTodo{ArchivedProjectID: id, Title: in.Title, Kind: kind, Status: "open"}
	if err := model.DB.Create(&t).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.todo.create", "archived_project", id, gin.H{"todo_id": t.ID, "kind": kind})
	ok(c, gin.H{"todo": t})
}

func ArchiveUpdateTodo(c *gin.Context) {
	id, _ := parseUint(c.Param("id"))
	tid, err := parseUint(c.Param("tid"))
	if err != nil {
		badRequest(c, "待办ID无效")
		return
	}
	var t model.ArchivedTodo
	if err := model.DB.Where("id = ? AND archived_project_id = ?", tid, id).First(&t).Error; err != nil {
		notFound(c, "待办不存在")
		return
	}
	var in struct {
		Title  string `json:"title"`
		Status string `json:"status"`
		Kind   string `json:"kind"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "参数错误")
		return
	}
	upd := map[string]interface{}{"updated_at": time.Now()}
	if in.Title != "" {
		upd["title"] = in.Title
	}
	if in.Status == "open" || in.Status == "done" {
		upd["status"] = in.Status
	}
	if in.Kind != "" {
		upd["kind"] = in.Kind
	}
	if err := model.DB.Model(&t).Updates(upd).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.todo.update", "archived_project", id, gin.H{"todo_id": tid, "status": in.Status})
	ok(c, gin.H{"message": "已更新"})
}

func ArchiveDeleteTodo(c *gin.Context) {
	id, _ := parseUint(c.Param("id"))
	tid, err := parseUint(c.Param("tid"))
	if err != nil {
		badRequest(c, "待办ID无效")
		return
	}
	if err := model.DB.Where("id = ? AND archived_project_id = ?", tid, id).Delete(&model.ArchivedTodo{}).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.todo.delete", "archived_project", id, gin.H{"todo_id": tid})
	ok(c, gin.H{"message": "已删除"})
}

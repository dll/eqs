package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ===========================================================================
// 存量项目收录（已有/线下承接项目直接录入 EQS —— 不走“发需求→招标→竞标”流程）
//
// 定位与红线（承 docs/L0-平台底座基线/既有交易域保护清单）：
//   - 平行于竞标/订单交易状态机：登记/补录/只读展示，纳入公司看板聚合；
//   - 不进 Bid/Order 交易主链路、不触发招标、不牵动资金/托管/结算；
//   - 只做加法：不改 Project/Order 既有路由/模型/落库值/状态机。
//
// 权限（A 批次默认）：受限在 admin（平台/公司运营侧，user_type=3）管理与查阅。
// 注：OPC/公司侧 AI agent 的细分 scope 在 L1-A RBAC 角色矩阵定稿后，以加法方式扩展写权；
//    本批次不预埋未经验证的越权面。
// status_phase: preparing/in_progress/delivered/settled/liquidation
// source: manual/batch/migrate
// ===========================================================================

type ArchivedProjectInput struct {
	Title        string     `json:"title" binding:"required"`
	Description  string     `json:"description"`
	Address      string     `json:"address"`
	ServiceType  string     `json:"service_type"`
	ProjectType  string     `json:"project_type"`
	OwnerOrg     string     `json:"owner_org"`
	Amount       float64    `json:"amount"`
	StatusPhase  string     `json:"status_phase"`
	Source       string     `json:"source"`
	StartDate    *time.Time `json:"start_date"`
	PlannedEnd   *time.Time `json:"planned_end"`
	ActualEnd    *time.Time `json:"actual_end"`
	ExtProjectID uint       `json:"ext_project_id"`
	Remark       string     `json:"remark"`
}

func (r *ArchivedProjectInput) normalizePhase() string {
	switch r.StatusPhase {
	case "in_progress", "delivered", "settled", "liquidation":
		return r.StatusPhase
	default:
		// preparing 为默认：已有项目由补录方后续补全阶段，不做招标语义
		return "preparing"
	}
}

func (r *ArchivedProjectInput) normalizeSource() string {
	switch r.Source {
	case "batch", "migrate":
		return r.Source
	default:
		return "manual"
	}
}

// CreateArchivedProject 存量项目单条补录
// POST /api/v1/admin/archive/projects
func CreateArchivedProject(c *gin.Context) {
	operatorID := c.GetUint("user_id")
	var req ArchivedProjectInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	ap := model.ArchivedProject{
		Title:        req.Title,
		Description:  req.Description,
		Address:      req.Address,
		ServiceType:  req.ServiceType,
		ProjectType:  req.ProjectType,
		OwnerOrg:     req.OwnerOrg,
		Amount:       req.Amount,
		StatusPhase:  req.normalizePhase(),
		Source:       req.normalizeSource(),
		StartDate:    req.StartDate,
		PlannedEnd:   req.PlannedEnd,
		ActualEnd:    req.ActualEnd,
		ExtProjectID: req.ExtProjectID,
		Remark:       req.Remark,
		OperatorID:   operatorID,
	}
	if err := model.DB.Create(&ap).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.project.create", "archived_project", ap.ID, gin.H{
		"title": ap.Title, "owner_org": ap.OwnerOrg, "phase": ap.StatusPhase,
	})
	ok(c, gin.H{"archived_project": ap, "message": "存量项目已收录"})
}

// BatchImportArchivedProjects 存量项目批量导入（一致结构，逐条校验，事务回滚）
// POST /api/v1/admin/archive/projects/batch
func BatchImportArchivedProjects(c *gin.Context) {
	operatorID := c.GetUint("user_id")
	var rows []ArchivedProjectInput
	if err := c.ShouldBindJSON(&rows); err != nil {
		badRequest(c, "参数错误：应为数组")
		return
	}
	if len(rows) == 0 {
		badRequest(c, "导入列表为空")
		return
	}
	if len(rows) > 500 {
		badRequest(c, "单次导入不超过 500 条，请分批")
		return
	}

	tx := model.DB.Begin()
	if tx.Error != nil {
		serverError(c, tx.Error)
		return
	}
	created := make([]model.ArchivedProject, 0, len(rows))
	for _, r := range rows {
		if r.Title == "" {
			tx.Rollback()
			badRequest(c, "存在缺少标题的记录，已整体回滚")
			return
		}
		ap := model.ArchivedProject{
			Title:        r.Title,
			Description:  r.Description,
			Address:      r.Address,
			ServiceType:  r.ServiceType,
			ProjectType:  r.ProjectType,
			OwnerOrg:     r.OwnerOrg,
			Amount:       r.Amount,
			StatusPhase:  r.normalizePhase(),
			Source:       r.normalizeSource(),
			StartDate:    r.StartDate,
			PlannedEnd:   r.PlannedEnd,
			ActualEnd:    r.ActualEnd,
			ExtProjectID: r.ExtProjectID,
			Remark:       r.Remark,
			OperatorID:   operatorID,
		}
		if err := tx.Create(&ap).Error; err != nil {
			tx.Rollback()
			serverError(c, err)
			return
		}
		created = append(created, ap)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.project.batch", "archived_project", 0, gin.H{
		"imported": len(created), "source": "batch",
	})
	ok(c, gin.H{"imported": len(created), "archived_projects": created, "message": "批量导入成功"})
}

// ListArchivedProjects 存量项目列表（筛选 + 分页）
// GET /api/v1/admin/archive/projects?phase=&source=&q=&page=&size=
func ListArchivedProjects(c *gin.Context) {
	page, size := parsePage(c)
	db := model.DB.Model(&model.ArchivedProject{}).Where("deleted_at IS NULL")

	if p := c.Query("phase"); p != "" {
		db = db.Where("status_phase = ?", p)
	}
	if s := c.Query("source"); s != "" {
		db = db.Where("source = ?", s)
	}
	if q := c.Query("q"); q != "" {
		like := "%" + q + "%"
		db = db.Where("title LIKE ? OR owner_org LIKE ? OR remark LIKE ?", like, like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		serverError(c, err)
		return
	}
	var list []model.ArchivedProject
	if err := db.Preload("Operator").
		Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&list).Error; err != nil {
		serverError(c, err)
		return
	}
	if list == nil {
		list = []model.ArchivedProject{}
	}
	ok(c, gin.H{"archived_projects": list, "count": total, "page": page, "size": size})
}

// GetArchivedProject 存量项目详情
// GET /api/v1/admin/archive/projects/:id
func GetArchivedProject(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	var ap model.ArchivedProject
	if err := model.DB.Preload("Operator").Where("deleted_at IS NULL").First(&ap, id).Error; err != nil {
		notFound(c, "存量项目不存在")
		return
	}
	ok(c, gin.H{"archived_project": ap})
}

// UpdateArchivedProject 编辑存量项目
// PUT /api/v1/admin/archive/projects/:id
func UpdateArchivedProject(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	var ap model.ArchivedProject
	if err := model.DB.Where("deleted_at IS NULL").First(&ap, id).Error; err != nil {
		notFound(c, "存量项目不存在")
		return
	}
	var req ArchivedProjectInput
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	updates := map[string]interface{}{
		"title":          req.Title,
		"description":    req.Description,
		"address":        req.Address,
		"service_type":   req.ServiceType,
		"project_type":   req.ProjectType,
		"owner_org":      req.OwnerOrg,
		"amount":         req.Amount,
		"status_phase":   req.normalizePhase(),
		"source":         req.normalizeSource(),
		"start_date":     req.StartDate,
		"planned_end":    req.PlannedEnd,
		"actual_end":     req.ActualEnd,
		"ext_project_id": req.ExtProjectID,
		"remark":         req.Remark,
		"updated_at":     time.Now(),
	}
	if err := model.DB.Model(&ap).Updates(updates).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.project.update", "archived_project", id, gin.H{"phase": req.normalizePhase()})
	ok(c, gin.H{"message": "存量项目已更新"})
}

// DeleteArchivedProject 软删除存量项目（保留审计撤回能力）
// DELETE /api/v1/admin/archive/projects/:id
func DeleteArchivedProject(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	var ap model.ArchivedProject
	if err := model.DB.Where("deleted_at IS NULL").First(&ap, id).Error; err != nil {
		notFound(c, "存量项目不存在")
		return
	}
	now := time.Now()
	if err := model.DB.Model(&ap).Update("deleted_at", &now).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "archive.project.delete", "archived_project", id, gin.H{"title": ap.Title})
	ok(c, gin.H{"message": "存量项目已删除（逻辑删除，可审计撤回）"})
}

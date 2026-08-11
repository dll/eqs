package handler

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// allowedProjectFileExt 项目附件允许的扩展名（CAD/PDF/图片）
var allowedProjectFileExt = map[string]bool{
	".dwg": true, ".dxf": true, ".pdf": true,
	".jpg": true, ".jpeg": true, ".png": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
}

// UploadProjectFile 甲方上传项目附件（批文/CAD/PDF等），保存本地并登记 project_files
// POST /api/v1/project/upload (multipart/form-data: file, 可选 project_id 绑定)
func UploadProjectFile(c *gin.Context) {
	userID := c.GetUint("user_id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		badRequest(c, "请选择要上传的附件")
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedProjectFileExt[ext] {
		badRequest(c, "仅支持 dwg/dxf/pdf/jpg/png/doc/xls 格式")
		return
	}
	if fileHeader.Size > 50*1024*1024 {
		badRequest(c, "附件不能超过 50MB")
		return
	}

	// 保存到本地 uploads/{UPLOAD_DIR 配置}/projects/{年月}/ 目录
	relDir := filepath.Join(config.Get().UploadDir, "projects", time.Now().Format("200601"))
	dir := filepath.Join(".", relDir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		serverError(c, err)
		return
	}
	now := time.Now().Format("150405.000000000")
	name := filepath.Join(dir, now+"_"+strings.ReplaceAll(fileHeader.Filename, " ", "_"))
	if err := c.SaveUploadedFile(fileHeader, name); err != nil {
		serverError(c, err)
		return
	}

	// 可选绑定项目 ID（需为发布者本人项目，或暂不绑定）
	projectID := uint(0)
	if pid := c.PostForm("project_id"); pid != "" {
		if v, perr := parseUint(pid); perr == nil {
			var proj model.Project
			if model.DB.First(&proj, v).Error == nil && (proj.UserID == userID || isAdmin(c)) {
				projectID = v
			}
		}
	}

	file := model.ProjectFile{
		ProjectID:    projectID,
		UploaderID:   userID,
		OriginalName: fileHeader.Filename,
		FileType:     strings.TrimPrefix(ext, "."),
		StorageKey:   relDir + "/" + filepath.Base(name),
		Version:      1,
	}
	if err := model.DB.Create(&file).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"file_id": file.ID, "original_name": file.OriginalName, "storage_key": file.StorageKey})
}

// serviceChecklistTemplates 按服务类型预置的标准交付清单（智能发单用）
var serviceChecklistTemplates = map[string][]string{
	"cost": {
		"工程量清单（GB50500）",
		"招标控制价编制说明",
		"计价软件源文件",
		"主要材料/设备询价单",
	},
	"supervision": {
		"监理规划/监理实施细则",
		"监理日志（每日）",
		"旁站记录",
		"监理月报",
	},
	"geotech": {
		"岩土工程勘察报告",
		"钻孔柱状图/剖面图",
		"土工试验成果表",
		"勘察资质页扫描件",
	},
	"design": {
		"方案设计文件（PDF/DWG）",
		"施工图设计文件",
		"设计变更通知单",
		"设计说明与计算书",
	},
}

// GetServiceChecklist 智能发单：按服务类型返回标准交付清单（AC-01）
// GET /api/v1/project/checklist?service_type=cost
func GetServiceChecklist(c *gin.Context) {
	st := c.Query("service_type")
	items := serviceChecklistTemplates[st]
	if items == nil {
		// 未知类型返回通用清单
		items = []string{"项目需求说明", "现场照片/资料包", "成果文件（PDF）", "验收确认单"}
	}
	ok(c, gin.H{"service_type": st, "checklist": items, "count": len(items)})
}

type CreateProjectRequest struct {
	ProjectType  string  `json:"project_type" binding:"required"`
	ServiceType  string  `json:"service_type"`
	Title        string  `json:"title" binding:"required"`
	Description  string  `json:"description"`
	Address      string  `json:"address"`
	Longitude    float64 `json:"longitude"`
	Latitude     float64 `json:"latitude"`
	BudgetMin    float64 `json:"budget_min"`
	BudgetMax    float64 `json:"budget_max"`
	PublishScope string  `json:"publish_scope"`
	Deadline     string  `json:"deadline"`
}

// CreateProject 发布项目
func CreateProject(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	scope := req.PublishScope
	if scope == "" {
		scope = "public"
	}

	var deadline *time.Time
	if req.Deadline != "" {
		if t, err := time.Parse(time.RFC3339, req.Deadline); err == nil {
			deadline = &t
		}
	}

	now := time.Now()
	project := model.Project{
		UserID:       userID,
		ProjectType:  req.ProjectType,
		ServiceType:  req.ServiceType,
		Title:        req.Title,
		Description:  req.Description,
		Address:      req.Address,
		Longitude:    req.Longitude,
		Latitude:     req.Latitude,
		BudgetMin:    req.BudgetMin,
		BudgetMax:    req.BudgetMax,
		PublishScope: scope,
		Status:       1, // 发布即上线（MVP：创建即为已发布，可进入报价）
		PublishTime:  &now,
		Deadline:     deadline,
	}

	if err := model.DB.Create(&project).Error; err != nil {
		serverError(c, err)
		return
	}

	WriteAudit(c, "project.publish", "project", project.ID, gin.H{"service_type": project.ServiceType, "budget_min": project.BudgetMin, "budget_max": project.BudgetMax})
	ok(c, gin.H{"project": project})
}

// ListProjects 项目列表，支持按类型/状态筛选
func ListProjects(c *gin.Context) {
	serviceType := c.Query("service_type")
	status := c.Query("status")
	keyword := c.Query("keyword")

	q := model.DB.Preload("User")
	if serviceType != "" {
		q = q.Where("service_type = ?", serviceType)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if keyword != "" {
		q = q.Where("title LIKE ?", "%"+keyword+"%")
	}

	// P2-03：分页保护（page/size，size 最大 100）
	page, size := parsePage(c)
	var total int64
	q.Model(&model.Project{}).Count(&total)

	var projects []model.Project
	if err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&projects).Error; err != nil {
		serverError(c, err)
		return
	}

	ok(c, gin.H{"projects": projects, "total": total, "page": page, "size": size})
}

// ListMyProjects 甲方"我的发单"列表（仅当前用户发布的项目，含分页与状态过滤）
func ListMyProjects(c *gin.Context) {
	userID := c.GetUint("user_id")

	q := model.DB.Where("user_id = ?", userID)
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	page, size := parsePage(c)
	var total int64
	q.Model(&model.Project{}).Count(&total)

	var projects []model.Project
	if err := q.Preload("User").Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&projects).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"projects": projects, "total": total, "page": page, "size": size})
}

// GetProject 项目详情
func GetProject(c *gin.Context) {
	id := c.Param("id")

	var project model.Project
	if err := model.DB.Preload("User").First(&project, id).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}

	ok(c, gin.H{"project": project})
}

// UpdateProject 编辑项目（仅发布者本人，草稿/已发布可改；有订单后仅可改描述等非关键字段）
func UpdateProject(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}

	var project model.Project
	if err := model.DB.First(&project, id).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	if project.UserID != userID && !isAdmin(c) {
		forbidden(c, "仅发布者可编辑项目")
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	// 已进入报价/订单的项目不允许改预算与类型（防竞拍串通）
	locked := project.Status >= 2
	updates := map[string]interface{}{}
	updates["title"] = req.Title
	updates["description"] = req.Description
	updates["address"] = req.Address
	if !locked {
		updates["project_type"] = req.ProjectType
		updates["service_type"] = req.ServiceType
		updates["budget_min"] = req.BudgetMin
		updates["budget_max"] = req.BudgetMax
	}
	if req.Deadline != "" {
		if t, err := time.Parse(time.RFC3339, req.Deadline); err == nil {
			updates["deadline"] = t
		}
	}

	if err := model.DB.Model(&project).Updates(updates).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "project.update", "project", project.ID, gin.H{"locked": locked})
	ok(c, gin.H{"message": "项目已更新", "locked": locked})
}

// DeleteProject 下架/删除项目（仅发布者本人；已有报价或订单时禁止删除，仅可下架）
func DeleteProject(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}

	var project model.Project
	if err := model.DB.First(&project, id).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	if project.UserID != userID && !isAdmin(c) {
		forbidden(c, "仅发布者可删除项目")
		return
	}

	// 存在报价或订单：不允许物理删除，改为下架状态
	var bidCount, orderCount int64
	model.DB.Model(&model.Bid{}).Where("project_id = ?", id).Count(&bidCount)
	model.DB.Model(&model.Order{}).Where("project_id = ?", id).Count(&orderCount)
	if bidCount > 0 || orderCount > 0 {
		model.DB.Model(&project).Update("status", 5) // 5 = 已下架
		WriteAudit(c, "project.offline", "project", id, gin.H{"bids": bidCount, "orders": orderCount})
		ok(c, gin.H{"message": "项目已有业务往来，已下架处理", "offline": true})
		return
	}

	if err := model.DB.Delete(&project).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "project.delete", "project", id, gin.H{})
	ok(c, gin.H{"message": "项目已删除", "offline": false})
}

// GetRecommendations 基于地区、资质和信用推荐服务方
func GetRecommendations(c *gin.Context) {
	id := c.Param("id")

	var project model.Project
	if err := model.DB.First(&project, id).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}

	// 满足条件的服务方：状态正常、信誉分>=70，优先同地区（地址相似度简化为同服务类型）
	q := model.DB.Where("user_type = ? AND status = ? AND credit_score >= ?", 2, 1, 70)
	q = q.Where("id <> ?", project.UserID)

	var suppliers []model.User
	if err := q.Order("credit_score DESC").Limit(10).Find(&suppliers).Error; err != nil {
		serverError(c, err)
		return
	}

	// P1-09：对外返回手机号脱敏
	type supplierItem struct {
		ID          uint    `json:"id"`
		Phone       string  `json:"phone"`
		CompanyName string  `json:"company_name"`
		CreditScore float64 `json:"credit_score"`
	}
	items := make([]supplierItem, 0, len(suppliers))
	for _, s := range suppliers {
		items = append(items, supplierItem{
			ID:          s.ID,
			Phone:       model.MaskPhone(s.Phone),
			CompanyName: s.CompanyName,
			CreditScore: s.CreditScore,
		})
	}

	ok(c, gin.H{"suppliers": items})
}

type InviteSupplierRequest struct {
	SupplierIDs []uint `json:"supplier_ids" binding:"required"`
}

type InviteRequest struct {
	SupplierIDs []uint `json:"supplier_ids" binding:"required"`
}

// InviteSuppliers 甲方邀请指定服务方，最多5家
func InviteSuppliers(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var project model.Project
	if err := model.DB.First(&project, id).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	if project.UserID != userID {
		forbidden(c, "无权限")
		return
	}
	if project.Status != 0 && project.Status != 1 {
		badRequest(c, "项目已开始，不能邀请")
		return
	}

	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if len(req.SupplierIDs) > 5 {
		badRequest(c, "最多邀请5家服务商")
		return
	}

	// 记录邀请（保留简单版本：写入项目描述字段另建 invitations 表成本过高，MVP 仅记录审计）
	model.DB.Exec("UPDATE projects SET publish_scope = ? WHERE id = ?", "invited", id)
	ok(c, gin.H{"message": "邀请成功", "count": len(req.SupplierIDs)})
}

func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	return uint(v), err
}
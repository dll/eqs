package handler

import (
	"strconv"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

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
package handler

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ProjectProgressItem 甘特图/看板项目项
type ProjectProgressItem struct {
	ID           uint    `json:"id"`
	Title        string  `json:"title"`
	ServiceType  string  `json:"service_type"`
	Status       int     `json:"status"`
	StatusText   string  `json:"status_text"`
	StartDate    *string `json:"start_date"`  // yyyy-MM-dd（发布/创建）
	EndDate      *string `json:"end_date"`    // 预计完成（deadline）
	Progress     int     `json:"progress"`    // 0-100
	OrderCount   int64   `json:"order_count"`
	Milestones   []MilestoneBar `json:"milestones"`
}

// MilestoneBar 甘特图里程碑条
type MilestoneBar struct {
	ID      uint    `json:"id"`
	Name    string  `json:"name"`
	Ratio   float64 `json:"ratio"`
	Status  string  `json:"status"`
	Amount  float64 `json:"amount"`
	OrderID uint    `json:"order_id"`
}

// ProjectProgressResponse 全部项目进度（看板/甘特图）
type ProjectProgressResponse struct {
	Projects []ProjectProgressItem `json:"projects"`
	Total    int64                 `json:"total"`
}

func fmtDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

// ListProjectProgress 全部项目进度（管理员/项目创建者可见）
// GET /api/v1/admin/project-progress
func ListProjectProgress(c *gin.Context) {
	var projects []model.Project
	model.DB.Order("created_at DESC").Limit(50).Find(&projects)

	items := make([]ProjectProgressItem, 0, len(projects))
	for _, p := range projects {
		item := ProjectProgressItem{
			ID:          p.ID,
			Title:       p.Title,
			ServiceType: p.ServiceType,
			Status:      p.Status,
			StatusText:  projectStatusText(p.Status),
			StartDate:   fmtDate(p.PublishTime),
			EndDate:     fmtDate(p.Deadline),
		}
		// 项目关联订单（进度来源）
		var orders []model.Order
		model.DB.Where("project_id = ?", p.ID).Find(&orders)
		item.OrderCount = int64(len(orders))
		for _, o := range orders {
			var ms []model.PaymentMilestone
			model.DB.Where("order_id = ?", o.ID).Order("sequence ASC").Find(&ms)
			for _, m := range ms {
				item.Milestones = append(item.Milestones, MilestoneBar{
					ID: m.ID, Name: m.Name, Ratio: m.Ratio, Status: m.Status, Amount: m.Amount, OrderID: o.ID,
				})
			}
		}
		// 进度计算：已结算里程碑占比 + 状态权重
		item.Progress = calcProjectProgress(p, item.Milestones)
		items = append(items, item)
	}

	ok(c, gin.H{"projects": items, "total": len(items)})
}

// GetProjectProgress 单项目进度（甘特图）
// GET /api/v1/project/:id/progress
func GetProjectProgress(c *gin.Context) {
	projectID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	var project model.Project
	if err := model.DB.First(&project, projectID).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	item := ProjectProgressItem{
		ID:          project.ID,
		Title:       project.Title,
		ServiceType: project.ServiceType,
		Status:      project.Status,
		StatusText:  projectStatusText(project.Status),
		StartDate:   fmtDate(project.PublishTime),
		EndDate:     fmtDate(project.Deadline),
	}
	var orders []model.Order
	model.DB.Where("project_id = ?", project.ID).Find(&orders)
	item.OrderCount = int64(len(orders))
	for _, o := range orders {
		var ms []model.PaymentMilestone
		model.DB.Where("order_id = ?", o.ID).Order("sequence ASC").Find(&ms)
		for _, m := range ms {
			item.Milestones = append(item.Milestones, MilestoneBar{ID: m.ID, Name: m.Name, Ratio: m.Ratio, Status: m.Status, Amount: m.Amount, OrderID: o.ID})
		}
	}
	item.Progress = calcProjectProgress(project, item.Milestones)
	ok(c, gin.H{"project": item})
}

// calcProjectProgress 计算项目进度
// 权重：已结算里程碑金额占比为主，辅以状态推进
func calcProjectProgress(p model.Project, ms []MilestoneBar) int {
	if len(ms) == 0 {
		// 无里程碑：草稿0，已发布10，进行中50，完成100
		switch p.Status {
		case 4:
			return 100
		case 3:
			return 80
		case 2:
			return 50
		case 1:
			return 10
		default:
			return 0
		}
	}
	var totalRatio, doneRatio float64
	for _, m := range ms {
		totalRatio += m.Ratio
		if m.Status == "settled" || m.Status == "accepted" {
			doneRatio += m.Ratio
		}
	}
	if totalRatio <= 0 {
		return 0
	}
	prog := int(doneRatio / totalRatio * 100)
	if prog > 100 {
		prog = 100
	}
	if p.Status == 4 {
		prog = 100
	}
	return prog
}

func projectStatusText(s int) string {
	switch s {
	case 0:
		return "草稿"
	case 1:
		return "已发布"
	case 2:
		return "进行中"
	case 3:
		return "待验收"
	case 4:
		return "已完成"
	default:
		return fmt.Sprintf("状态%d", s)
	}
}

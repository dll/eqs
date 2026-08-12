package handler

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ProjectProgressItem 甘特图/看板项目项
type ProjectProgressItem struct {
	ID            uint           `json:"id"`
	Title         string         `json:"title"`
	ServiceType   string         `json:"service_type"`
	Status        int            `json:"status"`
	StatusText    string         `json:"status_text"`
	StartDate     *string        `json:"start_date"` // yyyy-MM-dd（发布/创建）
	EndDate       *string        `json:"end_date"`   // 预计完成（deadline）
	Progress      int            `json:"progress"`   // 0-100
	OrderCount    int64          `json:"order_count"`
	ScheduleState string         `json:"schedule_state"` // ahead/on_time/late 提前/按时/滞后
	ScheduleText  string         `json:"schedule_text"`  // 提前/按时/滞后
	PlanProgress  int            `json:"plan_progress"`  // 按时间的计划进度（0-100）
	Stages        []ProjectStage `json:"stages"`         // 工程阶段甘特条
	Milestones    []MilestoneBar `json:"milestones"`
}

// ProjectStage 工程阶段条（甘特图用）
type ProjectStage struct {
	Name      string  `json:"name"`
	Order     int     `json:"order"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
	Progress  int     `json:"progress"`
	Done      bool    `json:"done"`
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

// fmtDateStr 日期格式化（nil 安全），返回字符串而非指针；AI 提示词等场景使用
func fmtDateStr(t *time.Time) string {
	if t == nil {
		return "未填写"
	}
	return t.Format("2006-01-02")
}

// ListProjectProgress 全部项目进度（管理员/项目创建者可见）
// GET /api/v1/admin/project-progress
func ListProjectProgress(c *gin.Context) {
	var projects []model.Project
	model.DB.Order("created_at DESC").Limit(50).Find(&projects)

	// P2-04：批量预加载，消除 N+1
	projectIDs := make([]uint, 0, len(projects))
	for _, p := range projects {
		projectIDs = append(projectIDs, p.ID)
	}
	// 一次查全部订单
	var allOrders []model.Order
	model.DB.Where("project_id IN ?", projectIDs).Find(&allOrders)
	orderByProject := map[uint][]model.Order{}
	for _, o := range allOrders {
		orderByProject[o.ProjectID] = append(orderByProject[o.ProjectID], o)
	}
	// 一次查全部里程碑
	var allMS []model.PaymentMilestone
	model.DB.Where("order_id IN ?", allOrderIDs(allOrders)).Order("sequence ASC").Find(&allMS)
	msByOrder := map[uint][]model.PaymentMilestone{}
	for _, m := range allMS {
		msByOrder[m.OrderID] = append(msByOrder[m.OrderID], m)
	}

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
		orders := orderByProject[p.ID]
		item.OrderCount = int64(len(orders))
		for _, o := range orders {
			for _, m := range msByOrder[o.ID] {
				item.Milestones = append(item.Milestones, MilestoneBar{
					ID: m.ID, Name: m.Name, Ratio: m.Ratio, Status: m.Status, Amount: m.Amount, OrderID: o.ID,
				})
			}
		}
		item.Progress = calcProjectProgress(p, item.Milestones)
		item.Stages = calcProjectStages(p, item.Progress)
		item.PlanProgress = calcPlanProgress(p)
		item.ScheduleState, item.ScheduleText = calcScheduleState(item.Progress, item.PlanProgress, p.Status)
		items = append(items, item)
	}

	ok(c, gin.H{"projects": items, "total": len(items)})
}

// allOrderIDs 提取订单 ID 列表
func allOrderIDs(orders []model.Order) []uint {
	ids := make([]uint, 0, len(orders))
	for _, o := range orders {
		ids = append(ids, o.ID)
	}
	if len(ids) == 0 {
		return []uint{0}
	}
	return ids
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
	item.Stages = calcProjectStages(project, item.Progress)
	item.PlanProgress = calcPlanProgress(project)
	item.ScheduleState, item.ScheduleText = calcScheduleState(item.Progress, item.PlanProgress, project.Status)
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

// stageDef 标准工程阶段定义（按服务类型）。每个阶段有名称与相对工期占比。
var stageDef = map[string][]struct {
	name string
	wt   float64
}{
	"cost": {
		{"需求梳理", 0.15}, {"工程量清单编制", 0.3}, {"控制价编制", 0.25}, {"成果审核", 0.2}, {"交付结算", 0.1},
	},
	"supervision": {
		{"监理规划", 0.15}, {"施工旁站", 0.35}, {"质量控制", 0.25}, {"监理月报", 0.15}, {"竣工验收", 0.1},
	},
	"geotech": {
		{"方案设计", 0.15}, {"外业钻孔", 0.3}, {"土工试验", 0.2}, {"报告编制", 0.25}, {"成果交付", 0.1},
	},
	"design": {
		{"方案设计", 0.2}, {"初步设计", 0.25}, {"施工图设计", 0.3}, {"设计交底", 0.15}, {"配合施工", 0.1},
	},
}

// defaultStageDef 未知服务类型的通用阶段
var defaultStageDef = []struct {
	name string
	wt   float64
}{
	{"需求确认", 0.15}, {"方案编制", 0.3}, {"实施推进", 0.3}, {"成果审核", 0.15}, {"交付验收", 0.1},
}

// calcProjectStages 按时间均分生成项目工程阶段（甘特图条）。
// 已发布前阶段 Done=false，进度按项目总进度分摊到各阶段。
func calcProjectStages(p model.Project, progress int) []ProjectStage {
	start := p.PublishTime
	end := p.Deadline
	now := time.Now()

	def := stageDef[p.ServiceType]
	if def == nil {
		def = defaultStageDef
	}

	// 无计划时间时给出基准区间（发布日起 30 天）
	var s, e time.Time
	if start != nil {
		s = *start
	} else {
		s = p.CreatedAt
	}
	if end != nil {
		e = *end
	} else {
		e = s.AddDate(0, 0, 30)
	}
	total := e.Sub(s).Seconds()
	if total <= 0 {
		total = 30 * 86400
	}

	stages := make([]ProjectStage, 0, len(def))
	acc := 0.0
	for i, d := range def {
		acc += d.wt
		ss := s.Add(time.Duration((acc - d.wt) / 1 * total * float64(time.Second)))
		se := s.Add(time.Duration(acc / 1 * total * float64(time.Second)))

		done := now.After(se)
		// 阶段进度：累计权重 vs 项目总进度
		stageProg := int(float64(progress) * acc)
		if stageProg > 100 {
			stageProg = 100
		}
		stages = append(stages, ProjectStage{
			Name:      d.name,
			Order:     i + 1,
			StartDate: fmtDate(&ss),
			EndDate:   fmtDate(&se),
			Progress:  stageProg,
			Done:      done,
		})
	}
	return stages
}

// calcPlanProgress 按时间流逝计算的计划进度（0-100）
// 基于发布日→截止日区间，越接近截止日计划进度越高。
func calcPlanProgress(p model.Project) int {
	start := p.PublishTime
	end := p.Deadline
	now := time.Now()
	if start == nil || end == nil {
		return 10 // 无计划时间的保守计划进度
	}
	total := end.Sub(*start).Seconds()
	if total <= 0 {
		return 100
	}
	elapsed := now.Sub(*start).Seconds()
	if elapsed <= 0 {
		return 0
	}
	prog := int(elapsed / total * 100)
	if prog > 100 {
		prog = 100
	}
	if p.Status == 4 {
		prog = 100
	}
	return prog
}

// calcScheduleState 计算进度状态：提前(ahead)/按时(on_time)/滞后(late)
// 规则：实际进度 vs 计划进度。
//   - 实际 - 计划 >= 15 → ahead（提前）
//   - 计划 - 实际 >= 15 → late（滞后）
//   - 否则 → on_time（按时）
//
// 已完成项目按 100 处理（始终按时或提前）。
func calcScheduleState(progress, planProgress, status int) (state, text string) {
	if status == 4 {
		return "ahead", "提前"
	}
	switch {
	case progress-planProgress >= 15:
		return "ahead", "提前"
	case planProgress-progress >= 15:
		return "late", "滞后"
	default:
		return "on_time", "按时"
	}
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

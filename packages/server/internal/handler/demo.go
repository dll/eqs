package handler

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func DemoSeedHandler(c *gin.Context) {
	mode := c.DefaultQuery("mode", "demo")
	if mode != "demo" && mode != "test" && mode != "training" {
		badRequest(c, "mode 参数无效")
		return
	}
	result := seedByMode(mode)
	ok(c, gin.H{"message": "演示数据生成成功", "mode": mode, "result": result})
}

func DemoCleanHandler(c *gin.Context) {
	cleanAll()
	ok(c, gin.H{"message": "演示数据已清理"})
}

func DemoToggleHandler(c *gin.Context) {
	var req struct {
		Enable *bool `json:"enable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Enable == nil {
		badRequest(c, "参数错误")
		return
	}
	enable := *req.Enable

	// 持久化到 system_configs
	var cfg model.SystemConfig
	err := model.DB.Where("config_key = ?", "demo.enabled").First(&cfg).Error
	now := time.Now()
	if err != nil {
		cfg = model.SystemConfig{
			ConfigKey:   "demo.enabled",
			ConfigValue: fmt.Sprintf("%v", enable),
			ValueType:   "bool",
			Description: "演示数据开关",
			IsPublic:    false,
			UpdatedAt:   now,
		}
		model.DB.Create(&cfg)
	} else {
		cfg.ConfigValue = fmt.Sprintf("%v", enable)
		cfg.UpdatedAt = now
		model.DB.Save(&cfg)
	}

	status := "disabled"
	if enable {
		status = "enabled"
	}
	WriteAudit(c, "admin.demo.toggle", "system", 0, gin.H{"enabled": enable})
	ok(c, gin.H{"message": "演示模式已" + status, "demo_mode": enable})
}

func DemoStatusHandler(c *gin.Context) {
	// 查询演示数据开关状态
	var cfg model.SystemConfig
	demoEnabled := false
	if err := model.DB.Where("config_key = ?", "demo.enabled").First(&cfg).Error; err == nil {
		demoEnabled = cfg.ConfigValue == "true" || cfg.ConfigValue == "1"
	}

	// 统计演示数据
	var userCount, projectCount, orderCount, disputeCount int64
	model.DB.Model(&model.User{}).Count(&userCount)
	model.DB.Model(&model.Project{}).Count(&projectCount)
	model.DB.Model(&model.Order{}).Count(&orderCount)
	model.DB.Model(&model.Dispute{}).Count(&disputeCount)

	ok(c, gin.H{
		"demo_mode":      demoEnabled,
		"user_count":     userCount,
		"project_count":  projectCount,
		"order_count":    orderCount,
		"dispute_count":  disputeCount,
	})
}

type seedResult struct {
	Users      int      `json:"users"`
	Projects   int      `json:"projects"`
	Orders     int      `json:"orders"`
	Bids       int      `json:"bids"`
	Payments   int      `json:"payments"`
	Disputes   int      `json:"disputes"`
	Qualifiers int      `json:"qualifications"`
	Attendance int      `json:"attendance"`
	Actions    []string `json:"actions"`
}

// 演示数据手机号前缀，seed 前清理，保证幂等
var demoPhones = []string{
	"13900001111", "13900002222", "13900003333",
	"13900004444", "13900005555", "13900006666", "13900007777",
}

// seedByMode 按模式生成演示数据，幂等（先清理演示用户及其关联数据）
func seedByMode(mode string) seedResult {
	r := seedResult{Actions: []string{}}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	// ===== 幂等清理：先清空全部业务表（彻底避免外键残留导致创建失败） =====
	cleanAll()

	// ===== 用户：甲方3 / 服务方3 / 管理员1 =====
	type seededUser struct {
		u      model.User
		client bool
	}
	users := []model.User{
		{Phone: "13900001111", UserType: 1, Status: 1, CreditScore: 100, CompanyName: "滁州城投建设有限公司"},
		{Phone: "13900002222", UserType: 1, Status: 1, CreditScore: 92, CompanyName: "南京市政设计研究院"},
		{Phone: "13900003333", UserType: 3, Status: 1, CreditScore: 100, CompanyName: "EQS平台运营"},
		{Phone: "13900004444", UserType: 2, Status: 1, CreditScore: 98, CompanyName: "安徽地勘勘察院"},
		{Phone: "13900005555", UserType: 2, Status: 1, CreditScore: 95, CompanyName: "江苏中衡造价咨询"},
		{Phone: "13900006666", UserType: 2, Status: 1, CreditScore: 88, CompanyName: "金陵工程监理有限公司"},
		{Phone: "13900007777", UserType: 2, Status: 1, CreditScore: 76, CompanyName: "远东设计工作室"},
	}
	for i := range users {
		if err := model.DB.Create(&users[i]).Error; err != nil {
			r.Actions = append(r.Actions, fmt.Sprintf("⚠️ 用户创建失败: %v", err))
			continue
		}
		if users[i].ID == 0 {
			r.Actions = append(r.Actions, fmt.Sprintf("⚠️ 用户 %s ID 未回填", users[i].Phone))
		}
	}
	r.Users = len(users)
	r.Actions = append(r.Actions, fmt.Sprintf("创建 %d 个演示用户（甲方3/服务方3/管理员1）", r.Users))

	clientID := users[0].ID     // 甲方
	client2ID := users[1].ID    // 甲方2
	supplierID := users[3].ID   // 地勘
	supplier2ID := users[4].ID  // 造价
	supplier3ID := users[5].ID  // 监理
	supplier4ID := users[6].ID  // 设计

	// ===== 项目（按模式数量不同） =====
	projects := []model.Project{
		{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "办公楼造价编制项目", Description: "编制办公楼土建安装工程预算", Address: "滁州市琅琊区", BudgetMin: 20000, BudgetMax: 60000, Status: 1, PublishScope: "public", PublishTime: &now},
		{UserID: clientID, ProjectType: "supervision", ServiceType: "supervision", Title: "住宅楼工程监理项目", Description: "小区1-3号楼施工监理", Address: "南京市江宁区", BudgetMin: 50000, BudgetMax: 150000, Status: 1, PublishScope: "public", PublishTime: &now},
		{UserID: clientID, ProjectType: "geotech", ServiceType: "geotech", Title: "地块地质勘察项目", Description: "商业地块岩土工程勘察", Address: "滁州市南谯区", BudgetMin: 30000, BudgetMax: 80000, Status: 1, PublishScope: "public", PublishTime: &now},
		{UserID: client2ID, ProjectType: "design", ServiceType: "design", Title: "市政道路初步设计", Description: "城市主干道方案设计与初设", Address: "合肥市包河区", BudgetMin: 40000, BudgetMax: 120000, Status: 1, PublishScope: "public", PublishTime: &now},
		{UserID: client2ID, ProjectType: "cost", ServiceType: "cost", Title: "厂房改造结算审核", Description: "既有厂房改造工程结算审核", Address: "芜湖市镜湖区", BudgetMin: 15000, BudgetMax: 45000, Status: 1, PublishScope: "invited", PublishTime: &now},
	}
	if mode != "demo" {
		// test/training 增加项目量
		projects = append(projects,
			model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "边界测试-小额预算", Description: "测试极小金额边界", Address: "测试地址", BudgetMin: 100, BudgetMax: 200, Status: 1, PublishScope: "public", PublishTime: &now},
			model.Project{UserID: client2ID, ProjectType: "geotech", ServiceType: "geotech", Title: "滨江地块二期勘察", Description: "二期场地详勘", Address: "马鞍山市", BudgetMin: 60000, BudgetMax: 150000, Status: 1, PublishScope: "public", PublishTime: &now},
			model.Project{UserID: clientID, ProjectType: "design", ServiceType: "design", Title: "教学楼建筑方案", Description: "九年一贯制学校方案设计", Address: "滁州市来安县", BudgetMin: 80000, BudgetMax: 200000, Status: 1, PublishScope: "public", PublishTime: &now},
		)
	}
	projectOK := 0
	for i := range projects {
		if err := model.DB.Create(&projects[i]).Error; err != nil {
			r.Actions = append(r.Actions, fmt.Sprintf("⚠️ 项目创建失败: %v", err))
			continue
		}
		projectOK++
	}
	r.Projects = projectOK
	r.Actions = append(r.Actions, fmt.Sprintf("创建 %d 个项目（覆盖造价/监理/地勘/设计）", r.Projects))

	// ===== 报价与订单（按模式生成不同数量的闭环订单） =====
	orderLimit := 3
	if mode == "test" {
		orderLimit = 4
	}
	if mode == "training" {
		orderLimit = 5
	}

	orderCount := 0
	for _, p := range projects {
		// 每个项目 1-3 个报价
		suppliers := []model.User{users[3], users[4], users[5], users[6]}
		for _, s := range suppliers[:1+rng.Intn(3)] {
			bidAmount := p.BudgetMin + float64(rng.Int63n(int64(p.BudgetMax-p.BudgetMin)))
			bid := model.Bid{ProjectID: p.ID, SupplierID: s.ID, Amount: float64(bidAmount), ServiceDays: 15 + rng.Intn(30), Status: "submitted"}
			model.DB.Create(&bid)
			r.Bids++
		}
		// 部分项目生成订单（完整闭环）
		if orderCount < orderLimit {
			var selected model.Bid
			model.DB.Where("project_id = ?", p.ID).Order("amount ASC").First(&selected)
			if selected.ID > 0 {
				model.DB.Model(&selected).Update("status", "selected")
				order := model.Order{ProjectID: p.ID, SupplierID: selected.SupplierID, SelectedBidID: selected.ID, Amount: selected.Amount, Status: 1}
				model.DB.Create(&order)
				orderCount++
				r.Orders++

				// 里程碑（节点金额合计=订单金额）
				model.DB.Create(&model.PaymentMilestone{OrderID: order.ID, Name: "合同预付款", Sequence: 1, Ratio: 30, Amount: selected.Amount * 0.3, Status: "submitted"})
				model.DB.Create(&model.PaymentMilestone{OrderID: order.ID, Name: "中期进度款", Sequence: 2, Ratio: 40, Amount: selected.Amount * 0.4, Status: "submitted"})
				model.DB.Create(&model.PaymentMilestone{OrderID: order.ID, Name: "验收尾款", Sequence: 3, Ratio: 30, Amount: selected.Amount * 0.3, Status: "pending"})

				// 合同（training 模式已签署）
				contractStatus := "signed"
				signedAt := &now
				if mode == "test" {
					contractStatus = "signing"
					signedAt = nil
				}
				contract := model.Contract{OrderID: order.ID, ContractNo: fmt.Sprintf("EQS-%d-%04d", time.Now().Year(), order.ID), TemplateVersion: "1.0", SignProvider: "mock", Status: contractStatus, SignedAt: signedAt}
				model.DB.Create(&contract)

				// 支付流水
				pt := float64(selected.Amount) * 0.3
				model.DB.Create(&model.PaymentTransaction{UserID: selected.SupplierID, OrderID: order.ID, Amount: pt, Type: "payment", Channel: "mock", ExternalTransactionID: fmt.Sprintf("PAY-MOCK-%d-%d", order.ID, time.Now().UnixNano()), Status: 1})

				// 交付物（training 模式含已交付）
				if mode == "training" {
					model.DB.Create(&model.Deliverable{OrderID: order.ID, MilestoneID: 0, Milestone: "预付款节点", FileURL: "/demo/report.pdf", FileName: "阶段性成果报告.pdf", Version: 1, Status: 1})
				}
				r.Payments++
				r.Actions = append(r.Actions, fmt.Sprintf("生成订单#%d 含里程碑/合同/支付", order.ID))
			}
		}
	}

	// ===== 资质（各服务方提交不同资质） =====
	qualTypes := []struct {
		sid uint
		t   string
		lv  string
	}{
		{supplierID, "工程勘察综合资质", "甲级"},
		{supplier2ID, "工程造价咨询", "甲级"},
		{supplier3ID, "工程监理综合资质", "甲级"},
		{supplier4ID, "工程设计行业资质", "乙级"},
	}
	for _, q := range qualTypes {
		model.DB.Create(&model.SupplierQualification{
			SupplierID: q.sid, QualificationType: q.t,
			CertificateNo: model.EncryptedString(fmt.Sprintf("ZZ-DEMO-%03d", rng.Intn(999))),
			Level:         q.lv, Scope: q.t, VerificationStatus: "approved",
		})
		r.Qualifiers++
	}
	r.Actions = append(r.Actions, fmt.Sprintf("生成 %d 个服务方资质（已核验）", r.Qualifiers))

	// ===== 打卡（demo/training） =====
	if mode != "test" {
		var firstOrder model.Order
		model.DB.First(&firstOrder)
		if firstOrder.ID > 0 {
			model.DB.Create(&model.AttendanceRecord{OrderID: firstOrder.ID, UserID: supplierID, CheckInAt: time.Now(), Longitude: 118.31, Latitude: 32.30})
			r.Attendance++
			r.Actions = append(r.Actions, "生成 1 条现场打卡记录")
		}
	}

	// ===== 争议（demo/training 有案例，test 无争议便于干净测试） =====
	if mode != "test" {
		var firstOrder model.Order
		model.DB.First(&firstOrder)
		if firstOrder.ID > 0 {
			dispute := model.Dispute{OrderID: firstOrder.ID, InitiatorID: clientID, Reason: "交付成果与合同约定不符", Claim: "要求重新核算并整改", Status: "evidence"}
			model.DB.Create(&dispute)
			r.Disputes++
			// 争议证据
			model.DB.Create(&model.DisputeEvidence{DisputeID: dispute.ID, UserID: clientID, FileID: 0, Content: "合同复印件与验收记录"})
			// 专家指派（user_type=4 需要专家，动态创建或跳过）
			r.Actions = append(r.Actions, fmt.Sprintf("生成 1 个争议案件#%d（含证据）", dispute.ID))
		}
	}

	// ===== 评价（demo/training 模式含已评价闭环，测试界面有评分数据） =====
	if mode != "test" {
		var completedOrder model.Order
		model.DB.Where("status = ?", 1).First(&completedOrder)
		if completedOrder.ID > 0 {
			model.DB.Create(&model.Review{OrderID: completedOrder.ID, ReviewerID: clientID, RevieweeID: supplierID, Rating: 5, Content: "配合度高，成果专业"})
			model.DB.Create(&model.Review{OrderID: completedOrder.ID, ReviewerID: supplierID, RevieweeID: clientID, Rating: 5, Content: "付款及时，沟通顺畅"})
			// 更多评分样本（供评分分布展示）
			var order2 model.Order
			model.DB.Where("status = ?", 1).Where("id <> ?", completedOrder.ID).First(&order2)
			if order2.ID > 0 {
				model.DB.Create(&model.Review{OrderID: order2.ID, ReviewerID: clientID, RevieweeID: supplier2ID, Rating: 4, Content: "专业细致，交付及时"})
				model.DB.Create(&model.Review{OrderID: order2.ID, ReviewerID: supplier2ID, RevieweeID: clientID, Rating: 5, Content: "流程规范，配合顺畅"})
			}
			r.Actions = append(r.Actions, "生成双方互评与评分样本（演示/培训用）")
		}
	}

	// ===== 交付模板（各服务类型标准模板，供模板库展示） =====
	tpls := []struct {
		svc, name, checklist string
	}{
		{"cost", "工程造价咨询交付模板", `["工程量清单（GB50500）","招标控制价编制说明","计价软件源文件","主要材料询价单"]`},
		{"supervision", "工程监理交付模板", `["监理规划/实施细则","监理日志（每日）","旁站记录","监理月报"]`},
		{"geotech", "岩土勘察交付模板", `["岩土工程勘察报告","钻孔柱状图/剖面图","土工试验成果表","勘察资质页"]`},
		{"design", "工程设计交付模板", `["方案设计文件","施工图设计文件","设计变更通知单","设计说明与计算书"]`},
	}
	for _, t := range tpls {
		model.DB.Create(&model.DeliveryTemplate{
			ServiceType: t.svc, Name: t.name, Version: "1.0",
			Checklist: t.checklist, Status: "active", EffectiveAt: &now,
		})
	}
	r.Actions = append(r.Actions, fmt.Sprintf("生成 %d 个标准交付模板", len(tpls)))

	// ===== 合同模板（电子签模板库） =====
	model.DB.Create(&model.ContractTemplate{
		ServiceType: "cost", Name: "工程造价咨询服务合同（标准版）", Version: "1.0",
		Content: "本合同由甲方与乙方就工程造价咨询服务达成……（标准条款）", Status: "active",
	})
	model.DB.Create(&model.ContractTemplate{
		ServiceType: "supervision", Name: "建设工程监理服务合同（标准版）", Version: "1.0",
		Content: "本合同由甲方与乙方就建设工程监理服务达成……（标准条款）", Status: "active",
	})
	r.Actions = append(r.Actions, "生成 2 个合同模板（电子签模板库）")

	// ===== 佣金（基于已生成订单，供结算中心展示） =====
	var txnOrders []model.Order
	model.DB.Where("status >= ?", 1).Limit(3).Find(&txnOrders)
	commCreated := 0
	for _, o := range txnOrders {
		rate := 8.0 // 8% 平台佣金
		model.DB.Create(&model.CommissionRecord{
			OrderID: o.ID, SupplierID: o.SupplierID, ProjectID: o.ProjectID,
			Amount: o.Amount, Rate: rate, Commission: o.Amount * rate / 100, Status: "pending",
		})
		commCreated++
	}
	r.Actions = append(r.Actions, fmt.Sprintf("生成 %d 条平台佣金记录（待收取）", commCreated))

	if mode == "training" {
		r.Actions = append(r.Actions, "培训模式：已含标准流程、完整闭环、结算与评价教学数据")
	}

	WriteAudit(nil, "admin.demo.seed", "system", 0, gin.H{"mode": mode, "result": r})
	return r
}

// cleanDemoUsers 清理演示手机号用户及其关联数据（保证 seed 幂等）
func cleanDemoUsers() {
	var demoUserIDs []uint
	model.DB.Model(&model.User{}).Where("phone IN ?", demoPhones).Pluck("id", &demoUserIDs)
	if len(demoUserIDs) == 0 {
		return
	}
	// 按用户清理关联数据（级联依赖顺序）
	model.DB.Where("user_id IN ?", demoUserIDs).Delete(&model.AttendanceRecord{})
	model.DB.Where("user_id IN ?", demoUserIDs).Delete(&model.Notification{})
	model.DB.Where("sender_id IN ? OR receiver_id IN ?", demoUserIDs, demoUserIDs).Delete(&model.Message{})
	model.DB.Where("reviewer_id IN ? OR reviewee_id IN ?", demoUserIDs, demoUserIDs).Delete(&model.Review{})
	model.DB.Where("supplier_id IN ?", demoUserIDs).Delete(&model.SupplierQualification{})
	model.DB.Where("supplier_id IN ?", demoUserIDs).Delete(&model.Bid{})
	model.DB.Where("user_id IN ?", demoUserIDs).Delete(&model.Project{})
	model.DB.Where("user_id IN ?", demoUserIDs).Delete(&model.User{})
}

func cleanAll() {
	model.DB.Where("1 = 1").Delete(&model.AttendanceRecord{})
	model.DB.Where("1 = 1").Delete(&model.DisputeExpertAssignment{})
	model.DB.Where("1 = 1").Delete(&model.DisputeEvidence{})
	model.DB.Where("1 = 1").Delete(&model.Dispute{})
	model.DB.Where("1 = 1").Delete(&model.PaymentTransaction{})
	model.DB.Where("1 = 1").Delete(&model.Deliverable{})
	model.DB.Where("1 = 1").Delete(&model.PaymentMilestone{})
	model.DB.Where("1 = 1").Delete(&model.Contract{})
	model.DB.Where("1 = 1").Delete(&model.FileAnnotation{})
	model.DB.Where("1 = 1").Delete(&model.ProjectFile{})
	model.DB.Where("1 = 1").Delete(&model.Bid{})
	model.DB.Where("1 = 1").Delete(&model.Order{})
	model.DB.Where("1 = 1").Delete(&model.SupplierQualification{})
	model.DB.Where("1 = 1").Delete(&model.Project{})
	model.DB.Where("1 = 1").Delete(&model.Review{})
	model.DB.Where("1 = 1").Delete(&model.Message{})
	model.DB.Where("1 = 1").Delete(&model.Notification{})
	model.DB.Where("1 = 1").Delete(&model.User{})
	model.DB.Where("1 = 1").Delete(&model.AuditLog{})
}

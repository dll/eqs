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
		badRequest(c, "mode 参数错误")
		return
	}
	result := seedByMode(mode)
	ok(c, gin.H{"message": "演示数据生成完成", "mode": mode, "result": result})
}

func DemoCleanHandler(c *gin.Context) {
	cleanAll()
	ok(c, gin.H{"message": "演示数据已清理"})
}

func DemoToggleHandler(c *gin.Context) {
	var req struct {
		Enable bool `json:"enable" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	status := "disabled"
	if req.Enable {
		status = "enabled"
	}
	ok(c, gin.H{"message": "演示模式已" + status, "demo_mode": req.Enable})
}

func DemoStatusHandler(c *gin.Context) {
	ok(c, gin.H{"demo_mode": false})
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

func seedByMode(mode string) seedResult {
	r := seedResult{Actions: []string{}}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	users := []model.User{
		{Phone: "13900001111", UserType: 1, Status: 1, CreditScore: 100, CompanyName: "演示甲方企业"},
		{Phone: "13900002222", UserType: 2, Status: 1, CreditScore: 100, CompanyName: "演示服务方企业"},
		{Phone: "13900003333", UserType: 3, Status: 1, CreditScore: 100, CompanyName: "平台管理"},
	}
	for i := range users {
		model.DB.Create(&users[i])
	}
	r.Users = len(users)
	r.Actions = append(r.Actions, fmt.Sprintf("创建 %d 个用户", r.Users))
	clientID := users[0].ID
	supplierID := users[1].ID
	now := time.Now()
	projects := []model.Project{
		{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "招标控制价编制项目", BudgetMin: 20000, BudgetMax: 60000, Status: 1, PublishScope: "public", PublishTime: &now},
		{UserID: clientID, ProjectType: "supervision", ServiceType: "supervision", Title: "住宅楼工程监理项目", BudgetMin: 50000, BudgetMax: 150000, Status: 1, PublishScope: "public", PublishTime: &now},
		{UserID: clientID, ProjectType: "geotech", ServiceType: "geotech", Title: "地块地质勘察项目", BudgetMin: 30000, BudgetMax: 80000, Status: 1, PublishScope: "public", PublishTime: &now},
	}
	if mode == "test" {
		projects = append(projects,
			model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "边界测试-最小预算", BudgetMin: 100, BudgetMax: 200, Status: 1, PublishScope: "public", PublishTime: &now},
			model.Project{UserID: clientID, ProjectType: "design", ServiceType: "design", Title: "工程设计测试", BudgetMin: 40000, BudgetMax: 120000, Status: 1, PublishScope: "public", PublishTime: &now},
		)
	}
	for i := range projects {
		model.DB.Create(&projects[i])
	}
	r.Projects = len(projects)
	r.Actions = append(r.Actions, fmt.Sprintf("创建 %d 个项目", r.Projects))
	for _, p := range projects {
		bidAmount := p.BudgetMin + float64(rng.Int63n(int64(p.BudgetMax-p.BudgetMin)))
		bid := model.Bid{ProjectID: p.ID, SupplierID: supplierID, Amount: float64(bidAmount), ServiceDays: 15 + rng.Intn(30), Status: "submitted"}
		model.DB.Create(&bid)
		r.Bids++
		if r.Orders < 3 && mode != "test" {
			model.DB.Model(&bid).Update("status", "selected")
			order := model.Order{ProjectID: p.ID, SupplierID: supplierID, SelectedBidID: bid.ID, Amount: float64(bidAmount), Status: 0}
			model.DB.Create(&order)
			r.Orders++
			model.DB.Create(&model.PaymentMilestone{OrderID: order.ID, Name: "开工款", Sequence: 1, Ratio: 40, Amount: float64(bidAmount) * 0.4, Status: "pending"})
			model.DB.Create(&model.PaymentMilestone{OrderID: order.ID, Name: "中期款", Sequence: 2, Ratio: 30, Amount: float64(bidAmount) * 0.3, Status: "pending"})
			model.DB.Create(&model.PaymentMilestone{OrderID: order.ID, Name: "验收款", Sequence: 3, Ratio: 30, Amount: float64(bidAmount) * 0.3, Status: "pending"})
			contract := model.Contract{OrderID: order.ID, ContractNo: fmt.Sprintf("EQS-%d-%d", time.Now().Year(), order.ID), TemplateVersion: "1.0", SignProvider: "mock", Status: "signed", SignedAt: &now}
			model.DB.Create(&contract)
			txn := model.PaymentTransaction{UserID: supplierID, OrderID: order.ID, Amount: float64(bidAmount), Type: "payment", Channel: "mock", ExternalTransactionID: fmt.Sprintf("PAY-MOCK-%d-%d", order.ID, time.Now().UnixNano()), Status: 0}
			model.DB.Create(&txn)
			r.Payments++
			r.Actions = append(r.Actions, fmt.Sprintf("订单 %d 完成", order.ID))
		}
	}
	qual := model.SupplierQualification{SupplierID: supplierID, QualificationType: "造价咨询甲级", CertificateNo: "ZZ-DEMO-" + fmt.Sprintf("%03d", rng.Intn(999)), Level: "甲级", Scope: "造价咨询", VerificationStatus: "pending"}
	model.DB.Create(&qual)
	r.Qualifiers++
	r.Actions = append(r.Actions, "服务方提交资质: 造价咨询甲级")
	if mode == "demo" {
		var firstOrder model.Order
		model.DB.First(&firstOrder)
		if firstOrder.ID > 0 {
			dispute := model.Dispute{OrderID: firstOrder.ID, InitiatorID: clientID, Reason: "交付物与合同约定不符", Claim: "要求重新核对", Status: "evidence"}
			model.DB.Create(&dispute)
			r.Disputes++
			r.Actions = append(r.Actions, fmt.Sprintf("订单 %d 发起争议", firstOrder.ID))
		}
	}
	if mode != "test" {
		var firstOrder model.Order
		model.DB.First(&firstOrder)
		if firstOrder.ID > 0 {
			att := model.AttendanceRecord{OrderID: firstOrder.ID, UserID: supplierID, CheckInAt: time.Now(), Longitude: 118.31, Latitude: 32.30}
			model.DB.Create(&att)
			r.Attendance++
			r.Actions = append(r.Actions, "服务方现场打卡")
		}
	}
	if mode == "training" {
		r.Actions = append(r.Actions, "培训模式：包含标准流程、部分完成、争议处理等教学案例")
	}
	WriteAudit(nil, "admin.demo.seed", "system", 0, gin.H{"mode": mode, "result": r})
	return r
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
	model.DB.Where("1 = 1").Delete(&model.User{})
	model.DB.Where("1 = 1").Delete(&model.AuditLog{})
}

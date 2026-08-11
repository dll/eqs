package handler

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupBizRouter 佣金 + 信用分测试路由
func setupBizRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	invalidatePublicCache()

	r := gin.New()
	api := r.Group("/api/v1")

	client := api.Group("")
	client.Use(AuthTestMiddleware(1, 1))
	{
		client.POST("/project/create", CreateProject)
		client.POST("/bid/:id/select", SelectBid)
		client.PUT("/order/:id/milestones", SetMilestones)
		client.POST("/contract/:id/sign", SignContract)
		client.POST("/milestone/:id/accept", ConfirmAcceptance)
		client.POST("/milestone/:id/settle", SettleMilestone)
		client.POST("/review/submit", SubmitReview)
	}

	supplier := api.Group("")
	supplier.Use(AuthTestMiddleware(2, 2))
	{
		supplier.POST("/bid/submit", SubmitBid)
	}

	contract := api.Group("")
	contract.Use(AuthTestMiddleware(1, 1))
	{
		contract.POST("/order/:id/contract", GenerateContract)
		contract.GET("/user/:id/reviews", GetUserReviews)
		contract.GET("/pay/balance", GetBalance)
	}

	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.GET("/admin/commission/list", AdminListCommissions)
		admin.POST("/admin/commission/:id/collect", AdminCollectCommission)
		admin.POST("/admin/config/upsert", AdminUpsertConfig)
	}

	// 公开配置
	api.GET("/config/public", PublicConfigs)
	return r
}

// buildSignedOrder 创建完整订单并签署，返回 orderID
func buildSignedOrder(t *testing.T, r *gin.Engine) uint {
	t.Helper()
	createTestUser(t, "13900000001", 1)
	createTestUser(t, "13900000002", 2)
	projectID := setupPublishedProject(t, r)

	w := doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 10000, "service_days": 15,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("报价失败: %d %s", w.Code, w.Body.String())
	}
	bidRaw := decodeBody(t, w)["bid"].(map[string]interface{})
	bidID := uint(bidRaw["id"].(float64))

	w = doJSONFull(t, r, "POST", "/api/v1/bid/"+strconv.Itoa(int(bidID))+"/select", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("中选失败: %d %s", w.Code, w.Body.String())
	}

	var order model.Order
	model.DB.Where("project_id = ?", projectID).First(&order)

	w = doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{
			{"name": "开工款", "ratio": 100},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("设置节点失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/contract", nil)
	contractRaw := decodeBody(t, w)["contract"].(map[string]interface{})
	contractID := uint(contractRaw["id"].(float64))

	w = doJSONFull(t, r, "POST", "/api/v1/contract/"+strconv.Itoa(int(contractID))+"/sign", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("签署失败: %d %s", w.Code, w.Body.String())
	}
	return order.ID
}

func TestCommission_GeneratedAfterSign(t *testing.T) {
	r := setupBizRouter()

	// 设置佣金比例 5%
	w := doJSONFull(t, r, "POST", "/api/v1/admin/config/upsert", map[string]interface{}{
		"config_key": "commission.rate", "config_value": "5", "value_type": "int", "is_public": true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("设置佣金比例失败: %d %s", w.Code, w.Body.String())
	}
	// 缓存失效后重新加载
	invalidatePublicCache()

	orderID := buildSignedOrder(t, r)

	// 验证佣金已生成
	var record model.CommissionRecord
	if err := model.DB.Where("order_id = ?", orderID).First(&record).Error; err != nil {
		t.Fatalf("佣金未生成: %v", err)
	}
	if record.Rate != 5 {
		t.Fatalf("佣金比例错误: %v", record.Rate)
	}
	// 10000 * 5% = 500
	if record.Commission != 500 {
		t.Fatalf("佣金金额错误: %v", record.Commission)
	}

	// 再次签署不应重复生成（幂等）
	calcAndCreateCommission(orderID)
	var count int64
	model.DB.Model(&model.CommissionRecord{}).Where("order_id = ?", orderID).Count(&count)
	if count != 1 {
		t.Fatalf("佣金重复生成: %d", count)
	}

	// 管理员查询列表
	w = doJSONFull(t, r, "GET", "/api/v1/admin/commission/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("佣金列表失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("期望1条佣金，得到 %v", out["count"])
	}

	// 标记收取
	idStr := strconv.FormatUint(uint64(record.ID), 10)
	w = doJSONFull(t, r, "POST", "/api/v1/admin/commission/"+idStr+"/collect", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("收取佣金失败: %d %s", w.Code, w.Body.String())
	}
	var after model.CommissionRecord
	model.DB.First(&after, record.ID)
	if after.Status != "collected" {
		t.Fatalf("佣金状态未更新: %s", after.Status)
	}
}

func TestCommission_ZeroRateNoRecord(t *testing.T) {
	r := setupBizRouter()
	// 默认佣金比例 0（初期免费），签署后不应生成佣金
	orderID := buildSignedOrder(t, r)

	var count int64
	model.DB.Model(&model.CommissionRecord{}).Where("order_id = ?", orderID).Count(&count)
	if count != 0 {
		t.Fatalf("佣金比例0不应生成佣金，得到 %d", count)
	}
}

func TestCreditScore_RecalcOnReview(t *testing.T) {
	model.InitTestDB()

	// 显式指定用户 ID，确保 reviewee_id=2 生效
	model.DB.Create(&model.User{ID: 2, Phone: "13900000002", UserType: 2, Status: 1, CreditScore: 100})

	// 创建一条评价
	model.DB.Create(&model.Review{OrderID: 1, ReviewerID: 1, RevieweeID: 2, Rating: 5})
	recalcUserCredit(2)

	var user model.User
	model.DB.First(&user, 2)
	// 满分评价：质量100*0.3 + 准时100*0.3 + 纠纷100*0.2 + 活跃20*0.1 + 履约100*0.1 = 92
	if user.CreditScore != 92 {
		t.Fatalf("满分评价后信用分应为92，得到 %v", user.CreditScore)
	}

	// 低分评价
	model.DB.Create(&model.Review{OrderID: 2, ReviewerID: 1, RevieweeID: 2, Rating: 1})
	recalcUserCredit(2)
	model.DB.First(&user, 2)
	// reviewScore = 100 - (5-3)*20 = 60（两条评价 avg=3）
	if user.CreditScore != 82 {
		t.Fatalf("低分评价后信用分应为82，得到 %v", user.CreditScore)
	}
}

func TestCreditScore_RecalcOnDisputeClose(t *testing.T) {
	model.InitTestDB()

	model.DB.Create(&model.User{Phone: "13900000001", UserType: 1, Status: 1, CreditScore: 100})
	model.DB.Create(&model.User{Phone: "13900000002", UserType: 2, Status: 1, CreditScore: 100})

	// 构造一单结案争议
	model.DB.Create(&model.Dispute{OrderID: 1, InitiatorID: 2, Reason: "交付不符", Status: "closed"})
	recalcUserCredit(2)

	var user model.User
	model.DB.First(&user, 2)
	// 五维：质量100*0.3 + 准时100*0.3 + 纠纷90*0.2 + 活跃20*0.1 + 履约100*0.1 = 90
	if user.CreditScore != 90 {
		t.Fatalf("结案争议后信用分应为90，得到 %v", user.CreditScore)
	}
}

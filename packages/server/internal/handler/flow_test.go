package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// AuthTestMiddleware 模拟登录中间件，注入 user_id 和 user_type
// 支持 X-Test-User-ID / X-Test-User-Type 头覆盖（便于同一 router 多角色测试）
func AuthTestMiddleware(userID, userType int) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := userID
		ut := userType
		if h := c.GetHeader("X-Test-User-ID"); h != "" {
			if v, err := strconv.Atoi(h); err == nil {
				uid = v
			}
		}
		if h := c.GetHeader("X-Test-User-Type"); h != "" {
			if v, err := strconv.Atoi(h); err == nil {
				ut = v
			}
		}
		c.Set("user_id", uint(uid))
		c.Set("user_type", ut)
		c.Next()
	}
}

func setupFlowRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")
	api.POST("/sms/send", SendSMS)
	api.POST("/auth/login", PhoneLogin)
	api.POST("/auth/wechat-login", WxLogin)

	client := api.Group("")
	client.Use(AuthTestMiddleware(1, 1))
	{
		client.POST("/project/create", CreateProject)
		client.GET("/project/list", ListProjects)
		client.GET("/project/:id", GetProject)
		client.GET("/project/:id/recommend", GetRecommendations)
		client.POST("/project/:id/invite", InviteSuppliers)
		client.POST("/bid/:id/select", SelectBid)
		client.GET("/order/list", ListMyOrders)
		client.GET("/order/:id", GetOrder)
		client.PUT("/order/:id/milestones", SetMilestones)
		client.POST("/milestone/:id/accept", ConfirmAcceptance)
		client.POST("/milestone/:id/settle", SettleMilestone)
		client.GET("/order/:id/disputes", ListDisputes)
		client.POST("/review/submit", SubmitReview)
	}

	supplier := api.Group("")
	supplier.Use(AuthTestMiddleware(2, 2))
	{
		supplier.POST("/bid/submit", SubmitBid)
		supplier.PUT("/bid/:id/withdraw", WithdrawBid)
		supplier.POST("/milestone/:id/deliver", UploadDeliverable)
	}

	shared := api.Group("")
	shared.Use(AuthTestMiddleware(1, 1))
	{
		shared.GET("/project/:id/bids", ListBids)
	}

	contract := api.Group("")
	contract.Use(AuthTestMiddleware(1, 1))
	{
		contract.GET("/contract/templates", ListContractTemplates)
		contract.POST("/order/:id/contract", GenerateContract)
		contract.POST("/contract/:id/sign", SignContract)
		contract.GET("/contract/:id/download", DownloadContract)
		contract.GET("/user/:id/reviews", GetUserReviews)
		contract.GET("/user/info", GetUserInfo)
		contract.PUT("/user/info", UpdateUserInfo)
		contract.POST("/pay/create", CreatePayment)
		contract.GET("/pay/transactions", ListPaymentTransactions)
		contract.GET("/pay/balance", GetBalance)
		contract.POST("/file/upload", UploadFile)
		contract.GET("/project/:id/files", ListFiles)
		contract.POST("/annotation/add", AddAnnotation)
		contract.GET("/annotation/list/:id", ListAnnotations)
		contract.PUT("/annotation/:id/resolve", ResolveAnnotation)
		contract.GET("/supplier/:id/qualifications", ListQualifications)
		contract.POST("/supplier/:id/qualifications", SubmitQualification)
		contract.POST("/qualification/:id/review", ReviewQualification)
		contract.POST("/attendance/checkin", CheckIn)
		contract.GET("/order/:id/attendance", ListAttendance)
		contract.POST("/dispute/:id/evidence", UploadDisputeEvidence)
		contract.GET("/dispute/:id", GetDispute)
		contract.POST("/dispute/:id/expert", AssignDisputeExpert)
		contract.POST("/dispute-expert/:id/opinion", SubmitExpertOpinion)
		contract.POST("/dispute/:id/close", CloseDispute)
		contract.POST("/dispute/create", CreateDispute)
	}
	return r
}

func doJSONFull(t *testing.T, r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// doJSONFullAuth 带指定用户认证的请求
func doJSONFullAuth(t *testing.T, r *gin.Engine, method, path string, body interface{}, userID, userType int) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", strconv.Itoa(userID))
	req.Header.Set("X-Test-User-Type", strconv.Itoa(userType))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}


func decodeBody(t *testing.T, w *httptest.ResponseRecorder) gin.H {
	t.Helper()
	var out gin.H
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("解析响应失败(%d): %v -> %s", w.Code, err, w.Body.String())
	}
	return out
}

// setupPublishedProject 创建甲方+服务方+已发布项目，返回项目ID
func setupPublishedProject(t *testing.T, r *gin.Engine) (projectID uint) {
	t.Helper()

	// 创建甲方项目（通过中间件 user_id=1 甲方）
	w := doJSONFull(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"project_type": "cost", "service_type": "cost",
		"title": "招标控制价编制", "budget_min": 10000, "budget_max": 50000,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("创建项目失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	proj := out["project"].(map[string]interface{})
	id := uint(proj["id"].(float64))

	// 发布项目（状态 0->1）
	model.DB.Model(&model.Project{}).Where("id = ?", id).Update("status", 1)
	return id
}

func TestFullLifecycle_ProjectToSettle(t *testing.T) {
	r := setupFlowRouter()

	// 准备甲方与服务方账号（中间件固定 1=甲方, 2=服务方）
	createTestUser(t, "13800000001", 1)
	createTestUser(t, "13800000002", 2)

	projectID := setupPublishedProject(t, r)

// 服务方报价
	w := doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 20000, "service_days": 30,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("报价失败: %d %s", w.Code, w.Body.String())
	}
	bidOut := decodeBody(t, w)
	bidRaw := bidOut["bid"].(map[string]interface{})
	bidID := uint(bidRaw["id"].(float64))

	// 甲方中选
	w = doJSONFull(t, r, "POST", "/api/v1/bid/"+strconv.Itoa(int(bidID))+"/select", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("中选失败: %d %s", w.Code, w.Body.String())
	}

	// 查询订单（含节点与资金状态）
	var order model.Order
	if err := model.DB.Where("project_id = ?", projectID).First(&order).Error; err != nil {
		t.Fatalf("订单未生成: %v", err)
	}

	// 设置付款节点
	w = doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{
			{"name": "开工款", "ratio": 30},
			{"name": "验收款", "ratio": 70},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("设置节点失败: %d %s", w.Code, w.Body.String())
	}

	// 生成合同并签署
	w = doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/contract", nil)
	out := decodeBody(t, w)
	contractRaw, ok := out["contract"].(map[string]interface{})
	if !ok {
		t.Fatalf("合同未生成: %s", w.Body.String())
	}
	contractID := uint(contractRaw["id"].(float64))

	w = doJSONFull(t, r, "POST", "/api/v1/contract/"+strconv.Itoa(int(contractID))+"/sign", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("签署失败: %d %s", w.Code, w.Body.String())
	}

	// 节点交付+验收流程
	var ms1, ms2 model.PaymentMilestone
	model.DB.Where("order_id = ? AND sequence = 1", order.ID).First(&ms1)
	model.DB.Where("order_id = ? AND sequence = 2", order.ID).First(&ms2)

	// 服务方上传交付物（中间件 user_id=1 走 client 组；交付应由服务方完成，测试用甲方模拟亦可走通）
	w = doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms1.ID))+"/deliver", map[string]interface{}{
		"file_name": "控制价清单.xlsx", "file_url": "https://cos.xxx/budget.xlsx",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("交付失败: %d %s", w.Code, w.Body.String())
	}

	// 甲方验收
	w = doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms1.ID))+"/accept", map[string]interface{}{
		"accept": true, "comment": "合格",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("验收失败: %d %s", w.Code, w.Body.String())
	}

	// 验收后结算
	w = doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms1.ID))+"/settle", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("结算失败: %d %s", w.Code, w.Body.String())
	}

	// 订单详情应含已结算节点
	w = doJSONFull(t, r, "GET", "/api/v1/order/"+strconv.Itoa(int(order.ID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("订单详情失败: %d %s", w.Code, w.Body.String())
	}
	out = decodeBody(t, w)
	msOut := out["milestones"].([]interface{})
	if len(msOut) != 2 {
		t.Fatalf("期望2个节点，得到 %d", len(msOut))
	}
}

func TestOrderMilestoneRatioMustBe100(t *testing.T) {
	r := setupFlowRouter()

	// 空 DB 快速构造：直接建甲方，项目，订单
	createTestUser(t, "13800000010", 1)
	createTestUser(t, "13800000011", 2)
	projectID := setupPublishedProject(t, r)

	// 直接构造订单（待签约）
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 10000, Status: 0}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	w := doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{
			{"name": "首期", "ratio": 40},
			{"name": "尾款", "ratio": 40},
		},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("节点比例合计非100%%应拒绝，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestBidDuplicateRejected(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13800000020", 2)
	projectID := setupPublishedProject(t, r)

	w := doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 1000, "service_days": 5,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("首次报价失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 800, "service_days": 5,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("重复报价应被拒绝，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestBidsMaskedForSupplier(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13800000090", 1)
	createTestUser(t, "13800000091", 2)
	projectID := setupPublishedProject(t, r)

	// 服务方（user_type=2）报价后查看排名应脱敏
	w := doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 8888, "service_days": 10,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("报价失败: %d %s", w.Code, w.Body.String())
	}

	// 单独以服务方视角路由验证脱敏（user_type=2），复用同一测试库
	gin.SetMode(gin.TestMode)
	r2 := gin.New()
	api2 := r2.Group("/api/v1")
	supplierView := api2.Group("")
	supplierView.Use(AuthTestMiddleware(2, 2))
	supplierView.GET("/project/:id/bids", ListBids)

	w = doJSONFull(t, r2, "GET", "/api/v1/project/"+strconv.Itoa(int(projectID))+"/bids", nil)
	out := decodeBody(t, w)
	bids := out["bids"].([]interface{})
	if len(bids) != 1 {
		t.Fatalf("期望1条脱敏报价，得到 %d", len(bids))
	}
	firstBid := bids[0].(map[string]interface{})
	if _, hasAmount := firstBid["amount"]; hasAmount {
		t.Fatalf("服务方视角不应暴露金额: %v", firstBid)
	}
}

func TestSetMilestones_NotClientForbidden(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13800000030", 1)
	createTestUser(t, "13800000031", 2)
	projectID := setupPublishedProject(t, r)

	// 服务方发起的项目（让服务方建项目：中间件固定甲方在中选，直接改项目归属）
	model.DB.Model(&model.Project{}).Where("id = ?", projectID).Update("user_id", 2)

	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 0}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatal(err)
	}
	w := doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{{"name": "全款", "ratio": 100}},
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("非甲方设置节点应被拒绝(403)，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestGetRecommendations(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13800000041", 1)
	createTestUser(t, "13800000042", 2)
	createTestUser(t, "13800000043", 1)
	projectID := setupPublishedProject(t, r)

	// 服务方压低信用分以测试筛选
	model.DB.Model(&model.User{}).Where("id = ?", 2).Update("credit_score", 60)

	w := doJSONFull(t, r, "GET", "/api/v1/project/"+strconv.Itoa(int(projectID))+"/recommend", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("推荐接口失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	suppliers := out["suppliers"].([]interface{})
	for _, s := range suppliers {
		sup := s.(map[string]interface{})
		if sup["credit_score"].(float64) < 70 {
			t.Fatalf("推荐结果不应含低信用服务方: %v", sup)
		}
	}
}

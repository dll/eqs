package handler

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// TestPhoneLogin_InvalidCode 无效验证码登录被拒（走 verifySMS false 分支）
func TestPhoneLogin_InvalidCode2(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13600000000", "code": "999999", "user_type": 1,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("错误验证码应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestPhoneLogin_SuccessNew 新用户登录走 findOrCreateUser 创建分支，再登录走查询分支
func TestPhoneLogin_TwoUsers(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13600000001", "code": "123456", "user_type": 2,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("登录失败: %d %s", w.Code, w.Body.String())
	}
	w2 := doJSONFull(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13600000002", "code": "123456", "user_type": 1,
	})
	if w2.Code != http.StatusOK {
		t.Fatalf("第二位用户登录失败: %d %s", w2.Code, w2.Body.String())
	}
}

// TestPhoneLogin_MissingParams 缺参 400
func TestPhoneLogin_MissingParams(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺参应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSignContract_SignAgain 幂等签署
func TestSignContract_SignAgain(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13610000001", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 1000, Status: 1}
	model.DB.Create(&order)

	w := doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/contract", nil)
	out := decodeBody(t, w)
	contractID := uint(out["contract"].(map[string]interface{})["id"].(float64))

	w = doJSONFull(t, r, "POST", "/api/v1/contract/"+strconv.Itoa(int(contractID))+"/sign", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("首次签署失败: %d %s", w.Code, w.Body.String())
	}
	w = doJSONFull(t, r, "POST", "/api/v1/contract/"+strconv.Itoa(int(contractID))+"/sign", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("重复签署失败: %d %s", w.Code, w.Body.String())
	}
}

// TestGenerateContract_Reuse 已存在合同草稿复用
func TestGenerateContract_Reuse(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13610000011", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 1000, Status: 1}
	model.DB.Create(&order)

	w := doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/contract", nil)
	out := decodeBody(t, w)
	contract1ID := uint(out["contract"].(map[string]interface{})["id"].(float64))

	w = doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/contract", nil)
	out = decodeBody(t, w)
	contract2ID := uint(out["contract"].(map[string]interface{})["id"].(float64))
	if contract1ID != contract2ID {
		t.Fatalf("重复生成应复用合同草稿: %d vs %d", contract1ID, contract2ID)
	}
}

// TestCreatePayment_WechatRealChannel 真实通道未签约时使用wechat被拒
func TestCreatePayment_WechatNoSign(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13610000003", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 100, Status: 1}
	model.DB.Create(&order)

	// 模拟真实通道配置（环境变量）
	t.Setenv("PAYMENT_PROVIDER", "wechat")
	defer t.Setenv("PAYMENT_PROVIDER", "mock")

	w := doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": order.ID, "amount": 100, "channel": "wechat",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("未就绪通道应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestPaymentNotify_AmountMismatch 回调金额不符不更新状态
func TestPaymentNotify_AmountMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	txn := model.PaymentTransaction{
		UserID: 1, OrderID: 1, Amount: 1000, Type: "payment", Channel: "mock",
		ExternalTransactionID: "PAY-1", Status: 0,
	}
	model.DB.Create(&txn)

	notify := gin.New()
	notify.POST("/pay/notify/:channel", PaymentNotify)
	w := doJSONFull(t, notify, "POST", "/pay/notify/mock", map[string]interface{}{
		"external_transaction_id": "PAY-1", "order_id": 1, "amount": 999, "result": "success",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("回调处理失败: %d %s", w.Code, w.Body.String())
	}

	var fetched model.PaymentTransaction
	model.DB.First(&fetched, txn.ID)
	if fetched.Status != 0 {
		t.Fatalf("金额不匹配不应更新状态，得到 %d", fetched.Status)
	}

	// 成功回调更新状态
	notify2 := gin.New()
	notify2.POST("/pay/notify/:channel", PaymentNotify)
	w = doJSONFull(t, notify2, "POST", "/pay/notify/mock", map[string]interface{}{
		"external_transaction_id": "PAY-1", "order_id": 1, "amount": 1000, "result": "success",
	})
	model.DB.First(&fetched, txn.ID)
	if fetched.Status != 1 {
		t.Fatalf("成功回调应置为1，得到 %d", fetched.Status)
	}

	// 幂等
	w = doJSONFull(t, notify2, "POST", "/pay/notify/mock", map[string]interface{}{
		"external_transaction_id": "PAY-1", "order_id": 1, "amount": 1000, "result": "success",
	})
	out := decodeBody(t, w)
	if out["idempotent"] != true {
		t.Fatalf("重复回调应幂等: %v", out)
	}
}

// TestCreatePayment_InvalidBody 缺参 400
func TestCreatePayment_InvalidBody(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺参应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestInviteSuppliers_StartedProject 项目已开始不能邀请
func TestInviteSuppliers_StartedProject(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13610000010", 1)
	projectID := setupPublishedProject(t, r)
	model.DB.Model(&model.Project{}).Where("id = ?", projectID).Update("status", 3)

	w := doJSONFull(t, r, "POST", "/api/v1/project/"+strconv.Itoa(int(projectID))+"/invite", map[string]interface{}{
		"supplier_ids": []uint{2},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("已开始项目不可邀请，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSubmitBid_NotBidding 项目不在报价期（状态非1）
func TestSubmitBid_NotBidding(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13610000020", 2)
	projectID := setupPublishedProject(t, r)
	model.DB.Model(&model.Project{}).Where("id = ?", projectID).Update("status", 3)

	supplierRouter := gin.New()
	sr := supplierRouter.Group("")
	sr.Use(AuthTestMiddleware(2, 2))
	sr.POST("/bid/submit", SubmitBid)
	w := doJSONFull(t, supplierRouter, "POST", "/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 100, "service_days": 1,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非报价期应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSubmitBid_InvalidBody 缺参 400
func TestSubmitBid_InvalidBody(t *testing.T) {
	model.InitTestDB()
	supplierRouter := gin.New()
	s := supplierRouter.Group("")
	s.Use(AuthTestMiddleware(2, 2))
	s.POST("/api/v1/bid/submit", SubmitBid)
	w := doJSONFull(t, supplierRouter, "POST", "/api/v1/bid/submit", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺参应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSkill 覆盖 ListContractTemplates 缺少 service_type
func TestListContractTemplates_All(t *testing.T) {
	model.InitTestDB()
	model.DB.Create(&model.ContractTemplate{
		Name: "默认", Version: "1.0", ServiceType: "geotech", Status: "active",
	})
	gin.SetMode(gin.TestMode)
	r := gin.New()
	api := r.Group("/api/v1")
	api.GET("/contract/templates", ListContractTemplates)
	w := doJSONFull(t, r, "GET", "/api/v1/contract/templates", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查询模板失败: %d %s", w.Code, w.Body.String())
	}
}

func TestListBids_NoProjects(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/project/99999/bids", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestResolveAnnotation_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "PUT", "/api/v1/annotation/77777/resolve", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的批注应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestAddAnnotation_FileNotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/annotation/add", map[string]interface{}{
		"file_id": 99999, "content": "x",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的文件应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestUploadFile_MissingField(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/file/upload", map[string]interface{}{
		"project_id": 1,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺字段应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestGetDispute_NotFound 不存在的争议
func TestGetDispute_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/dispute/7777", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的争议应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestGenerateContract_OrderNotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/order/8888/contract", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的订单应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSignContract_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/contract/8888/sign", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的合同应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSetMilestones_OrderNotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "PUT", "/api/v1/order/8888/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{{"name": "a", "ratio": 100}},
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的订单应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestUploadDeliverable_NodeNotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/milestone/8888/deliver", map[string]interface{}{
		"file_name": "x", "file_url": "y",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的节点应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestConfirmAcceptance_NodeNotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/milestone/8888/accept", map[string]interface{}{
		"accept": true,
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的节点应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSettleMilestone_NodeNotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/milestone/8888/settle", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的节点应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestCheckIn_NotSupplier 非服务方打卡被拒
func TestCheckIn_NotSupplier(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13610000060", 1)
	createTestUser(t, "13610000061", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 100, Status: 1}
	model.DB.Create(&order)

	// 甲方(1)身份打卡被拒
	w := doJSONFull(t, r, "POST", "/api/v1/attendance/checkin", map[string]interface{}{
		"order_id": order.ID,
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("非服务方打卡应403，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestListAttendance_InvalidOrder(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/order/abc/attendance", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法订单ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestAssignDisputeExpert_NotExists 指派不存在的专家 ID
func TestAssignDisputeExpert_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/dispute/1/expert", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺参数应400，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSubmitExpertOpinion_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/dispute-expert/8888/opinion", map[string]interface{}{
		"opinion": "x", "vote": "partial",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的指派应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSubmitExpertOpinion_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/dispute-expert/1/opinion", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺参数应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestCloseDispute_Idempotent 已结案重复结案幂等
func TestCloseDispute_Idempotent(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13610000001", 1)
	createTestUser(t, "13610000021", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 100, Status: 1}
	model.DB.Create(&order)
	dispute := model.Dispute{OrderID: order.ID, InitiatorID: 1, Status: "closed", Reason: "done"}
	model.DB.Create(&dispute)

	w := doJSONFull(t, r, "POST", "/api/v1/dispute/"+strconv.Itoa(int(dispute.ID))+"/close", map[string]interface{}{
		"resolution_type": "settlement",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("幂等结案失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["idempotent"] != true {
		t.Fatalf("已结案应幂等: %v", out)
	}

	// 不存在的争议
	w = doJSONFull(t, r, "POST", "/api/v1/dispute/9999/close", map[string]interface{}{
		"resolution_type": "settlement",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在争议应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestGetUserReviews_Empty 无评价返回0条
func TestGetUserReviews_Empty(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/user/55555/reviews", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查询评价失败: %d %s", w.Code, w.Body.String())
	}
}

func TestGetUserReviews_InvalidID(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/user/abc/reviews", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}
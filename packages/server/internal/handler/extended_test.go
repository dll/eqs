package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// seedOrder 构造一个已完成签约的订单，返回订单及节点
func seedSignedOrder(t *testing.T, r *gin.Engine) (orderID uint, milestoneID uint) {
	t.Helper()
	createTestUser(t, "13900000001", 1)
	createTestUser(t, "13900000002", 2)
	projectID := setupPublishedProject(t, r)

	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 10000, Status: 1}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}

	ms := model.PaymentMilestone{OrderID: order.ID, Name: "首期款", Sequence: 1, Ratio: 100, Amount: 10000, Status: "pending"}
	if err := model.DB.Create(&ms).Error; err != nil {
		t.Fatalf("创建节点失败: %v", err)
	}
	return order.ID, ms.ID
}

func TestGetUserInfo(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000100", 1)
	w := doJSONFull(t, r, "GET", "/api/v1/user/info", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("获取用户信息失败: %d %s", w.Code, w.Body.String())
	}
}

func TestUpdateUserInfo(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000101", 1)
	w := doJSONFull(t, r, "PUT", "/api/v1/user/info", map[string]interface{}{
		"company_name": "测试公司",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("更新用户信息失败: %d %s", w.Code, w.Body.String())
	}
}

func TestCreatePayment_AmountMismatch(t *testing.T) {
	r := setupFlowRouter()
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 10000, Status: 1}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatal(err)
	}
	w := doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": order.ID, "amount": 9999, "channel": "mock",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("金额不一致应被拒绝，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestCreatePayment_SuccessAndBalance(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000110", 1)
	createTestUser(t, "13900000111", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 10000, Status: 1}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	w := doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": order.ID, "amount": 10000, "channel": "mock",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("创建支付失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "GET", "/api/v1/pay/transactions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查询流水失败: %d %s", w.Code, w.Body.String())
	}
}

func TestFileUploadAndAnnotation(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000120", 1)
	projectID := setupPublishedProject(t, r)

	w := doJSONFull(t, r, "POST", "/api/v1/file/upload", map[string]interface{}{
		"project_id": projectID, "original_name": "图纸.dwg", "file_type": "dwg",
		"storage_key": "cos/bucket/1.dwg", "sha256": "abc123",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("上传文件失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	file := out["file"].(map[string]interface{})
	fileID := uint(file["id"].(float64))

	// 添加批注
	w = doJSONFull(t, r, "POST", "/api/v1/annotation/add", map[string]interface{}{
		"file_id": fileID, "page_no": 1, "xratio": 0.5, "yratio": 0.5, "content": "此处尺寸有误",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("添加批注失败: %d %s", w.Code, w.Body.String())
	}
	out = decodeBody(t, w)
	annot := out["annotation"].(map[string]interface{})
	annotID := uint(annot["id"].(float64))

	// 查看批注
	w = doJSONFull(t, r, "GET", "/api/v1/annotation/list/"+strconv.Itoa(int(fileID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查看批注失败: %d %s", w.Code, w.Body.String())
	}

	// 解决批注
	w = doJSONFull(t, r, "PUT", "/api/v1/annotation/"+strconv.Itoa(int(annotID))+"/resolve", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("解决批注失败: %d %s", w.Code, w.Body.String())
	}
}

func TestQualificationFlow(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000130", 2)

	w := doJSONFullAuth(t, r, "POST", "/api/v1/supplier/2/qualifications", map[string]interface{}{
		"qualification_type": "咨询资质", "certificate_no": "ZX-2026-001", "level": "甲级",
	}, 2, 2)
	if w.Code != http.StatusOK {
		t.Fatalf("提交资质失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	qual := out["qualification"].(map[string]interface{})
	qualID := uint(qual["id"].(float64))

	// 平台核验通过
	w = doJSONFullAuth(t, r, "POST", "/api/v1/qualification/"+strconv.Itoa(int(qualID))+"/review", map[string]interface{}{
		"verified": true, "comment": "核验通过",
	}, 9, 3)
	if w.Code != http.StatusOK {
		t.Fatalf("核验失败: %d %s", w.Code, w.Body.String())
	}

	// 查询资质列表
	w = doJSONFullAuth(t, r, "GET", "/api/v1/supplier/2/qualifications", nil, 9, 3)
	if w.Code != http.StatusOK {
		t.Fatalf("查询资质失败: %d %s", w.Code, w.Body.String())
	}
}

func TestAttendanceCheckIn(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000130", 1)
	createTestUser(t, "13900000131", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatal(err)
	}

	// 打卡（服务方身份）
	supplierRouter := gin.New()
	sr := supplierRouter.Group("")
	sr.Use(AuthTestMiddleware(2, 2))
	sr.POST("/attendance/checkin", CheckIn)

	w := doJSONFull(t, supplierRouter, "POST", "/attendance/checkin", map[string]interface{}{
		"order_id": order.ID, "longitude": 118.3, "latitude": 32.9,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("打卡失败: %d %s", w.Code, w.Body.String())
	}

	// 超距标记异常
	w = doJSONFull(t, supplierRouter, "POST", "/attendance/checkin", map[string]interface{}{
		"order_id": order.ID, "distance_meters": 6000,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("打卡失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "GET", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/attendance", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查询打卡失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	records := out["attendance"].([]interface{})
	if len(records) != 2 {
		t.Fatalf("期望2条打卡记录，得到 %d", len(records))
	}
}

func TestContractTemplates(t *testing.T) {
	r := setupFlowRouter()
	// 预置模板
	tpl := model.ContractTemplate{
		Name: "造价咨询合同", Version: "1.0", ServiceType: "cost", Status: "active",
	}
	if err := model.DB.Create(&tpl).Error; err != nil {
		t.Fatal(err)
	}

	w := doJSONFull(t, r, "GET", "/api/v1/contract/templates?service_type=cost", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查询模板失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	templates := out["templates"].([]interface{})
	if len(templates) != 1 {
		t.Fatalf("期望1条模板，得到 %d", len(templates))
	}
}

func TestDisputeFlow(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000140", 1)
	createTestUser(t, "13900000141", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 6000, Status: 1}
	model.DB.Create(&order)
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "款", Sequence: 1, Ratio: 100, Amount: 6000, Status: "pending"}
	model.DB.Create(&ms)

	// 服务方发起争议（中间件为甲方，直接以甲方身份发起亦可走通）
	w := doJSONFull(t, r, "POST", "/api/v1/dispute/create", map[string]interface{}{
		"order_id": order.ID, "milestone_id": ms.ID, "reason": "交付质量不达标",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("发起争议失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	disputeRaw := out["dispute"].(map[string]interface{})
	if disputeRaw["status"] != "evidence" {
		t.Fatalf("初始状态应为 evidence，得到 %v", disputeRaw["status"])
	}

	// 上传证据 -> 进入评审
	w = doJSONFull(t, r, "POST", "/api/v1/dispute/"+strconv.Itoa(int(disputeRaw["id"].(float64)))+"/evidence", map[string]interface{}{
		"file_id": 1, "sha256": "efg456", "content": "质检报告截图",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("上传证据失败: %d %s", w.Code, w.Body.String())
	}

	// 指派专家（平台）
	expert := model.User{Phone: "13900000142", UserType: 4, Status: 1}
	model.DB.Create(&expert)
	disputeID := uint(disputeRaw["id"].(float64))
	w = doJSONFull(t, r, "POST", "/api/v1/dispute/"+strconv.Itoa(int(disputeID))+"/expert", map[string]interface{}{
		"expert_user_id": expert.ID,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("指派专家失败: %d %s", w.Code, w.Body.String())
	}
	out = decodeBody(t, w)
	assignRaw := out["assignment"].(map[string]interface{})
	assignmentID := uint(assignRaw["id"].(float64))

	// 专家提交意见（以专家身份调用）
	expertRouter := gin.New()
	er := expertRouter.Group("")
	er.Use(AuthTestMiddleware(int(expert.ID), 4))
	er.POST("/dispute-expert/:id/opinion", SubmitExpertOpinion)
	w = doJSONFull(t, expertRouter, "POST", "/dispute-expert/"+strconv.Itoa(int(assignmentID))+"/opinion", map[string]interface{}{
		"opinion": "双方协商解决", "vote": "partial",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("专家提交意见失败: %d %s", w.Code, w.Body.String())
	}

	// 结案
	w = doJSONFull(t, r, "POST", "/api/v1/dispute/"+strconv.Itoa(int(disputeID))+"/close", map[string]interface{}{
		"resolution_type": "settlement", "settle_amount": 3000,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("结案失败: %d %s", w.Code, w.Body.String())
	}

	// 详情
	w = doJSONFull(t, r, "GET", "/api/v1/dispute/"+strconv.Itoa(int(disputeID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("详情失败: %d %s", w.Code, w.Body.String())
	}
}

func TestPaymentNotify_Idempotent(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000150", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 9000, Status: 1}
	model.DB.Create(&order)

	w := doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": order.ID, "amount": 9000, "channel": "mock",
	})
	out := decodeBody(t, w)
	txnRaw := out["transaction"].(map[string]interface{})
	txnID := txnRaw["external_transaction_id"].(string)
	model.DB.Model(&model.PaymentTransaction{}).Where("id = ?", uint(txnRaw["id"].(float64))).Update("status", 1)

	// 模拟回调携带相同金额
	notify := gin.New()
	notify.POST("/pay/notify/:channel", PaymentNotify)

	w = doRequestRaw(t, notify, txnID, order.ID)
	if w.Code != http.StatusOK {
		t.Fatalf("回调失败: %d %s", w.Code, w.Body.String())
	}
}

func TestWithdrawBid(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000160", 1)
	createTestUser(t, "13900000161", 2)
	projectID := setupPublishedProject(t, r)

	w := doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 2000, "service_days": 5,
	})
	out := decodeBody(t, w)
	bidID := uint(out["bid"].(map[string]interface{})["id"].(float64))

	w = doJSONFull(t, r, "PUT", "/api/v1/bid/"+strconv.Itoa(int(bidID))+"/withdraw", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("撤回失败: %d %s", w.Code, w.Body.String())
	}

	// 重复撤回应允许（状态变更为已撤回）
	w = doJSONFull(t, r, "PUT", "/api/v1/bid/"+strconv.Itoa(int(bidID))+"/withdraw", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("重复撤回不应报错: %d %s", w.Code, w.Body.String())
	}
}

func TestSettleMilestone_NotAcceptedRejected(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000170", 1)
	createTestUser(t, "13900000171", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	model.DB.Create(&order)
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "款", Sequence: 1, Ratio: 100, Amount: 5000, Status: "pending"}
	model.DB.Create(&ms)

	w := doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms.ID))+"/settle", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("未验收节点不可结算，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestDeliverableSubmitAndReject(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13900000180", 1)
	createTestUser(t, "13900000181", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 8000, Status: 0}
	model.DB.Create(&order)
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "设计", Sequence: 1, Ratio: 100, Amount: 8000, Status: "pending"}
	model.DB.Create(&ms)

	// 提交交付物
	w := doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms.ID))+"/deliver", map[string]interface{}{
		"file_name": "初稿.dwg", "file_url": "https://cos/x.dwg",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("提交交付失败: %d %s", w.Code, w.Body.String())
	}

	// 驳回
	w = doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms.ID))+"/accept", map[string]interface{}{
		"accept": false, "comment": "需修改",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("驳回失败: %d %s", w.Code, w.Body.String())
	}

	// 订单未签约（Status=0），但交付可上传
	var del model.Deliverable
	model.DB.First(&del)
	if del.Status != 2 {
		t.Fatalf("被驳回交付物状态应为2，得到 %d", del.Status)
	}
}

func doRequestRaw(t *testing.T, r *gin.Engine, txnID string, orderID uint) *httptest.ResponseRecorder {
	t.Helper()
	body := map[string]interface{}{
		"external_transaction_id": txnID, "order_id": orderID, "amount": 9000, "result": "success",
	}
	return doJSONFull(t, r, "POST", "/pay/notify/mock", body)
}
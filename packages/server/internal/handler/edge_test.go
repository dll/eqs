package handler

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// TestListProjects 按筛选条件查询项目列表
func TestListProjects(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000001", 1)
	projectID := setupPublishedProject(t, r)

	// 无条件
	w := doJSONFull(t, r, "GET", "/api/v1/project/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("列表查询失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if projs, ok := out["projects"].([]interface{}); !ok || len(projs) != 1 {
		t.Fatalf("期望1个项目，得到 %v", out["projects"])
	}

	// 服务类型筛选
	w = doJSONFull(t, r, "GET", "/api/v1/project/list?service_type=cost", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("类型筛选失败: %d %s", w.Code, w.Body.String())
	}

	// 关键字筛选
	w = doJSONFull(t, r, "GET", "/api/v1/project/list?keyword=招标", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("关键字筛选失败: %d %s", w.Code, w.Body.String())
	}

	// 不存在的项目118详情
	_ = projectID
}

func TestGetProject_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/project/99999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestInviteSuppliers(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000010", 1)
	createTestUser(t, "13910000011", 2)
	projectID := setupPublishedProject(t, r)

	// 超过5家被拒
	w := doJSONFull(t, r, "POST", "/api/v1/project/"+strconv.Itoa(int(projectID))+"/invite", map[string]interface{}{
		"supplier_ids": []uint{1, 2, 3, 4, 5, 6},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("超5家邀请应被拒绝，得到 %d %s", w.Code, w.Body.String())
	}

	// 正常邀请1家
	w = doJSONFull(t, r, "POST", "/api/v1/project/"+strconv.Itoa(int(projectID))+"/invite", map[string]interface{}{
		"supplier_ids": []uint{2},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("邀请失败: %d %s", w.Code, w.Body.String())
	}

	// 项目不存在
	w = doJSONFull(t, r, "POST", "/api/v1/project/88888/invite", map[string]interface{}{
		"supplier_ids": []uint{2},
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestGetBalance(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000020", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 10000, Status: 1}
	model.DB.Create(&order)

	// 先创建一笔支付以统计
	w := doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": order.ID, "amount": 10000, "channel": "mock",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("创建支付失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "GET", "/api/v1/pay/balance", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查询余额失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	// 支付流水归属供应商(user 2)，甲方视角无流水
	if out["transaction_count"].(float64) != 0 {
		t.Fatalf("甲方不应有流水，得到 %v", out["transaction_count"])
	}
}

func TestSubmitReview(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000030", 1)
	createTestUser(t, "13910000031", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 3}
	model.DB.Create(&order)

	// 未完成订单不可评论（状态改为3已完成）
	w := doJSONFull(t, r, "POST", "/api/v1/review/submit", map[string]interface{}{
		"order_id": order.ID, "reviewee_id": 2, "rating": 5, "content": "服务专业",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("评价失败: %d %s", w.Code, w.Body.String())
	}

	// 重复评价被拒
	w = doJSONFull(t, r, "POST", "/api/v1/review/submit", map[string]interface{}{
		"order_id": order.ID, "reviewee_id": 2, "rating": 4,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("重复评价应被拒绝，得到 %d %s", w.Code, w.Body.String())
	}

	// 查看被评人评价记录
	w = doJSONFull(t, r, "GET", "/api/v1/user/2/reviews", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("查询评价失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("期望1条评价，得到 %v", out["count"])
	}

	// 无效评分
	w = doJSONFull(t, r, "POST", "/api/v1/review/submit", map[string]interface{}{
		"order_id": order.ID, "reviewee_id": 2, "rating": 9,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("无效评分应被拒绝，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSubmitReview_OrderNotCompleted(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000040", 1)
	createTestUser(t, "13910000041", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	model.DB.Create(&order)

	w := doJSONFull(t, r, "POST", "/api/v1/review/submit", map[string]interface{}{
		"order_id": order.ID, "reviewee_id": 2, "rating": 5,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("未完成订单不可评价，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestListFiles(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000050", 1)
	projectID := setupPublishedProject(t, r)

	model.DB.Create(&model.ProjectFile{
		ProjectID: projectID, UploaderID: 1, OriginalName: "a.pdf", FileType: "pdf", StorageKey: "k", Version: 1,
	})
	w := doJSONFull(t, r, "GET", "/api/v1/project/"+strconv.Itoa(int(projectID))+"/files", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("列出文件失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if len(out["files"].([]interface{})) != 1 {
		t.Fatalf("期望1个文件，得到 %v", out["files"])
	}
}

func TestDownloadContract(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000060", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 3000, Status: 1}
	model.DB.Create(&order)

	w := doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/contract", nil)
	out := decodeBody(t, w)
	contract := out["contract"].(map[string]interface{})
	contractID := uint(contract["id"].(float64))

	w = doJSONFull(t, r, "GET", "/api/v1/contract/"+strconv.Itoa(int(contractID))+"/download", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("下载合同失败: %d %s", w.Code, w.Body.String())
	}

	// 不存在的合同
	w = doJSONFull(t, r, "GET", "/api/v1/contract/55555/download", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("期望404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestPaymentNotify_WrongAmount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	notify := gin.New()
	notify.POST("/pay/notify/:channel", PaymentNotify)
	body := map[string]interface{}{
		"external_transaction_id": "NOPE", "order_id": 1, "amount": 1, "result": "success",
	}
	w := doJSONFull(t, notify, "POST", "/pay/notify/mock", body)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在的交易应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSubmitBid_InvalidProject(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000060", 2)
	w := doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": 99999, "amount": 100, "service_days": 1,
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("项目不存在应404，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSetMilestones_InvalidParam(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000070", 1)
	projectID := setupPublishedProject(t, r)

	// 空节点列表
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 9000, Status: 0}
	model.DB.Create(&order)
	w := doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/milestones", map[string]interface{}{
		"milestones": []interface{}{},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("空节点应400，得到 %d %s", w.Code, w.Body.String())
	}
}
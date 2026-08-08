package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func TestSelectBid_NotOwnerForbidden(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13930000001", 1)
	createTestUser(t, "13930000002", 2)
	projectID := setupPublishedProject(t, r)

	w := doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 200, "service_days": 2,
	})
	out := decodeBody(t, w)
	bidID := uint(out["bid"].(map[string]interface{})["id"].(float64))

	// 项目归属改为其他人，甲方不再有权限
	model.DB.Model(&model.Project{}).Where("id = ?", projectID).Update("user_id", 99)

	clientRouter := gin.New()
	cr := clientRouter.Group("")
	cr.Use(AuthTestMiddleware(1, 1))
	cr.POST("/bid/:id/select", SelectBid)
	w = doJSONFull(t, clientRouter, "POST", "/bid/"+strconv.Itoa(int(bidID))+"/select", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非项目所有选中应403，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSelectBid_NotSubmittedRejected(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000010", 1)
	projectID := setupPublishedProject(t, r)

	// 直接构造 withdrawn 状态报价
	bid := model.Bid{ProjectID: projectID, SupplierID: 50, Amount: 1, Status: "withdrawn"}
	model.DB.Create(&bid)

	clientRouter := gin.New()
	cr := clientRouter.Group("")
	cr.Use(AuthTestMiddleware(1, 1))
	cr.POST("/api/v1/bid/:id/select", SelectBid)
	w := doJSONFull(t, clientRouter, "POST", "/api/v1/bid/"+strconv.Itoa(int(bid.ID))+"/select", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非submitted报价不可中选，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestCreateDispute_Forbidden(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000020", 1)
	createTestUser(t, "13910000021", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	model.DB.Create(&order)

	// 无关第三方（user 99）发起争议被拒
	outsider := gin.New()
	og := outsider.Group("")
	og.Use(AuthTestMiddleware(99, 1))
	og.POST("/dispute/create", CreateDispute)
	w := doJSONFull(t, outsider, "POST", "/dispute/create", map[string]interface{}{
		"order_id": order.ID, "reason": "无关",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("第三方发起争议应403，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestCreateDispute_DuplicateRejected(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000001", 1)
	createTestUser(t, "13910000021", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	model.DB.Create(&order)

	// 预置一个未结争议
	model.DB.Create(&model.Dispute{OrderID: order.ID, InitiatorID: 1, Status: "evidence", Reason: "x"})

	w := doJSONFull(t, r, "POST", "/api/v1/dispute/create", map[string]interface{}{
		"order_id": order.ID, "reason": "再次争议",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("已有未结争议应被拒，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSettleMilestone_DisputedFrozen(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000001", 1)
	createTestUser(t, "13910000021", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	model.DB.Create(&order)
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "款", Sequence: 1, Ratio: 100, Amount: 5000, Status: "accepted"}
	model.DB.Create(&ms)

	// 存在未结争议
	model.DB.Create(&model.Dispute{OrderID: order.ID, MilestoneID: ms.ID, InitiatorID: 1, Status: "review", Reason: "y"})

	w := doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms.ID))+"/settle", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("争议中节点不可结算，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestUploadDeliverable_WrongState(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000001", 1)
	createTestUser(t, "13910000021", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	model.DB.Create(&order)
	// 已结算节点不可再上传
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "款", Sequence: 1, Ratio: 100, Amount: 5000, Status: "settled"}
	model.DB.Create(&ms)

	w := doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms.ID))+"/deliver", map[string]interface{}{
		"file_name": "x.pdf", "file_url": "https://cos/x",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("已结算节点不可上传，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestConfirmAcceptance_NotSubmitted(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000001", 1)
	createTestUser(t, "13910000021", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 5000, Status: 1}
	model.DB.Create(&order)
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "款", Sequence: 1, Ratio: 100, Amount: 5000, Status: "pending"}
	model.DB.Create(&ms)

	w := doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms.ID))+"/accept", map[string]interface{}{
		"accept": true,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("未提交交付不可验收，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestInviteSuppliers_NotOwner 非业主邀请被拒
func TestInviteSuppliers_NotOwner(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000001", 1)
	createTestUser(t, "13910000021", 2)
	projectID := setupPublishedProject(t, r)
	model.DB.Model(&model.Project{}).Where("id = ?", projectID).Update("user_id", 99)

	outsider := gin.New()
	og := outsider.Group("")
	og.Use(AuthTestMiddleware(1, 1))
	og.POST("/api/v1/project/:id/invite", InviteSuppliers)
	w := doJSONFull(t, outsider, "POST", "/api/v1/project/"+strconv.Itoa(int(projectID))+"/invite", map[string]interface{}{
		"supplier_ids": []uint{2},
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("非业主邀请应403，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestGetRecommendations_NoProject 项目不存在
func TestGetRecommendations_NoProject(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/project/9999/recommend", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// 覆盖 CreateProject 中 deadline 解析分支
func TestCreateProject_WithDeadline(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13910000040", 1)
	w := doJSONFull(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"project_type": "cost", "title": "带截止日期",
		"deadline": "2026-12-31T00:00:00Z",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("创建项目失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	proj := out["project"].(map[string]interface{})
	if proj["deadline"] == nil {
		t.Fatalf("deadline 未解析: %v", proj)
	}
}

func setupProject(t *testing.T) uint {
	t.Helper()
	createTestUser(t, "13910000099", 1)
	user := model.DB.First(&model.User{}).Error
	_ = user
	proj := model.Project{UserID: 1, ProjectType: "cost", ServiceType: "cost", Title: "t", Status: 1}
	model.DB.Create(&proj)
	return proj.ID
}

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder) gin.H {
	t.Helper()
	return decodeBody(t, w)
}
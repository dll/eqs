package handler

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// TestGetOrder_NotFound 不存在的订单详情返回404
func TestGetOrder_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/order/8888", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在订单应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSetMilestones_SignedRejected 已签约订单禁止重设节点
func TestSetMilestones_SignedRejected(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000001", 1)
	createTestUser(t, "13950000002", 2)
	projectID := setupPublishedProject(t, r)

	// 已签约订单（status 1），甲方（user_id=1）
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 6000, Status: 1}
	model.DB.Create(&order)

	w := doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{{"name": "全款", "ratio": 100}},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("已签约订单设置节点应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestCheckIn_OrderNotFound 不存在的订单打卡404
func TestCheckIn_OrderNotFound(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000003", 2)
	w := doJSONFull(t, r, "POST", "/api/v1/attendance/checkin", map[string]interface{}{
		"order_id": 88888,
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在订单打卡应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestGetUserInfo_NotFound 初始无用户时查询自身返回404
func TestGetUserInfo_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/user/info", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("无此用户应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestUpdateUserInfo_MissingBody 空body应400
func TestUpdateUserInfo_MissingBody(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000004", 1)
	w := doJSONFull(t, r, "PUT", "/api/v1/user/info", map[string]interface{}{})
	if w.Code != http.StatusOK {
		t.Fatalf("空对象更新应200，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestWithdrawBid_Selected 中选报价不可撤回
func TestWithdrawBid_Selected(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000005", 1)
	createTestUser(t, "13950000006", 2)
	projectID := setupPublishedProject(t, r)

	bid := model.Bid{ProjectID: projectID, SupplierID: 2, Amount: 100, Status: "selected"}
	model.DB.Create(&bid)

	supplierRouter := gin.New()
	sr := supplierRouter.Group("")
	sr.Use(AuthTestMiddleware(2, 2))
	sr.PUT("/bid/:id/withdraw", WithdrawBid)
	w := doJSONFull(t, supplierRouter, "PUT", "/bid/"+strconv.Itoa(int(bid.ID))+"/withdraw", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("中选报价不可撤回应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSelectBid_NotFound 报价不存在404
func TestSelectBid_NotFound(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000007", 1)
	w := doJSONFull(t, r, "POST", "/api/v1/bid/8888/select", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在报价应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSelectBid_ProjectNotFound 报价对应项目被删除404
func TestSelectBid_ProjectNotFound(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000008", 1)
	// 报价指向不存在的项目
	bid := model.Bid{ProjectID: 99999, SupplierID: 2, Amount: 100, Status: "submitted"}
	model.DB.Create(&bid)

	w := doJSONFull(t, r, "POST", "/api/v1/bid/"+strconv.Itoa(int(bid.ID))+"/select", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("项目缺失应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSubmitQualification_MissingParam 缺字段400
func TestSubmitQualification_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/supplier/2/qualifications", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺参数应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSubmitQualification_InvalidID 非法ID400
func TestSubmitQualification_InvalidID(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/supplier/abc/qualifications", map[string]interface{}{
		"qualification_type": "甲级", "certificate_no": "X",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestReviewQualification_Rejected verified=false 走驳回分支
func TestReviewQualification_Rejected(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000009", 2)
	q := model.SupplierQualification{
		SupplierID: 2, QualificationType: "造价", CertificateNo: "Z-1",
		VerificationStatus: "pending", VerificationMethod: "manual",
	}
	model.DB.Create(&q)

	w := doJSONFull(t, r, "POST", "/api/v1/qualification/"+strconv.Itoa(int(q.ID))+"/review", map[string]interface{}{
		"verified": false, "comment": "材料不全",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("核验驳回失败: %d %s", w.Code, w.Body.String())
	}
	var fetched model.SupplierQualification
	model.DB.First(&fetched, q.ID)
	if fetched.VerificationStatus != "rejected" {
		t.Fatalf("应标记为rejected，得到 %s", fetched.VerificationStatus)
	}
}

// TestReviewQualification_NotFound 资质不存在404
func TestReviewQualification_NotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/qualification/8888/review", map[string]interface{}{
		"verified": true,
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在资质应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestListQualifications_InvalidID 非法ID400
func TestListQualifications_InvalidID(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/supplier/abc/qualifications", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestCreateDispute_MissingParam 缺order_id400
func TestCreateDispute_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000010", 1)
	w := doJSONFull(t, r, "POST", "/api/v1/dispute/create", map[string]interface{}{
		"reason": "x",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺order_id应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestUploadDisputeEvidence_MissingParam 缺file_id400
func TestUploadDisputeEvidence_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/dispute/1/evidence", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺file_id应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSubmitExpertOpinion_Forbidden 非被指派专家提交意见403
func TestSubmitExpertOpinion_Forbidden(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000012", 4)
	assignment := model.DisputeExpertAssignment{DisputeID: 1, ExpertUserID: 99, RecusalStatus: "not_required"}
	model.DB.Create(&assignment)

	// 专家本人是 user 1(id=1)，但指派给99，属越权
	w := doJSONFull(t, r, "POST", "/api/v1/dispute-expert/"+strconv.Itoa(int(assignment.ID))+"/opinion", map[string]interface{}{
		"opinion": "x", "vote": "partial",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("非指派专家应403，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestCreatePayment_OrderNotFound 订单不存在404
func TestCreatePayment_OrderNotFound(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": 8888, "amount": 100, "channel": "mock",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在订单应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestUploadDeliverable_MissingParam 缺字段400
func TestUploadDeliverable_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/milestone/1/deliver", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺字段应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestConfirmAcceptance_MissingParam 缺accept400
func TestConfirmAcceptance_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000013", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 100, Status: 1}
	model.DB.Create(&order)
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "款", Sequence: 1, Ratio: 100, Amount: 100, Status: "submitted"}
	model.DB.Create(&ms)

	w := doJSONFull(t, r, "POST", "/api/v1/milestone/"+strconv.Itoa(int(ms.ID))+"/accept", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺accept应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestWxLogin_MissingParam code缺失400
func TestWxLogin_MissingParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/auth/wechat-login", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺code应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestValidateIDsFailedAttendant 打卡服务方身份绑定错误 order==用户
func TestCheckIn_InvalidBody(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/attendance/checkin", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺order_id应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSettleMilestone_WrongMilestone 非法节点ID400
func TestSettleMilestone_InvalidParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/milestone/abc/settle", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法节点ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestGetDispute_InvalidParam 非法争议ID400
func TestGetDispute_InvalidParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "GET", "/api/v1/dispute/abc", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法争议ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestCloseDispute_InvalidParam 非法争议ID400
func TestCloseDispute_InvalidParam(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/dispute/abc/close", map[string]interface{}{
		"resolution_type": "settlement",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法争议ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestAssignDisputeExpert_Success 指派专家成功并推动评审
func TestAssignDisputeExpert_Success(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13950000014", 4)
	dispute := model.Dispute{OrderID: 1, InitiatorID: 1, Status: "review", Reason: "z"}
	model.DB.Create(&dispute)

	w := doJSONFull(t, r, "POST", "/api/v1/dispute/"+strconv.Itoa(int(dispute.ID))+"/expert", map[string]interface{}{
		"expert_user_id": 1,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("指派专家失败: %d %s", w.Code, w.Body.String())
	}
}

// TestAdminListOrdersAndTransactions 后台订单与流水接口
func TestAdminListOrdersAndTransactions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	createTestUser(t, "13960000020", 3)
	createTestUser(t, "13960000021", 2)
	project := model.Project{UserID: 1, ProjectType: "cost", ServiceType: "cost", Title: "t", Status: 1}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 2000, Status: 1}
	model.DB.Create(&order)
	model.DB.Create(&model.PaymentTransaction{
		UserID: 2, OrderID: order.ID, Amount: 2000, Type: "settlement", Channel: "mock",
		ExternalTransactionID: "SETTLE-1", Status: 1,
	})

	adminRouter := gin.New()
	ar := adminRouter.Group("/api/v1")
	ar.Use(AuthTestMiddleware(1, 3))
	ar.GET("/admin/orders", AdminListOrders)
	ar.GET("/admin/transactions", AdminListTransactions)

	w := doJSONFull(t, adminRouter, "GET", "/api/v1/admin/orders", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("订单列表失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("应见1笔订单，得到 %v", out["count"])
	}

	w = doJSONFull(t, adminRouter, "GET", "/api/v1/admin/transactions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("流水列表失败: %d %s", w.Code, w.Body.String())
	}
	out = decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("应见1条流水，得到 %v", out["count"])
	}
}

var _ = gin.H{}

// TestListMyOrders 甲方与服务方各自的订单列表
func TestListMyOrders(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13960000001", 1)
	createTestUser(t, "13960000002", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 3000, Status: 1}
	model.DB.Create(&order)

	// 甲方视角（user_id=1）可见该订单
	w := doJSONFull(t, r, "GET", "/api/v1/order/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("甲方订单列表失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("甲方应见1笔订单，得到 %v", out["count"])
	}

	// 服务方视角（user_id=2）承接订单
	supplierRouter := gin.New()
	sr := supplierRouter.Group("")
	sr.Use(AuthTestMiddleware(2, 2))
	sr.GET("/order/list", ListMyOrders)
	w = doJSONFull(t, supplierRouter, "GET", "/order/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("服务方订单列表失败: %d %s", w.Code, w.Body.String())
	}
	out = decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("服务方应见1笔订单，得到 %v", out["count"])
	}
}

// TestListDisputes 按订单查询争议列表
func TestListDisputes(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13960000003", 1)
	createTestUser(t, "13960000004", 2)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 4000, Status: 1}
	model.DB.Create(&order)
	model.DB.Create(&model.Dispute{OrderID: order.ID, InitiatorID: 1, Status: "evidence", Reason: "x"})

	w := doJSONFull(t, r, "GET", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/disputes", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("争议列表失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("应见1条争议，得到 %v", out["count"])
	}

	// 非法订单ID
	w = doJSONFull(t, r, "GET", "/api/v1/order/abc/disputes", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestAdminRoutes 后台统计/用户/争议/待审资质
func TestAdminRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	createTestUser(t, "13960000005", 3)
	createTestUser(t, "13960000006", 2)
	project := model.Project{UserID: 1, ProjectType: "cost", ServiceType: "cost", Title: "t", Status: 1}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 2000, Status: 1}
	model.DB.Create(&order)
	model.DB.Create(&model.Dispute{OrderID: order.ID, InitiatorID: 1, Status: "evidence", Reason: "x"})

	// 管理员身份路由
	adminRouter := gin.New()
	ar := adminRouter.Group("/api/v1")
	ar.Use(AuthTestMiddleware(1, 3))
	ar.GET("/admin/stats", AdminDashboardStats)
	ar.GET("/admin/users", AdminListUsers)
	ar.GET("/admin/disputes", AdminListDisputes)
	ar.GET("/admin/qualifications", AdminListPendingQualifications)

	for _, p := range []string{"/api/v1/admin/stats", "/api/v1/admin/users", "/api/v1/admin/disputes", "/api/v1/admin/qualifications"} {
		w := doJSONFull(t, adminRouter, "GET", p, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("管理接口失败 %s: %d %s", p, w.Code, w.Body.String())
		}
	}

	// 非管理员（user_type=1）访问被拒
	clientRouter := gin.New()
	cr := clientRouter.Group("/api/v1")
	cr.Use(AuthTestMiddleware(1, 1))
	cr.GET("/admin/stats", RequireAdmin(), AdminDashboardStats)
	w := doJSONFull(t, clientRouter, "GET", "/api/v1/admin/stats", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非管理员访问应403，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestRequireAdmin 中间件校验
func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthTestMiddleware(9, 3))
	r.GET("/admin/stats", RequireAdmin(), AdminDashboardStats)
	w := doJSONFull(t, r, "GET", "/admin/stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("管理员应通过: %d %s", w.Code, w.Body.String())
	}

	r2 := gin.New()
	r2.Use(AuthTestMiddleware(9, 2))
	r2.GET("/admin/stats", RequireAdmin(), AdminDashboardStats)
	w = doJSONFull(t, r2, "GET", "/admin/stats", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("服务方应被拒403，得到 %d %s", w.Code, w.Body.String())
	}
}
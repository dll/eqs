package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupV9Router V9 新功能测试路由（案例/计价/托管/批文校验）
func setupV9Router(userType int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")
	client := api.Group("")
	client.Use(AuthTestMiddleware(1, userType))
	{
		client.POST("/project/create", CreateProject)
		client.POST("/project/upload", UploadProjectFile)
		client.POST("/case/create", CreateCase)
		client.GET("/case/mine", ListMyCases)
		client.DELETE("/case/:id", DeleteCase)
		client.POST("/tools/cost-estimate", CostEstimate)
		client.GET("/order/:id/escrow", GetOrderEscrow)
		client.POST("/milestone/:id/settle", SettleMilestone)
		client.POST("/dispute/create", CreateDispute)
	}
	return r
}

// TestCreateProject_RequiresApproval PRD 3.1.1：预算≥50万需上传立项批文
func TestCreateProject_RequiresApproval(t *testing.T) {
	r := setupV9Router(1)
	uid := createTestUser(t, "13910000001", 1)

	// 无批文 → 400
	w := doJSON(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"project_type": "cost", "service_type": "cost", "title": "大额项目无批文",
		"budget_min": 600000, "budget_max": 800000, "approval_file_id": 0,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("无批文大额项目应返回400，得到 %d: %s", w.Code, w.Body.String())
	}

	// 先上传批文附件，再创建 → 200
	file := model.ProjectFile{UploaderID: uid, OriginalName: "批文.pdf", FileType: "pdf", StorageKey: "uploads/projects/202608/approval.pdf", Version: 1}
	if err := model.DB.Create(&file).Error; err != nil {
		t.Fatal(err)
	}
	w = doJSON(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"project_type": "cost", "service_type": "cost", "title": "大额项目带批文",
		"budget_min": 600000, "budget_max": 800000, "approval_file_id": file.ID,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("带批文大额项目应返回200，得到 %d: %s", w.Code, w.Body.String())
	}

	// 小金额无需批文 → 200
	w = doJSON(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"project_type": "cost", "service_type": "cost", "title": "小额项目",
		"budget_min": 10000, "budget_max": 20000,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("小额项目应返回200，得到 %d: %s", w.Code, w.Body.String())
	}
}

// TestCaseFlow 案例沉淀：仅已完成订单可沉淀；我的案例列表与删除
func TestCaseFlow(t *testing.T) {
	r := setupV9Router(2) // 服务方
	supplierID := createTestUser(t, "13910000002", 2)
	clientID := createTestUser(t, "13910000003", 1)

	// 创建订单（未完成 status=1）
	project := model.Project{UserID: clientID, Title: "测试项目", ServiceType: "cost", BudgetMin: 1000, BudgetMax: 5000, Status: 1}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: supplierID, Amount: 3000, Status: 1}
	model.DB.Create(&order)

	// 未完成订单不可沉淀 → 400
	w := doJSON(t, r, "POST", "/api/v1/case/create", map[string]interface{}{"order_id": order.ID, "title": "未完成订单案例"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("未完成订单沉淀案例应返回400，得到 %d: %s", w.Code, w.Body.String())
	}

	// 标记订单完成（status=3）
	model.DB.Model(&order).Update("status", 3)

	// 已完成订单可沉淀 → 200
	w = doJSON(t, r, "POST", "/api/v1/case/create", map[string]interface{}{"order_id": order.ID, "title": "办公楼造价成果案例"})
	if w.Code != http.StatusOK {
		t.Fatalf("已完成订单沉淀案例应返回200，得到 %d: %s", w.Code, w.Body.String())
	}

	// 我的案例列表
	w = doJSON(t, r, "GET", "/api/v1/case/mine", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("我的案例应返回200，得到 %d", w.Code)
	}
	var resp struct {
		Cases []model.CaseShowcase `json:"cases"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.Cases) != 1 || resp.Cases[0].Title != "办公楼造价成果案例" {
		t.Fatalf("我的案例数据异常: %+v", resp.Cases)
	}

	// 删除案例
	w = doJSON(t, r, "DELETE", "/api/v1/case/"+itoa(resp.Cases[0].ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("删除案例应返回200，得到 %d: %s", w.Code, w.Body.String())
	}
}

// TestEscrowFlow 资金托管台账：结算释放 + 争议冻结 + 明细查询
func TestEscrowFlow(t *testing.T) {
	r := setupV9Router(1) // 甲方（user_id=1，先创建以匹配中间件注入）
	clientID := createTestUser(t, "13910000005", 1)
	supplierID := createTestUser(t, "13910000004", 2)

	project := model.Project{UserID: clientID, Title: "托管测试项目", ServiceType: "geotech", BudgetMin: 1000, BudgetMax: 9000, Status: 2}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: supplierID, Amount: 6000, Status: 1}
	model.DB.Create(&order)
	ms := model.PaymentMilestone{OrderID: order.ID, Name: "预付款", Sequence: 1, Ratio: 50, Amount: 3000, Status: "accepted"}
	model.DB.Create(&ms)

	// 结算节点 → 托管释放记录
	w := doJSON(t, r, "POST", "/api/v1/milestone/"+itoa(ms.ID)+"/settle", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("结算应返回200，得到 %d: %s", w.Code, w.Body.String())
	}

	// 发起争议（该订单另一节点）→ 冻结
	ms2 := model.PaymentMilestone{OrderID: order.ID, Name: "尾款", Sequence: 2, Ratio: 50, Amount: 3000, Status: "submitted"}
	model.DB.Create(&ms2)
	w = doJSON(t, r, "POST", "/api/v1/dispute/create", map[string]interface{}{
		"order_id": order.ID, "milestone_id": ms2.ID, "reason": "交付不符", "claim": "重新核算",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("发起争议应返回200，得到 %d: %s", w.Code, w.Body.String())
	}

	// 查询托管明细：托管总额 6000，已释放 3000，冻结 3000，余额 3000
	w = doJSON(t, r, "GET", "/api/v1/order/"+itoa(order.ID)+"/escrow", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("托管明细应返回200，得到 %d", w.Code)
	}
	var esc struct {
		EscrowTotal float64 `json:"escrow_total"`
		Released    float64 `json:"released"`
		Frozen      float64 `json:"frozen"`
		Balance     float64 `json:"balance"`
		LedgerCount int     `json:"ledger_count"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &esc)
	if esc.EscrowTotal != 6000 || esc.Released != 3000 || esc.Frozen != 3000 || esc.Balance != 3000 {
		t.Fatalf("托管汇总异常: %+v", esc)
	}
	if esc.LedgerCount != 2 { // 1 release + 1 freeze
		t.Fatalf("台账记录数 = %d, 期望 2", esc.LedgerCount)
	}
}

// TestCostEstimateAPI 计价估算接口
func TestCostEstimateAPI(t *testing.T) {
	r := setupV9Router(1)
	createTestUser(t, "13910000006", 1)

	w := doJSON(t, r, "POST", "/api/v1/tools/cost-estimate", map[string]interface{}{
		"items": []map[string]interface{}{
			{"name": "挖土方", "unit": "m³", "quantity": 100, "unit_price": 30},
		},
		"rates": map[string]interface{}{},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("计价估算应返回200，得到 %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Result CostEstimateResult `json:"result"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Result.Subtotal != 3000 || resp.Result.Total <= 3000 {
		t.Fatalf("计价结果异常: %+v", resp.Result)
	}
}

func itoa(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

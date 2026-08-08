package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// TestSmokeCoreFlow 冒烟：验证服务启动路由完整 + 核心业务闭环（sqlite 替代 MySQL）
func TestSmokeCoreFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := model.InitTestDB()
	_ = db

	cfg := config.Load()
	r := setupRouter(db, cfg)

	// 1) 手机号登录（甲方）
	w := doSmoke(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13800000001", "code": "123456", "user_type": 1,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("登录失败: %d %s", w.Code, w.Body.String())
	}
	login := decodeSmoke(t, w)
	token := login["token"].(string)
	userID := uint(login["user"].(map[string]interface{})["id"].(float64))
	_ = userID

	clientHeader := map[string]string{"Authorization": "Bearer " + token}

	// 2) 甲方发布项目
	w = doSmokeH(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"project_type": "cost", "service_type": "cost", "title": "冒烟控制价项目",
		"budget_min": 10000, "budget_max": 50000,
	}, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("发布项目失败: %d %s", w.Code, w.Body.String())
	}
	proj := decodeSmoke(t, w)
	projectID := uint(proj["project"].(map[string]interface{})["id"].(float64))

	// 3) 服务方登录并报价
	w = doSmoke(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13800000002", "code": "123456", "user_type": 2,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("服务方登录失败: %d %s", w.Code, w.Body.String())
	}
	supplierLogin := decodeSmoke(t, w)
	supplierToken := supplierLogin["token"].(string)
	supplierHeader := map[string]string{"Authorization": "Bearer " + supplierToken}

	w = doSmokeH(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 20000, "service_days": 30,
	}, supplierHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("报价失败: %d %s", w.Code, w.Body.String())
	}
	bidOut := decodeSmoke(t, w)
	bidID := uint(bidOut["bid"].(map[string]interface{})["id"].(float64))

	// 4) 甲方中选
	w = doSmokeH(t, r, "POST", "/api/v1/bid/"+itoa(bidID)+"/select", nil, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("中选失败: %d %s", w.Code, w.Body.String())
	}

	// 5) 获取订单并设置节点
	var order model.Order
	if err := model.DB.Where("project_id = ?", projectID).First(&order).Error; err != nil {
		t.Fatalf("订单未生成: %v", err)
	}
	w = doSmokeH(t, r, "PUT", "/api/v1/order/"+itoa(order.ID)+"/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{
			{"name": "开工款", "ratio": 30},
			{"name": "验收款", "ratio": 70},
		},
	}, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("设置节点失败: %d %s", w.Code, w.Body.String())
	}

	// 6) 生成并签署合同
	w = doSmokeH(t, r, "POST", "/api/v1/order/"+itoa(order.ID)+"/contract", nil, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("生成合同失败: %d %s", w.Code, w.Body.String())
	}
	contractOut := decodeSmoke(t, w)
	contractID := uint(contractOut["contract"].(map[string]interface{})["id"].(float64))

	w = doSmokeH(t, r, "POST", "/api/v1/contract/"+itoa(contractID)+"/sign", nil, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("签署合同失败: %d %s", w.Code, w.Body.String())
	}

	// 7) 支付-结算（甲方向第三方提交支付单）
	w = doSmokeH(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": order.ID, "amount": order.Amount, "channel": "mock",
	}, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("支付失败: %d %s", w.Code, w.Body.String())
	}

	// 8) 服务方交付、甲方验收、结算
	var ms model.PaymentMilestone
	model.DB.Where("order_id = ? AND sequence = 1", order.ID).First(&ms)

	w = doSmokeH(t, r, "POST", "/api/v1/milestone/"+itoa(ms.ID)+"/deliver", map[string]interface{}{
		"file_name": "清单.xlsx", "file_url": "https://cos/budget.xlsx",
	}, supplierHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("交付失败: %d %s", w.Code, w.Body.String())
	}

	w = doSmokeH(t, r, "POST", "/api/v1/milestone/"+itoa(ms.ID)+"/accept", map[string]interface{}{
		"accept": true, "comment": "合格",
	}, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("验收失败: %d %s", w.Code, w.Body.String())
	}

	w = doSmokeH(t, r, "POST", "/api/v1/milestone/"+itoa(ms.ID)+"/settle", nil, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("结算失败: %d %s", w.Code, w.Body.String())
	}

	// 9) 订单详情应完整（节点、合同、资金）
	w = doSmokeH(t, r, "GET", "/api/v1/order/"+itoa(order.ID), nil, clientHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("订单详情失败: %d %s", w.Code, w.Body.String())
	}

	// 10) 管理员统计（需管理员账号）
	w = doSmoke(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13800000003", "code": "123456", "user_type": 3,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("管理员登录失败: %d %s", w.Code, w.Body.String())
	}
	adminLogin := decodeSmoke(t, w)
	adminHeader := map[string]string{"Authorization": "Bearer " + adminLogin["token"].(string)}
	w = doSmokeH(t, r, "GET", "/api/v1/admin/stats", nil, adminHeader)
	if w.Code != http.StatusOK {
		t.Fatalf("管理员统计失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeSmoke(t, w)
	if out["order_count"].(float64) != 1 {
		t.Fatalf("统计订单数应1，得到 %v", out["order_count"])
	}

	t.Log("冒烟通过：登录→发布→报价→中选→节点→合同→支付→交付→验收→结算→统计")
}

func itoa(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

func doSmoke(t *testing.T, r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	return doSmokeH(t, r, method, path, body, nil)
}

func doSmokeH(t *testing.T, r *gin.Engine, method, path string, body interface{}, header map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeSmoke(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("解析响应失败(%d): %v -> %s", w.Code, err, w.Body.String())
	}
	return out
}
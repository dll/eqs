package handler

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
)

// TestAuditLogWrites 关键状态变更可还原：发布、报价、中选、节点、签约、结算
func TestAuditLogWrites(t *testing.T) {
	r := setupFlowRouter()

	// 发布项目
	w := doJSONFull(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"project_type": "cost", "service_type": "cost",
		"title": "审计留痕项目", "budget_min": 10000, "budget_max": 50000,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("创建项目失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	projectID := uint(out["project"].(map[string]interface{})["id"].(float64))
	auditCount(t, "project.publish", 1)

	// 服务方报价（换路由组模拟服务方）
	w = doJSONFull(t, r, "POST", "/api/v1/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 12000, "service_days": 20,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("报价失败: %d %s", w.Code, w.Body.String())
	}
	bidOut := decodeBody(t, w)
	bidID := uint(bidOut["bid"].(map[string]interface{})["id"].(float64))
	auditCount(t, "bid.submit", 1)

	// 甲方中选
	w = doJSONFull(t, r, "POST", "/api/v1/bid/"+u64(bidID)+"/select", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("中选失败: %d %s", w.Code, w.Body.String())
	}
	auditCount(t, "bid.select", 1)

	// 设置节点
	var order model.Order
	model.DB.Where("project_id = ?", projectID).First(&order)
	w = doJSONFull(t, r, "PUT", "/api/v1/order/"+u64(order.ID)+"/milestones", map[string]interface{}{
		"milestones": []map[string]interface{}{
			{"name": "开工款", "ratio": 50},
			{"name": "验收款", "ratio": 50},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("设置节点失败: %d %s", w.Code, w.Body.String())
	}
	auditCount(t, "order.milestones", 1)

	// 生成并签署合同
	w = doJSONFull(t, r, "POST", "/api/v1/order/"+u64(order.ID)+"/contract", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("生成合同失败: %d %s", w.Code, w.Body.String())
	}
	cOut := decodeBody(t, w)
	contractID := uint(cOut["contract"].(map[string]interface{})["id"].(float64))
	w = doJSONFull(t, r, "POST", "/api/v1/contract/"+u64(contractID)+"/sign", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("签署失败: %d %s", w.Code, w.Body.String())
	}
	auditCount(t, "contract.sign", 1)

	// 创建支付单
	w = doJSONFull(t, r, "POST", "/api/v1/pay/create", map[string]interface{}{
		"order_id": order.ID, "amount": order.Amount, "channel": "mock",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("支付失败: %d %s", w.Code, w.Body.String())
	}
	auditCount(t, "payment.create", 1)

	// 完整性回溯：按时间倒序还原全部关键动作
	var logs []model.AuditLog
	model.DB.Order("created_at DESC").Find(&logs)
	actions := map[string]bool{}
	for _, l := range logs {
		actions[l.Action] = true
	}
	want := []string{"project.publish", "bid.submit", "bid.select", "order.milestones", "contract.sign", "payment.create"}
	for _, a := range want {
		if !actions[a] {
			t.Fatalf("审计日志缺动作: %s", a)
		}
	}
	if len(logs) < len(want) {
		t.Fatalf("审计记录过少: %d", len(logs))
	}
	t.Logf("审计留痕通过，共 %d 条", len(logs))
}

// TestAuditLog_LoginRecorded 登录动作记录在案
func TestAuditLog_LoginRecorded(t *testing.T) {
	r := setupFlowRouter()
	w := doJSON(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13800008888", "code": "123456", "user_type": 1,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("登录失败: %d %s", w.Code, w.Body.String())
	}
	var cnt int64
	model.DB.Model(&model.AuditLog{}).Where("action = ?", "user.login").Count(&cnt)
	if cnt < 1 {
		t.Fatalf("应有登录审计记录，实际 %d", cnt)
	}
}

// auditCount 断言某动作的审计记录数
func auditCount(t *testing.T, action string, want int64) {
	t.Helper()
	var cnt int64
	model.DB.Model(&model.AuditLog{}).Where("action = ?", action).Count(&cnt)
	if cnt != want {
		t.Fatalf("action=%s 应记录 %d 条，实际 %d", action, want, cnt)
	}
}

// u64 无符号整数转字符串
func u64(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
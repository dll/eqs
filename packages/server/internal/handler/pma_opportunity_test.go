package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func TestPMAOpportunityOfflineCaptureListFilterAndAudit(t *testing.T) {
	r := setupPMARouter()
	org := model.Organization{Name: "PMA商机组织", Code: "PMA-OPP", Type: "internal", Status: 1}
	if err := model.DB.Create(&org).Error; err != nil {
		t.Fatal(err)
	}
	if err := model.DB.Create(&model.OrgMember{OrgID: org.ID, UserID: 2, Role: "member", Status: 1}).Error; err != nil {
		t.Fatal(err)
	}

	w := doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities", map[string]interface{}{"title": "线下软件项目", "owner_org": "甲方A", "summary": "内部录入", "source": "offline"}, 2, 1)
	if w.Code != http.StatusOK {
		t.Fatalf("录入应成功，得到 %d: %s", w.Code, w.Body.String())
	}
	w = doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities", map[string]interface{}{"title": "线上候选项目", "source": "online"}, 2, 1)
	if w.Code != http.StatusOK {
		t.Fatalf("第二条录入应成功，得到 %d: %s", w.Code, w.Body.String())
	}
	w = doJSONFullAuth(t, r, "GET", "/api/v1/pma/opportunities?source=offline&q=线下", nil, 2, 1)
	if w.Code != http.StatusOK {
		t.Fatalf("列表筛选应成功，得到 %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "线下软件项目") || strings.Contains(w.Body.String(), "线上候选项目") {
		t.Fatalf("筛选结果不正确: %s", w.Body.String())
	}

	var auditCount int64
	model.DB.Model(&model.AuditLog{}).Where("action = ?", "pma.opportunity.create").Count(&auditCount)
	if auditCount != 2 {
		t.Fatalf("商机录入审计应为2条，得到 %d", auditCount)
	}

	w = doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities", map[string]interface{}{"title": "管理员越权"}, 3, 3)
	if w.Code != http.StatusForbidden {
		t.Fatalf("系统管理员不应隐式获得PMA商机录入权限，得到 %d", w.Code)
	}
}

func TestPMAOpportunityRejectsInvalidInput(t *testing.T) {
	r := setupPMARouter()
	org := model.Organization{Name: "PMA校验组织", Code: "PMA-VALID", Type: "internal", Status: 1}
	model.DB.Create(&org)
	model.DB.Create(&model.OrgMember{OrgID: org.ID, UserID: 2, Role: "member", Status: 1})
	w := doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities", map[string]interface{}{"title": "", "status": "bad"}, 2, 1)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法商机应400，得到 %d", w.Code)
	}
}

func setupPMAOpportunityStatusRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	r := gin.New()
	pma := r.Group("/api/v1/pma")
	pma.Use(AuthTestMiddleware(3, 3))
	pma.POST("/opportunities", CreatePMAOpportunity)
	pma.PUT("/opportunities/:id/status", UpdatePMAOpportunityStatus)
	return r
}

func TestPMAOpportunityStatusTransitionMatrixAuditAndPermission(t *testing.T) {
	r := setupPMAOpportunityStatusRouter()
	org := model.Organization{Name: "PMA状态组织", Code: "PMA-STATUS", Type: "internal", Status: 1}
	if err := model.DB.Create(&org).Error; err != nil {
		t.Fatal(err)
	}
	if err := model.DB.Create(&model.OrgMember{OrgID: org.ID, UserID: 2, Role: "member", Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	create := func(title, status string) uint {
		body := map[string]interface{}{"source": "test", "title": title}
		if status != "" {
			body["status"] = status
		}
		w := doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities", body, 2, 1)
		if w.Code != http.StatusOK {
			t.Fatalf("创建失败: %d %s", w.Code, w.Body.String())
		}
		var op model.PMAOpportunity
		if err := model.DB.Where("title = ?", title).First(&op).Error; err != nil {
			t.Fatal(err)
		}
		return op.ID
	}
	update := func(id uint, status, reason string, want int) {
		w := doJSONFullAuth(t, r, "PUT", "/api/v1/pma/opportunities/"+u64(id)+"/status", map[string]string{"status": status, "reason": reason}, 2, 1)
		if w.Code != want {
			t.Fatalf("转换到 %s 状态码=%d, body=%s", status, w.Code, w.Body.String())
		}
	}

	id := create("矩阵机会", "")
	update(id, "triaged", "初筛通过", http.StatusOK)
	update(id, "closed", "需求不匹配", http.StatusOK)
	update(id, "new", "误操作回退", http.StatusConflict)
	update(id, "closed", "重复关闭", http.StatusConflict)
	update(id, "triaged", "重复筛选", http.StatusConflict)
	update(id, "new", "", http.StatusBadRequest)
	var audit model.AuditLog
	if err := model.DB.Where("action = ? AND target_id = ?", "pma.opportunity.status", id).Order("id DESC").First(&audit).Error; err != nil {
		t.Fatal(err)
	}
	var detail map[string]interface{}
	if err := json.Unmarshal([]byte(audit.Detail), &detail); err != nil {
		t.Fatal(err)
	}
	if detail["from"] != "triaged" || detail["to"] != "closed" || detail["reason"] != "需求不匹配" {
		t.Fatalf("审计状态细节不完整: %v", detail)
	}
	principal, ok := detail["operator_principal"].(map[string]interface{})
	if !ok || principal["type"] != "user" || principal["user_id"] != float64(2) {
		t.Fatalf("审计operator principal不完整: %v", detail["operator_principal"])
	}

	id = create("初始筛选", "triaged")
	update(id, "new", "禁止回退", http.StatusConflict)
	id = create("初始关闭", "closed")
	update(id, "new", "禁止恢复", http.StatusConflict)
	update(id, "triaged", "禁止恢复", http.StatusConflict)
	w := doJSONFullAuth(t, r, "PUT", "/api/v1/pma/opportunities/"+u64(id)+"/status", map[string]string{"status": "new", "reason": "越权"}, 3, 3)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非PMA成员应403，得到%d", w.Code)
	}
}

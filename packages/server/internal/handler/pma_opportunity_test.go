package handler

import (
	"net/http"
	"strings"
	"testing"

	"github.com/eqs/server/internal/model"
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

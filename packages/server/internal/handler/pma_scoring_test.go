package handler

import (
	"github.com/eqs/server/internal/model"
	"net/http"
	"testing"
)

func TestPMAScoringAPIManualAuditAIDisabledAndInvalidScore(t *testing.T) {
	r := setupPMARouter()
	org := model.Organization{Name: "评分组织", Code: "SCORE-ORG", Type: "internal", Status: 1}
	model.DB.Create(&org)
	model.DB.Create(&model.OrgMember{OrgID: org.ID, UserID: 2, Role: "member", Status: 1})
	w := doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities", map[string]interface{}{"title": "评分商机", "owner_org": "SCORE-ORG"}, 2, 1)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}
	var op model.PMAOpportunity
	model.DB.Where("title = ?", "评分商机").First(&op)
	w = doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities/"+u64(op.ID)+"/scores", nil, 2, 1)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}
	var s model.PMAScore
	model.DB.Where("opportunity_id = ?", op.ID).First(&s)
	var item model.PMAScoreItem
	model.DB.Where("score_id = ?", s.ID).First(&item)
	w = doJSONFullAuth(t, r, "PUT", "/api/v1/pma/scores/"+u64(s.ID)+"/items/"+u64(item.ID), map[string]interface{}{"score": 6}, 2, 1)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid score: %d", w.Code)
	}
	w = doJSONFullAuth(t, r, "PUT", "/api/v1/pma/scores/"+u64(s.ID)+"/items/"+u64(item.ID), map[string]interface{}{"score": 5, "source": "manual"}, 2, 1)
	if w.Code != http.StatusOK {
		t.Fatal(w.Body.String())
	}
	w = doJSONFullAuth(t, r, "POST", "/api/v1/pma/scores/"+u64(s.ID)+"/ai-suggestions", map[string]interface{}{"suggestions": []string{"x"}, "confidence": 0.7}, 2, 1)
	if w.Code != http.StatusOK || !contains(w.Body.String(), `"applied":false`) {
		t.Fatalf("ai placeholder: %d %s", w.Code, w.Body.String())
	}
	var count int64
	model.DB.Model(&model.AuditLog{}).Where("action IN ?", []string{"pma.score.create", "pma.score.item.update", "pma.score.ai.suggest"}).Count(&count)
	if count != 3 {
		t.Fatalf("expected 3 score audit entries, got %d", count)
	}
}

func TestPMAScoringOrganizationIsolation(t *testing.T) {
	r := setupPMARouter()
	a := model.Organization{Name: "A", Code: "A-ORG", Type: "internal", Status: 1}
	b := model.Organization{Name: "B", Code: "B-ORG", Type: "internal", Status: 1}
	model.DB.Create(&a)
	model.DB.Create(&b)
	model.DB.Create(&model.OrgMember{OrgID: a.ID, UserID: 2, Status: 1})
	model.DB.Create(&model.OrgMember{OrgID: b.ID, UserID: 4, Status: 1})
	model.DB.Create(&model.PMAOpportunity{Title: "A机会", OwnerOrg: "A-ORG"})
	var op model.PMAOpportunity
	model.DB.Where("title = ?", "A机会").First(&op)
	model.DB.Create(&model.PMAScore{OpportunityID: op.ID, Version: 1})
	var s model.PMAScore
	model.DB.Where("opportunity_id = ?", op.ID).First(&s)
	w := doJSONFullAuth(t, r, "GET", "/api/v1/pma/scores/"+u64(s.ID), nil, 4, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cross-org read must be forbidden, got %d", w.Code)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func setupPMARouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	r := gin.New()
	api := r.Group("/api/v1")
	pma := api.Group("/pma")
	pma.Use(AuthTestMiddleware(3, 3))
	{
		pma.GET("/projects", PMAListProjects)
		pma.GET("/projects/:id/overview", PMAProjectOverview)
		pma.PUT("/projects/:id/todos/:tid/decision", PMADecideTodo)
		pma.POST("/opportunities", CreatePMAOpportunity)
		pma.GET("/opportunities", ListPMAOpportunities)
	}
	return r
}

func TestPMA_ProjectManagementViewAggregatesExistingProject(t *testing.T) {
	r := setupPMARouter()
	projectID := mkArchiveProject(t, setupArchiveManageRouter(), "PMA真实项目")
	model.DB.Create(&model.ArchivedMilestone{ArchivedProjectID: projectID, Name: "需求确认", Status: "done"})
	model.DB.Create(&model.ArchivedMilestone{ArchivedProjectID: projectID, Name: "阶段验收", Status: "todo"})
	model.DB.Create(&model.ArchivedRisk{ArchivedProjectID: projectID, Title: "延期风险", Level: "high", Status: "open"})
	todo := model.ArchivedTodo{ArchivedProjectID: projectID, Title: "OPC确认验收人", Kind: "decision", Status: "open"}
	model.DB.Create(&todo)

	w := doJSONFullAuth(t, r, "GET", "/api/v1/pma/projects/"+u64(projectID)+"/overview", nil, 3, 3)
	if w.Code != http.StatusOK {
		t.Fatalf("PMA概览应成功: %d %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	summary := body["summary"].(map[string]interface{})
	if int(summary["milestone_total"].(float64)) != 2 || int(summary["milestone_done"].(float64)) != 1 {
		t.Fatalf("里程碑聚合错误: %s", w.Body.String())
	}
	if int(summary["risk_open"].(float64)) != 1 || int(summary["todo_open"].(float64)) != 1 {
		t.Fatalf("风险/待办聚合错误: %s", w.Body.String())
	}

	w = doJSONFullAuth(t, r, "GET", "/api/v1/pma/projects", nil, 3, 3)
	if w.Code != http.StatusOK {
		t.Fatalf("PMA项目列表应成功: %d %s", w.Code, w.Body.String())
	}
	if len(decodeBody(t, w)["projects"].([]interface{})) != 1 {
		t.Fatalf("PMA项目列表应只返回现有项目: %s", w.Body.String())
	}
}

func TestPMA_OPCPermissionsAndDecisionAudit(t *testing.T) {
	r := setupPMARouter()
	projectID := mkArchiveProject(t, setupArchiveManageRouter(), "PMA权限项目")
	todo := model.ArchivedTodo{ArchivedProjectID: projectID, Title: "需要OPC决策", Kind: "decision", Status: "open"}
	model.DB.Create(&todo)

	// 非 internal 成员只能被拒绝，不能改变决策待办。
	w := doJSONFullAuth(t, r, "PUT", "/api/v1/pma/projects/"+u64(projectID)+"/todos/"+u64(todo.ID)+"/decision", map[string]interface{}{"status": "done"}, 1, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非OPC决策应403，得到 %d", w.Code)
	}
	// 平台管理员是系统运维角色，不能代替公司 OPC 真人定案。
	w = doJSONFullAuth(t, r, "PUT", "/api/v1/pma/projects/"+u64(projectID)+"/todos/"+u64(todo.ID)+"/decision", map[string]interface{}{"status": "done"}, 3, 3)
	if w.Code != http.StatusForbidden {
		t.Fatalf("系统管理员不得代替OPC决策，应403，得到 %d", w.Code)
	}

	org := model.Organization{Name: "PMA测试内部组织", Code: "PMA-TEST", Type: "internal", Status: 1}
	if err := model.DB.Create(&org).Error; err != nil {
		t.Fatal(err)
	}
	if err := model.DB.Create(&model.OrgMember{OrgID: org.ID, UserID: 2, Role: "member", Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	w = doJSONFullAuth(t, r, "GET", "/api/v1/pma/projects/"+u64(projectID)+"/overview", nil, 2, 1)
	if w.Code != http.StatusOK {
		t.Fatalf("internal成员应可查看项目，得到 %d: %s", w.Code, w.Body.String())
	}
	w = doJSONFullAuth(t, r, "PUT", "/api/v1/pma/projects/"+u64(projectID)+"/todos/"+u64(todo.ID)+"/decision", map[string]interface{}{"status": "done"}, 2, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("internal普通成员不应定案，得到 %d", w.Code)
	}

	if err := model.DB.Create(&model.OrgMember{OrgID: org.ID, UserID: 3, Role: "owner", Status: 1}).Error; err != nil {
		t.Fatal(err)
	}
	w = doJSONFullAuth(t, r, "PUT", "/api/v1/pma/projects/"+u64(projectID)+"/todos/"+u64(todo.ID)+"/decision", map[string]interface{}{"status": "done"}, 3, 1)
	if w.Code != http.StatusOK {
		t.Fatalf("OPC owner应可定案，得到 %d: %s", w.Code, w.Body.String())
	}
	var got model.ArchivedTodo
	model.DB.First(&got, todo.ID)
	if got.Status != "done" {
		t.Fatalf("待办状态未更新: %q", got.Status)
	}
	var auditCount int64
	model.DB.Model(&model.AuditLog{}).Where("action = ? AND target_id = ?", "pma.todo.decide", projectID).Count(&auditCount)
	if auditCount != 1 {
		t.Fatalf("PMA定案审计应为1条，得到 %d", auditCount)
	}
}

func TestPMA_OpportunityWorkflowIsControlled(t *testing.T) {
	r := setupPMARouter()
	w := doJSONFullAuth(t, r, "POST", "/api/v1/pma/opportunities", map[string]interface{}{"title": "管理员越权商机"}, 3, 3)
	if w.Code != http.StatusForbidden {
		t.Fatalf("PMA商机入口必须受控，系统管理员无internal成员身份应403，得到 %d", w.Code)
	}
}

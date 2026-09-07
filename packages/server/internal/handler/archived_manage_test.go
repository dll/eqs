package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupArchiveManageRouter 归档项目推进闭环(里程碑/风险/待办)测试路由
func setupArchiveManageRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	r := gin.New()
	api := r.Group("/api/v1")
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.POST("/admin/archive/projects", CreateArchivedProject)
		admin.GET("/admin/archive/projects/:id/overview", ArchiveGetOverview)
		admin.POST("/admin/archive/projects/:id/milestones", ArchiveAddMilestone)
		admin.PUT("/admin/archive/projects/:id/milestones/:mid", ArchiveUpdateMilestone)
		admin.DELETE("/admin/archive/projects/:id/milestones/:mid", ArchiveDeleteMilestone)
		admin.POST("/admin/archive/projects/:id/risks", ArchiveAddRisk)
		admin.PUT("/admin/archive/projects/:id/risks/:rid", ArchiveUpdateRisk)
		admin.DELETE("/admin/archive/projects/:id/risks/:rid", ArchiveDeleteRisk)
		admin.POST("/admin/archive/projects/:id/todos", ArchiveAddTodo)
		admin.PUT("/admin/archive/projects/:id/todos/:tid", ArchiveUpdateTodo)
		admin.DELETE("/admin/archive/projects/:id/todos/:tid", ArchiveDeleteTodo)
	}
	return r
}

// mkArchiveProject 建一个归档项目并返回 id
func mkArchiveProject(t *testing.T, r *gin.Engine, title string) uint {
	t.Helper()
	w := doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects", map[string]interface{}{
		"title": title, "owner_org": "CCIT-SIM", "status_phase": "in_progress",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("建归档项目失败 %d %s", w.Code, w.Body.String())
	}
	return uint(decodeBody(t, w)["archived_project"].(map[string]interface{})["id"].(float64))
}

func TestArchiveManage_Lifecycle(t *testing.T) {
	r := setupArchiveManageRouter()
	pid := mkArchiveProject(t, r, "第一单管理闭环-001")

	id := u64(pid)

	// 里程碑：加 2 个
	for _, name := range []string{"需求确认", "阶段验收"} {
		w := doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects/"+id+"/milestones", map[string]interface{}{"name": name, "sort_order": 1})
		if w.Code != http.StatusOK {
			t.Fatalf("加里程碑失败 %s", w.Body.String())
		}
	}
	// 完成第一个 → 进度 50%
	var ms model.ArchivedMilestone
	model.DB.Where("archived_project_id = ?", pid).Order("id ASC").First(&ms)
	w := doJSONFull(t, r, "PUT", "/api/v1/admin/archive/projects/"+id+"/milestones/"+u64(ms.ID), map[string]interface{}{"status": "done"})
	if w.Code != http.StatusOK {
		t.Fatalf("完成里程碑失败 %s", w.Body.String())
	}

	// 风险
	w = doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects/"+id+"/risks", map[string]interface{}{"level": "high", "title": "交付延期风险", "mitigation": "缓冲期"})
	if w.Code != http.StatusOK {
		t.Fatalf("加风险失败 %s", w.Body.String())
	}
	// 待办(决策)
	w = doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects/"+id+"/todos", map[string]interface{}{"title": "OPC 请定验收人", "kind": "decision"})
	if w.Code != http.StatusOK {
		t.Fatalf("加待办失败 %s", w.Body.String())
	}

	// 概览：应含 2 里程碑(1 done)/1 风险(open)/1 待办(决策)，进度 50
	w = doJSONFull(t, r, "GET", "/api/v1/admin/archive/projects/"+id+"/overview", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("概览失败 %s", w.Body.String())
	}
	b := decodeBody(t, w)
	msList := b["milestones"].([]interface{})
	rsList := b["risks"].([]interface{})
	tdList := b["todos"].([]interface{})
	sm := b["summary"].(map[string]interface{})
	if len(msList) != 2 || len(rsList) != 1 || len(tdList) != 1 {
		t.Fatalf("闭环数据不完整 ms=%d rs=%d td=%d", len(msList), len(rsList), len(tdList))
	}
	if int(sm["milestone_done"].(float64)) != 1 || int(sm["progress_percent"].(float64)) != 50 {
		t.Fatalf("进度/摘要错误: %s", w.Body.String())
	}
	// 待办 kind 应是 decision
	if (tdList[0].(map[string]interface{}))["kind"] != "decision" {
		t.Fatalf("待办 kind 应为 decision")
	}

	// 关闭风险 + 完成待办
	w = doJSONFull(t, r, "PUT", "/api/v1/admin/archive/projects/"+id+"/risks/"+u64(uint(rsList[0].(map[string]interface{})["id"].(float64))), map[string]interface{}{"status": "closed"})
	if w.Code != http.StatusOK {
		t.Fatalf("关闭风险失败 %s", w.Body.String())
	}
	w = doJSONFull(t, r, "PUT", "/api/v1/admin/archive/projects/"+id+"/todos/"+u64(uint(tdList[0].(map[string]interface{})["id"].(float64))), map[string]interface{}{"status": "done"})
	if w.Code != http.StatusOK {
		t.Fatalf("完成待办失败 %s", w.Body.String())
	}
	// 审计留痕：风险关闭 + 待办完成各有一条
	var cnt int64
	model.DB.Model(&model.AuditLog{}).Where("action = ? AND target_id = ?", "archive.risk.update", pid).Count(&cnt)
	if cnt == 0 {
		t.Fatalf("风险动作未留审计")
	}
}

func TestArchiveManage_PermissionAndGaps(t *testing.T) {
	r := setupArchiveManageRouter()
	pid := mkArchiveProject(t, r, "越权&不存在-002")
	// 非 admin 403
	w := doJSONFullAuth(t, r, "GET", "/api/v1/admin/archive/projects/"+u64(pid)+"/overview", nil, 1, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非 admin 应 403，得到 %d", w.Code)
	}
	// 不存在项目 → 404
	w = doJSONFull(t, r, "GET", "/api/v1/admin/archive/projects/999999/overview", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应 404，得到 %d", w.Code)
	}
	// 里程碑必填 title → 400
	w = doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects/"+u64(pid)+"/milestones", map[string]interface{}{"sort_order": 1})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺标题里程碑应 400，得到 %d", w.Code)
	}
}

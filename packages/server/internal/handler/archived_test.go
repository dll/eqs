package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupArchiveRouter 存量项目归档测试路由（A 批次：admin=user_type 3 平台/公司运营侧）
func setupArchiveRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	invalidatePublicCache()

	r := gin.New()
	api := r.Group("/api/v1")

	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.POST("/admin/archive/projects", CreateArchivedProject)
		admin.POST("/admin/archive/projects/batch", BatchImportArchivedProjects)
		admin.GET("/admin/archive/projects", ListArchivedProjects)
		admin.GET("/admin/archive/projects/:id", GetArchivedProject)
		admin.PUT("/admin/archive/projects/:id", UpdateArchivedProject)
		admin.DELETE("/admin/archive/projects/:id", DeleteArchivedProject)
	}
	return r
}

func TestArchiveProject_Create(t *testing.T) {
	r := setupArchiveRouter()

	w := doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects", map[string]interface{}{
		"title": "某市政道路勘察测绘（存量补录）", "description": "线下承接正在执行", "owner_org": "滁州某城投",
		"service_type": "geotech", "status_phase": "in_progress", "address": "滁州市",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("补录失败: %d %s", w.Code, w.Body.String())
	}
	raw := decodeBody(t, w)["archived_project"].(map[string]interface{})
	title := raw["title"].(string)
	if title == "" {
		t.Fatalf("返回项目缺少标题")
	}
	// 阶段回显 in_progress；缺省阶段应归一为 preparing
	if raw["status_phase"].(string) != "in_progress" {
		t.Fatalf("阶段未正确存档: %v", raw["status_phase"])
	}
	// 审计动作落库
	var cnt int64
	model.DB.Model(&model.AuditLog{}).
		Where("action = ? AND target_type = ?", "archive.project.create", "archived_project").Count(&cnt)
	if cnt == 0 {
		t.Fatalf("补录未留下审计")
	}
}

func TestArchiveProject_CreateDefaultPhase(t *testing.T) {
	r := setupArchiveRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects", map[string]interface{}{
		"title": "缺省阶段项目",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("补录失败: %d %s", w.Code, w.Body.String())
	}
	raw := decodeBody(t, w)["archived_project"].(map[string]interface{})
	if raw["status_phase"].(string) != "preparing" {
		t.Fatalf("缺省阶段应为 preparing: %v", raw["status_phase"])
	}
	if raw["source"].(string) != "manual" {
		t.Fatalf("缺省来源应为 manual: %v", raw["source"])
	}
}

func TestArchiveProject_PermissionDeniedForClient(t *testing.T) {
	r := setupArchiveRouter()
	// 甲方(1)写存量项目 → 403（命中既有 RequireAdmin 门禁，无需新增越权面）
	w := doJSONFullAuth(t, r, "POST", "/api/v1/admin/archive/projects", map[string]interface{}{
		"title": "不应允许甲方直接建档",
	}, 1, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("甲方应被拒(403)，得到 %d: %s", w.Code, w.Body.String())
	}
}

func TestArchiveProject_MissingTitleBadRequest(t *testing.T) {
	r := setupArchiveRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects", map[string]interface{}{
		"owner_org": "无标题",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺标题应 400: %d %s", w.Code, w.Body.String())
	}
}

func TestArchiveProject_BatchImportAndRollback(t *testing.T) {
	r := setupArchiveRouter()
	// 合法批量
	w := doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects/batch", []map[string]interface{}{
		{"title": "存量A-造价"}, {"title": "存量B-监理", "status_phase": "delivered"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("批量失败: %d %s", w.Code, w.Body.String())
	}
	raw := decodeBody(t, w)
	if int(raw["imported"].(float64)) != 2 {
		t.Fatalf("应导入 2 条: %s", w.Body.String())
	}
	var cnt int64
	model.DB.Model(&model.ArchivedProject{}).Count(&cnt)
	if cnt != 2 {
		t.Fatalf("批量后库存应为 2: %d", cnt)
	}

	// 非法批量（其中一条缺标题）→ 事务整体回滚，库存仍 2
	w = doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects/batch", []map[string]interface{}{
		{"title": "合法新增"},
		{"owner_org": "这条缺标题，应触发回滚"},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法批量应 400: %d %s", w.Code, w.Body.String())
	}
	model.DB.Model(&model.ArchivedProject{}).Count(&cnt)
	if cnt != 2 {
		t.Fatalf("非法批量应整体回滚，库存应仍 2: %d", cnt)
	}
}

func TestArchiveProject_ListFilterUpdateDelete(t *testing.T) {
	r := setupArchiveRouter()
	// 造 3 条
	for _, tt := range []map[string]interface{}{
		{"title": "项目甲", "owner_org": "单位A", "status_phase": "in_progress", "service_type": "cost"},
		{"title": "项目乙", "owner_org": "单位A", "status_phase": "delivered", "service_type": "supervision"},
		{"title": "项目丙", "owner_org": "单位B", "status_phase": "settled", "service_type": "design"},
	} {
		if w := doJSONFull(t, r, "POST", "/api/v1/admin/archive/projects", tt); w.Code != http.StatusOK {
			t.Fatalf("造数失败: %d %s", w.Code, w.Body.String())
		}
	}

	// 列表
	w := doJSONFull(t, r, "GET", "/api/v1/admin/archive/projects?page=1&size=20", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("列表失败: %d", w.Code)
	}
	if int(decodeBody(t, w)["count"].(float64)) != 3 {
		t.Fatalf("列表总数应为 3: %s", w.Body.String())
	}

	// 筛选 phase=delivered
	w = doJSONFull(t, r, "GET", "/api/v1/admin/archive/projects?phase=delivered", nil)
	if int(decodeBody(t, w)["count"].(float64)) != 1 {
		t.Fatalf("按阶段筛选应 1 条")
	}
	// 筛选 关键字 q=单位B
	w = doJSONFull(t, r, "GET", "/api/v1/admin/archive/projects?q=%E5%8D%95%E4%BD%8DB", nil)
	if int(decodeBody(t, w)["count"].(float64)) != 1 {
		t.Fatalf("关键字筛选应 1 条")
	}

	// 取首个 id
	var first model.ArchivedProject
	model.DB.First(&first)
	if first.ID == 0 {
		t.Fatalf("无存量项目")
	}
	idstr := uint2str(first.ID)

	// 详情
	w = doJSONFull(t, r, "GET", "/api/v1/admin/archive/projects/"+idstr, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("详情失败: %d %s", w.Code, w.Body.String())
	}

	// 更新
	w = doJSONFull(t, r, "PUT", "/api/v1/admin/archive/projects/"+idstr, map[string]interface{}{
		"title": first.Title + "-更新", "status_phase": "settled",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("更新失败: %d %s", w.Code, w.Body.String())
	}
	var got model.ArchivedProject
	model.DB.First(&got, first.ID)
	if got.StatusPhase != "settled" {
		t.Fatalf("更新后阶段应为 settled: %v", got.StatusPhase)
	}
	var acnt int64
	model.DB.Model(&model.AuditLog{}).
		Where("action = ? AND target_id = ?", "archive.project.update", first.ID).Count(&acnt)
	if acnt == 0 {
		t.Fatalf("更新未留审计")
	}

	// 删除（软删）
	w = doJSONFull(t, r, "DELETE", "/api/v1/admin/archive/projects/"+idstr, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("删除失败: %d %s", w.Code, w.Body.String())
	}
	var cnt int64
	model.DB.Model(&model.ArchivedProject{}).Where("deleted_at IS NULL").Count(&cnt)
	if cnt != 2 {
		t.Fatalf("软删后可见应 2: %d", cnt)
	}
	// 逻辑删除后详情应 404
	w = doJSONFull(t, r, "GET", "/api/v1/admin/archive/projects/"+idstr, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("已软删详情应 404: %d", w.Code)
	}
}

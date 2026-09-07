package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupOrgRouter 组织层测试路由
func setupOrgRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	r := gin.New()
	api := r.Group("/api/v1")
	client := api.Group("")
	client.Use(AuthTestMiddleware(1, 1))
	{
		client.POST("/org/create", OrgCreate)
		client.POST("/org/join", OrgJoin)
		client.GET("/org/mine", OrgMyList)
		client.GET("/org/:oid/members", OrgMembers)
	}
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.GET("/admin/agent-principals", AdminListAgentPrincipals)
		admin.POST("/admin/agent-principals", AdminRegisterAgentPrincipal)
	}
	return r
}

func TestOrg_CreateMyListMembers(t *testing.T) {
	r := setupOrgRouter()
	alice := createTestUser(t, "13900001001", 1)
	bob := createTestUser(t, "13900001002", 1)

	// alice 建组织（internal，用于 OPC/公司侧组织语义）
	w := doJSONFullAuth(t, r, "POST", "/api/v1/org/create", map[string]interface{}{"name": "CCIT-SIM-InternalOps", "type": "internal"}, int(alice), 1)
	if w.Code != http.StatusOK {
		t.Fatalf("建组织失败 %d %s", w.Code, w.Body.String())
	}
	orgID := uint(decodeBody(t, w)["organization"].(map[string]interface{})["id"].(float64))

	// org/mine 应列到此组织（alice 是 owner）
	w = doJSONFullAuth(t, r, "GET", "/api/v1/org/mine", nil, int(alice), 1)
	raw := decodeBody(t, w)
	if list, ok := raw["organizations"].([]interface{}); !ok || len(list) < 1 {
		t.Fatalf("org/mine 应至少 1 个组织: %s", w.Body.String())
	}

	// bob 普通加入
	w = doJSONFullAuth(t, r, "POST", "/api/v1/org/join", map[string]interface{}{"org_id": int(orgID)}, int(bob), 1)
	if w.Code != http.StatusOK {
		t.Fatalf("join 失败 %d %s", w.Code, w.Body.String())
	}
	// members：alice(owner)+bob(member)=2
	w = doJSONFullAuth(t, r, "GET", "/api/v1/org/"+u64(orgID)+"/members", nil, int(alice), 1)
	b := decodeBody(t, w)
	if mems, ok := b["members"].([]interface{}); !ok || len(mems) != 2 {
		t.Fatalf("应 2 成员: %s", w.Body.String())
	}
	// 非成员访问 → 403（carol 未加入）
	carol := createTestUser(t, "13900001004", 2)
	w = doJSONFullAuth(t, r, "GET", "/api/v1/org/"+u64(orgID)+"/members", nil, int(carol), 2)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非成员应 403，得到 %d", w.Code)
	}
}

func TestOrg_AgentPrincipalAdmin(t *testing.T) {
	r := setupOrgRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/admin/agent-principals", map[string]interface{}{
		"agent_key": "leader-eqs", "display": "EQS 执行", "scope_json": `{"read":["project","archive"]}`,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("登记 agent 失败 %d %s", w.Code, w.Body.String())
	}
	// 重复登记唯一键冲突 → 400
	w = doJSONFull(t, r, "POST", "/api/v1/admin/agent-principals", map[string]interface{}{"agent_key": "leader-eqs"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("重复 agent_key 应 400，得到 %d %s", w.Code, w.Body.String())
	}
	// 列表
	w = doJSONFull(t, r, "GET", "/api/v1/admin/agent-principals", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("agent list 失败 %d", w.Code)
	}
	raw := decodeBody(t, w)
	if list, ok := raw["agent_principals"].([]interface{}); !ok || len(list) < 1 {
		t.Fatalf("应至少 1 个 agent: %s", w.Body.String())
	}
	// 非 admin 不能登记
	w = doJSONFullAuth(t, r, "POST", "/api/v1/admin/agent-principals", map[string]interface{}{"agent_key": "x"}, 1, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非 admin 登记应 403，得到 %d", w.Code)
	}
}

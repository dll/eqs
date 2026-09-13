package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupAuditQueryRouter L1-A 审计检索只读端点测试路由（admin user_type=3）
func setupAuditQueryRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	invalidatePublicCache()

	r := gin.New()
	api := r.Group("/api/v1")
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.GET("/admin/audit/logs", ListAuditLogs)
	}
	return r
}

// seedAudit 直接落审计记录（模拟既有审计留痕，供查询端点读取）
func seedAudit(t *testing.T, action string, userID uint, targetType string, targetID uint) {
	t.Helper()
	if err := model.DB.Create(&model.AuditLog{
		UserID: userID, Action: action, TargetType: targetType, TargetID: targetID,
		Detail: "{\"seed\":true}", IP: "127.0.0.1",
	}).Error; err != nil {
		t.Fatalf("播种审计失败: %v", err)
	}
}

func TestAuditLogQuery_ListAndFilter(t *testing.T) {
	r := setupAuditQueryRouter()
	seedAudit(t, "project.publish", 1, "project", 100)
	seedAudit(t, "project.withdraw", 1, "project", 100)
	seedAudit(t, "archive.project.create", 3, "archived_project", 5)

	// 总数
	w := doJSONFull(t, r, "GET", "/api/v1/admin/audit/logs?page=1&size=20", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("审计查询失败: %d %s", w.Code, w.Body.String())
	}
	raw := decodeBody(t, w)
	if int(raw["count"].(float64)) != 3 {
		t.Fatalf("审计总数应为 3: %s", w.Body.String())
	}

	// 按 action 过滤
	w = doJSONFull(t, r, "GET", "/api/v1/admin/audit/logs?action=project.withdraw", nil)
	if int(decodeBody(t, w)["count"].(float64)) != 1 {
		t.Fatalf("按 action 过滤应 1 条")
	}

	// 按 target_type 过滤
	w = doJSONFull(t, r, "GET", "/api/v1/admin/audit/logs?target_type=archived_project", nil)
	if int(decodeBody(t, w)["count"].(float64)) != 1 {
		t.Fatalf("按 target_type 过滤应 1 条")
	}

	// 按 target_id 过滤
	w = doJSONFull(t, r, "GET", "/api/v1/admin/audit/logs?target_id=100", nil)
	if int(decodeBody(t, w)["count"].(float64)) != 2 {
		t.Fatalf("按 target_id 过滤应 2 条: %s", w.Body.String())
	}

	// 按时间范围（起=现在，应查不到过去播种）
	future := time.Now().Add(time.Hour).Format("2006-01-02 15:04")
	futureEnc := ""
	for _, ch := range future {
		if ch == ' ' {
			futureEnc += "+"
		} else {
			futureEnc += string(ch)
		}
	}
	w = doJSONFull(t, r, "GET", "/api/v1/admin/audit/logs?start="+futureEnc, nil)
	if int(decodeBody(t, w)["count"].(float64)) != 0 {
		t.Fatalf("未来起始时间应 0 条")
	}
}

func TestWriteAudit_PreservesRequestID(t *testing.T) {
	setupAuditQueryRouter()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set("user_id", uint(3))
	c.Set("request_id", "req-audit-123")
	WriteAudit(c, "config.upsert", "config", 7, gin.H{"key": "theme.default"})
	var entry model.AuditLog
	if err := model.DB.Order("id DESC").First(&entry).Error; err != nil {
		t.Fatalf("读取审计记录失败: %v", err)
	}
	if entry.RequestID != "req-audit-123" {
		t.Fatalf("request_id = %q", entry.RequestID)
	}
}
func TestAuditLogQuery_PermissionDeniedForClient(t *testing.T) {
	r := setupAuditQueryRouter()
	seedAudit(t, "project.publish", 1, "project", 101)
	// 甲方(1)查审计 → 403
	w := doJSONFullAuth(t, r, "GET", "/api/v1/admin/audit/logs", nil, 1, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非 admin 应 403，得到 %d: %s", w.Code, w.Body.String())
	}
}

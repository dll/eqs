package adapter

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func testRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := model.InitTestDB()
	now := time.Date(2026, 9, 18, 21, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	h := &handler{service: &service{db: db, now: func() time.Time { return now }}}
	r := gin.New()
	r.Use(gate(h.metadataResponse))
	group := r.Group("/adapter/v1")
	group.GET("/health", h.health)
	group.GET("/projects", h.projects)
	group.GET("/projects/:id/summary", h.projectSummary)
	group.GET("/alerts", h.alerts)
	group.GET("/version", h.version)
	return r
}

func request(t *testing.T, router http.Handler, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestAdapterEndpointsAndMetadata(t *testing.T) {
	r := testRouter(t)
	for _, path := range []string{"/adapter/v1/health", "/adapter/v1/projects", "/adapter/v1/alerts", "/adapter/v1/version"} {
		w := request(t, r, http.MethodGet, path)
		if w.Code != http.StatusOK {
			t.Fatalf("%s 返回 %d: %s", path, w.Code, w.Body.String())
		}
		var body response
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.ProjectID != ProjectID || body.Source != Source || body.SchemaVersion != SchemaVersion || body.ObservedAt.IsZero() {
			t.Fatalf("%s 缺少统一元数据: %+v", path, body)
		}
	}
}

func TestReadOnlyAndCrossProjectGuard(t *testing.T) {
	r := testRouter(t)
	if w := request(t, r, http.MethodPost, "/adapter/v1/health"); w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("写请求应返回 405，实际 %d", w.Code)
	}
	if w := request(t, r, http.MethodGet, "/adapter/v1/projects?project_id=other"); w.Code != http.StatusForbidden {
		t.Fatalf("跨项目请求应返回 403，实际 %d", w.Code)
	}
	if w := request(t, r, http.MethodGet, "/adapter/v1/projects?project_id=eqs"); w.Code != http.StatusOK {
		t.Fatalf("本项目请求应允许，实际 %d", w.Code)
	}
}

func TestUnavailableWhenDatabaseDisconnected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &handler{service: &service{db: nil, now: time.Now}}
	r := gin.New()
	r.Use(gate(h.metadataResponse))
	r.GET("/adapter/v1/health", h.health)
	w := request(t, r, http.MethodGet, "/adapter/v1/health")
	if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), `"status":"unavailable"`) {
		t.Fatalf("断连响应错误: %d %s", w.Code, w.Body.String())
	}
}

func TestVersionCompatibility(t *testing.T) {
	r := testRouter(t)
	w := request(t, r, http.MethodGet, "/adapter/v1/version")
	var body response
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.SchemaVersion != "v1" || body.AdapterVersion != AdapterVersion || body.EQSVersion == "" {
		t.Fatalf("版本响应不兼容: %+v", body)
	}
}

func TestProjectSummaryAndAlerts(t *testing.T) {
	r := testRouter(t)
	now := time.Date(2026, 9, 18, 21, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	deadline := now.Add(-time.Hour)
	project := model.Project{UserID: 1, ServiceType: "geotech", Title: "项目 13812345678", Description: "身份证 110101199001011234", Address: "北京市朝阳区幸福路", BudgetMin: 100000, BudgetMax: 200000, Deadline: &deadline, Status: 3, CreatedAt: now}
	if err := model.DB.Create(&project).Error; err != nil {
		t.Fatal(err)
	}
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 123456, Status: 1}
	if err := model.DB.Create(&order).Error; err != nil {
		t.Fatal(err)
	}
	if err := model.DB.Create(&model.PaymentTransaction{OrderID: order.ID, Status: 0, Amount: 123456}).Error; err != nil {
		t.Fatal(err)
	}
	if err := model.DB.Create(&model.Dispute{OrderID: order.ID, Status: "review", CreatedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	w := request(t, r, http.MethodGet, "/adapter/v1/projects/"+strconv.FormatUint(uint64(project.ID), 10)+"/summary")
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "13812345678") || strings.Contains(w.Body.String(), "123456") {
		t.Fatalf("项目摘要未正确脱敏: %d %s", w.Code, w.Body.String())
	}
	w = request(t, r, http.MethodGet, "/adapter/v1/projects?page=1&page_size=1&status=active")
	if w.Code != http.StatusOK {
		t.Fatalf("项目列表失败: %d %s", w.Code, w.Body.String())
	}
	w = request(t, r, http.MethodGet, "/adapter/v1/alerts?page_size=100&severity=warning")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "delivery_delay") {
		t.Fatalf("告警列表失败: %d %s", w.Code, w.Body.String())
	}
}

func TestInvalidParametersAndNotFound(t *testing.T) {
	r := testRouter(t)
	for _, path := range []string{"/adapter/v1/projects?page=0", "/adapter/v1/projects?page_size=101", "/adapter/v1/projects?status=bad", "/adapter/v1/alerts?severity=bad"} {
		if w := request(t, r, http.MethodGet, path); w.Code != http.StatusBadRequest {
			t.Fatalf("%s 应返回 400，实际 %d", path, w.Code)
		}
	}
	if w := request(t, r, http.MethodGet, "/adapter/v1/projects/not-a-project/summary"); w.Code != http.StatusBadRequest {
		t.Fatalf("非法项目 ID 应返回 400，实际 %d", w.Code)
	}
	if w := request(t, r, http.MethodGet, "/adapter/v1/projects/999999/summary"); w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应返回 404，实际 %d", w.Code)
	}
}

func TestStatusMappingsAndRegisteredRoutes(t *testing.T) {
	for input, expected := range map[int]string{0: "active", 4: "completed", 5: "cancelled", 99: "active"} {
		if got := projectStatus(input); got != expected {
			t.Fatalf("projectStatus(%d)=%s", input, got)
		}
	}
	for input, expected := range map[int]string{0: "pending_contract", 1: "in_progress", 2: "pending_acceptance", 3: "completed", 4: "disputed", 5: "offline", 6: "cancelled", 99: "unknown"} {
		if got := orderStatus(input); got != expected {
			t.Fatalf("orderStatus(%d)=%s", input, got)
		}
	}
	for _, status := range []string{"active", "completed", "cancelled"} {
		if len(projectStatusCodes(status)) == 0 {
			t.Fatalf("状态 %s 未映射", status)
		}
	}
	if projectStatusCodes("other") != nil {
		t.Fatal("未知状态不应映射")
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r, model.InitTestDB())
	if w := request(t, r, http.MethodHead, "/adapter/v1/version"); w.Code != http.StatusOK {
		t.Fatalf("注册的 HEAD 路由失败: %d", w.Code)
	}
	req := httptest.NewRequest(http.MethodGet, "/adapter/v1/health", nil)
	req.Header.Set("X-Project-ID", "other")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("请求头跨项目访问应返回 403，实际 %d", w.Code)
	}
}

func TestQualificationAlertAndStoredVersion(t *testing.T) {
	r := testRouter(t)
	now := time.Date(2026, 9, 18, 21, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	validTo := now.Add(10 * 24 * time.Hour)
	if err := model.DB.Create(&model.SupplierQualification{SupplierID: 1, ValidTo: &validTo}).Error; err != nil {
		t.Fatal(err)
	}
	if err := model.DB.Create(&model.SystemVersion{Version: "12.0.0", ReleasedAt: now}).Error; err != nil {
		t.Fatal(err)
	}
	if w := request(t, r, http.MethodGet, "/adapter/v1/alerts?severity=info"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "qualification_expire") {
		t.Fatalf("资质告警失败: %d %s", w.Code, w.Body.String())
	}
	if w := request(t, r, http.MethodGet, "/adapter/v1/version"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "12.0.0") {
		t.Fatalf("数据库版本读取失败: %d %s", w.Code, w.Body.String())
	}
}

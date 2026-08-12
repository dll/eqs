package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func setupDemoRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")

	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.POST("/admin/demo/seed", DemoSeedHandler)
		admin.POST("/admin/demo/clean", DemoCleanHandler)
		admin.POST("/admin/demo/toggle", DemoToggleHandler)
		admin.GET("/admin/demo/status", DemoStatusHandler)
	}
	return r
}

func TestDemo_Seed(t *testing.T) {
	r := setupDemoRouter()

	for _, mode := range []string{"demo", "test", "training"} {
		w := doJSONFull(t, r, "POST", "/api/v1/admin/demo/seed?mode="+mode, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("seed %s 失败: %d %s", mode, w.Code, w.Body.String())
		}
		out := decodeBody(t, w)
		if out["mode"] != mode {
			t.Fatalf("mode 应为 %s: %v", mode, out["mode"])
		}
	}

	// 非法 mode
	w := doJSONFull(t, r, "POST", "/api/v1/admin/demo/seed?mode=invalid", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法 mode 应400，得到 %d", w.Code)
	}
}

func TestDemo_Clean(t *testing.T) {
	r := setupDemoRouter()
	doJSONFull(t, r, "POST", "/api/v1/admin/demo/seed?mode=demo", nil)

	w := doJSONFull(t, r, "POST", "/api/v1/admin/demo/clean", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("clean 失败: %d %s", w.Code, w.Body.String())
	}

	var users, projects, admins int64
	model.DB.Model(&model.User{}).Count(&users)
	model.DB.Model(&model.User{}).Where("user_type = ?", 3).Count(&admins)
	model.DB.Model(&model.Project{}).Count(&projects)
	if projects != 0 {
		t.Fatalf("清理后应无项目: projects=%d", projects)
	}
	// 保留管理员账号（user_type=3），避免管理后台失效
	if admins == 0 {
		t.Fatalf("清理后应保留管理员账号")
	}
	nonAdmin := users - admins
	if nonAdmin != 0 {
		t.Fatalf("清理后非管理员用户应为0: users=%d admins=%d", users, admins)
	}
}

func TestDemo_Toggle(t *testing.T) {
	r := setupDemoRouter()

	w := doJSONFull(t, r, "POST", "/api/v1/admin/demo/toggle", map[string]interface{}{"enable": true})
	if w.Code != http.StatusOK {
		t.Fatalf("toggle 开启失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["demo_mode"] != true {
		t.Fatalf("demo_mode 应为 true: %v", out)
	}

	var cfg model.SystemConfig
	if err := model.DB.Where("config_key = ?", "demo.enabled").First(&cfg).Error; err != nil {
		t.Fatalf("开关未持久化: %v", err)
	}
	if cfg.ConfigValue != "true" {
		t.Fatalf("配置值应为 true: %s", cfg.ConfigValue)
	}

	// 参数缺失
	w = doJSONFull(t, r, "POST", "/api/v1/admin/demo/toggle", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺参应400，得到 %d", w.Code)
	}

	// 更新既有配置
	w = doJSONFull(t, r, "POST", "/api/v1/admin/demo/toggle", map[string]interface{}{"enable": false})
	if w.Code != http.StatusOK {
		t.Fatalf("toggle 关闭失败: %d %s", w.Code, w.Body.String())
	}
	out = decodeBody(t, w)
	if out["demo_mode"] != false {
		t.Fatalf("demo_mode 应为 false: %v", out)
	}
}

func TestDemo_Status(t *testing.T) {
	r := setupDemoRouter()

	w := doJSONFull(t, r, "GET", "/api/v1/admin/demo/status", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status 失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["demo_mode"] != false {
		t.Fatalf("默认应关闭: %v", out)
	}

	// 开启后再次查询
	doJSONFull(t, r, "POST", "/api/v1/admin/demo/toggle", map[string]interface{}{"enable": true})
	w = doJSONFull(t, r, "GET", "/api/v1/admin/demo/status", nil)
	out = decodeBody(t, w)
	if out["demo_mode"] != true {
		t.Fatalf("开启后应返回 true: %v", out)
	}
}

func TestDemo_AdminPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(1, 1))
	admin.Use(RequireAdmin())
	{
		admin.POST("/admin/demo/seed", DemoSeedHandler)
	}

	w := doJSONFull(t, r, "POST", "/api/v1/admin/demo/seed?mode=demo", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("普通用户访问 demo 应403，得到 %d", w.Code)
	}
}

// TestMonitor_Middlewares 覆盖监控中间件、P95 计算、统计接口与版本限流
func TestMonitor_Middlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	// 单独测试 computeP95 纯函数
	if p := computeP95(nil); p != 0 {
		t.Fatalf("空切片 P95 应为0: %v", p)
	}
	if p := computeP95([]time.Duration{1, 2, 3, 4}); p < 0 {
		t.Fatalf("P95 计算异常: %v", p)
	}

	r := gin.New()
	r.Use(MonitorMiddleware())
	api := r.Group("/api/v1")
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.GET("/admin/monitor/stats", MonitorStats)
		admin.GET("/admin/version/check", VersionRateLimit(), VersionCheck)
	}

	// 多次请求触发监控统计（含 404 错误路径）
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/never", nil)
		r.ServeHTTP(w, req)
	}

	// 查询监控统计
	var statsW *httptest.ResponseRecorder
	statsW = doJSONFull(t, r, "GET", "/api/v1/admin/monitor/stats", nil)
	if statsW.Code != http.StatusOK {
		t.Fatalf("monitor stats 失败: %d %s", statsW.Code, statsW.Body.String())
	}
	out := decodeBody(t, statsW)
	if _, ok := out["stats"]; !ok {
		t.Fatalf("缺少 stats 字段: %v", out)
	}
}

func TestMonitor_VersionRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")
	api.GET("/check", VersionRateLimit(), VersionCheck)

	// 首次请求放行
	w := doJSONFull(t, r, "GET", "/api/v1/check?current=1.0.0", nil)
	if w.Code == http.StatusTooManyRequests {
		t.Fatal("首次请求不应被限流")
	}
}
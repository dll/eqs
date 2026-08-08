package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupConfigRouter 配置中心测试路由
func setupConfigRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")

	auth := api.Group("")
	auth.Use(AuthTestMiddleware(1, 1))
	{
		auth.GET("/config/public", PublicConfigs)
		auth.GET("/config/user/prefs", GetUserPrefs)
		auth.PUT("/config/user/prefs", UpdateUserPrefs)
		auth.GET("/theme/list", ThemeList)
		auth.GET("/i18n/:lang", I18nMessages)
		auth.GET("/version/check", VersionCheck)
		auth.GET("/version/latest", VersionLatest)
	}

	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.GET("/admin/config/list", AdminListConfigs)
		admin.POST("/admin/config/upsert", AdminUpsertConfig)
		admin.DELETE("/admin/config/delete/:key", AdminDeleteConfig)
		admin.POST("/admin/version/publish", AdminPublishVersion)
		admin.GET("/admin/version/list", AdminListVersions)
	}
	return r
}

func TestConfigCenter_AdminCRUD(t *testing.T) {
	r := setupConfigRouter()

	w := doJSONFull(t, r, "POST", "/api/v1/admin/config/upsert", map[string]interface{}{
		"config_key": "theme.default", "config_value": "print",
		"value_type": "string", "description": "默认主题", "is_public": true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("新增配置失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "POST", "/api/v1/admin/config/upsert", map[string]interface{}{
		"config_key": "theme.default", "config_value": "light",
		"value_type": "string", "is_public": true,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("更新配置失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "GET", "/api/v1/admin/config/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("配置列表失败: %d", w.Code)
	}
	out := decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("应有1条配置，得到 %v", out["count"])
	}

	w = doJSONFull(t, r, "GET", "/api/v1/config/public", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("公开配置失败: %d", w.Code)
	}
	pub := decodeBody(t, w)
	configs := pub["configs"].(map[string]interface{})
	if configs["theme.default"] != "light" {
		t.Fatalf("theme.default 应为 light，得到 %v", configs["theme.default"])
	}

	w = doJSONFull(t, r, "DELETE", "/api/v1/admin/config/delete/theme.default", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("删除配置失败: %d", w.Code)
	}
}

func TestConfigCenter_UserPrefs(t *testing.T) {
	r := setupConfigRouter()

	w := doJSONFull(t, r, "GET", "/api/v1/config/user/prefs", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("读取偏好失败: %d", w.Code)
	}
	out := decodeBody(t, w)
	if out["theme"] != "print" || out["lang"] != "zh-CN" {
		t.Fatalf("默认偏好错误: %v", out)
	}

	w = doJSONFull(t, r, "PUT", "/api/v1/config/user/prefs", map[string]interface{}{
		"theme": "dark", "lang": "en-US",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("更新偏好失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "GET", "/api/v1/config/user/prefs", nil)
	out = decodeBody(t, w)
	if out["theme"] != "dark" || out["lang"] != "en-US" {
		t.Fatalf("偏好未保存: %v", out)
	}

	w = doJSONFull(t, r, "PUT", "/api/v1/config/user/prefs", map[string]interface{}{"theme": "invalid"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法主题应400，得到 %d", w.Code)
	}
}

func TestConfigCenter_ThemeAndI18n(t *testing.T) {
	r := setupConfigRouter()

	w := doJSONFull(t, r, "GET", "/api/v1/theme/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("主题列表失败: %d", w.Code)
	}
	out := decodeBody(t, w)
	if len(out["themes"].([]interface{})) != 3 {
		t.Fatalf("应有3个主题")
	}

	w = doJSONFull(t, r, "GET", "/api/v1/i18n/zh-CN", nil)
	out = decodeBody(t, w)
	if out["lang"] != "zh-CN" {
		t.Fatalf("语言错误: %v", out["lang"])
	}
	msgs := out["messages"].(map[string]interface{})
	if msgs["nav.home"] != "首页" {
		t.Fatalf("中文文案错误: %v", msgs["nav.home"])
	}

	w = doJSONFull(t, r, "GET", "/api/v1/i18n/en-US", nil)
	out = decodeBody(t, w)
	msgs = out["messages"].(map[string]interface{})
	if msgs["nav.home"] != "Home" {
		t.Fatalf("英文文案错误: %v", msgs["nav.home"])
	}
}

func TestConfigCenter_Version(t *testing.T) {
	r := setupConfigRouter()

	w := doJSONFull(t, r, "GET", "/api/v1/version/check?current=1.0.0", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("版本检查失败: %d", w.Code)
	}
	out := decodeBody(t, w)
	if out["update_available"] != false {
		t.Fatalf("无版本时应不提示更新: %v", out["update_available"])
	}

	w = doJSONFull(t, r, "POST", "/api/v1/admin/version/publish", map[string]interface{}{
		"version": "1.1.0", "build": 11, "platform": "all",
		"update_url": "https://example.com/update", "release_notes": "新增主题与国际化",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("发布版本失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "GET", "/api/v1/version/check?current=1.0.0", nil)
	out = decodeBody(t, w)
	if out["update_available"] != true {
		t.Fatalf("旧版本应提示更新: %v", out)
	}
	if out["version"] != "1.1.0" {
		t.Fatalf("版本号错误: %v", out["version"])
	}

	w = doJSONFull(t, r, "GET", "/api/v1/version/check?current=1.1.0", nil)
	out = decodeBody(t, w)
	if out["update_available"] != false {
		t.Fatalf("同版本不应提示更新: %v", out["update_available"])
	}

	w = doJSONFull(t, r, "GET", "/api/v1/admin/version/list", nil)
	out = decodeBody(t, w)
	if out["count"].(float64) != 1 {
		t.Fatalf("应有1条版本记录")
	}
}

func TestConfigCenter_AdminPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	// 普通用户（user_type=1）访问 admin 接口应被 RequireAdmin 拦截
	r := gin.New()
	api := r.Group("/api/v1")
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(1, 1))
	admin.Use(RequireAdmin())
	{
		admin.GET("/admin/config/list", AdminListConfigs)
	}

	w := doJSONFull(t, r, "GET", "/api/v1/admin/config/list", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("普通用户访问 admin 应403，得到 %d", w.Code)
	}
}

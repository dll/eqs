package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func setupConfigFullRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")

	shared := api.Group("")
	shared.Use(AuthTestMiddleware(1, 1))
	{
		shared.GET("/platform/links", PlatformLinks)
		shared.GET("/version/latest", VersionLatest)
		shared.PUT("/project/:id/theme", SetProjectTheme)
	}

	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.POST("/admin/version/publish", AdminPublishVersion)
	}
	return r
}

func TestPlatformLinks(t *testing.T) {
	r := setupConfigFullRouter()

	w := doJSONFull(t, r, "GET", "/api/v1/platform/links", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("platform/links 失败: %d", w.Code)
	}
	out := decodeBody(t, w)
	platforms := out["platforms"].([]interface{})
	if len(platforms) != 4 {
		t.Fatalf("应有 4 个平台入口: %v", platforms)
	}
}

func TestVersionLatest(t *testing.T) {
	r := setupConfigFullRouter()

	// 无版本信息时 404
	w := doJSONFull(t, r, "GET", "/api/v1/version/latest", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("无版本应404，得到 %d", w.Code)
	}

	// 发布后返回最新
	doJSONFull(t, r, "POST", "/api/v1/admin/version/publish", map[string]interface{}{
		"version": "2.0.0", "build": 20, "platform": "all",
	})
	w = doJSONFull(t, r, "GET", "/api/v1/version/latest", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("有版本应200，得到 %d", w.Code)
	}
	out := decodeBody(t, w)
	ver := out["version"].(map[string]interface{})
	if ver["version"] != "2.0.0" {
		t.Fatalf("最新版本号错误: %v", ver["version"])
	}
}

func TestAdminPublishVersion(t *testing.T) {
	r := setupConfigFullRouter()

	// 缺少必填字段
	w := doJSONFull(t, r, "POST", "/api/v1/admin/version/publish", map[string]interface{}{"build": 1})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺 version 应400，得到 %d", w.Code)
	}

	// platform 缺省默认 all
	w = doJSONFull(t, r, "POST", "/api/v1/admin/version/publish", map[string]interface{}{"version": "3.0.1"})
	if w.Code != http.StatusOK {
		t.Fatalf("发布失败: %d %s", w.Code, w.Body.String())
	}
}

func TestSetProjectTheme(t *testing.T) {
	r := setupConfigFullRouter()

	// 直接落库创建甲方(user_id=1)项目
	pID := createDirectProject(t, 1)

	// 合法主题
	w := doJSONFull(t, r, "PUT", "/api/v1/project/"+u64(pID)+"/theme", map[string]interface{}{"theme": "dark"})
	if w.Code != http.StatusOK {
		t.Fatalf("设置项目主题失败: %d %s", w.Code, w.Body.String())
	}

	// 非法主题
	w = doJSONFull(t, r, "PUT", "/api/v1/project/"+u64(pID)+"/theme", map[string]interface{}{"theme": "invalid"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法主题应400，得到 %d", w.Code)
	}

	// 缺参
	w = doJSONFull(t, r, "PUT", "/api/v1/project/"+u64(pID)+"/theme", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺 theme 应400，得到 %d", w.Code)
	}

	// 非法项目 ID
	w = doJSONFull(t, r, "PUT", "/api/v1/project/abc/theme", map[string]interface{}{"theme": "dark"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法项目ID应400，得到 %d", w.Code)
	}

	// 不存在的项目
	w = doJSONFull(t, r, "PUT", "/api/v1/project/99999/theme", map[string]interface{}{"theme": "dark"})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应404，得到 %d", w.Code)
	}
}

func TestSetProjectTheme_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")
	// 用 user_id=2（非项目所有者）
	api.PUT("/project/:id/theme", AuthTestMiddleware(2, 2), SetProjectTheme)

	// 由 user_id=1 创建项目（直接落库）
	pID := createDirectProject(t, 1)

	w := doJSONFull(t, r, "PUT", "/api/v1/project/"+u64(pID)+"/theme", map[string]interface{}{"theme": "dark"})
	if w.Code != http.StatusForbidden {
		t.Fatalf("非所有者应403，得到 %d", w.Code)
	}
}

// createDirectProject 直接落库创建项目并返回 ID
func createDirectProject(t *testing.T, ownerID uint) uint {
	t.Helper()
	p := model.Project{UserID: ownerID, ProjectType: "cost", ServiceType: "cost", Title: "主题测试项目",
		BudgetMin: 1000, BudgetMax: 5000, Status: 1, PublishScope: "public"}
	if err := model.DB.Create(&p).Error; err != nil {
		t.Fatalf("创建项目失败: %v", err)
	}
	return p.ID
}

func TestParseConfigValue(t *testing.T) {
	if v := parseConfigValue("42", "int"); v != 42 {
		t.Fatalf("int 解析错误: %v", v)
	}
	if v := parseConfigValue("abc", "int"); v != 0 {
		t.Fatalf("非法 int 应为0: %v", v)
	}
	if v := parseConfigValue("true", "bool"); v != true {
		t.Fatalf("bool 解析错误: %v", v)
	}
	if v := parseConfigValue("0", "bool"); v != false {
		t.Fatalf("bool 0 应为 false: %v", v)
	}
	if v := parseConfigValue(`{"a":1}`, "json"); v == nil {
		t.Fatal("json 解析失败")
	}
	if v := parseConfigValue("bad-json", "json"); v != nil {
		t.Fatalf("非法 json 应为 nil: %v", v)
	}
	if v := parseConfigValue("hello", "string"); v != "hello" {
		t.Fatalf("string 解析错误: %v", v)
	}
}

func TestGetCommissionRate_Branches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	// 清空缓存后写入不同类型配置
	invalidatePublicCache()
	model.DB.Where("1=1").Delete(&model.SystemConfig{})

	// int 类型
	model.DB.Create(&model.SystemConfig{ConfigKey: "commission.rate", ConfigValue: "5", ValueType: "int", IsPublic: true})
	invalidatePublicCache()
	if r := getCommissionRate(); r != 5 {
		t.Fatalf("int 费率应为5: %v", r)
	}

	// 越界 int
	model.DB.Model(&model.SystemConfig{}).Where("config_key = ?", "commission.rate").Update("config_value", "150")
	invalidatePublicCache()
	if r := getCommissionRate(); r != 0 {
		t.Fatalf("越界费率应为0: %v", r)
	}

	// string 类型
	model.DB.Model(&model.SystemConfig{}).Where("config_key = ?", "commission.rate").Updates(map[string]interface{}{"config_value": "8.5", "value_type": "string"})
	invalidatePublicCache()
	if r := getCommissionRate(); r != 8.5 {
		t.Fatalf("string 费率应为 8.5: %v", r)
	}

	// 越界 string
	model.DB.Model(&model.SystemConfig{}).Where("config_key = ?", "commission.rate").Update("config_value", "-3")
	invalidatePublicCache()
	if r := getCommissionRate(); r != 0 {
		t.Fatalf("负费率应为0: %v", r)
	}

	// 非法 string
	model.DB.Model(&model.SystemConfig{}).Where("config_key = ?", "commission.rate").Update("config_value", "abc")
	invalidatePublicCache()
	if r := getCommissionRate(); r != 0 {
		t.Fatalf("非法费率应为0: %v", r)
	}
}

func TestAdminCollectCommission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(3, 3))
	admin.Use(RequireAdmin())
	{
		admin.POST("/admin/commission/:id/collect", AdminCollectCommission)
		admin.GET("/admin/commission/list", AdminListCommissions)
	}

	// 非法 ID
	w := doJSONFull(t, r, "POST", "/api/v1/admin/commission/abc/collect", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d", w.Code)
	}

	// 不存在
	w = doJSONFull(t, r, "POST", "/api/v1/admin/commission/99999/collect", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在应404，得到 %d", w.Code)
	}

	// 创建佣金单并收取
	rec := model.CommissionRecord{OrderID: 1, SupplierID: 2, ProjectID: 1, Amount: 10000, Rate: 5, Commission: 500, Status: "pending"}
	if err := model.DB.Create(&rec).Error; err != nil {
		t.Fatalf("创建佣金单失败: %v", err)
	}
	w = doJSONFull(t, r, "POST", "/api/v1/admin/commission/"+u64(rec.ID)+"/collect", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("收取失败: %d %s", w.Code, w.Body.String())
	}

	// 幂等重复收取
	w = doJSONFull(t, r, "POST", "/api/v1/admin/commission/"+u64(rec.ID)+"/collect", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("重复收取失败: %d", w.Code)
	}
	out := decodeBody(t, w)
	if out["idempotent"] != true {
		t.Fatalf("重复收取应幂等: %v", out)
	}

	// 列表统计
	w = doJSONFull(t, r, "GET", "/api/v1/admin/commission/list?order_id=1&status=pending", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("列表失败: %d", w.Code)
	}
}
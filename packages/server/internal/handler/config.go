package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ==================== 配置缓存 ====================

type configCache struct {
	mu      sync.RWMutex
	configs map[string]interface{}
	updated time.Time
}

var publicCache = &configCache{configs: make(map[string]interface{})}

func loadPublicCache() {
	publicCache.mu.Lock()
	defer publicCache.mu.Unlock()
	var configs []model.SystemConfig
	model.DB.Where("is_public = ?", true).Find(&configs)
	m := make(map[string]interface{}, len(configs))
	for _, cfg := range configs {
		m[cfg.ConfigKey] = parseConfigValue(cfg.ConfigValue, cfg.ValueType)
	}
	publicCache.configs = m
	publicCache.updated = time.Now()
}

func invalidatePublicCache() {
	publicCache.mu.Lock()
	publicCache.configs = make(map[string]interface{})
	publicCache.mu.Unlock()
}

func getPublicCached() map[string]interface{} {
	publicCache.mu.RLock()
	if len(publicCache.configs) > 0 {
		defer publicCache.mu.RUnlock()
		return publicCache.configs
	}
	publicCache.mu.RUnlock()
	loadPublicCache()
	return publicCache.configs
}

// ==================== 系统配置中心 ====================

// AdminListConfigs 全部配置列表（管理员）
func AdminListConfigs(c *gin.Context) {
	var configs []model.SystemConfig
	model.DB.Order("config_key ASC").Find(&configs)
	ok(c, gin.H{"configs": configs, "count": len(configs)})
}

type UpsertConfigRequest struct {
	ConfigKey   string `json:"config_key" binding:"required"`
	ConfigValue string `json:"config_value"`
	ValueType   string `json:"value_type"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// AdminUpsertConfig 新增/更新配置（管理员）
func AdminUpsertConfig(c *gin.Context) {
	var req UpsertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	userID := c.GetUint("user_id")

	if req.ValueType == "" {
		req.ValueType = "string"
	}

	var cfg model.SystemConfig
	err := model.DB.Where("config_key = ?", req.ConfigKey).First(&cfg).Error
	if err != nil {
		cfg = model.SystemConfig{
			ConfigKey:   req.ConfigKey,
			ConfigValue: req.ConfigValue,
			ValueType:   req.ValueType,
			Description: req.Description,
			IsPublic:    req.IsPublic,
			UpdatedBy:   userID,
		}
		if err := model.DB.Create(&cfg).Error; err != nil {
			serverError(c, err)
			return
		}
	} else {
		updates := map[string]interface{}{
			"config_value": req.ConfigValue,
			"value_type":   req.ValueType,
			"description":  req.Description,
			"is_public":    req.IsPublic,
			"updated_by":   userID,
		}
		if err := model.DB.Model(&cfg).Updates(updates).Error; err != nil {
			serverError(c, err)
			return
		}
	}

	if req.IsPublic {
		invalidatePublicCache()
	}
	WriteAudit(c, "config.upsert", "config", cfg.ID, gin.H{"key": req.ConfigKey})
	ok(c, gin.H{"config": cfg, "message": "配置已保存"})
}

// AdminDeleteConfig 删除配置（管理员）
func AdminDeleteConfig(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		badRequest(c, "配置键无效")
		return
	}
	model.DB.Where("config_key = ?", key).Delete(&model.SystemConfig{})
	invalidatePublicCache()
	WriteAudit(c, "config.delete", "config", 0, gin.H{"key": key})
	ok(c, gin.H{"message": "配置已删除"})
}

// PublicConfigs 公开配置（所有用户可读）— 走缓存
func PublicConfigs(c *gin.Context) {
	result := getPublicCached()
	ok(c, gin.H{"configs": result})
}

// parseConfigValue 按类型解析配置值
func parseConfigValue(value, valueType string) interface{} {
	switch valueType {
	case "int":
		if v, err := strconv.Atoi(value); err == nil {
			return v
		}
		return 0
	case "bool":
		return value == "true" || value == "1"
	case "json":
		var v interface{}
		if err := json.Unmarshal([]byte(value), &v); err == nil {
			return v
		}
		return nil
	default:
		return value
	}
}

// ==================== 用户偏好 ====================

// GetUserPrefs 用户偏好（主题/语言）
func GetUserPrefs(c *gin.Context) {
	userID := c.GetUint("user_id")
	var setting model.UserSetting
	err := model.DB.Where("user_id = ?", userID).First(&setting).Error
	if err != nil {
		ok(c, gin.H{"theme": "print", "lang": "zh-CN"})
		return
	}
	ok(c, gin.H{"theme": setting.Theme, "lang": setting.Lang})
}

type UpdatePrefsRequest struct {
	Theme string `json:"theme"`
	Lang  string `json:"lang"`
}

// UpdateUserPrefs 更新用户偏好（主题/语言）
func UpdateUserPrefs(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req UpdatePrefsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	if req.Theme != "" && req.Theme != "print" && req.Theme != "dark" && req.Theme != "light" {
		badRequest(c, "主题不合法")
		return
	}
	if req.Lang != "" && req.Lang != "zh-CN" && req.Lang != "en-US" {
		badRequest(c, "语言不合法")
		return
	}

	var setting model.UserSetting
	err := model.DB.Where("user_id = ?", userID).First(&setting).Error
	now := time.Now()
	if err != nil {
		setting = model.UserSetting{
			UserID: userID, Theme: "print", Lang: "zh-CN",
			CreatedAt: now, UpdatedAt: now,
		}
		if req.Theme != "" {
			setting.Theme = req.Theme
		}
		if req.Lang != "" {
			setting.Lang = req.Lang
		}
		model.DB.Create(&setting)
	} else {
		updates := map[string]interface{}{"updated_at": now}
		if req.Theme != "" {
			updates["theme"] = req.Theme
			setting.Theme = req.Theme
		}
		if req.Lang != "" {
			updates["lang"] = req.Lang
			setting.Lang = req.Lang
		}
		model.DB.Model(&setting).Updates(updates)
	}

	WriteAudit(c, "config.prefs", "user", userID, gin.H{"theme": req.Theme, "lang": req.Lang})
	ok(c, gin.H{"theme": setting.Theme, "lang": setting.Lang, "message": "偏好已保存"})
}

// ==================== 主题 ====================

// ThemeList 可用主题列表
func ThemeList(c *gin.Context) {
	ok(c, gin.H{"themes": []gin.H{
		{"id": "print", "name": "打印主题", "description": "白底黑字，适合截图打印"},
		{"id": "dark", "name": "深色主题", "description": "深色背景，夜间友好"},
		{"id": "light", "name": "浅色主题", "description": "标准浅色界面"},
	}})
}

// SetProjectTheme 设置项目主题
func SetProjectTheme(c *gin.Context) {
	projectID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var req struct {
		Theme string `json:"theme" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if req.Theme != "print" && req.Theme != "dark" && req.Theme != "light" && req.Theme != "" {
		badRequest(c, "主题不合法")
		return
	}

	var project model.Project
	if err := model.DB.First(&project, projectID).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	if project.UserID != userID {
		forbidden(c, "仅项目所有者可设置主题")
		return
	}

	model.DB.Model(&project).Update("theme", req.Theme)
	WriteAudit(c, "project.theme", "project", projectID, gin.H{"theme": req.Theme})
	ok(c, gin.H{"message": "项目主题已设置", "theme": req.Theme})
}

// ==================== 国际化 ====================

// I18nMessages 获取语言文案
func I18nMessages(c *gin.Context) {
	lang := c.Param("lang")
	if lang != "zh-CN" && lang != "en-US" {
		lang = "zh-CN"
	}
	messages := loadI18n(lang)
	ok(c, gin.H{"lang": lang, "messages": messages})
}

// loadI18n 加载语言文案（内置最小集，前端可自行扩展）
func loadI18n(lang string) map[string]string {
	base := map[string]string{
		"app.name":        "工程快捷服务",
		"common.login":    "登录",
		"common.logout":   "退出登录",
		"common.confirm":  "确定",
		"common.cancel":   "取消",
		"common.save":     "保存",
		"common.delete":   "删除",
		"common.search":   "搜索",
		"common.status":   "状态",
		"common.action":   "操作",
		"common.loading":  "加载中...",
		"common.success":  "操作成功",
		"common.failed":   "操作失败",
		"common.empty":    "暂无数据",
		"nav.home":        "首页",
		"nav.project":     "项目",
		"nav.order":       "订单",
		"nav.mine":        "我的",
		"nav.dashboard":   "数据看板",
		"login.phone":     "请输入手机号",
		"login.code":      "验证码",
		"login.getCode":   "获取验证码",
		"login.type":      "角色",
		"project.create":  "发布项目",
		"project.list":    "项目列表",
		"project.detail":  "项目详情",
		"order.list":      "我的订单",
		"order.detail":    "订单详情",
		"theme.print":     "打印主题",
		"theme.dark":      "深色主题",
		"theme.light":     "浅色主题",
		"lang.zh-CN":      "中文",
		"lang.en-US":      "English",
		"version.update":  "发现新版本",
		"version.forced":  "请更新到最新版本",
		"error.network":   "网络异常",
		"error.unauth":    "未登录",
		"error.notFound":  "资源不存在",
		"error.server":    "服务器内部错误",
	}
	if lang == "en-US" {
		return map[string]string{
			"app.name":       "EQS Engineering Quick Service",
			"common.login":   "Login",
			"common.logout":  "Logout",
			"common.confirm": "Confirm",
			"common.cancel":  "Cancel",
			"common.save":    "Save",
			"common.delete":  "Delete",
			"common.search":  "Search",
			"common.status":  "Status",
			"common.action":  "Action",
			"common.loading": "Loading...",
			"common.success": "Success",
			"common.failed":  "Failed",
			"common.empty":   "No data",
			"nav.home":       "Home",
			"nav.project":    "Projects",
			"nav.order":      "Orders",
			"nav.mine":       "Mine",
			"nav.dashboard":  "Dashboard",
			"login.phone":    "Phone number",
			"login.code":     "Verification code",
			"login.getCode":  "Get code",
			"login.type":     "Role",
			"project.create": "Publish Project",
			"project.list":   "Project List",
			"project.detail": "Project Detail",
			"order.list":     "My Orders",
			"order.detail":   "Order Detail",
			"theme.print":    "Print",
			"theme.dark":     "Dark",
			"theme.light":    "Light",
			"lang.zh-CN":     "中文",
			"lang.en-US":     "English",
			"version.update": "New version available",
			"version.forced": "Please update to the latest version",
			"error.network":  "Network error",
			"error.unauth":   "Unauthorized",
			"error.notFound": "Not found",
			"error.server":   "Internal server error",
		}
	}
	return base
}

// ==================== 多端 ====================

// PlatformLinks 各端访问地址（从配置中心读取）
func PlatformLinks(c *gin.Context) {
	cfgs := getPublicCached()
	urls, _ := cfgs["multiplatform.urls"].(map[string]interface{})
	if urls == nil {
		urls = make(map[string]interface{})
	}

	getURL := func(key string) string {
		if v, ok := urls[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	ok(c, gin.H{"platforms": []gin.H{
		{"id": "h5", "name": "H5", "url": getURL("h5")},
		{"id": "mp-weixin", "name": "微信小程序", "url": getURL("mp-weixin")},
		{"id": "app-ios", "name": "iOS App", "url": getURL("app-ios")},
		{"id": "app-android", "name": "Android App", "url": getURL("app-android")},
	}})
}

// ==================== 版本 ====================

// VersionCheck 版本检查
func VersionCheck(c *gin.Context) {
	current := c.Query("current")
	platform := c.DefaultQuery("platform", "h5")

	var latest model.SystemVersion
	err := model.DB.Where("platform IN (?)", []string{platform, "all"}).
		Order("build DESC").First(&latest).Error
	if err != nil {
		ok(c, gin.H{"update_available": false, "message": "当前已是最新版本"})
		return
	}

	needUpdate := false
	if current != "" {
		needUpdate = compareVersions(current, latest.Version) < 0
	}

	ok(c, gin.H{
		"update_available": needUpdate,
		"version":          latest.Version,
		"build":            latest.Build,
		"update_url":       latest.UpdateURL,
		"release_notes":    latest.ReleaseNotes,
		"mandatory":        latest.Mandatory,
		"message":          fmt.Sprintf("最新版本 %s", latest.Version),
	})
}

// VersionLatest 最新版本信息
func VersionLatest(c *gin.Context) {
	platform := c.DefaultQuery("platform", "all")
	var latest model.SystemVersion
	err := model.DB.Where("platform IN (?)", []string{platform, "all"}).
		Order("build DESC").First(&latest).Error
	if err != nil {
		notFound(c, "暂无版本信息")
		return
	}
	ok(c, gin.H{"version": latest})
}

// compareVersions 比较版本号（语义化）
func compareVersions(a, b string) int {
	parse := func(v string) []int {
		var parts []int
		for _, p := range splitVersion(v) {
			n, _ := strconv.Atoi(p)
			parts = append(parts, n)
		}
		for len(parts) < 3 {
			parts = append(parts, 0)
		}
		return parts
	}
	pa, pb := parse(a), parse(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func splitVersion(v string) []string {
	var parts []string
	cur := ""
	for _, ch := range v {
		if ch >= '0' && ch <= '9' {
			cur += string(ch)
		} else if cur != "" {
			parts = append(parts, cur)
			cur = ""
		}
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	return parts
}

// AdminPublishVersion 发布新版本（管理员）
func AdminPublishVersion(c *gin.Context) {
	var req struct {
		Version      string `json:"version" binding:"required"`
		Build        int    `json:"build"`
		Platform     string `json:"platform"`
		UpdateURL    string `json:"update_url"`
		ReleaseNotes string `json:"release_notes"`
		Mandatory    bool   `json:"mandatory"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if req.Platform == "" {
		req.Platform = "all"
	}

	version := model.SystemVersion{
		Version:      req.Version,
		Build:        req.Build,
		Platform:     req.Platform,
		UpdateURL:    req.UpdateURL,
		ReleaseNotes: req.ReleaseNotes,
		Mandatory:    req.Mandatory,
		ReleasedAt:   time.Now(),
	}
	if err := model.DB.Create(&version).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "version.publish", "version", version.ID, gin.H{"version": req.Version, "platform": req.Platform})
	ok(c, gin.H{"version": version, "message": "版本已发布"})
}

// AdminListVersions 版本历史（管理员）
func AdminListVersions(c *gin.Context) {
	var versions []model.SystemVersion
	model.DB.Order("build DESC").Find(&versions)
	ok(c, gin.H{"versions": versions, "count": len(versions)})
}

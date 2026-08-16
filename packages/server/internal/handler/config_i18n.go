package handler

import (
	"github.com/gin-gonic/gin"
)

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

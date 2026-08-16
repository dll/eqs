package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

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

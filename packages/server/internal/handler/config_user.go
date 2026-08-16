package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

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

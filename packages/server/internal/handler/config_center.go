package handler

import (
	"encoding/json"
	"strconv"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

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

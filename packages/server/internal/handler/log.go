package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListMyLogs 当前用户自己的日志
// GET /api/v1/log/list
func ListMyLogs(c *gin.Context) {
	userID := c.GetUint("user_id")
	var logs []model.AuditLog
	model.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(100).Find(&logs)
	ok(c, gin.H{"logs": logs, "count": len(logs)})
}

// AdminListLogs 管理员查看所有日志（支持筛选）
// GET /api/v1/admin/log/list?user_id=&action=&page=&size=
func AdminListLogs(c *gin.Context) {
	query := model.DB.Model(&model.AuditLog{})
	if uid := c.Query("user_id"); uid != "" {
		if v, err := parseUint(uid); err == nil {
			query = query.Where("user_id = ?", v)
		}
	}
	if action := c.Query("action"); action != "" {
		query = query.Where("action LIKE ?", "%"+action+"%")
	}

	var total int64
	query.Count(&total)

	page := 1
	size := 50
	if p := c.Query("page"); p != "" {
		if v, err := parseUint(p); err == nil && v > 0 {
			page = int(v)
		}
	}
	if s := c.Query("size"); s != "" {
		if v, err := parseUint(s); err == nil && v > 0 && v <= 200 {
			size = int(v)
		}
	}

	var logs []model.AuditLog
	query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&logs)
	ok(c, gin.H{"logs": logs, "total": total, "page": page, "size": size})
}

// AdminRestoreConfig 管理员基于配置变更日志恢复配置
// POST /api/v1/admin/log/restore-config  body: {"log_id": 123}
// 依据 audit_log 中 config.upsert 的 detail 快照恢复配置值
func AdminRestoreConfig(c *gin.Context) {
	var req struct {
		LogID uint `json:"log_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var log model.AuditLog
	if err := model.DB.First(&log, req.LogID).Error; err != nil {
		notFound(c, "日志不存在")
		return
	}
	if log.TargetType != "config" {
		badRequest(c, "仅支持恢复配置类日志")
		return
	}

	// detail 为 JSON：{key, value, value_type}
	var detail struct {
		Key       string `json:"key"`
		Value     string `json:"value"`
		ValueType string `json:"value_type"`
	}
	if err := parseJSONString(log.Detail, &detail); err != nil || detail.Key == "" {
		badRequest(c, "日志详情缺少配置快照")
		return
	}

	// 恢复配置
	var cfg model.SystemConfig
	err := model.DB.Where("config_key = ?", detail.Key).First(&cfg).Error
	if err != nil {
		cfg = model.SystemConfig{ConfigKey: detail.Key, ValueType: detail.ValueType}
		model.DB.Create(&cfg)
	}
	cfg.ConfigValue = detail.Value
	cfg.ValueType = detail.ValueType
	model.DB.Save(&cfg)

	invalidatePublicCache()
	WriteAudit(c, "admin.config.restore", "config", cfg.ID, gin.H{"restored_from_log": log.ID, "key": detail.Key})
	ok(c, gin.H{"message": "配置已恢复", "key": detail.Key, "value": detail.Value})
}

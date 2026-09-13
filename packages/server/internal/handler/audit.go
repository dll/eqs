package handler

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// WriteAudit 记录关键状态变更审计日志（21.5：审计日志能够还原关键状态变更）
// P1-11：写入失败记录错误日志（不阻断主流程，但可观测）；DB 未初始化时安全跳过。
func WriteAudit(c *gin.Context, action, targetType string, targetID uint, detail interface{}) {
	detailBytes, err := json.Marshal(detail)
	if err != nil {
		detailBytes = []byte("{}")
	}

	var userID uint
	var ip string
	var requestID string
	if c != nil {
		userID = c.GetUint("user_id")
		ip = c.ClientIP()
		requestID = c.GetString("request_id")
	}

	if model.DB == nil {
		log.Printf("[audit] DB 未初始化，审计丢弃: %s/%s", action, targetType)
		return
	}

	entry := model.AuditLog{
		UserID:     userID,
		RequestID:  requestID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     string(detailBytes),
		IP:         ip,
	}

	// 带 3s 超时，避免审计写失败拖慢主流程
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := model.DB.WithContext(ctx).Create(&entry).Error; err != nil {
		log.Printf("[audit] 审计写入失败 action=%s target=%s/%d err=%v", action, targetType, targetID, err)
	}
}

// ListAuditLogs 审计检索（只读聚合；L1-A：补 OPC/监管可经 API 查阅审计的缺口）
// GET /api/v1/admin/audit/logs?user_id=&action=&target_type=&target_id=&start=&end=&page=&size=
// 仅平台/公司运营侧(RequireAdmin)可查；纯只读，不改写任何审计记录。
func ListAuditLogs(c *gin.Context) {
	page, size := parsePage(c)
	db := model.DB.Model(&model.AuditLog{})

	if v := c.Query("user_id"); v != "" {
		if id, err := parseUint(v); err == nil {
			db = db.Where("user_id = ?", id)
		}
	}
	if v := c.Query("action"); v != "" {
		db = db.Where("action = ?", v)
	}
	if v := c.Query("target_type"); v != "" {
		db = db.Where("target_type = ?", v)
	}
	if v := c.Query("target_id"); v != "" {
		if id, err := parseUint(v); err == nil {
			db = db.Where("target_id = ?", id)
		}
	}
	if v := c.Query("start"); v != "" {
		if t, err := time.Parse("2006-01-02 15:04", v); err == nil {
			db = db.Where("created_at >= ?", t)
		}
	}
	if v := c.Query("end"); v != "" {
		if t, err := time.Parse("2006-01-02 15:04", v); err == nil {
			db = db.Where("created_at <= ?", t)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		serverError(c, err)
		return
	}
	var logs []model.AuditLog
	if err := db.Order("created_at DESC").
		Offset((page - 1) * size).Limit(size).
		Find(&logs).Error; err != nil {
		serverError(c, err)
		return
	}
	if logs == nil {
		logs = []model.AuditLog{}
	}
	ok(c, gin.H{"logs": logs, "count": total, "page": page, "size": size})
}

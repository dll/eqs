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
	if c != nil {
		userID = c.GetUint("user_id")
		ip = c.ClientIP()
	}

	if model.DB == nil {
		log.Printf("[audit] DB 未初始化，审计丢弃: %s/%s", action, targetType)
		return
	}

	entry := model.AuditLog{
		UserID:     userID,
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

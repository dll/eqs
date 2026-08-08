package handler

import (
	"encoding/json"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// WriteAudit 记录关键状态变更审计日志（21.5：审计日志能够还原关键状态变更）
// 在 handler 关键节点调用，失败不影响主流程。
func WriteAudit(c *gin.Context, action, targetType string, targetID uint, detail interface{}) {
	detailBytes, err := json.Marshal(detail)
	if err != nil {
		detailBytes = []byte("{}")
	}

	entry := model.AuditLog{
		UserID:     c.GetUint("user_id"),
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Detail:     string(detailBytes),
		IP:         c.ClientIP(),
	}
	_ = model.DB.Create(&entry).Error
}

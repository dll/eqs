package handler

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ==================== 佣金计算（SRC-BIZ-01） ====================

// getCommissionRate 从配置中心读取佣金比例（%），默认 0（初期免费）
func getCommissionRate() float64 {
	cfgs := getPublicCached()
	switch v := cfgs["commission.rate"].(type) {
	case float64:
		if v >= 0 && v <= 100 {
			return v
		}
	case int:
		if v >= 0 && v <= 100 {
			return float64(v)
		}
	case string:
		var r float64
		if _, err := fmt.Sscanf(v, "%f", &r); err == nil && r >= 0 && r <= 100 {
			return r
		}
	}
	return 0
}

// calcAndCreateCommission 合同签署后按配置比例生成佣金单（幂等：同订单只生成一次）
func calcAndCreateCommission(orderID uint) {
	var order model.Order
	if err := model.DB.First(&order, orderID).Error; err != nil {
		return
	}

	// 幂等：已存在佣金单则跳过
	var count int64
	model.DB.Model(&model.CommissionRecord{}).Where("order_id = ?", orderID).Count(&count)
	if count > 0 {
		return
	}

	rate := getCommissionRate()
	if rate <= 0 {
		return
	}

	record := model.CommissionRecord{
		OrderID:    order.ID,
		SupplierID: order.SupplierID,
		ProjectID:  order.ProjectID,
		Amount:     order.Amount,
		Rate:       rate,
		Commission: order.Amount * rate / 100,
		Status:     "pending",
	}
	model.DB.Create(&record)
	WriteAudit(nil, "commission.create", "commission", record.ID,
		gin.H{"order_id": orderID, "rate": rate, "commission": record.Commission})
}

// AdminListCommissions 佣金记录列表（管理员）
func AdminListCommissions(c *gin.Context) {
	var records []model.CommissionRecord
	q := model.DB.Order("created_at DESC")
	if orderID := c.Query("order_id"); orderID != "" {
		q = q.Where("order_id = ?", orderID)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	q.Find(&records)

	var total float64
	var pending float64
	model.DB.Model(&model.CommissionRecord{}).Select("COALESCE(SUM(commission),0)").Scan(&total)
	model.DB.Model(&model.CommissionRecord{}).Where("status = ?", "pending").Select("COALESCE(SUM(commission),0)").Scan(&pending)

	ok(c, gin.H{"commissions": records, "count": len(records), "total_commission": total, "pending_commission": pending})
}

// AdminCollectCommission 标记佣金已收取（管理员）
func AdminCollectCommission(c *gin.Context) {
	id, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "佣金ID无效")
		return
	}
	var record model.CommissionRecord
	if err := model.DB.First(&record, id).Error; err != nil {
		notFound(c, "佣金记录不存在")
		return
	}
	if record.Status == "collected" {
		ok(c, gin.H{"message": "已收取", "idempotent": true})
		return
	}
	now := time.Now()
	model.DB.Model(&record).Updates(map[string]interface{}{
		"status":       "collected",
		"collected_at": now,
	})
	WriteAudit(c, "commission.collect", "commission", id, gin.H{"amount": record.Commission})
	ok(c, gin.H{"record": record, "message": "佣金已标记收取"})
}

// ==================== 信用评分动态重算（AC-10） ====================

// recalcUserCredit 按既有权重重算用户信用分（幂等：由原始数据推导，重复调用结果一致）
// score = 0.5*评价分 + 0.3*交付分 + 0.2*纠纷分
func recalcUserCredit(userID uint) {
	// 1. 评价分：平均评分 1-5 映射到 0-100
	var avg float64
	model.DB.Model(&model.Review{}).Where("reviewee_id = ?", userID).Select("AVG(rating)").Scan(&avg)
	reviewScore := 100.0
	if avg > 0 {
		reviewScore = 100 - (5-avg)*20
	}
	if reviewScore < 0 {
		reviewScore = 0
	}

	// 2. 交付分：已结算节点 / (已验收+已结算) 节点比例（按时交付率）
	var settled, accepted int64
	model.DB.Table("payment_milestones").
		Joins("JOIN orders ON orders.id = payment_milestones.order_id").
		Where("orders.supplier_id = ?", userID).
		Where("payment_milestones.status = ?", "settled").Count(&settled)
	model.DB.Table("payment_milestones").
		Joins("JOIN orders ON orders.id = payment_milestones.order_id").
		Where("orders.supplier_id = ?", userID).
		Where("payment_milestones.status IN (?)", []string{"settled", "accepted"}).Count(&accepted)
	deliveryScore := 100.0
	if accepted > 0 {
		deliveryScore = float64(settled) / float64(accepted) * 100
	}

	// 3. 纠纷分：每结案一单涉及纠纷扣 10 分（下限 0）
	var disputeCount int64
	model.DB.Model(&model.Dispute{}).Where("initiator_id = ? AND status = ?", userID, "closed").Count(&disputeCount)
	disputeScore := 100.0 - float64(disputeCount)*10
	if disputeScore < 0 {
		disputeScore = 0
	}

	score := reviewScore*0.5 + deliveryScore*0.3 + disputeScore*0.2
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	model.DB.Model(&model.User{}).Where("id = ?", userID).Update("credit_score", score)
	WriteAudit(nil, "credit.recalc", "user", userID, gin.H{"score": score})
}

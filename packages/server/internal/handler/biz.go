package handler

import (
	"fmt"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func getCommissionRate() float64 {
	var cfg model.SystemConfig
	if err := model.DB.Where("config_key = ?", "commission.rate").First(&cfg).Error; err != nil {
		return 0
	}
	var rate float64
	fmt.Sscanf(cfg.ConfigValue, "%f", &rate)
	// 费率以百分比表示（5 = 5%），范围 0-100
	if rate < 0 || rate > 100 {
		return 0
	}
	return rate
}

func calcAndCreateCommission(orderID uint) {
	var order model.Order
	if err := model.DB.First(&order, orderID).Error; err != nil {
		return
	}
	rate := getCommissionRate()
	if rate <= 0 {
		return
	}
	commission := order.Amount * rate / 100
	var cnt int64
	model.DB.Model(&model.CommissionRecord{}).Where("order_id = ?", orderID).Count(&cnt)
	if cnt > 0 {
		return
	}
	model.DB.Create(&model.CommissionRecord{
		OrderID:    orderID,
		SupplierID: order.SupplierID,
		ProjectID:  order.ProjectID,
		Amount:     order.Amount,
		Rate:       rate,
		Commission: commission,
		Status:     "pending",
	})
}

func AdminListCommissions(c *gin.Context) {
	var list []model.CommissionRecord
	model.DB.Preload("Order").Order("created_at DESC").Find(&list)
	ok(c, gin.H{"commissions": list, "count": len(list)})
}

func AdminCollectCommission(c *gin.Context) {
	commissionID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "佣金ID无效")
		return
	}
	var commission model.CommissionRecord
	if err := model.DB.First(&commission, commissionID).Error; err != nil {
		notFound(c, "佣金不存在")
		return
	}
	if commission.Status == "collected" {
		ok(c, gin.H{"message": "已收取", "idempotent": true})
		return
	}
	model.DB.Model(&commission).Update("status", "collected")
	WriteAudit(c, "commission.collect", "commission", commission.ID, gin.H{"amount": commission.Commission})
	ok(c, gin.H{"message": "佣金已收取"})
}

// recalcUserCredit P1-07：按 PRD 五维计算信用分
// score = 准时率30% + 质量30% + 纠纷20% + 活跃10% + 履约10%
func recalcUserCredit(userID uint) {
	// 1. 准时率 30%：已结算节点占已交付节点比例
	var settled, delivered int64
	model.DB.Table("payment_milestones").
		Joins("JOIN orders ON orders.id = payment_milestones.order_id").
		Where("orders.supplier_id = ?", userID).
		Where("payment_milestones.status = ?", "settled").Count(&settled)
	model.DB.Table("payment_milestones").
		Joins("JOIN orders ON orders.id = payment_milestones.order_id").
		Where("orders.supplier_id = ?", userID).
		Where("payment_milestones.status IN (?)", []string{"settled", "accepted"}).Count(&delivered)
	ontimeScore := 100.0
	if delivered > 0 {
		ontimeScore = float64(settled) / float64(delivered) * 100
	}

	// 2. 质量 30%：平均评价 1-5 映射到 0-100
	var avg float64
	model.DB.Model(&model.Review{}).Where("reviewee_id = ?", userID).Select("AVG(rating)").Scan(&avg)
	qualityScore := 100.0
	if avg > 0 {
		qualityScore = 100 - (5-avg)*20
	}
	if qualityScore < 0 {
		qualityScore = 0
	}

	// 3. 纠纷 20%：每结案纠纷扣 10 分，最低 0
	var disputeCount int64
	model.DB.Model(&model.Dispute{}).
		Joins("JOIN orders ON orders.id = disputes.order_id").
		Where("(disputes.initiator_id = ? OR orders.supplier_id = ?) AND disputes.status = ?", userID, userID, "closed").
		Count(&disputeCount)
	disputeScore := 100.0 - float64(disputeCount)*10
	if disputeScore < 0 {
		disputeScore = 0
	}

	// 4. 活跃度 10%：基于项目/订单/打卡/评价活动量
	var activityScore float64
	// 关联项目数
	var projCnt int64
	model.DB.Model(&model.Project{}).Where("user_id = ?", userID).Count(&projCnt)
	// 服务订单数
	var orderCnt int64
	model.DB.Model(&model.Order{}).Where("supplier_id = ?", userID).Count(&orderCnt)
	// 打卡数
	var attCnt int64
	model.DB.Model(&model.AttendanceRecord{}).Where("user_id = ?", userID).Count(&attCnt)
	// 评价数
	var reviewCnt int64
	model.DB.Model(&model.Review{}).Where("reviewer_id = ? OR reviewee_id = ?", userID, userID).Count(&reviewCnt)

	activity := float64(projCnt + orderCnt + attCnt + reviewCnt)
	activityScore = 100.0
	if activity < 5 {
		activityScore = activity / 5 * 100
	}

	// 5. 履约 10%：已完成订单占已签约订单比例（含作为甲方）
	var signedOrders, completedOrders int64
	model.DB.Model(&model.Order{}).
		Joins("JOIN projects ON projects.id = orders.project_id").
		Where("(orders.supplier_id = ? OR projects.user_id = ?) AND orders.status >= 1", userID, userID).Count(&signedOrders)
	model.DB.Model(&model.Order{}).
		Joins("JOIN projects ON projects.id = orders.project_id").
		Where("(orders.supplier_id = ? OR projects.user_id = ?) AND orders.status = 3", userID, userID).Count(&completedOrders)
	fulfillScore := 100.0
	if signedOrders > 0 {
		fulfillScore = float64(completedOrders) / float64(signedOrders) * 100
	}

	score := ontimeScore*0.3 + qualityScore*0.3 + disputeScore*0.2 + activityScore*0.1 + fulfillScore*0.1
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	model.DB.Model(&model.User{}).Where("id = ?", userID).Update("credit_score", score)
	WriteAudit(nil, "credit.recalc", "user", userID, gin.H{"score": score, "ontime": ontimeScore, "quality": qualityScore, "dispute": disputeScore, "activity": activityScore, "fulfill": fulfillScore})
}

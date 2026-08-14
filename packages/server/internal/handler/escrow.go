package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ==================== 资金托管台账 ====================
// recordEscrow 写入托管台账（冻结/释放/退款），供结算、争议流程联动。
func recordEscrow(orderID, milestoneID, disputeID, actorUserID uint, typ string, amount float64, note string) {
	if orderID == 0 || amount <= 0 {
		return
	}
	model.DB.Create(&model.EscrowLedger{
		OrderID:     orderID,
		MilestoneID: milestoneID,
		DisputeID:   disputeID,
		ActorUserID: actorUserID,
		Type:        typ,
		Amount:      amount,
		Note:        note,
	})
}

// escrowTotals 计算订单托管汇总（只读）
func escrowTotals(orderID uint) (total, released, frozen float64, ledger []model.EscrowLedger) {
	// 托管总量 = 里程碑金额合计（节点验收结算后按比例释放）
	var ms []model.PaymentMilestone
	model.DB.Where("order_id = ?", orderID).Find(&ms)
	for _, m := range ms {
		total += m.Amount
	}

	model.DB.Where("order_id = ?", orderID).Order("created_at ASC").Find(&ledger)
	for _, l := range ledger {
		switch l.Type {
		case "release":
			released += l.Amount
		case "freeze":
			frozen += l.Amount
		case "refund":
			frozen -= l.Amount
		}
	}
	return total, released, frozen, ledger
}

// GetOrderEscrow 订单资金托管明细（参与方/管理员可见）
// GET /api/v1/order/:id/escrow
func GetOrderEscrow(c *gin.Context) {
	orderID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "订单ID无效")
		return
	}
	order, allowed := getOrderForUser(c, orderID)
	if !allowed {
		forbidden(c, "无权查看该订单资金托管")
		return
	}
	total, released, frozen, ledger := escrowTotals(order.ID)
	balance := total - released
	if balance < 0 {
		balance = 0
	}
	ok(c, gin.H{
		"order_id":   order.ID,
		"escrow_total":  total, // 托管总量（按节点）
		"released":      released,
		"frozen":        frozen,
		"balance":       balance,
		"ledger":        ledger,
		"ledger_count":  len(ledger),
	})
}

// AdminListEscrowLedger 平台资金托管全量台账（对账）
// GET /api/v1/admin/escrow/ledger?page=&size=
func AdminListEscrowLedger(c *gin.Context) {
	query := model.DB.Model(&model.EscrowLedger{})
	if orderID := c.Query("order_id"); orderID != "" {
		if v, err := parseUint(orderID); err == nil {
			query = query.Where("order_id = ?", v)
		}
	}
	if typ := c.Query("type"); typ != "" {
		query = query.Where("type = ?", typ)
	}

	var total int64
	query.Count(&total)
	page, size := parsePage(c)

	var ledger []model.EscrowLedger
	query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&ledger)
	ok(c, gin.H{"ledger": ledger, "total": total, "page": page, "size": size})
}

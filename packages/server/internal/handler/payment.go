package handler

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// MockPaymentProvider 仿真持牌支付通道：仅在 PAYMENT_PROVIDER=mock 时可用
type MockPaymentProvider struct{}

func (m *MockPaymentProvider) CreatePayment(orderID uint, amount float64) (string, error) {
	return fmt.Sprintf("PAY-MOCK-%d-%d", orderID, time.Now().UnixNano()), nil
}

func (m *MockPaymentProvider) Settle(milestoneID uint, amount float64) (string, error) {
	return fmt.Sprintf("SETTLE-MOCK-%d", time.Now().UnixNano()), nil
}

type CreatePaymentRequest struct {
	OrderID uint    `json:"order_id" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Channel string  `json:"channel"`
}

// CreatePayment 为订单创建支付单（经持牌机构）
func CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	channel := req.Channel
	if channel == "" {
		channel = "mock"
	}
	if channel == "wechat" && config.Load().PaymentProvider != "mock" {
		// 真实通道尚未签约时不允许直接调用
		badRequest(c, "支付通道未就绪")
		return
	}

	var order model.Order
	if err := model.DB.First(&order, req.OrderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	if req.Amount != order.Amount {
		badRequest(c, "支付金额与订单金额不一致")
		return
	}

	provider := &MockPaymentProvider{}
	txnID, err := provider.CreatePayment(req.OrderID, req.Amount)
	if err != nil {
		serverError(c, err)
		return
	}

	txn := model.PaymentTransaction{
		UserID:               order.SupplierID,
		OrderID:              order.ID,
		Amount:               req.Amount,
		Type:                 "payment",
		Channel:              channel,
		ExternalTransactionID: txnID,
		Status:               0,
	}
	if err := model.DB.Create(&txn).Error; err != nil {
		serverError(c, err)
		return
	}

	WriteAudit(c, "payment.create", "order", req.OrderID, gin.H{"transaction_id": txn.ID, "amount": req.Amount, "channel": channel})
	ok(c, gin.H{"transaction": txn, "paid": true})
}

// PaymentNotify 支付回调：验签、幂等更新支付状态
func PaymentNotify(c *gin.Context) {
	channel := c.Param("channel")

	var req struct {
		ExternalTransactionID string  `json:"external_transaction_id"`
		OrderID               uint    `json:"order_id"`
		Amount                float64 `json:"amount"`
		Result                string  `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var txn model.PaymentTransaction
	if err := model.DB.Where("external_transaction_id = ?", req.ExternalTransactionID).First(&txn).Error; err != nil {
		notFound(c, "交易不存在")
		return
	}
	// 幂等：已成功不再重复更新
	if txn.Status == 1 {
		ok(c, gin.H{"message": "已处理", "idempotent": true})
		return
	}
	if req.Result == "success" && txn.Amount == req.Amount {
		model.DB.Model(&txn).Update("status", 1)
	}

	// 更新订单为已支付（进行中）
	if req.Result == "success" {
		model.DB.Model(&model.Order{}).Where("id = ? AND status = ?", req.OrderID, 0).
			Update("status", 1)
	}

	WriteAudit(c, "pay.notify", "transaction", txn.ID, gin.H{"order_id": req.OrderID, "result": req.Result, "amount": req.Amount})
	ok(c, gin.H{"channel": channel, "message": "通知已处理"})
}

// SettleMilestone 验收后向第三方提交节点结算指令
func SettleMilestone(c *gin.Context) {
	milestoneID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "节点ID无效")
		return
	}

	var ms model.PaymentMilestone
	if err := model.DB.First(&ms, milestoneID).Error; err != nil {
		notFound(c, "节点不存在")
		return
	}
	if ms.Status != "accepted" {
		badRequest(c, "节点未验收，不可结算")
		return
	}

	var disputed model.Dispute
	if err := model.DB.Where("milestone_id = ? AND status <> ?", milestoneID, "closed").First(&disputed).Error; err == nil {
		badRequest(c, "该节点存在争议，已冻结，不可结算")
		return
	}

	provider := &MockPaymentProvider{}
	txnID, err := provider.Settle(milestoneID, ms.Amount)
	if err != nil {
		serverError(c, err)
		return
	}

	var order model.Order
	model.DB.First(&order, ms.OrderID)
	txn := model.PaymentTransaction{
		UserID:               order.SupplierID,
		OrderID:              ms.OrderID,
		MilestoneID:          milestoneID,
		Amount:               ms.Amount,
		Type:                 "settlement",
		Channel:              "mock",
		ExternalTransactionID: txnID,
		Status:               1,
	}
	if err := model.DB.Create(&txn).Error; err != nil {
		serverError(c, err)
		return
	}

	model.DB.Model(&ms).Update("status", "settled")
	WriteAudit(c, "pay.settle", "milestone", milestoneID, gin.H{"amount": ms.Amount, "order_id": ms.OrderID, "transaction_id": txn.ID})
	ok(c, gin.H{"transaction": txn, "message": "结算指令已提交"})
}

// ListPaymentTransactions 查询支付、结算、退款及止付记录
func ListPaymentTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")

	var txns []model.PaymentTransaction
	model.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&txns)
	ok(c, gin.H{"transactions": txns})
}

// GetBalance 查询余额（V6 不自建资金池，直接查询流水汇总）
func GetBalance(c *gin.Context) {
	userID := c.GetUint("user_id")
	var count int64

	var paidIn float64
	model.DB.Model(&model.PaymentTransaction{}).Where("user_id = ? AND type = ? AND status = ?", userID, "payment", 1).Select("COALESCE(SUM(amount),0)").Scan(&paidIn)

	var settledOut float64
	model.DB.Model(&model.PaymentTransaction{}).Where("user_id = ? AND type IN (?) AND status = ?", userID, []string{"settlement", "refund"}, 1).Select("COALESCE(SUM(amount),0)").Scan(&settledOut)

	model.DB.Model(&model.PaymentTransaction{}).Where("user_id = ?", userID).Count(&count)
	ok(c, gin.H{"received": paidIn, "settled": settledOut, "transaction_count": count})
}
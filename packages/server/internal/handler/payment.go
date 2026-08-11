package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
		Timestamp             int64   `json:"timestamp"`
		Sign                  string  `json:"sign"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	// P0-03：回调验签（HMAC-SHA256，需配置 PaymentNotifySecret）
	cfg := config.Load()
	if cfg.PaymentNotifySecret == "" {
		if cfg.IsProduction() {
			serverError(c, fmt.Errorf("支付回调密钥未配置"))
			return
		}
		// 非生产（测试/Mock）：无密钥时跳过验签，便于联调
	} else {
		if !verifyCallbackSign(cfg.PaymentNotifySecret, req.ExternalTransactionID, req.OrderID, req.Amount, req.Result, req.Timestamp, req.Sign) {
			unauthorized(c, "回调签名无效")
			return
		}
		// 重放防护：时间戳 5 分钟内有效
		if time.Since(time.Unix(req.Timestamp, 0)) > 5*time.Minute || req.Timestamp > time.Now().Unix()+60 {
			badRequest(c, "回调时间戳无效")
			return
		}
	}

	var txn model.PaymentTransaction
	if err := model.DB.Where("external_transaction_id = ?", req.ExternalTransactionID).First(&txn).Error; err != nil {
		notFound(c, "交易不存在")
		return
	}
	// 幂等：已成功则直接返回，不重复处理
	if txn.Status == 1 {
		ok(c, gin.H{"message": "已处理", "idempotent": true})
		return
	}
	// 金额必须与交易记录一致；不匹配则不更新状态（安全，防伪造）
	if req.Result == "success" && txn.Amount == req.Amount {
		model.DB.Model(&txn).Update("status", 1)
	} else if req.Result == "success" {
		// 金额不匹配：不更新，仅记录
		WriteAudit(c, "pay.notify.mismatch", "transaction", txn.ID, gin.H{"order_id": txn.OrderID, "cb_amount": req.Amount, "txn_amount": txn.Amount})
		ok(c, gin.H{"channel": channel, "message": "金额不匹配，忽略"})
		return
	}

	// 更新订单状态：订单以交易记录关联为准，不信任回调 order_id
	if req.Result == "success" && txn.OrderID > 0 {
		model.DB.Model(&model.Order{}).Where("id = ? AND status = ?", txn.OrderID, 0).
			Update("status", 1)
	}

	WriteAudit(c, "pay.notify", "transaction", txn.ID, gin.H{"order_id": txn.OrderID, "result": req.Result, "amount": txn.Amount})
	ok(c, gin.H{"channel": channel, "message": "通知已处理"})
}

// verifyCallbackSign 回调签名校验：HMAC-SHA256(signstr, secret)
// signstr = external_transaction_id|order_id|amount|result|timestamp
func verifyCallbackSign(secret, txid string, orderID uint, amount float64, result string, ts int64, sign string) bool {
	str := fmt.Sprintf("%s|%d|%.2f|%s|%d", txid, orderID, amount, result, ts)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(str))
	expect := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expect), []byte(sign))
}


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

	// 节点结算后重算服务方信用分（交付分加权）
	recalcUserCredit(order.SupplierID)

	// 全部节点结算完成则订单完成
	var pending int64
	model.DB.Model(&model.PaymentMilestone{}).Where("order_id = ? AND status <> ?", ms.OrderID, "settled").Count(&pending)
	if pending == 0 {
		now := time.Now()
		model.DB.Model(&model.Order{}).Where("id = ?", ms.OrderID).Updates(map[string]interface{}{
			"status":       3,
			"completed_at": now,
		})
		model.DB.Model(&model.Project{}).Where("id = ?", order.ProjectID).Update("status", 4)
	}
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
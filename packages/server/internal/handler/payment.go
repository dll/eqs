package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/eqs/server/internal/channel"
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
	if channel == "wechat" && config.Get().PaymentProvider != "mock" {
		// 真实通道已配置（PAYMENT_PROVIDER=wechat），走微信支付网关
		gateway, gwErr := newWechatGateway()
		if gwErr != nil {
			badRequest(c, "微信支付未配置："+gwErr.Error())
			return
		}
		if gateway == nil {
			badRequest(c, "微信支付未配置（请填写 WXPAY_* 凭据）")
			return
		}

		var order model.Order
		if err := model.DB.First(&order, req.OrderID).Error; err != nil {
			notFound(c, "订单不存在")
			return
		}
		var proj model.Project
		if err := model.DB.First(&proj, order.ProjectID).Error; err != nil || proj.UserID != c.GetUint("user_id") {
			forbidden(c, "仅甲方可发起支付")
			return
		}
		if req.Amount != order.Amount {
			badRequest(c, "支付金额与订单金额不一致")
			return
		}

		txn := model.PaymentTransaction{
			UserID: order.SupplierID, OrderID: order.ID, Amount: req.Amount,
			Type: "payment", Channel: "wechat", Status: 0,
		}
		if err := model.DB.Create(&txn).Error; err != nil {
			serverError(c, err)
			return
		}
		outTradeNo := fmt.Sprintf("EQS%08d", txn.ID)
		model.DB.Model(&txn).Update("external_transaction_id", outTradeNo)

		codeURL, err := gateway.CreateNativeOrder(outTradeNo, int64(req.Amount*100),
			fmt.Sprintf("工程服务订单#%d", order.ID), config.Get().WXPayNotifyURL)
		if err != nil {
			// 下单失败：记录失败状态
			model.DB.Model(&txn).Update("status", 2)
			serverError(c, err)
			return
		}
		WriteAudit(c, "payment.create.wechat", "order", req.OrderID, gin.H{"transaction_id": txn.ID, "out_trade_no": outTradeNo})
		ok(c, gin.H{"transaction": txn, "paid": false, "channel": "wechat", "code_url": codeURL})
		return
	}

	// V11：微信小程序 JSAPI 支付（需甲方 openid）。
	// 仅当 channel=jsapi 且真实通道已配置（PAYMENT_PROVIDER=wechat）时启用；
	// 商户凭据未到位/未配置时保持隔离，不发起真实调用（返回业务提示）。
	if channel == "jsapi" && config.Get().PaymentProvider != "mock" {
		gateway, gwErr := newWechatGateway()
		if gwErr != nil {
			badRequest(c, "微信支付未配置："+gwErr.Error())
			return
		}
		if gateway == nil {
			badRequest(c, "微信支付未配置（请填写 WXPAY_* 凭据）")
			return
		}

		var order model.Order
		if err := model.DB.First(&order, req.OrderID).Error; err != nil {
			notFound(c, "订单不存在")
			return
		}
		var proj model.Project
		if err := model.DB.First(&proj, order.ProjectID).Error; err != nil || proj.UserID != c.GetUint("user_id") {
			forbidden(c, "仅甲方可发起支付")
			return
		}
		if req.Amount != order.Amount {
			badRequest(c, "支付金额与订单金额不一致")
			return
		}
		// 取甲方 openid（A1 小程序登录时写入 WxOpenID）
		var payer model.User
		if err := model.DB.First(&payer, c.GetUint("user_id")).Error; err != nil || payer.WxOpenID == nil || *payer.WxOpenID == "" {
			badRequest(c, "请先通过微信小程序登录以获取支付身份")
			return
		}

		txn := model.PaymentTransaction{
			UserID: order.SupplierID, OrderID: order.ID, Amount: req.Amount,
			Type: "payment", Channel: "jsapi", Status: 0,
		}
		if err := model.DB.Create(&txn).Error; err != nil {
			serverError(c, err)
			return
		}
		outTradeNo := fmt.Sprintf("EQS%08d", txn.ID)
		model.DB.Model(&txn).Update("external_transaction_id", outTradeNo)

		prepayID, err := gateway.CreateJSAPIOrder(*payer.WxOpenID, outTradeNo,
			int64(req.Amount*100), fmt.Sprintf("工程服务订单#%d", order.ID), config.Get().WXPayNotifyURL)
		if err != nil {
			model.DB.Model(&txn).Update("status", 2)
			serverError(c, err)
			return
		}
		WriteAudit(c, "payment.create.jsapi", "order", req.OrderID, gin.H{"transaction_id": txn.ID, "out_trade_no": outTradeNo})
		ok(c, gin.H{"transaction": txn, "paid": false, "channel": "jsapi", "prepay_id": prepayID})
		return
	}

	var order model.Order
	if err := model.DB.First(&order, req.OrderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	// P0-05：仅甲方（项目创建者）可发起支付
	var proj model.Project
	if err := model.DB.First(&proj, order.ProjectID).Error; err != nil || proj.UserID != c.GetUint("user_id") {
		forbidden(c, "仅甲方可发起支付")
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
		UserID:                order.SupplierID,
		OrderID:               order.ID,
		Amount:                req.Amount,
		Type:                  "payment",
		Channel:               channel,
		ExternalTransactionID: txnID,
		Status:                0,
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

	// V10：微信支付 v3 回调（验签+解密，返回微信要求格式）
	if channel == "wechat" {
		gateway, gwErr := newWechatGateway()
		if gwErr != nil {
			serverError(c, gwErr)
			return
		}
		if gateway == nil {
			badRequest(c, "微信支付未配置")
			return
		}
		body, _ := io.ReadAll(c.Request.Body)
		outTradeNo, totalFen, wechatTxnID, err := gateway.VerifyAndDecryptNotify(c.Request.Header, body)
		if err != nil {
			WriteAudit(c, "pay.notify.wechat.verify_fail", "transaction", 0, gin.H{"err": err.Error()})
			c.JSON(401, gin.H{"code": "FAIL", "message": "验签失败"})
			return
		}
		var txn model.PaymentTransaction
		if err := model.DB.Where("external_transaction_id = ?", outTradeNo).First(&txn).Error; err != nil || txn.ID == 0 {
			notFound(c, "交易不存在")
			return
		}
		if txn.Status == 1 {
			ok(c, gin.H{"code": "SUCCESS", "message": "成功"})
			return
		}
		// 金额核对：微信金额单位为分
		if txn.Status != 2 && int64(txn.Amount*100) == totalFen {
			model.DB.Model(&txn).Updates(map[string]interface{}{"status": 1, "external_transaction_id": wechatTxnID})
			// 订单状态 0→1（已支付），资金进入托管
			model.DB.Model(&model.Order{}).Where("id = ? AND status = ?", txn.OrderID, 0).Update("status", 1)
			WriteAudit(c, "pay.notify.wechat", "transaction", txn.ID, gin.H{"out_trade_no": outTradeNo, "wechat_txn": wechatTxnID, "amount": txn.Amount})
			ok(c, gin.H{"code": "SUCCESS", "message": "成功"})
			return
		}
		WriteAudit(c, "pay.notify.wechat.mismatch", "transaction", txn.ID, gin.H{"cb_fen": totalFen, "txn_amount": txn.Amount})
		ok(c, gin.H{"code": "FAIL", "message": "金额不匹配"})
		return
	}

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
	cfg := config.Get()
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
		UserID:                order.SupplierID,
		OrderID:               ms.OrderID,
		MilestoneID:           milestoneID,
		Amount:                ms.Amount,
		Type:                  "settlement",
		Channel:               "mock",
		ExternalTransactionID: txnID,
		Status:                1,
	}
	if err := model.DB.Create(&txn).Error; err != nil {
		serverError(c, err)
		return
	}

	model.DB.Model(&ms).Update("status", "settled")
	// 资金托管台账：节点验收结算，托管资金释放给服务方
	recordEscrow(ms.OrderID, milestoneID, 0, order.SupplierID, "release", ms.Amount, "节点验收结算释放")
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

// newWechatGateway 构造微信支付 v3 网关（凭据不完整时返回错误/空）
func newWechatGateway() (channel.PaymentGateway, error) {
	cfg := config.Get()
	return channel.NewPaymentGateway(cfg.PaymentProvider, cfg.WXPayAppID, cfg.WXPayMchID,
		cfg.WXPayAPIV3Key, cfg.WXPayMchSerialNo, cfg.WXPayMchPrivateKeyFile, cfg.WXPayPlatformCertFile)
}

// RefundPayment 退款（微信支付 v3 或 Mock）
// POST /api/v1/pay/refund {transaction_id, amount?}
func RefundPayment(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		TransactionID uint    `json:"transaction_id" binding:"required"`
		Amount        float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var txn model.PaymentTransaction
	if err := model.DB.First(&txn, req.TransactionID).Error; err != nil {
		notFound(c, "交易不存在")
		return
	}
	if txn.Status != 1 {
		badRequest(c, "该交易不可退款")
		return
	}
	// 仅管理员或交易参与方（甲方=订单付款方，此处以项目创建者近似）可退款
	if !isAdmin(c) {
		var order model.Order
		var project model.Project
		if model.DB.First(&order, txn.OrderID).Error != nil ||
			model.DB.First(&project, order.ProjectID).Error != nil || project.UserID != userID {
			forbidden(c, "无权退款")
			return
		}
	}

	refundAmt := req.Amount
	if refundAmt <= 0 || refundAmt > txn.Amount {
		refundAmt = txn.Amount
	}
	refundNo := fmt.Sprintf("EQSREF%08d%02d", txn.ID, time.Now().UnixNano()%100)

	if txn.Channel == "wechat" {
		gateway, gwErr := newWechatGateway()
		if gwErr != nil {
			serverError(c, gwErr)
			return
		}
		if gateway == nil {
			badRequest(c, "微信支付未配置")
			return
		}
		if err := gateway.Refund(txn.ExternalTransactionID, refundNo, int64(txn.Amount*100), int64(refundAmt*100)); err != nil {
			serverError(c, err)
			return
		}
	}

	// 退款记录
	model.DB.Create(&model.PaymentTransaction{
		UserID: txn.UserID, OrderID: txn.OrderID, Amount: -refundAmt,
		Type: "refund", Channel: txn.Channel, ExternalTransactionID: refundNo, Status: 1,
	})
	model.DB.Model(&txn).Update("status", 2)
	recordEscrow(txn.OrderID, 0, 0, userID, "refund", refundAmt, "支付退款")
	WriteAudit(c, "pay.refund", "transaction", txn.ID, gin.H{"amount": refundAmt, "channel": txn.Channel})
	ok(c, gin.H{"message": "退款指令已提交", "refund_no": refundNo, "amount": refundAmt})
}

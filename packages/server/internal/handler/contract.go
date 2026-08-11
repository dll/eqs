package handler

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListContractTemplates 按服务类型获取有效合同模板
func ListContractTemplates(c *gin.Context) {
	serviceType := c.Query("service_type")

	q := model.DB.Where("status = ?", "active")
	if serviceType != "" {
		q = q.Where("service_type = ?", serviceType)
	}

	var templates []model.ContractTemplate
	if err := q.Find(&templates).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"templates": templates})
}

// GenerateContract 根据订单、报价和付款节点生成合同草稿
func GenerateContract(c *gin.Context) {
	orderID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "订单ID无效")
		return
	}

	var order model.Order
	if err := model.DB.First(&order, orderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}

	// 已有合同草稿则不重复生成
	var existing model.Contract
	if err := model.DB.Where("order_id = ?", orderID).First(&existing).Error; err == nil {
		ok(c, gin.H{"contract": existing})
		return
	}

	var template model.ContractTemplate
	if err := model.DB.Where("service_type = ? AND status = ?", "cost", "active").First(&template).Error; err != nil {
		// 无模板时使用默认文本
		template = model.ContractTemplate{Name: "默认合同模板", Version: "1.0"}
	}

	contract := model.Contract{
		OrderID:         orderID,
		TemplateID:      template.ID,
		TemplateVersion: template.Version,
		ContractNo:      fmt.Sprintf("EQS-%d-%d", time.Now().Year(), orderID),
		SignProvider:    "mock",
		SignFlowID:      fmt.Sprintf("SIGN-%d-%d", time.Now().Unix(), orderID),
		Status:          "draft",
	}
	if err := model.DB.Create(&contract).Error; err != nil {
		serverError(c, err)
		return
	}

	ok(c, gin.H{"contract": contract})
}

// SignContract 发起第三方电子签署流程（mock 直接完成）
func SignContract(c *gin.Context) {
	contractID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "合同ID无效")
		return
	}

	var contract model.Contract
	if err := model.DB.First(&contract, contractID).Error; err != nil {
		notFound(c, "合同不存在")
		return
	}
	// P0-05：仅参与方可签署
	var order model.Order
	if err := model.DB.First(&order, contract.OrderID).Error; err != nil || !canAccessOrder(c, &order) {
		forbidden(c, "无权签署该合同")
		return
	}
	if contract.Status == "signed" {
		ok(c, gin.H{"message": "已签署", "contract": contract})
		return
	}

	now := time.Now()
	if err := model.DB.Model(&contract).Updates(map[string]interface{}{
		"status":   "signed",
		"signed_at": now,
	}).Error; err != nil {
		serverError(c, err)
		return
	}

	// 签署完成后订单进入待支付/进行中
	model.DB.Model(&model.Order{}).Where("id = ? AND status = ?", contract.OrderID, 0).
		Update("status", 1)
	model.DB.Model(&model.Order{}).Where("id = ?", contract.OrderID).Update("signed_at", now)

	// 合同签署后按配置生成平台佣金（幂等）
	calcAndCreateCommission(contract.OrderID)

	WriteAudit(c, "contract.sign", "contract", contractID, gin.H{"order_id": contract.OrderID})
	ok(c, gin.H{"contract": contract, "message": "签署完成"})
}

// SignNotify 电子签约回调：验签、幂等更新签署状态并归档
func SignNotify(c *gin.Context) {
	var req struct {
		SignFlowID string `json:"sign_flow_id"`
		OrderID    uint   `json:"order_id"`
		Result     string `json:"result"`
		Timestamp  int64  `json:"timestamp"`
		Sign       string `json:"sign"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	// P0-04：回调验签（HMAC-SHA256，需配置 ESignAPIKey）
	cfg := config.Load()
	if cfg.ESignAPIKey == "" {
		if cfg.IsProduction() {
			serverError(c, fmt.Errorf("电子签回调密钥未配置"))
			return
		}
		// 非生产（测试/Mock）：无密钥时跳过验签，便于联调
	} else {
		if !verifyCallbackSign(cfg.ESignAPIKey, req.SignFlowID, req.OrderID, 0, req.Result, req.Timestamp, req.Sign) {
			unauthorized(c, "回调签名无效")
			return
		}
		// 重放防护：时间戳 5 分钟内有效
		if time.Since(time.Unix(req.Timestamp, 0)) > 5*time.Minute || req.Timestamp > time.Now().Unix()+60 {
			badRequest(c, "回调时间戳无效")
			return
		}
	}

	var contract model.Contract
	if err := model.DB.Where("sign_flow_id = ?", req.SignFlowID).First(&contract).Error; err != nil {
		notFound(c, "签署流程不存在")
		return
	}
	if contract.Status == "signed" {
		ok(c, gin.H{"message": "已处理", "idempotent": true})
		return
	}

	if req.Result == "signed" {
		now := time.Now()
		model.DB.Model(&contract).Updates(map[string]interface{}{"status": "signed", "signed_at": now})
		// 订单以合同关联为准，不信任回调 order_id
		if contract.OrderID > 0 {
			model.DB.Model(&model.Order{}).Where("id = ? AND status = ?", contract.OrderID, 0).Update("status", 1)
		}
	}

	ok(c, gin.H{"message": "签署回调已处理"})
}


func DownloadContract(c *gin.Context) {
	contractID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "合同ID无效")
		return
	}

	var contract model.Contract
	if err := model.DB.First(&contract, contractID).Error; err != nil {
		notFound(c, "合同不存在")
		return
	}
	ok(c, gin.H{"contract": contract, "file_id": contract.ContractFileID})
}
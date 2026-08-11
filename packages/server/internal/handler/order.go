package handler

import (
	"strconv"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UpdateOrder 更新订单（甲方可改备注/期望工期等非资金字段；未签约前可调整金额需服务方确认）
func UpdateOrder(c *gin.Context) {
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
	if !canAccessOrder(c, &order) {
		forbidden(c, "无权操作该订单")
		return
	}

	var req struct {
		Remark    string  `json:"remark"`
		Amount    float64 `json:"amount"` // 未签约前可调整
		ExpectDays int    `json:"expect_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.ExpectDays > 0 {
		updates["expect_days"] = req.ExpectDays
	}
	// 未签约（status=0）时允许调整金额
	if req.Amount > 0 && order.Status == 0 {
		updates["amount"] = req.Amount
	}
	if len(updates) == 0 {
		badRequest(c, "没有可更新的字段")
		return
	}
	if err := model.DB.Model(&order).Updates(updates).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "order.update", "order", orderID, gin.H{"fields": len(updates)})
	ok(c, gin.H{"message": "订单已更新"})
}

// CancelOrder 取消订单（甲方或服务方均可；仅未签约状态可取消，已签约需走争议）
func CancelOrder(c *gin.Context) {
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
	if !canAccessOrder(c, &order) {
		forbidden(c, "无权操作该订单")
		return
	}
	if order.Status != 0 {
		badRequest(c, "仅未签约订单可取消，已签约请通过争议流程处理")
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	now := time.Now()
	if err := model.DB.Model(&order).Updates(map[string]interface{}{
		"status":      6, // 6 = 已取消
		"cancel_reason": req.Reason,
		"cancelled_at":  now,
	}).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "order.cancel", "order", orderID, gin.H{"reason": req.Reason})
	// 通知对端
	notifyOrderParty(&order, "订单已取消", "订单 #"+uint2str(orderID)+" 已被取消："+req.Reason)
	ok(c, gin.H{"message": "订单已取消"})
}

// uint2str 订单号转字符串（辅助通知）
func uint2str(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

// notifyOrderParty 向订单对端发送通知（甲方 or 服务方）
func notifyOrderParty(order *model.Order, title, content string) {
	if order == nil {
		return
	}
	var proj model.Project
	target := order.SupplierID
	if model.DB.First(&proj, order.ProjectID).Error == nil {
		// 通知非操作方：服务方固定，甲方=项目创建者
		_ = proj.UserID
	}
	CreateNotification(target, title, content, "order")
}

// GetOrder 订单详情：订单、合同、节点及资金状态
func GetOrder(c *gin.Context) {
	orderID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "订单ID无效")
		return
	}

	var order model.Order
	if err := model.DB.Preload("Project").Preload("Supplier").First(&order, orderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	// P0-05：仅管理员或订单参与方可查看
	if !canAccessOrder(c, &order) {
		forbidden(c, "无权查看该订单")
		return
	}

	var milestones []model.PaymentMilestone
	model.DB.Where("order_id = ?", orderID).Order("sequence ASC").Find(&milestones)

	var contract *model.Contract
	model.DB.Where("order_id = ?", orderID).First(&contract)

	var payments []model.PaymentTransaction
	model.DB.Where("order_id = ?", orderID).Order("created_at DESC").Find(&payments)

	ok(c, gin.H{
		"order":      order,
		"milestones": milestones,
		"contract":   contract,
		"payments":   payments,
	})
}

type Milestone struct {
	Name  string  `json:"name" binding:"required"`
	Ratio float64 `json:"ratio" binding:"required"`
}

type MilestoneRequest struct {
	Name  string  `json:"name" binding:"required"`
	Ratio float64 `json:"ratio" binding:"required"`
}

// SetMilestones 签约前设置付款节点，金额合计必须等于订单金额
func SetMilestones(c *gin.Context) {
	orderID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "订单ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var order model.Order
	if err := model.DB.First(&order, orderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	if order.Status != 0 {
		badRequest(c, "订单已签约，不能再设置付款节点")
		return
	}
	// 仅发布项目的甲方可设置节点
	var project model.Project
	if err := model.DB.First(&project, order.ProjectID).Error; err != nil {
		serverError(c, err)
		return
	}
	if project.UserID != userID {
		forbidden(c, "仅甲方可设置节点")
		return
	}

	var req struct {
		Milestones []Milestone `json:"milestones" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if len(req.Milestones) == 0 {
		badRequest(c, "至少设置一个节点")
		return
	}

	// 校验节点金额合计等于订单金额
	totalRatio := 0.0
	totalAmount := 0.0
	for _, m := range req.Milestones {
		totalRatio += m.Ratio
		totalAmount += order.Amount * m.Ratio / 100
	}
	if totalRatio-100 > 0.001 || 100-totalRatio > 0.001 {
		badRequest(c, "节点比例合计必须等于100%")
		return
	}

	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("order_id = ?", orderID).Delete(&model.PaymentMilestone{}).Error; err != nil {
			return err
		}
		for i, m := range req.Milestones {
			ms := model.PaymentMilestone{
				OrderID:  orderID,
				Name:     m.Name,
				Sequence: i + 1,
				Ratio:    m.Ratio,
				Amount:   order.Amount * m.Ratio / 100,
				Status:   "pending",
			}
			if err := tx.Create(&ms).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		serverError(c, err)
		return
	}

	WriteAudit(c, "order.milestones", "order", orderID, gin.H{"nodes": len(req.Milestones), "total_ratio": totalRatio})
	ok(c, gin.H{"message": "付款节点已设置"})
}

type UploadDeliverableRequest struct {
	FileName string `json:"file_name" binding:"required"`
	FileURL  string `json:"file_url" binding:"required"`
}

// UploadDeliverable 上传指定节点的阶段性交付物，MVP版交付物直接落库
func UploadDeliverable(c *gin.Context) {
	milestoneID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "节点ID无效")
		return
	}

	var req UploadDeliverableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var ms model.PaymentMilestone
	if err := model.DB.First(&ms, milestoneID).Error; err != nil {
		notFound(c, "节点不存在")
		return
	}
	if ms.Status != "pending" && ms.Status != "submitted" {
		badRequest(c, "当前节点不可上传交付物")
		return
	}
	// P0-05：仅订单服务方可上传交付物
	var order model.Order
	if err := model.DB.First(&order, ms.OrderID).Error; err != nil || order.SupplierID != c.GetUint("user_id") {
		forbidden(c, "仅服务方可上传交付物")
		return
	}

	// 版本号递增
	var maxVer int
	model.DB.Model(&model.Deliverable{}).Where("order_id = ? AND milestone_id = ?", ms.OrderID, milestoneID).
		Select("COALESCE(MAX(version),0)").Scan(&maxVer)

	deliverable := model.Deliverable{
		OrderID:     ms.OrderID,
		MilestoneID: ms.ID,
		Milestone:   ms.Name,
		FileURL:     req.FileURL,
		FileName:    req.FileName,
		Version:     maxVer + 1,
		Status:      0,
	}
	if err := model.DB.Create(&deliverable).Error; err != nil {
		serverError(c, err)
		return
	}

	// 更新节点为待验收
	model.DB.Model(&ms).Update("status", "submitted")
	ok(c, gin.H{"deliverable": deliverable})
}

type AcceptanceRequest struct {
	Accept  *bool  `json:"accept" binding:"required"`
	Comment string `json:"comment"`
}

// ConfirmAcceptance 甲方确认或驳回验收
func ConfirmAcceptance(c *gin.Context) {
	milestoneID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "节点ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var req AcceptanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	accept := *req.Accept

	var ms model.PaymentMilestone
	if err := model.DB.First(&ms, milestoneID).Error; err != nil {
		notFound(c, "节点不存在")
		return
	}
	if ms.Status != "submitted" {
		badRequest(c, "节点尚未提交交付物")
		return
	}
	// P0-05：仅项目甲方（订单甲方）可验收
	var order model.Order
	if err := model.DB.First(&order, ms.OrderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	var proj model.Project
	if err := model.DB.First(&proj, order.ProjectID).Error; err != nil || proj.UserID != userID {
		forbidden(c, "仅甲方可验收")
		return
	}

	now := time.Now()
	if accept {
		if err := model.DB.Model(&ms).Updates(map[string]interface{}{
			"status":      "accepted",
			"accepted_by": userID,
			"accepted_at": now,
		}).Error; err != nil {
			serverError(c, err)
			return
		}
		WriteAudit(c, "milestone.accept", "milestone", milestoneID, gin.H{"order_id": ms.OrderID, "accept": true})
		ok(c, gin.H{"message": "验收通过"})
		return
	}

	// 驳回：状态回到 pending，交付物标记驳回
	if err := model.DB.Model(&ms).Updates(map[string]interface{}{
		"status": "pending",
	}).Error; err != nil {
		serverError(c, err)
		return
	}
	model.DB.Model(&model.Deliverable{}).
		Where("order_id = ? AND milestone_id = ?", ms.OrderID, ms.ID).
		Update("status", 2)
	WriteAudit(c, "milestone.reject", "milestone", milestoneID, gin.H{"order_id": ms.OrderID, "accept": false, "comment": req.Comment})
	ok(c, gin.H{"message": "已驳回，请修改后重新提交"})
}
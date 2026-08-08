package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
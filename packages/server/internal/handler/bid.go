package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SubmitBidRequest struct {
	ProjectID      uint    `json:"project_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	ServiceDays    int     `json:"service_days"`
	ProposalFileID uint    `json:"proposal_file_id"`
}

// SubmitBid 服务方提交报价
func SubmitBid(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req SubmitBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	projectID := req.ProjectID

	var project model.Project
	if err := model.DB.First(&project, projectID).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	if project.Status != 1 {
		badRequest(c, "项目不在报价期")
		return
	}

	// 同一服务方不可重复报价
	var count int64
	model.DB.Model(&model.Bid{}).Where("project_id = ? AND supplier_id = ?", projectID, userID).Count(&count)
	if count > 0 {
		badRequest(c, "已报价，请勿重复提交")
		return
	}

	bid := model.Bid{
		ProjectID:      projectID,
		SupplierID:     userID,
		Amount:         req.Amount,
		ServiceDays:    req.ServiceDays,
		ProposalFileID: req.ProposalFileID,
		Status:         "submitted",
	}
	if err := model.DB.Create(&bid).Error; err != nil {
		serverError(c, err)
		return
	}

	WriteAudit(c, "bid.submit", "project", projectID, gin.H{"bid_id": bid.ID, "amount": bid.Amount})
	ok(c, gin.H{"bid": bid})
}

// ListBids 甲方查看报价；服务方仅查看脱敏排名
func ListBids(c *gin.Context) {
	projectID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	userType := c.GetInt("user_type")

	var project model.Project
	if err := model.DB.First(&project, projectID).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}

	var bids []model.Bid
	model.DB.Preload("Supplier").Where("project_id = ? AND status <> ?", projectID, "withdrawn").
		Order("amount ASC").Find(&bids)

	// 甲方可见全部；服务方仅见脱敏排名（隐藏金额与姓名）
	if userType == 2 {
		var masked []gin.H
		for i, b := range bids {
			masked = append(masked, gin.H{
				"rank":       i + 1,
				"service_days": b.ServiceDays,
				"anonymous":  true,
			})
		}
		ok(c, gin.H{"bids": masked})
		return
	}

	ok(c, gin.H{"bids": bids})
}

// ListMyBids 服务方"我的报价"列表（跨项目汇总自己提交的报价）
func ListMyBids(c *gin.Context) {
	userID := c.GetUint("user_id")
	// 仅服务方可查看自己的报价
	if c.GetInt("user_type") != 2 && !isAdmin(c) {
		forbidden(c, "仅服务方可查看报价记录")
		return
	}

	q := model.DB.Preload("Supplier").Where("supplier_id = ?", userID)
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	page, size := parsePage(c)
	var total int64
	q.Model(&model.Bid{}).Count(&total)

	var bids []model.Bid
	if err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&bids).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"bids": bids, "total": total, "page": page, "size": size})
}

// WithdrawBid 截止前撤回未中选报价
func WithdrawBid(c *gin.Context) {
	bidID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "报价ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var bid model.Bid
	if err := model.DB.First(&bid, bidID).Error; err != nil {
		notFound(c, "报价不存在")
		return
	}
	if bid.SupplierID != userID {
		forbidden(c, "无权限")
		return
	}
	if bid.Status == "selected" {
		badRequest(c, "中选报价不可撤回")
		return
	}

	model.DB.Model(&bid).Update("status", "withdrawn")
	ok(c, gin.H{"message": "已撤回"})
}

// SelectBid 甲方中选报价并生成待签约订单；同一项目仅可中选一次
func SelectBid(c *gin.Context) {
	bidID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "报价ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var bid model.Bid
	if err := model.DB.First(&bid, bidID).Error; err != nil {
		notFound(c, "报价不存在")
		return
	}
	var project model.Project
	if err := model.DB.First(&project, bid.ProjectID).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	if project.UserID != userID {
		forbidden(c, "无权限")
		return
	}
	if bid.Status != "submitted" {
		badRequest(c, "该报价不可中选")
		return
	}

	// 同项目仅可中选一次：检查是否已有待签约或进行中订单
	var count int64
	model.DB.Model(&model.Order{}).Where("project_id = ? AND status IN (0,1,2,3,4)", bid.ProjectID).Count(&count)
	if count > 0 {
		badRequest(c, "该项目已中选，不能重复选择")
		return
	}

	// 事务：更新报价为中选，生成待签约订单
	err = model.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Bid{}).Where("id = ?", bidID).Update("status", "selected").Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Bid{}).Where("project_id = ? AND id <> ? AND status = ?", bid.ProjectID, bidID, "submitted").
			Update("status", "rejected").Error; err != nil {
			return err
		}
		order := model.Order{
			ProjectID:     bid.ProjectID,
			SupplierID:    bid.SupplierID,
			SelectedBidID: bidID,
			Amount:        bid.Amount,
			Status:        0,
		}
		return tx.Create(&order).Error
	})
	if err != nil {
		serverError(c, err)
		return
	}

	// V10：接单通知——报价被中选后通知服务方（在 order 生成后）
	CreateNotification(bid.SupplierID, "报价已中选",
		"您在项目「"+project.Title+"」的报价已被甲方中选，请前往订单列表确认并签约。", "order")

	WriteAudit(c, "bid.select", "project", bid.ProjectID, gin.H{"bid_id": bidID, "supplier_id": bid.SupplierID, "amount": bid.Amount})
	ok(c, gin.H{"message": "中选成功，待签约"})
}
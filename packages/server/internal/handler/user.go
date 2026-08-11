package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// GetUserInfo 获取当前用户信息
func GetUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		notFound(c, "用户不存在")
		return
	}

	ok(c, gin.H{"user": user})
}

// UpdateUserInfo 更新用户资料
func UpdateUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input UserInfoRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		badRequest(c, "参数错误")
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if input.CompanyName != "" {
		updates["company_name"] = input.CompanyName
	}

	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		serverError(c, err)
		return
	}

	ok(c, gin.H{"message": "更新成功"})
}

// SubmitReview 订单完成双方互评，评分联动信用分
func SubmitReview(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		OrderID    uint   `json:"order_id" binding:"required"`
		RevieweeID uint   `json:"reviewee_id" binding:"required"`
		Rating     int    `json:"rating" binding:"required"`
		Content    string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		badRequest(c, "评分必须为1-5")
		return
	}

	var order model.Order
	if err := model.DB.First(&order, req.OrderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	if order.Status != 3 {
		badRequest(c, "订单尚未完成，不能评价")
		return
	}

	// 同一订单同向仅可评一次
	var count int64
	model.DB.Model(&model.Review{}).Where("order_id = ? AND reviewer_id = ? AND reviewee_id = ?",
		req.OrderID, userID, req.RevieweeID).Count(&count)
	if count > 0 {
		badRequest(c, "您已评价过该订单")
		return
	}

	review := model.Review{
		OrderID:    req.OrderID,
		ReviewerID: userID,
		RevieweeID: req.RevieweeID,
		Rating:     req.Rating,
		Content:    req.Content,
	}
	if err := model.DB.Create(&review).Error; err != nil {
		serverError(c, err)
		return
	}

	// 联动信用分：按评价/交付/纠纷既定权重重算（幂等）
	recalcUserCredit(req.RevieweeID)

	ok(c, gin.H{"review": review, "message": "评价成功"})
}

// GetUserReviews 查看某用户的评价
func GetUserReviews(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "用户ID无效")
		return
	}

	var reviews []model.Review
	model.DB.Where("reviewee_id = ?", userID).Order("created_at DESC").Find(&reviews)
	ok(c, gin.H{"reviews": reviews, "count": len(reviews)})
}

// ListProjectReviews 按项目查看评价（项目甲方/服务方/管理员可见）
func ListProjectReviews(c *gin.Context) {
	projectID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var project model.Project
	if err := model.DB.First(&project, projectID).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}
	// 权限：管理员、项目创建者、该项目订单服务方
	if !isAdmin(c) && project.UserID != userID {
		var cnt int64
		model.DB.Model(&model.Order{}).Where("project_id = ? AND supplier_id = ?", projectID, userID).Count(&cnt)
		if cnt == 0 {
			forbidden(c, "无权查看该项目评价")
			return
		}
	}

	// 找到项目关联订单的所有评价
	var orderIDs []uint
	model.DB.Model(&model.Order{}).Where("project_id = ?", projectID).Pluck("id", &orderIDs)
	var reviews []model.Review
	if len(orderIDs) > 0 {
		model.DB.Where("order_id IN ?", orderIDs).Order("created_at DESC").Find(&reviews)
	}
	ok(c, gin.H{"reviews": reviews, "count": len(reviews)})
}
package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// P0-05：对象级授权辅助函数

// isAdmin 当前用户是否为管理员（user_type=3）
func isAdmin(c *gin.Context) bool {
	return c.GetInt("user_type") == 3
}

// isOrderParticipant 判断用户是否为订单参与方（甲方=项目创建者 / 服务方=supplier）
func isOrderParticipant(userID uint, order *model.Order) bool {
	if order == nil {
		return false
	}
	// 服务方
	if order.SupplierID == userID {
		return true
	}
	// 甲方（项目创建者）
	var project model.Project
	if err := model.DB.First(&project, order.ProjectID).Error; err == nil && project.UserID == userID {
		return true
	}
	return false
}

// canAccessOrder 用户能否访问订单（管理员或参与方）
func canAccessOrder(c *gin.Context, order *model.Order) bool {
	if isAdmin(c) {
		return true
	}
	return isOrderParticipant(c.GetUint("user_id"), order)
}

// canAccessProject 用户能否访问项目（管理员或创建者）
func canAccessProject(c *gin.Context, project *model.Project) bool {
	if isAdmin(c) {
		return true
	}
	return project.UserID == c.GetUint("user_id")
}

// getOrderForUser 加载订单并校验访问权限；无权限返回 false
func getOrderForUser(c *gin.Context, orderID uint) (*model.Order, bool) {
	var order model.Order
	if err := model.DB.First(&order, orderID).Error; err != nil {
		return nil, false
	}
	if !canAccessOrder(c, &order) {
		return &order, false
	}
	return &order, true
}

// canAccessDispute 用户能否访问争议（管理员/订单参与方/指派专家）
func canAccessDispute(c *gin.Context, dispute *model.Dispute) bool {
	if isAdmin(c) {
		return true
	}
	userID := c.GetUint("user_id")
	// 发起人
	if dispute.InitiatorID == userID {
		return true
	}
	// 订单参与方
	var order model.Order
	if err := model.DB.First(&order, dispute.OrderID).Error; err == nil && isOrderParticipant(userID, &order) {
		return true
	}
	// 指派专家
	var cnt int64
	model.DB.Model(&model.DisputeExpertAssignment{}).
		Where("dispute_id = ? AND expert_user_id = ?", dispute.ID, userID).Count(&cnt)
	return cnt > 0
}

// getDisputeForUser 加载争议并校验权限
func getDisputeForUser(c *gin.Context, disputeID uint) (*model.Dispute, bool) {
	var dispute model.Dispute
	if err := model.DB.First(&dispute, disputeID).Error; err != nil {
		return nil, false
	}
	if !canAccessDispute(c, &dispute) {
		return &dispute, false
	}
	return &dispute, true
}

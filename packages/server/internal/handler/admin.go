package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListMyOrders 查询当前用户的订单（甲方=参与的/我发布的项目，服务方=承接的）
func ListMyOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	userType := c.GetInt("user_type")

	q := model.DB.Preload("Project").Preload("Supplier")
	if userType == 2 {
		q = q.Where("supplier_id = ?", userID)
	} else {
		q = q.Where("project_id IN (?)", model.DB.Model(&model.Project{}).Select("id").Where("user_id = ?", userID))
	}
	var orders []model.Order
	q.Order("created_at DESC").Find(&orders)
	ok(c, gin.H{"orders": orders, "count": len(orders)})
}

// AdminListDisputes 平台争议列表（含专家评审与调解状态）
func AdminListDisputes(c *gin.Context) {
	var disputes []model.Dispute
	model.DB.Order("created_at DESC").Find(&disputes)
	ok(c, gin.H{"disputes": disputes, "count": len(disputes)})
}

// ListDisputes 争议列表（按订单过滤，含证据与专家指派）
func ListDisputes(c *gin.Context) {
	orderID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "订单ID无效")
		return
	}

	var disputes []model.Dispute
	model.DB.Where("order_id = ?", orderID).Order("created_at DESC").Find(&disputes)
	ok(c, gin.H{"disputes": disputes, "count": len(disputes)})
}

// AdminPendingQualifications 待核验资质列表（平台）
func AdminListPendingQualifications(c *gin.Context) {
	var quals []model.SupplierQualification
	model.DB.Where("verification_status = ?", "pending").Order("created_at DESC").Find(&quals)
	ok(c, gin.H{"qualifications": quals, "count": len(quals)})
}

// AdminDashboardStats 后台总览统计
func AdminDashboardStats(c *gin.Context) {
	var userCount, projectCount, orderCount, disputeCount int64
	model.DB.Model(&model.User{}).Count(&userCount)
	model.DB.Model(&model.Project{}).Count(&projectCount)
	model.DB.Model(&model.Order{}).Count(&orderCount)
	model.DB.Model(&model.Dispute{}).Count(&disputeCount)

	var amount float64
	model.DB.Model(&model.PaymentTransaction{}).Where("type = ? AND status = ?", "settlement", 1).Select("COALESCE(SUM(amount),0)").Scan(&amount)

	ok(c, gin.H{
		"user_count":    userCount,
		"project_count": projectCount,
		"order_count":   orderCount,
		"dispute_count": disputeCount,
		"settled_amount": amount,
	})
}

// AdminListUsers 后台用户列表
func AdminListUsers(c *gin.Context) {
	var users []model.User
	model.DB.Order("created_at DESC").Find(&users)
	ok(c, gin.H{"users": users, "count": len(users)})
}

// AdminListOrders 后台全量订单列表
func AdminListOrders(c *gin.Context) {
	var orders []model.Order
	model.DB.Preload("Project").Preload("Supplier").Order("created_at DESC").Find(&orders)
	ok(c, gin.H{"orders": orders, "count": len(orders)})
}

// AdminListTransactions 后台全量资金流水
func AdminListTransactions(c *gin.Context) {
	var txns []model.PaymentTransaction
	model.DB.Order("created_at DESC").Find(&txns)
	ok(c, gin.H{"transactions": txns, "count": len(txns)})
}

// RequireAdmin 校验当前用户为管理员（user_type=3）
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetInt("user_type") != 3 {
			// 手动写403避免循环依赖（c.JSON 直接写）
			c.JSON(403, gin.H{"error": "forbidden", "message": "需要管理员权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
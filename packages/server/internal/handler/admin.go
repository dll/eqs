package handler

import (
	"time"

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

// AdminListDisputes 平台争议列表（支持状态筛选+分页）
// GET /api/v1/admin/disputes?status=&page=&size=
func AdminListDisputes(c *gin.Context) {
	query := model.DB.Model(&model.Dispute{})

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	page := 1
	size := 20
	if p := c.Query("page"); p != "" {
		if v, err := parseUint(p); err == nil && v > 0 {
			page = int(v)
		}
	}
	if s := c.Query("size"); s != "" {
		if v, err := parseUint(s); err == nil && v > 0 && v <= 100 {
			size = int(v)
		}
	}

	var disputes []model.Dispute
	query.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&disputes)
	ok(c, gin.H{"disputes": disputes, "count": len(disputes), "total": total, "page": page, "size": size})
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

// AdminOperationsStats 运营看板（V10：转化漏斗/状态分布/服务商活跃度）
// GET /api/v1/admin/operations-stats
func AdminOperationsStats(c *gin.Context) {
	// 用户结构
	var clientCount, supplierCount, expertCount int64
	model.DB.Model(&model.User{}).Where("user_type = ?", 1).Count(&clientCount)
	model.DB.Model(&model.User{}).Where("user_type = ?", 2).Count(&supplierCount)
	model.DB.Model(&model.User{}).Where("user_type = ?", 4).Count(&expertCount)

	// 项目状态分布（0草稿 1发布 2已接单 3进行中 4完成 5下架）
	type statusCount struct {
		Status int   `json:"status"`
		Count  int64 `json:"count"`
	}
	var projectDist []statusCount
	model.DB.Model(&model.Project{}).Select("status, COUNT(*) as count").Group("status").Scan(&projectDist)
	var orderDist []statusCount
	model.DB.Model(&model.Order{}).Select("status, COUNT(*) as count").Group("status").Scan(&orderDist)

	// 转化漏斗：发布项目 → 有报价 → 中选成单 → 已完成
	var published, withBid, completed int64
	model.DB.Model(&model.Project{}).Where("status >= ?", 1).Count(&published)
	model.DB.Model(&model.Bid{}).Where("status = ?", "selected").Distinct("project_id").Count(&withBid)
	model.DB.Model(&model.Order{}).Where("status = ?", 3).Count(&completed)

	// 服务商活跃度：近7天有报价/交付/打卡的服务商数
	since := time.Now().AddDate(0, 0, -7)
	var activeSuppliers int64
	model.DB.Model(&model.Bid{}).Where("created_at >= ?", since).Distinct("supplier_id").Count(&activeSuppliers)

	ok(c, gin.H{
		"users": gin.H{
			"clients": clientCount, "suppliers": supplierCount, "experts": expertCount,
		},
		"projects": projectDist,
		"orders":   orderDist,
		"funnel": gin.H{
			"published": published, "with_bid": withBid, "completed": completed,
		},
		"active_suppliers_7d": activeSuppliers,
	})
}

// AdminListUsers 后台用户列表（page/size 可选，默认全量）
func AdminListUsers(c *gin.Context) {
	var users []model.User
	q := model.DB.Select("id, phone, user_type, company_name, credit_score, status, created_at")
	// P2-03：传入 page/size 时启用分页，避免全表一次拉取
	if p, size := parsePage(c); c.Query("page") != "" {
		q = q.Offset((p - 1) * size).Limit(size)
	}
	q.Order("created_at DESC").Find(&users)
	// P1-09：手机号脱敏返回
	type userDTO struct {
		ID          uint    `json:"id"`
		Phone       string  `json:"phone"`
		UserType    int     `json:"user_type"`
		CompanyName string  `json:"company_name"`
		CreditScore float64 `json:"credit_score"`
		Status      int     `json:"status"`
		CreatedAt   string  `json:"created_at"`
	}
	dto := make([]userDTO, 0, len(users))
	for _, u := range users {
		dto = append(dto, userDTO{
			ID: u.ID, Phone: model.MaskPhone(u.Phone), UserType: u.UserType,
			CompanyName: u.CompanyName, CreditScore: u.CreditScore, Status: u.Status,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	ok(c, gin.H{"users": dto, "count": len(dto)})
}

// AdminGetUser 后台用户详情（含项目/订单/资质/评价统计）
func AdminGetUser(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "用户ID无效")
		return
	}

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		notFound(c, "用户不存在")
		return
	}

	var projectCount, orderAsOwner, orderAsSupplier, reviewCount, qualCount int64
	model.DB.Model(&model.Project{}).Where("user_id = ?", userID).Count(&projectCount)
	model.DB.Model(&model.Order{}).Where("supplier_id = ?", userID).Count(&orderAsSupplier)
	model.DB.Model(&model.Review{}).Where("reviewee_id = ?", userID).Count(&reviewCount)
	model.DB.Model(&model.SupplierQualification{}).Where("supplier_id = ?", userID).Count(&qualCount)
	// 甲方订单数：作为项目创建者的订单
	model.DB.Model(&model.Order{}).
		Joins("JOIN projects ON projects.id = orders.project_id").
		Where("projects.user_id = ?", userID).Count(&orderAsOwner)

	ok(c, gin.H{
		"user": gin.H{
			"id": user.ID, "phone": model.MaskPhone(user.Phone), "user_type": user.UserType,
			"company_name": user.CompanyName, "credit_score": user.CreditScore,
			"status": user.Status, "created_at": user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
		"stats": gin.H{
			"projects": projectCount, "orders_as_owner": orderAsOwner,
			"orders_as_supplier": orderAsSupplier, "reviews": reviewCount,
			"qualifications": qualCount,
		},
	})
}

// AdminUpdateUserStatus 后台启用/禁用用户
// PUT /api/v1/admin/users/:id/status  { "status": 1|0 }
func AdminUpdateUserStatus(c *gin.Context) {
	userID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "用户ID无效")
		return
	}
	// 防止管理员禁用自己
	operatorID := c.GetUint("user_id")
	if userID == operatorID {
		badRequest(c, "不能操作当前登录账号")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	if req.Status != 0 && req.Status != 1 {
		badRequest(c, "状态仅支持 0=禁用 1=启用")
		return
	}

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		notFound(c, "用户不存在")
		return
	}
	if err := model.DB.Model(&user).Update("status", req.Status).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "user.status", "user", userID, gin.H{"status": req.Status, "operator_id": operatorID})
	ok(c, gin.H{"message": "用户状态已更新", "status": req.Status})
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
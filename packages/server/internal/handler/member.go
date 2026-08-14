package handler

import (
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ==================== 会员体系（V10） ====================
// 等级权益（代码内定义，避免配置漂移）：
//   free   免费版：基础服务
//   silver 高级会员（¥99/月）：佣金 9.5 折、推荐加权 +5、专属标识
//   gold   企业会员（¥199/月）：佣金 9 折、推荐加权 +10、专属标识、优先派单
// 开通/续费采用模拟支付（订单生成即生效）；真实支付通道接入后改为走支付网关。

type memberLevelInfo struct {
	Level        string   `json:"level"`
	Name         string   `json:"name"`
	PricePerMonth float64 `json:"price_per_month"`
	Benefits     []string `json:"benefits"`
	CommissionDiscount float64 `json:"commission_discount"` // 佣金折扣系数（0.95=9.5折）
	RecommendBonus     float64 `json:"recommend_bonus"`     // 推荐综合评分加权
	Priority     bool     `json:"priority"`                 // 优先派单
}

var memberLevelDefs = map[string]memberLevelInfo{
	model.MemberLevelFree: {
		Level: model.MemberLevelFree, Name: "免费版", PricePerMonth: 0,
		Benefits: []string{"基础服务", "参与报价与接单"},
		CommissionDiscount: 1.0, RecommendBonus: 0,
	},
	model.MemberLevelSilver: {
		Level: model.MemberLevelSilver, Name: "高级会员", PricePerMonth: 99,
		Benefits: []string{"平台佣金 9.5 折", "智能派单推荐加权 +5", "会员专属标识"},
		CommissionDiscount: 0.95, RecommendBonus: 5,
	},
	model.MemberLevelGold: {
		Level: model.MemberLevelGold, Name: "企业会员", PricePerMonth: 199,
		Benefits: []string{"平台佣金 9 折", "智能派单推荐加权 +10", "会员专属标识", "优先派单"},
		CommissionDiscount: 0.9, RecommendBonus: 10, Priority: true,
	},
}

// memberLevelOf 返回用户当前生效的会员等级（过期视为 free）
func memberLevelOf(u *model.User) memberLevelInfo {
	lvl := model.MemberLevelFree
	if u.MemberLevel != "" && u.MemberLevel != model.MemberLevelFree {
		if u.MemberExpireAt != nil && u.MemberExpireAt.After(time.Now()) {
			lvl = u.MemberLevel
		}
	}
	def := memberLevelDefs[lvl]
	if def.Level == "" {
		return memberLevelDefs[model.MemberLevelFree]
	}
	return def
}

// ListMemberLevels 会员等级与权益
// GET /api/v1/member/levels
func ListMemberLevels(c *gin.Context) {
	levels := []memberLevelInfo{memberLevelDefs[model.MemberLevelFree], memberLevelDefs[model.MemberLevelSilver], memberLevelDefs[model.MemberLevelGold]}
	ok(c, gin.H{"levels": levels})
}

// GetMemberInfo 我的会员状态
// GET /api/v1/member/info
func GetMemberInfo(c *gin.Context) {
	var user model.User
	if err := model.DB.First(&user, c.GetUint("user_id")).Error; err != nil {
		notFound(c, "用户不存在")
		return
	}
	active := memberLevelOf(&user)
	ok(c, gin.H{
		"level":          user.MemberLevel,
		"level_name":     active.Name,
		"expire_at":      user.MemberExpireAt,
		"active":         user.MemberLevel != model.MemberLevelFree && user.MemberExpireAt != nil && user.MemberExpireAt.After(time.Now()),
		"benefits":       active.Benefits,
		"commission_discount": active.CommissionDiscount,
		"recommend_bonus":     active.RecommendBonus,
	})
}

type UpgradeMemberRequest struct {
	Level  string `json:"level" binding:"required"`
	Months int    `json:"months" binding:"required"`
}

// UpgradeMember 开通/续费会员（模拟支付即生效；记录订单供对账）
// POST /api/v1/member/upgrade
func UpgradeMember(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req UpgradeMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	def, found := memberLevelDefs[req.Level]
	if !found || req.Level == model.MemberLevelFree {
		badRequest(c, "会员等级无效")
		return
	}
	if req.Months < 1 || req.Months > 36 {
		badRequest(c, "开通月数需在 1-36 之间")
		return
	}
	amount := def.PricePerMonth * float64(req.Months)

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		notFound(c, "用户不存在")
		return
	}
	// 到期时间：未过期续费则累加，否则从当前时间起算
	base := time.Now()
	if user.MemberExpireAt != nil && user.MemberExpireAt.After(base) {
		base = *user.MemberExpireAt
	}
	expire := base.AddDate(0, req.Months, 0)

	// 模拟支付：直接生效并记录订单（真实支付通道接入后改为 pending + 回调生效）
	order := model.MembershipOrder{UserID: userID, Level: req.Level, Months: req.Months, Amount: amount, Status: "paid"}
	if err := model.DB.Create(&order).Error; err != nil {
		serverError(c, err)
		return
	}
	model.DB.Model(&user).Updates(map[string]interface{}{
		"member_level": req.Level, "member_expire_at": expire,
	})
	WriteAudit(c, "member.upgrade", "user", userID, gin.H{"level": req.Level, "months": req.Months, "amount": amount, "order_id": order.ID})
	ok(c, gin.H{"order_id": order.ID, "level": req.Level, "expire_at": expire, "amount": amount, "message": "会员开通成功（模拟支付）"})
}

// AdminListMembers 会员列表（平台）
// GET /api/v1/admin/members
func AdminListMembers(c *gin.Context) {
	page, size := parsePage(c)
	var total int64
	model.DB.Model(&model.User{}).Where("member_level <> ?", model.MemberLevelFree).Count(&total)
	var users []model.User
	model.DB.Where("member_level <> ?", model.MemberLevelFree).
		Order("member_expire_at DESC").Offset((page - 1) * size).Limit(size).Find(&users)
	ok(c, gin.H{"users": users, "total": total, "page": page, "size": size})
}

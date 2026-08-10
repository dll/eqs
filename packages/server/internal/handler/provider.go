package handler

import (
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListProviders 返回服务商列表（服务超市）
// GET /api/v1/provider/list?type=cost|supervision|geotech|design
// 公开接口：客户端以 silent401 调用，未登录也可浏览服务超市
func ListProviders(c *gin.Context) {
	db := model.DB
	if db == nil {
		serverError(c, nil)
		return
	}

	query := db.Model(&model.User{}).
		Where("user_type = ? AND status = ?", 2, 1)

	// 按服务类型筛选：通过已通过资质的服务方匹配
	if st := c.Query("type"); st != "" {
		sub := db.Model(&model.SupplierQualification{}).
			Select("DISTINCT supplier_id").
			Where("qualification_type = ? AND verification_status = ?", st, "approved")
		query = query.Where("id IN (?)", sub)
	}

	var providers []model.User
	if err := query.Order("credit_score DESC").Find(&providers).Error; err != nil {
		serverError(c, err)
		return
	}

	// 返回轻量结构
	type providerItem struct {
		ID          uint    `json:"id"`
		CompanyName string  `json:"company_name"`
		CreditScore float64 `json:"credit_score"`
	}
	items := make([]providerItem, 0, len(providers))
	for _, p := range providers {
		items = append(items, providerItem{
			ID:          p.ID,
			CompanyName: p.CompanyName,
			CreditScore: p.CreditScore,
		})
	}

	ok(c, gin.H{"providers": items})
}

// GetProvider 服务商详情
// GET /api/v1/provider/:id
func GetProvider(c *gin.Context) {
	db := model.DB
	if db == nil {
		serverError(c, nil)
		return
	}

	var provider model.User
	if err := db.First(&provider, c.Param("id")).Error; err != nil {
		notFound(c, "服务商不存在")
		return
	}
	if provider.UserType != 2 {
		notFound(c, "服务商不存在")
		return
	}

	// 附带已通过资质列表
	var quals []model.SupplierQualification
	db.Where("supplier_id = ? AND verification_status = ?", provider.ID, "approved").
		Find(&quals)

	type qualItem struct {
		QualificationType string `json:"qualification_type"`
		Level             string `json:"level"`
		Scope             string `json:"scope"`
	}
	qualItems := make([]qualItem, 0, len(quals))
	for _, q := range quals {
		qualItems = append(qualItems, qualItem{
			QualificationType: q.QualificationType,
			Level:             q.Level,
			Scope:             q.Scope,
		})
	}

	ok(c, gin.H{
		"provider": gin.H{
			"id":            provider.ID,
			"company_name":  provider.CompanyName,
			"credit_score":  provider.CreditScore,
			"qualifications": qualItems,
		},
	})
}

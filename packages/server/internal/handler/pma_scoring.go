package handler

import (
	"fmt"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"strconv"
)

func scoreAccess(c *gin.Context) bool { return requirePMAOpportunityAccess(c) }

// scoreForMember 同时校验评分存在、商机存在，并要求当前成员属于该商机的组织。
// PMAOpportunity.OwnerOrg 保存组织编码；不允许仅凭评分 ID 横向读取其他组织数据。
func scoreForMember(c *gin.Context, scoreID uint) (*model.PMAScore, bool) {
	var s model.PMAScore
	if model.DB.Preload("Items").First(&s, scoreID).Error != nil {
		notFound(c, "评分不存在")
		return nil, false
	}
	var op model.PMAOpportunity
	if model.DB.First(&op, s.OpportunityID).Error != nil {
		notFound(c, "商机不存在")
		return nil, false
	}
	var memberCount int64
	model.DB.Model(&model.OrgMember{}).Joins("JOIN organizations ON organizations.id = org_members.org_id").Where("org_members.user_id = ? AND org_members.status = 1 AND organizations.type = ? AND organizations.status = 1 AND organizations.code = ?", c.GetUint("user_id"), "internal", op.OwnerOrg).Count(&memberCount)
	if memberCount == 0 {
		forbidden(c, "无该商机评分权限")
		return nil, false
	}
	return &s, true
}
func scoreID(c *gin.Context) (uint, bool) {
	n, e := strconv.ParseUint(c.Param("id"), 10, 32)
	return uint(n), e == nil
}

func scoreItems() []model.PMAScoreItem {
	out := make([]model.PMAScoreItem, 0, 50)
	for _, c := range model.PMAScoreCategories {
		for n := 1; n <= 5; n++ {
			out = append(out, model.PMAScoreItem{Category: c.Name, CategoryKey: c.Key, Name: fmt.Sprintf("%s-%d", c.Name, n), Weight: c.Weight, Source: "manual"})
		}
	}
	return out
}

// CreatePMAScore 创建评分版本，版本只增不改，且不改变商机状态。
func CreatePMAScore(c *gin.Context) {
	if !scoreAccess(c) {
		return
	}
	oid, ok := scoreID(c)
	if !ok {
		badRequest(c, "商机ID无效")
		return
	}
	var op model.PMAOpportunity
	if model.DB.First(&op, oid).Error != nil {
		notFound(c, "商机不存在")
		return
	}
	var memberCount int64
	model.DB.Model(&model.OrgMember{}).Joins("JOIN organizations ON organizations.id = org_members.org_id").Where("org_members.user_id = ? AND org_members.status = 1 AND organizations.type = ? AND organizations.status = 1 AND organizations.code = ?", c.GetUint("user_id"), "internal", op.OwnerOrg).Count(&memberCount)
	if memberCount == 0 {
		forbidden(c, "无该商机评分权限")
		return
	}
	var max int
	model.DB.Model(&model.PMAScore{}).Where("opportunity_id = ?", oid).Select("COALESCE(MAX(version),0)").Scan(&max)
	s := model.PMAScore{OpportunityID: oid, Version: max + 1, CreatedBy: c.GetUint("user_id"), Items: scoreItems()}
	if err := model.DB.Create(&s).Error; err != nil {
		serverError(c, err)
		return
	}
	model.DB.Preload("Items").First(&s, s.ID)
	WriteAudit(c, "pma.score.create", "pma_score", s.ID, gin.H{"opportunity_id": oid, "version": s.Version})
	okJSON(c, s)
}

func GetPMAScore(c *gin.Context) {
	if !scoreAccess(c) {
		return
	}
	id, ok := scoreID(c)
	if !ok {
		badRequest(c, "评分ID无效")
		return
	}
	s, ok := scoreForMember(c, id)
	if !ok {
		return
	}
	okJSON(c, s)
}
func ListPMAScores(c *gin.Context) {
	if !scoreAccess(c) {
		return
	}
	oid, ok := scoreID(c)
	if !ok {
		badRequest(c, "商机ID无效")
		return
	}
	var op model.PMAOpportunity
	if model.DB.First(&op, oid).Error != nil {
		notFound(c, "商机不存在")
		return
	}
	var memberCount int64
	model.DB.Model(&model.OrgMember{}).Joins("JOIN organizations ON organizations.id = org_members.org_id").Where("org_members.user_id = ? AND org_members.status = 1 AND organizations.type = ? AND organizations.status = 1 AND organizations.code = ?", c.GetUint("user_id"), "internal", op.OwnerOrg).Count(&memberCount)
	if memberCount == 0 {
		forbidden(c, "无该商机评分权限")
		return
	}
	var ss []model.PMAScore
	model.DB.Where("opportunity_id = ?", oid).Preload("Items").Order("version DESC").Find(&ss)
	if ss == nil {
		ss = []model.PMAScore{}
	}
	okJSON(c, gin.H{"scores": ss})
}

func UpdatePMAScoreItem(c *gin.Context) {
	if !scoreAccess(c) {
		return
	}
	id, e1 := strconv.ParseUint(c.Param("id"), 10, 32)
	iid, e2 := strconv.ParseUint(c.Param("itemId"), 10, 32)
	if e1 != nil || e2 != nil {
		badRequest(c, "ID无效")
		return
	}
	var in struct {
		Score    *float64 `json:"score"`
		Evidence string   `json:"evidence"`
		Comment  string   `json:"comment"`
		Missing  string   `json:"missing"`
		Source   string   `json:"source"`
	}
	if c.ShouldBindJSON(&in) != nil {
		badRequest(c, "请求无效")
		return
	}
	if in.Score != nil && (*in.Score < 0 || *in.Score > 5) {
		badRequest(c, "score必须为0..5")
		return
	}
	if in.Source == "" {
		in.Source = "manual"
	}
	if in.Source != "manual" {
		badRequest(c, "逐项编辑仅允许manual")
		return
	}
	s, ok := scoreForMember(c, uint(id))
	if !ok {
		return
	}
	var item model.PMAScoreItem
	if model.DB.First(&item, iid).Error != nil || item.ScoreID != uint(id) {
		notFound(c, "评分项不存在")
		return
	}
	item.Score = in.Score
	item.Evidence = in.Evidence
	item.Comment = in.Comment
	item.Missing = in.Missing
	item.Source = in.Source
	model.DB.Save(&item)
	s.Blocked = false
	s.BlockedReason = ""
	s.ApplyGates()
	model.DB.Model(&s).Updates(map[string]interface{}{"total_score": s.TotalScore, "conclusion": s.Conclusion, "blocked": s.Blocked, "blocked_reason": s.BlockedReason})
	WriteAudit(c, "pma.score.item.update", "pma_score_item", item.ID, gin.H{"score_id": id, "score": item.Score, "source": "manual"})
	okJSON(c, item)
}

// SuggestPMAScoreAI 仅记录占位建议，不写入评分项最终分。
func SuggestPMAScoreAI(c *gin.Context) {
	if !scoreAccess(c) {
		return
	}
	id, ok := scoreID(c)
	if !ok {
		badRequest(c, "评分ID无效")
		return
	}
	if _, ok := scoreForMember(c, id); !ok {
		return
	}
	var in struct {
		Suggestions interface{} `json:"suggestions"`
		Confidence  *float64    `json:"confidence"`
		Gaps        string      `json:"gaps"`
		Evaluation  string      `json:"evaluation"`
	}
	if c.ShouldBindJSON(&in) != nil {
		badRequest(c, "请求无效")
		return
	}
	WriteAudit(c, "pma.score.ai.suggest", "pma_score", id, in)
	okJSON(c, gin.H{"score_id": id, "suggestions": in.Suggestions, "confidence": in.Confidence, "gaps": in.Gaps, "evaluation": in.Evaluation, "applied": false})
}

func okJSON(c *gin.Context, v interface{}) { c.JSON(200, gin.H{"data": v}) }

package handler

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

type CreateDisputeRequest struct {
	OrderID     uint   `json:"order_id" binding:"required"`
	MilestoneID uint   `json:"milestone_id"`
	Reason      string `json:"reason" binding:"required"`
	Claim       string `json:"claim"`
}

// CreateDispute 发起争议：先证据阶段，节点资金自动冻结
func CreateDispute(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CreateDisputeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var order model.Order
	if err := model.DB.Preload("Project").First(&order, req.OrderID).Error; err != nil {
		notFound(c, "订单不存在")
		return
	}
	if order.SupplierID != userID && order.Project.UserID != userID {
		forbidden(c, "仅合同双方可发起争议")
		return
	}

	// 同节点存在未结争议则拒绝
	var count int64
	q := model.DB.Model(&model.Dispute{}).Where("order_id = ? AND status <> ?", req.OrderID, "closed")
	if req.MilestoneID > 0 {
		q = q.Where("milestone_id = ?", req.MilestoneID)
	}
	q.Count(&count)
	if count > 0 {
		badRequest(c, "该订单已有未结争议")
		return
	}

	dispute := model.Dispute{
		OrderID:     req.OrderID,
		MilestoneID: req.MilestoneID,
		InitiatorID: userID,
		Reason:      req.Reason,
		Claim:       req.Claim,
		Status:      "evidence",
	}
	if err := model.DB.Create(&dispute).Error; err != nil {
		serverError(c, err)
		return
	}

	// 冻结对应节点：标记争议中（结算时检查 status<>closed 即拒绝）
	WriteAudit(c, "dispute.create", "dispute", dispute.ID, gin.H{"order_id": req.OrderID, "milestone_id": req.MilestoneID, "reason": req.Reason})
	ok(c, gin.H{"dispute": dispute, "message": "争议已发起，相关款项已冻结"})
}

// UpdateDispute 补充/更新争议信息（发起人可补充分歧点与诉求；证据阶段）
func UpdateDispute(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var req struct {
		Reason string `json:"reason"`
		Claim  string `json:"claim"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var dispute model.Dispute
	if err := model.DB.First(&dispute, disputeID).Error; err != nil {
		notFound(c, "争议不存在")
		return
	}
	if dispute.InitiatorID != userID && !isAdmin(c) {
		forbidden(c, "仅发起人可更新争议")
		return
	}
	if dispute.Status == "closed" {
		badRequest(c, "争议已结案，不可修改")
		return
	}

	updates := map[string]interface{}{}
	if req.Reason != "" {
		updates["reason"] = req.Reason
	}
	if req.Claim != "" {
		updates["claim"] = req.Claim
	}
	if len(updates) == 0 {
		badRequest(c, "没有可更新的字段")
		return
	}
	if err := model.DB.Model(&dispute).Updates(updates).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "dispute.update", "dispute", disputeID, gin.H{})
	ok(c, gin.H{"message": "争议信息已更新"})
}

// UploadDisputeEvidence 上传争议证据（文件+哈希）
func UploadDisputeEvidence(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var req struct {
		FileID  uint   `json:"file_id" binding:"required"`
		SHA256  string `json:"sha256"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	evidence := model.DisputeEvidence{
		DisputeID: disputeID,
		UserID:    userID,
		FileID:    req.FileID,
		SHA256:    req.SHA256,
		Content:   req.Content,
	}
	if err := model.DB.Create(&evidence).Error; err != nil {
		serverError(c, err)
		return
	}

	// 有证据后进入评审阶段
	model.DB.Model(&model.Dispute{}).Where("id = ? AND status = ?", disputeID, "evidence").
		Update("status", "review")
	ok(c, gin.H{"evidence": evidence})
}

// AutoAssignDisputeExperts 争议三专家自动评审（P1-08 增强）
// POST /api/v1/dispute/:id/auto-expert
// 平台自动从专家库随机选 3 名（排除已指派与利益冲突专家），创建指派并进入评审状态
func AutoAssignDisputeExperts(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}
	// 仅管理员可触发自动评审
	if !isAdmin(c) {
		forbidden(c, "仅管理员可指派专家")
		return
	}

	var dispute model.Dispute
	if err := model.DB.First(&dispute, disputeID).Error; err != nil {
		notFound(c, "争议不存在")
		return
	}
	if dispute.Status == "closed" {
		badRequest(c, "争议已结案，不可指派")
		return
	}

	// 已指派数量检查（每个争议最多 3 名专家）
	var existing int64
	model.DB.Model(&model.DisputeExpertAssignment{}).Where("dispute_id = ?", disputeID).Count(&existing)
	if existing >= 3 {
		badRequest(c, "该争议已完成专家指派（最多 3 名）")
		return
	}
	need := int(3 - existing)

	// 专家库：user_type=4 且状态正常；排除已指派专家
	sub := model.DB.Model(&model.DisputeExpertAssignment{}).Where("dispute_id = ?", disputeID).Select("expert_user_id")
	var experts []model.User
	model.DB.Where("user_type = ? AND status = ? AND id NOT IN (?)", 4, 1, sub).
		Order("credit_score DESC").
		Limit(need).
		Find(&experts)
	if len(experts) == 0 {
		badRequest(c, "暂无可用评审专家，请手动指派")
		return
	}

	created := 0
	for _, e := range experts {
		assignment := model.DisputeExpertAssignment{
			DisputeID:        disputeID,
			ExpertUserID:     e.ID,
			ConflictDeclared: 1, // 自动评审默认已披露（平台筛选排除冲突专家）
		}
		if err := model.DB.Create(&assignment).Error; err == nil {
			created++
			CreateNotification(e.ID, "争议评审任务", "您被指派参与争议 #"+uint2str(disputeID)+" 评审，请登录平台查看。", "dispute")
		}
	}
	if created == 0 {
		serverError(c, fmt.Errorf("专家指派失败"))
		return
	}

	// 更新争议状态为评审中
	model.DB.Model(&model.Dispute{}).Where("id = ?", disputeID).Update("status", "review")
	WriteAudit(c, "dispute.auto_expert", "dispute", disputeID, gin.H{"assigned": created, "total": int(existing) + created})
	ok(c, gin.H{"message": "已自动指派专家评审", "assigned": created, "total": int(existing) + created})
}

// AssignDisputeExpert 平台指派评审专家（管理员/平台）
func AssignDisputeExpert(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}

	var req struct {
		ExpertUserID    uint `json:"expert_user_id" binding:"required"`
		ConflictDeclared bool `json:"conflict_declared"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	// P1-08：仅管理员可指派专家；专家必须 user_type=4
	if !isAdmin(c) {
		forbidden(c, "仅管理员可指派专家")
		return
	}
	var expert model.User
	if err := model.DB.First(&expert, req.ExpertUserID).Error; err != nil || expert.UserType != 4 {
		badRequest(c, "专家不存在或非评审专家")
		return
	}
	// 利益冲突披露声明（模型字段为 int：0未披露/1已披露/2有冲突）
	conflict := 0
	if req.ConflictDeclared {
		conflict = 1
	}
	assignment := model.DisputeExpertAssignment{
		DisputeID:       disputeID,
		ExpertUserID:    req.ExpertUserID,
		ConflictDeclared: conflict,
	}
	if err := model.DB.Create(&assignment).Error; err != nil {
		serverError(c, err)
		return
	}
	// 更新争议状态为评审中
	model.DB.Model(&model.Dispute{}).Where("id = ?", disputeID).Update("status", "review")
	ok(c, gin.H{"assignment": assignment, "message": "专家已指派，利益冲突已披露"})
}

type SubmitExpertOpinionRequest struct {
	Opinion string `json:"opinion" binding:"required"`
	Vote    string `json:"vote" binding:"required"`
}

// SubmitExpertOpinion 专家提交评审意见
func SubmitExpertOpinion(c *gin.Context) {
	assignmentID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "指派ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var req SubmitExpertOpinionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var assignment model.DisputeExpertAssignment
	if err := model.DB.First(&assignment, assignmentID).Error; err != nil {
		notFound(c, "指派不存在")
		return
	}
	if assignment.ExpertUserID != userID {
		forbidden(c, "仅被指派专家可提交意见")
		return
	}

	now := time.Now()
	model.DB.Model(&assignment).Updates(map[string]interface{}{
		"opinion":      req.Opinion,
		"vote":         req.Vote,
		"submitted_at": now,
	})
	WriteAudit(c, "dispute.expert_opinion", "expert_assignment", assignmentID, gin.H{"dispute_id": assignment.DisputeID, "vote": req.Vote})
	ok(c, gin.H{"message": "评审意见已提交"})
}

// CloseDispute 争议结案：填写调解/评审结论，解冻资金并可按结论结算
func CloseDispute(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}

	var req struct {
		ResolutionType  string  `json:"resolution_type" binding:"required"` // settlement/agreement/award
		ResolutionFileID uint   `json:"resolution_file_id"`
		SettleAmount    float64 `json:"settle_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var dispute model.Dispute
	if err := model.DB.First(&dispute, disputeID).Error; err != nil {
		notFound(c, "争议不存在")
		return
	}
	// P0-05：仅参与方可提交证据
	if !canAccessDispute(c, &dispute) {
		forbidden(c, "无权操作该争议")
		return
	}
	if dispute.Status == "closed" {
		ok(c, gin.H{"message": "已结案", "idempotent": true})
		return
	}

	now := time.Now()
	model.DB.Model(&dispute).Updates(map[string]interface{}{
		"status":             "closed",
		"resolution_type":    req.ResolutionType,
		"resolution_file_id": req.ResolutionFileID,
		"closed_at":          now,
	})

	// 争议结案后重算发起方信用分（纠纷分加权，幂等）
	recalcUserCredit(dispute.InitiatorID)

	WriteAudit(c, "dispute.close", "dispute", disputeID, gin.H{"resolution_type": req.ResolutionType, "settle_amount": req.SettleAmount})
	ok(c, gin.H{"dispute": dispute, "message": "争议已结案，款项已解冻"})
}

// GetDispute 争议详情（含证据与专家意见）
func GetDispute(c *gin.Context) {
	disputeID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "争议ID无效")
		return
	}

	var dispute model.Dispute
	if err := model.DB.First(&dispute, disputeID).Error; err != nil {
		notFound(c, "争议不存在")
		return
	}
	// P0-05：仅参与方/专家/管理员可查看
	if !canAccessDispute(c, &dispute) {
		forbidden(c, "无权查看该争议")
		return
	}

	var evidence []model.DisputeEvidence
	model.DB.Where("dispute_id = ?", disputeID).Find(&evidence)

	var assignments []model.DisputeExpertAssignment
	model.DB.Where("dispute_id = ?", disputeID).Find(&assignments)

	ok(c, gin.H{
		"dispute":     dispute,
		"evidence":    evidence,
		"assignments": assignments,
		// P1-08：法定救济声明（平台专家评审≠法定仲裁，保留诉讼权利）
		"legal_remedy_notice": "本平台争议处理为专家评审与平台调解，不构成《仲裁法》意义上的法定仲裁；当事人保留依法向法院起诉或申请仲裁的权利。",
	})
}

// ListMyDisputes 我的争议列表（我发起或作为订单参与方）
// GET /api/v1/dispute/mine
func ListMyDisputes(c *gin.Context) {
	userID := c.GetUint("user_id")

	var disputes []model.Dispute
	// 我发起的，或我参与的订单（甲方=项目创建者 / 服务方=supplier）的争议
	model.DB.Where("initiator_id = ? OR order_id IN (?)", userID,
		model.DB.Model(&model.Order{}).
			Select("DISTINCT o.id").
			Table("orders o").
			Joins("JOIN projects p ON p.id = o.project_id").
			Where("o.supplier_id = ? OR p.user_id = ?", userID, userID),
	).Order("created_at DESC").Find(&disputes)

	ok(c, gin.H{"disputes": disputes, "count": len(disputes)})
}
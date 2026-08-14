package handler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eqs/server/internal/channel"
	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// ListQualifications 查询服务方资质列表
// 权限：仅服务方本人或管理员可查；支持 status 过滤（pending/approved/rejected/expired，缺省返回全部含已审核）
func ListQualifications(c *gin.Context) {
	supplierID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "服务方ID无效")
		return
	}

	// P0-05：对象级授权——仅本人或管理员可查看资质（含证书信息）
	userID := c.GetUint("user_id")
	if !isAdmin(c) && userID != supplierID {
		forbidden(c, "无权查看该服务方资质")
		return
	}

	query := model.DB.Where("supplier_id = ?", supplierID)
	if s := c.Query("status"); s != "" {
		query = query.Where("verification_status = ?", s)
	}
	var quals []model.SupplierQualification
	query.Order("created_at DESC").Find(&quals)
	ok(c, gin.H{"qualifications": quals, "count": len(quals)})
}

type SubmitQualificationRequest struct {
	QualificationType string `json:"qualification_type" binding:"required"`
	CertificateNo     string `json:"certificate_no" binding:"required"`
	Level             string `json:"level"`
	Scope             string `json:"scope"`
	// V6 必要字段
	IssuingAuthority string     `json:"issuing_authority"` // 发证机关
	IssueDate        *time.Time `json:"issue_date"`        // 签发日期
	ValidFrom        *time.Time `json:"valid_from"`
	ValidTo          *time.Time `json:"valid_to"`
	EvidenceFileID   uint       `json:"evidence_file_id"` // 扫描件附件（上传接口返回的 file_id）
}

// SubmitQualification 服务方提交资质用于核验
func SubmitQualification(c *gin.Context) {
	supplierID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "服务方ID无效")
		return
	}

	// P0-05：仅当前服务方可提交自己的资质
	if c.GetUint("user_id") != supplierID {
		forbidden(c, "仅可提交自己的资质")
		return
	}

	var req SubmitQualificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	// 有效期校验：有效期止不得早于签发日/当日（已过期资质直接拒绝，避免进入审核）
	if req.ValidTo != nil && req.ValidTo.Before(time.Now()) {
		badRequest(c, "资质已过有效期，请勿提交")
		return
	}
	if req.IssueDate != nil && req.ValidFrom != nil && req.ValidFrom.Before(*req.IssueDate) {
		badRequest(c, "有效期起始不得早于签发日期")
		return
	}
	if req.ValidFrom != nil && req.ValidTo != nil && req.ValidTo.Before(*req.ValidFrom) {
		badRequest(c, "有效期止不得早于有效期起始")
		return
	}

	// P1-09：证书号由 model.EncryptedString 透明加密存储
	qual := model.SupplierQualification{
		SupplierID:         supplierID,
		QualificationType:  req.QualificationType,
		CertificateNo:      model.EncryptedString(req.CertificateNo),
		Level:              req.Level,
		Scope:              req.Scope,
		IssuingAuthority:   req.IssuingAuthority,
		IssueDate:          req.IssueDate,
		ValidFrom:          req.ValidFrom,
		ValidTo:            req.ValidTo,
		EvidenceFileID:     req.EvidenceFileID,
		VerificationStatus: "pending",
	}
	if err := model.DB.Create(&qual).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"qualification": qual})
}

// ---- V6：扫描件上传（附件备份） ----

// allowedQualFileExt 允许的扫描件扩展名（图片/PDF）
var allowedQualFileExt = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".pdf": true,
}

// UploadQualificationFile 上传资质扫描件，保存到本地 uploads 目录并登记 project_files（附件备份）
// POST /api/v1/qualification/upload  (multipart/form-data: file)
func UploadQualificationFile(c *gin.Context) {
	userID := c.GetUint("user_id")

	fileHeader, err := c.FormFile("file")
	if err != nil {
		badRequest(c, "请选择要上传的扫描件")
		return
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedQualFileExt[ext] {
		badRequest(c, "仅支持 jpg/png/pdf 格式的扫描件")
		return
	}
	if fileHeader.Size > 10*1024*1024 {
		badRequest(c, "扫描件不能超过 10MB")
		return
	}

	// 保存到本地 uploads/{UPLOAD_DIR 配置}/{年月}/ 目录
	relDir := filepath.Join(config.Get().UploadDir, time.Now().Format("200601"))
	dir := filepath.Join(".", relDir)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		serverError(c, err)
		return
	}
	now := time.Now().Format("150405.000000000")
	name := filepath.Join(dir, now+"_"+strings.ReplaceAll(fileHeader.Filename, " ", "_"))
	if err := c.SaveUploadedFile(fileHeader, name); err != nil {
		serverError(c, err)
		return
	}

	// 登记为附件（项目文件表，作为备份可追溯）
	file := model.ProjectFile{
		UploaderID:   userID,
		OriginalName: fileHeader.Filename,
		FileType:     strings.TrimPrefix(ext, "."),
		StorageKey:   relDir + "/" + filepath.Base(name),
		Version:      1,
	}
	if err := model.DB.Create(&file).Error; err != nil {
		serverError(c, err)
		return
	}
	ok(c, gin.H{"file_id": file.ID, "original_name": file.OriginalName, "storage_key": file.StorageKey})
}

// GetQualification 资质详情（本人或管理员）
func GetQualification(c *gin.Context) {
	qualID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "资质ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var qual model.SupplierQualification
	if err := model.DB.First(&qual, qualID).Error; err != nil {
		notFound(c, "资质不存在")
		return
	}
	if !isAdmin(c) && qual.SupplierID != userID {
		forbidden(c, "无权查看该资质")
		return
	}
	ok(c, gin.H{"qualification": qual})
}

// UpdateQualification 编辑资质（仅本人；待审核/已驳回可改，已通过仅可改非关键信息）
func UpdateQualification(c *gin.Context) {
	qualID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "资质ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var qual model.SupplierQualification
	if err := model.DB.First(&qual, qualID).Error; err != nil {
		notFound(c, "资质不存在")
		return
	}
	if qual.SupplierID != userID && !isAdmin(c) {
		forbidden(c, "仅本人可编辑资质")
		return
	}
	if qual.VerificationStatus == "approved" {
		badRequest(c, "已审核通过的资质不可编辑，如需变更请重新提交")
		return
	}

	var req SubmitQualificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}
	// 有效期校验（同提交）
	if req.ValidTo != nil && req.ValidTo.Before(time.Now()) {
		badRequest(c, "资质已过有效期，请勿提交")
		return
	}

	updates := map[string]interface{}{
		"qualification_type":  req.QualificationType,
		"certificate_no":      string(model.EncryptedString(req.CertificateNo)),
		"level":               req.Level,
		"scope":               req.Scope,
		"issuing_authority":   req.IssuingAuthority,
		"issue_date":          req.IssueDate,
		"valid_from":          req.ValidFrom,
		"valid_to":            req.ValidTo,
		"evidence_file_id":    req.EvidenceFileID,
		"verification_status": "pending", // 修改后回到待审核
	}
	if err := model.DB.Model(&qual).Updates(updates).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "qualification.update", "qualification", qualID, gin.H{})
	ok(c, gin.H{"message": "资质已更新并重新进入审核"})
}

// DeleteQualification 删除资质（仅本人；仅待审核/已驳回可删，已通过不可删）
func DeleteQualification(c *gin.Context) {
	qualID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "资质ID无效")
		return
	}
	userID := c.GetUint("user_id")

	var qual model.SupplierQualification
	if err := model.DB.First(&qual, qualID).Error; err != nil {
		notFound(c, "资质不存在")
		return
	}
	if qual.SupplierID != userID && !isAdmin(c) {
		forbidden(c, "仅本人可删除资质")
		return
	}
	if qual.VerificationStatus == "approved" {
		badRequest(c, "已审核通过的资质不可删除")
		return
	}
	if err := model.DB.Delete(&qual).Error; err != nil {
		serverError(c, err)
		return
	}
	WriteAudit(c, "qualification.delete", "qualification", qualID, gin.H{})
	ok(c, gin.H{"message": "资质已删除"})
}

// ReviewQualification 平台核验资质（OCR/人工），结果更新核验状态
// V6：增加自动通过条件校验 + AI 辅助审核建议
func ReviewQualification(c *gin.Context) {
	// P0-05：仅管理员可审核资质
	if c.GetInt("user_type") != 3 {
		forbidden(c, "仅管理员可审核资质")
		return
	}
	qualID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "资质ID无效")
		return
	}
	reviewerID := c.GetUint("user_id")

	var req struct {
		Verified *bool  `json:"verified" binding:"required"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	var qual model.SupplierQualification
	if err := model.DB.First(&qual, qualID).Error; err != nil {
		notFound(c, "资质不存在")
		return
	}
	cfg := config.Get()

	// V6：自动通过条件校验（信息完整性 + 有效期）——通过必须满足硬性条件；
	// issues 同时供 AI 辅助审核做规则风险评估（仅算一次）
	gateIssues := checkQualificationGate(&qual)
	if *req.Verified && len(gateIssues) > 0 {
		badRequest(c, "不满足通过条件："+strings.Join(gateIssues, "；"))
		return
	}

	// V6：AI 辅助审核——对资质信息做自动风险评估，生成建议（未配置 AI 时按规则降级）
	aiSuggestion := aiAssistQualificationReview(&qual, gateIssues)
	// V10：OCR 识别扫描件（腾讯云 OCR；未配置凭据时跳过，不影响审核流程）
	ocrText := ocrAssistQualification(c, &qual, cfg)

	status := "rejected"
	if *req.Verified {
		status = "approved"
	}
	comment := strings.TrimSpace(req.Comment)
	prefix := ""
	if ocrText != "" {
		prefix += "OCR识别:" + ocrText + "。"
	}
	if aiSuggestion.Suggestion != "" {
		prefix += "AI建议:" + aiSuggestion.Suggestion + "。"
	}
	if prefix != "" {
		comment = prefix + "人工备注:" + comment
	}
	now := time.Now()
	model.DB.Model(&qual).Updates(map[string]interface{}{
		"verification_status": status,
		"verification_method": "ai",
		"review_comment":      comment,
		"reviewed_by":         reviewerID,
		"reviewed_at":         now,
	})
	WriteAudit(c, "qualification.review", "qualification", qualID, gin.H{"status": status, "reviewer_id": reviewerID, "ai_risk": aiSuggestion.Risk})
	ok(c, gin.H{"qualification": qual, "status": status, "ai_suggestion": aiSuggestion})
}

// checkQualificationGate 校验资质是否满足通过条件，返回未满足项列表
// 通过条件：证书号非空、发证机关非空、有效期未过期、附件扫描件已上传
func checkQualificationGate(q *model.SupplierQualification) []string {
	var issues []string
	if q.CertificateNo == "" {
		issues = append(issues, "证书编号为空")
	}
	if strings.TrimSpace(q.IssuingAuthority) == "" {
		issues = append(issues, "发证机关为空")
	}
	if q.ValidTo == nil || q.ValidTo.Before(time.Now()) {
		issues = append(issues, "资质已过有效期或未填写有效期")
	}
	if q.EvidenceFileID == 0 {
		issues = append(issues, "未上传扫描件附件")
	}
	return issues
}

// qualificationAISuggestion AI 辅助审核结果
type qualificationAISuggestion struct {
	Risk        string `json:"risk"`         // low/medium/high
	Suggestion  string `json:"suggestion"`   // 审核建议文案
	GeneratedBy string `json:"generated_by"` // ai/rules
}

// aiAssistQualificationReview AI 辅助审核：配置了大模型密钥时调用 LLM 评估，
// 否则基于规则（有效期/信息完整度）给出风险建议。issues 复用调用方已算好的通过条件检查结果。
func aiAssistQualificationReview(q *model.SupplierQualification, issues []string) qualificationAISuggestion {
	// 规则基线
	risk := "low"
	if len(issues) > 0 {
		risk = "high"
	}

	ai := loadAIProvider()
	if ai.APIKey == "" || ai.BaseURL == "" {
		suggestion := "规则校验通过，可人工复核后通过"
		if risk == "high" {
			suggestion = "规则校验未通过：" + strings.Join(issues, "；")
		}
		return qualificationAISuggestion{Risk: risk, Suggestion: suggestion, GeneratedBy: "rules"}
	}

	prompt := "你是建筑行业资质审核助手。请审核以下服务方资质信息，判断是否存在风险，只输出JSON：{\"risk\":\"low|medium|high\",\"suggestion\":\"审核建议\"}。资质类型：" + q.QualificationType +
		"；等级：" + q.Level + "；范围：" + q.Scope + "；发证机关：" + q.IssuingAuthority +
		"；签发日期：" + fmtDateStr(q.IssueDate) + "；有效期至：" + fmtDateStr(q.ValidTo) +
		"；证书编号：" + string(q.CertificateNo)
	if text, err := callAIModel(ai, prompt); err == nil && text != "" {
		// 尝试解析 JSON；失败则用原文作为建议
		var parsed qualificationAISuggestion
		if json.Unmarshal([]byte(text), &parsed) == nil && parsed.Risk != "" {
			parsed.GeneratedBy = "ai"
			return parsed
		}
		return qualificationAISuggestion{Risk: risk, Suggestion: strings.TrimSpace(text), GeneratedBy: "ai"}
	}
	return qualificationAISuggestion{Risk: risk, Suggestion: "AI 调用失败，请人工复核", GeneratedBy: "ai"}
}

// ocrAssistQualification V10：腾讯云 OCR 识别资质扫描件，返回识别文本（失败/未配置返回空串，不影响审核）
func ocrAssistQualification(c *gin.Context, qual *model.SupplierQualification, cfg *config.Config) string {
	if cfg.TencentOCRSecretID == "" || cfg.TencentOCRSecretKey == "" || qual.EvidenceFileID == 0 {
		return ""
	}
	var f model.ProjectFile
	if err := model.DB.First(&f, qual.EvidenceFileID).Error; err != nil {
		return ""
	}
	if len(f.StorageKey) > 7 && (f.StorageKey[:7] == "http://" || f.StorageKey[:8] == "https://") {
		return "" // 外部 URL 文件不本地读取
	}
	path := f.StorageKey
	if len(path) > 0 && path[0] != '/' {
		path = "./" + path
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	client := channel.NewOcrClient(cfg.TencentOCRSecretID, cfg.TencentOCRSecretKey)
	if client == nil {
		return ""
	}
	lines, err := client.RecognizeText(data)
	if err != nil || len(lines) == 0 {
		return ""
	}
	text := strings.Join(lines, "；")
	if len(text) > 200 {
		text = text[:200] + "…"
	}
	return text
}

package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// AIProvider 大模型配置（从环境变量读取）
type AIProvider struct {
	APIKey   string
	BaseURL  string
	Model    string
	Endpoint string
}

func loadAIProvider() *AIProvider {
	return &AIProvider{
		APIKey:   getEnvStr("ZHIPU_API_KEY"),
		BaseURL:  getEnvStr("ZHIPU_BASE_URL"),
		Model:    getEnvStr("ZHIPU_MODEL"),
		Endpoint: getEnvStr("AI_ANALYZE_ENDPOINT"),
	}
}

func getEnvStr(key string) string {
	return os.Getenv(key)
}

// AIAnalysis 分析结果
type AIAnalysis struct {
	Summary     string   `json:"summary"`
	RiskLevel   string   `json:"risk_level"` // low/medium/high
	Progress    int      `json:"progress"`
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
	GeneratedBy string   `json:"generated_by"` // ai / rules
}

// ProjectAnalysisItem 单项目分析
type ProjectAnalysisItem struct {
	ProjectID uint        `json:"project_id"`
	Title     string      `json:"title"`
	Analysis  AIAnalysis  `json:"analysis"`
}

// AIAnalyzeAllProjects 全量项目 AI 分析
// POST /api/v1/admin/ai/project-analysis
func AIAnalyzeAllProjects(c *gin.Context) {
	var projects []model.Project
	model.DB.Order("created_at DESC").Limit(20).Find(&projects)

	items := make([]ProjectAnalysisItem, 0, len(projects))
	for _, p := range projects {
		items = append(items, ProjectAnalysisItem{
			ProjectID: p.ID,
			Title:     p.Title,
			Analysis:  analyzeProjectByRules(p),
		})
	}

	// 尝试 AI 增强摘要
	ai := loadAIProvider()
	summary := ""
	source := "rules"
	if ai.APIKey != "" {
		if s, err := aiSummarizeProjects(ai, items); err == nil && s != "" {
			summary = s
			source = "ai"
		}
	}

	ok(c, gin.H{
		"items":        items,
		"summary":      summary,
		"generated_by": source,
		"total":        len(items),
	})
}

// AIAnalyzeProject 单项目问题解析
// POST /api/v1/project/:id/ai-analysis
func AIAnalyzeProject(c *gin.Context) {
	projectID, err := parseUint(c.Param("id"))
	if err != nil {
		badRequest(c, "项目ID无效")
		return
	}
	var project model.Project
	if err := model.DB.First(&project, projectID).Error; err != nil {
		notFound(c, "项目不存在")
		return
	}

	analysis := analyzeProjectByRules(project)

	ai := loadAIProvider()
	if ai.APIKey != "" {
		if s, err := aiAnalyzeOne(ai, project, analysis); err == nil && s != "" {
			analysis.Summary = s
			analysis.GeneratedBy = "ai"
		}
	}

	ok(c, gin.H{"project_id": project.ID, "title": project.Title, "analysis": analysis})
}

// analyzeProjectByRules 基于真实数据的规则分析
func analyzeProjectByRules(p model.Project) AIAnalysis {
	a := AIAnalysis{GeneratedBy: "rules", RiskLevel: "low", Issues: []string{}, Suggestions: []string{}}

	// 进度
	var orders []model.Order
	model.DB.Where("project_id = ?", p.ID).Find(&orders)
	totalMs, doneMs := 0, 0
	var msCount int64
	for _, o := range orders {
		var ms []model.PaymentMilestone
		model.DB.Where("order_id = ?", o.ID).Find(&ms)
		for _, m := range ms {
			msCount++
			totalMs++
			if m.Status == "settled" || m.Status == "accepted" {
				doneMs++
			}
		}
	}
	if totalMs > 0 {
		a.Progress = doneMs * 100 / totalMs
	} else {
		switch p.Status {
		case 4:
			a.Progress = 100
		case 3:
			a.Progress = 80
		case 2:
			a.Progress = 50
		case 1:
			a.Progress = 10
		default:
			a.Progress = 0
		}
	}

	// 风险：逾期
	if p.Deadline != nil && time.Now().After(*p.Deadline) && p.Status != 4 {
		a.RiskLevel = "high"
		a.Issues = append(a.Issues, "项目已超过截止日期尚未完成")
	}
	// 争议
	var disputeCount int64
	model.DB.Model(&model.Dispute{}).Where("project_id IN (?)",
		model.DB.Model(&model.Order{}).Select("id").Where("project_id = ?", p.ID),
	).Count(&disputeCount)
	if disputeCount > 0 {
		if a.RiskLevel != "high" {
			a.RiskLevel = "medium"
		}
		a.Issues = append(a.Issues, fmt.Sprintf("存在 %d 起争议", disputeCount))
	}
	// 里程碑停滞
	if msCount > 0 && doneMs == 0 && p.Status >= 2 {
		a.Issues = append(a.Issues, "里程碑无已结算节点，进度可能停滞")
	}
	// 无订单
	if len(orders) == 0 && p.Status >= 1 {
		a.Issues = append(a.Issues, "项目已发布但暂无订单/报价")
		a.Suggestions = append(a.Suggestions, "建议推送推荐服务方或扩大发布范围")
	}

	a.Summary = fmt.Sprintf("项目「%s」当前进度 %d%%，共 %d 个订单、%d 个里程碑，风险等级：%s",
		p.Title, a.Progress, len(orders), msCount, a.RiskLevel)
	if len(a.Issues) == 0 {
		a.Suggestions = append(a.Suggestions, "项目运行正常，建议按节点推进验收与结算")
	}
	return a
}

// ---- AI 大模型调用（有 key 时启用，否则降级规则分析） ----

func aiSummarizeProjects(ai *AIProvider, items []ProjectAnalysisItem) (string, error) {
	if ai.APIKey == "" || ai.BaseURL == "" {
		return "", fmt.Errorf("AI 未配置")
	}
	// 构造精简 prompt
	var b bytes.Buffer
	b.WriteString("请对以下工程服务项目列表进行总体分析（中文，100字内），指出整体进度与风险：\n")
	for _, it := range items {
		b.WriteString(fmt.Sprintf("- %s: 进度%d%%, 风险%s\n", it.Title, it.Analysis.Progress, it.Analysis.RiskLevel))
	}
	return callAIModel(ai, b.String())
}

func aiAnalyzeOne(ai *AIProvider, p model.Project, a AIAnalysis) (string, error) {
	if ai.APIKey == "" || ai.BaseURL == "" {
		return "", fmt.Errorf("AI 未配置")
	}
	prompt := fmt.Sprintf("工程服务项目「%s」进度%d%%，风险%s，问题：%v。请给出简要优化建议（中文，80字内）。",
		p.Title, a.Progress, a.RiskLevel, a.Issues)
	return callAIModel(ai, prompt)
}

// callAIModel 通用大模型调用（智谱 GLM 兼容 OpenAI 格式）
func callAIModel(ai *AIProvider, prompt string) (string, error) {
	endpoint := ai.Endpoint
	if endpoint == "" {
		endpoint = ai.BaseURL
	}
	if endpoint == "" {
		endpoint = "https://open.bigmodel.cn/api/paas/v4/chat/completions"
	}
	payload := map[string]interface{}{
		"model":    ai.Model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ai.APIKey)
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("AI API %d", resp.StatusCode)
	}
	var r struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(data, &r); err != nil || len(r.Choices) == 0 {
		return "", fmt.Errorf("AI 响应解析失败")
	}
	return r.Choices[0].Message.Content, nil
}

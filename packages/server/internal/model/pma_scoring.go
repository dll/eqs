package model

import "time"

// PMAScore 是独立于交易域的商机评分版本。
type PMAScore struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	OpportunityID uint           `json:"opportunity_id" gorm:"index;not null"`
	Version       int            `json:"version" gorm:"not null"`
	TotalScore    float64        `json:"total_score"`
	Conclusion    string         `json:"conclusion" gorm:"size:40"`
	AIConfidence  *float64       `json:"ai_confidence"`
	Gaps          string         `json:"gaps" gorm:"type:text"`
	Evaluation    string         `json:"evaluation" gorm:"type:text"`
	HardGates     string         `json:"hard_gates" gorm:"type:text"`
	BlockedReason string         `json:"blocked_reason" gorm:"type:text"`
	Blocked       bool           `json:"blocked"`
	CreatedBy     uint           `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Items         []PMAScoreItem `json:"items" gorm:"foreignKey:ScoreID;constraint:OnDelete:CASCADE"`
}

// PMAScoreItem 为十类五项的可审计评分明细；Score=nil 表示 unknown。
type PMAScoreItem struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ScoreID     uint      `json:"score_id" gorm:"index;not null"`
	Category    string    `json:"category" gorm:"size:40;not null"`
	CategoryKey string    `json:"category_key" gorm:"size:40;not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Score       *float64  `json:"score"`
	Weight      float64   `json:"weight"`
	Evidence    string    `json:"evidence" gorm:"type:text"`
	Comment     string    `json:"comment" gorm:"type:text"`
	Source      string    `json:"source" gorm:"size:10;not null"`
	Missing     string    `json:"missing" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var PMAScoreCategories = []struct {
	Key, Name string
	Weight    float64
}{
	{"requirements", "需求验收", 15}, {"technical_match", "技术匹配", 15}, {"source_credibility", "客户来源可信", 10},
	{"scope_control", "范围可控", 10}, {"schedule", "周期可行", 10}, {"resources", "资源可得", 10},
	{"business_value", "商业价值", 10}, {"compliance_security", "合规安全", 10}, {"budget_return", "预算收益", 5}, {"delivery_risk", "交付风险", 5},
}

func (s *PMAScore) Recalculate() {
	s.TotalScore = 0
	for _, c := range PMAScoreCategories {
		var sum float64
		var n int
		for _, i := range s.Items {
			if i.CategoryKey == c.Key && i.Score != nil {
				sum += *i.Score
				n++
			}
		}
		if n == 5 {
			s.TotalScore += (sum / 5) * c.Weight / 5
		}
	}
	if s.Blocked {
		s.Conclusion = "blocked"
	} else if s.TotalScore >= 80 {
		s.Conclusion = "建议进入OPC定案候选"
	} else if s.TotalScore >= 60 {
		s.Conclusion = "补充信息"
	} else {
		s.Conclusion = "不建议"
	}
}

func (s *PMAScore) ApplyGates() {
	if s.HardGates != "" {
		s.Blocked = true
		if s.BlockedReason == "" {
			s.BlockedReason = s.HardGates
		}
	}
	for _, i := range s.Items {
		if i.Score == nil && (i.CategoryKey == "requirements" || i.CategoryKey == "technical_match" || i.CategoryKey == "compliance_security" || i.CategoryKey == "schedule" || i.CategoryKey == "resources") {
			s.Blocked = true
			if s.BlockedReason == "" {
				s.BlockedReason = "关键项存在unknown"
			}
			break
		}
	}
	s.Recalculate()
}

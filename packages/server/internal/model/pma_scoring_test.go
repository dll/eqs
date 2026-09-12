package model

import "testing"

func TestPMAScoreTenByFiveWeightedCalculation(t *testing.T) {
	s := PMAScore{}
	for _, c := range PMAScoreCategories {
		for n := 0; n < 5; n++ {
			v := float64((n % 5) + 1)
			s.Items = append(s.Items, PMAScoreItem{CategoryKey: c.Key, Score: &v})
		}
	}
	s.ApplyGates()
	if s.TotalScore != 60 {
		t.Fatalf("expected weighted score 60, got %v", s.TotalScore)
	}
	if len(s.Items) != 50 {
		t.Fatalf("expected 50 items, got %d", len(s.Items))
	}
}

func TestPMAScoreCriticalUnknownAndHardGate(t *testing.T) {
	for _, key := range []string{"requirements", "technical_match", "compliance_security", "schedule", "resources"} {
		s := PMAScore{Items: []PMAScoreItem{{CategoryKey: key}}}
		s.ApplyGates()
		if !s.Blocked {
			t.Fatalf("critical unknown %s must block", key)
		}
	}
	s := PMAScore{HardGates: "违法/侵权高风险", Items: make([]PMAScoreItem, 0)}
	s.ApplyGates()
	if !s.Blocked || s.BlockedReason != "违法/侵权高风险" {
		t.Fatalf("hard gate must block with reason, got blocked=%v reason=%q", s.Blocked, s.BlockedReason)
	}
}

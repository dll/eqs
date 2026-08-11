package handler

import (
	"net/http"
	"testing"

	"github.com/eqs/server/internal/model"
)

// 验证 demo seed 在 SQLite 上生成的完整数据量
func TestDemo_SeedDataCount(t *testing.T) {
	r := setupDemoRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/admin/demo/seed?mode=demo", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("seed failed: %d %s", w.Code, w.Body.String())
	}
	var users, projects, orders, bids, milestones int64
	model.DB.Model(&model.User{}).Count(&users)
	model.DB.Model(&model.Project{}).Count(&projects)
	model.DB.Model(&model.Order{}).Count(&orders)
	model.DB.Model(&model.Bid{}).Count(&bids)
	model.DB.Model(&model.PaymentMilestone{}).Count(&milestones)
	t.Logf("users=%d projects=%d orders=%d bids=%d milestones=%d", users, projects, orders, bids, milestones)
	if projects != 5 {
		t.Errorf("projects 应为5，得到 %d", projects)
	}
	if orders != 3 {
		t.Errorf("orders 应为3，得到 %d", orders)
	}
	if bids < 3 {
		t.Errorf("bids 应>=3，得到 %d", bids)
	}
	if milestones != 9 {
		t.Errorf("milestones 应为9，得到 %d", milestones)
	}
	// 校验外键完整：所有 bid.project_id 存在
	var orphanBids int64
	model.DB.Model(&model.Bid{}).Where("project_id NOT IN (SELECT id FROM projects)").Count(&orphanBids)
	if orphanBids > 0 {
		t.Errorf("存在 %d 条孤儿 bid（project_id 无效）", orphanBids)
	}
}

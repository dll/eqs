package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupV10Router 会员体系测试路由
func setupV10Router() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	r := gin.New()
	api := r.Group("/api/v1")
	client := api.Group("")
	client.Use(AuthTestMiddleware(1, 2)) // 服务方
	{
		client.GET("/member/levels", ListMemberLevels)
		client.GET("/member/info", GetMemberInfo)
		client.POST("/member/upgrade", UpgradeMember)
	}
	return r
}

// TestMemberFlow 会员体系：等级列表 → 开通 → 状态生效 → 过期判定
func TestMemberFlow(t *testing.T) {
	r := setupV10Router()
	createTestUser(t, "13910000010", 2)

	// 等级列表
	w := doJSON(t, r, "GET", "/api/v1/member/levels", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("等级列表应200，得到 %d", w.Code)
	}
	var lvResp struct {
		Levels []memberLevelInfo `json:"levels"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &lvResp)
	if len(lvResp.Levels) != 3 {
		t.Fatalf("等级数 = %d, 期望 3", len(lvResp.Levels))
	}

	// 开通金牌会员 12 个月
	w = doJSON(t, r, "POST", "/api/v1/member/upgrade", map[string]interface{}{"level": "gold", "months": 12})
	if w.Code != http.StatusOK {
		t.Fatalf("开通会员应200，得到 %d: %s", w.Code, w.Body.String())
	}

	// 我的会员状态
	w = doJSON(t, r, "GET", "/api/v1/member/info", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("会员状态应200，得到 %d", w.Code)
	}
	var info struct {
		Level  string  `json:"level"`
		Active bool    `json:"active"`
		Discount float64 `json:"commission_discount"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &info)
	if info.Level != "gold" || !info.Active {
		t.Fatalf("会员状态异常: %+v", info)
	}
	if info.Discount != 0.9 {
		t.Fatalf("金牌佣金折扣 = %v, 期望 0.9", info.Discount)
	}

	// 非法等级
	w = doJSON(t, r, "POST", "/api/v1/member/upgrade", map[string]interface{}{"level": "platinum", "months": 1})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法等级应400，得到 %d", w.Code)
	}

	// 过期判定：直接改库为已过期
	model.DB.Model(&model.User{}).Where("phone = ?", "13910000010").
		Updates(map[string]interface{}{"member_expire_at": time.Now().Add(-time.Hour)})
	w = doJSON(t, r, "GET", "/api/v1/member/info", nil)
	var info2 struct {
		Active bool `json:"active"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &info2)
	if info2.Active {
		t.Fatal("过期会员应判定为未生效")
	}
}

// TestMemberCommissionDiscount 会员佣金折扣：金牌 9 折
func TestMemberCommissionDiscount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	uid := createTestUser(t, "13910000011", 2)
	model.DB.Model(&model.User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"member_level": "gold", "member_expire_at": time.Now().Add(30 * 24 * time.Hour),
	})
	model.DB.Create(&model.SystemConfig{ConfigKey: "commission.rate", ConfigValue: "8", ValueType: "string", Description: "佣金费率"})
	rate := effectiveCommissionRate(uid)
	if rate != 7.2 { // 8 * 0.9
		t.Fatalf("金牌佣金费率 = %v, 期望 7.2", rate)
	}
}

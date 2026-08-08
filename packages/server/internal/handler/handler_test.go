package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")
	api.POST("/auth/login", PhoneLogin)
	api.POST("/auth/wechat-login", WxLogin)
	api.POST("/project/create", CreateProject)
	api.GET("/project/list", ListProjects)
	api.GET("/project/:id", GetProject)
	api.POST("/bid/submit", SubmitBid)
	api.POST("/bid/:id/select", SelectBid)
	api.GET("/project/:id/bids", ListBids)
	return r
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf.Write(b)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// createTestUser 直接落库创建测试用户并返回其ID
func createTestUser(t *testing.T, phone string, userType int) uint {
	t.Helper()
	user := model.User{Phone: phone, UserType: userType, Status: 1, CreditScore: 100}
	if err := model.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	return user.ID
}

func TestPhoneLogin_InvalidCode(t *testing.T) {
	r := setupTestRouter(t)

	w := doJSON(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13800138000", "code": "000000", "user_type": 1,
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望400，得到 %d: %s", w.Code, w.Body.String())
	}
}

func TestPhoneLogin_Success(t *testing.T) {
	r := setupTestRouter(t)

	w := doJSON(t, r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"phone": "13800138001", "code": "123456", "user_type": 1,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Token string     `json:"token"`
		User  model.User `json:"user"`
		IsNew bool       `json:"isNew"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("未返回 token")
	}
	if !resp.IsNew {
		t.Fatal("首登应标记为新用户")
	}
}

func TestCreateProject_MissingField(t *testing.T) {
	r := setupTestRouter(t)

	w := doJSON(t, r, "POST", "/api/v1/project/create", map[string]interface{}{
		"title": "缺服务类型的项目",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("期望400，得到 %d: %s", w.Code, w.Body.String())
	}
}
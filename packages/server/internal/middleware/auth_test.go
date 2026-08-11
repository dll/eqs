package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func setup() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	r := gin.New()
	return r
}

// ensureUser 创建测试用户（若不存在）
func ensureUser(id uint, userType int) {
	var u model.User
	if err := model.DB.First(&u, id).Error; err != nil {
		model.DB.Create(&model.User{ID: id, Phone: "13" + fmt.Sprint(id%10000000000), UserType: userType, Status: 1})
	}
}

func TestAuth_MissingToken(t *testing.T) {
	r := setup()
	cfg := config.Load()
	r.Use(Auth(cfg))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望401，得到 %d", w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	r := setup()
	cfg := config.Load()
	r.Use(Auth(cfg))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望401，得到 %d", w.Code)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	r := setup()
	ensureUser(42, 2)
	cfg := config.Load()
	r.Use(Auth(cfg))

	var gotUserID uint
	var gotUserType int
	r.GET("/ping", func(c *gin.Context) {
		gotUserID = c.GetUint("user_id")
		gotUserType = c.GetInt("user_type")
		c.Status(http.StatusOK)
	})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   42,
		"user_type": 2,
		"exp":       time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString([]byte(cfg.JWTSecret))

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d", w.Code)
	}
	if gotUserID != 42 || gotUserType != 2 {
		t.Fatalf("期望注入 user_id=42 type=2，得到 %d %d", gotUserID, gotUserType)
	}
}

func TestAuth_NonMapClaims(t *testing.T) {
	r := setup()
	cfg := config.Load()
	r.Use(Auth(cfg))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// 使用自定义 Claims 类型（非 MapClaims）
	type customClaims struct {
		jwt.RegisteredClaims
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims{})
	tokenStr, _ := token.SignedString([]byte(cfg.JWTSecret))

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("非MapClaims应401，得到 %d", w.Code)
	}
}

func TestLogger_Run(t *testing.T) {
	r := setup()
	r.Use(Logger())
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d", w.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
}

func TestCORS_Headers(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	config.ResetCache()
	r := setup()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// 开发环境带来源：应回显该来源（非通配），保证凭证与调试可用
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://eqs.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://eqs.example.com" {
		t.Fatalf("开发环境应回显来源，得到 %q", got)
	}
}

func TestCORS_NoOrigin(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	config.ResetCache()
	r := setup()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// 无 Origin（同源/curl）：应返回通配，不拦截
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望200，得到 %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("无Origin应通配，得到 %q", got)
	}
}

func TestCORS_ProductionWhitelist(t *testing.T) {
	// 生产环境：白名单内放行、白名单外 403
	t.Setenv("APP_ENV", "production")
	t.Setenv("CORS_ALLOW_ORIGINS", "https://eqs-chzu.tech,https://www.eqs-chzu.tech")
	config.ResetCache()

	r := setup()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// 白名单来源
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://eqs-chzu.tech")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("白名单来源应放行，得到 %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://eqs-chzu.tech" {
		t.Fatalf("白名单来源应回显，得到 %q", got)
	}

	// 恶意来源
	req = httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非白名单来源应403，得到 %d", w.Code)
	}
}

func TestCORS_ProductionSameOriginIgnoringPort(t *testing.T) {
	// 生产环境经 nginx 代理：后端 Host 不带端口（$host），浏览器 Origin 带端口（:8091）
	// 必须按主机名判定同源并放行，否则同源登录会被误判为跨域 403
	t.Setenv("APP_ENV", "production")
	config.ResetCache()

	r := setup()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Host = "129.211.223.113" // nginx $host 去掉端口
	req.Header.Set("Origin", "http://129.211.223.113:8091")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("同源(去端口Host+带端口Origin)应放行，得到 %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://129.211.223.113:8091" {
		t.Fatalf("应回显来源，得到 %q", got)
	}
}

func TestCORS_OptionsPreflight(t *testing.T) {
	r := setup()
	r.Use(CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest("OPTIONS", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("预检应204，得到 %d", w.Code)
	}
}
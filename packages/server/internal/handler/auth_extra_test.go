package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func TestWxLogin(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/auth/wechat-login", map[string]interface{}{
		"code": "wx_code_001", "user_type": 2,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("微信登录失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["token"] == nil || out["token"] == "" {
		t.Fatalf("未返回token: %v", out)
	}

	// 二次登录不再是新用户
	w2 := doJSONFull(t, r, "POST", "/api/v1/auth/wechat-login", map[string]interface{}{
		"code": "wx_code_001", "user_type": 2,
	})
	out2 := decodeBody(t, w2)
	if out2["isNew"].(bool) {
		t.Fatalf("二次登录不应标记为新用户")
	}
}

func TestSendSMS(t *testing.T) {
	r := setupFlowRouter()
	w := doJSONFull(t, r, "POST", "/api/v1/sms/send", map[string]interface{}{
		"phone": "13700000001",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("发送验证码失败: %d %s", w.Code, w.Body.String())
	}

	// 缺手机号
	w = doJSONFull(t, r, "POST", "/api/v1/sms/send", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺手机号应400，得到 %d %s", w.Code, w.Body.String())
	}
}

func TestSignNotify(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	// 预置contract
	contract := model.Contract{
		OrderID: 1, SignFlowID: "flow_001", Status: "draft", SignProvider: "mock",
	}
	if err := model.DB.Create(&contract).Error; err != nil {
		t.Fatal(err)
	}

	notify := gin.New()
	notify.POST("/sign/notify", SignNotify)
	body := map[string]interface{}{
		"sign_flow_id": "flow_001", "order_id": 1, "result": "signed",
	}
	w := doJSONFull(t, notify, "POST", "/sign/notify", body)
	if w.Code != http.StatusOK {
		t.Fatalf("签约回调失败: %d %s", w.Code, w.Body.String())
	}

	// 幂等：重复回调
	w = doJSONFull(t, notify, "POST", "/sign/notify", body)
	if w.Code != http.StatusOK {
		t.Fatalf("幂等回调失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["idempotent"] != true {
		t.Fatalf("应标记幂等: %v", out)
	}

	// 不存在的流程
	body["sign_flow_id"] = "nope"
	w = doJSONFull(t, notify, "POST", "/sign/notify", body)
	if w.Code != http.StatusNotFound {
		t.Fatalf("未知流程应404，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestResponseHelpers 覆盖 created/unauthorized/serverError 未走到的响应分支
func TestResponseHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	r := gin.New()

	r.POST("/created", func(c *gin.Context) {
		created(c, gin.H{"ok": true})
	})
	r.POST("/unauthorized", func(c *gin.Context) {
		unauthorized(c, "未登录")
	})
	r.POST("/server-error", func(c *gin.Context) {
		serverError(c, errors.New("boom"))
	})
	r.POST("/server-error-nil", func(c *gin.Context) {
		serverError(c, nil)
	})

	req := func(path string) *httptest.ResponseRecorder {
		rr := httptest.NewRequest("POST", path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, rr)
		return w
	}

	if w := req("/created"); w.Code != http.StatusOK {
		t.Fatalf("created 应200，得到 %d", w.Code)
	}
	if w := req("/unauthorized"); w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized 应401，得到 %d", w.Code)
	}
	if w := req("/server-error"); w.Code != http.StatusInternalServerError {
		t.Fatalf("serverError 应500，得到 %d", w.Code)
	}
	if w := req("/server-error-nil"); w.Code != http.StatusInternalServerError {
		t.Fatalf("serverError(nil) 应500，得到 %d", w.Code)
	}
}

// TestSignContract_Resign 已签署合同重复签署返回成功
func TestSignContract_Resign(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13920000001", 1)
	projectID := setupPublishedProject(t, r)
	order := model.Order{ProjectID: projectID, SupplierID: 2, Amount: 2000, Status: 1}
	model.DB.Create(&order)

	w := doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/contract", nil)
	out := decodeBody(t, w)
	contractID := uint(out["contract"].(map[string]interface{})["id"].(float64))

	w = doJSONFull(t, r, "POST", "/api/v1/contract/"+strconv.Itoa(int(contractID))+"/sign", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("签署失败: %d %s", w.Code, w.Body.String())
	}
	// 重复签署幂等
	w = doJSONFull(t, r, "POST", "/api/v1/contract/"+strconv.Itoa(int(contractID))+"/sign", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("重复签署应幂等，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestWithdrawBid_OthersNotAllowed 越权撤回
func TestWithdrawBid_Forbidden(t *testing.T) {
	r := setupFlowRouter()
	createTestUser(t, "13920000010", 1)
	createTestUser(t, "13920000011", 2)
	projectID := setupPublishedProject(t, r)

	// 以服务方身份报价
	supplierRouter := gin.New()
	sr := supplierRouter.Group("")
	sr.Use(AuthTestMiddleware(2, 2))
	sr.POST("/bid/submit", SubmitBid)

	w := doJSONFull(t, supplierRouter, "POST", "/bid/submit", map[string]interface{}{
		"project_id": projectID, "amount": 100, "service_days": 1,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("报价失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	bidID := uint(out["bid"].(map[string]interface{})["id"].(float64))

	// 甲方(1)无权撤回服务方(2)报价
	clientRouter := gin.New()
	cr := clientRouter.Group("")
	cr.Use(AuthTestMiddleware(1, 1))
	cr.PUT("/bid/:id/withdraw", WithdrawBid)

	w = doJSONFull(t, clientRouter, "PUT", "/bid/"+strconv.Itoa(int(bidID))+"/withdraw", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("越权撤回应403，得到 %d %s", w.Code, w.Body.String())
	}
}

// TestSelectBid_Duplicate 同项目重复中选被拒
func TestSelectBid_Duplicate(t *testing.T) {
	r := setupFlowRouter()
	clientRouter := gin.New()
	cr := clientRouter.Group("")
	cr.Use(AuthTestMiddleware(1, 1))
	cr.POST("/api/v1/bid/:id/select", SelectBid)

	createTestUser(t, "13920000010", 1)
	createTestUser(t, "13920000011", 2)
	createTestUser(t, "13920000012", 2)
	projectID := setupPublishedProject(t, r)

	// 两位服务方分别报价（各自独立身份）
	supplierRouter2 := gin.New()
	s2r := supplierRouter2.Group("")
	s2r.Use(AuthTestMiddleware(2, 2))
	s2r.POST("/bid/submit", SubmitBid)

	supplierRouter3 := gin.New()
	s3r := supplierRouter3.Group("")
	s3r.Use(AuthTestMiddleware(3, 2))
	s3r.POST("/bid/submit", SubmitBid)

	bidIDs := []uint{}
	for _, rr := range []*gin.Engine{supplierRouter2, supplierRouter3} {
		w := doJSONFull(t, rr, "POST", "/bid/submit", map[string]interface{}{
			"project_id": projectID, "amount": 1000, "service_days": 3,
		})
		if w.Code != http.StatusOK {
			t.Fatalf("报价失败: %d %s", w.Code, w.Body.String())
		}
		out := decodeBody(t, w)
		bidIDs = append(bidIDs, uint(out["bid"].(map[string]interface{})["id"].(float64)))
	}

	// 选中第一个后，第二个被拒绝
	for i, id := range bidIDs {
		w := doJSONFull(t, clientRouter, "POST", "/api/v1/bid/"+strconv.Itoa(int(id))+"/select", nil)
		if i == 0 {
			if w.Code != http.StatusOK {
				t.Fatalf("首次中选应成功: %d %s", w.Code, w.Body.String())
			}
		} else {
			if w.Code != http.StatusBadRequest {
				t.Fatalf("重复中选应被拒(400)，得到 %d %s", w.Code, w.Body.String())
			}
		}
	}
}

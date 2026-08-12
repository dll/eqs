package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

// setupCoverRouter 覆盖未测试的 handler 路由
func setupCoverRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()

	r := gin.New()
	api := r.Group("/api/v1")

	// 管理员路由
	admin := api.Group("")
	admin.Use(AuthTestMiddleware(9, 3))
	{
		admin.GET("/admin/operations-stats", AdminOperationsStats)
		admin.GET("/admin/users/:id", AdminGetUser)
		admin.PUT("/admin/users/:id/status", AdminUpdateUserStatus)
		admin.POST("/admin/ai/project-analysis", AIAnalyzeAllProjects)
		admin.GET("/admin/project-progress", ListProjectProgress)
		admin.GET("/admin/log/list", AdminListLogs)
		admin.POST("/admin/log/restore-config", AdminRestoreConfig)
	}

	// 认证路由（身份通过 X-Test-User-* 头覆盖）
	shared := api.Group("")
	shared.Use(AuthTestMiddleware(1, 1))
	{
		shared.GET("/provider/list", ListProviders)
		shared.GET("/provider/:id", GetProvider)
		shared.GET("/bid/mine", ListMyBids)
		shared.GET("/message/list", ListMessages)
		shared.POST("/message/send", SendMessage)
		shared.GET("/message/:id", GetMessage)
		shared.DELETE("/message/:id", DeleteMessage)
		shared.PUT("/message/read/:id", MarkMessageRead)
		shared.GET("/notification/list", ListNotifications)
		shared.GET("/notification/unread-count", NotificationUnreadCount)
		shared.PUT("/notification/read/:id", MarkNotificationRead)
		shared.GET("/log/list", ListMyLogs)
		shared.GET("/delivery-templates", ListDeliveryTemplates)
		shared.GET("/delivery-templates/:id", GetDeliveryTemplate)
		shared.POST("/delivery-templates/:id/validate", ValidateDeliveryChecklist)
		shared.GET("/project/mine", ListMyProjects)
		shared.GET("/project/checklist", GetServiceChecklist)
		shared.PUT("/project/:id", UpdateProject)
		shared.PUT("/project/:id/withdraw", WithdrawProject)
		shared.PUT("/project/:id/abolish", AbolishProject)
		shared.DELETE("/project/:id", DeleteProject)
		shared.GET("/project/:id/progress", GetProjectProgress)
		shared.POST("/project/:id/ai-analysis", AIAnalyzeProject)
		shared.GET("/project/:id/reviews", ListProjectReviews)
		shared.PUT("/order/:id", UpdateOrder)
		shared.POST("/order/:id/cancel", CancelOrder)
		shared.PUT("/dispute/:id", UpdateDispute)
		shared.POST("/dispute/:id/auto-expert", AutoAssignDisputeExperts)
		shared.GET("/dispute/mine", ListMyDisputes)
		shared.GET("/file/:id/download", DownloadFile)
		shared.GET("/file/:id/preview", PreviewFile)
		shared.GET("/qualification/:id", GetQualification)
		shared.PUT("/qualification/:id", UpdateQualification)
		shared.DELETE("/qualification/:id", DeleteQualification)
	}
	return r
}

// TestCover_AdminOperationsStats 运营看板
func TestCover_AdminOperationsStats(t *testing.T) {
	r := setupCoverRouter()
	createTestUser(t, "13970000001", 1)
	createTestUser(t, "13970000002", 2)
	createTestUser(t, "13970000003", 4)
	project := model.Project{UserID: 1, ProjectType: "cost", ServiceType: "cost", Title: "t", Status: 1}
	model.DB.Create(&project)
	model.DB.Create(&model.Bid{ProjectID: project.ID, SupplierID: 2, Amount: 100, Status: "selected"})
	model.DB.Create(&model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 100, Status: 3})

	w := doJSONFull(t, r, "GET", "/api/v1/admin/operations-stats", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("运营看板失败: %d %s", w.Code, w.Body.String())
	}
}

// TestCover_AdminGetUserAndStatus 用户详情与状态更新
func TestCover_AdminGetUserAndStatus(t *testing.T) {
	r := setupCoverRouter()
	createTestUser(t, "13970000010", 3)
	targetID := createTestUser(t, "13970000011", 2)
	project := model.Project{UserID: targetID, ProjectType: "cost", ServiceType: "cost", Title: "t", Status: 1}
	model.DB.Create(&project)
	model.DB.Create(&model.Order{ProjectID: project.ID, SupplierID: targetID, Amount: 100, Status: 1})

	w := doJSONFull(t, r, "GET", "/api/v1/admin/users/"+strconv.Itoa(int(targetID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("用户详情失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "PUT", "/api/v1/admin/users/"+strconv.Itoa(int(targetID))+"/status", map[string]interface{}{"status": 1})
	if w.Code != http.StatusOK {
		t.Fatalf("启用用户失败: %d %s", w.Code, w.Body.String())
	}

	// 不能操作当前登录账号（管理员 user_id=9）
	w = doJSONFull(t, r, "PUT", "/api/v1/admin/users/9/status", map[string]interface{}{"status": 1})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("操作自己应400，得到 %d", w.Code)
	}
	// 非法状态
	w = doJSONFull(t, r, "PUT", "/api/v1/admin/users/"+strconv.Itoa(int(targetID))+"/status", map[string]interface{}{"status": 5})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法状态应400，得到 %d", w.Code)
	}
	// 不存在用户
	w = doJSONFull(t, r, "PUT", "/api/v1/admin/users/99999/status", map[string]interface{}{"status": 1})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在用户应404，得到 %d", w.Code)
	}
	// 非法 ID
	w = doJSONFull(t, r, "GET", "/api/v1/admin/users/abc", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d", w.Code)
	}
}

// TestCover_AI 全量/单项目 AI 分析（无 API key 走规则降级）
func TestCover_AI(t *testing.T) {
	r := setupCoverRouter()
	createTestUser(t, "13970000020", 1)
	project := model.Project{UserID: 1, ProjectType: "cost", ServiceType: "cost", Title: "测试项目", Status: 4}
	model.DB.Create(&project)

	w := doJSONFull(t, r, "POST", "/api/v1/admin/ai/project-analysis", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("全量AI分析失败: %d %s", w.Code, w.Body.String())
	}

	w = doJSONFull(t, r, "POST", "/api/v1/project/"+strconv.Itoa(int(project.ID))+"/ai-analysis", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("单项目AI分析失败: %d %s", w.Code, w.Body.String())
	}

	// 项目不存在
	w = doJSONFull(t, r, "POST", "/api/v1/project/99999/ai-analysis", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应404，得到 %d", w.Code)
	}

	// 非法 ID
	w = doJSONFull(t, r, "POST", "/api/v1/project/abc/ai-analysis", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d", w.Code)
	}
}

// TestCover_Messages 消息与通知全流程
func TestCover_Messages(t *testing.T) {
	r := setupCoverRouter()
	createTestUser(t, "13970000030", 2)
	createTestUser(t, "13970000031", 1)

	// 发送消息（发送者为 user_id=1 甲方）
	w := doJSONFull(t, r, "POST", "/api/v1/message/send", map[string]interface{}{
		"receiver_id": 2, "content": "你好，报价已提交",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("发送消息失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	msgID := uint(out["message"].(map[string]interface{})["id"].(float64))

	// 消息列表（带 other_id/order_id 过滤）
	w = doJSONFull(t, r, "GET", "/api/v1/message/list?other_id=2&order_id=1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("消息列表失败: %d %s", w.Code, w.Body.String())
	}
	w = doJSONFull(t, r, "GET", "/api/v1/message/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("消息列表失败: %d", w.Code)
	}

	// 消息详情（admin user_id=9 可查看任意消息）
	w = doJSONFullAuth(t, r, "GET", "/api/v1/message/"+strconv.Itoa(int(msgID)), nil, 9, 3)
	if w.Code != http.StatusOK {
		t.Fatalf("消息详情失败: %d %s", w.Code, w.Body.String())
	}

	// 删除消息（发送方=user1）
	w = doJSONFull(t, r, "DELETE", "/api/v1/message/"+strconv.Itoa(int(msgID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("删除消息失败: %d %s", w.Code, w.Body.String())
	}

	// 通知：创建、列表、未读数、标记已读
	model.DB.Create(&model.Notification{UserID: 1, Title: "t", Content: "c", Type: "system", IsRead: 0})
	var notif model.Notification
	model.DB.Where("user_id = ?", 1).First(&notif)

	w = doJSONFull(t, r, "GET", "/api/v1/notification/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("通知列表失败: %d", w.Code)
	}
	w = doJSONFull(t, r, "GET", "/api/v1/notification/unread-count", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("未读数失败: %d", w.Code)
	}
	w = doJSONFull(t, r, "PUT", "/api/v1/notification/read/"+strconv.Itoa(int(notif.ID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("标记已读失败: %d %s", w.Code, w.Body.String())
	}
	// 无权操作他人通知（通知属于 user_id=2，当前 user_id=1）
	model.DB.Create(&model.Notification{UserID: 2, Title: "t2", Content: "c2", Type: "system", IsRead: 0})
	var notif2 model.Notification
	model.DB.Where("user_id = ?", 2).First(&notif2)
	w = doJSONFull(t, r, "PUT", "/api/v1/notification/read/"+strconv.Itoa(int(notif2.ID)), nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("无权操作他人通知应403，得到 %d", w.Code)
	}
	// 无效消息ID
	w = doJSONFull(t, r, "PUT", "/api/v1/message/read/abc", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("无效消息ID应400，得到 %d", w.Code)
	}
	// 不存在消息
	w = doJSONFull(t, r, "GET", "/api/v1/message/99999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在消息应404，得到 %d", w.Code)
	}
}

// TestCover_Logs 个人日志、管理员日志、配置恢复
func TestCover_Logs(t *testing.T) {
	r := setupCoverRouter()
	createTestUser(t, "13970000040", 1)
	createTestUser(t, "13970000041", 3)

	// 记录审计日志（config 类型，供恢复）
	model.DB.Create(&model.AuditLog{
		UserID: 1, Action: "config.upsert", TargetType: "config", TargetID: 1,
		Detail: `{"key":"site.name","value":"新版","value_type":"string"}`,
	})

	w := doJSONFull(t, r, "GET", "/api/v1/log/list", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("个人日志失败: %d %s", w.Code, w.Body.String())
	}

	// 管理员日志（带筛选）
	w = doJSONFull(t, r, "GET", "/api/v1/admin/log/list?user_id=1&action=config&page=1&size=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("管理员日志失败: %d %s", w.Code, w.Body.String())
	}
	w = doJSONFull(t, r, "GET", "/api/v1/admin/log/list?user_id=abc&action=x&page=0&size=999", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("管理员日志筛选失败: %d", w.Code)
	}

	// 配置恢复
	var log model.AuditLog
	model.DB.Where("action = ?", "config.upsert").First(&log)
	w = doJSONFull(t, r, "POST", "/api/v1/admin/log/restore-config", map[string]interface{}{"log_id": log.ID})
	if w.Code != http.StatusOK {
		t.Fatalf("配置恢复失败: %d %s", w.Code, w.Body.String())
	}

	// 非 config 日志
	model.DB.Create(&model.AuditLog{UserID: 1, Action: "user.login", TargetType: "user", TargetID: 1, Detail: "{}"})
	var log2 model.AuditLog
	model.DB.Where("action = ?", "user.login").First(&log2)
	w = doJSONFull(t, r, "POST", "/api/v1/admin/log/restore-config", map[string]interface{}{"log_id": log2.ID})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非config日志应400，得到 %d", w.Code)
	}
	// 缺少 log_id
	w = doJSONFull(t, r, "POST", "/api/v1/admin/log/restore-config", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("缺log_id应400，得到 %d", w.Code)
	}
}

// TestCover_Providers 服务超市
func TestCover_Providers(t *testing.T) {
	r := setupCoverRouter()
	supplierID := createTestUser(t, "13970000050", 2)
	createTestUser(t, "13970000051", 1)
	model.DB.Create(&model.SupplierQualification{
		SupplierID: supplierID, QualificationType: "cost", Level: "甲级", Scope: "造价",
		VerificationStatus: "approved",
	})
	// 服务方订单与评价（providerStats）
	project := model.Project{UserID: 3, ProjectType: "cost", ServiceType: "cost", Title: "t", Status: 1}
	model.DB.Create(&project)
	model.DB.Create(&model.Order{ProjectID: project.ID, SupplierID: supplierID, Amount: 100, Status: 3})
	model.DB.Create(&model.Review{OrderID: 1, ReviewerID: 3, RevieweeID: supplierID, Rating: 5, Content: "good"})

	w := doJSONFull(t, r, "GET", "/api/v1/provider/list?type=cost", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("服务商列表失败: %d %s", w.Code, w.Body.String())
	}
	w = doJSONFull(t, r, "GET", "/api/v1/provider/"+strconv.Itoa(int(supplierID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("服务商详情失败: %d %s", w.Code, w.Body.String())
	}
	// 非服务商用户
	w = doJSONFull(t, r, "GET", "/api/v1/provider/2", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("非服务商应404，得到 %d", w.Code)
	}
	w = doJSONFull(t, r, "GET", "/api/v1/provider/99999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在服务商应404，得到 %d", w.Code)
	}
}

// TestCover_Templates 交付模板
func TestCover_Templates(t *testing.T) {
	r := setupCoverRouter()
	model.DB.Create(&model.DeliveryTemplate{
		ServiceType: "cost", Name: "造价交付标准", Version: "v1.0",
		Checklist: `["工程量清单","控制价说明"]`, Status: "active",
	})
	var tpl model.DeliveryTemplate
	model.DB.Where("status = ?", "active").First(&tpl)

	w := doJSONFull(t, r, "GET", "/api/v1/delivery-templates?service_type=cost", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("模板列表失败: %d %s", w.Code, w.Body.String())
	}
	w = doJSONFull(t, r, "GET", "/api/v1/delivery-templates/"+strconv.Itoa(int(tpl.ID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("模板详情失败: %d", w.Code)
	}
	// 校验清单（部分缺失）
	w = doJSONFull(t, r, "POST", "/api/v1/delivery-templates/"+strconv.Itoa(int(tpl.ID))+"/validate", map[string]interface{}{
		"checklist_result": map[string]interface{}{"工程量清单": true},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("校验清单失败: %d %s", w.Code, w.Body.String())
	}
	// 非法 ID / 不存在
	w = doJSONFull(t, r, "GET", "/api/v1/delivery-templates/abc", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法模板ID应400，得到 %d", w.Code)
	}
	w = doJSONFull(t, r, "POST", "/api/v1/delivery-templates/99999/validate", map[string]interface{}{})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在模板应404，得到 %d", w.Code)
	}
}

// TestCover_Progress 甘特图/看板进度
func TestCover_Progress(t *testing.T) {
	r := setupCoverRouter()
	createTestUser(t, "13970000060", 1)
	now := time.Now()
	project := model.Project{
		UserID: 1, ProjectType: "cost", ServiceType: "cost", Title: "进度项目",
		Status: 2, PublishTime: &now, Deadline: &now,
	}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 10000, Status: 2}
	model.DB.Create(&order)
	model.DB.Create(&model.PaymentMilestone{
		OrderID: order.ID, Name: "首期", Sequence: 1, Ratio: 50, Amount: 5000, Status: "settled",
	})
	model.DB.Create(&model.PaymentMilestone{
		OrderID: order.ID, Name: "尾款", Sequence: 2, Ratio: 50, Amount: 5000, Status: "pending",
	})

	// 管理员看板（覆盖全量 + 里程碑）
	w := doJSONFull(t, r, "GET", "/api/v1/admin/project-progress", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("全量进度失败: %d %s", w.Code, w.Body.String())
	}

	// 单项目进度
	w = doJSONFull(t, r, "GET", "/api/v1/project/"+strconv.Itoa(int(project.ID))+"/progress", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("单项目进度失败: %d %s", w.Code, w.Body.String())
	}
	w = doJSONFull(t, r, "GET", "/api/v1/project/abc/progress", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d", w.Code)
	}
	w = doJSONFull(t, r, "GET", "/api/v1/project/99999/progress", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应404，得到 %d", w.Code)
	}
}

// TestCover_ProjectOps 项目编辑/撤销/废除/删除/我的项目/清单
func TestCover_ProjectOps(t *testing.T) {
	r := setupCoverRouter()
	clientID := createTestUser(t, "13970000070", 1)
	createTestUser(t, "13970000071", 3)

	// 无业务往来项目（可撤销）
	project := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "撤", Status: 1}
	model.DB.Create(&project)

	w := doJSONFull(t, r, "PUT", "/api/v1/project/"+strconv.Itoa(int(project.ID))+"/withdraw", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("撤销失败: %d %s", w.Code, w.Body.String())
	}

	// 有报价项目（不能撤销，只能废除）
	project2 := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "废", Status: 2}
	model.DB.Create(&project2)
	model.DB.Create(&model.Bid{ProjectID: project2.ID, SupplierID: 2, Amount: 100, Status: "submitted"})

	w = doJSONFull(t, r, "PUT", "/api/v1/project/"+strconv.Itoa(int(project2.ID))+"/withdraw", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("有业务往来应400，得到 %d", w.Code)
	}
	w = doJSONFull(t, r, "PUT", "/api/v1/project/"+strconv.Itoa(int(project2.ID))+"/abolish", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("废除失败: %d %s", w.Code, w.Body.String())
	}

	// 我的项目列表
	w = doJSONFull(t, r, "GET", "/api/v1/project/mine?status=1&page=1&size=20", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("我的项目失败: %d %s", w.Code, w.Body.String())
	}

	// 服务清单
	w = doJSONFull(t, r, "GET", "/api/v1/project/checklist?service_type=cost", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("服务清单失败: %d", w.Code)
	}
	w = doJSONFull(t, r, "GET", "/api/v1/project/checklist?service_type=unknown", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("未知服务类型应200，得到 %d", w.Code)
	}

	// 编辑项目
	w = doJSONFull(t, r, "PUT", "/api/v1/project/"+strconv.Itoa(int(project2.ID)), map[string]interface{}{
		"project_type": "cost", "service_type": "cost", "title": "新标题",
		"description": "desc", "address": "addr", "budget_min": 100, "budget_max": 200,
		"deadline": time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
	})
	if w.Code != http.StatusOK {
		t.Fatalf("编辑项目失败: %d %s", w.Code, w.Body.String())
	}

	// 非发布者编辑被拒（user_id=9 非 owner）
	w = doJSONFullAuth(t, r, "PUT", "/api/v1/project/"+strconv.Itoa(int(project2.ID)), map[string]interface{}{
		"project_type": "cost", "service_type": "cost", "title": "越权",
	}, 9, 2)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非发布者编辑应403，得到 %d", w.Code)
	}

	// 管理员物理删除
	w = doJSONFullAuth(t, r, "DELETE", "/api/v1/project/"+strconv.Itoa(int(project2.ID)), nil, 9, 3)
	if w.Code != http.StatusOK {
		t.Fatalf("管理员删除失败: %d %s", w.Code, w.Body.String())
	}
	// 非管理员删除被拒
	w = doJSONFullAuth(t, r, "DELETE", "/api/v1/project/"+strconv.Itoa(int(project2.ID)), nil, 1, 1)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非管理员删除应403，得到 %d", w.Code)
	}
	// 不存在项目删除
	w = doJSONFullAuth(t, r, "DELETE", "/api/v1/project/99999", nil, 9, 3)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目删除应404，得到 %d", w.Code)
	}
}

// TestCover_OrderUpdateCancel 订单更新与取消
func TestCover_OrderUpdateCancel(t *testing.T) {
	r := setupCoverRouter()
	clientID := createTestUser(t, "13970000080", 1)
	createTestUser(t, "13970000081", 2)
	project := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "单", Status: 1}
	model.DB.Create(&project)

	// 待签约订单：更新金额/备注/工期
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 10000, Status: 0}
	model.DB.Create(&order)
	w := doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID)), map[string]interface{}{
		"remark": "加急", "amount": 12000, "expect_days": 30,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("更新订单失败: %d %s", w.Code, w.Body.String())
	}

	// 无更新字段
	w = doJSONFull(t, r, "PUT", "/api/v1/order/"+strconv.Itoa(int(order.ID)), map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("无字段应400，得到 %d", w.Code)
	}

	// 取消订单
	w = doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order.ID))+"/cancel", map[string]interface{}{"reason": "不再需要"})
	if w.Code != http.StatusOK {
		t.Fatalf("取消订单失败: %d %s", w.Code, w.Body.String())
	}
	// 已签约订单不可取消
	order2 := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 10000, Status: 1}
	model.DB.Create(&order2)
	w = doJSONFull(t, r, "POST", "/api/v1/order/"+strconv.Itoa(int(order2.ID))+"/cancel", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("已签约不可取消应400，得到 %d", w.Code)
	}
	// 非法 ID
	w = doJSONFull(t, r, "PUT", "/api/v1/order/abc", map[string]interface{}{"remark": "x"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d", w.Code)
	}
	// 不存在订单
	w = doJSONFull(t, r, "POST", "/api/v1/order/99999/cancel", map[string]interface{}{})
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在订单应404，得到 %d", w.Code)
	}
}

// TestCover_DisputeOps 争议更新/自动指派/我的争议
func TestCover_DisputeOps(t *testing.T) {
	r := setupCoverRouter()
	clientID := createTestUser(t, "13970000090", 1)
	createTestUser(t, "13970000091", 4)
	createTestUser(t, "13970000092", 4)
	project := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "争议项目", Status: 2}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 5000, Status: 2}
	model.DB.Create(&order)
	dispute := model.Dispute{OrderID: order.ID, InitiatorID: clientID, Reason: "质量", Status: "evidence"}
	model.DB.Create(&dispute)

	// 更新争议
	w := doJSONFull(t, r, "PUT", "/api/v1/dispute/"+strconv.Itoa(int(dispute.ID)), map[string]interface{}{
		"reason": "补充分歧", "claim": "要求返工",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("更新争议失败: %d %s", w.Code, w.Body.String())
	}
	// 空更新
	w = doJSONFull(t, r, "PUT", "/api/v1/dispute/"+strconv.Itoa(int(dispute.ID)), map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("空更新应400，得到 %d", w.Code)
	}

	// 自动指派专家（管理员 user_id=9）
	w = doJSONFullAuth(t, r, "POST", "/api/v1/dispute/"+strconv.Itoa(int(dispute.ID))+"/auto-expert", nil, 9, 3)
	if w.Code != http.StatusOK {
		t.Fatalf("自动指派失败: %d %s", w.Code, w.Body.String())
	}

	// 我的争议列表
	w = doJSONFull(t, r, "GET", "/api/v1/dispute/mine", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("我的争议失败: %d %s", w.Code, w.Body.String())
	}
}

// TestCover_Files 下载与预览
func TestCover_Files(t *testing.T) {
	r := setupCoverRouter()
	clientID := createTestUser(t, "13970000100", 1)
	project := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "文件", Status: 1}
	model.DB.Create(&project)

	// 外部 URL 文件（302 重定向）
	urlFile := model.ProjectFile{
		ProjectID: project.ID, UploaderID: clientID, OriginalName: "a.pdf", FileType: "pdf",
		StorageKey: "https://cdn.example.com/a.pdf", Version: 1,
	}
	model.DB.Create(&urlFile)
	w := doJSONFull(t, r, "GET", "/api/v1/file/"+strconv.Itoa(int(urlFile.ID))+"/download", nil)
	if w.Code != http.StatusFound {
		t.Fatalf("下载外部文件应302，得到 %d %s", w.Code, w.Body.String())
	}

	// 本地存储 key 文件（返回元信息）
	localFile := model.ProjectFile{
		ProjectID: project.ID, UploaderID: clientID, OriginalName: "b.dwg", FileType: "dwg",
		StorageKey: "uploads/projects/b.dwg", Version: 1,
	}
	model.DB.Create(&localFile)
	w = doJSONFull(t, r, "GET", "/api/v1/file/"+strconv.Itoa(int(localFile.ID))+"/download", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("下载本地文件失败: %d %s", w.Code, w.Body.String())
	}

	// 预览：不支持的类型
	w = doJSONFull(t, r, "GET", "/api/v1/file/"+strconv.Itoa(int(localFile.ID))+"/preview", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("不支持类型应400，得到 %d", w.Code)
	}
	// 预览：URL 文件 302
	w = doJSONFull(t, r, "GET", "/api/v1/file/"+strconv.Itoa(int(urlFile.ID))+"/preview", nil)
	if w.Code != http.StatusFound {
		t.Fatalf("预览URL文件应302，得到 %d", w.Code)
	}
	// 不存在文件
	w = doJSONFull(t, r, "GET", "/api/v1/file/99999/download", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在文件应404，得到 %d", w.Code)
	}
	// 非法 ID
	w = doJSONFull(t, r, "GET", "/api/v1/file/abc/download", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d", w.Code)
	}
}

// TestCover_QualificationOps 资质查看/编辑/删除
func TestCover_QualificationOps(t *testing.T) {
	r := setupCoverRouter()
	supplierID := createTestUser(t, "13970000110", 2)

	future := time.Now().AddDate(2, 0, 0)
	qual := model.SupplierQualification{
		SupplierID: supplierID, QualificationType: "造价", CertificateNo: "Z-2",
		VerificationStatus: "pending", ValidTo: &future,
	}
	model.DB.Create(&qual)

	// 查看（服务方本人，通过 X-Test-User-* 头覆盖 user_id=supplierID, user_type=2）
	w := doJSONFullAuth(t, r, "GET", "/api/v1/qualification/"+strconv.Itoa(int(qual.ID)), nil, int(supplierID), 2)
	if w.Code != http.StatusOK {
		t.Fatalf("查看资质失败: %d %s", w.Code, w.Body.String())
	}

	// 编辑（pending 可改）
	w = doJSONFullAuth(t, r, "PUT", "/api/v1/qualification/"+strconv.Itoa(int(qual.ID)), map[string]interface{}{
		"qualification_type": "监理", "certificate_no": "Z-2-new", "level": "乙级", "scope": "房建",
	}, int(supplierID), 2)
	if w.Code != http.StatusOK {
		t.Fatalf("编辑资质失败: %d %s", w.Code, w.Body.String())
	}

	// 删除
	w = doJSONFullAuth(t, r, "DELETE", "/api/v1/qualification/"+strconv.Itoa(int(qual.ID)), nil, int(supplierID), 2)
	if w.Code != http.StatusOK {
		t.Fatalf("删除资质失败: %d %s", w.Code, w.Body.String())
	}

	// 不存在
	w = doJSONFull(t, r, "GET", "/api/v1/qualification/99999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在资质应404，得到 %d", w.Code)
	}
	w = doJSONFull(t, r, "PUT", "/api/v1/qualification/abc", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("非法ID应400，得到 %d", w.Code)
	}
}

// TestCover_ProjectReviews 项目评价
func TestCover_ProjectReviews(t *testing.T) {
	r := setupCoverRouter()
	clientID := createTestUser(t, "13970000120", 1)
	project := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "评价", Status: 4}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 100, Status: 3}
	model.DB.Create(&order)
	model.DB.Create(&model.Review{OrderID: order.ID, ReviewerID: 2, RevieweeID: clientID, Rating: 5, Content: "ok"})

	// 项目创建者视角
	w := doJSONFull(t, r, "GET", "/api/v1/project/"+strconv.Itoa(int(project.ID))+"/reviews", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("项目评价失败: %d %s", w.Code, w.Body.String())
	}
	// 无订单项目评价
	project2 := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "评价2", Status: 4}
	model.DB.Create(&project2)
	w = doJSONFull(t, r, "GET", "/api/v1/project/"+strconv.Itoa(int(project2.ID))+"/reviews", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("无订单评价应200，得到 %d", w.Code)
	}
	w = doJSONFull(t, r, "GET", "/api/v1/project/99999/reviews", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("不存在项目应404，得到 %d", w.Code)
	}
}

// TestCover_ListMyBids 我的报价
func TestCover_ListMyBids(t *testing.T) {
	r := setupCoverRouter()
	supplierID := createTestUser(t, "13970000130", 2)
	createTestUser(t, "13970000131", 1)
	project := model.Project{UserID: 2, ProjectType: "cost", ServiceType: "cost", Title: "报价", Status: 1}
	model.DB.Create(&project)
	model.DB.Create(&model.Bid{ProjectID: project.ID, SupplierID: supplierID, Amount: 100, Status: "submitted"})

	w := doJSONFullAuth(t, r, "GET", "/api/v1/bid/mine?status=submitted&page=1&size=20", nil, int(supplierID), 2)
	if w.Code != http.StatusOK {
		t.Fatalf("我的报价失败: %d %s", w.Code, w.Body.String())
	}
	out := decodeBody(t, w)
	if out["total"].(float64) != 1 {
		t.Fatalf("应见1条报价，得到 %v", out["total"])
	}

	// 非服务方被拒
	w = doJSONFull(t, r, "GET", "/api/v1/bid/mine", nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("非服务方应403，得到 %d", w.Code)
	}
}

// TestCover_IsDemoPhone 演示手机号
func TestCover_IsDemoPhone(t *testing.T) {
	if !isDemoPhone("13900001111") {
		t.Fatal("1390000 前缀应判为演示号")
	}
	if isDemoPhone("13800138000") {
		t.Fatal("非演示号不应误判")
	}
	if isDemoPhone("1390000111") {
		t.Fatal("10位不应判为演示号")
	}
}

// TestCover_HelperFuncs 工具函数
func TestCover_HelperFuncs(t *testing.T) {
	var arr []string
	if err := parseJSONString(`["a","b"]`, &arr); err != nil || len(arr) != 2 {
		t.Fatalf("parseJSONString 失败: %v", err)
	}
	if err := parseJSONString("", &arr); err != nil {
		t.Fatalf("空串应返回 nil: %v", err)
	}
	if err := parseJSONString("not-json", &arr); err == nil {
		t.Fatal("非法 JSON 应报错")
	}

	if got := parseChecklist(`["a","b"]`); len(got) != 2 {
		t.Fatalf("parseChecklist 失败: %v", got)
	}
	if got := parseChecklist("bad"); len(got) != 0 {
		t.Fatalf("非法清单应返回空: %v", got)
	}

	if uint2str(123) != "123" {
		t.Fatal("uint2str 失败")
	}

	if fmtDate(nil) != nil {
		t.Fatal("fmtDate(nil) 应为 nil")
	}
	tm := time.Now()
	if fmtDate(&tm) == nil {
		t.Fatal("fmtDate 非 nil 应为指针")
	}
	if fmtDateStr(nil) != "未填写" {
		t.Fatal("fmtDateStr(nil) 应为未填写")
	}
	if fmtDateStr(&tm) == "未填写" {
		t.Fatal("fmtDateStr 应返回日期")
	}

	if ids := allOrderIDs(nil); len(ids) != 1 || ids[0] != 0 {
		t.Fatalf("allOrderIDs 空列表应返回[0]: %v", ids)
	}
	orders := []model.Order{{ID: 1}, {ID: 2}}
	if ids := allOrderIDs(orders); len(ids) != 2 {
		t.Fatalf("allOrderIDs 失败: %v", ids)
	}

	if projectStatusText(0) != "草稿" || projectStatusText(4) != "已完成" || projectStatusText(9) == "" {
		t.Fatal("projectStatusText 失败")
	}

	// calcProjectProgress 各分支
	p0 := model.Project{Status: 0}
	p4 := model.Project{Status: 4}
	if got := calcProjectProgress(p0, nil); got != 0 {
		t.Fatalf("草稿应0，得到 %d", got)
	}
	if got := calcProjectProgress(p4, nil); got != 100 {
		t.Fatalf("完成应100，得到 %d", got)
	}
	ms := []MilestoneBar{{Name: "a", Ratio: 50, Status: "settled"}, {Name: "b", Ratio: 50, Status: "pending"}}
	if got := calcProjectProgress(p0, ms); got != 50 {
		t.Fatalf("进度应50，得到 %d", got)
	}
	ms0 := []MilestoneBar{{Name: "a", Ratio: 0, Status: "settled"}}
	if got := calcProjectProgress(p0, ms0); got != 0 {
		t.Fatalf("ratio0 应0，得到 %d", got)
	}
	msOver := []MilestoneBar{{Name: "a", Ratio: 100, Status: "settled"}}
	if got := calcProjectProgress(p0, msOver); got != 100 {
		t.Fatalf("超100应100，得到 %d", got)
	}
}

// TestCover_VerifyCallbackSign 支付回调验签
func TestCover_VerifyCallbackSign(t *testing.T) {
	secret := "test-secret"
	txid := "TX-001"
	orderID := uint(42)
	amount := 1234.56
	result := "success"
	ts := time.Now().Unix()
	str := fmt.Sprintf("%s|%d|%.2f|%s|%d", txid, orderID, amount, result, ts)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(str))
	expect := hex.EncodeToString(mac.Sum(nil))

	if !verifyCallbackSign(secret, txid, orderID, amount, result, ts, expect) {
		t.Fatal("正确签名应通过")
	}
	if verifyCallbackSign(secret, txid, orderID, amount, result, ts, "wrong-sign") {
		t.Fatal("错误签名应失败")
	}
	if verifyCallbackSign(secret, txid, orderID, amount, "fail", ts, expect) {
		t.Fatal("不同结果签名应失败")
	}
}

// TestCover_ScopeHelpers scope 辅助：getOrderForUser/getDisputeForUser
func TestCover_ScopeHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	clientID := createTestUser(t, "13970000140", 1)
	createTestUser(t, "13970000141", 2)
	project := model.Project{UserID: clientID, ProjectType: "cost", ServiceType: "cost", Title: "scope", Status: 1}
	model.DB.Create(&project)
	order := model.Order{ProjectID: project.ID, SupplierID: 2, Amount: 100, Status: 1}
	model.DB.Create(&order)
	dispute := model.Dispute{OrderID: order.ID, InitiatorID: clientID, Reason: "r", Status: "evidence"}
	model.DB.Create(&dispute)

	mkCtx := func(userID, userType int) *gin.Context {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/", nil)
		ctx.Set("user_id", uint(userID))
		ctx.Set("user_type", userType)
		return ctx
	}

	// 甲方本人可访问订单
	if o, ok := getOrderForUser(mkCtx(int(clientID), 1), order.ID); !ok || o == nil {
		t.Fatal("甲方应可访问订单")
	}
	// 无关用户不可访问
	if _, ok := getOrderForUser(mkCtx(9, 1), order.ID); ok {
		t.Fatal("无关用户不应访问订单")
	}
	// 管理员可访问
	if o, ok := getOrderForUser(mkCtx(9, 3), order.ID); !ok || o == nil {
		t.Fatal("管理员应可访问订单")
	}
	// 不存在订单
	if _, ok := getOrderForUser(mkCtx(1, 1), 99999); ok {
		t.Fatal("不存在订单应返回 false")
	}

	// 发起人可访问争议
	if d, ok := getDisputeForUser(mkCtx(int(clientID), 1), dispute.ID); !ok || d == nil {
		t.Fatal("发起人应可访问争议")
	}
	// 无关用户不可访问
	if _, ok := getDisputeForUser(mkCtx(9, 1), dispute.ID); ok {
		t.Fatal("无关用户不应访问争议")
	}
	// 不存在争议
	if _, ok := getDisputeForUser(mkCtx(1, 1), 99999); ok {
		t.Fatal("不存在争议应返回 false")
	}
}

// TestCover_CleanDemoUsers 清理演示用户
func TestCover_CleanDemoUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	createTestUser(t, "13900001111", 1)
	createTestUser(t, "13900002222", 2)
	createTestUser(t, "13970000150", 3) // 非演示号管理员保留

	cleanDemoUsers()
	var demoCount int64
	model.DB.Model(&model.User{}).Where("phone LIKE ?", "1390000%").Count(&demoCount)
	if demoCount != 0 {
		t.Fatalf("演示用户应被清理，剩 %d", demoCount)
	}
	var adminCount int64
	model.DB.Model(&model.User{}).Where("user_type = ?", 3).Count(&adminCount)
	if adminCount != 1 {
		t.Fatalf("管理员应保留，得到 %d", adminCount)
	}

	// 空场景再次清理（demo 用户不存在时提前返回）
	cleanDemoUsers()
}

// TestCover_UploadProjectFile 项目附件上传（multipart）
func TestCover_UploadProjectFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	createTestUser(t, "13970000160", 1)
	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(AuthTestMiddleware(1, 1))
	api.POST("/project/upload", UploadProjectFile)

	// 隔离上传目录到临时相对路径（避免绝对路径 Join 问题），用后清理
	old := os.Getenv("UPLOAD_DIR")
	tmpRel := "uploads_cov_test"
	os.Setenv("UPLOAD_DIR", tmpRel)
	config.ResetCache()
	defer func() {
		os.Setenv("UPLOAD_DIR", old)
		config.ResetCache()
		os.RemoveAll(tmpRel)
	}()

	body, contentType := buildMultipartBody(t, "design.pdf", []byte("%PDF-1.4 test"))
	w := doMultipart(t, r, "/api/v1/project/upload", body, contentType)
	if w.Code != http.StatusOK {
		t.Fatalf("上传项目附件失败: %d %s", w.Code, w.Body.String())
	}
}

// TestCover_UploadQualificationFile 资质扫描件上传
func TestCover_UploadQualificationFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	model.InitTestDB()
	createTestUser(t, "13970000170", 2)
	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(AuthTestMiddleware(2, 2))
	api.POST("/qualification/upload", UploadQualificationFile)

	old := os.Getenv("UPLOAD_DIR")
	tmpRel := "uploads_cov_test"
	os.Setenv("UPLOAD_DIR", tmpRel)
	config.ResetCache()
	defer func() {
		os.Setenv("UPLOAD_DIR", old)
		config.ResetCache()
		os.RemoveAll(tmpRel)
	}()

	body, contentType := buildMultipartBody(t, "cert.jpg", []byte{0xFF, 0xD8, 0xFF, 0xE0})
	w := doMultipart(t, r, "/api/v1/qualification/upload", body, contentType)
	if w.Code != http.StatusOK {
		t.Fatalf("上传资质扫描件失败: %d %s", w.Code, w.Body.String())
	}
}

// buildMultipartBody 构造 multipart/form-data 请求体
func buildMultipartBody(t *testing.T, filename string, content []byte) (bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("创建 multipart 失败: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("写入 multipart 失败: %v", err)
	}
	mw.Close()
	return buf, mw.FormDataContentType()
}

// doMultipart 发送 multipart 请求
func doMultipart(t *testing.T, r *gin.Engine, path string, body bytes.Buffer, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", path, &body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}


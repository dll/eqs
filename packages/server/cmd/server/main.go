package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/handler"
	"github.com/eqs/server/internal/middleware"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// 启动即构造进程内配置单例（后续热路径经 config.Get() 复用同一实例）
	cfg := config.Get()

	// P0-07：生产环境拒绝弱默认凭据/密钥启动
	if cfg.IsProduction() {
		weak := []string{"root", "eqs-secret-key", "", "123456"}
		for _, w := range weak {
			if cfg.JWTSecret == w || cfg.DBPassword == w {
				log.Fatalf("生产环境禁止使用默认/弱凭据：请设置强 JWT_SECRET 与 DB_PASSWORD")
			}
		}
		if cfg.DBPassword == "root" {
			log.Fatalf("生产环境禁止默认数据库密码 root")
		}
		// P1-09：生产环境必须配置字段加密密钥，否则敏感字段会以明文落库
		if cfg.DataEncryptionKey == "" {
			log.Fatalf("生产环境必须设置 DATA_ENCRYPTION_KEY（AES-256-GCM 敏感字段加密密钥）")
		}
	}

	// P1-09：注入敏感字段加密密钥（开发环境未配置时字段以明文存取）
	model.InitFieldCrypto(cfg.DataEncryptionKey)

	db, err := model.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// P2-08：版本化数据库迁移（仅 MySQL 驱动执行；SQLite 由 AutoMigrate 维护）
	if cfg.DBDriver == "mysql" {
		if err := model.ApplyMigrations(db); err != nil {
			log.Printf("[migration] 迁移执行失败（继续启动）: %v", err)
		}
	}

	if cfg.DBDriver == "sqlite" {
		log.Printf("SQLite 模式：使用本地文件库 %s，Redis 校验降级为内置模拟", cfg.DBName)
	} else {
		_ = model.InitRedis(cfg)
	}

	r := setupRouter(db, cfg)

	// 可信代理：仅信任本机 nginx/Caddy（127.0.0.1），防止伪造 X-Forwarded-For 错误归因 IP
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// 优雅停机：SIGTERM/SIGINT 时等待当前请求完成（最多 10s）再退出
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}
	log.Println("Server exited")
}

func setupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(handler.MonitorMiddleware())

	api := r.Group("/api/v1")
	{
		// Auth
		api.POST("/sms/send", handler.SendSMS)
		api.POST("/auth/login", handler.PhoneLogin)
		api.POST("/auth/wechat-login", handler.WxLogin)

		// Licensed provider callbacks (公开，验签后处理)
		api.POST("/pay/notify/:channel", handler.PaymentNotify)
		api.POST("/sign/notify", handler.SignNotify)

		// V7 公开接口（无需登录）
		api.GET("/config/public", handler.PublicConfigs)
		api.GET("/theme/list", handler.ThemeList)
		api.GET("/i18n/:lang", handler.I18nMessages)
		api.GET("/platform/links", handler.PlatformLinks)
		api.GET("/version/check", handler.VersionRateLimit(), handler.VersionCheck)
		api.GET("/version/latest", handler.VersionLatest)
		// V8 服务超市（公开浏览）
		api.GET("/provider/list", handler.ListProviders)
		api.GET("/provider/:id", handler.GetProvider)

		// Protected routes
		auth := api.Group("")
		auth.Use(middleware.Auth(cfg))
		{
			// Progress（单项目进度甘特图，需登录）
			auth.GET("/project/:id/progress", handler.GetProjectProgress)
			// V8 AI 单项目问题解析
			auth.POST("/project/:id/ai-analysis", handler.AIAnalyzeProject)
			// V8 个人日志
			auth.GET("/log/list", handler.ListMyLogs)
			// P1-06 标准交付模板
			auth.GET("/delivery-templates", handler.ListDeliveryTemplates)
			auth.GET("/delivery-templates/:id", handler.GetDeliveryTemplate)
			auth.POST("/delivery-templates/:id/validate", handler.ValidateDeliveryChecklist)
			// User
			auth.GET("/user/info", handler.GetUserInfo)
			auth.PUT("/user/info", handler.UpdateUserInfo)

			// Project
			auth.POST("/project/create", handler.CreateProject)
			auth.GET("/project/list", handler.ListProjects)
			auth.GET("/project/:id/recommend", handler.GetRecommendations)
			auth.GET("/project/:id", handler.GetProject)
			auth.PUT("/project/:id", handler.UpdateProject)
			auth.DELETE("/project/:id", handler.DeleteProject)
			auth.POST("/project/:id/invite", handler.InviteSuppliers)
			auth.GET("/project/:id/reviews", handler.ListProjectReviews)

			// Bid
			auth.POST("/bid/submit", handler.SubmitBid)
			auth.GET("/project/:id/bids", handler.ListBids)
			auth.PUT("/bid/:id/withdraw", handler.WithdrawBid)
			auth.POST("/bid/:id/select", handler.SelectBid)

			// Order
			auth.GET("/order/list", handler.ListMyOrders)
			auth.GET("/order/:id", handler.GetOrder)
			auth.PUT("/order/:id", handler.UpdateOrder)
			auth.POST("/order/:id/cancel", handler.CancelOrder)
			auth.PUT("/order/:id/milestones", handler.SetMilestones)
			auth.POST("/milestone/:id/deliver", handler.UploadDeliverable)
			auth.POST("/milestone/:id/accept", handler.ConfirmAcceptance)

			// Contract
			auth.GET("/contract/templates", handler.ListContractTemplates)
			auth.POST("/order/:id/contract", handler.GenerateContract)
			auth.POST("/contract/:id/sign", handler.SignContract)
			auth.GET("/contract/:id/download", handler.DownloadContract)

			// Payment（经持牌机构，非托管）
			auth.POST("/pay/create", handler.CreatePayment)
			auth.POST("/milestone/:id/settle", handler.SettleMilestone)
			auth.GET("/pay/transactions", handler.ListPaymentTransactions)
			auth.GET("/pay/balance", handler.GetBalance)

			// Dispute（专家评审+平台调解）
			auth.POST("/dispute/create", handler.CreateDispute)
			auth.GET("/order/:id/disputes", handler.ListDisputes)
			auth.GET("/dispute/mine", handler.ListMyDisputes)
			auth.POST("/dispute/:id/evidence", handler.UploadDisputeEvidence)
			auth.GET("/dispute/:id", handler.GetDispute)
			auth.PUT("/dispute/:id", handler.UpdateDispute)
			auth.POST("/dispute/:id/expert", handler.AssignDisputeExpert)
			auth.POST("/dispute-expert/:id/opinion", handler.SubmitExpertOpinion)
			auth.POST("/dispute/:id/close", handler.CloseDispute)

			// Attendance（现场打卡）
			auth.POST("/attendance/checkin", handler.CheckIn)
			auth.GET("/order/:id/attendance", handler.ListAttendance)

			// Qualification（资质核验）
			auth.GET("/supplier/:id/qualifications", handler.ListQualifications)
			auth.POST("/supplier/:id/qualifications", handler.SubmitQualification)
			auth.GET("/qualification/:id", handler.GetQualification)
			auth.PUT("/qualification/:id", handler.UpdateQualification)
			auth.DELETE("/qualification/:id", handler.DeleteQualification)
			auth.POST("/qualification/:id/review", handler.ReviewQualification)
			// V6：资质扫描件上传（附件备份，multipart file 字段）
			auth.POST("/qualification/upload", handler.UploadQualificationFile)

			// File（文件与批注）
			auth.POST("/file/upload", handler.UploadFile)
			auth.GET("/file/:id/download", handler.DownloadFile)
			auth.GET("/project/:id/files", handler.ListFiles)
			auth.POST("/annotation/add", handler.AddAnnotation)
			auth.GET("/annotation/list/:id", handler.ListAnnotations)
			auth.PUT("/annotation/:id/resolve", handler.ResolveAnnotation)

			// Review
			auth.POST("/review/submit", handler.SubmitReview)
			auth.GET("/user/:id/reviews", handler.GetUserReviews)

			// Message / Notification（V8 补齐）
			auth.POST("/message/send", handler.SendMessage)
			auth.GET("/message/list", handler.ListMessages)
			auth.GET("/message/:id", handler.GetMessage)
			auth.DELETE("/message/:id", handler.DeleteMessage)
			auth.PUT("/message/read/:id", handler.MarkMessageRead)
			auth.GET("/notification/list", handler.ListNotifications)
			auth.PUT("/notification/read/:id", handler.MarkNotificationRead)
			// V7 用户偏好（需登录）
			auth.GET("/config/user/prefs", handler.GetUserPrefs)
			auth.PUT("/config/user/prefs", handler.UpdateUserPrefs)
			auth.PUT("/project/:id/theme", handler.SetProjectTheme)
		}

		// Admin routes
		admin := api.Group("")
		admin.Use(middleware.Auth(cfg))
		admin.Use(handler.RequireAdmin())
		{
			admin.GET("/admin/stats", handler.AdminDashboardStats)
			admin.GET("/admin/users", handler.AdminListUsers)
			admin.PUT("/admin/users/:id/status", handler.AdminUpdateUserStatus)
			admin.GET("/admin/orders", handler.AdminListOrders)
			admin.GET("/admin/transactions", handler.AdminListTransactions)
			admin.GET("/admin/disputes", handler.AdminListDisputes)
			admin.GET("/admin/qualifications", handler.AdminListPendingQualifications)
			// Demo 数据管理（系统管理员功能）
			admin.POST("/admin/demo/seed", handler.DemoSeedHandler)
			admin.POST("/admin/demo/clean", handler.DemoCleanHandler)
			admin.POST("/admin/demo/toggle", handler.DemoToggleHandler)
			admin.GET("/admin/demo/status", handler.DemoStatusHandler)
			// V7 配置中心（管理员）
			admin.GET("/admin/config/list", handler.AdminListConfigs)
			admin.POST("/admin/config/upsert", handler.AdminUpsertConfig)
			admin.DELETE("/admin/config/delete/:key", handler.AdminDeleteConfig)
			admin.POST("/admin/version/publish", handler.AdminPublishVersion)
			admin.GET("/admin/version/list", handler.AdminListVersions)
			admin.GET("/admin/monitor/stats", handler.MonitorStats)
			// 佣金管理（SRC-BIZ-01）
			admin.GET("/admin/commission/list", handler.AdminListCommissions)
			admin.POST("/admin/commission/:id/collect", handler.AdminCollectCommission)
			// V8 数据看板（甘特图/看板）
			admin.GET("/admin/project-progress", handler.ListProjectProgress)
			// V8 AI 全量项目分析
			admin.POST("/admin/ai/project-analysis", handler.AIAnalyzeAllProjects)
			// V8 日志管理（管理员查看/恢复）
			admin.GET("/admin/log/list", handler.AdminListLogs)
			admin.POST("/admin/log/restore-config", handler.AdminRestoreConfig)
		}
	}

	return r
}

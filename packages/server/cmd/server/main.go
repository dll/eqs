package main

import (
	"log"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/handler"
	"github.com/eqs/server/internal/middleware"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := model.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if cfg.DBDriver == "sqlite" {
		log.Printf("SQLite 模式：使用本地文件库 %s，Redis 校验降级为内置模拟", cfg.DBName)
	} else {
		_ = model.InitRedis(cfg)
	}

	r := setupRouter(db, cfg)
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	api := r.Group("/api/v1")
	{
		// Auth
		api.POST("/sms/send", handler.SendSMS)
		api.POST("/auth/login", handler.PhoneLogin)
		api.POST("/auth/wechat-login", handler.WxLogin)

		// Licensed provider callbacks (公开，验签后处理)
		api.POST("/pay/notify/:channel", handler.PaymentNotify)
		api.POST("/sign/notify", handler.SignNotify)

		// Protected routes
		auth := api.Group("")
		auth.Use(middleware.Auth(cfg))
		{
			// User
			auth.GET("/user/info", handler.GetUserInfo)
			auth.PUT("/user/info", handler.UpdateUserInfo)

			// Project
			auth.POST("/project/create", handler.CreateProject)
			auth.GET("/project/list", handler.ListProjects)
			auth.GET("/project/:id/recommend", handler.GetRecommendations)
			auth.GET("/project/:id", handler.GetProject)
			auth.POST("/project/:id/invite", handler.InviteSuppliers)

			// Bid
			auth.POST("/bid/submit", handler.SubmitBid)
			auth.GET("/project/:id/bids", handler.ListBids)
			auth.PUT("/bid/:id/withdraw", handler.WithdrawBid)
			auth.POST("/bid/:id/select", handler.SelectBid)

			// Order
			auth.GET("/order/list", handler.ListMyOrders)
			auth.GET("/order/:id", handler.GetOrder)
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
			auth.POST("/dispute/:id/evidence", handler.UploadDisputeEvidence)
			auth.GET("/dispute/:id", handler.GetDispute)
			auth.POST("/dispute/:id/expert", handler.AssignDisputeExpert)
			auth.POST("/dispute-expert/:id/opinion", handler.SubmitExpertOpinion)
			auth.POST("/dispute/:id/close", handler.CloseDispute)

			// Attendance（现场打卡）
			auth.POST("/attendance/checkin", handler.CheckIn)
			auth.GET("/order/:id/attendance", handler.ListAttendance)

			// Qualification（资质核验）
			auth.GET("/supplier/:id/qualifications", handler.ListQualifications)
			auth.POST("/supplier/:id/qualifications", handler.SubmitQualification)
			auth.POST("/qualification/:id/review", handler.ReviewQualification)

			// File（文件与批注）
			auth.POST("/file/upload", handler.UploadFile)
			auth.GET("/project/:id/files", handler.ListFiles)
			auth.POST("/annotation/add", handler.AddAnnotation)
			auth.GET("/annotation/list/:id", handler.ListAnnotations)
			auth.PUT("/annotation/:id/resolve", handler.ResolveAnnotation)

			// Review
			auth.POST("/review/submit", handler.SubmitReview)
			auth.GET("/user/:id/reviews", handler.GetUserReviews)
		}

		// Admin routes
		admin := api.Group("")
		admin.Use(middleware.Auth(cfg))
		admin.Use(handler.RequireAdmin())
		{
			admin.GET("/admin/stats", handler.AdminDashboardStats)
			admin.GET("/admin/users", handler.AdminListUsers)
			admin.GET("/admin/orders", handler.AdminListOrders)
			admin.GET("/admin/transactions", handler.AdminListTransactions)
			admin.GET("/admin/disputes", handler.AdminListDisputes)
			admin.GET("/admin/qualifications", handler.AdminListPendingQualifications)
		}
	}

	return r
}
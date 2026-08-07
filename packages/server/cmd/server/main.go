package main

import (
	"log"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/handler"
	"github.com/eqs/server/internal/middleware"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := model.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	_ = model.InitRedis(cfg)

	r := setupRouter(db, cfg)
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(db *model.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	api := r.Group("/api/v1")
	{
		// Auth
		api.POST("/sms/send", handler.SendSMS)
		api.POST("/auth/login", handler.Login)

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
			auth.GET("/project/:id", handler.GetProject)
			auth.POST("/project/:id/apply", handler.ApplyProject)
			auth.POST("/project/:id/select", handler.SelectSupplier)

			// Order
			auth.POST("/order/create", handler.CreateOrder)
			auth.POST("/order/:id/deliver", handler.UploadDeliverable)
			auth.POST("/order/:id/confirm", handler.ConfirmDelivery)

			// Payment
			auth.POST("/pay/recharge", handler.Recharge)
			auth.POST("/pay/withdraw", handler.Withdraw)
			auth.GET("/pay/records", handler.GetPayRecords)
		}
	}

	return r
}

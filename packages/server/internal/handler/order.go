package handler

import (
	"net/http"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	ProjectID  uint    `json:"project_id" binding:"required"`
	SupplierID uint    `json:"supplier_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required"`
}

func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := model.Order{
		ProjectID:  req.ProjectID,
		SupplierID: req.SupplierID,
		Amount:     req.Amount,
		Status:     0,
	}

	if err := model.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func UploadDeliverable(c *gin.Context) {
	// TODO: Implement file upload logic
	c.JSON(http.StatusOK, gin.H{"message": "上传成功"})
}

func ConfirmDelivery(c *gin.Context) {
	id := c.Param("id")

	if err := model.DB.Model(&model.Order{}).Where("id = ?", id).Update("status", 3).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "确认失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验收成功"})
}

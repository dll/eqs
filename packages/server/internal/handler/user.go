package handler

import (
	"net/http"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUserInfo(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input struct {
		CompanyName string `json:"company_name"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	model.DB.Model(&model.User{}).Where("id = ?", userID).Update("company_name", input.CompanyName)
	c.JSON(http.StatusOK, gin.H{"message": "更新成功"})
}

package handler

import (
	"net/http"
	"time"

	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	UserType int    `json:"user_type" binding:"required"`
}

func SendSMS(c *gin.Context) {
	// TODO: Implement SMS sending logic
	c.JSON(http.StatusOK, gin.H{"message": "验证码已发送"})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Verify SMS code
	// For now, accept any code "123456"
	if req.Code != "123456" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误"})
		return
	}

	var user model.User
	result := model.DB.Where("phone = ?", req.Phone).First(&user)
	if result.Error != nil {
		user = model.User{
			Phone:    req.Phone,
			UserType: req.UserType,
			Status:   1,
		}
		model.DB.Create(&user)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"user_type": user.UserType,
		"exp":       time.Now().Add(72 * time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("eqs-secret-key"))

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

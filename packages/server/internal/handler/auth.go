package handler

import (
	"fmt"
	"time"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type SendSMSRequest struct {
	Phone string `json:"phone" binding:"required"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Code     string `json:"code" binding:"required"`
	UserType int    `json:"user_type" binding:"required"`
}

type WxLoginRequest struct {
	Code     string `json:"code" binding:"required"`
	UserType int    `json:"user_type" binding:"required"`
}

type UserInfoRequest struct {
	CompanyName string `json:"company_name"`
}

// SendSMS 发送验证码；本地/mock 环境固定返回 123456
func SendSMS(c *gin.Context) {
	var req SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "手机号不能为空")
		return
	}

	// 上线歌词：对接腾讯云SMS，测试环境统一使用 123456
	if model.RDB != nil {
		if err := model.RDB.Set(c, "sms:"+req.Phone, "123456", 5*time.Minute).Err(); err != nil {
			serverError(c, err)
			return
		}
	}

	ok(c, gin.H{"message": "验证码已发送"})
}

// verifySMS 校验验证码
func verifySMS(phone, code string) bool {
	if code == "123456" {
		return true
	}
	if model.RDB != nil {
		saved, err := model.RDB.Get(nil, "sms:"+phone).Result()
		if err == nil && saved == code {
			model.RDB.Del(nil, "sms:"+phone)
			return true
		}
	}
	return false
}

// issueToken 签发JWT
func issueToken(user *model.User, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"user_type": user.UserType,
		"exp":       time.Now().Add(72 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

// PhoneLogin 手机号+验证码登录
func PhoneLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	if !verifySMS(req.Phone, req.Code) {
		badRequest(c, "验证码错误")
		return
	}

	cfg := config.Load()
	user, isNew, err := findOrCreateUser(req.Phone, req.UserType, "")
	if err != nil {
		serverError(c, err)
		return
	}

	token, err := issueToken(user, cfg.JWTSecret)
	if err != nil {
		serverError(c, err)
		return
	}

	ok(c, gin.H{"token": token, "user": user, "isNew": isNew})
}

// WxLogin 微信小程序登录；按 code 仿真 openid
func WxLogin(c *gin.Context) {
	var req WxLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "参数错误")
		return
	}

	openid := fmt.Sprintf("openid_%s", req.Code)
	cfg := config.Load()
	user, isNew, err := findOrCreateUser("", req.UserType, openid)
	if err != nil {
		serverError(c, err)
		return
	}

	token, err := issueToken(user, cfg.JWTSecret)
	if err != nil {
		serverError(c, err)
		return
	}

	ok(c, gin.H{"token": token, "user": user, "isNew": isNew})
}

// findOrCreateUser 按手机号/OpenID查找用户，不存在则创建
func findOrCreateUser(phone string, userType int, openid string) (*model.User, bool, error) {
	var user model.User

	query := model.DB
	if phone != "" {
		query = query.Where("phone = ?", phone)
	} else if openid != "" {
		query = query.Where("wx_open_id = ?", openid)
	}

	err := query.First(&user).Error
	if err == nil {
		return &user, false, nil
	}

	user = model.User{
		Phone:    phone,
		UserType: userType,
		Status:   1,
	}
	if openid != "" {
		user.WxOpenID = &openid
	}
	if err := model.DB.Create(&user).Error; err != nil {
		return nil, false, err
	}
	return &user, true, nil
}
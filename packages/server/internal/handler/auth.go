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
	UserType int    `json:"user_type"`
}

type WxLoginRequest struct {
	Code     string `json:"code" binding:"required"`
	UserType int    `json:"user_type"`
}

type UserInfoRequest struct {
	CompanyName string `json:"company_name"`
}

// 公共注册允许的角色（甲方/服务方）；管理员(3)/专家(4)必须受控创建，禁止自注册
var publicRoles = map[int]bool{1: true, 2: true}

const smsFreqLimit = 60 * time.Second   // 发送间隔
const smsFailLockLimit = 5              // 连续失败锁定次数

// SendSMS 发送验证码；生产环境真实发送（未接SMS时禁用），非生产固定 123456
func SendSMS(c *gin.Context) {
	var req SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "手机号不能为空")
		return
	}

	cfg := config.Load()

	// 频控：60 秒内不允许重复发送
	if model.RDB != nil {
		last, err := model.RDB.Get(c, "sms:last:"+req.Phone).Result()
		if err == nil {
			if t, terr := time.Parse(time.RFC3339, last); terr == nil && time.Since(t) < smsFreqLimit {
				badRequest(c, "发送过于频繁，请稍后再试")
				return
			}
		}
		// 失败锁定检查
		fails, _ := model.RDB.Get(c, "sms:fail:"+req.Phone).Int()
		if fails >= smsFailLockLimit {
			badRequest(c, "尝试次数过多，请稍后再试")
			return
		}
	}

	code := "123456"
	if cfg.IsProduction() {
		// 生产环境：固定验证码禁用；未配置 SMS 时拒绝发送（避免固定码泄漏）
		if cfg.SMSAppKey == "" {
			badRequest(c, "短信服务未配置")
			return
		}
		code = fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
		// TODO: 对接腾讯云 SMS 实际发送
	}

	if model.RDB != nil {
		model.RDB.Set(c, "sms:"+req.Phone, code, 5*time.Minute)
		model.RDB.Set(c, "sms:last:"+req.Phone, time.Now().Format(time.RFC3339), smsFreqLimit)
		model.RDB.Del(c, "sms:fail:"+req.Phone)
	}

	ok(c, gin.H{"message": "验证码已发送"})
}

// verifySMS 校验验证码；生产环境禁止固定码，校验失败记录并锁定
func verifySMS(phone, code string) bool {
	cfg := config.Load()

	// 非生产环境允许固定测试码
	if !cfg.IsProduction() && code == "123456" {
		return true
	}

	if model.RDB == nil {
		return false
	}

	// 失败锁定
	fails, _ := model.RDB.Get(nil, "sms:fail:"+phone).Int()
	if fails >= smsFailLockLimit {
		return false
	}

	saved, err := model.RDB.Get(nil, "sms:"+phone).Result()
	if err != nil || saved != code {
		model.RDB.Incr(nil, "sms:fail:"+phone)
		model.RDB.Expire(nil, "sms:fail:"+phone, 30*time.Minute)
		return false
	}
	model.RDB.Del(nil, "sms:"+phone)
	model.RDB.Del(nil, "sms:fail:"+phone)
	return true
}

// issueToken 签发JWT（HS256）
func issueToken(user *model.User, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   user.ID,
		"user_type": user.UserType,
		"exp":       time.Now().Add(72 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(secret))
}

// PhoneLogin 手机号+验证码登录；角色以数据库为准
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

	// 用户被禁用则拒绝登录
	if user.Status != 1 {
		badRequest(c, "账号已被禁用")
		return
	}

	token, err := issueToken(user, cfg.JWTSecret)
	if err != nil {
		serverError(c, err)
		return
	}

	WriteAudit(c, "user.login", "user", user.ID, gin.H{"phone": req.Phone})
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

	if user.Status != 1 {
		badRequest(c, "账号已被禁用")
		return
	}

	token, err := issueToken(user, cfg.JWTSecret)
	if err != nil {
		serverError(c, err)
		return
	}

	WriteAudit(c, "user.login", "user", user.ID, gin.H{"openid": openid, "is_new": isNew})
	ok(c, gin.H{"token": token, "user": user, "isNew": isNew})
}

// findOrCreateUser 按手机号/OpenID查找用户；不存在则创建
// P0-01 修复：仅允许公共角色(1/2)自注册；管理员/专家必须预创建
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

	// 自注册仅允许甲方/服务方
	if !publicRoles[userType] {
		userType = 1 // 默认甲方；管理员/专家需受控创建
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
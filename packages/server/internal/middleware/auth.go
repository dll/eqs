package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/eqs/server/internal/config"
	"github.com/eqs/server/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken 令牌无效/签名错误/非 HS256
	ErrInvalidToken = errors.New("invalid token")
	// ErrUserNotFound 用户不存在或已删除
	ErrUserNotFound = errors.New("user not found")
	// ErrUserDisabled 用户已被禁用
	ErrUserDisabled = errors.New("user disabled")
)

// ValidateToken 校验 JWT（HS256 白名单）并返回用户身份。
// 供 HTTP 头鉴权与 SSE 查询参数鉴权复用，保证两处规则一致：
//   - 仅接受 HS256；
//   - 用户状态实时查库（禁用/删除的令牌立即失效）；
//   - 角色以数据库为准。
func ValidateToken(cfg *config.Config, tokenStr string) (userID uint, userType int, err error) {
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// P1-10：仅接受 HS256
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || t.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return 0, 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, ErrInvalidToken
	}

	idVal, ok := claims["user_id"].(float64)
	if !ok {
		return 0, 0, ErrInvalidToken
	}

	// P1-10：用户状态实时校验——被禁用/删除的令牌失效
	var user model.User
	if err := model.DB.First(&user, uint(idVal)).Error; err != nil {
		return 0, 0, ErrUserNotFound
	}
	if user.Status != 1 {
		return 0, 0, ErrUserDisabled
	}
	// 角色以数据库为准（防令牌角色与库不一致）
	return user.ID, user.UserType, nil
}

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userID, userType, err := ValidateToken(cfg, tokenStr)
		if err != nil {
			msg := "invalid token"
			if err == ErrUserNotFound {
				msg = "user not found"
			} else if err == ErrUserDisabled {
				msg = "user disabled"
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("user_type", userType)
		c.Next()
	}
}

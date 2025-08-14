package auth

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"

	"teamspeak-one-click-deploy/api"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/users"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var config = DefaultConfig()

// RegisterRoutes 注册认证相关路由
func RegisterRoutes(router *gin.Engine) {
	router.POST(api.LoginPath, loginHandler)
	router.POST(api.LogoutPath, logoutHandler)
	router.GET(api.UserInfoPath, authMiddlewareWithGin(), infoHandler)
}

// loginHandler 处理用户登录
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrBadJSONBody, utils.ErrorMessage(utils.ErrBadJSONBody))
		return
	}

	// 从数据库获取用户
	var user users.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailGin(c, http.StatusUnauthorized, utils.ErrInvalidCredentials, utils.ErrorMessage(utils.ErrInvalidCredentials))
			return
		}
		log.Printf("Error querying user: %v", err)
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		utils.FailGin(c, http.StatusUnauthorized, utils.ErrInvalidCredentials, utils.ErrorMessage(utils.ErrInvalidCredentials))
		return
	}

	// 检查用户状态
	if !user.IsActive() {
		utils.FailGin(c, http.StatusForbidden, utils.ErrAccountDisabled, utils.ErrorMessage(utils.ErrAccountDisabled))
		return
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	if err := database.DB.Save(&user).Error; err != nil {
		log.Printf("Error updating last login time: %v", err)
		// 不返回错误，继续生成令牌
	}

	// 生成 JWT 令牌
	token, expiresAt, err := generateToken(&user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	// 返回登录成功响应
	utils.OKGin(c, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: &UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			LastLogin: *user.LastLogin,
		},
	})
}

// infoHandler 获取当前用户信息
func infoHandler(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, ok := c.Get(string(utils.UserIDKey))
	if !ok || userID == 0 {
		utils.FailGin(c, http.StatusUnauthorized, utils.ErrInvalidToken, utils.ErrorMessage(utils.ErrInvalidToken))
		return
	}

	// 从数据库获取用户信息
	var user users.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailGin(c, http.StatusNotFound, utils.ErrUserNotFound, utils.ErrorMessage(utils.ErrUserNotFound))
			return
		}
		log.Printf("Error fetching user: %v", err)
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	// 返回用户信息
	utils.OKGin(c, UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		LastLogin: *user.LastLogin,
	})
}

// logoutHandler 处理用户登出
func logoutHandler(c *gin.Context) {
	// 在基于 JWT 的系统中，登出通常由前端删除 token 实现
	utils.OKGin(c, map[string]string{"message": "Logout successful"})
}

// generateToken 生成 JWT 令牌
func generateToken(user *users.User) (string, time.Time, error) {
	// 设置令牌过期时间
	expiresAt := time.Now().Add(config.ExpiresIn)

	// 创建声明
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "teamspeak-one-click-deploy",
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ParseToken 解析并验证 JWT 令牌
func ParseToken(tokenString string) (*Claims, error) {
	// 移除 Bearer 前缀
	tokenString = strings.TrimPrefix(tokenString, config.TokenPrefix)

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(utils.ErrorMessage(utils.ErrInvalidToken))
		}
		return []byte(config.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New(utils.ErrorMessage(utils.ErrInvalidToken))
}

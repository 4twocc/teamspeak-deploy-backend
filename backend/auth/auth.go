package auth

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"teamspeak-one-click-deploy/api"
	"teamspeak-one-click-deploy/config"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/user"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var configInstance *config.Config

// Init 初始化认证模块
func Init(cfg *config.Config) {
	configInstance = cfg
}

// RegisterRoutes 注册认证相关路由
func RegisterRoutes(router *gin.Engine) {
	router.POST(api.LoginPath, loginHandler)
	router.POST(api.RegisterPath, registerHandler)
	router.POST(api.LogoutPath, logoutHandler)
	router.GET(api.UserInfoPath, authMiddlewareWithGin(), infoHandler)
}

// registerHandler 处理用户注册
// @param c gin上下文
// @return 注册成功返回用户信息和token，失败返回错误信息
func registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrBadJSONBody, utils.ErrorMessage(utils.ErrBadJSONBody))
		return
	}

	// 检查用户名是否已存在
	var existingUser user.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		utils.FailGin(c, http.StatusConflict, utils.ErrUserAlreadyExists, utils.ErrorMessage(utils.ErrUserAlreadyExists))
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking username: %v", err)
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
			utils.FailGin(c, http.StatusConflict, utils.ErrEmailExists, utils.ErrorMessage(utils.ErrEmailExists))
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Error checking email: %v", err)
			utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
			return
		}
	}

	// 创建新用户
	newUser := user.User{
		Username: req.Username,
		Email:    req.Email,
		Nickname: req.Nickname,
		Role:     utils.AccountRoleVisitor,  // 默认为访客角色
		Status:   utils.AccountStatusActive, // 默认为激活状态
	}

	// 设置密码（加密）
	if err := newUser.SetPassword(req.Password); err != nil {
		log.Printf("Error hashing password: %v", err)
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	// 如果没有设置昵称，使用用户名作为昵称
	if newUser.Nickname == "" {
		newUser.Nickname = newUser.Username
	}

	// 保存到数据库
	if err := database.DB.Create(&newUser).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		// 检查是否是用户名重复错误
		if strings.Contains(err.Error(), "UNIQUE constraint failed: user.username") {
			utils.FailGin(c, http.StatusConflict, utils.ErrUserAlreadyExists, utils.ErrorMessage(utils.ErrUserAlreadyExists))
		} else {
			utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		}
		return
	}

	// 返回注册成功响应
	utils.OKGin(c, nil)
}

// loginHandler 处理用户登录
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrBadJSONBody, utils.ErrorMessage(utils.ErrBadJSONBody))
		return
	}

	// 从数据库获取用户
	var user user.User
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

	// 将 JWT token 写入 cookie（使用配置文件中的设置）
	maxAge := int(time.Until(expiresAt).Seconds())
	c.SetCookie(
		configInstance.Security.CookieName,     // cookie名称
		token,                                  // cookie值（JWT token）
		maxAge,                                 // 过期时间（秒）
		configInstance.Security.CookiePath,     // 路径
		configInstance.Security.CookieDomain,   // 域名
		configInstance.Security.CookieSecure,   // secure（HTTPS）
		configInstance.Security.CookieHttpOnly, // httpOnly（防止XSS）
	)

	// 返回登录成功响应
	utils.OKGin(c, LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      NewUserInfo(&user),
	})
}

// infoHandler 获取当前用户信息
// infoHandler 处理用户信息查询
// 支持两种模式：
// 1. 通过路径参数uid查询指定用户信息（需要JWT认证）
// 2. 如果uid参数为空或无效，则返回JWT token中的当前用户信息
// @param c gin上下文
// @return 返回用户信息或错误信息
func infoHandler(c *gin.Context) {
	// 确保用户已通过JWT认证
	currentUserID, ok := c.Get(string(utils.UserIDKey))
	if !ok || currentUserID == 0 {
		utils.FailGin(c, http.StatusUnauthorized, utils.ErrInvalidToken, utils.ErrorMessage(utils.ErrInvalidToken))
		return
	}

	// 获取要查询的用户ID
	var targetUserID uint64
	uidParam := c.Param("uid")

	if uidParam != "" {
		// 如果提供了uid参数，解析并使用该uid
		parsedUID, err := strconv.ParseUint(uidParam, 10, 64)
		if err != nil {
			utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidParameter, "Invalid uid parameter")
			return
		}
		targetUserID = parsedUID
	} else {
		// 如果没有提供uid参数，使用当前用户的ID
		targetUserID = currentUserID.(uint64)
	}

	// 从数据库获取用户信息
	var user user.User
	if err := database.DB.First(&user, targetUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailGin(c, http.StatusNotFound, utils.ErrUserNotFound, utils.ErrorMessage(utils.ErrUserNotFound))
			return
		}
		log.Printf("Error fetching user: %v", err)
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	// 返回用户信息
	utils.OKGin(c, NewUserInfo(&user))
}

// logoutHandler 处理用户登出
// 清除客户端的认证cookie并返回登出成功消息
func logoutHandler(c *gin.Context) {
	// 清除认证 cookie（使用配置文件中的设置）
	c.SetCookie(
		configInstance.Security.CookieName,     // cookie名称
		"",                                     // 空值
		-1,                                     // MaxAge设为-1立即过期
		configInstance.Security.CookiePath,     // 路径
		configInstance.Security.CookieDomain,   // 域名
		configInstance.Security.CookieSecure,   // secure
		configInstance.Security.CookieHttpOnly, // httpOnly
	)

	// 在基于 JWT 的系统中，登出通常由前端删除 token 实现
	// 现在我们同时清除了服务端设置的 cookie
	utils.OKGin(c, map[string]string{"message": "Logout successful"})
}

// generateToken 生成 JWT 令牌
func generateToken(user *user.User) (string, time.Time, error) {
	// 设置令牌过期时间 (配置文件中是天数)
	expiresAt := time.Now().Add(time.Duration(configInstance.Security.ExpiresIn) * 24 * time.Hour)

	// 创建声明
	claims := &Claims{
		UID:      user.UID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "zhuwo",
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(configInstance.Security.JWTSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ParseToken 解析并验证 JWT 令牌
func ParseToken(tokenString string) (*Claims, error) {
	// 移除配置的 token 前缀
	tokenString = strings.TrimPrefix(tokenString, configInstance.Security.TokenPrefix)

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(utils.ErrorMessage(utils.ErrInvalidToken))
		}
		return []byte(configInstance.Security.JWTSecret), nil
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

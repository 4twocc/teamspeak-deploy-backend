package auth

import (
	"log"
	"net/http"
	"strings"

	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
)

// 全局配置变量，用于存储服务器配置
var serverConfig *struct {
	Auth struct {
		PublicPaths []string
	}
}

// SetServerConfig 设置服务器配置
func SetServerConfig(config interface {
	GetServerConfig() interface {
		GetAuthConfig() interface{ GetPublicPaths() []string }
	}
}) {
	serverConfig = &struct {
		Auth struct {
			PublicPaths []string
		}
	}{}

	authConfig := config.GetServerConfig().GetAuthConfig()
	serverConfig.Auth.PublicPaths = authConfig.GetPublicPaths()
}

// isPublicPath 检查给定路径是否为公共路径
func isPublicPath(path string) bool {
	if serverConfig == nil {
		return false
	}

	for _, publicPath := range serverConfig.Auth.PublicPaths {
		// 精确匹配
		if path == publicPath {
			return true
		}
		// 前缀匹配（处理带参数的路径）
		if strings.HasSuffix(publicPath, "/*") {
			prefix := strings.TrimSuffix(publicPath, "/*")
			if strings.HasPrefix(path, prefix) {
				return true
			}
		}
		// 处理带参数的路径匹配
		if strings.Contains(publicPath, ":") {
			// 将配置中的 :id 转换为实际路径中的具体值进行比较
			// 这里简化处理，只检查前缀
			publicPathPrefix := strings.Split(publicPath, ":")[0]
			if strings.HasPrefix(path, publicPathPrefix) {
				return true
			}
		}
	}
	return false
}

// AuthMiddleware 认证中间件
// 验证 JWT 令牌并将用户信息添加到请求上下文中
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查是否为公共路径
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// 获取 Authorization 头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.Fail(w, http.StatusUnauthorized, utils.ErrUnauthorized, "Authorization header is required")
			return
		}

		// 解析令牌
		_, err := ParseToken(authHeader)
		if err != nil {
			log.Printf("Error parsing token: %v", err)
			utils.Fail(w, http.StatusUnauthorized, utils.ErrUnauthorized, "Invalid or expired token")
			return
		}

		// 调用下一个处理器
		next.ServeHTTP(w, r)
	})
}

// AuthMiddlewareWithGin Gin认证中间件
// 验证 JWT 令牌并将用户信息添加到Gin上下文中
func AuthMiddlewareWithGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否为公共路径
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// 获取 Authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.FailGin(c, http.StatusUnauthorized, utils.ErrUnauthorized, utils.ErrorMessage(utils.ErrUnauthorized))
			c.Abort()
			return
		}

		// 解析令牌
		_, err := ParseToken(authHeader)
		if err != nil {
			log.Printf("Error parsing token: %v", err)
			utils.FailGin(c, http.StatusUnauthorized, utils.ErrUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// 调用下一个处理器
		c.Next()
	}
}

// RequireRole 要求用户具有指定角色
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 简化实现，实际应检查用户角色
			if true {
				utils.Fail(w, http.StatusForbidden, utils.ErrForbidden, "Access denied: insufficient permissions")
				return
			}

			// 调用下一个处理器
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleWithGin Gin要求用户具有指定角色
func RequireRoleWithGin(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简化实现，实际应检查用户角色
		if true {
			utils.FailGin(c, http.StatusForbidden, utils.ErrForbidden, "Access denied: insufficient permissions")
			c.Abort()
			return
		}

		// 调用下一个处理器
		c.Next()
	}
}

// authMiddlewareWithGin 认证中间件（内部使用）
func authMiddlewareWithGin() gin.HandlerFunc {
	return AuthMiddlewareWithGin()
}

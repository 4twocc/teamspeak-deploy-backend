// backend/users/users.go
package user

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"teamspeak-one-click-deploy/api"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var req struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     uint8  `json:"role"`
}

// RegisterRoutes 注册用户管理路由
func RegisterRoutes(router *gin.Engine) {
	router.GET(api.UsersListPath, listHandler)
	router.GET(api.UsersPagePath, usersPagedHandler)
	router.POST(api.UsersAddPath, addHandler)
	router.DELETE(api.UsersRemovePath, removeHandler)
}

// Initialize 初始化用户服务
func Initialize() error {
	// 确保数据库表已创建
	if err := ensureTables(database.DB); err != nil {
		return fmt.Errorf("failed to ensure users table exist: %w", err)
	}
	return nil
}

// ensureTables 确保用户相关的表已创建
func ensureTables(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 自动迁移用户相关表
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to auto migrate users table: %w", err)
	}

	return nil
}

// listHandler 获取用户列表
func listHandler(c *gin.Context) {
	var users []User
	if err := database.DB.Find(&users).Error; err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, utils.ErrorMessage(utils.ErrInternalServer))
		return
	}

	// 不返回密码
	for i := range users {
		users[i].Password = ""
	}

	utils.OKGin(c, users)
}

// usersPagedHandler 分页获取用户列表
func usersPagedHandler(c *gin.Context) {
	q := c.Request.URL.Query()
	page, err := strconv.Atoi(q.Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(q.Get("pageSize"))
	switch {
	case err != nil || pageSize <= 0:
		pageSize = 10
	case pageSize > 100:
		pageSize = 100
	}

	var total int64
	var users []User

	// 获取总数
	if err := database.DB.Model(&User{}).Count(&total).Error; err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Failed to count users")
		return
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := database.DB.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Failed to fetch users")
		return
	}

	// 不返回密码
	for i := range users {
		users[i].Password = ""
	}

	utils.OKGin(c, map[string]any{
		"list":  users,
		"total": total,
		"page":  page,
		"pages": int(math.Ceil(float64(total) / float64(pageSize))),
	})
}

// addHandler 添加用户
func addHandler(c *gin.Context) {
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrBadJSONBody, "Invalid request body")
		return
	}

	// 验证必填字段
	if req.Username == "" || req.Password == "" || req.Email == "" {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrMissingParameter, "Username, password and email are required")
		return
	}

	// 验证角色
	if req.Role == 0 {
		req.Role = utils.AccountStatusVisitor // 默认角色
	} else if req.Role != utils.AccountStatusAdmin && req.Role != utils.AccountStatusOperator {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInvalidRole, "Invalid role, must be 'admin' or 'operator'")
		return
	}

	// 检查用户名是否已存在
	var existingUser User
	err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		utils.FailGin(c, http.StatusConflict, utils.ErrUserAlreadyExists, "Username already exists")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrUserAlreadyExists, "Failed to check username availability")
		return
	}

	// 创建用户
	user := User{
		Username: req.Username,
		Email:    req.Email,
		Role:     req.Role,
		Status:   utils.AccountStatusActive,
	}

	if err := user.SetPassword(req.Password); err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Failed to set password")
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Failed to create user")
		return
	}

	// 不返回密码
	user.Password = ""
	utils.OKGin(c, user)
}

// removeHandler 删除用户
func removeHandler(c *gin.Context) {
	idStr := c.Query("id")
	if idStr == "" {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrMissingParameter, "User ID is required")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.FailGin(c, http.StatusBadRequest, utils.ErrInternalServer, "Invalid user ID format")
		return
	}

	// 检查用户是否存在
	var user User
	if err := database.DB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.FailGin(c, http.StatusNotFound, utils.ErrUserNotFound, "User not found")
		} else {
			utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Failed to find user")
		}
		return
	}

	// 删除用户
	if err := database.DB.Delete(&user).Error; err != nil {
		utils.FailGin(c, http.StatusInternalServerError, utils.ErrInternalServer, "Failed to delete user")
		return
	}

	utils.OKGin(c, map[string]any{"message": "User deleted successfully"})
}

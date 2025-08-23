package user

import (
	"teamspeak-one-click-deploy/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	UID       uint           `json:"uid" gorm:"primaryKey;autoIncrement"`
	Username  string         `json:"username" gorm:"size:50;uniqueIndex;not null"`
	Nickname  string         `json:"nickname" gorm:"size:50"`
	Password  string         `json:"-" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"size:100"`
	Role      uint8          `json:"role" gorm:"size:20;default:8"`
	Status    uint8          `json:"status" gorm:"size:20;default:0"`
	LastLogin *time.Time     `json:"last_login,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前的钩子
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return
}

// BeforeUpdate 更新前的钩子
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	u.UpdatedAt = time.Now()
	return
}

// SetPassword 设置密码（加密）
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsAdmin 检查是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == utils.AccountRoleAdmin
}

// IsAdmin 检查是否是运营者
func (u *User) IsOperator() bool {
	return u.Role == utils.AccountRoleOperator
}

// IsActive 检查用户是否激活
func (u *User) IsActive() bool {
	return u.Status == utils.AccountStatusActive
}

// IsBanned 检查用户是否禁用
func (u *User) IsBanned() bool {
	return u.Status == utils.AccountStatusBanned
}

// IsLocked 检查用户是否锁定
func (u *User) IsLocked() bool {
	return u.Status == utils.AccountStatusLocked
}

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UsersAuthDao is the data access object for the table users_auth.
type UsersAuthDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  UsersAuthColumns   // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// UsersAuthColumns defines and stores column names for the table users_auth.
type UsersAuthColumns struct {
	UserId                string // ID=users.id
	PasswordHash          string // argon2id/bcrypt
	PasswordAlgo          string //
	MfaEnabled            string // MFA
	MfaSecret             string // TOTPKMS
	LastPasswordChangedAt string //
	CreatedAt             string //
	UpdatedAt             string //
}

// usersAuthColumns holds the columns for the table users_auth.
var usersAuthColumns = UsersAuthColumns{
	UserId:                "user_id",
	PasswordHash:          "password_hash",
	PasswordAlgo:          "password_algo",
	MfaEnabled:            "mfa_enabled",
	MfaSecret:             "mfa_secret",
	LastPasswordChangedAt: "last_password_changed_at",
	CreatedAt:             "created_at",
	UpdatedAt:             "updated_at",
}

// NewUsersAuthDao creates and returns a new DAO object for table data access.
func NewUsersAuthDao(handlers ...gdb.ModelHandler) *UsersAuthDao {
	return &UsersAuthDao{
		group:    "default",
		table:    "users_auth",
		columns:  usersAuthColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UsersAuthDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UsersAuthDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UsersAuthDao) Columns() UsersAuthColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UsersAuthDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UsersAuthDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *UsersAuthDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

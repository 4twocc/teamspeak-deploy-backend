// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// LoginAuditDao is the data access object for the table login_audit.
type LoginAuditDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  LoginAuditColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// LoginAuditColumns defines and stores column names for the table login_audit.
type LoginAuditColumns struct {
	Id        string //
	UserId    string // ID
	Identity  string // //
	Provider  string //
	IpAddress string // IPIPv4/IPv6
	UserAgent string // UA
	Success   string //
	ErrorCode string //
	CreatedAt string //
}

// loginAuditColumns holds the columns for the table login_audit.
var loginAuditColumns = LoginAuditColumns{
	Id:        "id",
	UserId:    "user_id",
	Identity:  "identity",
	Provider:  "provider",
	IpAddress: "ip_address",
	UserAgent: "user_agent",
	Success:   "success",
	ErrorCode: "error_code",
	CreatedAt: "created_at",
}

// NewLoginAuditDao creates and returns a new DAO object for table data access.
func NewLoginAuditDao(handlers ...gdb.ModelHandler) *LoginAuditDao {
	return &LoginAuditDao{
		group:    "default",
		table:    "login_audit",
		columns:  loginAuditColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *LoginAuditDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *LoginAuditDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *LoginAuditDao) Columns() LoginAuditColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *LoginAuditDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *LoginAuditDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *LoginAuditDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// VerificationTokensDao is the data access object for the table verification_tokens.
type VerificationTokensDao struct {
	table    string                    // table is the underlying table name of the DAO.
	group    string                    // group is the database configuration group name of the current DAO.
	columns  VerificationTokensColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler        // handlers for customized model modification.
}

// VerificationTokensColumns defines and stores column names for the table verification_tokens.
type VerificationTokensColumns struct {
	Id         string //
	UserId     string // ID=users.id
	TokenType  string //
	TokenHash  string // SHA-256
	ExpiresAt  string //
	ConsumedAt string // NULL
	Metadata   string //
	CreatedAt  string //
}

// verificationTokensColumns holds the columns for the table verification_tokens.
var verificationTokensColumns = VerificationTokensColumns{
	Id:         "id",
	UserId:     "user_id",
	TokenType:  "token_type",
	TokenHash:  "token_hash",
	ExpiresAt:  "expires_at",
	ConsumedAt: "consumed_at",
	Metadata:   "metadata",
	CreatedAt:  "created_at",
}

// NewVerificationTokensDao creates and returns a new DAO object for table data access.
func NewVerificationTokensDao(handlers ...gdb.ModelHandler) *VerificationTokensDao {
	return &VerificationTokensDao{
		group:    "default",
		table:    "verification_tokens",
		columns:  verificationTokensColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *VerificationTokensDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *VerificationTokensDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *VerificationTokensDao) Columns() VerificationTokensColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *VerificationTokensDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *VerificationTokensDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *VerificationTokensDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

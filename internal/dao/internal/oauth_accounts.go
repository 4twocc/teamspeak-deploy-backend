// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OauthAccountsDao is the data access object for the table oauth_accounts.
type OauthAccountsDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  OauthAccountsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// OauthAccountsColumns defines and stores column names for the table oauth_accounts.
type OauthAccountsColumns struct {
	Id                string //
	UserId            string // ID=users.id
	Provider          string // github/google/wechat
	ProviderAccountId string // ID
	AccessToken       string //
	RefreshToken      string //
	TokenType         string //
	ExpiresAt         string //
	Scope             string // Scope
	ProfileJson       string //
	CreatedAt         string //
	UpdatedAt         string //
}

// oauthAccountsColumns holds the columns for the table oauth_accounts.
var oauthAccountsColumns = OauthAccountsColumns{
	Id:                "id",
	UserId:            "user_id",
	Provider:          "provider",
	ProviderAccountId: "provider_account_id",
	AccessToken:       "access_token",
	RefreshToken:      "refresh_token",
	TokenType:         "token_type",
	ExpiresAt:         "expires_at",
	Scope:             "scope",
	ProfileJson:       "profile_json",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewOauthAccountsDao creates and returns a new DAO object for table data access.
func NewOauthAccountsDao(handlers ...gdb.ModelHandler) *OauthAccountsDao {
	return &OauthAccountsDao{
		group:    "default",
		table:    "oauth_accounts",
		columns:  oauthAccountsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OauthAccountsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OauthAccountsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OauthAccountsDao) Columns() OauthAccountsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OauthAccountsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OauthAccountsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *OauthAccountsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

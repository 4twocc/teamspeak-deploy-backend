package utils

// ValidationError 表示验证错误
type ValidationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// 错误码定义
const (
	// 通用
	Success    = 0
	ErrUnknown = 1000 + iota
	ErrInvalidRequest
	ErrInvalidParameter
	ErrUnauthorized
	ErrForbidden
	ErrNotFound
	ErrMethodNotAllowed
	ErrBadJSONBody
	ErrMissingParameter
	ErrInternalServer
	ErrServiceUnavailable
	ErrTooManyRequests
	ErrRequestTimeout
	ErrInvalidRole
	ErrUserAlreadyExists
	ErrUserNotFound
	ErrAccountDisabled
	ErrEmailExists

	// 数据库相关错误
	ErrDatabase = 2000 + iota
	ErrDatabaseConnection
	ErrDatabaseQuery

	// 认证相关错误
	ErrAuth = 3000 + iota
	ErrInvalidCredentials
	ErrTokenExpired
	ErrInvalidToken

	// 业务相关错误
	ErrBusiness = 4000 + iota
	ErrTSInstanceIsNil
	ErrTSInstanceDirEmpty
	ErrTSInstancePortConflict

	// 监控相关错误
	ErrMonitor = 5000 + iota
	ErrCollectorInit
	ErrNoSystemMetrics
	ErrNoBusinessMetrics
	ErrCollectStatsFailed
	ErrInvalidDuration
	ErrDurationTooLong

	// 部署相关错误
	ErrDeploy = 6000 + iota
	ErrDeployInProgress
)

// ErrorMessage 根据错误码返回错误信息
func ErrorMessage(code int) string {
	messages := map[int]string{
		// 通用错误
		Success:               "success",
		ErrUnknown:            "unknown error",
		ErrInvalidRequest:     "invalid request",
		ErrInvalidParameter:   "invalid parameter",
		ErrUnauthorized:       "unauthorized",
		ErrForbidden:          "forbidden",
		ErrNotFound:           "not found",
		ErrMethodNotAllowed:   "method not allowed",
		ErrBadJSONBody:        "bad JSON body",
		ErrMissingParameter:   "missing parameter",
		ErrInternalServer:     "internal server error",
		ErrServiceUnavailable: "service unavailable",
		ErrTooManyRequests:    "too many requests",
		ErrRequestTimeout:     "request timeout",
		ErrInvalidRole:        "invalid role",
		ErrUserAlreadyExists:  "user already exists",
		ErrUserNotFound:       "user not found",
		ErrAccountDisabled:    "account disabled",
		ErrEmailExists:        "email already exists",

		// 数据库相关错误
		ErrDatabase:           "database error",
		ErrDatabaseConnection: "database connection error",
		ErrDatabaseQuery:      "database query error",

		// 认证相关错误
		ErrAuth:               "authentication error",
		ErrInvalidCredentials: "invalid credentials",
		ErrTokenExpired:       "token expired",
		ErrInvalidToken:       "invalid token",

		// 业务相关错误
		ErrTSInstanceIsNil:        "teamspeak instance is nil",
		ErrTSInstanceDirEmpty:     "teamspeak instance directory is empty",
		ErrTSInstancePortConflict: "teamspeak instance port conflict",

		// 监控相关错误
		ErrCollectorInit:      "collector initialization failed",
		ErrNoSystemMetrics:    "no system metrics available",
		ErrNoBusinessMetrics:  "no business metrics available",
		ErrCollectStatsFailed: "failed to collect statistics",
		ErrInvalidDuration:    "invalid duration parameter",
		ErrDurationTooLong:    "duration parameter too long",

		// 部署相关错误
		ErrDeploy:           "deployment error",
		ErrDeployInProgress: "deployment already in progress",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}
	return messages[ErrUnknown]
}

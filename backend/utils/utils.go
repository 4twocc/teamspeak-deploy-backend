// backend/utils/utils.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse 统一的HTTP响应结构
// code: 0 表示成功，非0表示业务错误码
// message: 可读的消息
// data: 具体数据负载
// timestamp: 服务端时间（UTC）
type APIResponse struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Data      any       `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// PaginationMeta 列表型响应的分页元信息（可按需扩展）
type PaginationMeta struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	Total    int `json:"total,omitempty"`
	Pages    int `json:"pages,omitempty"`
}

// WriteJSON 写出统一响应 (net/http)
func WriteJSON(w http.ResponseWriter, status int, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	})
}

// OK 成功响应（HTTP 200, 业务码 0）(net/http)
func OK(w http.ResponseWriter, data any) {
	WriteJSON(w, http.StatusOK, Success, "ok", data)
}

// OKGin 成功响应（HTTP 200, 业务码 0）(gin)
func OKGin(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:      Success,
		Message:   "ok",
		Data:      data,
		Timestamp: time.Now().UTC(),
	})
}

// PaginatedOK 成功分页响应（HTTP 200, 业务码 0，data 为对象，包含 list 与 meta）(net/http)
func PaginatedOK(w http.ResponseWriter, list any, meta PaginationMeta) {
	payload := map[string]any{
		"data": list,
		"meta": meta,
	}
	WriteJSON(w, http.StatusOK, Success, "ok", payload)
}

// Fail 失败响应（自定义HTTP状态与业务错误码）(net/http)
func Fail(w http.ResponseWriter, httpStatus int, code int, message string) {
	WriteJSON(w, httpStatus, code, message, nil)
}

// FailGin 失败响应（自定义HTTP状态与业务错误码）(gin)
func FailGin(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, APIResponse{
		Code:      code,
		Message:   message,
		Data:      nil,
		Timestamp: time.Now().UTC(),
	})
}

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// FormatDuration 将秒数格式化为易读的持续时间
func FormatDuration(seconds uint64) string {
	duration := time.Duration(seconds) * time.Second
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds = uint64(duration.Seconds()) % 60

	result := ""
	if days > 0 {
		result += fmt.Sprintf("%d天", days)
	}
	if hours > 0 || days > 0 {
		result += fmt.Sprintf("%d小时", hours)
	}
	if minutes > 0 || hours > 0 || days > 0 {
		result += fmt.Sprintf("%d分钟", minutes)
	}
	result += fmt.Sprintf("%d秒", seconds)
	return result
}

// ContainsString 检查字符串切片中是否包含指定字符串
func ContainsString(slice []string, str string) bool {
	return slices.Contains(slice, str)
}

// MapToQueryString 将 map 转换为 URL 查询字符串
func MapToQueryString(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}

	var query string
	for k, v := range params {
		query += k + "=" + v + "&"
	}
	return query[:len(query)-1] // 移除最后一个 &
}

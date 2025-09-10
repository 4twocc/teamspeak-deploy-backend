package instance

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"teamspeak-one-click-deploy/api"
	"teamspeak-one-click-deploy/database"
	"teamspeak-one-click-deploy/utils"

	"github.com/gin-gonic/gin"
)

// 定义 context key 类型，避免使用内置字符串类型作为 key
type contextKey string

const (
	idKey contextKey = "id"
)

// Handler 实例管理 HTTP 处理器
type Handler struct {
	service *Service
}

// NewHandler 创建新的实例处理器
func NewHandler() *Handler {
	return &Handler{
		service: NewService(database.DB, NewAlertManager(database.DB, nil)),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// 实例管理
	router.GET(api.InstancesListPath, h.listInstancesGin)
	router.POST(api.InstancesCreatePath, h.createInstanceGin)
	router.GET(api.InstancesDetailPath, h.getInstanceGin)
	router.PUT(api.InstancesUpdatePath, h.updateInstanceGin)
	router.DELETE(api.InstancesDeletePath, h.deleteInstanceGin)
	router.POST(api.InstancesStartPath, h.startInstanceGin)
	router.POST(api.InstancesStopPath, h.stopInstanceGin)
	router.POST(api.InstancesRestartPath, h.restartInstanceGin)
	router.GET(api.InstancesLogsPath, h.getInstanceLogsGin)
	router.GET(api.InstancesResourcesPath, h.getInstanceResourcesGin)
}

// 基于 gin 的 handler 方法
func (h *Handler) listInstancesGin(c *gin.Context) {
	h.listInstances(c.Writer, c.Request)
}

func (h *Handler) createInstanceGin(c *gin.Context) {
	h.createInstance(c.Writer, c.Request)
}

func (h *Handler) getInstanceGin(c *gin.Context) {
	// 构造一个新的请求，将 gin 的参数注入到请求上下文中
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.getInstance(c.Writer, req)
}

func (h *Handler) updateInstanceGin(c *gin.Context) {
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.updateInstance(c.Writer, req)
}

func (h *Handler) deleteInstanceGin(c *gin.Context) {
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.deleteInstance(c.Writer, req)
}

func (h *Handler) startInstanceGin(c *gin.Context) {
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.startInstance(c.Writer, req)
}

func (h *Handler) stopInstanceGin(c *gin.Context) {
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.stopInstance(c.Writer, req)
}

func (h *Handler) restartInstanceGin(c *gin.Context) {
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.restartInstance(c.Writer, req)
}

func (h *Handler) getInstanceLogsGin(c *gin.Context) {
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.getInstanceLogs(c.Writer, req)
}

func (h *Handler) getInstanceResourcesGin(c *gin.Context) {
	req := c.Request.WithContext(context.WithValue(c.Request.Context(), idKey, c.Param("id")))
	h.getInstanceResources(c.Writer, req)
}

// getInstanceResources 获取实例资源使用情况
// @Summary 获取实例资源使用情况
// @Description 获取指定 TeamSpeak 实例的 CPU、内存、磁盘与网络等资源使用情况
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 200 {object} utils.APIResponse{data=ResourceUsage} "成功获取资源使用情况"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在或未运行"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id}/resources [get]
func (h *Handler) getInstanceResources(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 获取实例资源使用情况
	resources, err := h.service.GetInstanceResources(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get resources for instance %s: %v", id, err)
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInstanceNotFound) {
			status = http.StatusNotFound
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	utils.OK(w, resources)
}

// listInstances 获取实例列表
// @Summary 获取实例列表
// @Description 分页获取 TeamSpeak 实例列表
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param page query int false "页码，默认为1" default(1)
// @Param page_size query int false "每页条数，默认为10，最大100" default(10)
// @Success 200 {object} utils.APIResponse{data=[]Instance} "成功获取实例列表"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances [get]
func (h *Handler) listInstances(w http.ResponseWriter, r *http.Request) {
	// 解析分页参数
	page := 1
	pageSize := 10
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			if ps > 100 {
				ps = 100 // 限制最大100条
			}
			pageSize = ps
		}
	}

	// 构造查询参数
	filter := &InstanceFilter{
		Page:     page,
		PageSize: pageSize,
	}

	// 获取实例列表
	instances, total, err := h.service.ListInstances(r.Context(), filter)
	if err != nil {
		log.Printf("Failed to list instances: %v", err)
		utils.Fail(w, http.StatusInternalServerError, utils.ErrInternalServer, err.Error())
		return
	}

	meta := utils.PaginationMeta{
		Page:     page,
		PageSize: pageSize,
		Total:    int(total),
		Pages:    (int(total) + pageSize - 1) / pageSize,
	}

	utils.PaginatedOK(w, instances, meta)
}

// createInstance 创建实例
// @Summary 创建实例
// @Description 创建一个新的 TeamSpeak 实例
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param instance body CreateInstanceInput true "实例创建信息"
// @Success 201 {object} utils.APIResponse{data=Instance} "成功创建实例"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 409 {object} utils.APIResponse "实例已存在"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances [post]
func (h *Handler) createInstance(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var input CreateInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		utils.Fail(w, http.StatusBadRequest, utils.ErrBadJSONBody, "Invalid request body")
		return
	}

	// 创建实例
	instance, err := h.service.CreateInstance(r.Context(), &input)
	if err != nil {
		log.Printf("Failed to create instance: %v", err)
		status := http.StatusInternalServerError
		switch {
		case errors.As(err, new(*utils.ValidationError)):
			status = http.StatusBadRequest
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	utils.OK(w, instance)
}

// getInstance 获取实例详情
// @Summary 获取实例详情
// @Description 根据ID获取 TeamSpeak 实例详情
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 200 {object} utils.APIResponse{data=Instance} "成功获取实例详情"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id} [get]
func (h *Handler) getInstance(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 获取实例详情
	instance, err := h.service.GetInstance(r.Context(), id)
	if err != nil {
		log.Printf("Failed to get instance %s: %v", id, err)
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInstanceNotFound) {
			status = http.StatusNotFound
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	utils.OK(w, instance)
}

// updateInstance 更新实例
// @Summary 更新实例
// @Description 根据ID更新 TeamSpeak 实例信息
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Param instance body UpdateInstanceInput true "实例更新信息"
// @Success 200 {object} utils.APIResponse{data=Instance} "成功更新实例"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id} [put]
func (h *Handler) updateInstance(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 解析请求体
	var input UpdateInstanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		utils.Fail(w, http.StatusBadRequest, utils.ErrBadJSONBody, "Invalid request body")
		return
	}

	// 更新实例
	instance, err := h.service.UpdateInstance(r.Context(), id, &input)
	if err != nil {
		log.Printf("Failed to update instance %s: %v", id, err)
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrInstanceNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrInvalidStatus):
			status = http.StatusBadRequest
		case errors.As(err, new(*utils.ValidationError)):
			status = http.StatusBadRequest
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	utils.OK(w, instance)
}

// deleteInstance 删除实例
// @Summary 删除实例
// @Description 根据ID删除 TeamSpeak 实例
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 204 {object} utils.APIResponse "成功删除实例"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id} [delete]
func (h *Handler) deleteInstance(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 删除实例
	err := h.service.DeleteInstance(r.Context(), id)
	if err != nil {
		log.Printf("Failed to delete instance %s: %v", id, err)
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrInstanceNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrOperationNotAllowed):
			status = http.StatusForbidden
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// startInstance 启动实例
// @Summary 启动实例
// @Description 启动指定的 TeamSpeak 实例
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 200 {object} utils.APIResponse{data=Instance} "成功启动实例"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在"
// @Failure 409 {object} utils.APIResponse "实例状态冲突"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id}/start [post]
func (h *Handler) startInstance(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 启动实例
	err := h.service.StartInstance(r.Context(), id)
	if err != nil {
		log.Printf("Failed to start instance %s: %v", id, err)
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrInstanceNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrOperationNotAllowed):
			status = http.StatusForbidden
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// stopInstance 停止实例
// @Summary 停止实例
// @Description 停止指定的 TeamSpeak 实例
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 200 {object} utils.APIResponse{data=Instance} "成功停止实例"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在"
// @Failure 409 {object} utils.APIResponse "实例状态冲突"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id}/stop [post]
func (h *Handler) stopInstance(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 停止实例
	err := h.service.StopInstance(r.Context(), id)
	if err != nil {
		log.Printf("Failed to stop instance %s: %v", id, err)
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrInstanceNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrOperationNotAllowed):
			status = http.StatusForbidden
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// restartInstance 重启实例
// @Summary 重启实例
// @Description 重启指定的 TeamSpeak 实例
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Success 200 {object} utils.APIResponse{data=Instance} "成功重启实例"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在"
// @Failure 409 {object} utils.APIResponse "实例状态冲突"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id}/restart [post]
func (h *Handler) restartInstance(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 重启实例
	err := h.service.RestartInstance(r.Context(), id)
	if err != nil {
		log.Printf("Failed to restart instance %s: %v", id, err)
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrInstanceNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrOperationNotAllowed):
			status = http.StatusForbidden
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// getInstanceLogs 获取实例日志
// @Summary 获取实例日志
// @Description 获取指定 TeamSpeak 实例的最新日志记录
// @Tags 实例管理
// @Accept json
// @Produce json
// @Param id path string true "实例ID"
// @Param limit query int false "返回日志条数，默认为100，最大1000" minimum(1) maximum(1000)
// @Success 200 {object} utils.APIResponse{data=[]InstanceLog} "成功获取实例日志"
// @Failure 400 {object} utils.APIResponse "请求参数错误"
// @Failure 401 {object} utils.APIResponse "未授权"
// @Failure 404 {object} utils.APIResponse "实例不存在"
// @Failure 500 {object} utils.APIResponse "服务器内部错误"
// @Security BearerAuth
// @Router /instances/{id}/logs [get]
func (h *Handler) getInstanceLogs(w http.ResponseWriter, r *http.Request) {
	// 获取实例ID
	id := r.Context().Value(idKey).(string)
	if id == "" {
		utils.Fail(w, http.StatusBadRequest, utils.ErrMissingParameter, "Instance ID is required")
		return
	}

	// 解析 limit 参数
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 1000 {
				l = 1000 // 限制最大1000条
			}
			limit = l
		}
	}

	// 获取实例日志
	logs, err := h.service.GetInstanceLogs(r.Context(), id, limit)
	if err != nil {
		// 这里使用标准log，因为handler层通常不直接持有logService
		log.Printf("Failed to get logs for instance %s: %v", id, err)
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInstanceNotFound) {
			status = http.StatusNotFound
		}
		utils.Fail(w, status, utils.ErrInternalServer, err.Error())
		return
	}

	utils.OK(w, logs)
}

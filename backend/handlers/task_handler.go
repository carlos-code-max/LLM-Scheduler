package handlers

import (
	"strconv"

	"llm-scheduler/models"
	"llm-scheduler/services"
	"llm-scheduler/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *services.TaskService
	logger      *logrus.Logger
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *services.TaskService, logger *logrus.Logger) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		logger:      logger,
	}
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req models.TaskCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	// 设置默认优先级
	if req.Priority == 0 {
		req.Priority = models.TaskPriorityMedium
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create task")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "任务创建成功", task)
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	task, err := h.taskService.GetTask(id)
	if err != nil {
		if err.Error() == "task not found" {
			utils.NotFound(c, "任务不存在")
			return
		}
		h.logger.WithError(err).Error("Failed to get task")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, task)
}

// ListTasks 获取任务列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	var req models.TaskListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大页面大小
	}

	tasks, total, err := h.taskService.ListTasks(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list tasks")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessPaged(c, tasks, total, req.Page, req.PageSize)
}

// UpdateTask 更新任务
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req models.TaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	task, err := h.taskService.UpdateTask(id, &req)
	if err != nil {
		if err.Error() == "task not found" {
			utils.NotFound(c, "任务不存在")
			return
		}
		h.logger.WithError(err).Error("Failed to update task")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "任务更新成功", task)
}

// CancelTask 取消任务
func (h *TaskHandler) CancelTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.taskService.CancelTask(c.Request.Context(), id); err != nil {
		if err.Error() == "task not found" {
			utils.NotFound(c, "任务不存在")
			return
		}
		h.logger.WithError(err).Error("Failed to cancel task")
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "任务已取消", nil)
}

// RetryTask 重试任务
func (h *TaskHandler) RetryTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.taskService.RetryTask(c.Request.Context(), id); err != nil {
		if err.Error() == "task not found" {
			utils.NotFound(c, "任务不存在")
			return
		}
		h.logger.WithError(err).Error("Failed to retry task")
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "任务已重新提交", nil)
}

// GetTaskStats 获取任务统计
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	stats, err := h.taskService.GetTaskStats()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get task stats")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

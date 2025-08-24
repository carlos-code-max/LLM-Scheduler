package handlers

import (
	"strconv"

	"llm-scheduler/services"
	"llm-scheduler/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	statsService *services.StatsService
	logger       *logrus.Logger
}

// NewStatsHandler 创建统计处理器
func NewStatsHandler(statsService *services.StatsService, logger *logrus.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		logger:       logger,
	}
}

// GetDashboardStats 获取 Dashboard 统计数据
func (h *StatsHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.statsService.GetDashboardStats()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get dashboard stats")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetTaskStatsByDate 按日期获取任务统计
func (h *StatsHandler) GetTaskStatsByDate(c *gin.Context) {
	daysStr := c.Query("days")
	days := 7 // 默认7天
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	stats, err := h.statsService.GetTaskStatsByDate(days)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get task stats by date")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetTaskStatsByModel 按模型获取任务统计
func (h *StatsHandler) GetTaskStatsByModel(c *gin.Context) {
	stats, err := h.statsService.GetTaskStatsByModel()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get task stats by model")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetTaskStatsByType 按任务类型获取统计
func (h *StatsHandler) GetTaskStatsByType(c *gin.Context) {
	stats, err := h.statsService.GetTaskStatsByType()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get task stats by type")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

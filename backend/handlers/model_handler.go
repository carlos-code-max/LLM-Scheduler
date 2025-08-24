package handlers

import (
	"fmt"
	"strconv"

	"llm-scheduler/models"
	"llm-scheduler/services"
	"llm-scheduler/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ModelHandler 模型处理器
type ModelHandler struct {
	modelService *services.ModelService
	logger       *logrus.Logger
}

// NewModelHandler 创建模型处理器
func NewModelHandler(modelService *services.ModelService, logger *logrus.Logger) *ModelHandler {
	return &ModelHandler{
		modelService: modelService,
		logger:       logger,
	}
}

// CreateModel 创建模型
func (h *ModelHandler) CreateModel(c *gin.Context) {
	var model models.Model
	if err := c.ShouldBindJSON(&model); err != nil {
		utils.ValidationError(c, err)
		return
	}

	// 验证必填字段
	if model.Name == "" {
		utils.BadRequest(c, "模型名称不能为空")
		return
	}
	if model.Type == "" {
		utils.BadRequest(c, "模型类型不能为空")
		return
	}
	if model.Config == nil {
		model.Config = make(models.ModelConfig)
	}

	createdModel, err := h.modelService.CreateModel(&model)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create model")
		if err.Error() == fmt.Sprintf("model with name '%s' already exists", model.Name) {
			utils.BadRequest(c, err.Error())
			return
		}
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "模型创建成功", createdModel)
}

// GetModel 获取模型详情
func (h *ModelHandler) GetModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的模型ID")
		return
	}

	model, err := h.modelService.GetModel(id)
	if err != nil {
		if err.Error() == "model not found" {
			utils.NotFound(c, "模型不存在")
			return
		}
		h.logger.WithError(err).Error("Failed to get model")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, model)
}

// ListModels 获取模型列表
func (h *ModelHandler) ListModels(c *gin.Context) {
	var modelType *models.ModelType
	var status *models.ModelStatus

	if typeStr := c.Query("type"); typeStr != "" {
		mt := models.ModelType(typeStr)
		modelType = &mt
	}

	if statusStr := c.Query("status"); statusStr != "" {
		ms := models.ModelStatus(statusStr)
		status = &ms
	}

	models_list, err := h.modelService.ListModels(modelType, status)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list models")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, models_list)
}

// UpdateModel 更新模型
func (h *ModelHandler) UpdateModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的模型ID")
		return
	}

	var updates models.Model
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ValidationError(c, err)
		return
	}

	model, err := h.modelService.UpdateModel(id, &updates)
	if err != nil {
		if err.Error() == "model not found" {
			utils.NotFound(c, "模型不存在")
			return
		}
		h.logger.WithError(err).Error("Failed to update model")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "模型更新成功", model)
}

// DeleteModel 删除模型
func (h *ModelHandler) DeleteModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的模型ID")
		return
	}

	if err := h.modelService.DeleteModel(id); err != nil {
		h.logger.WithError(err).Error("Failed to delete model")
		utils.BadRequest(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "模型删除成功", nil)
}

// UpdateModelStatus 更新模型状态
func (h *ModelHandler) UpdateModelStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		utils.BadRequest(c, "无效的模型ID")
		return
	}

	var req struct {
		Status models.ModelStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ValidationError(c, err)
		return
	}

	if err := h.modelService.UpdateModelStatus(id, req.Status); err != nil {
		h.logger.WithError(err).Error("Failed to update model status")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "模型状态更新成功", nil)
}

// GetModelStats 获取模型统计
func (h *ModelHandler) GetModelStats(c *gin.Context) {
	stats, err := h.modelService.GetModelStats()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get model stats")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

// GetAvailableModels 获取可用模型
func (h *ModelHandler) GetAvailableModels(c *gin.Context) {
	models_list, err := h.modelService.GetAvailableModels()
	if err != nil {
		h.logger.WithError(err).Error("Failed to get available models")
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.Success(c, models_list)
}

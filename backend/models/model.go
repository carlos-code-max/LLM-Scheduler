package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ModelType 模型类型枚举
type ModelType string

const (
	ModelTypeOpenAI ModelType = "openai"
	ModelTypeLocal  ModelType = "local"
	ModelTypeCustom ModelType = "custom"
)

// ModelStatus 模型状态枚举
type ModelStatus string

const (
	ModelStatusOnline      ModelStatus = "online"
	ModelStatusOffline     ModelStatus = "offline"
	ModelStatusMaintenance ModelStatus = "maintenance"
)

// ModelConfig 模型配置，存储为 JSON
type ModelConfig map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (mc *ModelConfig) Scan(value interface{}) error {
	if value == nil {
		*mc = make(ModelConfig)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal ModelConfig: %v", value)
	}

	return json.Unmarshal(bytes, mc)
}

// Value 实现 driver.Valuer 接口
func (mc ModelConfig) Value() (driver.Value, error) {
	if mc == nil {
		return nil, nil
	}
	return json.Marshal(mc)
}

// Model 模型表结构
type Model struct {
	ID              uint64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string      `json:"name" gorm:"type:varchar(255);uniqueIndex;not null"`
	Type            ModelType   `json:"type" gorm:"type:enum('openai','local','custom');not null"`
	Config          ModelConfig `json:"config" gorm:"type:json;not null"`
	Status          ModelStatus `json:"status" gorm:"type:enum('online','offline','maintenance');default:offline"`
	MaxWorkers      int         `json:"max_workers" gorm:"default:1"`
	CurrentWorkers  int         `json:"current_workers" gorm:"default:0"`
	TotalRequests   uint64      `json:"total_requests" gorm:"default:0"`
	SuccessRequests uint64      `json:"success_requests" gorm:"default:0"`
	CreatedAt       time.Time   `json:"created_at"`
	Updated         time.Time   `json:"updated_at"`

	// 关联关系
	Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:ModelID"`
}

// TableName 指定表名
func (Model) TableName() string {
	return "models"
}

// GetSuccessRate 计算成功率
func (m *Model) GetSuccessRate() float64 {
	if m.TotalRequests == 0 {
		return 0
	}
	return float64(m.SuccessRequests) / float64(m.TotalRequests) * 100
}

// IsAvailable 检查模型是否可用
func (m *Model) IsAvailable() bool {
	return m.Status == ModelStatusOnline && m.CurrentWorkers < m.MaxWorkers
}

// GetConfigValue 获取配置值
func (m *Model) GetConfigValue(key string) (interface{}, bool) {
	value, exists := m.Config[key]
	return value, exists
}

// SetConfigValue 设置配置值
func (m *Model) SetConfigValue(key string, value interface{}) {
	if m.Config == nil {
		m.Config = make(ModelConfig)
	}
	m.Config[key] = value
}

// BeforeCreate GORM 钩子：创建前
func (m *Model) BeforeCreate(tx *gorm.DB) error {
	if m.Config == nil {
		m.Config = make(ModelConfig)
	}
	return nil
}

// BeforeUpdate GORM 钩子：更新前
func (m *Model) BeforeUpdate(tx *gorm.DB) error {
	m.Updated = time.Now()
	return nil
}

// ModelStats 模型统计信息
type ModelStats struct {
	Model
	PendingTasks  int64   `json:"pending_tasks"`
	RunningTasks  int64   `json:"running_tasks"`
	SuccessRate   float64 `json:"success_rate"`
	AvgResponseMs int64   `json:"avg_response_ms"`
}

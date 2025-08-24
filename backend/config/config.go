package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Queue    QueueConfig    `mapstructure:"queue"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Models   ModelsConfig   `mapstructure:"models"`
}

// AppConfig 应用基本配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Env     string `mapstructure:"env"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	Charset         string        `mapstructure:"charset"`
	ParseTime       bool          `mapstructure:"parse_time"`
	Loc             string        `mapstructure:"loc"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	Password     string `mapstructure:"password"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// QueueConfig 队列配置
type QueueConfig struct {
	HighPriorityQueue   string        `mapstructure:"high_priority_queue"`
	MediumPriorityQueue string        `mapstructure:"medium_priority_queue"`
	LowPriorityQueue    string        `mapstructure:"low_priority_queue"`
	DelayedQueue        string        `mapstructure:"delayed_queue"`
	ProcessingQueue     string        `mapstructure:"processing_queue"`
	MaxQueueSize        int           `mapstructure:"max_queue_size"`
	TaskTimeout         time.Duration `mapstructure:"task_timeout"`
	MaxRetries          int           `mapstructure:"max_retries"`
	RetryDelay          time.Duration `mapstructure:"retry_delay"`
}

// WorkerConfig Worker 配置
type WorkerConfig struct {
	DefaultWorkers    int           `mapstructure:"default_workers"`
	MaxWorkers        int           `mapstructure:"max_workers"`
	WorkerTimeout     time.Duration `mapstructure:"worker_timeout"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level       string `mapstructure:"level"`
	Format      string `mapstructure:"format"`
	Output      string `mapstructure:"output"`
	FilePath    string `mapstructure:"file_path"`
	MaxSize     int    `mapstructure:"max_size"`
	MaxAge      int    `mapstructure:"max_age"`
	MaxBackups  int    `mapstructure:"max_backups"`
	Compress    bool   `mapstructure:"compress"`
}

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           string   `mapstructure:"max_age"`
}

// ModelsConfig 模型配置
type ModelsConfig struct {
	OpenAI OpenAIConfig `mapstructure:"openai"`
	Local  LocalConfig  `mapstructure:"local"`
}

// OpenAIConfig OpenAI 配置
type OpenAIConfig struct {
	BaseURL    string        `mapstructure:"base_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

// LocalConfig 本地模型配置
type LocalConfig struct {
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvPrefix("LLM_SCHEDULER")

	// 环境变量映射
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.username", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.database", "DB_NAME")
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetDSN 获取数据库连接字符串
func (db *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		db.Username,
		db.Password,
		db.Host,
		db.Port,
		db.Database,
		db.Charset,
		db.ParseTime,
		db.Loc,
	)
}

// GetRedisAddr 获取 Redis 地址
func (r *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

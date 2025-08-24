package queue

import (
	"context"
	"fmt"

	"llm-scheduler/config"

	"github.com/go-redis/redis/v8"
)

// InitRedis 初始化 Redis 连接
func InitRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	// 测试连接
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return rdb, nil
}

// HealthCheck Redis 健康检查
func HealthCheck(client *redis.Client) error {
	ctx := context.Background()
	return client.Ping(ctx).Err()
}

// GetRedisInfo 获取 Redis 信息
func GetRedisInfo(client *redis.Client) (map[string]string, error) {
	ctx := context.Background()
	info, err := client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	// 解析 info 字符串并返回关键信息
	result := map[string]string{
		"redis_version": "",
		"used_memory":   "",
		"connected_clients": "",
		"total_commands_processed": "",
	}

	// 这里可以解析 info 字符串，提取关键信息
	// 为了简化，直接返回原始信息
	result["raw_info"] = info

	return result, nil
}

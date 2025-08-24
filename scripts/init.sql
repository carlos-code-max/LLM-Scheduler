-- LLM Scheduler 数据库初始化脚本
-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS llm_scheduler CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE llm_scheduler;

-- 模型表
CREATE TABLE IF NOT EXISTS models (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE COMMENT '模型名称',
    type ENUM('openai', 'local', 'custom') NOT NULL COMMENT '模型类型',
    config JSON NOT NULL COMMENT '模型配置（API Key、参数等）',
    status ENUM('online', 'offline', 'maintenance') DEFAULT 'offline' COMMENT '模型状态',
    max_workers INT DEFAULT 1 COMMENT '最大并发 Worker 数量',
    current_workers INT DEFAULT 0 COMMENT '当前活跃 Worker 数量',
    total_requests BIGINT DEFAULT 0 COMMENT '总请求次数',
    success_requests BIGINT DEFAULT 0 COMMENT '成功请求次数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_type_status (type, status),
    INDEX idx_updated_at (updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模型配置表';

-- 任务表
CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    model_id BIGINT NOT NULL COMMENT '关联模型ID',
    type VARCHAR(50) NOT NULL COMMENT '任务类型',
    input TEXT NOT NULL COMMENT '输入内容',
    output TEXT COMMENT '输出内容（完成后填充）',
    status ENUM('pending', 'running', 'completed', 'failed', 'cancelled') DEFAULT 'pending' COMMENT '任务状态',
    priority TINYINT DEFAULT 1 COMMENT '优先级（1-低，2-中，3-高）',
    retry_count INT DEFAULT 0 COMMENT '已重试次数',
    max_retries INT DEFAULT 3 COMMENT '最大重试次数',
    error_message TEXT COMMENT '错误信息',
    started_at DATETIME COMMENT '开始执行时间',
    completed_at DATETIME COMMENT '完成时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
    INDEX idx_model_status (model_id, status),
    INDEX idx_status_priority (status, priority DESC),
    INDEX idx_created_at (created_at DESC),
    INDEX idx_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务表';

-- 任务日志表
CREATE TABLE IF NOT EXISTS task_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id BIGINT NOT NULL COMMENT '对应任务ID',
    level ENUM('info', 'warn', 'error', 'debug') DEFAULT 'info' COMMENT '日志级别',
    message TEXT NOT NULL COMMENT '日志内容',
    data JSON COMMENT '附加数据',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '记录时间',
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    INDEX idx_task_created (task_id, created_at DESC),
    INDEX idx_level_created (level, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务日志表';

-- 系统统计表
CREATE TABLE IF NOT EXISTS system_stats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    stat_date DATE NOT NULL COMMENT '统计日期',
    total_tasks INT DEFAULT 0 COMMENT '总任务数',
    completed_tasks INT DEFAULT 0 COMMENT '完成任务数',
    failed_tasks INT DEFAULT 0 COMMENT '失败任务数',
    avg_processing_time_ms INT DEFAULT 0 COMMENT '平均处理时间(毫秒)',
    queue_length INT DEFAULT 0 COMMENT '队列长度',
    active_models INT DEFAULT 0 COMMENT '活跃模型数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_stat_date (stat_date),
    INDEX idx_stat_date (stat_date DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统统计表';

-- 插入初始模型配置
INSERT INTO models (name, type, config, status, max_workers) VALUES 
(
    'gpt-3.5-turbo',
    'openai',
    JSON_OBJECT(
        'api_key', 'your-openai-api-key',
        'base_url', 'https://api.openai.com/v1',
        'model', 'gpt-3.5-turbo',
        'max_tokens', 4096,
        'temperature', 0.7
    ),
    'offline',
    3
),
(
    'gpt-4',
    'openai', 
    JSON_OBJECT(
        'api_key', 'your-openai-api-key',
        'base_url', 'https://api.openai.com/v1',
        'model', 'gpt-4',
        'max_tokens', 8192,
        'temperature', 0.7
    ),
    'offline',
    2
),
(
    'local-llama',
    'local',
    JSON_OBJECT(
        'model_path', '/models/llama-2-7b-chat',
        'host', 'localhost',
        'port', 8000,
        'max_tokens', 2048,
        'temperature', 0.7
    ),
    'offline',
    1
);

-- 创建示例任务（用于测试）
INSERT INTO tasks (model_id, type, input, priority) VALUES
(1, 'text-generation', '请写一个关于人工智能的简短介绍', 2),
(1, 'translation', 'Translate this text to Chinese: Hello, how are you?', 1),
(2, 'summarization', '请总结以下文本的主要内容：人工智能是计算机科学的一个分支，它试图理解智能的实质，并生产出一种新的能以人类智能相似方式做出反应的智能机器。', 3);

-- 创建视图：任务统计
CREATE VIEW v_task_stats AS
SELECT 
    DATE(created_at) as date,
    status,
    COUNT(*) as count,
    AVG(CASE 
        WHEN completed_at IS NOT NULL AND started_at IS NOT NULL 
        THEN TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000
        ELSE NULL 
    END) as avg_processing_time_ms
FROM tasks 
GROUP BY DATE(created_at), status;

-- 创建视图：模型状态统计
CREATE VIEW v_model_stats AS
SELECT 
    m.id,
    m.name,
    m.type,
    m.status,
    m.current_workers,
    m.max_workers,
    m.total_requests,
    m.success_requests,
    ROUND(
        CASE WHEN m.total_requests > 0 
        THEN (m.success_requests * 100.0 / m.total_requests) 
        ELSE 0 END, 2
    ) as success_rate,
    COUNT(t.id) as pending_tasks
FROM models m
LEFT JOIN tasks t ON m.id = t.model_id AND t.status = 'pending'
GROUP BY m.id, m.name, m.type, m.status, m.current_workers, m.max_workers, m.total_requests, m.success_requests;

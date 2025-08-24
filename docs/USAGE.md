# LLM Scheduler 使用指南

## 📖 目录

1. [快速开始](#快速开始)
2. [系统架构](#系统架构)
3. [功能介绍](#功能介绍)
4. [API 接口](#api-接口)
5. [配置说明](#配置说明)
6. [开发指南](#开发指南)
7. [常见问题](#常见问题)

## 🚀 快速开始

### 使用 Docker Compose (推荐)

1. **克隆项目**
```bash
git clone https://github.com/your-org/llm-scheduler.git
cd llm-scheduler
```

2. **启动所有服务**
```bash
# 使用启动脚本
chmod +x scripts/start.sh
./scripts/start.sh

# 或者直接使用 docker-compose
docker-compose up -d
```

3. **访问服务**
- Dashboard: http://localhost:3000
- API: http://localhost:8080
- 健康检查: http://localhost:8080/api/v1/system/health

### 开发环境

```bash
# 启动开发环境
chmod +x scripts/dev.sh
./scripts/dev.sh
```

## 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  React Frontend │────│   Go API Server │────│  MySQL Database │
│   (Dashboard)   │    │   (Gin/Fiber)   │    │   (任务/模型)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                       ┌─────────────────┐
                       │  Redis Queue    │
                       │  (任务调度)      │
                       └─────────────────┘
```

### 核心组件

- **API 层**: 提供 REST 接口，处理用户请求
- **任务调度层**: 基于 Redis 的优先级队列系统
- **Worker 管理**: 并发处理任务的工作进程
- **持久化层**: MySQL 存储任务数据和结果
- **前端界面**: React Dashboard 提供可视化管理

## 🎯 功能介绍

### 1. 任务管理

#### 支持的任务类型
- **text-generation**: 文本生成
- **translation**: 文本翻译
- **summarization**: 文本摘要
- **embedding**: 文本向量化
- **custom**: 自定义任务类型

#### 任务优先级
- **高优先级 (3)**: 紧急任务，优先处理
- **中优先级 (2)**: 普通任务，默认级别
- **低优先级 (1)**: 批量任务，资源空闲时处理

#### 任务状态流转
```
Pending → Running → Completed/Failed/Cancelled
           ↓
        (可重试)
```

### 2. 模型管理

#### 支持的模型类型
- **OpenAI**: GPT-3.5, GPT-4 等 OpenAI 模型
- **Local**: 本地部署的开源模型 (LLaMA, ChatGLM 等)
- **Custom**: 自定义模型接口

#### 模型配置示例

**OpenAI 模型配置**:
```json
{
  "api_key": "your-openai-api-key",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-3.5-turbo",
  "max_tokens": 4096,
  "temperature": 0.7
}
```

**本地模型配置**:
```json
{
  "host": "localhost",
  "port": 8000,
  "model_path": "/models/llama-2-7b-chat",
  "max_tokens": 2048,
  "temperature": 0.7
}
```

### 3. 队列调度

#### 调度策略
- 优先级调度: 高 → 中 → 低
- 同优先级内 FIFO (先进先出)
- 并发控制: 每模型可配置最大 Worker 数
- 反压机制: 队列过长时自动限流

#### 重试机制
- 失败任务自动重试
- 可配置最大重试次数
- 指数退避延迟

## 🔌 API 接口

### 任务相关接口

#### 创建任务
```http
POST /api/v1/tasks
Content-Type: application/json

{
  "model_id": 1,
  "type": "text-generation",
  "input": "写一个关于人工智能的简短介绍",
  "priority": 2
}
```

#### 获取任务列表
```http
GET /api/v1/tasks?page=1&page_size=20&status=pending
```

#### 获取任务详情
```http
GET /api/v1/tasks/{id}
```

#### 取消任务
```http
DELETE /api/v1/tasks/{id}
```

#### 重试任务
```http
POST /api/v1/tasks/{id}/retry
```

### 模型相关接口

#### 创建模型
```http
POST /api/v1/models
Content-Type: application/json

{
  "name": "gpt-3.5-turbo",
  "type": "openai",
  "config": {
    "api_key": "your-key",
    "model": "gpt-3.5-turbo"
  },
  "max_workers": 3
}
```

#### 获取模型列表
```http
GET /api/v1/models
```

#### 更新模型状态
```http
PUT /api/v1/models/{id}/status
Content-Type: application/json

{
  "status": "online"
}
```

### 统计接口

#### Dashboard 统计
```http
GET /api/v1/stats/dashboard
```

#### 按日期统计
```http
GET /api/v1/stats/tasks/date?days=7
```

## ⚙️ 配置说明

### 后端配置文件 (backend/config.yaml)

```yaml
app:
  name: "LLM Scheduler"
  version: "1.0.0"
  env: "development"

server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 3306
  username: "llm_user"
  password: "llm_password"
  database: "llm_scheduler"

redis:
  host: "localhost"
  port: 6379
  db: 0

queue:
  max_queue_size: 10000
  task_timeout: "300s"
  max_retries: 3
  retry_delay: "60s"

worker:
  default_workers: 5
  max_workers: 50
```

### 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|---------|
| `DB_HOST` | 数据库主机 | localhost |
| `DB_PORT` | 数据库端口 | 3306 |
| `DB_USER` | 数据库用户名 | llm_user |
| `DB_PASSWORD` | 数据库密码 | llm_password |
| `REDIS_HOST` | Redis 主机 | localhost |
| `REDIS_PORT` | Redis 端口 | 6379 |
| `REACT_APP_API_URL` | API 地址 | http://localhost:8080 |

## 🛠️ 开发指南

### 后端开发

1. **安装依赖**
```bash
cd backend
go mod tidy
```

2. **启动开发服务器**
```bash
go run main.go
```

3. **运行测试**
```bash
go test ./...
```

### 前端开发

1. **安装依赖**
```bash
cd frontend
npm install
```

2. **启动开发服务器**
```bash
npm start
```

3. **构建生产版本**
```bash
npm run build
```

### 数据库迁移

```bash
# 连接数据库
mysql -u root -p

# 导入初始化脚本
source scripts/init.sql
```

## ❓ 常见问题

### Q: 如何添加新的模型类型？

A: 需要在以下文件中添加支持：
1. `backend/models/model.go` - 添加模型类型枚举
2. `backend/worker/worker.go` - 实现对应的执行逻辑
3. `frontend/src/types/index.ts` - 更新前端类型定义

### Q: 任务一直处于 Pending 状态怎么办？

A: 检查以下几点：
1. 模型是否已上线 (status = online)
2. 模型是否有可用的 Worker 槽位
3. 队列是否正常工作
4. 查看后端日志是否有错误

### Q: 如何扩展任务类型？

A: 1. 在数据库中添加新的任务类型
2. 在 Worker 中实现对应的处理逻辑
3. 在前端添加创建界面

### Q: 如何监控系统性能？

A: 可以通过以下方式：
1. Dashboard 页面查看实时统计
2. 系统管理页面查看健康状态
3. 查看后端日志文件
4. 使用 `docker-compose logs -f` 查看容器日志

### Q: 生产环境部署注意事项？

A: 1. 修改数据库和 Redis 密码
2. 使用 HTTPS
3. 配置反向代理
4. 设置适当的资源限制
5. 配置日志轮转
6. 设置监控和告警

## 📞 支持

如有问题，请：
1. 查看 [GitHub Issues](https://github.com/your-org/llm-scheduler/issues)
2. 提交新的 Issue
3. 参与社区讨论

## 🤝 贡献

欢迎提交 Pull Request！请先阅读 [贡献指南](CONTRIBUTING.md)。

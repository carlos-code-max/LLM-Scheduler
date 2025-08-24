# LLM Scheduler - Open Source LLM Task Management Platform

## Project Overview

LLM Scheduler is an open-source large language model scheduling and task management tool designed for developers and enterprises. It provides unified multi-model task management, priority scheduling, and status tracking capabilities.

## Key Features

- **Multi-Model Support** - Support for OpenAI API, local LLaMA, and various other large language models
- **Intelligent Scheduling** - Priority-based task queue scheduling with rate limiting and backpressure mechanisms
- **Visual Management** - Built-in Dashboard with real-time task and model status display
- **Plugin Extensibility** - Support for custom scheduling strategies and task types
- **Ready to Deploy** - One-click Docker deployment with complete deployment solutions

## Technical Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  React Frontend │────│   Go API Server │────│  MySQL Database │
│   (Dashboard)   │    │   (Gin/Fiber)   │    │ (Tasks/Models)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                       ┌─────────────────┐
                       │  Redis Queue    │
                       │ (Task Scheduling│
                       └─────────────────┘
```

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone git@github.com:carlos-code-max/LLM-Scheduler.git
cd LLM-Scheduler

# Start all services
docker-compose up -d

# Access Dashboard
open http://localhost:3000
```

### Manual Deployment

1. **Start Backend Service**
```bash
cd backend
go mod tidy
go run main.go
```

2. **Start Frontend**
```bash
cd frontend
npm install
npm start
```

3. **Configure Database**
```bash
# Import database schema
mysql -u root -p < scripts/init.sql
```

## API Documentation

### Task Management

- `POST /api/tasks` - Submit new task
- `GET /api/tasks` - Get task list
- `GET /api/tasks/:id` - Get task details
- `PUT /api/tasks/:id/priority` - Adjust task priority

### Model Management

- `GET /api/models` - Get model list
- `POST /api/models` - Register new model
- `PUT /api/models/:id` - Update model configuration

## 📋 Task Types

- **text-generation** - Text generation tasks
- **embedding** - Text vectorization
- **translation** - Text translation
- **summarization** - Text summarization
- **custom** - Custom task types

## Configuration

Main configuration file: `backend/config.yaml`

```yaml
server:
  port: 8080
  
database:
  host: localhost
  port: 3306
  username: root
  password: password
  database: llm_scheduler

redis:
  host: localhost
  port: 6379
  db: 0
```

## Contributing

Issues and Pull Requests are welcome!

1. Fork this repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

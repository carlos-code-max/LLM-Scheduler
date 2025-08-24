#!/bin/bash

# LLM Scheduler 启动脚本

echo "🚀 Starting LLM Scheduler..."

# 检查 Docker 和 Docker Compose 是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# 创建必要的目录
mkdir -p logs
mkdir -p data/mysql
mkdir -p data/redis

echo "📁 Created necessary directories"

# 停止现有容器（如果有）
echo "🛑 Stopping existing containers..."
docker-compose down

# 清理孤儿容器
docker-compose down --remove-orphans

# 构建并启动所有服务
echo "🏗️ Building and starting services..."
docker-compose up -d --build

# 等待服务启动
echo "⏳ Waiting for services to start..."
sleep 30

# 检查服务状态
echo "🔍 Checking service status..."
docker-compose ps

# 检查健康状态
echo "🏥 Checking health status..."
until curl -f http://localhost:8080/api/v1/system/health >/dev/null 2>&1; do
    echo "Waiting for backend to be healthy..."
    sleep 5
done

echo "✅ Backend is healthy!"

until curl -f http://localhost:3000/health >/dev/null 2>&1; do
    echo "Waiting for frontend to be healthy..."
    sleep 5
done

echo "✅ Frontend is healthy!"

echo ""
echo "🎉 LLM Scheduler is now running!"
echo ""
echo "📊 Dashboard: http://localhost:3000"
echo "🔧 API: http://localhost:8080"
echo "📋 API Documentation: http://localhost:8080/api/v1"
echo ""
echo "🐳 Docker containers:"
docker-compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
echo ""
echo "📝 To view logs: docker-compose logs -f"
echo "🛑 To stop: docker-compose down"
echo ""

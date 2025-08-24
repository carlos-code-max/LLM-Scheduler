#!/bin/bash

# LLM Scheduler 开发环境启动脚本

echo "🛠️ Starting LLM Scheduler in Development Mode..."

# 检查必要的工具
check_command() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 is not installed. Please install $1 first."
        exit 1
    fi
}

check_command go
check_command node
check_command npm
check_command docker
check_command docker-compose

echo "✅ All required tools are available"

# 启动基础服务（MySQL + Redis）
echo "🐳 Starting database and cache services..."
docker-compose up -d mysql redis

echo "⏳ Waiting for database and Redis to be ready..."
sleep 15

# 检查数据库连接
until docker-compose exec mysql mysqladmin ping -h"localhost" --silent; do
    echo "Waiting for MySQL to be ready..."
    sleep 2
done
echo "✅ MySQL is ready!"

# 检查 Redis 连接  
until docker-compose exec redis redis-cli ping | grep -q PONG; do
    echo "Waiting for Redis to be ready..."
    sleep 2
done
echo "✅ Redis is ready!"

# 在后台启动后端开发服务器
echo "🏃 Starting backend in development mode..."
cd backend
go mod tidy
go run main.go &
BACKEND_PID=$!
echo "Backend PID: $BACKEND_PID"
cd ..

# 等待后端启动
echo "⏳ Waiting for backend to start..."
sleep 10

# 在后台启动前端开发服务器
echo "🎨 Starting frontend in development mode..."
cd frontend
npm install
npm start &
FRONTEND_PID=$!
echo "Frontend PID: $FRONTEND_PID"
cd ..

echo ""
echo "🎉 LLM Scheduler development environment is starting!"
echo ""
echo "📊 Frontend (React): http://localhost:3000"
echo "🔧 Backend (Go): http://localhost:8080"
echo "🗄️ MySQL: localhost:3306"
echo "💾 Redis: localhost:6379"
echo ""
echo "📝 Backend PID: $BACKEND_PID"
echo "📝 Frontend PID: $FRONTEND_PID"
echo ""

# 创建停止脚本
cat > scripts/stop-dev.sh << EOF
#!/bin/bash
echo "🛑 Stopping development services..."
kill $BACKEND_PID 2>/dev/null && echo "✅ Backend stopped" || echo "⚠️ Backend was not running"
kill $FRONTEND_PID 2>/dev/null && echo "✅ Frontend stopped" || echo "⚠️ Frontend was not running"
docker-compose stop mysql redis
echo "✅ Database services stopped"
rm -f scripts/stop-dev.sh
echo "🏁 Development environment stopped!"
EOF
chmod +x scripts/stop-dev.sh

echo "🛑 To stop development environment: ./scripts/stop-dev.sh"
echo "📋 To view backend logs: tail -f backend/logs/app.log (when available)"
echo "📋 To view database logs: docker-compose logs -f mysql"
echo "📋 To view Redis logs: docker-compose logs -f redis"

# 等待用户中断
trap 'echo ""; echo "🛑 Shutting down..."; ./scripts/stop-dev.sh; exit 0' INT

echo ""
echo "Press Ctrl+C to stop all services..."
wait

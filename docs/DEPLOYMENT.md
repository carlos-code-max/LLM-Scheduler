# LLM Scheduler 部署指南

## 📋 目录

1. [系统要求](#系统要求)
2. [Docker Compose 部署 (推荐)](#docker-compose-部署-推荐)
3. [手动部署](#手动部署)
4. [生产环境配置](#生产环境配置)
5. [性能优化](#性能优化)
6. [监控和日志](#监控和日志)
7. [故障排除](#故障排除)

## 🖥️ 系统要求

### 最低配置
- CPU: 2 核
- 内存: 4GB RAM
- 存储: 20GB 可用空间
- 操作系统: Ubuntu 20.04+ / CentOS 7+ / Windows 10+

### 推荐配置
- CPU: 4 核
- 内存: 8GB RAM
- 存储: 50GB 可用空间 (SSD)
- 操作系统: Ubuntu 22.04 LTS

### 软件依赖
- Docker 20.10+
- Docker Compose 2.0+
- Git 2.0+

## 🐳 Docker Compose 部署 (推荐)

### 1. 下载项目

```bash
git clone https://github.com/your-org/llm-scheduler.git
cd llm-scheduler
```

### 2. 配置环境变量

创建 `.env` 文件：

```bash
# 数据库配置
MYSQL_ROOT_PASSWORD=your_secure_password_here
MYSQL_DATABASE=llm_scheduler
MYSQL_USER=llm_user
MYSQL_PASSWORD=your_mysql_password_here

# Redis 配置
REDIS_PASSWORD=your_redis_password_here

# 应用配置
APP_ENV=production
API_BASE_URL=https://your-domain.com

# JWT 密钥（如果启用认证）
JWT_SECRET=your_jwt_secret_here
```

### 3. 生产环境配置

修改 `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: llm-scheduler-mysql
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: ${MYSQL_DATABASE}
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    ports:
      - "127.0.0.1:3306:3306"  # 只绑定本地接口
    volumes:
      - mysql_data:/var/lib/mysql
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
      - ./config/mysql.cnf:/etc/mysql/conf.d/custom.cnf
    networks:
      - llm-scheduler-internal
    restart: unless-stopped
    command: --default-authentication-plugin=mysql_native_password
    
  redis:
    image: redis:7-alpine
    container_name: llm-scheduler-redis
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    ports:
      - "127.0.0.1:6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - llm-scheduler-internal
    restart: unless-stopped

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.prod
    container_name: llm-scheduler-backend
    environment:
      - GIN_MODE=release
      - APP_ENV=production
      - DB_HOST=mysql
      - DB_PASSWORD=${MYSQL_PASSWORD}
      - REDIS_HOST=redis
      - REDIS_PASSWORD=${REDIS_PASSWORD}
    depends_on:
      - mysql
      - redis
    volumes:
      - ./logs:/app/logs
      - ./config/config.prod.yaml:/app/config.yaml
    networks:
      - llm-scheduler-internal
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/v1/system/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
      args:
        - REACT_APP_API_URL=${API_BASE_URL}
    container_name: llm-scheduler-frontend
    depends_on:
      - backend
    networks:
      - llm-scheduler-internal
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    container_name: llm-scheduler-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.conf:/etc/nginx/nginx.conf
      - ./config/ssl:/etc/nginx/ssl
      - ./logs/nginx:/var/log/nginx
    depends_on:
      - frontend
      - backend
    networks:
      - llm-scheduler-internal
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:

networks:
  llm-scheduler-internal:
    driver: bridge
```

### 4. Nginx 配置

创建 `config/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                   '$status $body_bytes_sent "$http_referer" '
                   '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log warn;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;

    # Gzip 压缩
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # 限制请求大小
    client_max_body_size 10M;

    # 后端代理
    upstream backend {
        server backend:8080;
    }

    # 前端代理
    upstream frontend {
        server frontend:80;
    }

    # HTTPS 重定向
    server {
        listen 80;
        server_name your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    # HTTPS 服务器
    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        # SSL 配置
        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384;
        ssl_prefer_server_ciphers off;

        # 安全头
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header Referrer-Policy "no-referrer-when-downgrade" always;
        add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

        # API 代理
        location /api/ {
            proxy_pass http://backend/api/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }

        # 前端
        location / {
            proxy_pass http://frontend/;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_cache_bypass $http_upgrade;
        }
    }
}
```

### 5. 启动生产环境

```bash
# 启动服务
docker-compose -f docker-compose.prod.yml up -d

# 查看状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f
```

## 🔧 手动部署

### 1. 准备环境

```bash
# 安装 Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 安装 Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# 安装 MySQL
sudo apt update
sudo apt install mysql-server

# 安装 Redis
sudo apt install redis-server
```

### 2. 数据库设置

```bash
# 配置 MySQL
sudo mysql_secure_installation
mysql -u root -p < scripts/init.sql

# 启动 Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

### 3. 构建后端

```bash
cd backend
go mod tidy
go build -o llm-scheduler main.go

# 创建配置文件
cp config.yaml config.prod.yaml
# 编辑配置文件...

# 创建 systemd 服务
sudo tee /etc/systemd/system/llm-scheduler.service > /dev/null <<EOF
[Unit]
Description=LLM Scheduler Backend
After=network.target mysql.service redis.service

[Service]
Type=simple
User=llm-scheduler
WorkingDirectory=/opt/llm-scheduler
ExecStart=/opt/llm-scheduler/llm-scheduler
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable llm-scheduler
sudo systemctl start llm-scheduler
```

### 4. 构建前端

```bash
cd frontend
npm install
REACT_APP_API_URL=https://your-domain.com npm run build

# 配置 Nginx
sudo cp build/* /var/www/html/
```

## ⚡ 生产环境配置

### 1. 安全配置

```yaml
# backend/config.prod.yaml
app:
  env: "production"

database:
  # 使用强密码
  password: "your_very_secure_password"
  # 启用 SSL
  tls: true

logging:
  level: "warn"  # 减少日志输出
  
cors:
  allow_origins: ["https://your-domain.com"]
```

### 2. 性能优化

```yaml
# 数据库连接池
database:
  max_idle_conns: 25
  max_open_conns: 100
  conn_max_lifetime: "1h"

# Redis 连接池
redis:
  pool_size: 20
  min_idle_conns: 10

# Worker 配置
worker:
  default_workers: 10
  max_workers: 100
```

### 3. 备份策略

```bash
#!/bin/bash
# backup.sh

# 数据库备份
mysqldump -u root -p llm_scheduler > /backup/llm_scheduler_$(date +%Y%m%d_%H%M%S).sql

# Redis 备份
cp /var/lib/redis/dump.rdb /backup/redis_$(date +%Y%m%d_%H%M%S).rdb

# 日志归档
tar -czf /backup/logs_$(date +%Y%m%d).tar.gz logs/

# 清理旧备份 (保留7天)
find /backup -name "*.sql" -mtime +7 -delete
find /backup -name "*.rdb" -mtime +7 -delete
find /backup -name "*.tar.gz" -mtime +7 -delete
```

## 📊 监控和日志

### 1. 健康检查

```bash
# API 健康检查
curl -f https://your-domain.com/api/v1/system/health

# 数据库连接检查
mysql -u llm_user -p -e "SELECT 1"

# Redis 连接检查
redis-cli ping
```

### 2. 日志配置

```yaml
# backend/config.prod.yaml
logging:
  level: "info"
  format: "json"
  output: "file"
  file_path: "/var/log/llm-scheduler/app.log"
  max_size: 100  # MB
  max_age: 30    # days
  max_backups: 10
  compress: true
```

### 3. Prometheus 监控 (可选)

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      
  grafana:
    image: grafana/grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana

volumes:
  grafana_data:
```

## 🔍 故障排除

### 常见问题

1. **数据库连接失败**
```bash
# 检查数据库状态
sudo systemctl status mysql
mysql -u llm_user -p -e "SELECT 1"
```

2. **Redis 连接失败**
```bash
# 检查 Redis 状态
sudo systemctl status redis
redis-cli ping
```

3. **任务不执行**
```bash
# 检查 Worker 状态
curl https://your-domain.com/api/v1/stats/dashboard
# 查看后端日志
tail -f /var/log/llm-scheduler/app.log
```

4. **前端无法访问**
```bash
# 检查 Nginx 状态
sudo systemctl status nginx
# 检查配置
sudo nginx -t
```

### 性能调优

1. **数据库优化**
```sql
-- 创建索引
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
CREATE INDEX idx_tasks_status ON tasks(status);

-- 查看慢查询
SHOW VARIABLES LIKE 'slow_query_log';
```

2. **Redis 优化**
```bash
# Redis 配置优化
echo 'maxmemory 2gb' >> /etc/redis/redis.conf
echo 'maxmemory-policy allkeys-lru' >> /etc/redis/redis.conf
```

3. **系统资源监控**
```bash
# CPU 和内存使用情况
htop

# 磁盘使用情况
df -h

# 网络连接
netstat -tulpn
```

## 📞 获取帮助

如果遇到部署问题，请：

1. 查看 [故障排除文档](TROUBLESHOOTING.md)
2. 检查 [GitHub Issues](https://github.com/your-org/llm-scheduler/issues)
3. 提交详细的错误报告

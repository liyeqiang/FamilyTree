#!/bin/bash

# Linux Redis 启动脚本

echo "🐧 启动 Redis 服务 (Linux)"
echo "========================="

# 检查是否以 root 权限运行
if [[ $EUID -eq 0 ]]; then
    SUDO=""
else
    SUDO="sudo"
fi

# 检查 Redis 是否已安装
if ! command -v redis-server &> /dev/null; then
    echo "❌ Redis 未安装，请先运行安装脚本: ./install_redis_linux.sh"
    exit 1
fi

# 检测服务名称（不同发行版可能不同）
SERVICE_NAME=""
if systemctl list-unit-files | grep -q "redis-server.service"; then
    SERVICE_NAME="redis-server"
elif systemctl list-unit-files | grep -q "redis.service"; then
    SERVICE_NAME="redis"
else
    echo "⚠️  未检测到 Redis 系统服务，尝试手动启动..."
fi

# 检查 Redis 服务状态
if [ -n "$SERVICE_NAME" ]; then
    if systemctl is-active --quiet $SERVICE_NAME; then
        echo "✅ Redis 服务已经在运行"
        echo "状态: $(systemctl is-active $SERVICE_NAME)"
        
        # 显示服务信息
        echo ""
        echo "📊 服务信息:"
        $SUDO systemctl status $SERVICE_NAME --no-pager -l
        
        exit 0
    fi
    
    # 使用 systemctl 启动服务
    echo "🚀 使用 systemctl 启动 Redis 服务..."
    $SUDO systemctl start $SERVICE_NAME
    
    # 等待服务启动
    sleep 2
    
    # 检查服务状态
    if systemctl is-active --quiet $SERVICE_NAME; then
        echo "✅ Redis 服务启动成功!"
        
        # 显示服务信息
        echo ""
        echo "📊 服务信息:"
        $SUDO systemctl status $SERVICE_NAME --no-pager -l
        
    else
        echo "❌ Redis 服务启动失败"
        echo "错误日志:"
        $SUDO journalctl -u $SERVICE_NAME --no-pager -l -n 10
        exit 1
    fi
    
else
    # 手动启动 Redis
    echo "🚀 手动启动 Redis 服务..."
    
    # 检查是否已经在运行
    if pgrep -x "redis-server" > /dev/null; then
        echo "⚠️  Redis 进程已经在运行"
        echo "进程ID: $(pgrep -x redis-server)"
        exit 0
    fi
    
    # 配置文件路径
    CONFIG_FILE=""
    if [ -f "/etc/redis/redis.conf" ]; then
        CONFIG_FILE="/etc/redis/redis.conf"
        echo "📄 使用配置文件: $CONFIG_FILE"
    elif [ -f "../config/redis.conf" ]; then
        CONFIG_FILE="../config/redis.conf"
        echo "📄 使用配置文件: $CONFIG_FILE"
    else
        echo "📄 使用默认配置"
    fi
    
    # 创建日志目录
    $SUDO mkdir -p /var/log/redis
    $SUDO chown redis:redis /var/log/redis 2>/dev/null || true
    
    # 启动 Redis 服务
    if [ -n "$CONFIG_FILE" ]; then
        # 使用配置文件启动
        $SUDO redis-server "$CONFIG_FILE" --daemonize yes
    else
        # 使用默认配置启动
        $SUDO redis-server --daemonize yes --logfile /var/log/redis/redis.log --dir /var/lib/redis/
    fi
    
    # 等待服务启动
    sleep 2
    
    # 检查服务是否启动成功
    if pgrep -x "redis-server" > /dev/null; then
        echo "✅ Redis 服务启动成功!"
        echo "进程ID: $(pgrep -x redis-server)"
    else
        echo "❌ Redis 服务启动失败"
        echo "请检查日志文件: /var/log/redis/redis.log"
        exit 1
    fi
fi

# 测试连接
echo ""
echo "🔗 测试 Redis 连接..."
if redis-cli ping > /dev/null 2>&1; then
    echo "✅ Redis 连接测试成功"
    echo "服务地址: localhost:6379"
    
    # 显示 Redis 信息
    echo ""
    echo "📊 Redis 信息:"
    redis-cli info server | grep -E "redis_version|process_id|tcp_port|uptime_in_seconds"
    
else
    echo "⚠️  Redis 连接测试失败，请检查配置"
fi

echo ""
echo "🛠️  常用命令:"
if [ -n "$SERVICE_NAME" ]; then
    echo "  停止服务: sudo systemctl stop $SERVICE_NAME"
    echo "  重启服务: sudo systemctl restart $SERVICE_NAME"
    echo "  查看状态: sudo systemctl status $SERVICE_NAME"
    echo "  查看日志: sudo journalctl -u $SERVICE_NAME -f"
else
    echo "  停止服务: ./stop_redis_linux.sh"
    echo "  查看日志: tail -f /var/log/redis/redis.log"
fi
echo "  连接客户端: redis-cli" 
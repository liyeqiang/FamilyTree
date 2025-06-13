#!/bin/bash

# macOS Redis 启动脚本

echo "🍎 启动 Redis 服务 (macOS)"
echo "========================="

# 检查 Redis 是否已安装
if ! command -v redis-server &> /dev/null; then
    echo "❌ Redis 未安装，请先运行安装脚本: ./install_redis_macos.sh"
    exit 1
fi

# 检查 Redis 是否已经在运行
if pgrep -x "redis-server" > /dev/null; then
    echo "⚠️  Redis 服务已经在运行"
    echo "进程ID: $(pgrep -x redis-server)"
    echo "如需重启，请先运行停止脚本: ./stop_redis_macos.sh"
    exit 0
fi

# 配置文件路径
CONFIG_FILE=""
if [ -f "~/redis/config/redis.conf" ]; then
    CONFIG_FILE="~/redis/config/redis.conf"
    echo "📄 使用配置文件: $CONFIG_FILE"
elif [ -f "../config/redis.conf" ]; then
    CONFIG_FILE="../config/redis.conf"
    echo "📄 使用配置文件: $CONFIG_FILE"
else
    echo "📄 使用默认配置"
fi

# 创建日志目录
mkdir -p ~/redis/logs

# 启动 Redis 服务
echo "🚀 启动 Redis 服务..."

if [ -n "$CONFIG_FILE" ]; then
    # 使用配置文件启动
    nohup redis-server "$CONFIG_FILE" > ~/redis/logs/redis.log 2>&1 &
else
    # 使用默认配置启动
    nohup redis-server --daemonize yes --logfile ~/redis/logs/redis.log --dir ~/redis/ > /dev/null 2>&1 &
fi

# 等待服务启动
sleep 2

# 检查服务是否启动成功
if pgrep -x "redis-server" > /dev/null; then
    echo "✅ Redis 服务启动成功!"
    echo "进程ID: $(pgrep -x redis-server)"
    echo "日志文件: ~/redis/logs/redis.log"
    
    # 测试连接
    if redis-cli ping > /dev/null 2>&1; then
        echo "✅ Redis 连接测试成功"
        echo "服务地址: localhost:6379"
    else
        echo "⚠️  Redis 连接测试失败，请检查配置"
    fi
    
    echo ""
    echo "📊 Redis 信息:"
    redis-cli info server | grep -E "redis_version|process_id|tcp_port"
    
else
    echo "❌ Redis 服务启动失败"
    echo "请检查日志文件: ~/redis/logs/redis.log"
    exit 1
fi

echo ""
echo "🛠️  常用命令:"
echo "  停止服务: ./stop_redis_macos.sh"
echo "  连接客户端: redis-cli"
echo "  查看日志: tail -f ~/redis/logs/redis.log" 
#!/bin/bash

# macOS Redis 停止脚本

echo "🍎 停止 Redis 服务 (macOS)"
echo "========================="

# 检查 Redis 是否在运行
if ! pgrep -x "redis-server" > /dev/null; then
    echo "⚠️  Redis 服务未运行"
    exit 0
fi

echo "🛑 正在停止 Redis 服务..."

# 获取 Redis 进程ID
REDIS_PID=$(pgrep -x "redis-server")
echo "Redis 进程ID: $REDIS_PID"

# 方法1: 使用 redis-cli 优雅关闭
if command -v redis-cli &> /dev/null; then
    echo "📤 尝试使用 redis-cli 优雅关闭..."
    if redis-cli shutdown > /dev/null 2>&1; then
        echo "✅ Redis 服务已优雅关闭"
    else
        echo "⚠️  redis-cli 关闭失败，尝试其他方法..."
        
        # 方法2: 发送 SIGTERM 信号
        echo "📤 发送 SIGTERM 信号..."
        kill -TERM $REDIS_PID
        
        # 等待进程结束
        sleep 3
        
        # 检查进程是否还在运行
        if pgrep -x "redis-server" > /dev/null; then
            echo "⚠️  进程仍在运行，强制终止..."
            # 方法3: 强制终止
            kill -KILL $REDIS_PID
            sleep 1
        fi
    fi
else
    # 直接使用 kill 命令
    echo "📤 发送 SIGTERM 信号..."
    kill -TERM $REDIS_PID
    
    # 等待进程结束
    sleep 3
    
    # 检查进程是否还在运行
    if pgrep -x "redis-server" > /dev/null; then
        echo "⚠️  进程仍在运行，强制终止..."
        kill -KILL $REDIS_PID
        sleep 1
    fi
fi

# 最终检查
if pgrep -x "redis-server" > /dev/null; then
    echo "❌ Redis 服务停止失败"
    echo "请手动终止进程: kill -9 $(pgrep -x redis-server)"
    exit 1
else
    echo "✅ Redis 服务已成功停止"
fi

# 清理临时文件（可选）
read -p "是否清理 Redis 临时文件? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理临时文件..."
    
    # 清理可能的临时文件
    [ -f ~/redis/dump.rdb ] && rm ~/redis/dump.rdb && echo "  ✅ 删除 dump.rdb"
    [ -f ~/redis/redis.pid ] && rm ~/redis/redis.pid && echo "  ✅ 删除 redis.pid"
    [ -f /tmp/redis.sock ] && rm /tmp/redis.sock && echo "  ✅ 删除 redis.sock"
    
    echo "✅ 清理完成"
fi

echo ""
echo "🛠️  相关命令:"
echo "  启动服务: ./start_redis_macos.sh"
echo "  重新安装: ./install_redis_macos.sh"
echo "  查看日志: tail ~/redis/logs/redis.log" 
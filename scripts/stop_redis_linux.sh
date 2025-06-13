#!/bin/bash

# Linux Redis 停止脚本

echo "🐧 停止 Redis 服务 (Linux)"
echo "========================="

# 检查是否以 root 权限运行
if [[ $EUID -eq 0 ]]; then
    SUDO=""
else
    SUDO="sudo"
fi

# 检测服务名称（不同发行版可能不同）
SERVICE_NAME=""
if systemctl list-unit-files | grep -q "redis-server.service"; then
    SERVICE_NAME="redis-server"
elif systemctl list-unit-files | grep -q "redis.service"; then
    SERVICE_NAME="redis"
fi

# 使用 systemctl 停止服务
if [ -n "$SERVICE_NAME" ]; then
    # 检查服务状态
    if ! systemctl is-active --quiet $SERVICE_NAME; then
        echo "⚠️  Redis 服务未运行"
        exit 0
    fi
    
    echo "🛑 使用 systemctl 停止 Redis 服务..."
    $SUDO systemctl stop $SERVICE_NAME
    
    # 等待服务停止
    sleep 2
    
    # 检查服务状态
    if ! systemctl is-active --quiet $SERVICE_NAME; then
        echo "✅ Redis 服务已成功停止"
        
        # 显示服务状态
        echo ""
        echo "📊 服务状态:"
        $SUDO systemctl status $SERVICE_NAME --no-pager -l
        
    else
        echo "❌ Redis 服务停止失败"
        echo "尝试强制停止..."
        
        # 强制停止服务
        $SUDO systemctl kill $SERVICE_NAME
        sleep 2
        
        if ! systemctl is-active --quiet $SERVICE_NAME; then
            echo "✅ Redis 服务已强制停止"
        else
            echo "❌ 无法停止 Redis 服务"
            exit 1
        fi
    fi
    
else
    # 手动停止 Redis
    echo "🛑 手动停止 Redis 服务..."
    
    # 检查 Redis 是否在运行
    if ! pgrep -x "redis-server" > /dev/null; then
        echo "⚠️  Redis 服务未运行"
        exit 0
    fi
    
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
            $SUDO kill -TERM $REDIS_PID
            
            # 等待进程结束
            sleep 3
            
            # 检查进程是否还在运行
            if pgrep -x "redis-server" > /dev/null; then
                echo "⚠️  进程仍在运行，强制终止..."
                # 方法3: 强制终止
                $SUDO kill -KILL $REDIS_PID
                sleep 1
            fi
        fi
    else
        # 直接使用 kill 命令
        echo "📤 发送 SIGTERM 信号..."
        $SUDO kill -TERM $REDIS_PID
        
        # 等待进程结束
        sleep 3
        
        # 检查进程是否还在运行
        if pgrep -x "redis-server" > /dev/null; then
            echo "⚠️  进程仍在运行，强制终止..."
            $SUDO kill -KILL $REDIS_PID
            sleep 1
        fi
    fi
    
    # 最终检查
    if pgrep -x "redis-server" > /dev/null; then
        echo "❌ Redis 服务停止失败"
        echo "请手动终止进程: sudo kill -9 $(pgrep -x redis-server)"
        exit 1
    else
        echo "✅ Redis 服务已成功停止"
    fi
fi

# 清理临时文件（可选）
read -p "是否清理 Redis 临时文件? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🧹 清理临时文件..."
    
    # 清理可能的临时文件
    [ -f /var/lib/redis/dump.rdb ] && $SUDO rm /var/lib/redis/dump.rdb && echo "  ✅ 删除 dump.rdb"
    [ -f /var/run/redis/redis-server.pid ] && $SUDO rm /var/run/redis/redis-server.pid && echo "  ✅ 删除 redis-server.pid"
    [ -f /tmp/redis.sock ] && $SUDO rm /tmp/redis.sock && echo "  ✅ 删除 redis.sock"
    
    echo "✅ 清理完成"
fi

echo ""
echo "🛠️  相关命令:"
if [ -n "$SERVICE_NAME" ]; then
    echo "  启动服务: sudo systemctl start $SERVICE_NAME"
    echo "  重启服务: sudo systemctl restart $SERVICE_NAME"
    echo "  查看状态: sudo systemctl status $SERVICE_NAME"
    echo "  禁用开机启动: sudo systemctl disable $SERVICE_NAME"
else
    echo "  启动服务: ./start_redis_linux.sh"
    echo "  重新安装: ./install_redis_linux.sh"
fi
echo "  查看日志: tail /var/log/redis/redis.log" 
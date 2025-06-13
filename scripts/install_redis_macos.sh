#!/bin/bash

# macOS Redis 安装脚本
# 使用 Homebrew 安装 Redis

echo "🍎 macOS Redis 安装脚本"
echo "========================"

# 检查是否安装了 Homebrew
if ! command -v brew &> /dev/null; then
    echo "❌ 未检测到 Homebrew，正在安装..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # 添加 Homebrew 到 PATH
    echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
    eval "$(/opt/homebrew/bin/brew shellenv)"
else
    echo "✅ 检测到 Homebrew"
fi

# 更新 Homebrew
echo "🔄 更新 Homebrew..."
brew update

# 安装 Redis
echo "📦 安装 Redis..."
if brew list redis &> /dev/null; then
    echo "✅ Redis 已经安装"
    echo "当前版本: $(redis-server --version)"
    
    # 询问是否升级
    read -p "是否要升级到最新版本? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        brew upgrade redis
    fi
else
    brew install redis
fi

# 检查安装是否成功
if command -v redis-server &> /dev/null; then
    echo "✅ Redis 安装成功!"
    echo "版本: $(redis-server --version)"
    
    # 创建配置文件目录
    mkdir -p ~/redis/config
    
    # 复制配置文件
    if [ -f "../config/redis.conf" ]; then
        cp ../config/redis.conf ~/redis/config/
        echo "✅ Redis 配置文件已复制到 ~/redis/config/"
    else
        echo "⚠️  未找到 Redis 配置文件，将使用默认配置"
    fi
    
    echo ""
    echo "🚀 安装完成！"
    echo "启动 Redis: ./start_redis_macos.sh"
    echo "停止 Redis: ./stop_redis_macos.sh"
    echo "或者使用 Homebrew 服务:"
    echo "  启动: brew services start redis"
    echo "  停止: brew services stop redis"
    echo "  重启: brew services restart redis"
    
else
    echo "❌ Redis 安装失败"
    exit 1
fi 
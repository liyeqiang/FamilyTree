#!/bin/bash

# Linux Redis 安装脚本
# 支持 Ubuntu/Debian 和 CentOS/RHEL/Fedora

echo "🐧 Linux Redis 安装脚本"
echo "======================="

# 检测 Linux 发行版
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
    VER=$VERSION_ID
elif type lsb_release >/dev/null 2>&1; then
    OS=$(lsb_release -si)
    VER=$(lsb_release -sr)
else
    OS=$(uname -s)
    VER=$(uname -r)
fi

echo "检测到系统: $OS $VER"

# 检查是否以 root 权限运行
if [[ $EUID -eq 0 ]]; then
    SUDO=""
else
    SUDO="sudo"
    echo "⚠️  需要管理员权限来安装软件包"
fi

# 检查 Redis 是否已安装
if command -v redis-server &> /dev/null; then
    echo "✅ Redis 已经安装"
    echo "当前版本: $(redis-server --version)"
    
    read -p "是否要重新安装? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
fi

# 根据发行版安装 Redis
case $OS in
    *"Ubuntu"*|*"Debian"*)
        echo "📦 使用 APT 安装 Redis..."
        
        # 更新包列表
        echo "🔄 更新软件包列表..."
        $SUDO apt update
        
        # 安装 Redis
        echo "📥 安装 Redis..."
        $SUDO apt install -y redis-server redis-tools
        
        # 启用服务
        $SUDO systemctl enable redis-server
        ;;
        
    *"CentOS"*|*"Red Hat"*|*"Fedora"*|*"Rocky"*|*"AlmaLinux"*)
        echo "📦 使用 YUM/DNF 安装 Redis..."
        
        # 检查包管理器
        if command -v dnf &> /dev/null; then
            PKG_MGR="dnf"
        else
            PKG_MGR="yum"
        fi
        
        # 安装 EPEL 仓库 (CentOS/RHEL)
        if [[ $OS == *"CentOS"* ]] || [[ $OS == *"Red Hat"* ]]; then
            echo "📥 安装 EPEL 仓库..."
            $SUDO $PKG_MGR install -y epel-release
        fi
        
        # 安装 Redis
        echo "📥 安装 Redis..."
        $SUDO $PKG_MGR install -y redis
        
        # 启用服务
        $SUDO systemctl enable redis
        ;;
        
    *"Arch"*)
        echo "📦 使用 Pacman 安装 Redis..."
        
        # 更新包数据库
        $SUDO pacman -Sy
        
        # 安装 Redis
        $SUDO pacman -S --noconfirm redis
        
        # 启用服务
        $SUDO systemctl enable redis
        ;;
        
    *)
        echo "❌ 不支持的 Linux 发行版: $OS"
        echo "请手动安装 Redis 或使用源码编译安装"
        exit 1
        ;;
esac

# 检查安装是否成功
if command -v redis-server &> /dev/null; then
    echo "✅ Redis 安装成功!"
    echo "版本: $(redis-server --version)"
    
    # 创建配置文件目录
    $SUDO mkdir -p /etc/redis
    $SUDO mkdir -p /var/log/redis
    $SUDO mkdir -p /var/lib/redis
    
    # 设置权限
    $SUDO chown redis:redis /var/log/redis /var/lib/redis 2>/dev/null || true
    
    # 复制配置文件
    if [ -f "../config/redis.conf" ]; then
        $SUDO cp ../config/redis.conf /etc/redis/
        echo "✅ Redis 配置文件已复制到 /etc/redis/"
    else
        echo "⚠️  未找到 Redis 配置文件，将使用默认配置"
    fi
    
    # 配置防火墙（如果需要）
    read -p "是否配置防火墙允许 Redis 端口 6379? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # UFW (Ubuntu/Debian)
        if command -v ufw &> /dev/null; then
            $SUDO ufw allow 6379/tcp
            echo "✅ UFW 防火墙规则已添加"
        # Firewalld (CentOS/RHEL/Fedora)
        elif command -v firewall-cmd &> /dev/null; then
            $SUDO firewall-cmd --permanent --add-port=6379/tcp
            $SUDO firewall-cmd --reload
            echo "✅ Firewalld 防火墙规则已添加"
        else
            echo "⚠️  未检测到防火墙管理工具"
        fi
    fi
    
    echo ""
    echo "🚀 安装完成！"
    echo "启动 Redis: ./start_redis_linux.sh"
    echo "停止 Redis: ./stop_redis_linux.sh"
    echo "或者使用 systemctl:"
    echo "  启动: sudo systemctl start redis"
    echo "  停止: sudo systemctl stop redis"
    echo "  重启: sudo systemctl restart redis"
    echo "  状态: sudo systemctl status redis"
    
else
    echo "❌ Redis 安装失败"
    exit 1
fi 
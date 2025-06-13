# Redis 安装和管理脚本

本目录包含了在不同操作系统上安装、启动和停止 Redis 的脚本。

## 📁 脚本文件

### Windows 系统
- `install_redis.bat` - Redis 安装脚本
- `start_redis.bat` - Redis 启动脚本  
- `stop_redis.bat` - Redis 停止脚本

### macOS 系统
- `install_redis_macos.sh` - Redis 安装脚本（使用 Homebrew）
- `start_redis_macos.sh` - Redis 启动脚本
- `stop_redis_macos.sh` - Redis 停止脚本

### Linux 系统
- `install_redis_linux.sh` - Redis 安装脚本（支持多个发行版）
- `start_redis_linux.sh` - Redis 启动脚本
- `stop_redis_linux.sh` - Redis 停止脚本

## 🚀 使用方法

### Windows

```cmd
# 安装 Redis
scripts\install_redis.bat

# 启动 Redis
scripts\start_redis.bat

# 停止 Redis
scripts\stop_redis.bat
```

### macOS

```bash
# 添加执行权限
chmod +x scripts/*.sh

# 安装 Redis
./scripts/install_redis_macos.sh

# 启动 Redis
./scripts/start_redis_macos.sh

# 停止 Redis
./scripts/stop_redis_macos.sh
```

### Linux

```bash
# 添加执行权限
chmod +x scripts/*.sh

# 安装 Redis
./scripts/install_redis_linux.sh

# 启动 Redis
./scripts/start_redis_linux.sh

# 停止 Redis
./scripts/stop_redis_linux.sh
```

## 📋 系统要求

### Windows
- Windows 10/11 或 Windows Server 2016+
- PowerShell 5.0+
- 管理员权限（用于安装）

### macOS
- macOS 10.14+ (Mojave)
- Homebrew（脚本会自动安装）
- Xcode Command Line Tools

### Linux
支持的发行版：
- **Ubuntu/Debian**: 使用 APT 包管理器
- **CentOS/RHEL/Fedora**: 使用 YUM/DNF 包管理器
- **Arch Linux**: 使用 Pacman 包管理器
- **Rocky Linux/AlmaLinux**: 使用 YUM/DNF 包管理器

## ⚙️ 配置文件

所有脚本都会尝试使用项目中的 `config/redis.conf` 配置文件。如果找不到，将使用默认配置。

### 配置文件位置
- **Windows**: `%USERPROFILE%\redis\config\redis.conf`
- **macOS**: `~/redis/config/redis.conf`
- **Linux**: `/etc/redis/redis.conf`

## 📊 服务管理

### macOS (Homebrew)
```bash
# 使用 Homebrew 服务管理
brew services start redis    # 启动
brew services stop redis     # 停止
brew services restart redis  # 重启
brew services list           # 查看状态
```

### Linux (systemd)
```bash
# 使用 systemctl 管理
sudo systemctl start redis-server    # 启动
sudo systemctl stop redis-server     # 停止
sudo systemctl restart redis-server  # 重启
sudo systemctl status redis-server   # 查看状态
sudo systemctl enable redis-server   # 开机启动
sudo systemctl disable redis-server  # 禁用开机启动
```

## 🔧 常用命令

### 连接 Redis
```bash
redis-cli                    # 连接到本地 Redis
redis-cli -h host -p port    # 连接到远程 Redis
redis-cli ping               # 测试连接
```

### 查看信息
```bash
redis-cli info               # 查看 Redis 信息
redis-cli info server        # 查看服务器信息
redis-cli info memory        # 查看内存使用
redis-cli monitor            # 监控命令执行
```

### 数据操作
```bash
redis-cli flushall           # 清空所有数据
redis-cli save               # 手动保存数据
redis-cli shutdown           # 关闭 Redis 服务
```

## 📝 日志文件

### Windows
- 日志位置: `%USERPROFILE%\redis\logs\redis.log`

### macOS
- 日志位置: `~/redis/logs/redis.log`

### Linux
- 系统服务日志: `sudo journalctl -u redis-server -f`
- 文件日志: `/var/log/redis/redis.log`

## 🛠️ 故障排除

### 端口被占用
```bash
# 查看端口占用
netstat -tulpn | grep 6379    # Linux/macOS
netstat -ano | findstr 6379   # Windows

# 杀死占用进程
sudo kill -9 <PID>            # Linux/macOS
taskkill /F /PID <PID>         # Windows
```

### 权限问题
```bash
# Linux/macOS: 确保 Redis 用户有正确权限
sudo chown -R redis:redis /var/lib/redis
sudo chown -R redis:redis /var/log/redis

# 检查 SELinux (CentOS/RHEL)
sudo setsebool -P redis_enable_notify on
```

### 内存不足
```bash
# 检查内存使用
redis-cli info memory

# 设置最大内存限制
redis-cli config set maxmemory 256mb
redis-cli config set maxmemory-policy allkeys-lru
```

## 🔐 安全配置

### 设置密码
在 `redis.conf` 中添加：
```
requirepass your_secure_password
```

### 绑定IP
```
bind 127.0.0.1 ::1  # 仅本地访问
# bind 0.0.0.0      # 允许所有IP访问（不推荐）
```

### 禁用危险命令
```
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command DEBUG ""
```

## 📞 支持

如果遇到问题，请检查：
1. 系统日志和 Redis 日志
2. 防火墙设置
3. 端口占用情况
4. 权限配置
5. 配置文件语法

更多信息请参考 [Redis 官方文档](https://redis.io/documentation)。 
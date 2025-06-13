@echo off
echo 正在启动 Redis 服务器...

REM 检查 Redis 是否已安装
where redis-server >nul 2>nul
if %errorlevel% neq 0 (
    echo Redis 未安装，请先安装 Redis
    exit /b 1
)

REM 启动 Redis 服务器
redis-server ..\config\redis.conf

echo Redis 服务器已启动 
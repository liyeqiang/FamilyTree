@echo off
echo 正在停止 Redis 服务器...

REM 检查 Redis 是否已安装
where redis-cli >nul 2>nul
if %errorlevel% neq 0 (
    echo Redis 未安装，请先安装 Redis
    exit /b 1
)

REM 停止 Redis 服务器
redis-cli -a your_secure_password shutdown

echo Redis 服务器已停止 
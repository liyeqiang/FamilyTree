@echo off
echo 正在安装 Redis...

REM 检查是否已安装 Redis
where redis-server >nul 2>nul
if %errorlevel% equ 0 (
    echo Redis 已经安装
    exit /b 0
)

REM 下载 Redis
echo 正在下载 Redis...
powershell -Command "& {Invoke-WebRequest -Uri 'https://github.com/microsoftarchive/redis/releases/download/win-3.0.504/Redis-x64-3.0.504.msi' -OutFile 'Redis-x64-3.0.504.msi'}"

REM 安装 Redis
echo 正在安装 Redis...
msiexec /i Redis-x64-3.0.504.msi /qn

REM 等待安装完成
timeout /t 30

REM 删除安装文件
del Redis-x64-3.0.504.msi

echo Redis 安装完成 
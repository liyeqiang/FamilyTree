@echo off
echo 正在执行数据库迁移...

REM 检查 SQLite 是否已安装
where sqlite3 >nul 2>nul
if %errorlevel% neq 0 (
    echo SQLite 未安装，请先安装 SQLite
    exit /b 1
)

REM 执行迁移脚本
sqlite3 familytree.db < ..\sql\migrations\add_burial_place.sql

echo 数据库迁移完成 
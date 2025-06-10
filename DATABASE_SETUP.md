# 数据库设置指南

## 选项1: 使用MySQL

### 1. 安装MySQL
- 下载并安装 [MySQL Community Server](https://dev.mysql.com/downloads/mysql/)
- 或者使用XAMPP、WAMP等集成环境

### 2. 启动MySQL服务
- Windows: 在服务管理器中启动MySQL服务
- 或者使用XAMPP控制面板启动MySQL

### 3. 创建数据库
```sql
-- 连接到MySQL (使用命令行或GUI工具如MySQL Workbench)
mysql -u root -p

-- 创建数据库
CREATE DATABASE familytree CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户（可选）
CREATE USER 'familytree_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON familytree.* TO 'familytree_user'@'localhost';
FLUSH PRIVILEGES;
```

### 4. 初始化数据库表
```bash
# 进入项目目录
cd D:\Orange\coding\FamilyTree

# 导入数据库结构
mysql -u root -p familytree < sql/init.sql
```

### 5. 设置环境变量（可选）
创建 `.env` 文件或设置环境变量：
```bash
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_NAME=familytree
DB_USER=root
DB_PASSWORD=your_mysql_password
```

## 选项2: 使用SQLite（简单测试）

如果只是想快速测试，我们可以修改代码使用SQLite：

1. 修改 `go.mod` 添加SQLite驱动
2. 使用内存数据库进行测试

## 选项3: 使用Docker

```bash
# 运行MySQL容器
docker run --name familytree-mysql \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -e MYSQL_DATABASE=familytree \
  -e MYSQL_USER=familytree_user \
  -e MYSQL_PASSWORD=userpassword \
  -p 3306:3306 \
  -d mysql:8.0

# 等待MySQL启动完成
sleep 30

# 导入数据库结构
docker exec -i familytree-mysql mysql -u root -prootpassword familytree < sql/init.sql
```

## 快速测试连接

运行测试连接程序：
```bash
go run test_db.go
``` 
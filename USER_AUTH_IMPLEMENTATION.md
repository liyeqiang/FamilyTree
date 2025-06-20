# 家族树系统用户认证实现说明

## 📋 功能概述

我已经为您的家族树系统添加了完整的用户认证和多用户支持功能，包括：

### ✅ 已实现功能

1. **用户管理系统**
   - 用户注册和登录
   - JWT令牌认证
   - 密码加密存储（bcrypt）
   - 用户资料管理
   - 密码修改功能

2. **家族树隔离**
   - 每个用户拥有独立的家族树
   - 用户只能访问自己的数据
   - 支持多个家族树管理
   - 默认家族树设置

3. **安全认证**
   - JWT令牌验证
   - 认证中间件保护API
   - 自动令牌刷新
   - 安全的密码存储

## 🏗️ 系统架构

### 数据库结构

#### 新增表
```sql
-- 用户表
users (
    user_id, username, email, password, 
    full_name, avatar, is_active, 
    created_at, updated_at
)

-- 用户家族树关联表
user_family_trees (
    user_id, family_tree_id, family_tree_name, 
    description, root_person_id, is_default,
    created_at, updated_at
)
```

#### 修改的表
所有现有表都添加了 `user_id` 和 `family_tree_id` 字段以支持多用户隔离。

### 代码结构

```
models/
├── models.go          # 添加了用户和认证相关的数据模型

interfaces/
├── interfaces.go      # 新增用户和认证服务接口

services/
├── auth_service.go    # 认证服务（注册、登录、令牌管理）
├── user_service.go    # 用户服务（用户信息管理）

handlers/
├── auth_handler.go    # 认证HTTP处理器

repository/
├── user_repository.go # 用户数据访问层

pkg/middleware/
├── auth.go           # JWT认证中间件

static/
├── login.html        # 用户登录注册页面
```

## 🔧 API 接口

### 认证相关接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/auth/register` | POST | 用户注册 |
| `/api/v1/auth/login` | POST | 用户登录 |
| `/api/v1/auth/refresh` | POST | 刷新令牌 |
| `/api/v1/auth/logout` | POST | 用户登出 |

### 用户管理接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/v1/user/profile` | GET | 获取用户资料 |
| `/api/v1/user/profile` | PUT | 更新用户资料 |
| `/api/v1/user/password` | PUT | 修改密码 |
| `/api/v1/user/validate` | GET | 验证令牌 |

### 现有接口保护
所有家族树相关的接口现在都需要认证，并且只能访问用户自己的数据。

## 🎯 使用指南

### 1. 启动应用
```bash
go build -o familytree.exe
./familytree.exe
```

### 2. 访问登录页面
打开浏览器访问：`http://localhost:8080/static/login.html`

### 3. 注册新用户
- 填写用户名、邮箱、姓名和密码
- 系统会自动为新用户创建默认家族树

### 4. 登录系统
- 支持用户名或邮箱登录
- 登录成功后获得JWT令牌

### 5. 使用API
在HTTP请求头中添加：
```
Authorization: Bearer <your-jwt-token>
```

## 🔐 安全特性

1. **密码安全**
   - 使用bcrypt哈希算法
   - 密码不会以明文存储或传输

2. **JWT令牌**
   - 24小时有效期
   - 支持刷新令牌（7天有效期）
   - 安全的签名验证

3. **数据隔离**
   - 每个用户只能访问自己的数据
   - 中间件级别的权限控制

4. **输入验证**
   - 邮箱格式验证
   - 密码长度要求
   - 用户名唯一性检查

## 🧪 测试用例

### 演示用户
系统包含一个演示用户：
- 用户名：`demo`
- 邮箱：`demo@example.com`
- 密码：`demo123`

### API测试示例

1. **注册新用户**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "email": "newuser@example.com", 
    "password": "password123",
    "full_name": "New User"
  }'
```

2. **用户登录**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "password": "password123"
  }'
```

3. **访问用户资料**
```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer <your-token>"
```

## 🚀 下一步扩展

### 可以进一步实现的功能：

1. **家族树管理API**
   - 创建新家族树
   - 切换默认家族树
   - 家族树分享功能

2. **权限系统**
   - 家族树访问权限
   - 协作编辑功能
   - 只读权限分享

3. **高级功能**
   - 用户头像上传
   - 邮箱验证
   - 密码重置
   - 双因素认证

4. **数据迁移**
   - 现有数据关联到用户
   - 批量导入功能
   - 数据备份恢复

## 📝 配置说明

### JWT密钥
在生产环境中，请修改 `pkg/middleware/auth.go` 中的 `JWTSecret` 为安全的随机密钥：

```go
var JWTSecret = []byte("your-secure-secret-key-for-production")
```

### 数据库迁移
系统启动时会自动执行数据库迁移，创建必要的用户表和索引。

## ❓ 常见问题

**Q: 如何重置用户密码？**
A: 目前需要直接修改数据库，后续可以实现邮箱重置功能。

**Q: 现有数据会受影响吗？**
A: 不会，现有数据会自动关联到演示用户，保持兼容性。

**Q: 如何备份用户数据？**
A: SQLite数据库文件包含所有数据，定期备份 `familytree.db` 文件即可。

## 🎉 结语

您的家族树系统现在已经具备了完整的多用户支持！用户可以：
- 注册自己的账户
- 登录后创建和管理自己的家族树
- 安全地访问只属于自己的数据
- 通过美观的Web界面进行操作

系统采用了现代化的JWT认证机制，确保数据安全和用户体验。您可以继续在这个基础上添加更多功能，如家族树分享、协作编辑等高级特性。 
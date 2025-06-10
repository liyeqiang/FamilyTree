# 🌳 家谱系统 (Family Tree System)

一个功能完整的家谱管理系统，支持个人信息管理、家族关系查询、家族树构建等功能。

## ✨ 特性

- 📊 **完整的个人信息管理** - 姓名、性别、出生/死亡日期、职业等
- 👨‍👩‍👧‍👦 **家族关系管理** - 父母、子女、兄弟姐妹、配偶关系
- 🌲 **家族树构建** - 可视化家族结构，支持多代查询
- 🔍 **智能搜索** - 按姓名、职业、备注等字段搜索
- 📄 **分页查询** - 高效处理大量数据
- 🚀 **双模式运行** - 内存演示模式 + SQLite持久化模式
- 🌐 **RESTful API** - 标准HTTP接口，支持CORS
- 📱 **响应式界面** - 现代化Web界面，支持移动设备

## 🚀 快速开始

### 运行模式

系统支持两种运行模式：

#### 1. 演示模式（推荐）
无需数据库配置，数据存储在内存中，适合演示和测试：
```bash
go run main.go demo
# 或者
go run main.go
```

#### 2. SQLite模式
数据持久化存储在SQLite数据库中：
```bash
# 需要CGO支持，Windows下可能需要额外配置
go run main.go sqlite
```

### 访问系统

启动后访问：
- **主页**: http://localhost:8080
- **API文档**: http://localhost:8080
- **健康检查**: http://localhost:8080/health

## 📖 API 文档

### 个人信息管理

| 方法 | 路径 | 说明 |
|-----|------|------|
| `GET` | `/api/v1/individuals` | 获取所有个人信息 |
| `POST` | `/api/v1/individuals` | 创建个人信息 |
| `GET` | `/api/v1/individuals/{id}` | 获取指定个人信息 |
| `PUT` | `/api/v1/individuals/{id}` | 更新个人信息 |
| `DELETE` | `/api/v1/individuals/{id}` | 删除个人信息 |

### 家族关系查询

| 方法 | 路径 | 说明 |
|-----|------|------|
| `GET` | `/api/v1/individuals/{id}/children` | 获取子女 |
| `GET` | `/api/v1/individuals/{id}/parents` | 获取父母 |
| `GET` | `/api/v1/individuals/{id}/siblings` | 获取兄弟姐妹 |
| `GET` | `/api/v1/individuals/{id}/spouses` | 获取配偶 |
| `GET` | `/api/v1/individuals/{id}/ancestors` | 获取祖先 |
| `GET` | `/api/v1/individuals/{id}/descendants` | 获取后代 |
| `GET` | `/api/v1/individuals/{id}/family-tree` | 获取家族树 |

## 📊 示例数据

系统预置了以下示例数据：

- **张伟** (ID: 1) - 工程师，1950年出生，家族族长
- **李丽** (ID: 2) - 教师，1955年出生，张伟的妻子
- **张明** (ID: 3) - 医生，1975年出生，张伟和李丽的儿子
- **王美** (ID: 4) - 护士，1978年出生，张明的妻子
- **张小宝** (ID: 5) - 2005年出生，张明和王美的儿子

## 🛠️ 项目结构

```
FamilyTree/
├── main.go              # 主程序入口（合并版本）
├── init_db.go          # 数据库初始化工具
├── go.mod              # Go模块文件
├── env.example         # 环境变量示例
├── models/             # 数据模型
├── interfaces/         # 接口定义
├── repository/         # 数据访问层
├── services/           # 业务逻辑层
├── handlers/           # HTTP处理器
├── config/             # 配置管理
└── sql/               # SQL脚本
```

## 🔧 开发说明

### 环境要求

- Go 1.21+
- SQLite模式需要CGO支持

### 依赖

```go
require (
    github.com/gorilla/mux v1.8.0
    github.com/mattn/go-sqlite3 v1.14.28
)
```

### 编译

```bash
# 演示模式（无需CGO）
go build -o familytree main.go

# SQLite模式（需要CGO）
CGO_ENABLED=1 go build -o familytree main.go
```

## 🌟 核心功能

### 1. 数据验证
- 姓名必填验证
- 性别一致性验证（父亲必须是男性，母亲必须是女性）
- 家族关系循环检测
- 数据完整性约束

### 2. 关系查询
- 支持多代祖先/后代查询
- 兄弟姐妹关系自动去重
- 配偶关系管理
- 家族树递归构建

### 3. 搜索功能
- 模糊搜索支持
- 多字段搜索（姓名、职业、备注）
- 分页查询优化

### 4. API设计
- RESTful风格
- 统一错误处理
- CORS跨域支持
- 请求日志记录

## 📝 更新日志

### v2.0.0 (最新)
- ✅ 合并演示模式和SQLite模式到单个可执行文件
- ✅ 命令行参数支持模式选择
- ✅ 统一的API接口和路由配置
- ✅ 优化的错误处理和类型安全
- ✅ 现代化的Web界面

### v1.0.0
- ✅ 基础个人信息管理
- ✅ 家族关系查询
- ✅ SQLite数据库支持
- ✅ RESTful API

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📄 许可证

MIT License

---

**享受构建你的数字家族树！** 🌳✨


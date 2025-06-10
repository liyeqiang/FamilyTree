# 族谱系统使用说明

这是一个用Go语言实现的族谱管理系统，提供了完整的API来管理家族成员信息、关系和事件。

## 项目结构

```
familytree/
├── config/           # 配置文件
│   └── database.go   # 数据库配置
├── models/           # 数据模型
│   └── models.go     # 所有数据结构定义
├── interfaces/       # 接口定义
│   └── interfaces.go # 服务和存储库接口
├── services/         # 业务逻辑层
│   └── individual_service.go # 个人信息服务
├── repository/       # 数据访问层
│   └── mysql_repository.go   # MySQL数据库实现
├── handlers/         # HTTP处理器
│   └── individual_handler.go # 个人信息API处理器
├── sql/              # 数据库脚本
│   └── init.sql      # 数据库初始化脚本
├── go.mod            # Go模块定义
├── main.go           # 主程序入口
├── env.example       # 环境变量示例
└── README.md         # 项目说明
```

## 快速开始

### 1. 环境准备

确保您的系统已安装：
- Go 1.21 或更高版本
- MySQL 5.7 或更高版本

### 2. 克隆项目

```bash
git clone <repository-url>
cd familytree
```

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 数据库设置

1. 创建MySQL数据库：
```sql
CREATE DATABASE familytree CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 运行初始化脚本：
```bash
mysql -u root -p familytree < sql/init.sql
```

### 5. 配置环境变量

复制环境变量示例文件并修改配置：
```bash
cp env.example .env
# 编辑 .env 文件，设置数据库连接信息
```

### 6. 运行应用

```bash
go run main.go
```

服务器将在 `http://localhost:8080` 启动。

## API 使用说明

### 基础URL
```
http://localhost:8080/api/v1
```

### 个人信息管理

#### 创建个人信息
```bash
POST /individuals
Content-Type: application/json

{
    "full_name": "张三",
    "gender": "男",
    "birth_date": "1990-01-15",
    "occupation": "工程师",
    "notes": "备注信息"
}
```

#### 获取个人信息
```bash
GET /individuals/{id}
```

#### 更新个人信息
```bash
PUT /individuals/{id}
Content-Type: application/json

{
    "full_name": "张三丰",
    "occupation": "武林高手"
}
```

#### 删除个人信息
```bash
DELETE /individuals/{id}
```

#### 搜索个人信息
```bash
GET /individuals?q=张&limit=20&offset=0
```

### 家族关系查询

#### 获取子女
```bash
GET /individuals/{id}/children
```

#### 获取父母
```bash
GET /individuals/{id}/parents
```

#### 获取兄弟姐妹
```bash
GET /individuals/{id}/siblings
```

#### 获取配偶
```bash
GET /individuals/{id}/spouses
```

#### 获取祖先
```bash
GET /individuals/{id}/ancestors?generations=3
```

#### 获取后代
```bash
GET /individuals/{id}/descendants?generations=3
```

#### 获取家族树
```bash
GET /individuals/{id}/family-tree?generations=3
```

## 数据模型说明

### 个人信息 (Individual)
- `individual_id`: 个人唯一标识符
- `full_name`: 姓名
- `gender`: 性别（男/女/其他/未知）
- `birth_date`: 出生日期
- `death_date`: 死亡日期
- `occupation`: 职业
- `notes`: 备注信息
- `father_id`, `mother_id`: 父母ID引用

### 家庭关系 (Family)
- `family_id`: 家庭唯一标识符
- `husband_id`, `wife_id`: 夫妻ID引用
- `marriage_date`: 结婚日期
- `divorce_date`: 离婚日期

### 子女关系 (Child)
- `child_id`: 子女关系唯一标识符
- `family_id`: 家庭ID引用
- `individual_id`: 个人ID引用
- `relationship_to_parents`: 与父母的关系类型

### 事件 (Event)
- `event_id`: 事件唯一标识符
- `individual_id`: 关联的个人ID
- `event_type`: 事件类型
- `event_date`: 事件日期
- `description`: 事件描述

### 地点 (Place)
- `place_id`: 地点唯一标识符
- `place_name`: 地点名称
- `latitude`, `longitude`: 经纬度坐标

## 扩展功能

系统设计采用分层架构，便于扩展：

1. **添加新的实体类型**：在 `models/` 中定义新的结构体
2. **添加新的业务逻辑**：在 `services/` 中实现相应的服务
3. **添加新的API接口**：在 `handlers/` 中创建HTTP处理器
4. **支持其他数据库**：在 `repository/` 中实现相应的存储库

## 注意事项

1. **数据一致性**：系统会验证父母关系的逻辑一致性
2. **级联删除**：删除个人时会检查是否有子女引用
3. **性能优化**：数据库已创建必要的索引以提高查询性能
4. **并发安全**：使用数据库事务确保数据一致性

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库服务是否启动
   - 验证连接参数是否正确
   - 确认数据库用户权限

2. **API返回404错误**
   - 检查URL路径是否正确
   - 确认服务器是否正常启动

3. **数据验证错误**
   - 检查请求JSON格式是否正确
   - 验证必填字段是否提供

### 日志查看

应用启动后会在控制台输出详细的请求日志，包括：
- HTTP方法和路径
- 客户端IP地址
- 响应时间

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 许可证

MIT License 
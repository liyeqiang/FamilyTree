1. 数据库表设计

我们将使用关系型数据库来存储祖谱数据。以下是建议的数据库表结构：

Individuals (个人信息表)

| 列名             | 数据类型       | 约束                      | 描述                                   |
| ---------------- | -------------- | ------------------------- | -------------------------------------- |
| individual_id   | INT            | PRIMARY KEY, AUTO_INCREMENT | 个人唯一标识符                           |
| first_name      | VARCHAR(255)   | NOT NULL                  | 名                                       |
| middle_name     | VARCHAR(255)   |                           | 中间名                                   |
| last_name       | VARCHAR(255)   | NOT NULL                  | 姓                                       |
| gender           | ENUM('男', '女', '其他', '未知') |                           | 性别                                     |
| birth_date      | DATE           |                           | 出生日期                                 |
| birth_place_id | INT            | FOREIGN KEY references Places(place_id) | 出生地点ID，关联地点表                     |
| death_date      | DATE           |                           | 死亡日期                                 |
| death_place_id | INT            | FOREIGN KEY references Places(place_id) | 死亡地点ID，关联地点表                     |
| occupation       | VARCHAR(255)   |                           | 职业                                     |
| notes            | TEXT           |                           | 备注信息                                 |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                 |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                 |

Families (家庭关系表)

| 列名             | 数据类型       | 约束                      | 描述                                   |
| ---------------- | -------------- | ------------------------- | -------------------------------------- |
| family_id       | INT            | PRIMARY KEY, AUTO_INCREMENT | 家庭唯一标识符                           |
| husband_id      | INT            | FOREIGN KEY references Individuals(individual_id) | 丈夫ID，关联个人信息表                   |
| wife_id         | INT            | FOREIGN KEY references Individuals(individual_id) | 妻子ID，关联个人信息表                   |
| marriage_date   | DATE           |                           | 结婚日期                                 |
| marriage_place_id | INT            | FOREIGN KEY references Places(place_id) | 结婚地点ID，关联地点表                     |
| divorce_date    | DATE           |                           | 离婚日期                                 |
| notes            | TEXT           |                           | 备注信息                                 |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                 |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                 |   

Children (子女关系表，用于连接家庭和个人)

| 列名             | 数据类型       | 约束                      | 描述                                             |
| ---------------- | -------------- | ------------------------- | ------------------------------------------------ |
| child_id        | INT            | PRIMARY KEY, AUTO_INCREMENT | 子女关系唯一标识符                               |
| family_id       | INT            | FOREIGN KEY references Families(family_id) | 家庭ID，关联家庭关系表                             |
| individual_id   | INT            | FOREIGN KEY references Individuals(individual_id) | 子女个人ID，关联个人信息表                           |
| relationship_to_parents | VARCHAR(50)    |                           | 与父母的关系，例如：亲生、收养、继子/女等 |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                         |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                         |
| UNIQUE KEY unique_child_family (family_id, individual_id) |                           | 确保一个孩子在一个家庭中只记录一次                 |   

Events (事件表，记录个人经历的重要事件)

| 列名             | 数据类型       | 约束                      | 描述                                   |
| ---------------- | -------------- | ------------------------- | -------------------------------------- |
| event_id        | INT            | PRIMARY KEY, AUTO_INCREMENT | 事件唯一标识符                           |
| individual_id   | INT            | FOREIGN KEY references Individuals(individual_id) | 关联的个人ID                             |
| event_type      | VARCHAR(100)   | NOT NULL                  | 事件类型，例如：出生、死亡、婚姻、洗礼、葬礼、户口普查等 |
| event_date      | DATE           |                           | 事件发生日期                             |
| event_place_id | INT            | FOREIGN KEY references Places(place_id) | 事件发生地点ID，关联地点表                 |
| description      | TEXT           |                           | 事件描述                                 |
| notes            | TEXT           |                           | 备注信息                                 |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                 |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                 |   

Places (地点表)   

| 列名             | 数据类型       | 约束                      | 描述                                   |
| ---------------- | -------------- | ------------------------- | -------------------------------------- |
| place_id        | INT            | PRIMARY KEY, AUTO_INCREMENT | 地点唯一标识符                           |
| place_name      | VARCHAR(255)   | NOT NULL, UNIQUE          | 地点名称，例如：省、市、县、村庄等       |
| latitude         | DECIMAL(10, 6) |                           | 纬度                                     |
| longitude        | DECIMAL(10, 6) |                           | 经度                                     |
| notes            | TEXT           |                           | 备注信息                                 |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                 |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                 |   

Sources (信息来源表)

| 列名             | 数据类型       | 约束                      | 描述                                   |
| ---------------- | -------------- | ------------------------- | -------------------------------------- |
| source_id       | INT            | PRIMARY KEY, AUTO_INCREMENT | 来源唯一标识符                           |
| title            | VARCHAR(255)   | NOT NULL                  | 来源标题，例如：出生证明、死亡证明、族谱记录、访谈等 |
| author           | VARCHAR(255)   |                           | 作者/提供者                              |
| publication_year | SMALLINT       |                           | 出版年份                                 |
| publisher        | VARCHAR(255)   |                           | 出版社/机构                              |
| location         | VARCHAR(255)   |                           | 来源存储位置，例如：档案馆、图书馆、URL等 |
| notes            | TEXT           |                           | 备注信息                                 |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                 |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                 |

Citations (引用表，用于连接信息来源和个人、家庭或事件)

| 列名             | 数据类型       | 约束                      | 描述                                             |
| ---------------- | -------------- | ------------------------- | ------------------------------------------------ |
| citation_id     | INT            | PRIMARY KEY, AUTO_INCREMENT | 引用唯一标识符                                   |
| source_id       | INT            | FOREIGN KEY references Sources(source_id) | 来源ID，关联信息来源表                             |
| entity_type     | ENUM('Individual', 'Family', 'Event') | NOT NULL                  | 引用的实体类型，可以是个人、家庭或事件         |
| entity_id       | INT            | NOT NULL                  | 引用的实体ID，根据 entity_type 决定关联的表     |
| page_number     | VARCHAR(50)    |                           | 页码/引用位置                                  |
| notes            | TEXT           |                           | 备注信息                                         |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                         |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                         |
| INDEX (entity_type, entity_id) |                           | 为快速查找特定实体的引用创建索引                 |

Notes (通用备注表，用于存储不属于特定列的额外信息)

| 列名             | 数据类型       | 约束                      | 描述                                             |
| ---------------- | -------------- | ------------------------- | ------------------------------------------------ |
| note_id         | INT            | PRIMARY KEY, AUTO_INCREMENT | 备注唯一标识符                                   |
| entity_type     | ENUM('Individual', 'Family', 'Event', 'Source', 'Place') | NOT NULL                  | 备注关联的实体类型                               |
| entity_id       | INT            | NOT NULL                  | 备注关联的实体ID                                 |
| note_text       | TEXT           | NOT NULL                  | 备注内容                                         |
| created_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP  | 创建时间                                         |
| updated_at      | TIMESTAMP      | DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                                         |
| INDEX (entity_type, entity_id) |                           | 为快速查找特定实体的备注创建索引                 |


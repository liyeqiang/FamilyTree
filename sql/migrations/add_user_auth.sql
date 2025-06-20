-- 用户认证系统迁移脚本
-- 添加用户表和家族树关联表

-- 1. 用户表
CREATE TABLE IF NOT EXISTS users (
    user_id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    full_name TEXT NOT NULL,
    avatar TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 2. 用户家族树关联表
CREATE TABLE IF NOT EXISTS user_family_trees (
    user_id INTEGER NOT NULL,
    family_tree_id INTEGER PRIMARY KEY AUTOINCREMENT,
    family_tree_name TEXT NOT NULL,
    description TEXT,
    root_person_id INTEGER,
    is_default BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (root_person_id) REFERENCES individuals(individual_id) ON DELETE SET NULL
);

-- 3. 为现有的individuals表添加user_id和family_tree_id字段
-- 注意：SQLite不支持ADD COLUMN IF NOT EXISTS，所以我们需要检查列是否存在
-- 这里我们使用更安全的方法：尝试添加列，如果失败就忽略错误

-- 添加用户ID字段到individuals表
-- （在SQLite中，我们需要使用ALTER TABLE）
ALTER TABLE individuals ADD COLUMN user_id INTEGER;
ALTER TABLE individuals ADD COLUMN family_tree_id INTEGER DEFAULT 1;

-- 添加外键约束（注意：SQLite的外键约束在ALTER TABLE中可能不被支持）
-- 我们将在应用层处理这些约束

-- 4. 为families表添加用户关联
ALTER TABLE families ADD COLUMN user_id INTEGER;
ALTER TABLE families ADD COLUMN family_tree_id INTEGER DEFAULT 1;

-- 5. 为其他表添加用户关联
ALTER TABLE events ADD COLUMN user_id INTEGER;
ALTER TABLE events ADD COLUMN family_tree_id INTEGER DEFAULT 1;

ALTER TABLE places ADD COLUMN user_id INTEGER;
ALTER TABLE places ADD COLUMN family_tree_id INTEGER DEFAULT 1;

ALTER TABLE sources ADD COLUMN user_id INTEGER;
ALTER TABLE sources ADD COLUMN family_tree_id INTEGER DEFAULT 1;

-- 6. 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_family_trees_user ON user_family_trees(user_id);
CREATE INDEX IF NOT EXISTS idx_user_family_trees_name ON user_family_trees(family_tree_name);
CREATE INDEX IF NOT EXISTS idx_individuals_user_family ON individuals(user_id, family_tree_id);
CREATE INDEX IF NOT EXISTS idx_families_user_family ON families(user_id, family_tree_id);
CREATE INDEX IF NOT EXISTS idx_events_user_family ON events(user_id, family_tree_id);
CREATE INDEX IF NOT EXISTS idx_places_user_family ON places(user_id, family_tree_id);
CREATE INDEX IF NOT EXISTS idx_sources_user_family ON sources(user_id, family_tree_id);

-- 7. 创建触发器更新updated_at字段
CREATE TRIGGER IF NOT EXISTS update_users_updated_at 
    AFTER UPDATE ON users
    FOR EACH ROW 
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE user_id = NEW.user_id;
END;

CREATE TRIGGER IF NOT EXISTS update_user_family_trees_updated_at 
    AFTER UPDATE ON user_family_trees
    FOR EACH ROW 
BEGIN
    UPDATE user_family_trees SET updated_at = CURRENT_TIMESTAMP WHERE family_tree_id = NEW.family_tree_id;
END;

-- 8. 插入默认的演示用户（可选）
-- 密码是 'demo123' 的bcrypt哈希值
INSERT OR IGNORE INTO users (user_id, username, email, password, full_name, is_active) VALUES
(1, 'demo', 'demo@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iKXYoM.UvTrAAWHJoMIXVjHmU3jG', '演示用户', 1);

-- 9. 为演示用户创建默认家族树
INSERT OR IGNORE INTO user_family_trees (user_id, family_tree_id, family_tree_name, description, is_default) VALUES
(1, 1, '张氏家族', '演示家族树数据', 1);

-- 10. 更新现有数据，将它们关联到演示用户
UPDATE individuals SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE families SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE events SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE places SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE sources SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL; 
-- User authentication system migration (English version)

-- 1. Users table
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

-- 2. User family trees association table
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

-- 3. Add user_id and family_tree_id to existing tables
ALTER TABLE individuals ADD COLUMN user_id INTEGER;
ALTER TABLE individuals ADD COLUMN family_tree_id INTEGER DEFAULT 1;

ALTER TABLE families ADD COLUMN user_id INTEGER;
ALTER TABLE families ADD COLUMN family_tree_id INTEGER DEFAULT 1;

ALTER TABLE events ADD COLUMN user_id INTEGER;
ALTER TABLE events ADD COLUMN family_tree_id INTEGER DEFAULT 1;

ALTER TABLE places ADD COLUMN user_id INTEGER;
ALTER TABLE places ADD COLUMN family_tree_id INTEGER DEFAULT 1;

ALTER TABLE sources ADD COLUMN user_id INTEGER;
ALTER TABLE sources ADD COLUMN family_tree_id INTEGER DEFAULT 1;

-- 4. Create indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_user_family_trees_user ON user_family_trees(user_id);
CREATE INDEX IF NOT EXISTS idx_individuals_user_family ON individuals(user_id, family_tree_id);
CREATE INDEX IF NOT EXISTS idx_families_user_family ON families(user_id, family_tree_id);

-- 5. Create triggers
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

-- 6. Insert demo user (password is 'demo123' hashed with bcrypt)
INSERT OR IGNORE INTO users (user_id, username, email, password, full_name, is_active) VALUES
(1, 'demo', 'demo@example.com', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iKXYoM.UvTrAAWHJoMIXVjHmU3jG', 'Demo User', 1);

-- 7. Create default family tree for demo user
INSERT OR IGNORE INTO user_family_trees (user_id, family_tree_id, family_tree_name, description, is_default) VALUES
(1, 1, 'Zhang Family Tree', 'Demo family tree data', 1);

-- 8. Update existing data to associate with demo user
UPDATE individuals SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE families SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE events SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE places SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL;
UPDATE sources SET user_id = 1, family_tree_id = 1 WHERE user_id IS NULL; 
-- SQLite 家谱系统数据库初始化脚本

-- 启用外键约束
PRAGMA foreign_keys = ON;

-- 1. 个人信息表
CREATE TABLE IF NOT EXISTS individuals (
    individual_id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    gender TEXT CHECK(gender IN ('male', 'female', 'unknown')) NOT NULL DEFAULT 'unknown',
    birth_date DATE,
    birth_place_id INTEGER,
    death_date DATE,
    death_place_id INTEGER,
    occupation TEXT,
    notes TEXT,
    photo_url TEXT,
    father_id INTEGER,
    mother_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (birth_place_id) REFERENCES places(place_id),
    FOREIGN KEY (death_place_id) REFERENCES places(place_id),
    FOREIGN KEY (father_id) REFERENCES individuals(individual_id),
    FOREIGN KEY (mother_id) REFERENCES individuals(individual_id)
);

-- 2. 地点信息表
CREATE TABLE IF NOT EXISTS places (
    place_id INTEGER PRIMARY KEY AUTOINCREMENT,
    place_name TEXT NOT NULL,
    place_type TEXT,
    country TEXT,
    state_province TEXT,
    city TEXT,
    address TEXT,
    latitude REAL,
    longitude REAL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 3. 家庭关系表
CREATE TABLE IF NOT EXISTS families (
    family_id INTEGER PRIMARY KEY AUTOINCREMENT,
    husband_id INTEGER,
    wife_id INTEGER,
    marriage_order INTEGER DEFAULT 1,
    marriage_date DATE,
    marriage_place_id INTEGER,
    divorce_date DATE,
    divorce_place_id INTEGER,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (husband_id) REFERENCES individuals(individual_id),
    FOREIGN KEY (wife_id) REFERENCES individuals(individual_id),
    FOREIGN KEY (marriage_place_id) REFERENCES places(place_id),
    FOREIGN KEY (divorce_place_id) REFERENCES places(place_id)
);

-- 4. 子女关系表
CREATE TABLE IF NOT EXISTS children (
    family_id INTEGER NOT NULL,
    individual_id INTEGER NOT NULL,
    relationship_type TEXT DEFAULT 'biological',
    birth_order INTEGER,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (family_id, individual_id),
    FOREIGN KEY (family_id) REFERENCES families(family_id) ON DELETE CASCADE,
    FOREIGN KEY (individual_id) REFERENCES individuals(individual_id) ON DELETE CASCADE
);

-- 5. 事件表
CREATE TABLE IF NOT EXISTS events (
    event_id INTEGER PRIMARY KEY AUTOINCREMENT,
    individual_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    event_date DATE,
    place_id INTEGER,
    description TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (individual_id) REFERENCES individuals(individual_id) ON DELETE CASCADE,
    FOREIGN KEY (place_id) REFERENCES places(place_id)
);

-- 6. 信息来源表
CREATE TABLE IF NOT EXISTS sources (
    source_id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT,
    publication_date DATE,
    publisher TEXT,
    source_type TEXT,
    repository_name TEXT,
    call_number TEXT,
    description TEXT,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 7. 引用表
CREATE TABLE IF NOT EXISTS citations (
    citation_id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id INTEGER NOT NULL,
    entity_type TEXT NOT NULL CHECK(entity_type IN ('individual', 'family', 'event', 'place')),
    entity_id INTEGER NOT NULL,
    page_number TEXT,
    confidence_level INTEGER CHECK(confidence_level BETWEEN 1 AND 5) DEFAULT 3,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (source_id) REFERENCES sources(source_id) ON DELETE CASCADE
);

-- 8. 备注表
CREATE TABLE IF NOT EXISTS notes (
    note_id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_type TEXT NOT NULL CHECK(entity_type IN ('individual', 'family', 'event', 'place')),
    entity_id INTEGER NOT NULL,
    note_text TEXT NOT NULL,
    note_type TEXT DEFAULT 'general',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_individuals_name ON individuals(full_name);
CREATE INDEX IF NOT EXISTS idx_individuals_father ON individuals(father_id);
CREATE INDEX IF NOT EXISTS idx_individuals_mother ON individuals(mother_id);
CREATE INDEX IF NOT EXISTS idx_individuals_birth_date ON individuals(birth_date);
CREATE INDEX IF NOT EXISTS idx_places_name ON places(place_name);
CREATE INDEX IF NOT EXISTS idx_families_husband ON families(husband_id);
CREATE INDEX IF NOT EXISTS idx_families_wife ON families(wife_id);
CREATE INDEX IF NOT EXISTS idx_events_individual ON events(individual_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_date ON events(event_date);
CREATE INDEX IF NOT EXISTS idx_citations_entity ON citations(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_notes_entity ON notes(entity_type, entity_id);

-- 插入演示数据
-- 地点数据
INSERT OR IGNORE INTO places (place_id, place_name, place_type, country, state_province, city) VALUES
(1, '北京市', 'city', '中国', '北京市', '北京市'),
(2, '上海市', 'city', '中国', '上海市', '上海市'),
(3, '广州市', 'city', '中国', '广东省', '广州市'),
(4, '深圳市', 'city', '中国', '广东省', '深圳市'),
(5, '杭州市', 'city', '中国', '浙江省', '杭州市'),
(6, '南京市', 'city', '中国', '江苏省', '南京市'),
(7, '西安市', 'city', '中国', '陕西省', '西安市'),
(8, '成都市', 'city', '中国', '四川省', '成都市'),
(9, '武汉市', 'city', '中国', '湖北省', '武汉市'),
(10, '天津市', 'city', '中国', '天津市', '天津市');

-- 个人信息数据（包含多代人和复杂关系）
INSERT OR IGNORE INTO individuals (individual_id, full_name, gender, birth_date, death_date, birth_place_id, occupation, notes) VALUES
-- 第一代（祖辈）
(1, '张德高', 'male', '1920-03-15', '1995-08-20', 1, '商人', '张家始祖，经营茶叶生意'),
(2, '李秀英', 'female', '1925-07-10', '2000-12-05', 2, '家庭主妇', '张德高的第一任妻子'),
(3, '王桂花', 'female', '1928-11-22', '2005-03-18', 3, '裁缝', '张德高的第二任妻子'),

-- 第二代（父辈）
(4, '张建国', 'male', '1945-05-01', NULL, 1, '工程师', '张德高与李秀英的长子'),
(5, '张建军', 'male', '1947-09-12', NULL, 1, '教师', '张德高与李秀英的次子'),
(6, '张建华', 'female', '1950-02-28', NULL, 1, '医生', '张德高与李秀英的女儿'),
(7, '张建民', 'male', '1952-12-08', NULL, 1, '农民', '张德高与王桂花的儿子'),
(8, '张建设', 'male', '1955-06-15', NULL, 1, '司机', '张德高与王桂花的儿子'),

-- 第二代的配偶
(9, '陈美丽', 'female', '1948-04-20', NULL, 4, '护士', '张建国的第一任妻子'),
(10, '刘芳', 'female', '1952-08-30', NULL, 5, '会计', '张建国的第二任妻子'),
(11, '赵敏', 'female', '1950-01-15', NULL, 6, '银行职员', '张建军的妻子'),
(12, '李强', 'male', '1948-10-05', NULL, 7, '工人', '张建华的丈夫'),
(13, '孙丽', 'female', '1955-03-25', NULL, 8, '店员', '张建民的妻子'),
(14, '周红', 'female', '1958-07-18', NULL, 9, '厨师', '张建设的妻子'),

-- 第三代（子辈）
(15, '张伟', 'male', '1970-06-10', NULL, 1, '软件工程师', '张建国与陈美丽的儿子'),
(16, '张丽', 'female', '1972-09-15', NULL, 1, '律师', '张建国与陈美丽的女儿'),
(17, '张强', 'male', '1975-12-20', NULL, 1, '医生', '张建国与刘芳的儿子'),
(18, '张敏', 'female', '1978-03-08', NULL, 1, '教师', '张建国与刘芳的女儿'),
(19, '张军', 'male', '1973-05-12', NULL, 1, '警察', '张建军与赵敏的儿子'),
(20, '张华', 'female', '1976-11-25', NULL, 1, '设计师', '张建军与赵敏的女儿'),
(21, '李明', 'male', '1971-08-30', NULL, 7, '程序员', '张建华与李强的儿子'),
(22, '李娜', 'female', '1974-01-18', NULL, 7, '翻译', '张建华与李强的女儿'),
(23, '张勇', 'male', '1977-04-22', NULL, 1, '销售员', '张建民与孙丽的儿子'),
(24, '张静', 'female', '1980-10-14', NULL, 1, '会计', '张建民与孙丽的女儿'),
(25, '张涛', 'male', '1982-07-05', NULL, 1, '司机', '张建设与周红的儿子'),

-- 第三代的配偶
(26, '王美', 'female', '1972-04-15', NULL, 2, '护士', '张伟的妻子'),
(27, '陈刚', 'male', '1970-11-20', NULL, 3, '经理', '张丽的丈夫'),
(28, '李雪', 'female', '1977-02-28', NULL, 4, '医生', '张强的妻子'),
(29, '刘涛', 'male', '1976-09-10', NULL, 5, '工程师', '张敏的丈夫'),
(30, '赵琳', 'female', '1975-06-18', NULL, 6, '记者', '张军的妻子'),
(31, '孙伟', 'male', '1974-12-03', NULL, 7, '商人', '张华的丈夫'),
(32, '周芳', 'female', '1973-08-25', NULL, 8, '老师', '李明的妻子'),
(33, '吴强', 'male', '1972-05-14', NULL, 9, '律师', '李娜的丈夫'),
(34, '马丽', 'female', '1979-01-30', NULL, 10, '销售', '张勇的妻子'),
(35, '何军', 'male', '1978-11-08', NULL, 1, '技术员', '张静的丈夫'),
(36, '郑美', 'female', '1984-03-12', NULL, 2, '文员', '张涛的妻子'),

-- 第四代（孙辈）
(37, '张小明', 'male', '1995-08-20', NULL, 1, '学生', '张伟与王美的儿子'),
(38, '张小丽', 'female', '1998-12-15', NULL, 1, '学生', '张伟与王美的女儿'),
(39, '陈小强', 'male', '1996-05-10', NULL, 3, '学生', '张丽与陈刚的儿子'),
(40, '陈小敏', 'female', '1999-09-25', NULL, 3, '学生', '张丽与陈刚的女儿'),
(41, '张小华', 'male', '2000-02-14', NULL, 1, '学生', '张强与李雪的儿子'),
(42, '刘小雨', 'female', '2001-07-08', NULL, 5, '学生', '张敏与刘涛的女儿'),
(43, '张小军', 'male', '1997-11-30', NULL, 1, '学生', '张军与赵琳的儿子'),
(44, '孙小花', 'female', '2002-04-18', NULL, 7, '学生', '张华与孙伟的女儿'),
(45, '李小东', 'male', '1999-10-22', NULL, 7, '学生', '李明与周芳的儿子'),
(46, '吴小燕', 'female', '2003-01-05', NULL, 9, '学生', '李娜与吴强的女儿'),
(47, '张小勇', 'male', '2005-06-12', NULL, 1, '学生', '张勇与马丽的儿子'),
(48, '何小静', 'female', '2007-09-28', NULL, 1, '学生', '张静与何军的女儿'),
(49, '张小涛', 'male', '2010-03-15', NULL, 1, '学生', '张涛与郑美的儿子');

-- 更新父母关系
-- 第二代的父母关系
UPDATE individuals SET father_id = 1, mother_id = 2 WHERE individual_id IN (4, 5, 6);
UPDATE individuals SET father_id = 1, mother_id = 3 WHERE individual_id IN (7, 8);

-- 第三代的父母关系
UPDATE individuals SET father_id = 4, mother_id = 9 WHERE individual_id IN (15, 16);
UPDATE individuals SET father_id = 4, mother_id = 10 WHERE individual_id IN (17, 18);
UPDATE individuals SET father_id = 5, mother_id = 11 WHERE individual_id IN (19, 20);
UPDATE individuals SET father_id = 12, mother_id = 6 WHERE individual_id IN (21, 22);
UPDATE individuals SET father_id = 7, mother_id = 13 WHERE individual_id IN (23, 24);
UPDATE individuals SET father_id = 8, mother_id = 14 WHERE individual_id = 25;

-- 第四代的父母关系
UPDATE individuals SET father_id = 15, mother_id = 26 WHERE individual_id IN (37, 38);
UPDATE individuals SET father_id = 27, mother_id = 16 WHERE individual_id IN (39, 40);
UPDATE individuals SET father_id = 17, mother_id = 28 WHERE individual_id = 41;
UPDATE individuals SET father_id = 29, mother_id = 18 WHERE individual_id = 42;
UPDATE individuals SET father_id = 19, mother_id = 30 WHERE individual_id = 43;
UPDATE individuals SET father_id = 31, mother_id = 20 WHERE individual_id = 44;
UPDATE individuals SET father_id = 21, mother_id = 32 WHERE individual_id = 45;
UPDATE individuals SET father_id = 33, mother_id = 22 WHERE individual_id = 46;
UPDATE individuals SET father_id = 23, mother_id = 34 WHERE individual_id = 47;
UPDATE individuals SET father_id = 35, mother_id = 24 WHERE individual_id = 48;
UPDATE individuals SET father_id = 25, mother_id = 36 WHERE individual_id = 49;

-- 家庭关系数据（包含多妻制情况）
INSERT OR IGNORE INTO families (family_id, husband_id, wife_id, marriage_order, marriage_date, marriage_place_id, notes) VALUES
-- 第一代家庭
(1, 1, 2, 1, '1943-10-01', 1, '张德高的第一次婚姻'),
(2, 1, 3, 2, '1951-05-15', 1, '张德高的第二次婚姻'),

-- 第二代家庭
(3, 4, 9, 1, '1968-03-20', 1, '张建国的第一次婚姻'),
(4, 4, 10, 2, '1974-11-08', 1, '张建国的第二次婚姻'),
(5, 5, 11, 1, '1972-06-15', 1, '张建军与赵敏的婚姻'),
(6, 12, 6, 1, '1970-09-25', 7, '李强与张建华的婚姻'),
(7, 7, 13, 1, '1976-04-12', 1, '张建民与孙丽的婚姻'),
(8, 8, 14, 1, '1980-08-30', 1, '张建设与周红的婚姻'),

-- 第三代家庭
(9, 15, 26, 1, '1994-05-20', 1, '张伟与王美的婚姻'),
(10, 27, 16, 1, '1995-09-10', 3, '陈刚与张丽的婚姻'),
(11, 17, 28, 1, '1999-07-18', 4, '张强与李雪的婚姻'),
(12, 29, 18, 1, '2000-12-25', 5, '刘涛与张敏的婚姻'),
(13, 19, 30, 1, '1996-10-14', 6, '张军与赵琳的婚姻'),
(14, 31, 20, 1, '1998-03-08', 7, '孙伟与张华的婚姻'),
(15, 21, 32, 1, '1997-11-22', 8, '李明与周芳的婚姻'),
(16, 33, 22, 1, '1998-06-30', 9, '吴强与李娜的婚姻'),
(17, 23, 34, 1, '2002-04-15', 10, '张勇与马丽的婚姻'),
(18, 35, 24, 1, '2003-08-20', 1, '何军与张静的婚姻'),
(19, 25, 36, 1, '2008-12-12', 2, '张涛与郑美的婚姻');

-- 子女关系数据
INSERT OR IGNORE INTO children (family_id, individual_id, relationship_type, birth_order) VALUES
-- 第一代的子女
(1, 4, 'biological', 1),
(1, 5, 'biological', 2),
(1, 6, 'biological', 3),
(2, 7, 'biological', 1),
(2, 8, 'biological', 2),

-- 第二代的子女
(3, 15, 'biological', 1),
(3, 16, 'biological', 2),
(4, 17, 'biological', 1),
(4, 18, 'biological', 2),
(5, 19, 'biological', 1),
(5, 20, 'biological', 2),
(6, 21, 'biological', 1),
(6, 22, 'biological', 2),
(7, 23, 'biological', 1),
(7, 24, 'biological', 2),
(8, 25, 'biological', 1),

-- 第三代的子女
(9, 37, 'biological', 1),
(9, 38, 'biological', 2),
(10, 39, 'biological', 1),
(10, 40, 'biological', 2),
(11, 41, 'biological', 1),
(12, 42, 'biological', 1),
(13, 43, 'biological', 1),
(14, 44, 'biological', 1),
(15, 45, 'biological', 1),
(16, 46, 'biological', 1),
(17, 47, 'biological', 1),
(18, 48, 'biological', 1),
(19, 49, 'biological', 1);

-- 事件数据
INSERT OR IGNORE INTO events (individual_id, event_type, event_date, place_id, description) VALUES
-- 出生事件
(1, 'birth', '1920-03-15', 1, '张德高出生'),
(2, 'birth', '1925-07-10', 2, '李秀英出生'),
(3, 'birth', '1928-11-22', 3, '王桂花出生'),
(4, 'birth', '1945-05-01', 1, '张建国出生'),
(15, 'birth', '1970-06-10', 1, '张伟出生'),
(37, 'birth', '1995-08-20', 1, '张小明出生'),

-- 婚姻事件
(1, 'marriage', '1943-10-01', 1, '张德高与李秀英结婚'),
(2, 'marriage', '1943-10-01', 1, '李秀英与张德高结婚'),
(1, 'marriage', '1951-05-15', 1, '张德高与王桂花结婚'),
(3, 'marriage', '1951-05-15', 1, '王桂花与张德高结婚'),
(4, 'marriage', '1968-03-20', 1, '张建国与陈美丽结婚'),
(9, 'marriage', '1968-03-20', 1, '陈美丽与张建国结婚'),
(4, 'marriage', '1974-11-08', 1, '张建国与刘芳结婚'),
(10, 'marriage', '1974-11-08', 1, '刘芳与张建国结婚'),
(15, 'marriage', '1994-05-20', 1, '张伟与王美结婚'),
(26, 'marriage', '1994-05-20', 1, '王美与张伟结婚'),

-- 死亡事件
(1, 'death', '1995-08-20', 1, '张德高去世'),
(2, 'death', '2000-12-05', 2, '李秀英去世'),
(3, 'death', '2005-03-18', 3, '王桂花去世'),

-- 教育事件
(15, 'education', '1988-09-01', 1, '张伟开始上小学'),
(15, 'education', '1994-09-01', 1, '张伟大学毕业'),
(37, 'education', '2001-09-01', 1, '张小明开始上小学'),
(37, 'education', '2013-09-01', 1, '张小明开始上高中'),

-- 职业事件
(15, 'career', '1994-10-01', 1, '张伟开始工作'),
(4, 'career', '1965-07-01', 1, '张建国开始工作'),

-- 其他重要事件
(1, 'business', '1950-01-01', 1, '张德高创办茶叶生意'),
(15, 'achievement', '2010-05-15', 1, '张伟获得优秀员工奖'),
(37, 'achievement', '2013-06-01', 1, '张小明中考优秀');

-- 创建更新时间触发器
CREATE TRIGGER IF NOT EXISTS update_individuals_updated_at 
    AFTER UPDATE ON individuals
    FOR EACH ROW 
BEGIN
    UPDATE individuals SET updated_at = CURRENT_TIMESTAMP WHERE individual_id = NEW.individual_id;
END;

CREATE TRIGGER IF NOT EXISTS update_places_updated_at 
    AFTER UPDATE ON places
    FOR EACH ROW 
BEGIN
    UPDATE places SET updated_at = CURRENT_TIMESTAMP WHERE place_id = NEW.place_id;
END;

CREATE TRIGGER IF NOT EXISTS update_families_updated_at 
    AFTER UPDATE ON families
    FOR EACH ROW 
BEGIN
    UPDATE families SET updated_at = CURRENT_TIMESTAMP WHERE family_id = NEW.family_id;
END;

CREATE TRIGGER IF NOT EXISTS update_events_updated_at 
    AFTER UPDATE ON events
    FOR EACH ROW 
BEGIN
    UPDATE events SET updated_at = CURRENT_TIMESTAMP WHERE event_id = NEW.event_id;
END;

CREATE TRIGGER IF NOT EXISTS update_sources_updated_at 
    AFTER UPDATE ON sources
    FOR EACH ROW 
BEGIN
    UPDATE sources SET updated_at = CURRENT_TIMESTAMP WHERE source_id = NEW.source_id;
END;

CREATE TRIGGER IF NOT EXISTS update_citations_updated_at 
    AFTER UPDATE ON citations
    FOR EACH ROW 
BEGIN
    UPDATE citations SET updated_at = CURRENT_TIMESTAMP WHERE citation_id = NEW.citation_id;
END;

CREATE TRIGGER IF NOT EXISTS update_notes_updated_at 
    AFTER UPDATE ON notes
    FOR EACH ROW 
BEGIN
    UPDATE notes SET updated_at = CURRENT_TIMESTAMP WHERE note_id = NEW.note_id;
END; 
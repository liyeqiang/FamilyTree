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

-- 插入示例数据
INSERT OR IGNORE INTO places (place_id, place_name, place_type, country, state_province, city) VALUES
(1, '北京市', 'city', '中国', '北京市', '北京市'),
(2, '上海市', 'city', '中国', '上海市', '上海市'),
(3, '广州市', 'city', '中国', '广东省', '广州市');

INSERT OR IGNORE INTO individuals (individual_id, full_name, gender, birth_date, birth_place_id, occupation, notes) VALUES
(1, '张伟', 'male', '1950-01-15', 1, '工程师', '家族族长'),
(2, '李丽', 'female', '1955-03-20', 2, '教师', '张伟的妻子'),
(3, '张明', 'male', '1975-06-10', 1, '医生', '张伟和李丽的儿子'),
(4, '王美', 'female', '1978-09-15', 3, '护士', '张明的妻子'),
(5, '张小宝', 'male', '2005-12-25', 1, '', '张明和王美的儿子');

-- 更新父母关系
UPDATE individuals SET father_id = 1, mother_id = 2 WHERE individual_id = 3;
UPDATE individuals SET father_id = 3, mother_id = 4 WHERE individual_id = 5;

INSERT OR IGNORE INTO families (family_id, husband_id, wife_id, marriage_order, marriage_date, marriage_place_id) VALUES
(1, 1, 2, 1, '1974-05-01', 1),
(2, 3, 4, 1, '2003-10-08', 1);

INSERT OR IGNORE INTO children (family_id, individual_id, relationship_type, birth_order) VALUES
(1, 3, 'biological', 1),
(2, 5, 'biological', 1);

INSERT OR IGNORE INTO events (individual_id, event_type, event_date, place_id, description) VALUES
(1, 'birth', '1950-01-15', 1, '出生'),
(2, 'birth', '1955-03-20', 2, '出生'),
(3, 'birth', '1975-06-10', 1, '出生'),
(1, 'marriage', '1974-05-01', 1, '与李丽结婚'),
(2, 'marriage', '1974-05-01', 1, '与张伟结婚');

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
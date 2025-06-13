-- 添加 burial_place_id 字段
ALTER TABLE individuals ADD COLUMN burial_place_id INTEGER REFERENCES places(place_id);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_individuals_burial_place ON individuals(burial_place_id); 
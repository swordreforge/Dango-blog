-- 创建文章-标签关联表
CREATE TABLE IF NOT EXISTS passage_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    passage_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (passage_id) REFERENCES passages(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
    UNIQUE(passage_id, tag_id)
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_passage_tags_passage_id ON passage_tags(passage_id);
CREATE INDEX IF NOT EXISTS idx_passage_tags_tag_id ON passage_tags(tag_id);
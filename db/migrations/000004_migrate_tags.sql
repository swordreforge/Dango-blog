-- 迁移脚本：将 passages.tags 中的标签迁移到 passage_tags 关联表

-- 注意：由于当前所有文章的 tags 字段都是空的 '[]'，
-- 这个脚本主要是为了演示如何迁移数据。

-- 如果有文章有标签，可以使用以下SQL进行迁移：
-- INSERT INTO passage_tags (passage_id, tag_id, created_at)
-- SELECT p.id, t.id, datetime('now')
-- FROM passages p
-- JOIN tags t ON json_extract(p.tags, '$') LIKE '%' || t.name || '%'
-- WHERE p.tags != '[]' AND p.tags IS NOT NULL AND p.tags != '';

-- 由于当前没有文章标签数据，我们为"测试标签"创建一些示例关联
-- 假设前几篇文章关联到"测试标签"

-- 查找"测试标签"的ID
-- SELECT id FROM tags WHERE name = '测试标签';

-- 为文章ID 1-5 关联到"测试标签"（如果存在）
-- INSERT OR IGNORE INTO passage_tags (passage_id, tag_id, created_at)
-- SELECT p.id, t.id, datetime('now')
-- FROM passages p
-- CROSS JOIN tags t
-- WHERE t.name = '测试标签' AND p.id BETWEEN 1 AND 5;

-- 完成迁移后，可以清空 passages.tags 字段（可选）
-- UPDATE passages SET tags = '[]';
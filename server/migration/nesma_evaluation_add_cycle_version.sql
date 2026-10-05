-- 为NESMA评估表添加项目周期和需求版本关联字段
-- 执行时间：2025-01-17

-- 1. 添加新字段
ALTER TABLE nesma_evaluations 
ADD COLUMN cycle_id BIGINT NOT NULL DEFAULT 0 COMMENT '项目周期ID',
ADD COLUMN requirement_version_id BIGINT NOT NULL DEFAULT 0 COMMENT '需求版本ID';

-- 2. 添加索引
CREATE INDEX idx_nesma_evaluations_cycle_id ON nesma_evaluations(cycle_id);
CREATE INDEX idx_nesma_evaluations_requirement_version_id ON nesma_evaluations(requirement_version_id);

-- 3. 添加外键约束（可选，如果需要强制引用完整性）
-- ALTER TABLE nesma_evaluations 
-- ADD CONSTRAINT fk_nesma_evaluations_cycle_id 
-- FOREIGN KEY (cycle_id) REFERENCES nesma_project_cycles(id) ON DELETE RESTRICT ON UPDATE CASCADE;

-- ALTER TABLE nesma_evaluations 
-- ADD CONSTRAINT fk_nesma_evaluations_requirement_version_id 
-- FOREIGN KEY (requirement_version_id) REFERENCES nesma_requirement_versions(id) ON DELETE RESTRICT ON UPDATE CASCADE;

-- 4. 为现有数据设置默认值（如果需要）
-- 注意：在生产环境中，你可能需要先创建默认的周期和版本，然后更新现有评估记录
-- UPDATE nesma_evaluations SET cycle_id = 1, requirement_version_id = 1 WHERE cycle_id = 0;

-- 5. 移除默认值约束（避免新记录使用0值）
-- ALTER TABLE nesma_evaluations ALTER COLUMN cycle_id DROP DEFAULT;
-- ALTER TABLE nesma_evaluations ALTER COLUMN requirement_version_id DROP DEFAULT;

-- 查询验证
SELECT 
    table_name,
    column_name,
    column_type,
    is_nullable,
    column_default,
    column_comment
FROM 
    information_schema.columns 
WHERE 
    table_schema = DATABASE() 
    AND table_name = 'nesma_evaluations' 
    AND column_name IN ('cycle_id', 'requirement_version_id');
-- 修复现有评估记录的周期和版本关联
-- 执行时间：2025-07-18

-- 第一步：检查现有评估数据
SELECT 
    id, 
    project_id, 
    evaluation_name, 
    cycle_id, 
    requirement_version_id,
    created_at
FROM nesma_evaluations 
WHERE cycle_id IS NULL OR requirement_version_id IS NULL
ORDER BY created_at DESC;

-- 第二步：为每个项目创建默认周期（如果不存在）
INSERT INTO nesma_project_cycles (project_id, name, description, status, phase, created_at, updated_at)
SELECT DISTINCT 
    p.id,
    '默认周期',
    '系统自动创建的默认周期，用于兼容历史评估数据',
    'active',
    'requirement',
    NOW(),
    NOW()
FROM nesma_projects p
WHERE NOT EXISTS (
    SELECT 1 FROM nesma_project_cycles c WHERE c.project_id = p.id
);

-- 第三步：为每个周期创建默认版本（如果不存在）
INSERT INTO nesma_requirement_versions (cycle_id, version, version_type, created_by, summary, status, created_at, updated_at)
SELECT DISTINCT
    c.id,
    'v1.0',
    'initial',
    'system',
    '系统自动创建的默认版本，用于兼容历史评估数据',
    'active',
    NOW(),
    NOW()
FROM nesma_project_cycles c
WHERE NOT EXISTS (
    SELECT 1 FROM nesma_requirement_versions v WHERE v.cycle_id = c.id
);

-- 第四步：更新现有的评估记录
UPDATE nesma_evaluations e
SET 
    cycle_id = (
        SELECT c.id 
        FROM nesma_project_cycles c 
        WHERE c.project_id = e.project_id 
        ORDER BY c.created_at ASC 
        LIMIT 1
    ),
    requirement_version_id = (
        SELECT v.id
        FROM nesma_requirement_versions v
        JOIN nesma_project_cycles c ON v.cycle_id = c.id
        WHERE c.project_id = e.project_id
        ORDER BY v.created_at ASC
        LIMIT 1
    )
WHERE cycle_id IS NULL OR requirement_version_id IS NULL;

-- 第五步：验证修复结果
SELECT 
    '修复后检查' as status,
    COUNT(*) as total_evaluations,
    COUNT(cycle_id) as has_cycle,
    COUNT(requirement_version_id) as has_version,
    COUNT(CASE WHEN cycle_id IS NULL OR requirement_version_id IS NULL THEN 1 END) as still_null
FROM nesma_evaluations;

-- 第六步：显示修复的评估详情
SELECT 
    e.id,
    e.evaluation_name,
    p.name as project_name,
    c.name as cycle_name,
    v.version as version_name,
    e.updated_at
FROM nesma_evaluations e
JOIN nesma_projects p ON e.project_id = p.id
LEFT JOIN nesma_project_cycles c ON e.cycle_id = c.id
LEFT JOIN nesma_requirement_versions v ON e.requirement_version_id = v.id
ORDER BY e.updated_at DESC;
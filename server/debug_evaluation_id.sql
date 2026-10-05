-- 检查NESMA评估记录
SELECT 
    id,
    project_id,
    cycle_id,
    requirement_version_id,
    evaluation_name,
    status,
    created_at
FROM nesma_evaluations 
ORDER BY created_at DESC 
LIMIT 10;

-- 检查功能点记录
SELECT 
    id,
    evaluation_id,
    requirement_id,
    function_type,
    function_name,
    created_at
FROM nesma_function_points 
ORDER BY created_at DESC 
LIMIT 10;

-- 检查评估ID=12是否存在
SELECT COUNT(*) as count FROM nesma_evaluations WHERE id = 12;

-- 查看所有评估ID
SELECT id, evaluation_name, status FROM nesma_evaluations ORDER BY id;
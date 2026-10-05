 -- PostgreSQL pgvector表依赖功能脚本
-- 用于NESMA智能文档生成系统的向量存储功能
-- 注意：此文件在数据表创建之后执行

-- 创建向量查询优化函数
CREATE OR REPLACE FUNCTION vector_similarity_search(
    query_vector vector,
    similarity_threshold float8 DEFAULT 0.7,
    max_results int DEFAULT 20
) RETURNS TABLE (
    id int,
    vector_id varchar,
    content text,
    similarity float8
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        vs.id,
        vs.vector_id,
        vs.content,
        cosine_similarity(vs.embedding::vector, query_vector) as similarity
    FROM nesma_vector_stores vs
    WHERE vs.is_active = true
    AND cosine_similarity(vs.embedding::vector, query_vector) >= similarity_threshold
    ORDER BY vs.embedding::vector <=> query_vector
    LIMIT max_results;
END;
$$ LANGUAGE plpgsql;

-- 创建向量相似度搜索的存储过程
CREATE OR REPLACE FUNCTION hybrid_vector_search(
    query_text text,
    query_vector vector,
    content_type_filter varchar DEFAULT NULL,
    project_id_filter int DEFAULT NULL,
    user_id_filter int DEFAULT NULL,
    similarity_threshold float8 DEFAULT 0.7,
    max_results int DEFAULT 20
) RETURNS TABLE (
    id int,
    vector_id varchar,
    content text,
    content_type varchar,
    similarity float8,
    search_rank float8
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        vs.id,
        vs.vector_id,
        vs.content,
        vs.content_type,
        cosine_similarity(vs.embedding::vector, query_vector) as similarity,
        -- 组合文本搜索和向量搜索的分数
        (cosine_similarity(vs.embedding::vector, query_vector) * 0.7 + 
         CASE 
             WHEN vs.content ILIKE '%' || query_text || '%' THEN 0.3
             ELSE 0.0
         END) as search_rank
    FROM nesma_vector_stores vs
    WHERE vs.is_active = true
    AND (content_type_filter IS NULL OR vs.content_type = content_type_filter)
    AND (project_id_filter IS NULL OR vs.project_id = project_id_filter)
    AND (user_id_filter IS NULL OR vs.user_id = user_id_filter)
    AND cosine_similarity(vs.embedding::vector, query_vector) >= similarity_threshold
    ORDER BY search_rank DESC, vs.embedding::vector <=> query_vector
    LIMIT max_results;
END;
$$ LANGUAGE plpgsql;

-- 创建向量存储统计视图
CREATE OR REPLACE VIEW vector_store_stats AS
SELECT 
    content_type,
    COUNT(*) as total_count,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_count,
    AVG(LENGTH(content)) as avg_content_length,
    MIN(created_at) as first_created,
    MAX(updated_at) as last_updated
FROM nesma_vector_stores
GROUP BY content_type;

-- 创建向量索引状态视图
CREATE OR REPLACE VIEW vector_index_stats AS
SELECT 
    index_name,
    target_table,
    dimensions,
    index_type,
    is_built,
    status,
    build_time,
    CASE 
        WHEN build_time IS NOT NULL THEN 
            EXTRACT(EPOCH FROM (build_time - created_at)) * 1000
        ELSE NULL
    END as build_duration_ms
FROM nesma_vector_indexes
ORDER BY created_at DESC;

-- 创建Agent任务统计视图
CREATE OR REPLACE VIEW agent_task_stats AS
SELECT 
    agent_id,
    task_type,
    status,
    COUNT(*) as task_count,
    AVG(processing_time) as avg_processing_time,
    MIN(processing_time) as min_processing_time,
    MAX(processing_time) as max_processing_time,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count
FROM nesma_agent_tasks
GROUP BY agent_id, task_type, status;

-- 创建会话记忆统计视图（修复JSONB函数问题）
CREATE OR REPLACE VIEW session_memory_stats AS
SELECT 
    user_id,
    COUNT(*) as total_sessions,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_sessions,
    AVG(
        CASE 
            WHEN conversation_history IS NOT NULL THEN 
                jsonb_array_length(conversation_history)
            ELSE 0
        END
    ) as avg_conversation_length,
    MIN(created_at) as first_session,
    MAX(updated_at) as last_activity
FROM nesma_session_memories
GROUP BY user_id;

-- 创建触发器（在表创建后执行）
DO $$
BEGIN
    -- 检查触发器是否已存在，避免重复创建
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.triggers 
        WHERE trigger_name = 'update_vector_store_timestamp'
    ) THEN
        CREATE TRIGGER update_vector_store_timestamp
            BEFORE UPDATE ON nesma_vector_stores
            FOR EACH ROW
            EXECUTE FUNCTION update_vector_embedding_timestamp();
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.triggers 
        WHERE trigger_name = 'update_embedding_timestamp'
    ) THEN
        CREATE TRIGGER update_embedding_timestamp
            BEFORE UPDATE ON nesma_embeddings
            FOR EACH ROW
            EXECUTE FUNCTION update_vector_embedding_timestamp();
    END IF;
END $$;

-- 创建向量索引（可选，根据数据量决定）
-- 注意：这些索引创建可能需要较长时间，可以根据需要手动执行

-- 为向量存储表创建HNSW索引（余弦相似度）
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_vector_stores_hnsw_cosine 
-- ON nesma_vector_stores USING hnsw (embedding vector_cosine_ops);

-- 为向量存储表创建IVFFlat索引（L2距离）
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_vector_stores_ivfflat_l2
-- ON nesma_vector_stores USING ivfflat (embedding vector_l2_ops) WITH (lists = 100);

-- 为嵌入向量表创建索引
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_embeddings_hnsw_cosine
-- ON nesma_embeddings USING hnsw (vector vector_cosine_ops);

-- 创建常用查询的索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_vector_stores_active 
ON nesma_vector_stores (is_active) WHERE is_active = true;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_vector_stores_content_type 
ON nesma_vector_stores (content_type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_vector_stores_user_id 
ON nesma_vector_stores (user_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_vector_stores_project_id 
ON nesma_vector_stores (project_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_agent_tasks_status 
ON nesma_agent_tasks (status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_agent_tasks_agent_id 
ON nesma_agent_tasks (agent_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_session_memories_user_id 
ON nesma_session_memories (user_id);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_nesma_session_memories_active 
ON nesma_session_memories (is_active) WHERE is_active = true;

-- 完成表依赖功能初始化
SELECT 'pgvector table-dependent features initialization completed' as status;
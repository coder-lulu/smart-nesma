 -- PostgreSQL pgvector扩展初始化脚本
-- 用于NESMA智能文档生成系统的向量存储功能

-- 创建pgvector扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 创建向量存储相关的自定义类型和函数
CREATE OR REPLACE FUNCTION cosine_similarity(a vector, b vector) 
RETURNS float8 AS $$
BEGIN
    RETURN 1 - (a <=> b);
END;
$$ LANGUAGE plpgsql;

-- 创建向量存储表的索引
-- 注意：这些索引将在应用启动后根据实际数据创建

-- 为向量存储表创建HNSW索引（余弦相似度）
-- CREATE INDEX CONCURRENTLY ON nesma_vector_stores USING hnsw (embedding vector_cosine_ops);

-- 为向量存储表创建IVFFlat索引（L2距离）
-- CREATE INDEX CONCURRENTLY ON nesma_vector_stores USING ivfflat (embedding vector_l2_ops) WITH (lists = 100);

-- 为嵌入向量表创建索引
-- CREATE INDEX CONCURRENTLY ON nesma_embeddings USING hnsw (vector vector_cosine_ops);

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

-- 创建向量维度验证函数
CREATE OR REPLACE FUNCTION validate_vector_dimension(
    vector_data vector,
    expected_dimension int
) RETURNS boolean AS $$
BEGIN
    RETURN vector_dims(vector_data) = expected_dimension;
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

-- 创建会话记忆统计视图
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

-- 创建触发器函数用于自动更新向量索引
CREATE OR REPLACE FUNCTION update_vector_embedding_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
CREATE TRIGGER update_vector_store_timestamp
    BEFORE UPDATE ON nesma_vector_stores
    FOR EACH ROW
    EXECUTE FUNCTION update_vector_embedding_timestamp();

CREATE TRIGGER update_embedding_timestamp
    BEFORE UPDATE ON nesma_embeddings
    FOR EACH ROW
    EXECUTE FUNCTION update_vector_embedding_timestamp();

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

-- 创建向量数据清理函数
CREATE OR REPLACE FUNCTION cleanup_expired_vectors()
RETURNS int AS $$
DECLARE
    deleted_count int;
BEGIN
    -- 删除过期的向量数据
    DELETE FROM nesma_vector_stores 
    WHERE expires_at IS NOT NULL 
    AND expires_at < NOW();
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    -- 删除过期的会话记忆
    DELETE FROM nesma_session_memories 
    WHERE expires_at < NOW();
    
    -- 删除过期的Agent消息
    DELETE FROM nesma_agent_messages 
    WHERE expires_at IS NOT NULL 
    AND expires_at < NOW();
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 创建定期清理任务
-- 注意：这需要pg_cron扩展，如果不可用则注释掉
-- SELECT cron.schedule('cleanup-vectors', '0 2 * * *', 'SELECT cleanup_expired_vectors();');

-- 创建向量搜索性能优化建议函数
CREATE OR REPLACE FUNCTION vector_search_performance_hints()
RETURNS TABLE (
    suggestion text,
    priority int,
    description text
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        'Create HNSW index' as suggestion,
        1 as priority,
        'Create HNSW index for better vector search performance' as description
    WHERE NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE tablename = 'nesma_vector_stores' 
        AND indexname LIKE '%hnsw%'
    )
    
    UNION ALL
    
    SELECT 
        'Update table statistics' as suggestion,
        2 as priority,
        'Run ANALYZE on vector tables to update statistics' as description
    WHERE (
        SELECT last_analyze 
        FROM pg_stat_user_tables 
        WHERE relname = 'nesma_vector_stores'
    ) < NOW() - INTERVAL '1 day'
    
    UNION ALL
    
    SELECT 
        'Consider partitioning' as suggestion,
        3 as priority,
        'Consider partitioning large vector tables by date or content type' as description
    WHERE (
        SELECT n_tup_ins + n_tup_upd + n_tup_del 
        FROM pg_stat_user_tables 
        WHERE relname = 'nesma_vector_stores'
    ) > 1000000;
END;
$$ LANGUAGE plpgsql;

-- 授予必要的权限（根据实际用户调整）
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO nesma_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO nesma_user;
-- GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO nesma_user;

-- 完成初始化
SELECT 'pgvector extension initialization completed' as status;
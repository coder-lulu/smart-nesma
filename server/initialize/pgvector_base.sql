 -- PostgreSQL pgvector扩展基础初始化脚本
-- 用于NESMA智能文档生成系统的向量存储功能
-- 注意：此文件在数据表创建之前执行

-- 创建pgvector扩展
CREATE EXTENSION IF NOT EXISTS vector;

-- 创建向量存储相关的自定义类型和函数
CREATE OR REPLACE FUNCTION cosine_similarity(a vector, b vector) 
RETURNS float8 AS $$
BEGIN
    RETURN 1 - (a <=> b);
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

-- 创建触发器函数用于自动更新向量索引
CREATE OR REPLACE FUNCTION update_vector_embedding_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建向量数据清理函数（通用版本，不依赖特定表）
CREATE OR REPLACE FUNCTION cleanup_expired_vectors()
RETURNS int AS $$
DECLARE
    deleted_count int := 0;
    temp_count int;
BEGIN
    -- 删除过期的向量数据（如果表存在）
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'nesma_vector_stores') THEN
        EXECUTE 'DELETE FROM nesma_vector_stores WHERE expires_at IS NOT NULL AND expires_at < NOW()';
        GET DIAGNOSTICS temp_count = ROW_COUNT;
        deleted_count := deleted_count + temp_count;
    END IF;
    
    -- 删除过期的会话记忆（如果表存在）
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'nesma_session_memories') THEN
        EXECUTE 'DELETE FROM nesma_session_memories WHERE expires_at < NOW()';
        GET DIAGNOSTICS temp_count = ROW_COUNT;
        deleted_count := deleted_count + temp_count;
    END IF;
    
    -- 删除过期的Agent消息（如果表存在）
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'nesma_agent_messages') THEN
        EXECUTE 'DELETE FROM nesma_agent_messages WHERE expires_at IS NOT NULL AND expires_at < NOW()';
        GET DIAGNOSTICS temp_count = ROW_COUNT;
        deleted_count := deleted_count + temp_count;
    END IF;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 创建向量搜索性能优化建议函数
CREATE OR REPLACE FUNCTION vector_search_performance_hints()
RETURNS TABLE (
    suggestion text,
    priority int,
    description text
) AS $$
BEGIN
    -- 检查是否需要创建HNSW索引
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'nesma_vector_stores') THEN
        IF NOT EXISTS (
            SELECT 1 FROM pg_indexes 
            WHERE tablename = 'nesma_vector_stores' 
            AND indexname LIKE '%hnsw%'
        ) THEN
            RETURN QUERY
            SELECT 
                'Create HNSW index' as suggestion,
                1 as priority,
                'Create HNSW index for better vector search performance' as description;
        END IF;
        
        -- 检查表统计信息是否需要更新
        IF (
            SELECT COALESCE(last_analyze, '1970-01-01'::timestamp)
            FROM pg_stat_user_tables 
            WHERE relname = 'nesma_vector_stores'
        ) < NOW() - INTERVAL '1 day' THEN
            RETURN QUERY
            SELECT 
                'Update table statistics' as suggestion,
                2 as priority,
                'Run ANALYZE on vector tables to update statistics' as description;
        END IF;
        
        -- 检查是否需要分区
        IF (
            SELECT COALESCE(n_tup_ins + n_tup_upd + n_tup_del, 0)
            FROM pg_stat_user_tables 
            WHERE relname = 'nesma_vector_stores'
        ) > 1000000 THEN
            RETURN QUERY
            SELECT 
                'Consider partitioning' as suggestion,
                3 as priority,
                'Consider partitioning large vector tables by date or content type' as description;
        END IF;
    ELSE
        RETURN QUERY
        SELECT 
            'Tables not created yet' as suggestion,
            0 as priority,
            'Vector storage tables have not been created yet' as description;
    END IF;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- 完成基础初始化
SELECT 'pgvector base extension initialization completed' as status;
--
-- PostgreSQL database dump
--

\restrict mxvsBJZAJJ3og5W8Lt2GPbCiDFFtBZ1JSJz8QTZrbKfBxQJdIPyZgy4cZSTXXyV

-- Dumped from database version 17.11 (Debian 17.11-1.pgdg13+2)
-- Dumped by pg_dump version 17.11 (Debian 17.11-1.pgdg13+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: cleanup_expired_vectors(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.cleanup_expired_vectors() RETURNS integer
    LANGUAGE plpgsql
    AS $$
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
$$;


--
-- Name: cosine_similarity(public.vector, public.vector); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.cosine_similarity(a public.vector, b public.vector) RETURNS double precision
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN 1 - (a <=> b);
END;
$$;


--
-- Name: hybrid_vector_search(text, public.vector, character varying, integer, integer, double precision, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.hybrid_vector_search(query_text text, query_vector public.vector, content_type_filter character varying DEFAULT NULL::character varying, project_id_filter integer DEFAULT NULL::integer, user_id_filter integer DEFAULT NULL::integer, similarity_threshold double precision DEFAULT 0.7, max_results integer DEFAULT 20) RETURNS TABLE(id integer, vector_id character varying, content text, content_type character varying, similarity double precision, search_rank double precision)
    LANGUAGE plpgsql
    AS $$
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
$$;


--
-- Name: update_vector_embedding_timestamp(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_vector_embedding_timestamp() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: validate_vector_dimension(public.vector, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.validate_vector_dimension(vector_data public.vector, expected_dimension integer) RETURNS boolean
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN vector_dims(vector_data) = expected_dimension;
END;
$$;


--
-- Name: vector_search_performance_hints(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.vector_search_performance_hints() RETURNS TABLE(suggestion text, priority integer, description text)
    LANGUAGE plpgsql
    AS $$
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
$$;


--
-- Name: vector_similarity_search(public.vector, double precision, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.vector_similarity_search(query_vector public.vector, similarity_threshold double precision DEFAULT 0.7, max_results integer DEFAULT 20) RETURNS TABLE(id integer, vector_id character varying, content text, similarity double precision)
    LANGUAGE plpgsql
    AS $$
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
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: nesma_agent_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_agent_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    task_id character varying(100) NOT NULL,
    agent_id character varying(100) NOT NULL,
    task_type character varying(50) NOT NULL,
    priority bigint DEFAULT 1,
    status character varying(50) DEFAULT 'pending'::character varying,
    input_data jsonb,
    output_data jsonb,
    error_message text,
    session_id character varying(255),
    user_id bigint,
    project_id bigint,
    parent_task_id character varying(100),
    retry_count bigint DEFAULT 0,
    max_retries bigint DEFAULT 3,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    processing_time bigint,
    metadata jsonb
);


--
-- Name: COLUMN nesma_agent_tasks.task_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.task_id IS '任务唯一标识';


--
-- Name: COLUMN nesma_agent_tasks.agent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.agent_id IS '负责的Agent ID';


--
-- Name: COLUMN nesma_agent_tasks.task_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.task_type IS '任务类型';


--
-- Name: COLUMN nesma_agent_tasks.priority; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.priority IS '任务优先级:1-低,2-中,3-高,4-紧急';


--
-- Name: COLUMN nesma_agent_tasks.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.status IS '任务状态:pending,processing,completed,failed,cancelled';


--
-- Name: COLUMN nesma_agent_tasks.input_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.input_data IS '任务输入数据';


--
-- Name: COLUMN nesma_agent_tasks.output_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.output_data IS '任务输出数据';


--
-- Name: COLUMN nesma_agent_tasks.error_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.error_message IS '错误信息';


--
-- Name: COLUMN nesma_agent_tasks.session_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.session_id IS '会话ID';


--
-- Name: COLUMN nesma_agent_tasks.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.user_id IS '用户ID';


--
-- Name: COLUMN nesma_agent_tasks.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.project_id IS '项目ID';


--
-- Name: COLUMN nesma_agent_tasks.parent_task_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.parent_task_id IS '父任务ID';


--
-- Name: COLUMN nesma_agent_tasks.retry_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.retry_count IS '重试次数';


--
-- Name: COLUMN nesma_agent_tasks.max_retries; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.max_retries IS '最大重试次数';


--
-- Name: COLUMN nesma_agent_tasks.start_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.start_time IS '开始时间';


--
-- Name: COLUMN nesma_agent_tasks.end_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.end_time IS '结束时间';


--
-- Name: COLUMN nesma_agent_tasks.processing_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.processing_time IS '处理时间(毫秒)';


--
-- Name: COLUMN nesma_agent_tasks.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_tasks.metadata IS '任务元数据';


--
-- Name: agent_task_stats; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.agent_task_stats AS
 SELECT agent_id,
    task_type,
    status,
    count(*) AS task_count,
    avg(processing_time) AS avg_processing_time,
    min(processing_time) AS min_processing_time,
    max(processing_time) AS max_processing_time,
    count(
        CASE
            WHEN ((status)::text = 'completed'::text) THEN 1
            ELSE NULL::integer
        END) AS completed_count,
    count(
        CASE
            WHEN ((status)::text = 'failed'::text) THEN 1
            ELSE NULL::integer
        END) AS failed_count
   FROM public.nesma_agent_tasks
  GROUP BY agent_id, task_type, status;


--
-- Name: ai_business_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_business_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    analysis_id bigint NOT NULL,
    business_value numeric,
    roi_estimation numeric,
    strategic_alignment numeric,
    user_experience_score numeric,
    usability_analysis text,
    accessibility_analysis text,
    process_efficiency numeric,
    process_automation numeric,
    business_logic_complexity numeric,
    market_fit numeric,
    competitive_advantage numeric,
    innovation_level numeric,
    business_risk text,
    regulatory_compliance numeric,
    ai_business_assessment text,
    business_recommendations text,
    confidence_score numeric
);


--
-- Name: COLUMN ai_business_analyses.analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.analysis_id IS '分析ID';


--
-- Name: COLUMN ai_business_analyses.business_value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.business_value IS '业务价值评分';


--
-- Name: COLUMN ai_business_analyses.roi_estimation; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.roi_estimation IS '投资回报率估算';


--
-- Name: COLUMN ai_business_analyses.strategic_alignment; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.strategic_alignment IS '战略一致性';


--
-- Name: COLUMN ai_business_analyses.user_experience_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.user_experience_score IS '用户体验评分';


--
-- Name: COLUMN ai_business_analyses.usability_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.usability_analysis IS '可用性分析JSON';


--
-- Name: COLUMN ai_business_analyses.accessibility_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.accessibility_analysis IS '可访问性分析JSON';


--
-- Name: COLUMN ai_business_analyses.process_efficiency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.process_efficiency IS '流程效率';


--
-- Name: COLUMN ai_business_analyses.process_automation; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.process_automation IS '流程自动化程度';


--
-- Name: COLUMN ai_business_analyses.business_logic_complexity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.business_logic_complexity IS '业务逻辑复杂度';


--
-- Name: COLUMN ai_business_analyses.market_fit; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.market_fit IS '市场适配度';


--
-- Name: COLUMN ai_business_analyses.competitive_advantage; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.competitive_advantage IS '竞争优势';


--
-- Name: COLUMN ai_business_analyses.innovation_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.innovation_level IS '创新水平';


--
-- Name: COLUMN ai_business_analyses.business_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.business_risk IS '业务风险分析JSON';


--
-- Name: COLUMN ai_business_analyses.regulatory_compliance; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.regulatory_compliance IS '法规合规性';


--
-- Name: COLUMN ai_business_analyses.ai_business_assessment; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.ai_business_assessment IS 'AI业务评估';


--
-- Name: COLUMN ai_business_analyses.business_recommendations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.business_recommendations IS '业务建议';


--
-- Name: COLUMN ai_business_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_business_analyses.confidence_score IS '置信度分数';


--
-- Name: ai_business_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_business_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_business_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_business_analyses_id_seq OWNED BY public.ai_business_analyses.id;


--
-- Name: ai_compliance_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_compliance_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    analysis_id bigint NOT NULL,
    compliance_standard character varying(100),
    standard_version character varying(50),
    overall_compliance numeric,
    compliance_level character varying(50),
    compliance_items text,
    non_compliance_items text,
    partial_compliance_items text,
    compliance_gaps text,
    remediation_plan text,
    ai_compliance_assessment text,
    compliance_recommendations text,
    confidence_score numeric
);


--
-- Name: COLUMN ai_compliance_analyses.analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.analysis_id IS '分析ID';


--
-- Name: COLUMN ai_compliance_analyses.compliance_standard; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.compliance_standard IS '合规标准';


--
-- Name: COLUMN ai_compliance_analyses.standard_version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.standard_version IS '标准版本';


--
-- Name: COLUMN ai_compliance_analyses.overall_compliance; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.overall_compliance IS '总体合规性';


--
-- Name: COLUMN ai_compliance_analyses.compliance_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.compliance_level IS '合规等级';


--
-- Name: COLUMN ai_compliance_analyses.compliance_items; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.compliance_items IS '合规项目JSON';


--
-- Name: COLUMN ai_compliance_analyses.non_compliance_items; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.non_compliance_items IS '不合规项目JSON';


--
-- Name: COLUMN ai_compliance_analyses.partial_compliance_items; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.partial_compliance_items IS '部分合规项目JSON';


--
-- Name: COLUMN ai_compliance_analyses.compliance_gaps; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.compliance_gaps IS '合规缺口JSON';


--
-- Name: COLUMN ai_compliance_analyses.remediation_plan; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.remediation_plan IS '整改计划JSON';


--
-- Name: COLUMN ai_compliance_analyses.ai_compliance_assessment; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.ai_compliance_assessment IS 'AI合规评估';


--
-- Name: COLUMN ai_compliance_analyses.compliance_recommendations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.compliance_recommendations IS '合规建议';


--
-- Name: COLUMN ai_compliance_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_compliance_analyses.confidence_score IS '置信度分数';


--
-- Name: ai_compliance_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_compliance_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_compliance_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_compliance_analyses_id_seq OWNED BY public.ai_compliance_analyses.id;


--
-- Name: ai_functional_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_functional_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    analysis_id bigint NOT NULL,
    total_function_points numeric,
    data_function_points numeric,
    transactional_function_points numeric,
    ilf_count bigint,
    eif_count bigint,
    ei_count bigint,
    eo_count bigint,
    eq_count bigint,
    low_complexity_count bigint,
    medium_complexity_count bigint,
    high_complexity_count bigint,
    functional_coverage numeric,
    functional_completeness numeric,
    functional_consistency numeric,
    function_type_distribution text,
    complexity_analysis text,
    functional_gaps text,
    ai_assessment text,
    confidence_score numeric,
    quality_indicators text
);


--
-- Name: COLUMN ai_functional_analyses.analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.analysis_id IS '分析ID';


--
-- Name: COLUMN ai_functional_analyses.total_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.total_function_points IS '总功能点数';


--
-- Name: COLUMN ai_functional_analyses.data_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.data_function_points IS '数据功能点数';


--
-- Name: COLUMN ai_functional_analyses.transactional_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.transactional_function_points IS '事务功能点数';


--
-- Name: COLUMN ai_functional_analyses.ilf_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.ilf_count IS '内部逻辑文件数量';


--
-- Name: COLUMN ai_functional_analyses.eif_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.eif_count IS '外部接口文件数量';


--
-- Name: COLUMN ai_functional_analyses.ei_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.ei_count IS '外部输入数量';


--
-- Name: COLUMN ai_functional_analyses.eo_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.eo_count IS '外部输出数量';


--
-- Name: COLUMN ai_functional_analyses.eq_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.eq_count IS '外部查询数量';


--
-- Name: COLUMN ai_functional_analyses.low_complexity_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.low_complexity_count IS '低复杂度功能数';


--
-- Name: COLUMN ai_functional_analyses.medium_complexity_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.medium_complexity_count IS '中复杂度功能数';


--
-- Name: COLUMN ai_functional_analyses.high_complexity_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.high_complexity_count IS '高复杂度功能数';


--
-- Name: COLUMN ai_functional_analyses.functional_coverage; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.functional_coverage IS '功能覆盖度';


--
-- Name: COLUMN ai_functional_analyses.functional_completeness; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.functional_completeness IS '功能完整性';


--
-- Name: COLUMN ai_functional_analyses.functional_consistency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.functional_consistency IS '功能一致性';


--
-- Name: COLUMN ai_functional_analyses.function_type_distribution; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.function_type_distribution IS '功能类型分布分析JSON';


--
-- Name: COLUMN ai_functional_analyses.complexity_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.complexity_analysis IS '复杂度分析JSON';


--
-- Name: COLUMN ai_functional_analyses.functional_gaps; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.functional_gaps IS '功能缺口分析JSON';


--
-- Name: COLUMN ai_functional_analyses.ai_assessment; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.ai_assessment IS 'AI评估结果';


--
-- Name: COLUMN ai_functional_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.confidence_score IS '置信度分数';


--
-- Name: COLUMN ai_functional_analyses.quality_indicators; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_functional_analyses.quality_indicators IS '质量指标JSON';


--
-- Name: ai_functional_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_functional_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_functional_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_functional_analyses_id_seq OWNED BY public.ai_functional_analyses.id;


--
-- Name: ai_project_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_project_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    evaluation_id bigint NOT NULL,
    analysis_version character varying(50),
    analysis_type character varying(50),
    status character varying(50) DEFAULT 'processing'::character varying,
    progress numeric,
    start_time timestamp with time zone,
    completion_time timestamp with time zone,
    total_requirements bigint,
    processed_requirements bigint,
    requirement_levels text,
    project_context text,
    ai_model character varying(100),
    max_tokens bigint,
    temperature numeric,
    batch_size bigint,
    overall_score numeric,
    complexity_level character varying(50),
    risk_level character varying(50),
    recommendation_level character varying(50)
);


--
-- Name: COLUMN ai_project_analyses.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.project_id IS '项目ID';


--
-- Name: COLUMN ai_project_analyses.evaluation_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.evaluation_id IS '评估ID';


--
-- Name: COLUMN ai_project_analyses.analysis_version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.analysis_version IS '分析版本';


--
-- Name: COLUMN ai_project_analyses.analysis_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.analysis_type IS '分析类型:comprehensive,focused,incremental';


--
-- Name: COLUMN ai_project_analyses.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.status IS '分析状态:processing,completed,failed';


--
-- Name: COLUMN ai_project_analyses.progress; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.progress IS '分析进度0-100';


--
-- Name: COLUMN ai_project_analyses.start_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.start_time IS '分析开始时间';


--
-- Name: COLUMN ai_project_analyses.completion_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.completion_time IS '分析完成时间';


--
-- Name: COLUMN ai_project_analyses.total_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.total_requirements IS '总需求数';


--
-- Name: COLUMN ai_project_analyses.processed_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.processed_requirements IS '已处理需求数';


--
-- Name: COLUMN ai_project_analyses.requirement_levels; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.requirement_levels IS '需求层级分布JSON';


--
-- Name: COLUMN ai_project_analyses.project_context; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.project_context IS '项目上下文信息';


--
-- Name: COLUMN ai_project_analyses.ai_model; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.ai_model IS '使用的AI模型';


--
-- Name: COLUMN ai_project_analyses.max_tokens; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.max_tokens IS '最大Token数';


--
-- Name: COLUMN ai_project_analyses.temperature; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.temperature IS 'AI温度参数';


--
-- Name: COLUMN ai_project_analyses.batch_size; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.batch_size IS '分批处理大小';


--
-- Name: COLUMN ai_project_analyses.overall_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.overall_score IS '总体质量得分';


--
-- Name: COLUMN ai_project_analyses.complexity_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.complexity_level IS '复杂度等级';


--
-- Name: COLUMN ai_project_analyses.risk_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.risk_level IS '风险等级';


--
-- Name: COLUMN ai_project_analyses.recommendation_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_project_analyses.recommendation_level IS '建议等级';


--
-- Name: ai_project_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_project_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_project_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_project_analyses_id_seq OWNED BY public.ai_project_analyses.id;


--
-- Name: ai_quality_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_quality_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    analysis_id bigint NOT NULL,
    overall_quality numeric,
    requirement_quality numeric,
    design_quality numeric,
    correctness numeric,
    completeness numeric,
    consistency numeric,
    clarity numeric,
    traceability numeric,
    maintainability numeric,
    modularity numeric,
    reusability numeric,
    testability numeric,
    critical_issues text,
    quality_gaps text,
    improvement_areas text,
    ai_quality_assessment text,
    quality_recommendations text,
    confidence_score numeric
);


--
-- Name: COLUMN ai_quality_analyses.analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.analysis_id IS '分析ID';


--
-- Name: COLUMN ai_quality_analyses.overall_quality; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.overall_quality IS '总体质量';


--
-- Name: COLUMN ai_quality_analyses.requirement_quality; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.requirement_quality IS '需求质量';


--
-- Name: COLUMN ai_quality_analyses.design_quality; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.design_quality IS '设计质量';


--
-- Name: COLUMN ai_quality_analyses.correctness; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.correctness IS '正确性';


--
-- Name: COLUMN ai_quality_analyses.completeness; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.completeness IS '完整性';


--
-- Name: COLUMN ai_quality_analyses.consistency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.consistency IS '一致性';


--
-- Name: COLUMN ai_quality_analyses.clarity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.clarity IS '清晰性';


--
-- Name: COLUMN ai_quality_analyses.traceability; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.traceability IS '可追溯性';


--
-- Name: COLUMN ai_quality_analyses.maintainability; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.maintainability IS '可维护性';


--
-- Name: COLUMN ai_quality_analyses.modularity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.modularity IS '模块化程度';


--
-- Name: COLUMN ai_quality_analyses.reusability; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.reusability IS '可重用性';


--
-- Name: COLUMN ai_quality_analyses.testability; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.testability IS '可测试性';


--
-- Name: COLUMN ai_quality_analyses.critical_issues; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.critical_issues IS '关键问题JSON';


--
-- Name: COLUMN ai_quality_analyses.quality_gaps; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.quality_gaps IS '质量缺口JSON';


--
-- Name: COLUMN ai_quality_analyses.improvement_areas; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.improvement_areas IS '改进领域JSON';


--
-- Name: COLUMN ai_quality_analyses.ai_quality_assessment; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.ai_quality_assessment IS 'AI质量评估';


--
-- Name: COLUMN ai_quality_analyses.quality_recommendations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.quality_recommendations IS '质量建议';


--
-- Name: COLUMN ai_quality_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_quality_analyses.confidence_score IS '置信度分数';


--
-- Name: ai_quality_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_quality_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_quality_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_quality_analyses_id_seq OWNED BY public.ai_quality_analyses.id;


--
-- Name: ai_recommendation_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_recommendation_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    analysis_id bigint NOT NULL,
    recommendation_type character varying(100),
    priority character varying(50),
    impact_level character varying(50),
    title character varying(200),
    description text,
    rationale text,
    implementation_steps text,
    estimated_effort numeric,
    expected_benefit text,
    dependencies text,
    constraints text,
    risks text,
    ai_generated_recommendation text,
    confidence_score numeric,
    urgency character varying(50)
);


--
-- Name: COLUMN ai_recommendation_analyses.analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.analysis_id IS '分析ID';


--
-- Name: COLUMN ai_recommendation_analyses.recommendation_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.recommendation_type IS '建议类型';


--
-- Name: COLUMN ai_recommendation_analyses.priority; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.priority IS '优先级';


--
-- Name: COLUMN ai_recommendation_analyses.impact_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.impact_level IS '影响程度';


--
-- Name: COLUMN ai_recommendation_analyses.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.title IS '建议标题';


--
-- Name: COLUMN ai_recommendation_analyses.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.description IS '建议描述';


--
-- Name: COLUMN ai_recommendation_analyses.rationale; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.rationale IS '建议理由';


--
-- Name: COLUMN ai_recommendation_analyses.implementation_steps; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.implementation_steps IS '实施步骤JSON';


--
-- Name: COLUMN ai_recommendation_analyses.estimated_effort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.estimated_effort IS '估算工作量';


--
-- Name: COLUMN ai_recommendation_analyses.expected_benefit; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.expected_benefit IS '预期收益';


--
-- Name: COLUMN ai_recommendation_analyses.dependencies; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.dependencies IS '依赖关系JSON';


--
-- Name: COLUMN ai_recommendation_analyses.constraints; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.constraints IS '约束条件JSON';


--
-- Name: COLUMN ai_recommendation_analyses.risks; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.risks IS '相关风险JSON';


--
-- Name: COLUMN ai_recommendation_analyses.ai_generated_recommendation; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.ai_generated_recommendation IS 'AI生成的建议';


--
-- Name: COLUMN ai_recommendation_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.confidence_score IS '置信度分数';


--
-- Name: COLUMN ai_recommendation_analyses.urgency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_recommendation_analyses.urgency IS '紧急程度';


--
-- Name: ai_recommendation_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_recommendation_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_recommendation_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_recommendation_analyses_id_seq OWNED BY public.ai_recommendation_analyses.id;


--
-- Name: ai_risk_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_risk_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    analysis_id bigint NOT NULL,
    overall_risk numeric,
    risk_level character varying(50),
    technical_risk numeric,
    implementation_risk numeric,
    integration_risk numeric,
    schedule_risk numeric,
    budget_risk numeric,
    resource_risk numeric,
    business_risk numeric,
    market_risk numeric,
    compliance_risk numeric,
    identified_risks text,
    risk_mitigation_plans text,
    contingency_plans text,
    ai_risk_assessment text,
    risk_recommendations text,
    confidence_score numeric
);


--
-- Name: COLUMN ai_risk_analyses.analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.analysis_id IS '分析ID';


--
-- Name: COLUMN ai_risk_analyses.overall_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.overall_risk IS '总体风险';


--
-- Name: COLUMN ai_risk_analyses.risk_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.risk_level IS '风险等级';


--
-- Name: COLUMN ai_risk_analyses.technical_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.technical_risk IS '技术风险';


--
-- Name: COLUMN ai_risk_analyses.implementation_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.implementation_risk IS '实现风险';


--
-- Name: COLUMN ai_risk_analyses.integration_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.integration_risk IS '集成风险';


--
-- Name: COLUMN ai_risk_analyses.schedule_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.schedule_risk IS '进度风险';


--
-- Name: COLUMN ai_risk_analyses.budget_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.budget_risk IS '预算风险';


--
-- Name: COLUMN ai_risk_analyses.resource_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.resource_risk IS '资源风险';


--
-- Name: COLUMN ai_risk_analyses.business_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.business_risk IS '业务风险';


--
-- Name: COLUMN ai_risk_analyses.market_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.market_risk IS '市场风险';


--
-- Name: COLUMN ai_risk_analyses.compliance_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.compliance_risk IS '合规风险';


--
-- Name: COLUMN ai_risk_analyses.identified_risks; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.identified_risks IS '识别的风险JSON';


--
-- Name: COLUMN ai_risk_analyses.risk_mitigation_plans; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.risk_mitigation_plans IS '风险缓解计划JSON';


--
-- Name: COLUMN ai_risk_analyses.contingency_plans; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.contingency_plans IS '应急计划JSON';


--
-- Name: COLUMN ai_risk_analyses.ai_risk_assessment; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.ai_risk_assessment IS 'AI风险评估';


--
-- Name: COLUMN ai_risk_analyses.risk_recommendations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.risk_recommendations IS '风险建议';


--
-- Name: COLUMN ai_risk_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_risk_analyses.confidence_score IS '置信度分数';


--
-- Name: ai_risk_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_risk_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_risk_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_risk_analyses_id_seq OWNED BY public.ai_risk_analyses.id;


--
-- Name: ai_service_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_service_metrics (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    service_name character varying(50) NOT NULL,
    metric_type character varying(50) NOT NULL,
    metric_value numeric NOT NULL,
    request_count bigint DEFAULT 0,
    error_count bigint DEFAULT 0,
    total_tokens bigint DEFAULT 0,
    prompt_tokens bigint DEFAULT 0,
    completion_tokens bigint DEFAULT 0,
    average_latency numeric DEFAULT 0,
    p95_latency numeric DEFAULT 0,
    p99_latency numeric DEFAULT 0,
    error_messages jsonb,
    metadata jsonb,
    recorded_at timestamp with time zone,
    user_id bigint NOT NULL,
    model_name character varying(100) NOT NULL,
    request_type character varying(50) NOT NULL,
    status character varying(20) NOT NULL,
    response_time bigint NOT NULL,
    token_usage bigint DEFAULT 0,
    error_message text,
    request_payload json
);


--
-- Name: ai_service_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_service_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_service_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_service_metrics_id_seq OWNED BY public.ai_service_metrics.id;


--
-- Name: ai_technical_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ai_technical_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    analysis_id bigint NOT NULL,
    architecture_type character varying(100),
    technology_stack text,
    integration_complexity character varying(50),
    performance_requirements text,
    scalability_analysis text,
    reliability_analysis text,
    technical_debt text,
    security_considerations text,
    maintenance_complexity numeric,
    development_effort numeric,
    testing_effort numeric,
    deployment_complexity numeric,
    technical_feasibility numeric,
    implementation_risk numeric,
    technical_innovation numeric,
    ai_technical_assessment text,
    technical_recommendations text,
    confidence_score numeric
);


--
-- Name: COLUMN ai_technical_analyses.analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.analysis_id IS '分析ID';


--
-- Name: COLUMN ai_technical_analyses.architecture_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.architecture_type IS '架构类型';


--
-- Name: COLUMN ai_technical_analyses.technology_stack; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.technology_stack IS '技术栈JSON';


--
-- Name: COLUMN ai_technical_analyses.integration_complexity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.integration_complexity IS '集成复杂度';


--
-- Name: COLUMN ai_technical_analyses.performance_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.performance_requirements IS '性能需求JSON';


--
-- Name: COLUMN ai_technical_analyses.scalability_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.scalability_analysis IS '可扩展性分析JSON';


--
-- Name: COLUMN ai_technical_analyses.reliability_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.reliability_analysis IS '可靠性分析JSON';


--
-- Name: COLUMN ai_technical_analyses.technical_debt; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.technical_debt IS '技术债务分析JSON';


--
-- Name: COLUMN ai_technical_analyses.security_considerations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.security_considerations IS '安全考虑JSON';


--
-- Name: COLUMN ai_technical_analyses.maintenance_complexity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.maintenance_complexity IS '维护复杂度';


--
-- Name: COLUMN ai_technical_analyses.development_effort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.development_effort IS '开发工作量估算';


--
-- Name: COLUMN ai_technical_analyses.testing_effort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.testing_effort IS '测试工作量估算';


--
-- Name: COLUMN ai_technical_analyses.deployment_complexity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.deployment_complexity IS '部署复杂度';


--
-- Name: COLUMN ai_technical_analyses.technical_feasibility; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.technical_feasibility IS '技术可行性';


--
-- Name: COLUMN ai_technical_analyses.implementation_risk; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.implementation_risk IS '实现风险';


--
-- Name: COLUMN ai_technical_analyses.technical_innovation; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.technical_innovation IS '技术创新度';


--
-- Name: COLUMN ai_technical_analyses.ai_technical_assessment; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.ai_technical_assessment IS 'AI技术评估';


--
-- Name: COLUMN ai_technical_analyses.technical_recommendations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.technical_recommendations IS '技术建议';


--
-- Name: COLUMN ai_technical_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.ai_technical_analyses.confidence_score IS '置信度分数';


--
-- Name: ai_technical_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ai_technical_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ai_technical_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ai_technical_analyses_id_seq OWNED BY public.ai_technical_analyses.id;


--
-- Name: casbin_rule; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.casbin_rule (
    id bigint NOT NULL,
    ptype character varying(100),
    v0 character varying(100),
    v1 character varying(100),
    v2 character varying(100),
    v3 character varying(100),
    v4 character varying(100),
    v5 character varying(100)
);


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.casbin_rule_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.casbin_rule_id_seq OWNED BY public.casbin_rule.id;


--
-- Name: chat_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_messages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    session_id bigint NOT NULL,
    role character varying(20) NOT NULL,
    content text NOT NULL,
    metadata json,
    token_count bigint DEFAULT 0
);


--
-- Name: chat_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.chat_messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: chat_messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_messages_id_seq OWNED BY public.chat_messages.id;


--
-- Name: chat_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat_sessions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    project_id bigint,
    title character varying(255) NOT NULL,
    description text,
    model_name character varying(50) NOT NULL,
    is_active boolean DEFAULT true
);


--
-- Name: chat_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.chat_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: chat_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_sessions_id_seq OWNED BY public.chat_sessions.id;


--
-- Name: exa_attachment_category; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_attachment_category (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(255) DEFAULT NULL::character varying,
    pid bigint DEFAULT 0
);


--
-- Name: COLUMN exa_attachment_category.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_attachment_category.name IS '分类名称';


--
-- Name: COLUMN exa_attachment_category.pid; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_attachment_category.pid IS '父节点ID';


--
-- Name: exa_attachment_category_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_attachment_category_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_attachment_category_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_attachment_category_id_seq OWNED BY public.exa_attachment_category.id;


--
-- Name: exa_customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_customers (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    customer_name text,
    customer_phone_data text,
    sys_user_id bigint,
    sys_user_authority_id bigint
);


--
-- Name: COLUMN exa_customers.customer_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.customer_name IS '客户名';


--
-- Name: COLUMN exa_customers.customer_phone_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.customer_phone_data IS '客户手机号';


--
-- Name: COLUMN exa_customers.sys_user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.sys_user_id IS '管理ID';


--
-- Name: COLUMN exa_customers.sys_user_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_customers.sys_user_authority_id IS '管理角色ID';


--
-- Name: exa_customers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_customers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_customers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_customers_id_seq OWNED BY public.exa_customers.id;


--
-- Name: exa_file_chunks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_file_chunks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    exa_file_id bigint,
    file_chunk_number bigint,
    file_chunk_path text
);


--
-- Name: exa_file_chunks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_file_chunks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_file_chunks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_file_chunks_id_seq OWNED BY public.exa_file_chunks.id;


--
-- Name: exa_file_upload_and_downloads; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_file_upload_and_downloads (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    class_id bigint DEFAULT 0,
    url text,
    tag text,
    key text
);


--
-- Name: COLUMN exa_file_upload_and_downloads.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.name IS '文件名';


--
-- Name: COLUMN exa_file_upload_and_downloads.class_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.class_id IS '分类id';


--
-- Name: COLUMN exa_file_upload_and_downloads.url; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.url IS '文件地址';


--
-- Name: COLUMN exa_file_upload_and_downloads.tag; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.tag IS '文件标签';


--
-- Name: COLUMN exa_file_upload_and_downloads.key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.exa_file_upload_and_downloads.key IS '编号';


--
-- Name: exa_file_upload_and_downloads_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_file_upload_and_downloads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_file_upload_and_downloads_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_file_upload_and_downloads_id_seq OWNED BY public.exa_file_upload_and_downloads.id;


--
-- Name: exa_files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exa_files (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    file_name text,
    file_md5 text,
    file_path text,
    chunk_total bigint,
    is_finish boolean
);


--
-- Name: exa_files_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.exa_files_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: exa_files_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.exa_files_id_seq OWNED BY public.exa_files.id;


--
-- Name: gva_announcements_info; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gva_announcements_info (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text,
    content text,
    user_id bigint,
    attachments jsonb
);


--
-- Name: COLUMN gva_announcements_info.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.title IS '公告标题';


--
-- Name: COLUMN gva_announcements_info.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.content IS '公告内容';


--
-- Name: COLUMN gva_announcements_info.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.user_id IS '发布者';


--
-- Name: COLUMN gva_announcements_info.attachments; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.gva_announcements_info.attachments IS '相关附件';


--
-- Name: gva_announcements_info_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.gva_announcements_info_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: gva_announcements_info_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.gva_announcements_info_id_seq OWNED BY public.gva_announcements_info.id;


--
-- Name: jwt_blacklists; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jwt_blacklists (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    jwt text
);


--
-- Name: COLUMN jwt_blacklists.jwt; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.jwt_blacklists.jwt IS 'jwt';


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jwt_blacklists_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jwt_blacklists_id_seq OWNED BY public.jwt_blacklists.id;


--
-- Name: nesma_agent_interactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_agent_interactions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    session_id character varying(255) NOT NULL,
    source_agent character varying(100) NOT NULL,
    target_agent character varying(100) NOT NULL,
    message_type character varying(50) NOT NULL,
    message_content jsonb,
    response_content jsonb,
    processing_time_ms bigint,
    status character varying(50) DEFAULT 'success'::character varying,
    error_message text
);


--
-- Name: COLUMN nesma_agent_interactions.session_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.session_id IS '会话ID';


--
-- Name: COLUMN nesma_agent_interactions.source_agent; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.source_agent IS '源Agent';


--
-- Name: COLUMN nesma_agent_interactions.target_agent; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.target_agent IS '目标Agent';


--
-- Name: COLUMN nesma_agent_interactions.message_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.message_type IS '消息类型';


--
-- Name: COLUMN nesma_agent_interactions.message_content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.message_content IS '消息内容';


--
-- Name: COLUMN nesma_agent_interactions.response_content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.response_content IS '响应内容';


--
-- Name: COLUMN nesma_agent_interactions.processing_time_ms; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.processing_time_ms IS '处理时间(毫秒)';


--
-- Name: COLUMN nesma_agent_interactions.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.status IS '状态';


--
-- Name: COLUMN nesma_agent_interactions.error_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_interactions.error_message IS '错误信息';


--
-- Name: nesma_agent_interactions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_agent_interactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_agent_interactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_agent_interactions_id_seq OWNED BY public.nesma_agent_interactions.id;


--
-- Name: nesma_agent_messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_agent_messages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    message_id character varying(100) NOT NULL,
    session_id character varying(255) NOT NULL,
    source_agent_id character varying(100) NOT NULL,
    target_agent_id character varying(100),
    message_type character varying(50) NOT NULL,
    content_type character varying(50),
    content jsonb NOT NULL,
    status character varying(50) DEFAULT 'sent'::character varying,
    priority bigint DEFAULT 1,
    response_to_id character varying(100),
    expires_at timestamp with time zone,
    metadata jsonb,
    delivered_at timestamp with time zone,
    read_at timestamp with time zone
);


--
-- Name: COLUMN nesma_agent_messages.message_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.message_id IS '消息唯一标识';


--
-- Name: COLUMN nesma_agent_messages.session_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.session_id IS '会话ID';


--
-- Name: COLUMN nesma_agent_messages.source_agent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.source_agent_id IS '源Agent ID';


--
-- Name: COLUMN nesma_agent_messages.target_agent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.target_agent_id IS '目标Agent ID';


--
-- Name: COLUMN nesma_agent_messages.message_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.message_type IS '消息类型:REQUEST,RESPONSE,NOTIFICATION,BROADCAST';


--
-- Name: COLUMN nesma_agent_messages.content_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.content_type IS '内容类型:text,json,binary';


--
-- Name: COLUMN nesma_agent_messages.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.content IS '消息内容';


--
-- Name: COLUMN nesma_agent_messages.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.status IS '消息状态:sent,delivered,read,failed';


--
-- Name: COLUMN nesma_agent_messages.priority; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.priority IS '消息优先级';


--
-- Name: COLUMN nesma_agent_messages.response_to_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.response_to_id IS '响应的消息ID';


--
-- Name: COLUMN nesma_agent_messages.expires_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.expires_at IS '过期时间';


--
-- Name: COLUMN nesma_agent_messages.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.metadata IS '消息元数据';


--
-- Name: COLUMN nesma_agent_messages.delivered_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.delivered_at IS '送达时间';


--
-- Name: COLUMN nesma_agent_messages.read_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agent_messages.read_at IS '阅读时间';


--
-- Name: nesma_agent_messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_agent_messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_agent_messages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_agent_messages_id_seq OWNED BY public.nesma_agent_messages.id;


--
-- Name: nesma_agent_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_agent_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_agent_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_agent_tasks_id_seq OWNED BY public.nesma_agent_tasks.id;


--
-- Name: nesma_agents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_agents (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    agent_id character varying(100) NOT NULL,
    name character varying(200) NOT NULL,
    description text,
    agent_type character varying(50) NOT NULL,
    capabilities jsonb,
    status character varying(50) DEFAULT 'active'::character varying,
    version character varying(50) DEFAULT '1.0.0'::character varying,
    configuration jsonb,
    max_concurrency bigint DEFAULT 3,
    last_heartbeat timestamp with time zone,
    processing_count bigint DEFAULT 0,
    total_processed bigint DEFAULT 0,
    average_response_time bigint DEFAULT 0
);


--
-- Name: COLUMN nesma_agents.agent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.agent_id IS 'Agent唯一标识';


--
-- Name: COLUMN nesma_agents.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.name IS 'Agent名称';


--
-- Name: COLUMN nesma_agents.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.description IS 'Agent描述';


--
-- Name: COLUMN nesma_agents.agent_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.agent_type IS 'Agent类型:REQUIREMENT_ANALYSIS,NESMA_EVALUATION,KNOWLEDGE_RETRIEVAL,DOCUMENT_GENERATION,COORDINATION';


--
-- Name: COLUMN nesma_agents.capabilities; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.capabilities IS 'Agent能力配置';


--
-- Name: COLUMN nesma_agents.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.status IS 'Agent状态:active,inactive,busy,error';


--
-- Name: COLUMN nesma_agents.version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.version IS 'Agent版本';


--
-- Name: COLUMN nesma_agents.configuration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.configuration IS 'Agent配置参数';


--
-- Name: COLUMN nesma_agents.max_concurrency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.max_concurrency IS '最大并发处理数';


--
-- Name: COLUMN nesma_agents.last_heartbeat; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.last_heartbeat IS '最后心跳时间';


--
-- Name: COLUMN nesma_agents.processing_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.processing_count IS '当前处理中的任务数';


--
-- Name: COLUMN nesma_agents.total_processed; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.total_processed IS '总处理任务数';


--
-- Name: COLUMN nesma_agents.average_response_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_agents.average_response_time IS '平均响应时间(毫秒)';


--
-- Name: nesma_agents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_agents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_agents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_agents_id_seq OWNED BY public.nesma_agents.id;


--
-- Name: nesma_analysis_histories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_analysis_histories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    analysis_type character varying(50) NOT NULL,
    analysis_date timestamp with time zone NOT NULL,
    quality_score numeric,
    function_points numeric,
    requirement_count bigint,
    complexity_index numeric,
    quality_delta numeric,
    fp_delta numeric,
    requirement_delta bigint,
    project_analysis_id bigint,
    evaluation_id bigint
);


--
-- Name: COLUMN nesma_analysis_histories.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.project_id IS '项目ID';


--
-- Name: COLUMN nesma_analysis_histories.analysis_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.analysis_type IS '分析类型';


--
-- Name: COLUMN nesma_analysis_histories.analysis_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.analysis_date IS '分析日期';


--
-- Name: COLUMN nesma_analysis_histories.quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.quality_score IS '质量评分';


--
-- Name: COLUMN nesma_analysis_histories.function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.function_points IS '功能点数';


--
-- Name: COLUMN nesma_analysis_histories.requirement_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.requirement_count IS '需求数量';


--
-- Name: COLUMN nesma_analysis_histories.complexity_index; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.complexity_index IS '复杂度指数';


--
-- Name: COLUMN nesma_analysis_histories.quality_delta; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.quality_delta IS '质量变化';


--
-- Name: COLUMN nesma_analysis_histories.fp_delta; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.fp_delta IS '功能点变化';


--
-- Name: COLUMN nesma_analysis_histories.requirement_delta; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.requirement_delta IS '需求数变化';


--
-- Name: COLUMN nesma_analysis_histories.project_analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.project_analysis_id IS '项目分析ID';


--
-- Name: COLUMN nesma_analysis_histories.evaluation_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_analysis_histories.evaluation_id IS '评估ID';


--
-- Name: nesma_analysis_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_analysis_histories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_analysis_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_analysis_histories_id_seq OWNED BY public.nesma_analysis_histories.id;


--
-- Name: nesma_case_studies; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_case_studies (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_name character varying(255) NOT NULL,
    domain character varying(100),
    requirements_summary text,
    nesma_results jsonb,
    lessons_learned text,
    embedding text,
    total_function_points numeric,
    project_duration bigint,
    team_size bigint,
    complexity character varying(50),
    success boolean DEFAULT true
);


--
-- Name: COLUMN nesma_case_studies.project_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.project_name IS '项目名称';


--
-- Name: COLUMN nesma_case_studies.domain; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.domain IS '项目领域';


--
-- Name: COLUMN nesma_case_studies.requirements_summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.requirements_summary IS '需求摘要';


--
-- Name: COLUMN nesma_case_studies.nesma_results; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.nesma_results IS 'NESMA评估结果';


--
-- Name: COLUMN nesma_case_studies.lessons_learned; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.lessons_learned IS '经验教训';


--
-- Name: COLUMN nesma_case_studies.embedding; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.embedding IS '向量表示(JSON格式)';


--
-- Name: COLUMN nesma_case_studies.total_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.total_function_points IS '总功能点数';


--
-- Name: COLUMN nesma_case_studies.project_duration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.project_duration IS '项目持续时间(天)';


--
-- Name: COLUMN nesma_case_studies.team_size; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.team_size IS '团队规模';


--
-- Name: COLUMN nesma_case_studies.complexity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.complexity IS '项目复杂度';


--
-- Name: COLUMN nesma_case_studies.success; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_case_studies.success IS '项目是否成功';


--
-- Name: nesma_case_studies_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_case_studies_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_case_studies_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_case_studies_id_seq OWNED BY public.nesma_case_studies.id;


--
-- Name: nesma_compact_memories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_compact_memories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    project_id bigint,
    summary text NOT NULL,
    interaction_count bigint DEFAULT 0,
    last_updated timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    importance numeric DEFAULT 1
);


--
-- Name: COLUMN nesma_compact_memories.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_compact_memories.user_id IS '用户ID';


--
-- Name: COLUMN nesma_compact_memories.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_compact_memories.project_id IS '项目ID';


--
-- Name: COLUMN nesma_compact_memories.summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_compact_memories.summary IS '记忆摘要';


--
-- Name: COLUMN nesma_compact_memories.interaction_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_compact_memories.interaction_count IS '交互次数';


--
-- Name: COLUMN nesma_compact_memories.last_updated; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_compact_memories.last_updated IS '最后更新时间';


--
-- Name: COLUMN nesma_compact_memories.importance; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_compact_memories.importance IS '重要性评分';


--
-- Name: nesma_compact_memories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_compact_memories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_compact_memories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_compact_memories_id_seq OWNED BY public.nesma_compact_memories.id;


--
-- Name: nesma_complexity_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_complexity_metrics (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    evaluation_id bigint NOT NULL,
    metric_type character varying(100) NOT NULL,
    metric_name character varying(200) NOT NULL,
    metric_value numeric NOT NULL,
    threshold_value numeric,
    score numeric,
    weight numeric DEFAULT 1,
    calculation_formula character varying(500),
    calculation_details jsonb,
    quality_indicator character varying(50),
    impact_level character varying(50)
);


--
-- Name: COLUMN nesma_complexity_metrics.evaluation_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.evaluation_id IS '评估ID';


--
-- Name: COLUMN nesma_complexity_metrics.metric_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.metric_type IS '指标类型';


--
-- Name: COLUMN nesma_complexity_metrics.metric_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.metric_name IS '指标名称';


--
-- Name: COLUMN nesma_complexity_metrics.metric_value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.metric_value IS '指标值';


--
-- Name: COLUMN nesma_complexity_metrics.threshold_value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.threshold_value IS '阈值';


--
-- Name: COLUMN nesma_complexity_metrics.score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.score IS '评分';


--
-- Name: COLUMN nesma_complexity_metrics.weight; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.weight IS '权重';


--
-- Name: COLUMN nesma_complexity_metrics.calculation_formula; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.calculation_formula IS '计算公式';


--
-- Name: COLUMN nesma_complexity_metrics.calculation_details; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.calculation_details IS '计算详情';


--
-- Name: COLUMN nesma_complexity_metrics.quality_indicator; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.quality_indicator IS '质量指示器:excellent,good,average,poor';


--
-- Name: COLUMN nesma_complexity_metrics.impact_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_complexity_metrics.impact_level IS '影响级别:low,medium,high,critical';


--
-- Name: nesma_complexity_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_complexity_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_complexity_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_complexity_metrics_id_seq OWNED BY public.nesma_complexity_metrics.id;


--
-- Name: nesma_doc_batches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_doc_batches (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    total_count bigint NOT NULL,
    success_count bigint DEFAULT 0,
    failed_count bigint DEFAULT 0,
    status character varying(50) DEFAULT 'pending'::character varying,
    progress bigint DEFAULT 0,
    config jsonb,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_by bigint NOT NULL
);


--
-- Name: COLUMN nesma_doc_batches.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.project_id IS '项目ID';


--
-- Name: COLUMN nesma_doc_batches.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.name IS '批次名称';


--
-- Name: COLUMN nesma_doc_batches.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.description IS '批次描述';


--
-- Name: COLUMN nesma_doc_batches.total_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.total_count IS '总文档数';


--
-- Name: COLUMN nesma_doc_batches.success_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.success_count IS '成功数量';


--
-- Name: COLUMN nesma_doc_batches.failed_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.failed_count IS '失败数量';


--
-- Name: COLUMN nesma_doc_batches.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.status IS '批次状态：pending/processing/completed/failed';


--
-- Name: COLUMN nesma_doc_batches.progress; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.progress IS '批次进度0-100';


--
-- Name: COLUMN nesma_doc_batches.config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.config IS '批次配置';


--
-- Name: COLUMN nesma_doc_batches.started_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.started_at IS '开始时间';


--
-- Name: COLUMN nesma_doc_batches.completed_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.completed_at IS '完成时间';


--
-- Name: COLUMN nesma_doc_batches.created_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_batches.created_by IS '创建人ID';


--
-- Name: nesma_doc_batches_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_doc_batches_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_doc_batches_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_doc_batches_id_seq OWNED BY public.nesma_doc_batches.id;


--
-- Name: nesma_doc_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_doc_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(255) NOT NULL,
    description text,
    type character varying(50) NOT NULL,
    format character varying(50) NOT NULL,
    category character varying(100),
    template_path character varying(500),
    preview_path character varying(500),
    config jsonb,
    variables jsonb,
    is_default boolean DEFAULT false,
    is_active boolean DEFAULT true,
    usage_count bigint DEFAULT 0,
    created_by bigint NOT NULL,
    content text
);


--
-- Name: COLUMN nesma_doc_templates.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.name IS '模板名称';


--
-- Name: COLUMN nesma_doc_templates.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.description IS '模板描述';


--
-- Name: COLUMN nesma_doc_templates.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.type IS '模板类型：word/excel/pdf';


--
-- Name: COLUMN nesma_doc_templates.format; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.format IS '模板格式：requirement_spec/nesma_report/business_summary';


--
-- Name: COLUMN nesma_doc_templates.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.category IS '模板分类';


--
-- Name: COLUMN nesma_doc_templates.template_path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.template_path IS '模板文件路径';


--
-- Name: COLUMN nesma_doc_templates.preview_path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.preview_path IS '预览图路径';


--
-- Name: COLUMN nesma_doc_templates.config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.config IS '模板配置';


--
-- Name: COLUMN nesma_doc_templates.variables; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.variables IS '模板变量';


--
-- Name: COLUMN nesma_doc_templates.is_default; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.is_default IS '是否默认模板';


--
-- Name: COLUMN nesma_doc_templates.is_active; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.is_active IS '是否激活';


--
-- Name: COLUMN nesma_doc_templates.usage_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.usage_count IS '使用次数';


--
-- Name: COLUMN nesma_doc_templates.created_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.created_by IS '创建人ID';


--
-- Name: COLUMN nesma_doc_templates.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_doc_templates.content IS '模板内容';


--
-- Name: nesma_doc_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_doc_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_doc_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_doc_templates_id_seq OWNED BY public.nesma_doc_templates.id;


--
-- Name: nesma_documents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_documents (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    template_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    type character varying(50) NOT NULL,
    format character varying(50) NOT NULL,
    file_path character varying(500),
    file_size bigint,
    status character varying(50) DEFAULT 'pending'::character varying,
    progress bigint DEFAULT 0,
    error_msg text,
    config jsonb,
    version character varying(50),
    generated_at timestamp with time zone,
    created_by bigint NOT NULL,
    batch_id bigint,
    cycle_id bigint,
    version_id bigint
);


--
-- Name: COLUMN nesma_documents.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.project_id IS '项目ID';


--
-- Name: COLUMN nesma_documents.template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.template_id IS '模板ID（支持内置模板如builtin_excel）';


--
-- Name: COLUMN nesma_documents.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.name IS '文档名称';


--
-- Name: COLUMN nesma_documents.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.description IS '文档描述';


--
-- Name: COLUMN nesma_documents.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.type IS '文档类型：word/excel/pdf';


--
-- Name: COLUMN nesma_documents.format; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.format IS '文档格式：requirement_spec/nesma_report/business_summary';


--
-- Name: COLUMN nesma_documents.file_path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.file_path IS '文件路径';


--
-- Name: COLUMN nesma_documents.file_size; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.file_size IS '文件大小';


--
-- Name: COLUMN nesma_documents.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.status IS '生成状态：pending/generating/completed/failed';


--
-- Name: COLUMN nesma_documents.progress; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.progress IS '生成进度0-100';


--
-- Name: COLUMN nesma_documents.error_msg; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.error_msg IS '错误信息';


--
-- Name: COLUMN nesma_documents.config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.config IS '生成配置';


--
-- Name: COLUMN nesma_documents.version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.version IS '文档版本';


--
-- Name: COLUMN nesma_documents.generated_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.generated_at IS '生成完成时间';


--
-- Name: COLUMN nesma_documents.created_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.created_by IS '创建人ID';


--
-- Name: COLUMN nesma_documents.batch_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.batch_id IS '批次ID（可选）';


--
-- Name: COLUMN nesma_documents.cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.cycle_id IS '项目周期ID';


--
-- Name: COLUMN nesma_documents.version_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_documents.version_id IS '需求版本ID';


--
-- Name: nesma_documents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_documents_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_documents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_documents_id_seq OWNED BY public.nesma_documents.id;


--
-- Name: nesma_embeddings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_embeddings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    embedding_id character varying(100) NOT NULL,
    content text NOT NULL,
    vector text NOT NULL,
    dimensions bigint NOT NULL,
    model character varying(100),
    token_count bigint,
    cost numeric,
    provider character varying(50),
    metadata jsonb,
    hash character varying(64),
    is_active boolean DEFAULT true
);


--
-- Name: COLUMN nesma_embeddings.embedding_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.embedding_id IS '嵌入向量唯一标识';


--
-- Name: COLUMN nesma_embeddings.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.content IS '原始内容';


--
-- Name: COLUMN nesma_embeddings.vector; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.vector IS '向量数据';


--
-- Name: COLUMN nesma_embeddings.dimensions; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.dimensions IS '向量维度';


--
-- Name: COLUMN nesma_embeddings.model; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.model IS '使用的模型';


--
-- Name: COLUMN nesma_embeddings.token_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.token_count IS 'token数量';


--
-- Name: COLUMN nesma_embeddings.cost; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.cost IS '成本';


--
-- Name: COLUMN nesma_embeddings.provider; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.provider IS '提供商';


--
-- Name: COLUMN nesma_embeddings.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.metadata IS '元数据';


--
-- Name: COLUMN nesma_embeddings.hash; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.hash IS '内容哈希';


--
-- Name: COLUMN nesma_embeddings.is_active; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_embeddings.is_active IS '是否激活';


--
-- Name: nesma_embeddings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_embeddings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_embeddings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_embeddings_id_seq OWNED BY public.nesma_embeddings.id;


--
-- Name: nesma_evaluation_factors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_evaluation_factors (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    evaluation_id bigint NOT NULL,
    factor_config text NOT NULL,
    nesma_version character varying(20) DEFAULT 'v2.2'::character varying
);


--
-- Name: COLUMN nesma_evaluation_factors.evaluation_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_factors.evaluation_id IS '评估ID';


--
-- Name: COLUMN nesma_evaluation_factors.factor_config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_factors.factor_config IS '因子配置JSON';


--
-- Name: COLUMN nesma_evaluation_factors.nesma_version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_factors.nesma_version IS 'NESMA版本';


--
-- Name: nesma_evaluation_factors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_evaluation_factors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_evaluation_factors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_evaluation_factors_id_seq OWNED BY public.nesma_evaluation_factors.id;


--
-- Name: nesma_evaluation_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_evaluation_history (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    current_eval_id bigint NOT NULL,
    previous_eval_id bigint NOT NULL,
    fp_difference numeric,
    fp_change_percent numeric,
    complexity_change numeric,
    change_category character varying(100),
    change_summary text,
    change_reasons jsonb,
    impact_analysis jsonb
);


--
-- Name: COLUMN nesma_evaluation_history.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.project_id IS '项目ID';


--
-- Name: COLUMN nesma_evaluation_history.current_eval_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.current_eval_id IS '当前评估ID';


--
-- Name: COLUMN nesma_evaluation_history.previous_eval_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.previous_eval_id IS '之前评估ID';


--
-- Name: COLUMN nesma_evaluation_history.fp_difference; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.fp_difference IS '功能点差异';


--
-- Name: COLUMN nesma_evaluation_history.fp_change_percent; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.fp_change_percent IS '功能点变化百分比';


--
-- Name: COLUMN nesma_evaluation_history.complexity_change; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.complexity_change IS '复杂度变化';


--
-- Name: COLUMN nesma_evaluation_history.change_category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.change_category IS '变化类别';


--
-- Name: COLUMN nesma_evaluation_history.change_summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.change_summary IS '变化总结';


--
-- Name: COLUMN nesma_evaluation_history.change_reasons; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.change_reasons IS '变化原因';


--
-- Name: COLUMN nesma_evaluation_history.impact_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluation_history.impact_analysis IS '影响分析';


--
-- Name: nesma_evaluation_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_evaluation_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_evaluation_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_evaluation_history_id_seq OWNED BY public.nesma_evaluation_history.id;


--
-- Name: nesma_evaluations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_evaluations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    evaluation_name character varying(200) NOT NULL,
    evaluation_version character varying(50) NOT NULL,
    evaluation_type character varying(50) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying,
    total_function_points numeric,
    data_function_points numeric,
    transactional_fp numeric,
    adjustment_factor numeric DEFAULT 1,
    adjusted_function_points numeric,
    simple_function_count bigint,
    average_function_count bigint,
    complex_function_count bigint,
    evaluation_config jsonb,
    nesma_rules character varying(50) DEFAULT 'v2.2'::character varying,
    confidence_score numeric,
    accuracy_score numeric,
    compliance_score numeric,
    evaluator_id bigint,
    start_time timestamp with time zone,
    completion_time timestamp with time zone,
    duration bigint,
    evaluation_details jsonb,
    validation_results jsonb,
    recommendation_list jsonb,
    review_status character varying(50) DEFAULT 'pending'::character varying,
    reviewer_id bigint,
    review_comments text,
    review_time timestamp with time zone,
    cycle_id bigint,
    requirement_version_id bigint,
    error_message text,
    error_code character varying(50),
    error_details text,
    progress numeric DEFAULT 0,
    current_phase character varying(100),
    total_steps bigint DEFAULT 0,
    completed_steps bigint DEFAULT 0,
    processed_requirements bigint DEFAULT 0,
    total_requirements bigint DEFAULT 0,
    progress_message character varying(500),
    ai_analysis_id bigint,
    quality_score numeric,
    overall_grade character varying(50),
    complexity_level character varying(50),
    risk_level character varying(50),
    evaluation_summary text
);


--
-- Name: COLUMN nesma_evaluations.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.project_id IS '项目ID';


--
-- Name: COLUMN nesma_evaluations.evaluation_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.evaluation_name IS '评估名称';


--
-- Name: COLUMN nesma_evaluations.evaluation_version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.evaluation_version IS '评估版本';


--
-- Name: COLUMN nesma_evaluations.evaluation_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.evaluation_type IS '评估类型:initial,detailed,final';


--
-- Name: COLUMN nesma_evaluations.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.status IS '评估状态:pending,processing,completed,failed';


--
-- Name: COLUMN nesma_evaluations.total_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.total_function_points IS '总功能点数';


--
-- Name: COLUMN nesma_evaluations.data_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.data_function_points IS '数据功能点数';


--
-- Name: COLUMN nesma_evaluations.transactional_fp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.transactional_fp IS '事务功能点数';


--
-- Name: COLUMN nesma_evaluations.adjustment_factor; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.adjustment_factor IS '调整因子';


--
-- Name: COLUMN nesma_evaluations.adjusted_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.adjusted_function_points IS '调整后功能点数';


--
-- Name: COLUMN nesma_evaluations.simple_function_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.simple_function_count IS '简单功能数量';


--
-- Name: COLUMN nesma_evaluations.average_function_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.average_function_count IS '平均功能数量';


--
-- Name: COLUMN nesma_evaluations.complex_function_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.complex_function_count IS '复杂功能数量';


--
-- Name: COLUMN nesma_evaluations.evaluation_config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.evaluation_config IS '评估配置参数';


--
-- Name: COLUMN nesma_evaluations.nesma_rules; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.nesma_rules IS 'NESMA规则版本';


--
-- Name: COLUMN nesma_evaluations.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.confidence_score IS '评估置信度分数';


--
-- Name: COLUMN nesma_evaluations.accuracy_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.accuracy_score IS '评估准确度分数';


--
-- Name: COLUMN nesma_evaluations.compliance_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.compliance_score IS '标准合规度分数';


--
-- Name: COLUMN nesma_evaluations.evaluator_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.evaluator_id IS '评估人员ID';


--
-- Name: COLUMN nesma_evaluations.start_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.start_time IS '评估开始时间';


--
-- Name: COLUMN nesma_evaluations.completion_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.completion_time IS '评估完成时间';


--
-- Name: COLUMN nesma_evaluations.duration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.duration IS '评估耗时(秒)';


--
-- Name: COLUMN nesma_evaluations.evaluation_details; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.evaluation_details IS '详细评估结果';


--
-- Name: COLUMN nesma_evaluations.validation_results; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.validation_results IS '验证结果';


--
-- Name: COLUMN nesma_evaluations.recommendation_list; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.recommendation_list IS '改进建议列表';


--
-- Name: COLUMN nesma_evaluations.review_status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.review_status IS '审核状态';


--
-- Name: COLUMN nesma_evaluations.reviewer_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.reviewer_id IS '审核人员ID';


--
-- Name: COLUMN nesma_evaluations.review_comments; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.review_comments IS '审核意见';


--
-- Name: COLUMN nesma_evaluations.review_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.review_time IS '审核时间';


--
-- Name: COLUMN nesma_evaluations.cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.cycle_id IS '项目周期ID';


--
-- Name: COLUMN nesma_evaluations.requirement_version_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.requirement_version_id IS '需求版本ID';


--
-- Name: COLUMN nesma_evaluations.error_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.error_message IS '错误信息';


--
-- Name: COLUMN nesma_evaluations.error_code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.error_code IS '错误代码';


--
-- Name: COLUMN nesma_evaluations.error_details; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.error_details IS '错误详情';


--
-- Name: COLUMN nesma_evaluations.progress; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.progress IS '评估进度百分比0-100';


--
-- Name: COLUMN nesma_evaluations.current_phase; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.current_phase IS '当前阶段';


--
-- Name: COLUMN nesma_evaluations.total_steps; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.total_steps IS '总步骤数';


--
-- Name: COLUMN nesma_evaluations.completed_steps; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.completed_steps IS '已完成步骤数';


--
-- Name: COLUMN nesma_evaluations.processed_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.processed_requirements IS '已处理需求数';


--
-- Name: COLUMN nesma_evaluations.total_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.total_requirements IS '总需求数';


--
-- Name: COLUMN nesma_evaluations.progress_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.progress_message IS '进度消息';


--
-- Name: COLUMN nesma_evaluations.ai_analysis_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.ai_analysis_id IS 'AI分析ID';


--
-- Name: COLUMN nesma_evaluations.quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.quality_score IS '质量评分';


--
-- Name: COLUMN nesma_evaluations.overall_grade; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.overall_grade IS '总体评级';


--
-- Name: COLUMN nesma_evaluations.complexity_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.complexity_level IS '复杂度级别';


--
-- Name: COLUMN nesma_evaluations.risk_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.risk_level IS '风险级别';


--
-- Name: COLUMN nesma_evaluations.evaluation_summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_evaluations.evaluation_summary IS '评估总结';


--
-- Name: nesma_evaluations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_evaluations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_evaluations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_evaluations_id_seq OWNED BY public.nesma_evaluations.id;


--
-- Name: nesma_function_points; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_function_points (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    evaluation_id bigint NOT NULL,
    requirement_id bigint,
    function_type character varying(50) NOT NULL,
    function_name character varying(200) NOT NULL,
    function_desc text,
    data_elements bigint,
    file_types bigint,
    record_elements bigint,
    complexity_level character varying(20) NOT NULL,
    weight_factor numeric NOT NULL,
    calculated_points numeric NOT NULL,
    identification_method character varying(50),
    confidence_level numeric,
    detection_rules jsonb,
    is_validated boolean DEFAULT false,
    validation_status character varying(50) DEFAULT 'pending'::character varying,
    validation_notes text
);


--
-- Name: COLUMN nesma_function_points.evaluation_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.evaluation_id IS '评估ID';


--
-- Name: COLUMN nesma_function_points.requirement_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.requirement_id IS '关联需求ID';


--
-- Name: COLUMN nesma_function_points.function_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.function_type IS '功能类型:ILF,EIF,EI,EO,EQ';


--
-- Name: COLUMN nesma_function_points.function_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.function_name IS '功能名称';


--
-- Name: COLUMN nesma_function_points.function_desc; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.function_desc IS '功能描述';


--
-- Name: COLUMN nesma_function_points.data_elements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.data_elements IS '数据元素数';


--
-- Name: COLUMN nesma_function_points.file_types; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.file_types IS '文件类型数';


--
-- Name: COLUMN nesma_function_points.record_elements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.record_elements IS '记录元素数';


--
-- Name: COLUMN nesma_function_points.complexity_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.complexity_level IS '复杂度级别:Low,Average,High';


--
-- Name: COLUMN nesma_function_points.weight_factor; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.weight_factor IS '权重因子';


--
-- Name: COLUMN nesma_function_points.calculated_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.calculated_points IS '计算得出的功能点数';


--
-- Name: COLUMN nesma_function_points.identification_method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.identification_method IS '识别方法:manual,auto,ai_assisted';


--
-- Name: COLUMN nesma_function_points.confidence_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.confidence_level IS '识别置信度';


--
-- Name: COLUMN nesma_function_points.detection_rules; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.detection_rules IS '检测规则';


--
-- Name: COLUMN nesma_function_points.is_validated; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.is_validated IS '是否已验证';


--
-- Name: COLUMN nesma_function_points.validation_status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.validation_status IS '验证状态';


--
-- Name: COLUMN nesma_function_points.validation_notes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_function_points.validation_notes IS '验证备注';


--
-- Name: nesma_function_points_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_function_points_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_function_points_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_function_points_id_seq OWNED BY public.nesma_function_points.id;


--
-- Name: nesma_knowledge_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_knowledge_entries (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title character varying(500) NOT NULL,
    content text NOT NULL,
    category character varying(100) NOT NULL,
    domain character varying(100),
    tags jsonb,
    embedding text,
    confidence_score numeric DEFAULT 1,
    usage_count bigint DEFAULT 0,
    version character varying(50),
    status character varying(50) DEFAULT 'active'::character varying,
    source character varying(255),
    author character varying(255)
);


--
-- Name: COLUMN nesma_knowledge_entries.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.title IS '知识条目标题';


--
-- Name: COLUMN nesma_knowledge_entries.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.content IS '知识内容';


--
-- Name: COLUMN nesma_knowledge_entries.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.category IS '分类：NESMA_STANDARD,BEST_PRACTICE,CASE_STUDY,RULE';


--
-- Name: COLUMN nesma_knowledge_entries.domain; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.domain IS '适用领域';


--
-- Name: COLUMN nesma_knowledge_entries.tags; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.tags IS '标签';


--
-- Name: COLUMN nesma_knowledge_entries.embedding; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.embedding IS '向量表示(JSON格式)';


--
-- Name: COLUMN nesma_knowledge_entries.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.confidence_score IS '置信度评分';


--
-- Name: COLUMN nesma_knowledge_entries.usage_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.usage_count IS '使用次数';


--
-- Name: COLUMN nesma_knowledge_entries.version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.version IS '版本号';


--
-- Name: COLUMN nesma_knowledge_entries.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.status IS '状态';


--
-- Name: COLUMN nesma_knowledge_entries.source; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.source IS '来源';


--
-- Name: COLUMN nesma_knowledge_entries.author; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_entries.author IS '作者';


--
-- Name: nesma_knowledge_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_knowledge_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_knowledge_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_knowledge_entries_id_seq OWNED BY public.nesma_knowledge_entries.id;


--
-- Name: nesma_knowledge_rules; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_knowledge_rules (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    rule_name character varying(255) NOT NULL,
    condition_expression text NOT NULL,
    action_expression text NOT NULL,
    priority bigint DEFAULT 1,
    is_active boolean DEFAULT true,
    description text,
    category character varying(100)
);


--
-- Name: COLUMN nesma_knowledge_rules.rule_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_rules.rule_name IS '规则名称';


--
-- Name: COLUMN nesma_knowledge_rules.condition_expression; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_rules.condition_expression IS '规则条件表达式';


--
-- Name: COLUMN nesma_knowledge_rules.action_expression; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_rules.action_expression IS '规则动作表达式';


--
-- Name: COLUMN nesma_knowledge_rules.priority; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_rules.priority IS '优先级';


--
-- Name: COLUMN nesma_knowledge_rules.is_active; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_rules.is_active IS '是否激活';


--
-- Name: COLUMN nesma_knowledge_rules.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_rules.description IS '规则描述';


--
-- Name: COLUMN nesma_knowledge_rules.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_knowledge_rules.category IS '规则分类';


--
-- Name: nesma_knowledge_rules_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_knowledge_rules_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_knowledge_rules_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_knowledge_rules_id_seq OWNED BY public.nesma_knowledge_rules.id;


--
-- Name: nesma_nesma_evaluations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_nesma_evaluations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    cycle_id bigint NOT NULL,
    evaluation_type character varying(50) NOT NULL,
    overall_grade character varying(10),
    total_afp numeric,
    total_ufp numeric,
    quality_score numeric,
    nesma_compliance numeric,
    industry_ranking character varying(50),
    investment_grade character varying(20),
    ei_count bigint,
    eo_count bigint,
    eq_count bigint,
    ilf_count bigint,
    eif_count bigint,
    low_complexity_count bigint,
    avg_complexity_count bigint,
    high_complexity_count bigint,
    measurement_consistency numeric,
    documentation_completeness numeric,
    traceability_score numeric,
    evaluation_result jsonb,
    function_point_analysis jsonb,
    quality_evaluation jsonb,
    compliance_check jsonb,
    benchmarking_analysis jsonb,
    key_findings jsonb,
    critical_recommendations jsonb,
    improvement_roadmap jsonb,
    evaluator_model character varying(50),
    evaluation_standard character varying(50),
    confidence_level numeric,
    evaluation_date timestamp with time zone NOT NULL
);


--
-- Name: COLUMN nesma_nesma_evaluations.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.project_id IS '项目ID';


--
-- Name: COLUMN nesma_nesma_evaluations.cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.cycle_id IS '周期ID';


--
-- Name: COLUMN nesma_nesma_evaluations.evaluation_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.evaluation_type IS '评估类型：comprehensive-全面评估,compliance-合规检查,benchmark-基准对比';


--
-- Name: COLUMN nesma_nesma_evaluations.overall_grade; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.overall_grade IS '总体评级';


--
-- Name: COLUMN nesma_nesma_evaluations.total_afp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.total_afp IS '总调整功能点数';


--
-- Name: COLUMN nesma_nesma_evaluations.total_ufp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.total_ufp IS '总未调整功能点数';


--
-- Name: COLUMN nesma_nesma_evaluations.quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.quality_score IS '质量评分(0-100)';


--
-- Name: COLUMN nesma_nesma_evaluations.nesma_compliance; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.nesma_compliance IS 'NESMA合规评分(0-100)';


--
-- Name: COLUMN nesma_nesma_evaluations.industry_ranking; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.industry_ranking IS '行业排名';


--
-- Name: COLUMN nesma_nesma_evaluations.investment_grade; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.investment_grade IS '投资级别';


--
-- Name: COLUMN nesma_nesma_evaluations.ei_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.ei_count IS '外部输入数量';


--
-- Name: COLUMN nesma_nesma_evaluations.eo_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.eo_count IS '外部输出数量';


--
-- Name: COLUMN nesma_nesma_evaluations.eq_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.eq_count IS '外部查询数量';


--
-- Name: COLUMN nesma_nesma_evaluations.ilf_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.ilf_count IS '内部逻辑文件数量';


--
-- Name: COLUMN nesma_nesma_evaluations.eif_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.eif_count IS '外部接口文件数量';


--
-- Name: COLUMN nesma_nesma_evaluations.low_complexity_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.low_complexity_count IS '低复杂度数量';


--
-- Name: COLUMN nesma_nesma_evaluations.avg_complexity_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.avg_complexity_count IS '中等复杂度数量';


--
-- Name: COLUMN nesma_nesma_evaluations.high_complexity_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.high_complexity_count IS '高复杂度数量';


--
-- Name: COLUMN nesma_nesma_evaluations.measurement_consistency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.measurement_consistency IS '测量一致性(0-100)';


--
-- Name: COLUMN nesma_nesma_evaluations.documentation_completeness; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.documentation_completeness IS '文档完整性(0-100)';


--
-- Name: COLUMN nesma_nesma_evaluations.traceability_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.traceability_score IS '可追溯性评分(0-100)';


--
-- Name: COLUMN nesma_nesma_evaluations.evaluation_result; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.evaluation_result IS '详细评估结果';


--
-- Name: COLUMN nesma_nesma_evaluations.function_point_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.function_point_analysis IS '功能点分析';


--
-- Name: COLUMN nesma_nesma_evaluations.quality_evaluation; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.quality_evaluation IS '质量评估';


--
-- Name: COLUMN nesma_nesma_evaluations.compliance_check; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.compliance_check IS '合规检查';


--
-- Name: COLUMN nesma_nesma_evaluations.benchmarking_analysis; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.benchmarking_analysis IS '基准对比分析';


--
-- Name: COLUMN nesma_nesma_evaluations.key_findings; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.key_findings IS '关键发现';


--
-- Name: COLUMN nesma_nesma_evaluations.critical_recommendations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.critical_recommendations IS '关键建议';


--
-- Name: COLUMN nesma_nesma_evaluations.improvement_roadmap; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.improvement_roadmap IS '改进路线图';


--
-- Name: COLUMN nesma_nesma_evaluations.evaluator_model; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.evaluator_model IS '评估AI模型';


--
-- Name: COLUMN nesma_nesma_evaluations.evaluation_standard; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.evaluation_standard IS '评估标准';


--
-- Name: COLUMN nesma_nesma_evaluations.confidence_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.confidence_level IS '置信水平(0-1)';


--
-- Name: COLUMN nesma_nesma_evaluations.evaluation_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_nesma_evaluations.evaluation_date IS '评估日期';


--
-- Name: nesma_nesma_evaluations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_nesma_evaluations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_nesma_evaluations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_nesma_evaluations_id_seq OWNED BY public.nesma_nesma_evaluations.id;


--
-- Name: nesma_optimization_suggestions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_optimization_suggestions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    requirement_id bigint,
    suggestion_type character varying(50) NOT NULL,
    category character varying(50) NOT NULL,
    title character varying(200) NOT NULL,
    description text NOT NULL,
    rationale text,
    expected_outcome text,
    priority character varying(20) NOT NULL,
    impact character varying(20),
    effort character varying(20),
    action_items jsonb,
    implementation_notes text,
    timeline character varying(100),
    resource_requirements text,
    status character varying(20) DEFAULT 'pending'::character varying,
    assigned_to character varying(100),
    start_date timestamp with time zone,
    due_date timestamp with time zone,
    completed_date timestamp with time zone,
    actual_outcome text,
    effectiveness_score numeric,
    generated_by character varying(50),
    ai_model character varying(50),
    confidence_score numeric
);


--
-- Name: COLUMN nesma_optimization_suggestions.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.project_id IS '项目ID';


--
-- Name: COLUMN nesma_optimization_suggestions.requirement_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.requirement_id IS '需求ID(针对单个需求的建议)';


--
-- Name: COLUMN nesma_optimization_suggestions.suggestion_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.suggestion_type IS '建议类型：project-项目级,requirement-需求级,process-流程级';


--
-- Name: COLUMN nesma_optimization_suggestions.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.category IS '建议类别：quality-质量改进,efficiency-效率提升,risk-风险控制,compliance-合规性';


--
-- Name: COLUMN nesma_optimization_suggestions.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.title IS '建议标题';


--
-- Name: COLUMN nesma_optimization_suggestions.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.description IS '建议描述';


--
-- Name: COLUMN nesma_optimization_suggestions.rationale; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.rationale IS '建议理由';


--
-- Name: COLUMN nesma_optimization_suggestions.expected_outcome; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.expected_outcome IS '预期结果';


--
-- Name: COLUMN nesma_optimization_suggestions.priority; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.priority IS '优先级：High-高,Medium-中,Low-低';


--
-- Name: COLUMN nesma_optimization_suggestions.impact; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.impact IS '影响程度：High-高,Medium-中,Low-低';


--
-- Name: COLUMN nesma_optimization_suggestions.effort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.effort IS '实施工作量：High-高,Medium-中,Low-低';


--
-- Name: COLUMN nesma_optimization_suggestions.action_items; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.action_items IS '行动项列表';


--
-- Name: COLUMN nesma_optimization_suggestions.implementation_notes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.implementation_notes IS '实施注意事项';


--
-- Name: COLUMN nesma_optimization_suggestions.timeline; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.timeline IS '时间线';


--
-- Name: COLUMN nesma_optimization_suggestions.resource_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.resource_requirements IS '资源需求';


--
-- Name: COLUMN nesma_optimization_suggestions.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.status IS '状态：pending-待处理,in_progress-进行中,completed-已完成,rejected-已拒绝';


--
-- Name: COLUMN nesma_optimization_suggestions.assigned_to; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.assigned_to IS '指派给';


--
-- Name: COLUMN nesma_optimization_suggestions.start_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.start_date IS '开始日期';


--
-- Name: COLUMN nesma_optimization_suggestions.due_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.due_date IS '截止日期';


--
-- Name: COLUMN nesma_optimization_suggestions.completed_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.completed_date IS '完成日期';


--
-- Name: COLUMN nesma_optimization_suggestions.actual_outcome; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.actual_outcome IS '实际结果';


--
-- Name: COLUMN nesma_optimization_suggestions.effectiveness_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.effectiveness_score IS '有效性评分(0-100)';


--
-- Name: COLUMN nesma_optimization_suggestions.generated_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.generated_by IS '生成方式：ai-AI生成,manual-手动添加';


--
-- Name: COLUMN nesma_optimization_suggestions.ai_model; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.ai_model IS 'AI模型';


--
-- Name: COLUMN nesma_optimization_suggestions.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_optimization_suggestions.confidence_score IS '置信度(0-1)';


--
-- Name: nesma_optimization_suggestions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_optimization_suggestions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_optimization_suggestions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_optimization_suggestions_id_seq OWNED BY public.nesma_optimization_suggestions.id;


--
-- Name: nesma_project_analyses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_project_analyses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    cycle_id bigint NOT NULL,
    version_id bigint NOT NULL,
    analysis_type character varying(50) NOT NULL,
    analysis_scope character varying(100),
    analysis_date timestamp with time zone NOT NULL,
    overall_grade character varying(10),
    health_score numeric,
    success_probability numeric,
    quality_score numeric,
    total_requirements bigint,
    total_function_points numeric,
    complexity_index numeric,
    analysis_result jsonb,
    key_strengths jsonb,
    major_risks jsonb,
    recommendations jsonb,
    ai_model character varying(50),
    confidence_score numeric
);


--
-- Name: COLUMN nesma_project_analyses.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.project_id IS '项目ID';


--
-- Name: COLUMN nesma_project_analyses.cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.cycle_id IS '周期ID';


--
-- Name: COLUMN nesma_project_analyses.version_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.version_id IS '版本ID';


--
-- Name: COLUMN nesma_project_analyses.analysis_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.analysis_type IS '分析类型：project_overview-项目概览,requirement_analysis-需求分析,quality_assessment-质量评估';


--
-- Name: COLUMN nesma_project_analyses.analysis_scope; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.analysis_scope IS '分析范围';


--
-- Name: COLUMN nesma_project_analyses.analysis_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.analysis_date IS '分析日期';


--
-- Name: COLUMN nesma_project_analyses.overall_grade; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.overall_grade IS '总体评级：A+,A,A-,B+,B,B-,C+,C,C-,D,F';


--
-- Name: COLUMN nesma_project_analyses.health_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.health_score IS '项目健康分(0-100)';


--
-- Name: COLUMN nesma_project_analyses.success_probability; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.success_probability IS '成功概率(0-100)';


--
-- Name: COLUMN nesma_project_analyses.quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.quality_score IS '质量评分(0-100)';


--
-- Name: COLUMN nesma_project_analyses.total_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.total_requirements IS '总需求数量';


--
-- Name: COLUMN nesma_project_analyses.total_function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.total_function_points IS '总功能点数';


--
-- Name: COLUMN nesma_project_analyses.complexity_index; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.complexity_index IS '复杂度指数';


--
-- Name: COLUMN nesma_project_analyses.analysis_result; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.analysis_result IS '详细分析结果';


--
-- Name: COLUMN nesma_project_analyses.key_strengths; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.key_strengths IS '核心优势';


--
-- Name: COLUMN nesma_project_analyses.major_risks; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.major_risks IS '主要风险';


--
-- Name: COLUMN nesma_project_analyses.recommendations; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.recommendations IS '改进建议';


--
-- Name: COLUMN nesma_project_analyses.ai_model; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.ai_model IS '使用的AI模型';


--
-- Name: COLUMN nesma_project_analyses.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_analyses.confidence_score IS 'AI置信度(0-1)';


--
-- Name: nesma_project_analyses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_project_analyses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_project_analyses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_project_analyses_id_seq OWNED BY public.nesma_project_analyses.id;


--
-- Name: nesma_project_cycles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_project_cycles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    status character varying(50) DEFAULT 'planning'::character varying,
    phase character varying(50),
    settings jsonb,
    requirement_count bigint DEFAULT 0,
    completed_count bigint DEFAULT 0,
    progress numeric DEFAULT 0
);


--
-- Name: COLUMN nesma_project_cycles.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.project_id IS '项目ID';


--
-- Name: COLUMN nesma_project_cycles.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.name IS '周期名称，如：一期、二期、三期';


--
-- Name: COLUMN nesma_project_cycles.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.description IS '周期描述';


--
-- Name: COLUMN nesma_project_cycles.start_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.start_date IS '周期开始时间';


--
-- Name: COLUMN nesma_project_cycles.end_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.end_date IS '周期结束时间';


--
-- Name: COLUMN nesma_project_cycles.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.status IS '周期状态：planning-规划中,active-进行中,completed-已完成,suspended-暂停';


--
-- Name: COLUMN nesma_project_cycles.phase; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.phase IS '当前阶段：需求分析,设计,开发,测试,上线';


--
-- Name: COLUMN nesma_project_cycles.settings; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.settings IS '周期特有配置';


--
-- Name: COLUMN nesma_project_cycles.requirement_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.requirement_count IS '需求总数';


--
-- Name: COLUMN nesma_project_cycles.completed_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.completed_count IS '已完成需求数';


--
-- Name: COLUMN nesma_project_cycles.progress; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_project_cycles.progress IS '完成进度百分比';


--
-- Name: nesma_project_cycles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_project_cycles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_project_cycles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_project_cycles_id_seq OWNED BY public.nesma_project_cycles.id;


--
-- Name: nesma_projects; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_projects (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name character varying(255) NOT NULL,
    description text,
    owner_id bigint NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying,
    domain character varying(100),
    domain_tags jsonb,
    settings jsonb,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    active_cycle_id bigint
);


--
-- Name: COLUMN nesma_projects.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.name IS '项目名称';


--
-- Name: COLUMN nesma_projects.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.description IS '项目描述';


--
-- Name: COLUMN nesma_projects.owner_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.owner_id IS '项目负责人ID';


--
-- Name: COLUMN nesma_projects.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.status IS '项目状态';


--
-- Name: COLUMN nesma_projects.domain; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.domain IS '项目领域（金融、电商、医疗等）';


--
-- Name: COLUMN nesma_projects.domain_tags; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.domain_tags IS '领域标签';


--
-- Name: COLUMN nesma_projects.settings; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.settings IS '项目配置';


--
-- Name: COLUMN nesma_projects.start_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.start_date IS '项目开始时间';


--
-- Name: COLUMN nesma_projects.end_date; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.end_date IS '项目结束时间';


--
-- Name: COLUMN nesma_projects.active_cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_projects.active_cycle_id IS '当前激活的周期ID';


--
-- Name: nesma_projects_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_projects_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_projects_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_projects_id_seq OWNED BY public.nesma_projects.id;


--
-- Name: nesma_requirement_analysis; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_requirement_analysis (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    requirement_id bigint NOT NULL,
    ai_expanded_content text,
    nesma_category character varying(100),
    complexity_score numeric,
    function_points numeric,
    confidence_level numeric,
    knowledge_base_refs jsonb,
    analysis_version character varying(50),
    analysis_details jsonb,
    review_status character varying(50) DEFAULT 'pending'::character varying,
    review_comments text
);


--
-- Name: COLUMN nesma_requirement_analysis.requirement_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.requirement_id IS '需求ID';


--
-- Name: COLUMN nesma_requirement_analysis.ai_expanded_content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.ai_expanded_content IS 'AI扩展的需求内容';


--
-- Name: COLUMN nesma_requirement_analysis.nesma_category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.nesma_category IS 'NESMA分类(ILF,EIF,EI,EO,EQ)';


--
-- Name: COLUMN nesma_requirement_analysis.complexity_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.complexity_score IS '复杂度评分';


--
-- Name: COLUMN nesma_requirement_analysis.function_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.function_points IS '功能点数';


--
-- Name: COLUMN nesma_requirement_analysis.confidence_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.confidence_level IS '置信度';


--
-- Name: COLUMN nesma_requirement_analysis.knowledge_base_refs; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.knowledge_base_refs IS '引用的知识库条目';


--
-- Name: COLUMN nesma_requirement_analysis.analysis_version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.analysis_version IS '分析版本';


--
-- Name: COLUMN nesma_requirement_analysis.analysis_details; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.analysis_details IS '详细分析结果';


--
-- Name: COLUMN nesma_requirement_analysis.review_status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.review_status IS '审核状态';


--
-- Name: COLUMN nesma_requirement_analysis.review_comments; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis.review_comments IS '审核意见';


--
-- Name: nesma_requirement_analysis_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_requirement_analysis_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_requirement_analysis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_requirement_analysis_id_seq OWNED BY public.nesma_requirement_analysis.id;


--
-- Name: nesma_requirement_analysis_tasks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_requirement_analysis_tasks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    cycle_id bigint NOT NULL,
    source_version_id bigint NOT NULL,
    target_version_id bigint,
    task_type character varying(50) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying,
    progress bigint DEFAULT 0,
    config jsonb,
    priority bigint DEFAULT 5,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    duration bigint,
    total_count bigint DEFAULT 0,
    processed_count bigint DEFAULT 0,
    success_count bigint DEFAULT 0,
    failed_count bigint DEFAULT 0,
    skipped_count bigint DEFAULT 0,
    error_msg text,
    error_details jsonb,
    result jsonb,
    summary text,
    ai_call_count bigint DEFAULT 0,
    ai_call_duration bigint,
    ai_token_usage bigint,
    quality_score numeric,
    accuracy_rate numeric,
    requirement_ids jsonb
);


--
-- Name: COLUMN nesma_requirement_analysis_tasks.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.project_id IS '项目ID';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.cycle_id IS '周期ID';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.source_version_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.source_version_id IS '源版本ID';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.target_version_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.target_version_id IS '目标版本ID';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.task_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.task_type IS '任务类型：requirement_analysis-需求分析,description_generation-描述生成,flowchart_generation-流程图生成';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.status IS '任务状态：pending-等待中,running-执行中,completed-已完成,failed-失败,cancelled-已取消';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.progress; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.progress IS '执行进度0-100';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.config; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.config IS '任务配置参数';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.priority; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.priority IS '任务优先级1-10，数字越小优先级越高';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.start_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.start_time IS '任务开始时间';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.end_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.end_time IS '任务结束时间';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.duration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.duration IS '执行耗时（秒）';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.total_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.total_count IS '总需求数量';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.processed_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.processed_count IS '已处理数量';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.success_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.success_count IS '成功处理数量';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.failed_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.failed_count IS '失败处理数量';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.skipped_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.skipped_count IS '跳过处理数量';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.error_msg; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.error_msg IS '错误信息';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.error_details; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.error_details IS '详细错误信息';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.result; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.result IS '任务执行结果';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.summary IS '任务执行摘要';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.ai_call_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.ai_call_count IS 'AI调用次数';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.ai_call_duration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.ai_call_duration IS 'AI调用总耗时（毫秒）';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.ai_token_usage; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.ai_token_usage IS 'AI Token使用量';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.quality_score IS '输出质量评分（0-100）';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.accuracy_rate; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.accuracy_rate IS '准确率（0-1）';


--
-- Name: COLUMN nesma_requirement_analysis_tasks.requirement_ids; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_analysis_tasks.requirement_ids IS '需求ID列表';


--
-- Name: nesma_requirement_analysis_tasks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_requirement_analysis_tasks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_requirement_analysis_tasks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_requirement_analysis_tasks_id_seq OWNED BY public.nesma_requirement_analysis_tasks.id;


--
-- Name: nesma_requirement_optimizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_requirement_optimizations (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    requirement_id bigint NOT NULL,
    project_id bigint NOT NULL,
    optimization_type character varying(50) NOT NULL,
    original_title character varying(500),
    original_description text,
    original_quality_score numeric,
    optimized_title character varying(500),
    optimized_description text,
    optimized_quality_score numeric,
    improvement_summary jsonb,
    quality_comparison jsonb,
    optimization_rationale text,
    user_decision_required boolean DEFAULT false,
    decision_points jsonb,
    user_decision character varying(20),
    user_decision_time timestamp with time zone,
    user_decision_reason text,
    applied_to_requirement boolean DEFAULT false,
    applied_time timestamp with time zone,
    ai_model character varying(50),
    analysis_confidence numeric,
    generated_at timestamp with time zone NOT NULL
);


--
-- Name: COLUMN nesma_requirement_optimizations.requirement_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.requirement_id IS '需求ID';


--
-- Name: COLUMN nesma_requirement_optimizations.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.project_id IS '项目ID';


--
-- Name: COLUMN nesma_requirement_optimizations.optimization_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.optimization_type IS '优化类型：title-标题优化,description-描述优化,structure-结构优化,quality-质量提升';


--
-- Name: COLUMN nesma_requirement_optimizations.original_title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.original_title IS '原始标题';


--
-- Name: COLUMN nesma_requirement_optimizations.original_description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.original_description IS '原始描述';


--
-- Name: COLUMN nesma_requirement_optimizations.original_quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.original_quality_score IS '原始质量评分';


--
-- Name: COLUMN nesma_requirement_optimizations.optimized_title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.optimized_title IS '优化后标题';


--
-- Name: COLUMN nesma_requirement_optimizations.optimized_description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.optimized_description IS '优化后描述';


--
-- Name: COLUMN nesma_requirement_optimizations.optimized_quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.optimized_quality_score IS '优化后质量评分';


--
-- Name: COLUMN nesma_requirement_optimizations.improvement_summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.improvement_summary IS '改进摘要';


--
-- Name: COLUMN nesma_requirement_optimizations.quality_comparison; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.quality_comparison IS '质量对比';


--
-- Name: COLUMN nesma_requirement_optimizations.optimization_rationale; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.optimization_rationale IS '优化理由';


--
-- Name: COLUMN nesma_requirement_optimizations.user_decision_required; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.user_decision_required IS '是否需要用户决策';


--
-- Name: COLUMN nesma_requirement_optimizations.decision_points; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.decision_points IS '决策点';


--
-- Name: COLUMN nesma_requirement_optimizations.user_decision; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.user_decision IS '用户决策：accepted-接受,rejected-拒绝,pending-待决策';


--
-- Name: COLUMN nesma_requirement_optimizations.user_decision_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.user_decision_time IS '用户决策时间';


--
-- Name: COLUMN nesma_requirement_optimizations.user_decision_reason; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.user_decision_reason IS '用户决策理由';


--
-- Name: COLUMN nesma_requirement_optimizations.applied_to_requirement; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.applied_to_requirement IS '是否已应用到需求';


--
-- Name: COLUMN nesma_requirement_optimizations.applied_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.applied_time IS '应用时间';


--
-- Name: COLUMN nesma_requirement_optimizations.ai_model; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.ai_model IS '使用的AI模型';


--
-- Name: COLUMN nesma_requirement_optimizations.analysis_confidence; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.analysis_confidence IS '分析置信度(0-1)';


--
-- Name: COLUMN nesma_requirement_optimizations.generated_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_optimizations.generated_at IS '生成时间';


--
-- Name: nesma_requirement_optimizations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_requirement_optimizations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_requirement_optimizations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_requirement_optimizations_id_seq OWNED BY public.nesma_requirement_optimizations.id;


--
-- Name: nesma_requirement_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_requirement_versions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    cycle_id bigint NOT NULL,
    version character varying(20) NOT NULL,
    version_type character varying(50) NOT NULL,
    created_by character varying(100) NOT NULL,
    summary character varying(500),
    description text,
    changes jsonb,
    status character varying(50) DEFAULT 'draft'::character varying,
    requirement_count bigint DEFAULT 0,
    analyzed_count bigint DEFAULT 0,
    optimized_count bigint DEFAULT 0,
    analysis_progress numeric DEFAULT 0,
    analysis_task_id bigint,
    analysis_start_time timestamp with time zone,
    analysis_end_time timestamp with time zone,
    analysis_duration bigint,
    quality_score numeric,
    confidence_score numeric,
    review_status character varying(50) DEFAULT 'pending'::character varying,
    review_notes text
);


--
-- Name: COLUMN nesma_requirement_versions.cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.cycle_id IS '所属周期ID';


--
-- Name: COLUMN nesma_requirement_versions.version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.version IS '版本号，如：v1.0, v1.1, v1.2';


--
-- Name: COLUMN nesma_requirement_versions.version_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.version_type IS '版本类型：initial-初始版本,analyzed-AI分析版本,optimized-人工优化版本,finalized-最终版本';


--
-- Name: COLUMN nesma_requirement_versions.created_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.created_by IS '创建方式：user-用户创建,ai_analysis-AI分析,manual_edit-手动编辑';


--
-- Name: COLUMN nesma_requirement_versions.summary; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.summary IS '版本变更摘要';


--
-- Name: COLUMN nesma_requirement_versions.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.description IS '详细版本说明';


--
-- Name: COLUMN nesma_requirement_versions.changes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.changes IS '变更记录详情';


--
-- Name: COLUMN nesma_requirement_versions.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.status IS '版本状态：draft-草稿,active-激活,archived-归档';


--
-- Name: COLUMN nesma_requirement_versions.requirement_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.requirement_count IS '该版本需求总数';


--
-- Name: COLUMN nesma_requirement_versions.analyzed_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.analyzed_count IS '已分析需求数';


--
-- Name: COLUMN nesma_requirement_versions.optimized_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.optimized_count IS '已优化需求数';


--
-- Name: COLUMN nesma_requirement_versions.analysis_progress; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.analysis_progress IS '分析进度百分比';


--
-- Name: COLUMN nesma_requirement_versions.analysis_task_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.analysis_task_id IS '关联的分析任务ID';


--
-- Name: COLUMN nesma_requirement_versions.analysis_start_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.analysis_start_time IS '分析开始时间';


--
-- Name: COLUMN nesma_requirement_versions.analysis_end_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.analysis_end_time IS '分析结束时间';


--
-- Name: COLUMN nesma_requirement_versions.analysis_duration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.analysis_duration IS '分析耗时（秒）';


--
-- Name: COLUMN nesma_requirement_versions.quality_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.quality_score IS '版本质量评分（0-100）';


--
-- Name: COLUMN nesma_requirement_versions.confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.confidence_score IS 'AI分析置信度（0-1）';


--
-- Name: COLUMN nesma_requirement_versions.review_status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.review_status IS '评审状态：pending-待评审,approved-已通过,rejected-已拒绝';


--
-- Name: COLUMN nesma_requirement_versions.review_notes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirement_versions.review_notes IS '评审意见';


--
-- Name: nesma_requirement_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_requirement_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_requirement_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_requirement_versions_id_seq OWNED BY public.nesma_requirement_versions.id;


--
-- Name: nesma_requirements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_requirements (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    project_id bigint NOT NULL,
    parent_id bigint,
    level bigint NOT NULL,
    title character varying(500) NOT NULL,
    description text,
    priority bigint DEFAULT 3,
    status character varying(50) DEFAULT 'pending'::character varying,
    order_index bigint DEFAULT 0,
    domain_tags jsonb,
    code character varying(100),
    category character varying(100),
    complexity character varying(50),
    estimate_hours numeric,
    actual_hours numeric,
    business_value text,
    acceptance_criteria text,
    notes text,
    import_batch character varying(100),
    import_source character varying(100),
    construction_period character varying(100),
    version bigint DEFAULT 0,
    afp numeric,
    ufp numeric,
    function_type character varying(10),
    reuse_level character varying(20),
    modification_type character varying(20),
    cycle_id bigint,
    version_id bigint,
    ai_analysis_status character varying(50) DEFAULT 'pending'::character varying,
    ai_description text,
    ai_generated_title character varying(500),
    ai_complexity_score numeric,
    ai_confidence_score numeric,
    ai_analysis_time timestamp with time zone,
    ai_analysis_log jsonb,
    similar_requirements jsonb,
    recommended_afp numeric,
    recommended_ufp numeric,
    ai_optimization_applied boolean DEFAULT false,
    ai_optimization_applied_at timestamp with time zone,
    complexity_level character varying(50),
    knowledge_references text,
    a_idescription text,
    mermaid_diagram text,
    mermaid_generated_at timestamp with time zone
);


--
-- Name: COLUMN nesma_requirements.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.project_id IS '项目ID';


--
-- Name: COLUMN nesma_requirements.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.parent_id IS '父需求ID，支持层级结构';


--
-- Name: COLUMN nesma_requirements.level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.level IS '需求层级：1-一级功能模块,2-二级功能模块,3-三级功能模块,4-功能点计数项';


--
-- Name: COLUMN nesma_requirements.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.title IS '需求标题';


--
-- Name: COLUMN nesma_requirements.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.description IS '需求描述';


--
-- Name: COLUMN nesma_requirements.priority; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.priority IS '优先级1-5';


--
-- Name: COLUMN nesma_requirements.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.status IS '需求状态';


--
-- Name: COLUMN nesma_requirements.order_index; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.order_index IS '排序索引';


--
-- Name: COLUMN nesma_requirements.domain_tags; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.domain_tags IS '领域标签';


--
-- Name: COLUMN nesma_requirements.code; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.code IS '需求编号';


--
-- Name: COLUMN nesma_requirements.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.category IS '功能分类';


--
-- Name: COLUMN nesma_requirements.complexity; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.complexity IS '复杂度：简单/中等/复杂';


--
-- Name: COLUMN nesma_requirements.estimate_hours; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.estimate_hours IS '预估工时';


--
-- Name: COLUMN nesma_requirements.actual_hours; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.actual_hours IS '实际工时';


--
-- Name: COLUMN nesma_requirements.business_value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.business_value IS '业务价值';


--
-- Name: COLUMN nesma_requirements.acceptance_criteria; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.acceptance_criteria IS '验收标准';


--
-- Name: COLUMN nesma_requirements.notes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.notes IS '备注';


--
-- Name: COLUMN nesma_requirements.import_batch; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.import_batch IS '导入批次号';


--
-- Name: COLUMN nesma_requirements.import_source; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.import_source IS '导入来源';


--
-- Name: COLUMN nesma_requirements.construction_period; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.construction_period IS '建设周期';


--
-- Name: COLUMN nesma_requirements.version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.version IS '版本';


--
-- Name: COLUMN nesma_requirements.afp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.afp IS '调整后功能点数';


--
-- Name: COLUMN nesma_requirements.ufp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ufp IS '未调整功能点数';


--
-- Name: COLUMN nesma_requirements.function_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.function_type IS '功能类别：EQ,EI,EO,ILF,EIF';


--
-- Name: COLUMN nesma_requirements.reuse_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.reuse_level IS '重用程度：高,中,低';


--
-- Name: COLUMN nesma_requirements.modification_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.modification_type IS '修改类型：新增,优化,删除';


--
-- Name: COLUMN nesma_requirements.cycle_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.cycle_id IS '所属项目周期ID';


--
-- Name: COLUMN nesma_requirements.version_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.version_id IS '所属需求版本ID';


--
-- Name: COLUMN nesma_requirements.ai_analysis_status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_analysis_status IS 'AI分析状态：pending-待分析,analyzing-分析中,completed-已完成,failed-失败';


--
-- Name: COLUMN nesma_requirements.ai_description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_description IS 'AI生成的需求描述';


--
-- Name: COLUMN nesma_requirements.ai_generated_title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_generated_title IS 'AI优化的需求标题';


--
-- Name: COLUMN nesma_requirements.ai_complexity_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_complexity_score IS 'AI评估的复杂度得分（0-100）';


--
-- Name: COLUMN nesma_requirements.ai_confidence_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_confidence_score IS 'AI分析置信度（0-1）';


--
-- Name: COLUMN nesma_requirements.ai_analysis_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_analysis_time IS 'AI分析完成时间';


--
-- Name: COLUMN nesma_requirements.ai_analysis_log; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_analysis_log IS 'AI分析过程日志';


--
-- Name: COLUMN nesma_requirements.similar_requirements; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.similar_requirements IS '相似需求推荐';


--
-- Name: COLUMN nesma_requirements.recommended_afp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.recommended_afp IS 'AI推荐的AFP值';


--
-- Name: COLUMN nesma_requirements.recommended_ufp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.recommended_ufp IS 'AI推荐的UFP值';


--
-- Name: COLUMN nesma_requirements.ai_optimization_applied; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_optimization_applied IS '是否已应用AI优化建议';


--
-- Name: COLUMN nesma_requirements.ai_optimization_applied_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.ai_optimization_applied_at IS 'AI优化应用时间';


--
-- Name: COLUMN nesma_requirements.complexity_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.complexity_level IS '复杂度级别';


--
-- Name: COLUMN nesma_requirements.knowledge_references; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.knowledge_references IS '知识库引用';


--
-- Name: COLUMN nesma_requirements.a_idescription; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.a_idescription IS 'AI生成的需求描述';


--
-- Name: COLUMN nesma_requirements.mermaid_diagram; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.mermaid_diagram IS 'Mermaid流程图代码';


--
-- Name: COLUMN nesma_requirements.mermaid_generated_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_requirements.mermaid_generated_at IS '流程图生成时间';


--
-- Name: nesma_requirements_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_requirements_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_requirements_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_requirements_id_seq OWNED BY public.nesma_requirements.id;


--
-- Name: nesma_session_memories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_session_memories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    session_id character varying(255) NOT NULL,
    user_id bigint NOT NULL,
    conversation_history jsonb,
    expires_at timestamp with time zone,
    is_active boolean DEFAULT true
);


--
-- Name: COLUMN nesma_session_memories.session_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_session_memories.session_id IS '会话ID';


--
-- Name: COLUMN nesma_session_memories.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_session_memories.user_id IS '用户ID';


--
-- Name: COLUMN nesma_session_memories.conversation_history; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_session_memories.conversation_history IS '对话历史';


--
-- Name: COLUMN nesma_session_memories.expires_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_session_memories.expires_at IS '过期时间';


--
-- Name: COLUMN nesma_session_memories.is_active; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_session_memories.is_active IS '是否激活';


--
-- Name: nesma_session_memories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_session_memories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_session_memories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_session_memories_id_seq OWNED BY public.nesma_session_memories.id;


--
-- Name: nesma_system_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_system_activities (
    id bigint NOT NULL,
    type character varying(50) NOT NULL,
    category character varying(100) NOT NULL,
    title character varying(255) NOT NULL,
    message text,
    user_id bigint,
    user_name character varying(100),
    project_id bigint,
    project_name character varying(255),
    entity_type character varying(50),
    entity_id bigint,
    metadata text,
    ip_address character varying(45),
    user_agent character varying(500),
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: COLUMN nesma_system_activities.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.type IS '活动类型(success/info/warning/error)';


--
-- Name: COLUMN nesma_system_activities.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.category IS '活动分类(project/requirement/analysis/system)';


--
-- Name: COLUMN nesma_system_activities.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.title IS '活动标题';


--
-- Name: COLUMN nesma_system_activities.message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.message IS '活动详细消息';


--
-- Name: COLUMN nesma_system_activities.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.user_id IS '操作用户ID';


--
-- Name: COLUMN nesma_system_activities.user_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.user_name IS '操作用户名';


--
-- Name: COLUMN nesma_system_activities.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.project_id IS '关联项目ID';


--
-- Name: COLUMN nesma_system_activities.project_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.project_name IS '关联项目名称';


--
-- Name: COLUMN nesma_system_activities.entity_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.entity_type IS '关联实体类型';


--
-- Name: COLUMN nesma_system_activities.entity_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.entity_id IS '关联实体ID';


--
-- Name: COLUMN nesma_system_activities.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.metadata IS '额外元数据';


--
-- Name: COLUMN nesma_system_activities.ip_address; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.ip_address IS '操作IP地址';


--
-- Name: COLUMN nesma_system_activities.user_agent; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_system_activities.user_agent IS '用户代理';


--
-- Name: nesma_system_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_system_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_system_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_system_activities_id_seq OWNED BY public.nesma_system_activities.id;


--
-- Name: nesma_validation_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_validation_items (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    evaluation_id bigint NOT NULL,
    validation_rule character varying(200) NOT NULL,
    rule_description text,
    validation_result character varying(50) NOT NULL,
    expected_value character varying(200),
    actual_value character varying(200),
    deviation_level character varying(50),
    validation_notes text,
    validation_details jsonb,
    is_critical boolean DEFAULT false,
    resolution_status character varying(50) DEFAULT 'open'::character varying,
    resolution_notes text
);


--
-- Name: COLUMN nesma_validation_items.evaluation_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.evaluation_id IS '评估ID';


--
-- Name: COLUMN nesma_validation_items.validation_rule; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.validation_rule IS '验证规则';


--
-- Name: COLUMN nesma_validation_items.rule_description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.rule_description IS '规则描述';


--
-- Name: COLUMN nesma_validation_items.validation_result; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.validation_result IS '验证结果:pass,fail,warning,skip';


--
-- Name: COLUMN nesma_validation_items.expected_value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.expected_value IS '期望值';


--
-- Name: COLUMN nesma_validation_items.actual_value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.actual_value IS '实际值';


--
-- Name: COLUMN nesma_validation_items.deviation_level; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.deviation_level IS '偏差级别';


--
-- Name: COLUMN nesma_validation_items.validation_notes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.validation_notes IS '验证备注';


--
-- Name: COLUMN nesma_validation_items.validation_details; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.validation_details IS '验证详情';


--
-- Name: COLUMN nesma_validation_items.is_critical; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.is_critical IS '是否关键问题';


--
-- Name: COLUMN nesma_validation_items.resolution_status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.resolution_status IS '解决状态';


--
-- Name: COLUMN nesma_validation_items.resolution_notes; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_validation_items.resolution_notes IS '解决备注';


--
-- Name: nesma_validation_items_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_validation_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_validation_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_validation_items_id_seq OWNED BY public.nesma_validation_items.id;


--
-- Name: nesma_vector_indexes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_vector_indexes (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    index_name character varying(100) NOT NULL,
    target_table character varying(100) NOT NULL,
    column_name character varying(100) NOT NULL,
    dimensions bigint NOT NULL,
    index_type character varying(50),
    index_method character varying(50),
    configuration jsonb,
    is_built boolean DEFAULT false,
    build_time timestamp with time zone,
    status character varying(50) DEFAULT 'pending'::character varying
);


--
-- Name: COLUMN nesma_vector_indexes.index_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.index_name IS '索引名称';


--
-- Name: COLUMN nesma_vector_indexes.target_table; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.target_table IS '表名';


--
-- Name: COLUMN nesma_vector_indexes.column_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.column_name IS '列名';


--
-- Name: COLUMN nesma_vector_indexes.dimensions; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.dimensions IS '向量维度';


--
-- Name: COLUMN nesma_vector_indexes.index_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.index_type IS '索引类型:cosine,euclidean,dot_product';


--
-- Name: COLUMN nesma_vector_indexes.index_method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.index_method IS '索引方法:ivfflat,hnsw';


--
-- Name: COLUMN nesma_vector_indexes.configuration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.configuration IS '索引配置';


--
-- Name: COLUMN nesma_vector_indexes.is_built; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.is_built IS '是否已构建';


--
-- Name: COLUMN nesma_vector_indexes.build_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.build_time IS '构建时间';


--
-- Name: COLUMN nesma_vector_indexes.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_indexes.status IS '状态:pending,building,built,failed';


--
-- Name: nesma_vector_indexes_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_vector_indexes_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_vector_indexes_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_vector_indexes_id_seq OWNED BY public.nesma_vector_indexes.id;


--
-- Name: nesma_vector_memories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_vector_memories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    project_id bigint,
    content text NOT NULL,
    embedding text,
    metadata jsonb,
    relevance_score numeric DEFAULT 1,
    memory_type character varying(50),
    expires_at timestamp with time zone
);


--
-- Name: COLUMN nesma_vector_memories.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.user_id IS '用户ID';


--
-- Name: COLUMN nesma_vector_memories.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.project_id IS '项目ID';


--
-- Name: COLUMN nesma_vector_memories.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.content IS '记忆内容';


--
-- Name: COLUMN nesma_vector_memories.embedding; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.embedding IS '向量表示(JSON格式)';


--
-- Name: COLUMN nesma_vector_memories.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.metadata IS '元数据';


--
-- Name: COLUMN nesma_vector_memories.relevance_score; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.relevance_score IS '相关性评分';


--
-- Name: COLUMN nesma_vector_memories.memory_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.memory_type IS '记忆类型';


--
-- Name: COLUMN nesma_vector_memories.expires_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_memories.expires_at IS '过期时间';


--
-- Name: nesma_vector_memories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_vector_memories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_vector_memories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_vector_memories_id_seq OWNED BY public.nesma_vector_memories.id;


--
-- Name: nesma_vector_queries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_vector_queries (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    query_id character varying(100) NOT NULL,
    query_vector text NOT NULL,
    query_content text,
    result_count bigint,
    results jsonb,
    threshold numeric,
    query_time bigint,
    user_id bigint NOT NULL,
    project_id bigint,
    session_id character varying(255),
    query_type character varying(50),
    metadata jsonb
);


--
-- Name: COLUMN nesma_vector_queries.query_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.query_id IS '查询唯一标识';


--
-- Name: COLUMN nesma_vector_queries.query_vector; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.query_vector IS '查询向量';


--
-- Name: COLUMN nesma_vector_queries.query_content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.query_content IS '查询内容';


--
-- Name: COLUMN nesma_vector_queries.result_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.result_count IS '结果数量';


--
-- Name: COLUMN nesma_vector_queries.results; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.results IS '查询结果';


--
-- Name: COLUMN nesma_vector_queries.threshold; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.threshold IS '相似度阈值';


--
-- Name: COLUMN nesma_vector_queries.query_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.query_time IS '查询时间(毫秒)';


--
-- Name: COLUMN nesma_vector_queries.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.user_id IS '用户ID';


--
-- Name: COLUMN nesma_vector_queries.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.project_id IS '项目ID';


--
-- Name: COLUMN nesma_vector_queries.session_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.session_id IS '会话ID';


--
-- Name: COLUMN nesma_vector_queries.query_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.query_type IS '查询类型:similarity,keyword,hybrid';


--
-- Name: COLUMN nesma_vector_queries.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_queries.metadata IS '查询元数据';


--
-- Name: nesma_vector_queries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_vector_queries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_vector_queries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_vector_queries_id_seq OWNED BY public.nesma_vector_queries.id;


--
-- Name: nesma_vector_stores; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_vector_stores (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    vector_id character varying(100) NOT NULL,
    content text NOT NULL,
    embedding text,
    metadata jsonb,
    content_type character varying(50),
    source character varying(200),
    project_id bigint,
    user_id bigint NOT NULL,
    tags jsonb,
    hash character varying(64),
    version bigint DEFAULT 1,
    is_active boolean DEFAULT true,
    expires_at timestamp with time zone
);


--
-- Name: COLUMN nesma_vector_stores.vector_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.vector_id IS '向量唯一标识';


--
-- Name: COLUMN nesma_vector_stores.content; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.content IS '原始内容';


--
-- Name: COLUMN nesma_vector_stores.embedding; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.embedding IS '向量表示(JSON格式)';


--
-- Name: COLUMN nesma_vector_stores.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.metadata IS '元数据';


--
-- Name: COLUMN nesma_vector_stores.content_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.content_type IS '内容类型:text,document,code,knowledge';


--
-- Name: COLUMN nesma_vector_stores.source; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.source IS '来源';


--
-- Name: COLUMN nesma_vector_stores.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.project_id IS '项目ID';


--
-- Name: COLUMN nesma_vector_stores.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.user_id IS '用户ID';


--
-- Name: COLUMN nesma_vector_stores.tags; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.tags IS '标签';


--
-- Name: COLUMN nesma_vector_stores.hash; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.hash IS '内容哈希';


--
-- Name: COLUMN nesma_vector_stores.version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.version IS '版本号';


--
-- Name: COLUMN nesma_vector_stores.is_active; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.is_active IS '是否激活';


--
-- Name: COLUMN nesma_vector_stores.expires_at; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_vector_stores.expires_at IS '过期时间';


--
-- Name: nesma_vector_stores_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_vector_stores_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_vector_stores_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_vector_stores_id_seq OWNED BY public.nesma_vector_stores.id;


--
-- Name: nesma_workflow_executions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_workflow_executions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    execution_id character varying(100) NOT NULL,
    workflow_id character varying(100) NOT NULL,
    session_id character varying(255) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying,
    input_data jsonb,
    output_data jsonb,
    current_step character varying(100),
    execution_log jsonb,
    error_message text,
    start_time timestamp with time zone,
    end_time timestamp with time zone,
    total_duration bigint,
    user_id bigint,
    project_id bigint,
    metadata jsonb
);


--
-- Name: COLUMN nesma_workflow_executions.execution_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.execution_id IS '执行唯一标识';


--
-- Name: COLUMN nesma_workflow_executions.workflow_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.workflow_id IS '工作流ID';


--
-- Name: COLUMN nesma_workflow_executions.session_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.session_id IS '会话ID';


--
-- Name: COLUMN nesma_workflow_executions.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.status IS '执行状态:pending,running,completed,failed,cancelled';


--
-- Name: COLUMN nesma_workflow_executions.input_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.input_data IS '输入数据';


--
-- Name: COLUMN nesma_workflow_executions.output_data; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.output_data IS '输出数据';


--
-- Name: COLUMN nesma_workflow_executions.current_step; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.current_step IS '当前步骤';


--
-- Name: COLUMN nesma_workflow_executions.execution_log; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.execution_log IS '执行日志';


--
-- Name: COLUMN nesma_workflow_executions.error_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.error_message IS '错误信息';


--
-- Name: COLUMN nesma_workflow_executions.start_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.start_time IS '开始时间';


--
-- Name: COLUMN nesma_workflow_executions.end_time; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.end_time IS '结束时间';


--
-- Name: COLUMN nesma_workflow_executions.total_duration; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.total_duration IS '总耗时(毫秒)';


--
-- Name: COLUMN nesma_workflow_executions.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.user_id IS '用户ID';


--
-- Name: COLUMN nesma_workflow_executions.project_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.project_id IS '项目ID';


--
-- Name: COLUMN nesma_workflow_executions.metadata; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflow_executions.metadata IS '执行元数据';


--
-- Name: nesma_workflow_executions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_workflow_executions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_workflow_executions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_workflow_executions_id_seq OWNED BY public.nesma_workflow_executions.id;


--
-- Name: nesma_workflows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.nesma_workflows (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    workflow_id character varying(100) NOT NULL,
    name character varying(200) NOT NULL,
    description text,
    definition jsonb NOT NULL,
    status character varying(50) DEFAULT 'active'::character varying,
    version character varying(50) DEFAULT '1.0.0'::character varying,
    category character varying(100),
    tags jsonb,
    is_template boolean DEFAULT false,
    usage_count bigint DEFAULT 0,
    created_by bigint
);


--
-- Name: COLUMN nesma_workflows.workflow_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.workflow_id IS '工作流唯一标识';


--
-- Name: COLUMN nesma_workflows.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.name IS '工作流名称';


--
-- Name: COLUMN nesma_workflows.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.description IS '工作流描述';


--
-- Name: COLUMN nesma_workflows.definition; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.definition IS '工作流定义(JSON格式)';


--
-- Name: COLUMN nesma_workflows.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.status IS '工作流状态:active,inactive,draft';


--
-- Name: COLUMN nesma_workflows.version; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.version IS '工作流版本';


--
-- Name: COLUMN nesma_workflows.category; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.category IS '工作流分类';


--
-- Name: COLUMN nesma_workflows.tags; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.tags IS '标签';


--
-- Name: COLUMN nesma_workflows.is_template; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.is_template IS '是否为模板';


--
-- Name: COLUMN nesma_workflows.usage_count; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.usage_count IS '使用次数';


--
-- Name: COLUMN nesma_workflows.created_by; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.nesma_workflows.created_by IS '创建者ID';


--
-- Name: nesma_workflows_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.nesma_workflows_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: nesma_workflows_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.nesma_workflows_id_seq OWNED BY public.nesma_workflows.id;


--
-- Name: recommendation_adoptions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recommendation_adoptions (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    project_id bigint NOT NULL,
    recommendation_id text NOT NULL,
    title text NOT NULL,
    content text,
    category character varying(100),
    confidence numeric DEFAULT 0,
    knowledge_refs jsonb,
    user_feedback text,
    adopted_at timestamp with time zone,
    created_at timestamp with time zone
);


--
-- Name: recommendation_adoptions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.recommendation_adoptions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: recommendation_adoptions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.recommendation_adoptions_id_seq OWNED BY public.recommendation_adoptions.id;


--
-- Name: recommendation_histories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recommendation_histories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    recommendation_type character varying(50) NOT NULL,
    item_type character varying(50) NOT NULL,
    item_id bigint NOT NULL,
    score numeric NOT NULL,
    reason text,
    algorithm character varying(50),
    context jsonb,
    is_clicked boolean DEFAULT false,
    is_interacted boolean DEFAULT false,
    feedback_score numeric,
    feedback_comment text
);


--
-- Name: recommendation_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.recommendation_histories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: recommendation_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.recommendation_histories_id_seq OWNED BY public.recommendation_histories.id;


--
-- Name: requirement_analysis_results; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.requirement_analysis_results (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    project_id bigint NOT NULL,
    requirement_id bigint NOT NULL,
    analysis_type character varying(50) NOT NULL,
    input_text text NOT NULL,
    result json NOT NULL,
    suggestions jsonb,
    quality_score numeric DEFAULT 0,
    processing_time bigint DEFAULT 0,
    model_name character varying(50),
    version character varying(20) DEFAULT '1.0'::character varying,
    status character varying(20) DEFAULT 'completed'::character varying,
    error_message text,
    original_text text NOT NULL,
    score numeric DEFAULT 0,
    ai_model character varying(50)
);


--
-- Name: requirement_analysis_results_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.requirement_analysis_results_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: requirement_analysis_results_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.requirement_analysis_results_id_seq OWNED BY public.requirement_analysis_results.id;


--
-- Name: session_memory_stats; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.session_memory_stats AS
 SELECT user_id,
    count(*) AS total_sessions,
    count(
        CASE
            WHEN (is_active = true) THEN 1
            ELSE NULL::integer
        END) AS active_sessions,
    avg(
        CASE
            WHEN (conversation_history IS NOT NULL) THEN jsonb_array_length(conversation_history)
            ELSE 0
        END) AS avg_conversation_length,
    min(created_at) AS first_session,
    max(updated_at) AS last_activity
   FROM public.nesma_session_memories
  GROUP BY user_id;


--
-- Name: sys_apis; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_apis (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    path text,
    description text,
    api_group text,
    method text DEFAULT 'POST'::text
);


--
-- Name: COLUMN sys_apis.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.path IS 'api路径';


--
-- Name: COLUMN sys_apis.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.description IS 'api中文描述';


--
-- Name: COLUMN sys_apis.api_group; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.api_group IS 'api组';


--
-- Name: COLUMN sys_apis.method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_apis.method IS '方法';


--
-- Name: sys_apis_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_apis_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_apis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_apis_id_seq OWNED BY public.sys_apis.id;


--
-- Name: sys_authorities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_authorities (
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    authority_id bigint NOT NULL,
    authority_name text,
    parent_id bigint,
    default_router text DEFAULT 'dashboard'::text
);


--
-- Name: COLUMN sys_authorities.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.authority_id IS '角色ID';


--
-- Name: COLUMN sys_authorities.authority_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.authority_name IS '角色名';


--
-- Name: COLUMN sys_authorities.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.parent_id IS '父角色ID';


--
-- Name: COLUMN sys_authorities.default_router; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authorities.default_router IS '默认菜单';


--
-- Name: sys_authorities_authority_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_authorities_authority_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_authorities_authority_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_authorities_authority_id_seq OWNED BY public.sys_authorities.authority_id;


--
-- Name: sys_authority_btns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_authority_btns (
    authority_id bigint,
    sys_menu_id bigint,
    sys_base_menu_btn_id bigint
);


--
-- Name: COLUMN sys_authority_btns.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_btns.authority_id IS '角色ID';


--
-- Name: COLUMN sys_authority_btns.sys_menu_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_btns.sys_menu_id IS '菜单ID';


--
-- Name: COLUMN sys_authority_btns.sys_base_menu_btn_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_btns.sys_base_menu_btn_id IS '菜单按钮ID';


--
-- Name: sys_authority_menus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_authority_menus (
    sys_base_menu_id bigint NOT NULL,
    sys_authority_authority_id bigint NOT NULL
);


--
-- Name: COLUMN sys_authority_menus.sys_authority_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_authority_menus.sys_authority_authority_id IS '角色ID';


--
-- Name: sys_auto_code_histories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_auto_code_histories (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    table_name text,
    package text,
    request text,
    struct_name text,
    abbreviation text,
    business_db text,
    description text,
    templates text,
    "Injections" text,
    flag bigint,
    api_ids text,
    menu_id bigint,
    export_template_id bigint,
    package_id bigint
);


--
-- Name: COLUMN sys_auto_code_histories.table_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.table_name IS '表名';


--
-- Name: COLUMN sys_auto_code_histories.package; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.package IS '模块名/插件名';


--
-- Name: COLUMN sys_auto_code_histories.request; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.request IS '前端传入的结构化信息';


--
-- Name: COLUMN sys_auto_code_histories.struct_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.struct_name IS '结构体名称';


--
-- Name: COLUMN sys_auto_code_histories.abbreviation; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.abbreviation IS '结构体名称缩写';


--
-- Name: COLUMN sys_auto_code_histories.business_db; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.business_db IS '业务库';


--
-- Name: COLUMN sys_auto_code_histories.description; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.description IS 'Struct中文名称';


--
-- Name: COLUMN sys_auto_code_histories.templates; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.templates IS '模板信息';


--
-- Name: COLUMN sys_auto_code_histories."Injections"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories."Injections" IS '注入路径';


--
-- Name: COLUMN sys_auto_code_histories.flag; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.flag IS '[0:创建,1:回滚]';


--
-- Name: COLUMN sys_auto_code_histories.api_ids; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.api_ids IS 'api表注册内容';


--
-- Name: COLUMN sys_auto_code_histories.menu_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.menu_id IS '菜单ID';


--
-- Name: COLUMN sys_auto_code_histories.export_template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.export_template_id IS '导出模板ID';


--
-- Name: COLUMN sys_auto_code_histories.package_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_histories.package_id IS '包ID';


--
-- Name: sys_auto_code_histories_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_auto_code_histories_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_auto_code_histories_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_auto_code_histories_id_seq OWNED BY public.sys_auto_code_histories.id;


--
-- Name: sys_auto_code_packages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_auto_code_packages (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    "desc" text,
    label text,
    template text,
    package_name text,
    module text
);


--
-- Name: COLUMN sys_auto_code_packages."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages."desc" IS '描述';


--
-- Name: COLUMN sys_auto_code_packages.label; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages.label IS '展示名';


--
-- Name: COLUMN sys_auto_code_packages.template; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages.template IS '模版';


--
-- Name: COLUMN sys_auto_code_packages.package_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_auto_code_packages.package_name IS '包名';


--
-- Name: sys_auto_code_packages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_auto_code_packages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_auto_code_packages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_auto_code_packages_id_seq OWNED BY public.sys_auto_code_packages.id;


--
-- Name: sys_base_menu_btns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_base_menu_btns (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    "desc" text,
    sys_base_menu_id bigint
);


--
-- Name: COLUMN sys_base_menu_btns.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_btns.name IS '按钮关键key';


--
-- Name: COLUMN sys_base_menu_btns.sys_base_menu_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_btns.sys_base_menu_id IS '菜单ID';


--
-- Name: sys_base_menu_btns_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_base_menu_btns_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menu_btns_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_base_menu_btns_id_seq OWNED BY public.sys_base_menu_btns.id;


--
-- Name: sys_base_menu_parameters; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_base_menu_parameters (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    sys_base_menu_id bigint,
    type text,
    key text,
    value text
);


--
-- Name: COLUMN sys_base_menu_parameters.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_parameters.type IS '地址栏携带参数为params还是query';


--
-- Name: COLUMN sys_base_menu_parameters.key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_parameters.key IS '地址栏携带参数的key';


--
-- Name: COLUMN sys_base_menu_parameters.value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menu_parameters.value IS '地址栏携带参数的值';


--
-- Name: sys_base_menu_parameters_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_base_menu_parameters_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menu_parameters_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_base_menu_parameters_id_seq OWNED BY public.sys_base_menu_parameters.id;


--
-- Name: sys_base_menus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_base_menus (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    menu_level bigint,
    parent_id bigint,
    path text,
    name text,
    hidden boolean,
    component text,
    sort bigint,
    active_name text,
    keep_alive boolean,
    default_menu boolean,
    title text,
    icon text,
    close_tab boolean,
    transition_type text
);


--
-- Name: COLUMN sys_base_menus.parent_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.parent_id IS '父菜单ID';


--
-- Name: COLUMN sys_base_menus.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.path IS '路由path';


--
-- Name: COLUMN sys_base_menus.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.name IS '路由name';


--
-- Name: COLUMN sys_base_menus.hidden; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.hidden IS '是否在列表隐藏';


--
-- Name: COLUMN sys_base_menus.component; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.component IS '对应前端文件路径';


--
-- Name: COLUMN sys_base_menus.sort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.sort IS '排序标记';


--
-- Name: COLUMN sys_base_menus.active_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.active_name IS '高亮菜单';


--
-- Name: COLUMN sys_base_menus.keep_alive; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.keep_alive IS '是否缓存';


--
-- Name: COLUMN sys_base_menus.default_menu; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.default_menu IS '是否是基础路由（开发中）';


--
-- Name: COLUMN sys_base_menus.title; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.title IS '菜单名';


--
-- Name: COLUMN sys_base_menus.icon; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.icon IS '菜单图标';


--
-- Name: COLUMN sys_base_menus.close_tab; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.close_tab IS '自动关闭tab';


--
-- Name: COLUMN sys_base_menus.transition_type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_base_menus.transition_type IS '路由切换动画';


--
-- Name: sys_base_menus_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_base_menus_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_base_menus_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_base_menus_id_seq OWNED BY public.sys_base_menus.id;


--
-- Name: sys_data_authority_id; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_data_authority_id (
    sys_authority_authority_id bigint NOT NULL,
    data_authority_id_authority_id bigint NOT NULL
);


--
-- Name: COLUMN sys_data_authority_id.sys_authority_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_data_authority_id.sys_authority_authority_id IS '角色ID';


--
-- Name: COLUMN sys_data_authority_id.data_authority_id_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_data_authority_id.data_authority_id_authority_id IS '角色ID';


--
-- Name: sys_dictionaries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_dictionaries (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    type text,
    status boolean,
    "desc" text
);


--
-- Name: COLUMN sys_dictionaries.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries.name IS '字典名（中）';


--
-- Name: COLUMN sys_dictionaries.type; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries.type IS '字典名（英）';


--
-- Name: COLUMN sys_dictionaries.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries.status IS '状态';


--
-- Name: COLUMN sys_dictionaries."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionaries."desc" IS '描述';


--
-- Name: sys_dictionaries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_dictionaries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_dictionaries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_dictionaries_id_seq OWNED BY public.sys_dictionaries.id;


--
-- Name: sys_dictionary_details; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_dictionary_details (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    label text,
    value text,
    extend text,
    status boolean,
    sort bigint,
    sys_dictionary_id bigint
);


--
-- Name: COLUMN sys_dictionary_details.label; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.label IS '展示值';


--
-- Name: COLUMN sys_dictionary_details.value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.value IS '字典值';


--
-- Name: COLUMN sys_dictionary_details.extend; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.extend IS '扩展值';


--
-- Name: COLUMN sys_dictionary_details.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.status IS '启用状态';


--
-- Name: COLUMN sys_dictionary_details.sort; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.sort IS '排序标记';


--
-- Name: COLUMN sys_dictionary_details.sys_dictionary_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_dictionary_details.sys_dictionary_id IS '关联标记';


--
-- Name: sys_dictionary_details_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_dictionary_details_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_dictionary_details_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_dictionary_details_id_seq OWNED BY public.sys_dictionary_details.id;


--
-- Name: sys_export_template_condition; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_export_template_condition (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    template_id text,
    "from" text,
    "column" text,
    operator text
);


--
-- Name: COLUMN sys_export_template_condition.template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition.template_id IS '模板标识';


--
-- Name: COLUMN sys_export_template_condition."from"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition."from" IS '条件取的key';


--
-- Name: COLUMN sys_export_template_condition."column"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition."column" IS '作为查询条件的字段';


--
-- Name: COLUMN sys_export_template_condition.operator; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_condition.operator IS '操作符';


--
-- Name: sys_export_template_condition_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_export_template_condition_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_template_condition_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_export_template_condition_id_seq OWNED BY public.sys_export_template_condition.id;


--
-- Name: sys_export_template_join; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_export_template_join (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    template_id text,
    joins text,
    "table" text,
    "on" text
);


--
-- Name: COLUMN sys_export_template_join.template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join.template_id IS '模板标识';


--
-- Name: COLUMN sys_export_template_join.joins; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join.joins IS '关联';


--
-- Name: COLUMN sys_export_template_join."table"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join."table" IS '关联表';


--
-- Name: COLUMN sys_export_template_join."on"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_template_join."on" IS '关联条件';


--
-- Name: sys_export_template_join_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_export_template_join_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_template_join_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_export_template_join_id_seq OWNED BY public.sys_export_template_join.id;


--
-- Name: sys_export_templates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_export_templates (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    db_name text,
    name text,
    table_name text,
    template_id text,
    template_info text,
    "limit" bigint,
    "order" text
);


--
-- Name: COLUMN sys_export_templates.db_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.db_name IS '数据库名称';


--
-- Name: COLUMN sys_export_templates.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.name IS '模板名称';


--
-- Name: COLUMN sys_export_templates.table_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.table_name IS '表名称';


--
-- Name: COLUMN sys_export_templates.template_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates.template_id IS '模板标识';


--
-- Name: COLUMN sys_export_templates."limit"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates."limit" IS '导出限制';


--
-- Name: COLUMN sys_export_templates."order"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_export_templates."order" IS '排序';


--
-- Name: sys_export_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_export_templates_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_export_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_export_templates_id_seq OWNED BY public.sys_export_templates.id;


--
-- Name: sys_ignore_apis; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_ignore_apis (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    path text,
    method text DEFAULT 'POST'::text
);


--
-- Name: COLUMN sys_ignore_apis.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ignore_apis.path IS 'api路径';


--
-- Name: COLUMN sys_ignore_apis.method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_ignore_apis.method IS '方法';


--
-- Name: sys_ignore_apis_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_ignore_apis_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_ignore_apis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_ignore_apis_id_seq OWNED BY public.sys_ignore_apis.id;


--
-- Name: sys_operation_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_operation_records (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    ip text,
    method text,
    path text,
    status bigint,
    latency bigint,
    agent text,
    error_message text,
    body text,
    resp text,
    user_id bigint
);


--
-- Name: COLUMN sys_operation_records.ip; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.ip IS '请求ip';


--
-- Name: COLUMN sys_operation_records.method; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.method IS '请求方法';


--
-- Name: COLUMN sys_operation_records.path; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.path IS '请求路径';


--
-- Name: COLUMN sys_operation_records.status; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.status IS '请求状态';


--
-- Name: COLUMN sys_operation_records.latency; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.latency IS '延迟';


--
-- Name: COLUMN sys_operation_records.agent; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.agent IS '代理';


--
-- Name: COLUMN sys_operation_records.error_message; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.error_message IS '错误信息';


--
-- Name: COLUMN sys_operation_records.body; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.body IS '请求Body';


--
-- Name: COLUMN sys_operation_records.resp; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.resp IS '响应Body';


--
-- Name: COLUMN sys_operation_records.user_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_operation_records.user_id IS '用户id';


--
-- Name: sys_operation_records_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_operation_records_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_operation_records_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_operation_records_id_seq OWNED BY public.sys_operation_records.id;


--
-- Name: sys_params; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_params (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text,
    key text,
    value text,
    "desc" text
);


--
-- Name: COLUMN sys_params.name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params.name IS '参数名称';


--
-- Name: COLUMN sys_params.key; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params.key IS '参数键';


--
-- Name: COLUMN sys_params.value; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params.value IS '参数值';


--
-- Name: COLUMN sys_params."desc"; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_params."desc" IS '参数说明';


--
-- Name: sys_params_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_params_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_params_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_params_id_seq OWNED BY public.sys_params.id;


--
-- Name: sys_user_authority; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_user_authority (
    sys_user_id bigint NOT NULL,
    sys_authority_authority_id bigint NOT NULL
);


--
-- Name: COLUMN sys_user_authority.sys_authority_authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_user_authority.sys_authority_authority_id IS '角色ID';


--
-- Name: sys_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sys_users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    uuid text,
    username text,
    password text,
    nick_name text DEFAULT '系统用户'::text,
    header_img text DEFAULT 'https://qmplusimg.henrongyi.top/gva_header.jpg'::text,
    authority_id bigint DEFAULT 888,
    phone text,
    email text,
    enable bigint DEFAULT 1,
    origin_setting text
);


--
-- Name: COLUMN sys_users.uuid; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.uuid IS '用户UUID';


--
-- Name: COLUMN sys_users.username; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.username IS '用户登录名';


--
-- Name: COLUMN sys_users.password; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.password IS '用户登录密码';


--
-- Name: COLUMN sys_users.nick_name; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.nick_name IS '用户昵称';


--
-- Name: COLUMN sys_users.header_img; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.header_img IS '用户头像';


--
-- Name: COLUMN sys_users.authority_id; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.authority_id IS '用户角色ID';


--
-- Name: COLUMN sys_users.phone; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.phone IS '用户手机号';


--
-- Name: COLUMN sys_users.email; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.email IS '用户邮箱';


--
-- Name: COLUMN sys_users.enable; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.enable IS '用户是否被冻结 1正常 2冻结';


--
-- Name: COLUMN sys_users.origin_setting; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.sys_users.origin_setting IS '配置';


--
-- Name: sys_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.sys_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: sys_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.sys_users_id_seq OWNED BY public.sys_users.id;


--
-- Name: user_behaviors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_behaviors (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    user_id bigint NOT NULL,
    action character varying(50) NOT NULL,
    item_type character varying(50) NOT NULL,
    item_id bigint NOT NULL,
    session_id character varying(255),
    context jsonb,
    duration bigint DEFAULT 0,
    device_type character varying(50),
    user_agent character varying(500),
    ip_address character varying(50)
);


--
-- Name: user_behaviors_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_behaviors_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_behaviors_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_behaviors_id_seq OWNED BY public.user_behaviors.id;


--
-- Name: vector_index_stats; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vector_index_stats AS
 SELECT index_name,
    target_table,
    dimensions,
    index_type,
    is_built,
    status,
    build_time,
        CASE
            WHEN (build_time IS NOT NULL) THEN (EXTRACT(epoch FROM (build_time - created_at)) * (1000)::numeric)
            ELSE NULL::numeric
        END AS build_duration_ms
   FROM public.nesma_vector_indexes
  ORDER BY created_at DESC;


--
-- Name: vector_store_stats; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.vector_store_stats AS
 SELECT content_type,
    count(*) AS total_count,
    count(
        CASE
            WHEN (is_active = true) THEN 1
            ELSE NULL::integer
        END) AS active_count,
    avg(length(content)) AS avg_content_length,
    min(created_at) AS first_created,
    max(updated_at) AS last_updated
   FROM public.nesma_vector_stores
  GROUP BY content_type;


--
-- Name: ai_business_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_business_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_business_analyses_id_seq'::regclass);


--
-- Name: ai_compliance_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_compliance_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_compliance_analyses_id_seq'::regclass);


--
-- Name: ai_functional_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_functional_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_functional_analyses_id_seq'::regclass);


--
-- Name: ai_project_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_project_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_project_analyses_id_seq'::regclass);


--
-- Name: ai_quality_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_quality_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_quality_analyses_id_seq'::regclass);


--
-- Name: ai_recommendation_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_recommendation_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_recommendation_analyses_id_seq'::regclass);


--
-- Name: ai_risk_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_risk_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_risk_analyses_id_seq'::regclass);


--
-- Name: ai_service_metrics id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_service_metrics ALTER COLUMN id SET DEFAULT nextval('public.ai_service_metrics_id_seq'::regclass);


--
-- Name: ai_technical_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_technical_analyses ALTER COLUMN id SET DEFAULT nextval('public.ai_technical_analyses_id_seq'::regclass);


--
-- Name: casbin_rule id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casbin_rule ALTER COLUMN id SET DEFAULT nextval('public.casbin_rule_id_seq'::regclass);


--
-- Name: chat_messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_messages ALTER COLUMN id SET DEFAULT nextval('public.chat_messages_id_seq'::regclass);


--
-- Name: chat_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_sessions ALTER COLUMN id SET DEFAULT nextval('public.chat_sessions_id_seq'::regclass);


--
-- Name: exa_attachment_category id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_attachment_category ALTER COLUMN id SET DEFAULT nextval('public.exa_attachment_category_id_seq'::regclass);


--
-- Name: exa_customers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_customers ALTER COLUMN id SET DEFAULT nextval('public.exa_customers_id_seq'::regclass);


--
-- Name: exa_file_chunks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_chunks ALTER COLUMN id SET DEFAULT nextval('public.exa_file_chunks_id_seq'::regclass);


--
-- Name: exa_file_upload_and_downloads id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_upload_and_downloads ALTER COLUMN id SET DEFAULT nextval('public.exa_file_upload_and_downloads_id_seq'::regclass);


--
-- Name: exa_files id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_files ALTER COLUMN id SET DEFAULT nextval('public.exa_files_id_seq'::regclass);


--
-- Name: gva_announcements_info id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gva_announcements_info ALTER COLUMN id SET DEFAULT nextval('public.gva_announcements_info_id_seq'::regclass);


--
-- Name: jwt_blacklists id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jwt_blacklists ALTER COLUMN id SET DEFAULT nextval('public.jwt_blacklists_id_seq'::regclass);


--
-- Name: nesma_agent_interactions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_interactions ALTER COLUMN id SET DEFAULT nextval('public.nesma_agent_interactions_id_seq'::regclass);


--
-- Name: nesma_agent_messages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_messages ALTER COLUMN id SET DEFAULT nextval('public.nesma_agent_messages_id_seq'::regclass);


--
-- Name: nesma_agent_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_tasks ALTER COLUMN id SET DEFAULT nextval('public.nesma_agent_tasks_id_seq'::regclass);


--
-- Name: nesma_agents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agents ALTER COLUMN id SET DEFAULT nextval('public.nesma_agents_id_seq'::regclass);


--
-- Name: nesma_analysis_histories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_analysis_histories ALTER COLUMN id SET DEFAULT nextval('public.nesma_analysis_histories_id_seq'::regclass);


--
-- Name: nesma_case_studies id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_case_studies ALTER COLUMN id SET DEFAULT nextval('public.nesma_case_studies_id_seq'::regclass);


--
-- Name: nesma_compact_memories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_compact_memories ALTER COLUMN id SET DEFAULT nextval('public.nesma_compact_memories_id_seq'::regclass);


--
-- Name: nesma_complexity_metrics id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_complexity_metrics ALTER COLUMN id SET DEFAULT nextval('public.nesma_complexity_metrics_id_seq'::regclass);


--
-- Name: nesma_doc_batches id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_doc_batches ALTER COLUMN id SET DEFAULT nextval('public.nesma_doc_batches_id_seq'::regclass);


--
-- Name: nesma_doc_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_doc_templates ALTER COLUMN id SET DEFAULT nextval('public.nesma_doc_templates_id_seq'::regclass);


--
-- Name: nesma_documents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_documents ALTER COLUMN id SET DEFAULT nextval('public.nesma_documents_id_seq'::regclass);


--
-- Name: nesma_embeddings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_embeddings ALTER COLUMN id SET DEFAULT nextval('public.nesma_embeddings_id_seq'::regclass);


--
-- Name: nesma_evaluation_factors id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_evaluation_factors ALTER COLUMN id SET DEFAULT nextval('public.nesma_evaluation_factors_id_seq'::regclass);


--
-- Name: nesma_evaluation_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_evaluation_history ALTER COLUMN id SET DEFAULT nextval('public.nesma_evaluation_history_id_seq'::regclass);


--
-- Name: nesma_evaluations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_evaluations ALTER COLUMN id SET DEFAULT nextval('public.nesma_evaluations_id_seq'::regclass);


--
-- Name: nesma_function_points id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_function_points ALTER COLUMN id SET DEFAULT nextval('public.nesma_function_points_id_seq'::regclass);


--
-- Name: nesma_knowledge_entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_knowledge_entries ALTER COLUMN id SET DEFAULT nextval('public.nesma_knowledge_entries_id_seq'::regclass);


--
-- Name: nesma_knowledge_rules id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_knowledge_rules ALTER COLUMN id SET DEFAULT nextval('public.nesma_knowledge_rules_id_seq'::regclass);


--
-- Name: nesma_nesma_evaluations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_nesma_evaluations ALTER COLUMN id SET DEFAULT nextval('public.nesma_nesma_evaluations_id_seq'::regclass);


--
-- Name: nesma_optimization_suggestions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_optimization_suggestions ALTER COLUMN id SET DEFAULT nextval('public.nesma_optimization_suggestions_id_seq'::regclass);


--
-- Name: nesma_project_analyses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_project_analyses ALTER COLUMN id SET DEFAULT nextval('public.nesma_project_analyses_id_seq'::regclass);


--
-- Name: nesma_project_cycles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_project_cycles ALTER COLUMN id SET DEFAULT nextval('public.nesma_project_cycles_id_seq'::regclass);


--
-- Name: nesma_projects id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_projects ALTER COLUMN id SET DEFAULT nextval('public.nesma_projects_id_seq'::regclass);


--
-- Name: nesma_requirement_analysis id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_analysis ALTER COLUMN id SET DEFAULT nextval('public.nesma_requirement_analysis_id_seq'::regclass);


--
-- Name: nesma_requirement_analysis_tasks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_analysis_tasks ALTER COLUMN id SET DEFAULT nextval('public.nesma_requirement_analysis_tasks_id_seq'::regclass);


--
-- Name: nesma_requirement_optimizations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_optimizations ALTER COLUMN id SET DEFAULT nextval('public.nesma_requirement_optimizations_id_seq'::regclass);


--
-- Name: nesma_requirement_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_versions ALTER COLUMN id SET DEFAULT nextval('public.nesma_requirement_versions_id_seq'::regclass);


--
-- Name: nesma_requirements id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirements ALTER COLUMN id SET DEFAULT nextval('public.nesma_requirements_id_seq'::regclass);


--
-- Name: nesma_session_memories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_session_memories ALTER COLUMN id SET DEFAULT nextval('public.nesma_session_memories_id_seq'::regclass);


--
-- Name: nesma_system_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_system_activities ALTER COLUMN id SET DEFAULT nextval('public.nesma_system_activities_id_seq'::regclass);


--
-- Name: nesma_validation_items id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_validation_items ALTER COLUMN id SET DEFAULT nextval('public.nesma_validation_items_id_seq'::regclass);


--
-- Name: nesma_vector_indexes id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_indexes ALTER COLUMN id SET DEFAULT nextval('public.nesma_vector_indexes_id_seq'::regclass);


--
-- Name: nesma_vector_memories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_memories ALTER COLUMN id SET DEFAULT nextval('public.nesma_vector_memories_id_seq'::regclass);


--
-- Name: nesma_vector_queries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_queries ALTER COLUMN id SET DEFAULT nextval('public.nesma_vector_queries_id_seq'::regclass);


--
-- Name: nesma_vector_stores id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_stores ALTER COLUMN id SET DEFAULT nextval('public.nesma_vector_stores_id_seq'::regclass);


--
-- Name: nesma_workflow_executions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_workflow_executions ALTER COLUMN id SET DEFAULT nextval('public.nesma_workflow_executions_id_seq'::regclass);


--
-- Name: nesma_workflows id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_workflows ALTER COLUMN id SET DEFAULT nextval('public.nesma_workflows_id_seq'::regclass);


--
-- Name: recommendation_adoptions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recommendation_adoptions ALTER COLUMN id SET DEFAULT nextval('public.recommendation_adoptions_id_seq'::regclass);


--
-- Name: recommendation_histories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recommendation_histories ALTER COLUMN id SET DEFAULT nextval('public.recommendation_histories_id_seq'::regclass);


--
-- Name: requirement_analysis_results id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.requirement_analysis_results ALTER COLUMN id SET DEFAULT nextval('public.requirement_analysis_results_id_seq'::regclass);


--
-- Name: sys_apis id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_apis ALTER COLUMN id SET DEFAULT nextval('public.sys_apis_id_seq'::regclass);


--
-- Name: sys_authorities authority_id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authorities ALTER COLUMN authority_id SET DEFAULT nextval('public.sys_authorities_authority_id_seq'::regclass);


--
-- Name: sys_auto_code_histories id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_histories ALTER COLUMN id SET DEFAULT nextval('public.sys_auto_code_histories_id_seq'::regclass);


--
-- Name: sys_auto_code_packages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_packages ALTER COLUMN id SET DEFAULT nextval('public.sys_auto_code_packages_id_seq'::regclass);


--
-- Name: sys_base_menu_btns id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_btns ALTER COLUMN id SET DEFAULT nextval('public.sys_base_menu_btns_id_seq'::regclass);


--
-- Name: sys_base_menu_parameters id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_parameters ALTER COLUMN id SET DEFAULT nextval('public.sys_base_menu_parameters_id_seq'::regclass);


--
-- Name: sys_base_menus id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menus ALTER COLUMN id SET DEFAULT nextval('public.sys_base_menus_id_seq'::regclass);


--
-- Name: sys_dictionaries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionaries ALTER COLUMN id SET DEFAULT nextval('public.sys_dictionaries_id_seq'::regclass);


--
-- Name: sys_dictionary_details id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionary_details ALTER COLUMN id SET DEFAULT nextval('public.sys_dictionary_details_id_seq'::regclass);


--
-- Name: sys_export_template_condition id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_condition ALTER COLUMN id SET DEFAULT nextval('public.sys_export_template_condition_id_seq'::regclass);


--
-- Name: sys_export_template_join id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_join ALTER COLUMN id SET DEFAULT nextval('public.sys_export_template_join_id_seq'::regclass);


--
-- Name: sys_export_templates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_templates ALTER COLUMN id SET DEFAULT nextval('public.sys_export_templates_id_seq'::regclass);


--
-- Name: sys_ignore_apis id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_ignore_apis ALTER COLUMN id SET DEFAULT nextval('public.sys_ignore_apis_id_seq'::regclass);


--
-- Name: sys_operation_records id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_operation_records ALTER COLUMN id SET DEFAULT nextval('public.sys_operation_records_id_seq'::regclass);


--
-- Name: sys_params id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_params ALTER COLUMN id SET DEFAULT nextval('public.sys_params_id_seq'::regclass);


--
-- Name: sys_users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_users ALTER COLUMN id SET DEFAULT nextval('public.sys_users_id_seq'::regclass);


--
-- Name: user_behaviors id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_behaviors ALTER COLUMN id SET DEFAULT nextval('public.user_behaviors_id_seq'::regclass);


--
-- Data for Name: ai_business_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_business_analyses (id, created_at, updated_at, deleted_at, analysis_id, business_value, roi_estimation, strategic_alignment, user_experience_score, usability_analysis, accessibility_analysis, process_efficiency, process_automation, business_logic_complexity, market_fit, competitive_advantage, innovation_level, business_risk, regulatory_compliance, ai_business_assessment, business_recommendations, confidence_score) FROM stdin;
\.


--
-- Data for Name: ai_compliance_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_compliance_analyses (id, created_at, updated_at, deleted_at, analysis_id, compliance_standard, standard_version, overall_compliance, compliance_level, compliance_items, non_compliance_items, partial_compliance_items, compliance_gaps, remediation_plan, ai_compliance_assessment, compliance_recommendations, confidence_score) FROM stdin;
\.


--
-- Data for Name: ai_functional_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_functional_analyses (id, created_at, updated_at, deleted_at, analysis_id, total_function_points, data_function_points, transactional_function_points, ilf_count, eif_count, ei_count, eo_count, eq_count, low_complexity_count, medium_complexity_count, high_complexity_count, functional_coverage, functional_completeness, functional_consistency, function_type_distribution, complexity_analysis, functional_gaps, ai_assessment, confidence_score, quality_indicators) FROM stdin;
\.


--
-- Data for Name: ai_project_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_project_analyses (id, created_at, updated_at, deleted_at, project_id, evaluation_id, analysis_version, analysis_type, status, progress, start_time, completion_time, total_requirements, processed_requirements, requirement_levels, project_context, ai_model, max_tokens, temperature, batch_size, overall_score, complexity_level, risk_level, recommendation_level) FROM stdin;
\.


--
-- Data for Name: ai_quality_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_quality_analyses (id, created_at, updated_at, deleted_at, analysis_id, overall_quality, requirement_quality, design_quality, correctness, completeness, consistency, clarity, traceability, maintainability, modularity, reusability, testability, critical_issues, quality_gaps, improvement_areas, ai_quality_assessment, quality_recommendations, confidence_score) FROM stdin;
\.


--
-- Data for Name: ai_recommendation_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_recommendation_analyses (id, created_at, updated_at, deleted_at, analysis_id, recommendation_type, priority, impact_level, title, description, rationale, implementation_steps, estimated_effort, expected_benefit, dependencies, constraints, risks, ai_generated_recommendation, confidence_score, urgency) FROM stdin;
\.


--
-- Data for Name: ai_risk_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_risk_analyses (id, created_at, updated_at, deleted_at, analysis_id, overall_risk, risk_level, technical_risk, implementation_risk, integration_risk, schedule_risk, budget_risk, resource_risk, business_risk, market_risk, compliance_risk, identified_risks, risk_mitigation_plans, contingency_plans, ai_risk_assessment, risk_recommendations, confidence_score) FROM stdin;
\.


--
-- Data for Name: ai_service_metrics; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_service_metrics (id, created_at, updated_at, deleted_at, service_name, metric_type, metric_value, request_count, error_count, total_tokens, prompt_tokens, completion_tokens, average_latency, p95_latency, p99_latency, error_messages, metadata, recorded_at, user_id, model_name, request_type, status, response_time, token_usage, error_message, request_payload) FROM stdin;
\.


--
-- Data for Name: ai_technical_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.ai_technical_analyses (id, created_at, updated_at, deleted_at, analysis_id, architecture_type, technology_stack, integration_complexity, performance_requirements, scalability_analysis, reliability_analysis, technical_debt, security_considerations, maintenance_complexity, development_effort, testing_effort, deployment_complexity, technical_feasibility, implementation_risk, technical_innovation, ai_technical_assessment, technical_recommendations, confidence_score) FROM stdin;
\.


--
-- Data for Name: casbin_rule; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.casbin_rule (id, ptype, v0, v1, v2, v3, v4, v5) FROM stdin;
6263	p	999	/ai-optimized	GET			
6264	p	999	/ai-status	PUT			
6265	p	999	/compare-versions	POST			
6266	p	999	/ai-analysis-stats	GET			
6267	p	999	/:id/ai-analysis	GET			
6268	p	999	/nesma/generator/level4/validate	POST			
6269	p	999	/nesma/generator/level4/stats/:cycleId	GET			
6270	p	999	/nesma/generator/level4/history/:cycleId	GET			
6271	p	999	/nesma/generator/level4/batch-create	POST			
6272	p	999	/nesma/generator/level4/create	POST			
6273	p	999	/nesma/generator/level4/generate	POST			
6274	p	999	/nesma/generator/level4/stats/{cycleId}	GET			
6275	p	999	/nesma/analysis/level3/stats/:cycleId	GET			
6276	p	999	/nesma/analysis/level3/history/:cycleId	GET			
6277	p	999	/nesma/analysis/level3/batch-create-expansions	POST			
6278	p	999	/nesma/analysis/level3/batch-apply-optimizations	POST			
6279	p	999	/nesma/analysis/level3/create-expansion	POST			
6280	p	999	/nesma/analysis/level3/apply-optimization	POST			
6281	p	999	/nesma/analysis/level3/analyze	POST			
6282	p	999	/monitoring/report/export	GET			
6283	p	999	/monitoring/cost/analysis	GET			
6284	p	999	/monitoring/response-time/distribution	GET			
6285	p	999	/monitoring/model/comparison	GET			
6286	p	999	/monitoring/system/resources	GET			
6287	p	999	/monitoring/health	GET			
6288	p	999	/monitoring/metrics	GET			
6289	p	999	/monitoring/time-series	GET			
6290	p	999	/monitoring/token/stats	GET			
6291	p	999	/monitoring/error/analysis	GET			
6292	p	999	/monitoring/service/stats	GET			
6293	p	999	/nesma/generator/mermaid/preview	POST			
6294	p	999	/nesma/generator/mermaid/validate	POST			
6295	p	999	/nesma/generator/mermaid/stats/{cycleId}	GET			
6296	p	999	/nesma/generator/mermaid/history/{cycleId}	GET			
6297	p	999	/nesma/generator/mermaid/batch-generate	POST			
6298	p	999	/nesma/generator/mermaid/apply	POST			
6299	p	999	/nesma/generator/mermaid/generate	POST			
6300	p	999	/nesma/generator/description/validate	POST			
6301	p	999	/nesma/generator/description/preview	POST			
6302	p	999	/nesma/generator/description/stats/{cycleId}	GET			
6303	p	999	/nesma/generator/description/history/{cycleId}	GET			
6304	p	999	/nesma/generator/description/batch-generate	POST			
6305	p	999	/nesma/generator/description/apply	POST			
6306	p	999	/nesma/generator/description/generate	POST			
6307	p	999	/nesma/requirement/max-version/:projectId	GET			
6308	p	999	/nesma/analysis/progress/:taskId	GET			
6309	p	999	/nesma/analysis/analyze	POST			
6310	p	999	/nesma/project-cycle/set-active	POST			
6311	p	999	/nesma/project-cycle/project/:projectId	GET			
6312	p	999	/nesma/project-cycle/list	GET			
6313	p	999	/nesma/project-cycle/:id	GET			
6314	p	999	/nesma/project-cycle/status	PUT			
6315	p	999	/nesma/project-cycle/:id	DELETE			
6316	p	999	/nesma/project-cycle	PUT			
6317	p	999	/nesma/project-cycle	POST			
6318	p	999	/nesma/recommendation/adopt	POST			
6319	p	999	/nesma/recommendation/ai	POST			
6320	p	999	/nesma/recommendation/config	GET			
6321	p	999	/nesma/recommendation/personalized	POST			
6322	p	999	/nesma/recommendation/stats	GET			
6323	p	999	/nesma/recommendation/feedback	POST			
6324	p	999	/nesma/recommendation/behavior	POST			
6325	p	999	/nesma/recommendation/items	GET			
6326	p	999	/nesma/requirement/max-version/2	GET			
6327	p	999	/requirement/delete-by-condition	POST			
6328	p	999	/agent/:id/logs	GET			
6329	p	999	/agent/:id/tasks	GET			
6330	p	999	/nesma-evaluation/:id/recalculate	POST			
6331	p	999	/nesma-evaluation/project/:projectId/summary	GET			
6332	p	999	/nesma-evaluation/:id/export	GET			
6333	p	999	/nesma-evaluation/review	POST			
6334	p	999	/nesma-evaluation/validation-item/:id	PUT			
6335	p	999	/nesma-evaluation/validation-items	GET			
6336	p	999	/nesma-evaluation/complexity-metrics	GET			
6337	p	999	/nesma-evaluation/:id/stats	GET			
6338	p	999	/nesma-evaluation/function-point/:id	DELETE			
6339	p	999	/nesma-evaluation/function-point/:id	PUT			
6340	p	999	/nesma-evaluation/function-point	POST			
6341	p	999	/nesma-evaluation/function-points	GET			
6342	p	999	/nesma-evaluation/list	GET			
6343	p	999	/nesma-evaluation/:id	GET			
6344	p	999	/nesma-evaluation/start	POST			
6345	p	999	/nesma-evaluation/create	POST			
6346	p	999	/requirement-analysis/compare	POST			
6347	p	999	/requirement-analysis/stats	GET			
6348	p	999	/requirement-analysis/config	GET			
6349	p	999	/requirement-analysis/quick-analyze	POST			
6350	p	999	/requirement-analysis/result/:analysisId	GET			
6351	p	999	/requirement-analysis/history	GET			
6352	p	999	/requirement-analysis/batch-analyze	POST			
6353	p	999	/requirement-analysis/analyze	POST			
6354	p	999	/chat/session/:sessionId/export	GET			
6355	p	999	/chat/stats	GET			
6356	p	999	/chat/session/:sessionId/messages	GET			
6357	p	999	/chat/session/:sessionId	GET			
6358	p	999	/chat/sessions	GET			
6359	p	999	/chat/session/:sessionId/clear	POST			
6360	p	999	/chat/session/:sessionId	DELETE			
6361	p	999	/chat/session/:sessionId	PUT			
6362	p	999	/chat/session	POST			
6363	p	999	/chat/send	POST			
6364	p	999	/ai/services/circuit-breaker/stats	GET			
6365	p	999	/ai/services/config	GET			
6366	p	999	/ai/services/health	GET			
6367	p	999	/ai/services	GET			
6368	p	999	/ai/services/circuit-breaker/reset	POST			
6369	p	999	/ai/services/test	POST			
6370	p	999	/ai/services/switch	POST			
6371	p	999	/nesma/template/batch-delete	POST			
6372	p	999	/nesma/template/upload	POST			
6373	p	999	/nesma/template/:id/activate	POST			
6374	p	999	/nesma/template/:id/default	POST			
6375	p	999	/nesma/template/variables	GET			
6376	p	999	/nesma/template/options	GET			
6377	p	999	/nesma/template/list	GET			
6378	p	999	/nesma/template/:id	GET			
6379	p	999	/nesma/template/:id	DELETE			
6380	p	999	/nesma/template	PUT			
6381	p	999	/nesma/template	POST			
6382	p	999	/nesma/document/batch-delete	POST			
6383	p	999	/nesma/document/stats	GET			
6384	p	999	/nesma/document/progress	GET			
6385	p	999	/nesma/document/download	GET			
6386	p	999	/nesma/document/preview	POST			
6387	p	999	/nesma/document/batch-generate	POST			
6388	p	999	/nesma/document/generate	POST			
6389	p	999	/nesma/document/list	GET			
6390	p	999	/nesma/document/:id	GET			
6391	p	999	/nesma/document/:id	DELETE			
6392	p	999	/nesma/document	PUT			
6393	p	999	/nesma/document	POST			
6394	p	999	/agent/statistics/workflow	GET			
6395	p	999	/agent/execution/:id/cancel	POST			
6396	p	999	/agent/execution/:id	GET			
6397	p	999	/agent/workflow/:id/execute	POST			
6398	p	999	/agent/workflows	GET			
6399	p	999	/agent/workflow/:id	GET			
6400	p	999	/agent/workflow	POST			
6401	p	999	/agent/:id/heartbeat	POST			
6402	p	999	/agent/:id/status	PUT			
6403	p	999	/agent/list	GET			
6404	p	999	/agent/:id	GET			
6405	p	999	/agent/register	POST			
6406	p	999	/knowledge/recommended	GET			
6407	p	999	/knowledge/tags/popular	GET			
6408	p	999	/knowledge/import/standard	POST			
6409	p	999	/knowledge/statistics	GET			
6410	p	999	/knowledge/cases/search	POST			
6411	p	999	/knowledge/cases	GET			
6412	p	999	/knowledge/case/:id	GET			
6413	p	999	/knowledge/case/:id	DELETE			
6414	p	999	/knowledge/case/:id	PUT			
6415	p	999	/knowledge/case	POST			
6416	p	999	/knowledge/rules	GET			
6417	p	999	/knowledge/rule/:id	GET			
6418	p	999	/knowledge/rule/:id	DELETE			
6419	p	999	/knowledge/rule/:id	PUT			
6420	p	999	/knowledge/rule	POST			
6421	p	999	/knowledge/entries/search	POST			
6422	p	999	/knowledge/entries	GET			
6423	p	999	/knowledge/entry/:id	GET			
6424	p	999	/knowledge/entry/:id	DELETE			
6425	p	999	/knowledge/entry/:id	PUT			
6426	p	999	/knowledge/entry	POST			
6427	p	999	/nesma/requirement/move	POST			
6428	p	999	/nesma/requirement/batch-update-order	POST			
6429	p	999	/nesma/requirement/import-excel	POST			
6430	p	999	/nesma/requirement/batch-delete	POST			
6431	p	999	/nesma/requirement/parent-options	GET			
6432	p	999	/nesma/requirement/stats	GET			
6433	p	999	/nesma/requirement/tree	GET			
6434	p	999	/nesma/requirement/list	GET			
6435	p	999	/nesma/requirement/:id	GET			
6436	p	999	/nesma/requirement/:id	DELETE			
6437	p	999	/nesma/requirement	PUT			
6438	p	999	/nesma/requirement	POST			
6439	p	999	/nesma/project/stats	GET			
6440	p	999	/nesma/project/list	GET			
6441	p	999	/nesma/project/:id	GET			
6442	p	999	/nesma/project/restore	POST			
6443	p	999	/nesma/project/archive	POST			
6444	p	999	/nesma/project	PUT			
6445	p	999	/nesma/project/delete-batch	DELETE			
6446	p	999	/nesma/project	DELETE			
6447	p	999	/nesma/project	POST			
6448	p	999	/attachmentCategory/deleteCategory	POST			
6449	p	999	/attachmentCategory/addCategory	POST			
6450	p	999	/attachmentCategory/getCategoryList	GET			
6451	p	999	/sysParams/getSysParam	GET			
6452	p	999	/sysParams/getSysParamsList	GET			
6453	p	999	/sysParams/findSysParams	GET			
6454	p	999	/sysParams/updateSysParams	PUT			
6455	p	999	/sysParams/deleteSysParamsByIds	DELETE			
6456	p	999	/sysParams/deleteSysParams	DELETE			
6457	p	999	/sysParams/createSysParams	POST			
6458	p	999	/info/getInfoList	GET			
6459	p	999	/info/findInfo	GET			
6460	p	999	/info/updateInfo	PUT			
6461	p	999	/info/deleteInfoByIds	DELETE			
6462	p	999	/info/deleteInfo	DELETE			
6463	p	999	/info/createInfo	POST			
6464	p	999	/sysExportTemplate/importExcel	POST			
6465	p	999	/sysExportTemplate/exportTemplate	GET			
6466	p	999	/sysExportTemplate/exportExcel	GET			
6467	p	999	/sysExportTemplate/getSysExportTemplateList	GET			
6468	p	999	/sysExportTemplate/findSysExportTemplate	GET			
6469	p	999	/sysExportTemplate/updateSysExportTemplate	PUT			
6470	p	999	/sysExportTemplate/deleteSysExportTemplateByIds	DELETE			
6471	p	999	/sysExportTemplate/deleteSysExportTemplate	DELETE			
6472	p	999	/sysExportTemplate/createSysExportTemplate	POST			
6473	p	999	/authorityBtn/canRemoveAuthorityBtn	POST			
6474	p	999	/authorityBtn/getAuthorityBtn	POST			
6475	p	999	/authorityBtn/setAuthorityBtn	POST			
6476	p	999	/email/sendEmail	POST			
6477	p	999	/email/emailTest	POST			
6478	p	999	/simpleUploader/mergeFileMd5	GET			
6479	p	999	/simpleUploader/checkFileMd5	GET			
6480	p	999	/simpleUploader/upload	POST			
6481	p	999	/sysOperationRecord/deleteSysOperationRecordByIds	DELETE			
6482	p	999	/sysOperationRecord/deleteSysOperationRecord	DELETE			
6483	p	999	/sysOperationRecord/getSysOperationRecordList	GET			
6484	p	999	/sysOperationRecord/findSysOperationRecord	GET			
6485	p	999	/sysOperationRecord/createSysOperationRecord	POST			
6486	p	999	/sysDictionary/getSysDictionaryList	GET			
6487	p	999	/sysDictionary/findSysDictionary	GET			
6488	p	999	/sysDictionary/updateSysDictionary	PUT			
6489	p	999	/sysDictionary/deleteSysDictionary	DELETE			
6490	p	999	/sysDictionary/createSysDictionary	POST			
6491	p	999	/sysDictionaryDetail/getSysDictionaryDetailList	GET			
6492	p	999	/sysDictionaryDetail/findSysDictionaryDetail	GET			
6493	p	999	/sysDictionaryDetail/deleteSysDictionaryDetail	DELETE			
6494	p	999	/sysDictionaryDetail/createSysDictionaryDetail	POST			
6495	p	999	/sysDictionaryDetail/updateSysDictionaryDetail	PUT			
6496	p	999	/autoCode/addFunc	POST			
6497	p	999	/autoCode/delSysHistory	POST			
6498	p	999	/autoCode/getSysHistory	POST			
6499	p	999	/autoCode/rollback	POST			
6500	p	999	/autoCode/getMeta	POST			
6501	p	999	/autoCode/delPackage	POST			
6502	p	999	/autoCode/getPackage	POST			
6503	p	999	/autoCode/getTemplates	GET			
6504	p	999	/autoCode/createPackage	POST			
6505	p	999	/autoCode/mcpList	POST			
6506	p	999	/autoCode/mcpTest	POST			
6507	p	999	/autoCode/mcp	POST			
6508	p	999	/autoCode/pubPlug	POST			
6509	p	999	/autoCode/installPlugin	POST			
6510	p	999	/autoCode/getColumn	GET			
6511	p	999	/autoCode/preview	POST			
6512	p	999	/autoCode/createTemp	POST			
6513	p	999	/autoCode/getTables	GET			
6514	p	999	/autoCode/getDB	GET			
6515	p	999	/customer/customerList	GET			
6516	p	999	/customer/customer	GET			
6517	p	999	/customer/customer	DELETE			
6518	p	999	/customer/customer	POST			
6519	p	999	/customer/customer	PUT			
6520	p	999	/system/setSystemConfig	POST			
6521	p	999	/system/getSystemConfig	POST			
6522	p	999	/system/getServerInfo	POST			
6523	p	999	/fileUploadAndDownload/importURL	POST			
6524	p	999	/fileUploadAndDownload/getFileList	POST			
6525	p	999	/fileUploadAndDownload/editFileName	POST			
6526	p	999	/fileUploadAndDownload/deleteFile	POST			
6527	p	999	/fileUploadAndDownload/upload	POST			
6528	p	999	/fileUploadAndDownload/removeChunk	POST			
6529	p	999	/fileUploadAndDownload/breakpointContinueFinish	POST			
6530	p	999	/fileUploadAndDownload/breakpointContinue	POST			
6531	p	999	/fileUploadAndDownload/findFile	GET			
6532	p	999	/menu/addMenuAuthority	POST			
6533	p	999	/menu/getMenuAuthority	POST			
6534	p	999	/menu/getBaseMenuTree	POST			
6535	p	999	/menu/getMenuList	POST			
6536	p	999	/menu/getBaseMenuById	POST			
6537	p	999	/menu/updateBaseMenu	POST			
6538	p	999	/menu/deleteBaseMenu	POST			
6539	p	999	/menu/getMenu	POST			
6540	p	999	/menu/addBaseMenu	POST			
6541	p	999	/casbin/getPolicyPathByAuthorityId	POST			
6542	p	999	/casbin/updateCasbin	POST			
6543	p	999	/authority/setDataAuthority	POST			
6544	p	999	/authority/getAuthorityList	POST			
6545	p	999	/authority/updateAuthority	PUT			
6546	p	999	/authority/deleteAuthority	POST			
6547	p	999	/authority/createAuthority	POST			
6548	p	999	/authority/copyAuthority	POST			
6549	p	999	/api/ignoreApi	POST			
6550	p	999	/api/enterSyncApi	POST			
6551	p	999	/api/getApiGroups	GET			
6552	p	999	/api/syncApi	GET			
6553	p	999	/api/deleteApisByIds	DELETE			
6554	p	999	/api/getApiById	POST			
6555	p	999	/api/getAllApis	POST			
6556	p	999	/api/getApiList	POST			
6557	p	999	/api/updateApi	POST			
6558	p	999	/api/deleteApi	POST			
6559	p	999	/api/createApi	POST			
6560	p	999	/user/setSelfSetting	PUT			
6561	p	999	/user/resetPassword	POST			
6562	p	999	/user/setUserAuthority	POST			
6563	p	999	/user/changePassword	POST			
6564	p	999	/user/setUserAuthorities	POST			
6565	p	999	/user/getUserInfo	GET			
6566	p	999	/user/setSelfInfo	PUT			
6567	p	999	/user/setUserInfo	PUT			
6568	p	999	/user/getUserList	POST			
6569	p	999	/user/admin_register	POST			
6570	p	999	/user/deleteUser	DELETE			
6571	p	999	/jwt/jsonInBlacklist	POST			
\.


--
-- Data for Name: chat_messages; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.chat_messages (id, created_at, updated_at, deleted_at, session_id, role, content, metadata, token_count) FROM stdin;
\.


--
-- Data for Name: chat_sessions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.chat_sessions (id, created_at, updated_at, deleted_at, user_id, project_id, title, description, model_name, is_active) FROM stdin;
1	2026-10-05 22:29:30.763186+00	2026-10-05 22:29:30.763186+00	\N	1	1	新对话 2026/10/6 06:29:30		deepseek-chat	t
\.


--
-- Data for Name: exa_attachment_category; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.exa_attachment_category (id, created_at, updated_at, deleted_at, name, pid) FROM stdin;
\.


--
-- Data for Name: exa_customers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.exa_customers (id, created_at, updated_at, deleted_at, customer_name, customer_phone_data, sys_user_id, sys_user_authority_id) FROM stdin;
\.


--
-- Data for Name: exa_file_chunks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.exa_file_chunks (id, created_at, updated_at, deleted_at, exa_file_id, file_chunk_number, file_chunk_path) FROM stdin;
\.


--
-- Data for Name: exa_file_upload_and_downloads; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.exa_file_upload_and_downloads (id, created_at, updated_at, deleted_at, name, class_id, url, tag, key) FROM stdin;
\.


--
-- Data for Name: exa_files; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.exa_files (id, created_at, updated_at, deleted_at, file_name, file_md5, file_path, chunk_total, is_finish) FROM stdin;
\.


--
-- Data for Name: gva_announcements_info; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.gva_announcements_info (id, created_at, updated_at, deleted_at, title, content, user_id, attachments) FROM stdin;
\.


--
-- Data for Name: jwt_blacklists; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.jwt_blacklists (id, created_at, updated_at, deleted_at, jwt) FROM stdin;
\.


--
-- Data for Name: nesma_agent_interactions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_agent_interactions (id, created_at, updated_at, deleted_at, session_id, source_agent, target_agent, message_type, message_content, response_content, processing_time_ms, status, error_message) FROM stdin;
\.


--
-- Data for Name: nesma_agent_messages; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_agent_messages (id, created_at, updated_at, deleted_at, message_id, session_id, source_agent_id, target_agent_id, message_type, content_type, content, status, priority, response_to_id, expires_at, metadata, delivered_at, read_at) FROM stdin;
\.


--
-- Data for Name: nesma_agent_tasks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_agent_tasks (id, created_at, updated_at, deleted_at, task_id, agent_id, task_type, priority, status, input_data, output_data, error_message, session_id, user_id, project_id, parent_task_id, retry_count, max_retries, start_time, end_time, processing_time, metadata) FROM stdin;
\.


--
-- Data for Name: nesma_agents; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_agents (id, created_at, updated_at, deleted_at, agent_id, name, description, agent_type, capabilities, status, version, configuration, max_concurrency, last_heartbeat, processing_count, total_processed, average_response_time) FROM stdin;
1	2026-10-05 18:19:52.971076+00	2026-10-05 22:09:56.011442+00	\N	requirement-analysis-agent	需求细化Agent	负责需求扩充与细化，结合知识库提供上下文，分析用户输入的一二三级需求，生成详细的需求规格说明	REQUIREMENT_ANALYSIS	["requirement_expansion", "context_analysis", "requirement_refinement", "knowledge_integration"]	active	1.0.0	{"model": "deepseek-chat", "max_tokens": 2000, "temperature": 0.7, "prompt_template": "requirement_analysis_prompt"}	5	2026-10-05 18:19:52.969411+00	0	0	0
2	2026-10-05 18:19:53.002033+00	2026-10-05 22:09:56.025942+00	\N	nesma-evaluation-agent	NESMA评估Agent	负责功能点识别与度量，应用NESMA标准规则，对需求进行功能点计算和复杂度评估	NESMA_EVALUATION	["function_point_identification", "complexity_assessment", "nesma_calculation", "standard_compliance"]	active	1.0.0	{"nesma_version": "2.2", "prompt_template": "nesma_evaluation_prompt", "calculation_mode": "detailed", "validation_level": "strict"}	3	2026-10-05 18:19:53.000324+00	0	0	0
3	2026-10-05 18:19:53.019026+00	2026-10-05 22:09:56.036607+00	\N	document-generation-agent	文档生成Agent	负责多格式文档生成，遵循最佳实践模板，生成需求规格说明书、NESMA评估报告等标准文档	DOCUMENT_GENERATION	["document_templating", "multi_format_export", "content_structuring", "format_validation"]	active	1.0.0	{"quality_check": true, "auto_formatting": true, "template_version": "1.0", "supported_formats": ["docx", "pdf", "excel", "html"]}	4	2026-10-05 18:19:53.016793+00	0	0	0
4	2026-10-05 18:19:53.043331+00	2026-10-05 22:09:56.046077+00	\N	memory-agent	记忆Agent	负责长期记忆、上下文检索与智能遗忘，管理系统的记忆存储和检索，提供上下文感知能力	MEMORY_MANAGEMENT	["vector_storage", "context_retrieval", "intelligent_forgetting", "memory_compression"]	active	1.0.0	{"memory_types": ["vector", "compact", "session"], "max_memory_size": 10000, "retention_policy": "adaptive", "similarity_threshold": 0.7}	10	2026-10-05 18:19:53.041652+00	0	0	0
5	2026-10-05 18:19:53.062705+00	2026-10-05 22:09:56.056181+00	\N	knowledge-base-agent	知识库Agent	负责知识检索、规则匹配、案例推荐，管理NESMA标准库、最佳实践库和历史案例库	KNOWLEDGE_RETRIEVAL	["knowledge_search", "rule_matching", "case_recommendation", "content_validation"]	active	1.0.0	{"cache_enabled": true, "ranking_method": "relevance_score", "search_algorithm": "semantic", "knowledge_sources": ["nesma_standards", "best_practices", "case_studies", "rules"]}	8	2026-10-05 18:19:53.060296+00	0	0	0
\.


--
-- Data for Name: nesma_analysis_histories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_analysis_histories (id, created_at, updated_at, deleted_at, project_id, analysis_type, analysis_date, quality_score, function_points, requirement_count, complexity_index, quality_delta, fp_delta, requirement_delta, project_analysis_id, evaluation_id) FROM stdin;
\.


--
-- Data for Name: nesma_case_studies; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_case_studies (id, created_at, updated_at, deleted_at, project_name, domain, requirements_summary, nesma_results, lessons_learned, embedding, total_function_points, project_duration, team_size, complexity, success) FROM stdin;
\.


--
-- Data for Name: nesma_compact_memories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_compact_memories (id, created_at, updated_at, deleted_at, user_id, project_id, summary, interaction_count, last_updated, importance) FROM stdin;
\.


--
-- Data for Name: nesma_complexity_metrics; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_complexity_metrics (id, created_at, updated_at, deleted_at, evaluation_id, metric_type, metric_name, metric_value, threshold_value, score, weight, calculation_formula, calculation_details, quality_indicator, impact_level) FROM stdin;
\.


--
-- Data for Name: nesma_doc_batches; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_doc_batches (id, created_at, updated_at, deleted_at, project_id, name, description, total_count, success_count, failed_count, status, progress, config, started_at, completed_at, created_by) FROM stdin;
\.


--
-- Data for Name: nesma_doc_templates; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_doc_templates (id, created_at, updated_at, deleted_at, name, description, type, format, category, template_path, preview_path, config, variables, is_default, is_active, usage_count, created_by, content) FROM stdin;
1	2026-10-05 18:19:52.279632+00	2026-10-05 18:19:52.279632+00	\N	标准需求规格说明书模板	基于NESMA标准的需求规格说明书模板，包含完整的功能点分析框架	word	requirement_spec	标准模板			\N	\N	t	t	0	1	需求规格说明书模板\n\n1. 项目概述\n   - 项目名称: {{project_name}}\n   - 项目描述: {{project_description}}\n   - 项目版本: {{project_version}}\n   - 项目经理: {{project_manager}}\n   - 开发团队: {{development_team}}\n\n2. 功能需求分析\n   2.1 业务功能概述\n   - 核心业务流程\n   - 用户角色定义\n   - 系统边界确定\n\n   2.2 功能点识别\n   - 数据功能点 (Data Functions)\n     * 内部逻辑文件 (ILF): {{ilf_list}}\n     * 外部接口文件 (EIF): {{eif_list}}\n   \n   - 事务功能点 (Transaction Functions)  \n     * 外部输入 (EI): {{ei_list}}\n     * 外部输出 (EO): {{eo_list}}\n     * 外部查询 (EQ): {{eq_list}}\n\n3. NESMA功能点评估\n   3.1 复杂度评估标准\n   - 低复杂度 (Low): 3-4个权重\n   - 中等复杂度 (Average): 4-6个权重  \n   - 高复杂度 (High): 6-7个权重\n\n   3.2 功能点统计表\n   | 类型 | 低复杂度 | 中等复杂度 | 高复杂度 | 小计 |\n   |------|----------|------------|----------|------|\n   | ILF  | {{ilf_low}} × 7 | {{ilf_avg}} × 10 | {{ilf_high}} × 15 | {{ilf_total}} |\n   | EIF  | {{eif_low}} × 5 | {{eif_avg}} × 7 | {{eif_high}} × 10 | {{eif_total}} |\n   | EI   | {{ei_low}} × 3 | {{ei_avg}} × 4 | {{ei_high}} × 6 | {{ei_total}} |\n   | EO   | {{eo_low}} × 4 | {{eo_avg}} × 5 | {{eo_high}} × 7 | {{eo_total}} |\n   | EQ   | {{eq_low}} × 3 | {{eq_avg}} × 4 | {{eq_high}} × 6 | {{eq_total}} |\n\n   总计未调整功能点: {{total_unadjusted_fp}} FP\n\n4. 技术要求\n   - 系统架构: {{system_architecture}}\n   - 技术栈: {{technology_stack}}\n   - 性能要求: {{performance_requirements}}\n   - 安全要求: {{security_requirements}}\n\n5. 验收标准\n   - 功能验收标准\n   - 性能验收标准\n   - 安全验收标准
2	2026-10-05 18:19:52.298248+00	2026-10-05 18:19:52.298248+00	\N	NESMA功能点评估报告模板	专业的NESMA功能点评估报告模板，符合国际标准	excel	nesma_report	评估报告			\N	\N	t	t	0	1	NESMA功能点评估报告\n\n项目信息\n========\n项目名称: {{project_name}}\n评估日期: {{evaluation_date}}\n评估人员: {{evaluator_name}}\n评估版本: {{evaluation_version}}\n项目阶段: {{project_phase}}\n\n评估摘要\n========\n评估方法: NESMA Function Point Analysis v2.2\n评估范围: {{evaluation_scope}}\n评估精度: {{evaluation_accuracy}}\n\n功能点统计\n==========\n\n1. 数据功能点 (Data Functions)\n   内部逻辑文件 (ILF):\n   - 低复杂度: {{ilf_low_count}} 个 × 7 FP = {{ilf_low_fp}} FP\n   - 中等复杂度: {{ilf_avg_count}} 个 × 10 FP = {{ilf_avg_fp}} FP  \n   - 高复杂度: {{ilf_high_count}} 个 × 15 FP = {{ilf_high_fp}} FP\n   小计: {{ilf_total_fp}} FP\n\n   外部接口文件 (EIF):\n   - 低复杂度: {{eif_low_count}} 个 × 5 FP = {{eif_low_fp}} FP\n   - 中等复杂度: {{eif_avg_count}} 个 × 7 FP = {{eif_avg_fp}} FP\n   - 高复杂度: {{eif_high_count}} 个 × 10 FP = {{eif_high_fp}} FP\n   小计: {{eif_total_fp}} FP\n\n2. 事务功能点 (Transaction Functions)\n   外部输入 (EI):\n   - 低复杂度: {{ei_low_count}} 个 × 3 FP = {{ei_low_fp}} FP\n   - 中等复杂度: {{ei_avg_count}} 个 × 4 FP = {{ei_avg_fp}} FP\n   - 高复杂度: {{ei_high_count}} 个 × 6 FP = {{ei_high_fp}} FP\n   小计: {{ei_total_fp}} FP\n\n   外部输出 (EO):\n   - 低复杂度: {{eo_low_count}} 个 × 4 FP = {{eo_low_fp}} FP\n   - 中等复杂度: {{eo_avg_count}} 个 × 5 FP = {{eo_avg_fp}} FP\n   - 高复杂度: {{eo_high_count}} 个 × 7 FP = {{eo_high_fp}} FP\n   小计: {{eo_total_fp}} FP\n\n   外部查询 (EQ):\n   - 低复杂度: {{eq_low_count}} 个 × 3 FP = {{eq_low_fp}} FP\n   - 中等复杂度: {{eq_avg_count}} 个 × 4 FP = {{eq_avg_fp}} FP\n   - 高复杂度: {{eq_high_count}} 个 × 6 FP = {{eq_high_fp}} FP\n   小计: {{eq_total_fp}} FP\n\n功能点汇总\n==========\n数据功能点总计: {{data_fp_total}} FP ({{data_fp_percentage}}%)\n事务功能点总计: {{transaction_fp_total}} FP ({{transaction_fp_percentage}}%)\n未调整功能点总计: {{unadjusted_fp_total}} FP\n\n调整因子评估\n============\n技术复杂度调整因子: {{technical_complexity_factor}}\n环境复杂度调整因子: {{environment_complexity_factor}}\n综合调整因子: {{overall_adjustment_factor}}\n\n最终结果\n========\n调整后功能点: {{adjusted_fp_total}} FP\n工作量估算: {{effort_estimate}} 人天\n开发周期估算: {{duration_estimate}} 天\n成本估算: {{cost_estimate}} 元\n\n质量指标\n========\n识别置信度: {{identification_confidence}}%\n评估准确度: {{evaluation_accuracy}}%\n合规性评分: {{compliance_score}}/100\n\n建议和说明\n==========\n{{recommendations_and_notes}}
3	2026-10-05 18:19:52.31347+00	2026-10-05 18:19:52.31347+00	\N	业务需求汇总表模板	便于项目管理的业务需求汇总表格模板	excel	business_summary	管理表格			\N	\N	t	t	0	1	业务需求汇总表\n\n项目基本信息\n============\n项目编号: {{project_code}}\n项目名称: {{project_name}}\n所属部门: {{department}}\n项目经理: {{project_manager}}\n业务负责人: {{business_owner}}\n技术负责人: {{technical_lead}}\n预计开始日期: {{start_date}}\n预计结束日期: {{end_date}}\n项目状态: {{project_status}}\n\n需求清单\n========\n序号 | 需求ID | 需求名称 | 需求描述 | 优先级 | 业务价值 | 功能点类型 | 复杂度 | 预估工时 | 责任人 | 状态 | 备注\n-----|--------|----------|----------|--------|----------|------------|--------|----------|--------|------|------\n1    | REQ001 | 用户注册 | 系统用户注册功能 | 高 | 核心 | EI | 中 | 8h | 张三 | 开发中 | 包含邮箱验证\n2    | REQ002 | 用户登录 | 系统用户登录功能 | 高 | 核心 | EI | 低 | 4h | 李四 | 已完成 | 支持多种登录方式\n3    | REQ003 | 用户信息管理 | 用户个人信息维护 | 中 | 重要 | ILF | 中 | 12h | 王五 | 待开发 | \n4    | REQ004 | 数据查询 | 业务数据查询展示 | 高 | 核心 | EO | 中 | 16h | 赵六 | 开发中 | 支持多条件查询\n5    | REQ005 | 报表生成 | 统计报表生成导出 | 中 | 重要 | EO | 高 | 24h | 孙七 | 待开发 | 支持多种格式\n\n功能点统计\n==========\n功能点类型分布:\n- 内部逻辑文件 (ILF): {{ilf_count}} 个\n- 外部接口文件 (EIF): {{eif_count}} 个  \n- 外部输入 (EI): {{ei_count}} 个\n- 外部输出 (EO): {{eo_count}} 个\n- 外部查询 (EQ): {{eq_count}} 个\n\n复杂度分布:\n- 低复杂度: {{low_complexity_count}} 个 ({{low_complexity_percentage}}%)\n- 中等复杂度: {{avg_complexity_count}} 个 ({{avg_complexity_percentage}}%)\n- 高复杂度: {{high_complexity_count}} 个 ({{high_complexity_percentage}}%)\n\n总功能点数: {{total_function_points}} FP\n数据功能点: {{data_function_points}} FP ({{data_percentage}}%)\n事务功能点: {{transaction_function_points}} FP ({{transaction_percentage}}%)\n\n项目估算\n========\n预估指标:\n- 总工作量: {{total_effort}} 人天\n- 开发周期: {{development_duration}} 天\n- 测试周期: {{testing_duration}} 天\n- 总项目周期: {{total_duration}} 天\n- 项目成本: {{project_cost}} 元\n\n人力资源需求:\n- 项目经理: {{pm_count}} 人\n- 系统分析师: {{analyst_count}} 人  \n- 开发工程师: {{developer_count}} 人\n- 测试工程师: {{tester_count}} 人\n- UI/UX设计师: {{designer_count}} 人\n\n风险评估\n========\n主要风险:\n1. {{risk_1}}\n2. {{risk_2}}\n3. {{risk_3}}\n\n缓解措施:\n1. {{mitigation_1}}\n2. {{mitigation_2}}\n3. {{mitigation_3}}\n\n里程碑计划\n==========\n里程碑 | 计划完成日期 | 主要交付物 | 责任人 | 状态\n-------|--------------|------------|--------|------\n需求分析完成 | {{milestone_1_date}} | 需求规格说明书 | {{milestone_1_owner}} | {{milestone_1_status}}\n系统设计完成 | {{milestone_2_date}} | 系统设计文档 | {{milestone_2_owner}} | {{milestone_2_status}}\n开发完成 | {{milestone_3_date}} | 系统代码 | {{milestone_3_owner}} | {{milestone_3_status}}\n测试完成 | {{milestone_4_date}} | 测试报告 | {{milestone_4_owner}} | {{milestone_4_status}}\n上线部署 | {{milestone_5_date}} | 生产系统 | {{milestone_5_owner}} | {{milestone_5_status}}\n\n更新日志\n========\n版本 | 更新日期 | 更新内容 | 更新人\n-----|----------|----------|--------\nv1.0 | {{update_date_1}} | 初始版本创建 | {{updater_1}}\nv1.1 | {{update_date_2}} | 添加新需求REQ006-REQ010 | {{updater_2}}\nv1.2 | {{update_date_3}} | 修订功能点评估 | {{updater_3}}
\.


--
-- Data for Name: nesma_documents; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_documents (id, created_at, updated_at, deleted_at, project_id, template_id, name, description, type, format, file_path, file_size, status, progress, error_msg, config, version, generated_at, created_by, batch_id, cycle_id, version_id) FROM stdin;
\.


--
-- Data for Name: nesma_embeddings; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_embeddings (id, created_at, updated_at, deleted_at, embedding_id, content, vector, dimensions, model, token_count, cost, provider, metadata, hash, is_active) FROM stdin;
\.


--
-- Data for Name: nesma_evaluation_factors; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_evaluation_factors (id, created_at, updated_at, deleted_at, evaluation_id, factor_config, nesma_version) FROM stdin;
\.


--
-- Data for Name: nesma_evaluation_history; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_evaluation_history (id, created_at, updated_at, deleted_at, project_id, current_eval_id, previous_eval_id, fp_difference, fp_change_percent, complexity_change, change_category, change_summary, change_reasons, impact_analysis) FROM stdin;
\.


--
-- Data for Name: nesma_evaluations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_evaluations (id, created_at, updated_at, deleted_at, project_id, evaluation_name, evaluation_version, evaluation_type, status, total_function_points, data_function_points, transactional_fp, adjustment_factor, adjusted_function_points, simple_function_count, average_function_count, complex_function_count, evaluation_config, nesma_rules, confidence_score, accuracy_score, compliance_score, evaluator_id, start_time, completion_time, duration, evaluation_details, validation_results, recommendation_list, review_status, reviewer_id, review_comments, review_time, cycle_id, requirement_version_id, error_message, error_code, error_details, progress, current_phase, total_steps, completed_steps, processed_requirements, total_requirements, progress_message, ai_analysis_id, quality_score, overall_grade, complexity_level, risk_level, evaluation_summary) FROM stdin;
1	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	一期需求评估 · 人工演示样例	1.0	detailed	completed	34	12	22	1.1	37.4	5	3	0	{}	{}	0.9	0.9	0.9	1	2026-10-05 22:04:43.736265+00	2026-10-05 22:09:43.736265+00	300	{}	[]	[]	pending	\N	\N	\N	1	1	\N	\N	\N	100	completed	5	5	14	14	人工录入的展示记录，未执行 AI 计算。	\N	90	A	medium	low	此评估记录仅用于展示界面；计数与调整因子均为人工填写的演示数据。
\.


--
-- Data for Name: nesma_function_points; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_function_points (id, created_at, updated_at, deleted_at, evaluation_id, requirement_id, function_type, function_name, function_desc, data_elements, file_types, record_elements, complexity_level, weight_factor, calculated_points, identification_method, confidence_level, detection_rules, is_validated, validation_status, validation_notes) FROM stdin;
\.


--
-- Data for Name: nesma_knowledge_entries; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_knowledge_entries (id, created_at, updated_at, deleted_at, title, content, category, domain, tags, embedding, confidence_score, usage_count, version, status, source, author) FROM stdin;
2	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	需求可测试性检查	需求应明确输入、处理规则、输出与异常路径；验收条件应可以通过测试验证。	BEST_PRACTICE	需求工程	["验收", "质量"]	\N	0.9	8	1.0	active	人工编写演示条目	蜂巢工作室
3	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	智慧园区工单业务边界	示例工单流程包含受理、分类、分派、处理反馈、用户确认与关闭。实际计量前需明确系统边界。	RULE	智慧园区	["工单", "业务边界"]	\N	0.85	6	1.0	active	虚构园区演示案例	蜂巢工作室
1	2026-10-05 22:09:43.736265+00	2026-10-05 22:31:13.939157+00	\N	功能点分类速查：EI / EO / EQ / ILF / EIF	EI 表示外部输入，EO 表示外部输出，EQ 表示外部查询，ILF 表示内部逻辑文件，EIF 表示外部接口文件。应结合业务边界和数据处理特征进行人工识别。	NESMA_STANDARD	通用软件	["功能点", "人工审核"]	\N	0.9	13	1.0	active	蜂巢工作室演示整理	蜂巢工作室
\.


--
-- Data for Name: nesma_knowledge_rules; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_knowledge_rules (id, created_at, updated_at, deleted_at, rule_name, condition_expression, action_expression, priority, is_active, description, category) FROM stdin;
\.


--
-- Data for Name: nesma_nesma_evaluations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_nesma_evaluations (id, created_at, updated_at, deleted_at, project_id, cycle_id, evaluation_type, overall_grade, total_afp, total_ufp, quality_score, nesma_compliance, industry_ranking, investment_grade, ei_count, eo_count, eq_count, ilf_count, eif_count, low_complexity_count, avg_complexity_count, high_complexity_count, measurement_consistency, documentation_completeness, traceability_score, evaluation_result, function_point_analysis, quality_evaluation, compliance_check, benchmarking_analysis, key_findings, critical_recommendations, improvement_roadmap, evaluator_model, evaluation_standard, confidence_level, evaluation_date) FROM stdin;
\.


--
-- Data for Name: nesma_optimization_suggestions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_optimization_suggestions (id, created_at, updated_at, deleted_at, project_id, requirement_id, suggestion_type, category, title, description, rationale, expected_outcome, priority, impact, effort, action_items, implementation_notes, timeline, resource_requirements, status, assigned_to, start_date, due_date, completed_date, actual_outcome, effectiveness_score, generated_by, ai_model, confidence_score) FROM stdin;
\.


--
-- Data for Name: nesma_project_analyses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_project_analyses (id, created_at, updated_at, deleted_at, project_id, cycle_id, version_id, analysis_type, analysis_scope, analysis_date, overall_grade, health_score, success_probability, quality_score, total_requirements, total_function_points, complexity_index, analysis_result, key_strengths, major_risks, recommendations, ai_model, confidence_score) FROM stdin;
\.


--
-- Data for Name: nesma_project_cycles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_project_cycles (id, created_at, updated_at, deleted_at, project_id, name, description, start_date, end_date, status, phase, settings, requirement_count, completed_count, progress) FROM stdin;
2	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	二期 · 数据运营提升	运营看板与跨系统协同，演示周期。	2026-07-01 00:00:00+00	2026-12-31 00:00:00+00	planned	phase2	{}	0	0	0
1	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	一期 · 基础能力建设	资产台账与工单闭环，演示周期。	2026-01-01 00:00:00+00	2026-06-30 00:00:00+00	active	phase1	{}	13	5	38
\.


--
-- Data for Name: nesma_projects; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_projects (id, created_at, updated_at, deleted_at, name, description, owner_id, status, domain, domain_tags, settings, start_date, end_date, active_cycle_id) FROM stdin;
1	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	蜂巢工作室 · 智慧园区演示	用于开源文档展示的虚构项目，覆盖园区资产、企业服务、工单协同与运营统计。所有数据均为人工录入演示样例。	1	active	智慧园区	["园区运营", "企业服务", "演示数据"]	{}	2026-01-01 00:00:00+00	2026-12-31 00:00:00+00	1
\.


--
-- Data for Name: nesma_requirement_analysis; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_requirement_analysis (id, created_at, updated_at, deleted_at, requirement_id, ai_expanded_content, nesma_category, complexity_score, function_points, confidence_level, knowledge_base_refs, analysis_version, analysis_details, review_status, review_comments) FROM stdin;
\.


--
-- Data for Name: nesma_requirement_analysis_tasks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_requirement_analysis_tasks (id, created_at, updated_at, deleted_at, project_id, cycle_id, source_version_id, target_version_id, task_type, status, progress, config, priority, start_time, end_time, duration, total_count, processed_count, success_count, failed_count, skipped_count, error_msg, error_details, result, summary, ai_call_count, ai_call_duration, ai_token_usage, quality_score, accuracy_rate, requirement_ids) FROM stdin;
\.


--
-- Data for Name: nesma_requirement_optimizations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_requirement_optimizations (id, created_at, updated_at, deleted_at, requirement_id, project_id, optimization_type, original_title, original_description, original_quality_score, optimized_title, optimized_description, optimized_quality_score, improvement_summary, quality_comparison, optimization_rationale, user_decision_required, decision_points, user_decision, user_decision_time, user_decision_reason, applied_to_requirement, applied_time, ai_model, analysis_confidence, generated_at) FROM stdin;
\.


--
-- Data for Name: nesma_requirement_versions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_requirement_versions (id, created_at, updated_at, deleted_at, cycle_id, version, version_type, created_by, summary, description, changes, status, requirement_count, analyzed_count, optimized_count, analysis_progress, analysis_task_id, analysis_start_time, analysis_end_time, analysis_duration, quality_score, confidence_score, review_status, review_notes) FROM stdin;
1	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	1.0	original	蜂巢工作室	人工录入演示基线	开源截图专用需求版本，不包含真实客户信息。	{}	active	13	0	0	0	\N	\N	\N	\N	\N	\N	pending	\N
2	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	1.1	optimized	蜂巢工作室	人工整理的迭代示例	用于展示版本界面，未声称由 AI 生成。	{}	draft	0	0	0	0	\N	\N	\N	\N	\N	\N	pending	\N
\.


--
-- Data for Name: nesma_requirements; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_requirements (id, created_at, updated_at, deleted_at, project_id, parent_id, level, title, description, priority, status, order_index, domain_tags, code, category, complexity, estimate_hours, actual_hours, business_value, acceptance_criteria, notes, import_batch, import_source, construction_period, version, afp, ufp, function_type, reuse_level, modification_type, cycle_id, version_id, ai_analysis_status, ai_description, ai_generated_title, ai_complexity_score, ai_confidence_score, ai_analysis_time, ai_analysis_log, similar_requirements, recommended_afp, recommended_ufp, ai_optimization_applied, ai_optimization_applied_at, complexity_level, knowledge_references, a_idescription, mermaid_diagram, mermaid_generated_at) FROM stdin;
4	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	3	4	新增设备档案	输入设备基础信息并保存为台账记录。	4	completed	1	[]	ASSET.01.01	\N	简单	\N	\N	建立统一资产底账。	必填字段校验通过后保存，并返回唯一设备编号。	人工录入的计数样例，需按实际边界复核。	\N	\N	\N	0	3.3	3	EI	低	新增	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
5	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	3	4	查询设备台账	按名称、类型和状态检索设备，支持分页。	4	completed	2	[]	ASSET.01.02	\N	中等	\N	\N	提高设备检索效率。	组合筛选结果正确，分页稳定。	演示样例	\N	\N	\N	0	4.4	4	EQ	中	新增	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
6	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	3	4	维护设备状态	记录设备启用、停用和维修状态变化。	3	pending	3	[]	ASSET.01.03	\N	中等	\N	\N	保留设备状态轨迹。	状态变更留有日志。	演示样例	\N	\N	\N	0	4.4	4	EI	中	优化	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
7	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	3	4	设备基础信息集	维护业务边界内设备主数据。	4	completed	4	[]	ASSET.01.04	\N	简单	\N	\N	支撑园区资产数据管理。	数据元素定义和逻辑文件边界经人工确认。	演示样例	\N	\N	\N	0	7.7	7	ILF	低	新增	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
10	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	9	4	提交企业服务诉求	企业登记诉求内容、类别和联系方式。	5	completed	1	[]	SERVICE.01.01	\N	简单	\N	\N	\N	\N	\N	\N	\N	\N	0	3.3	3	EI	低	新增	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	flowchart LR; A[提交诉求] --> B[校验信息]; B --> C[创建工单]; C --> D[分派处理]; D --> E[反馈确认]; E --> F[关闭工单]	\N
11	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	9	4	工单状态查询	查询工单处理状态与反馈记录。	4	completed	2	[]	SERVICE.01.02	\N	简单	\N	\N	\N	\N	\N	\N	\N	\N	0	3.3	3	EQ	中	新增	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N		\N
12	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	9	4	服务统计报表	统计各类别诉求数量、完成率和处理周期。	3	pending	3	[]	SERVICE.01.03	\N	中等	\N	\N	\N	\N	\N	\N	\N	\N	0	5.5	5	EO	低	新增	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N		\N
13	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	9	4	企业基础信息引用	引用外部企业信息库的基础资料。	3	pending	4	[]	SERVICE.01.04	\N	简单	\N	\N	\N	\N	\N	\N	\N	\N	0	5.5	5	EIF	高	新增	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N		\N
1	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	\N	1	智慧园区运营	整合资产管理、企业服务与运营决策。	3	pending	1	["演示"]	PARK	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	0	\N	\N	\N	\N	\N	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
2	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	1	2	园区资产管理	维护设备与房产台账，支持全生命周期记录。	4	pending	1	[]	PARK.ASSET	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	0	\N	\N	\N	\N	\N	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
3	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	2	3	设备台账管理	登记设备编号、名称、类型、位置和运行状态，支持条件查询及信息变更。	4	pending	1	[]	PARK.ASSET.01	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	0	\N	\N	\N	\N	\N	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
8	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	1	2	企业服务协同	为入驻企业提供诉求受理和服务协作。	4	pending	2	[]	PARK.SERVICE	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	0	\N	\N	\N	\N	\N	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
9	2026-10-05 22:09:43.736265+00	2026-10-05 22:09:43.736265+00	\N	1	8	3	服务工单闭环	受理企业诉求，分派、处理、反馈并关闭服务工单。	5	pending	1	[]	PARK.SERVICE.01	\N	\N	\N	\N	\N	\N	\N	\N	\N	\N	0	\N	\N	\N	\N	\N	1	1	pending	\N	\N	\N	\N	\N	\N	\N	\N	\N	f	\N	\N	\N	\N	\N	\N
\.


--
-- Data for Name: nesma_session_memories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_session_memories (id, created_at, updated_at, deleted_at, session_id, user_id, conversation_history, expires_at, is_active) FROM stdin;
\.


--
-- Data for Name: nesma_system_activities; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_system_activities (id, type, category, title, message, user_id, user_name, project_id, project_name, entity_type, entity_id, metadata, ip_address, user_agent, created_at, updated_at, deleted_at) FROM stdin;
\.


--
-- Data for Name: nesma_validation_items; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_validation_items (id, created_at, updated_at, deleted_at, evaluation_id, validation_rule, rule_description, validation_result, expected_value, actual_value, deviation_level, validation_notes, validation_details, is_critical, resolution_status, resolution_notes) FROM stdin;
\.


--
-- Data for Name: nesma_vector_indexes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_vector_indexes (id, created_at, updated_at, deleted_at, index_name, target_table, column_name, dimensions, index_type, index_method, configuration, is_built, build_time, status) FROM stdin;
\.


--
-- Data for Name: nesma_vector_memories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_vector_memories (id, created_at, updated_at, deleted_at, user_id, project_id, content, embedding, metadata, relevance_score, memory_type, expires_at) FROM stdin;
\.


--
-- Data for Name: nesma_vector_queries; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_vector_queries (id, created_at, updated_at, deleted_at, query_id, query_vector, query_content, result_count, results, threshold, query_time, user_id, project_id, session_id, query_type, metadata) FROM stdin;
\.


--
-- Data for Name: nesma_vector_stores; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_vector_stores (id, created_at, updated_at, deleted_at, vector_id, content, embedding, metadata, content_type, source, project_id, user_id, tags, hash, version, is_active, expires_at) FROM stdin;
\.


--
-- Data for Name: nesma_workflow_executions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_workflow_executions (id, created_at, updated_at, deleted_at, execution_id, workflow_id, session_id, status, input_data, output_data, current_step, execution_log, error_message, start_time, end_time, total_duration, user_id, project_id, metadata) FROM stdin;
\.


--
-- Data for Name: nesma_workflows; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.nesma_workflows (id, created_at, updated_at, deleted_at, workflow_id, name, description, definition, status, version, category, tags, is_template, usage_count, created_by) FROM stdin;
\.


--
-- Data for Name: recommendation_adoptions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.recommendation_adoptions (id, user_id, project_id, recommendation_id, title, content, category, confidence, knowledge_refs, user_feedback, adopted_at, created_at) FROM stdin;
\.


--
-- Data for Name: recommendation_histories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.recommendation_histories (id, created_at, updated_at, deleted_at, user_id, recommendation_type, item_type, item_id, score, reason, algorithm, context, is_clicked, is_interacted, feedback_score, feedback_comment) FROM stdin;
\.


--
-- Data for Name: requirement_analysis_results; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.requirement_analysis_results (id, created_at, updated_at, deleted_at, user_id, project_id, requirement_id, analysis_type, input_text, result, suggestions, quality_score, processing_time, model_name, version, status, error_message, original_text, score, ai_model) FROM stdin;
\.


--
-- Data for Name: sys_apis; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_apis (id, created_at, updated_at, deleted_at, path, description, api_group, method) FROM stdin;
1	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/jwt/jsonInBlacklist	jwt加入黑名单(退出，必选)	jwt	POST
2	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/deleteUser	删除用户	系统用户	DELETE
3	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/admin_register	用户注册	系统用户	POST
4	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/getUserList	获取用户列表	系统用户	POST
5	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/setUserInfo	设置用户信息	系统用户	PUT
6	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/setSelfInfo	设置自身信息(必选)	系统用户	PUT
7	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/getUserInfo	获取自身信息(必选)	系统用户	GET
8	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/setUserAuthorities	设置权限组	系统用户	POST
9	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/changePassword	修改密码（建议选择)	系统用户	POST
10	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/setUserAuthority	修改用户角色(必选)	系统用户	POST
11	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/resetPassword	重置用户密码	系统用户	POST
12	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/user/setSelfSetting	用户界面配置	系统用户	PUT
13	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/createApi	创建api	api	POST
14	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/deleteApi	删除Api	api	POST
15	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/updateApi	更新Api	api	POST
16	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/getApiList	获取api列表	api	POST
17	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/getAllApis	获取所有api	api	POST
18	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/getApiById	获取api详细信息	api	POST
19	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/deleteApisByIds	批量删除api	api	DELETE
20	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/syncApi	获取待同步API	api	GET
21	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/getApiGroups	获取路由组	api	GET
22	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/enterSyncApi	确认同步API	api	POST
23	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/api/ignoreApi	忽略API	api	POST
24	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authority/copyAuthority	拷贝角色	角色	POST
25	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authority/createAuthority	创建角色	角色	POST
26	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authority/deleteAuthority	删除角色	角色	POST
27	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authority/updateAuthority	更新角色信息	角色	PUT
28	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authority/getAuthorityList	获取角色列表	角色	POST
29	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authority/setDataAuthority	设置角色资源权限	角色	POST
30	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/casbin/updateCasbin	更改角色api权限	casbin	POST
31	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/casbin/getPolicyPathByAuthorityId	获取权限列表	casbin	POST
32	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/addBaseMenu	新增菜单	菜单	POST
33	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/getMenu	获取菜单树(必选)	菜单	POST
34	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/deleteBaseMenu	删除菜单	菜单	POST
35	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/updateBaseMenu	更新菜单	菜单	POST
36	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/getBaseMenuById	根据id获取菜单	菜单	POST
37	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/getMenuList	分页获取基础menu列表	菜单	POST
38	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/getBaseMenuTree	获取用户动态路由	菜单	POST
39	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/getMenuAuthority	获取指定角色menu	菜单	POST
40	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/menu/addMenuAuthority	增加menu和角色关联关系	菜单	POST
41	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/findFile	寻找目标文件（秒传）	分片上传	GET
42	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/breakpointContinue	断点续传	分片上传	POST
43	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/breakpointContinueFinish	断点续传完成	分片上传	POST
44	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/removeChunk	上传完成移除文件	分片上传	POST
45	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/upload	文件上传（建议选择）	文件上传与下载	POST
46	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/deleteFile	删除文件	文件上传与下载	POST
47	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/editFileName	文件名或者备注编辑	文件上传与下载	POST
48	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/getFileList	获取上传文件列表	文件上传与下载	POST
49	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/fileUploadAndDownload/importURL	导入URL	文件上传与下载	POST
50	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/system/getServerInfo	获取服务器信息	系统服务	POST
51	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/system/getSystemConfig	获取配置文件内容	系统服务	POST
52	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/system/setSystemConfig	设置配置文件内容	系统服务	POST
53	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/customer/customer	更新客户	客户	PUT
54	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/customer/customer	创建客户	客户	POST
55	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/customer/customer	删除客户	客户	DELETE
56	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/customer/customer	获取单一客户	客户	GET
57	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/customer/customerList	获取客户列表	客户	GET
58	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/getDB	获取所有数据库	代码生成器	GET
59	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/getTables	获取数据库表	代码生成器	GET
60	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/createTemp	自动化代码	代码生成器	POST
61	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/preview	预览自动化代码	代码生成器	POST
62	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/getColumn	获取所选table的所有字段	代码生成器	GET
63	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/installPlugin	安装插件	代码生成器	POST
64	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/pubPlug	打包插件	代码生成器	POST
65	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/mcp	自动生成 MCP Tool 模板	代码生成器	POST
66	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/mcpTest	MCP Tool 测试	代码生成器	POST
67	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/mcpList	获取 MCP ToolList	代码生成器	POST
68	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/createPackage	配置模板	模板配置	POST
69	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/getTemplates	获取模板文件	模板配置	GET
70	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/getPackage	获取所有模板	模板配置	POST
71	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/delPackage	删除模板	模板配置	POST
72	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/getMeta	获取meta信息	代码生成器历史	POST
73	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/rollback	回滚自动生成代码	代码生成器历史	POST
74	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/getSysHistory	查询回滚记录	代码生成器历史	POST
75	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/delSysHistory	删除回滚记录	代码生成器历史	POST
76	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/autoCode/addFunc	增加模板方法	代码生成器历史	POST
77	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionaryDetail/updateSysDictionaryDetail	更新字典内容	系统字典详情	PUT
78	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionaryDetail/createSysDictionaryDetail	新增字典内容	系统字典详情	POST
79	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionaryDetail/deleteSysDictionaryDetail	删除字典内容	系统字典详情	DELETE
80	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionaryDetail/findSysDictionaryDetail	根据ID获取字典内容	系统字典详情	GET
81	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionaryDetail/getSysDictionaryDetailList	获取字典内容列表	系统字典详情	GET
82	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionary/createSysDictionary	新增字典	系统字典	POST
83	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionary/deleteSysDictionary	删除字典	系统字典	DELETE
84	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionary/updateSysDictionary	更新字典	系统字典	PUT
85	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionary/findSysDictionary	根据ID获取字典（建议选择）	系统字典	GET
86	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysDictionary/getSysDictionaryList	获取字典列表	系统字典	GET
87	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysOperationRecord/createSysOperationRecord	新增操作记录	操作记录	POST
88	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysOperationRecord/findSysOperationRecord	根据ID获取操作记录	操作记录	GET
89	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysOperationRecord/getSysOperationRecordList	获取操作记录列表	操作记录	GET
90	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysOperationRecord/deleteSysOperationRecord	删除操作记录	操作记录	DELETE
91	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysOperationRecord/deleteSysOperationRecordByIds	批量删除操作历史	操作记录	DELETE
92	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/simpleUploader/upload	插件版分片上传	断点续传(插件版)	POST
93	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/simpleUploader/checkFileMd5	文件完整度验证	断点续传(插件版)	GET
94	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/simpleUploader/mergeFileMd5	上传完成合并文件	断点续传(插件版)	GET
95	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/email/emailTest	发送测试邮件	email	POST
96	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/email/sendEmail	发送邮件	email	POST
97	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authorityBtn/setAuthorityBtn	设置按钮权限	按钮权限	POST
98	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authorityBtn/getAuthorityBtn	获取已有按钮权限	按钮权限	POST
99	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/authorityBtn/canRemoveAuthorityBtn	删除按钮	按钮权限	POST
100	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/createSysExportTemplate	新增导出模板	导出模板	POST
101	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/deleteSysExportTemplate	删除导出模板	导出模板	DELETE
102	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/deleteSysExportTemplateByIds	批量删除导出模板	导出模板	DELETE
103	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/updateSysExportTemplate	更新导出模板	导出模板	PUT
104	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/findSysExportTemplate	根据ID获取导出模板	导出模板	GET
105	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/getSysExportTemplateList	获取导出模板列表	导出模板	GET
106	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/exportExcel	导出Excel	导出模板	GET
107	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/exportTemplate	下载模板	导出模板	GET
108	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysExportTemplate/importExcel	导入Excel	导出模板	POST
109	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/info/createInfo	新建公告	公告	POST
110	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/info/deleteInfo	删除公告	公告	DELETE
111	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/info/deleteInfoByIds	批量删除公告	公告	DELETE
112	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/info/updateInfo	更新公告	公告	PUT
113	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/info/findInfo	根据ID获取公告	公告	GET
114	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/info/getInfoList	获取公告列表	公告	GET
115	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysParams/createSysParams	新建参数	参数管理	POST
116	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysParams/deleteSysParams	删除参数	参数管理	DELETE
117	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysParams/deleteSysParamsByIds	批量删除参数	参数管理	DELETE
118	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysParams/updateSysParams	更新参数	参数管理	PUT
119	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysParams/findSysParams	根据ID获取参数	参数管理	GET
120	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysParams/getSysParamsList	获取参数列表	参数管理	GET
121	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/sysParams/getSysParam	获取参数列表	参数管理	GET
122	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/attachmentCategory/getCategoryList	分类列表	媒体库分类	GET
123	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/attachmentCategory/addCategory	添加/编辑分类	媒体库分类	POST
124	2025-07-05 10:16:11.56073+00	2025-07-05 10:16:11.56073+00	\N	/attachmentCategory/deleteCategory	删除分类	媒体库分类	POST
125	2025-07-05 11:21:43.737396+00	2025-07-05 11:21:43.737396+00	\N	/nesma/project	创建nesma项目	NESMA	POST
126	2025-07-05 11:22:26.293142+00	2025-07-05 11:22:26.293142+00	\N	/nesma/project	删除nesma项目	NESMA	DELETE
127	2025-07-05 11:23:05.894896+00	2025-07-05 11:23:05.894896+00	\N	/nesma/project/delete-batch	批量删除nesma项目	NESMA	DELETE
128	2025-07-05 11:23:27.062338+00	2025-07-05 11:23:27.062338+00	\N	/nesma/project	更新nesma项目	NESMA	PUT
129	2025-07-05 11:23:52.22018+00	2025-07-05 11:23:52.22018+00	\N	/nesma/project/archive	归档项目	NESMA	POST
130	2025-07-05 11:24:16.213367+00	2025-07-05 11:24:16.213367+00	\N	/nesma/project/restore	恢复项目	NESMA	POST
132	2025-07-05 11:25:00.492912+00	2025-07-05 11:25:00.492912+00	\N	/nesma/project/list	获取nesma列表	NESMA	GET
133	2025-07-05 11:25:33.069299+00	2025-07-05 11:25:33.069299+00	\N	/nesma/project/stats	获取nesma项目统计信息	NESMA	GET
131	2025-07-05 11:24:42.988004+00	2025-07-05 11:28:45.174297+00	\N	/nesma/project/:id	获取nesma详情	NESMA	GET
134	2025-07-05 12:05:06.704786+00	2025-07-05 12:05:06.704786+00	\N	/nesma/requirement	创建项目需求	NESMA	POST
135	2025-07-05 12:08:59.665918+00	2025-07-05 12:08:59.665918+00	\N	/nesma/requirement	更新需求	NESMA	PUT
136	2025-07-05 12:08:59.671781+00	2025-07-05 12:08:59.671781+00	\N	/nesma/requirement/:id	删除需求	NESMA	DELETE
137	2025-07-05 12:08:59.708439+00	2025-07-05 12:08:59.708439+00	\N	/nesma/requirement/:id	获取需求详情	NESMA	GET
138	2025-07-05 12:08:59.710063+00	2025-07-05 12:08:59.710063+00	\N	/nesma/requirement/list	获取需求列表	NESMA	GET
139	2025-07-05 12:08:59.711798+00	2025-07-05 12:08:59.711798+00	\N	/nesma/requirement/tree	获取需求树	NESMA	GET
140	2025-07-05 12:08:59.713508+00	2025-07-05 12:08:59.713508+00	\N	/nesma/requirement/stats	获取需求统计	NESMA	GET
141	2025-07-05 12:08:59.714558+00	2025-07-05 12:08:59.714558+00	\N	/nesma/requirement/parent-options	获取父需求选项	NESMA	GET
143	2025-07-05 12:08:59.716472+00	2025-07-05 12:08:59.716472+00	\N	/nesma/requirement/import-excel	Excel导入	NESMA	POST
144	2025-07-05 12:08:59.717601+00	2025-07-05 12:08:59.717601+00	\N	/nesma/requirement/batch-update-order	批量更新排序	NESMA	POST
145	2025-07-05 12:08:59.719462+00	2025-07-05 12:08:59.719462+00	\N	/nesma/requirement/move	移动需求	NESMA	POST
142	2025-07-05 12:08:59.7156+00	2025-07-09 16:18:31.583344+00	\N	/nesma/requirement/batch-delete	批量删除	NESMA	POST
146	2025-07-05 12:16:35.308511+00	2025-07-05 12:16:35.308511+00	2025-07-05 12:26:24.932964+00	/nesma/requirement/:projectId	获取项目下需求分析统计	NESMA	GET
147	2025-07-06 04:08:12.636886+00	2025-07-06 04:08:12.636886+00	\N	/knowledge/entry	创建知识条目	知识库	POST
148	2025-07-06 04:08:12.643538+00	2025-07-06 04:08:12.643538+00	\N	/knowledge/entry/:id	更新知识条目  	知识库	PUT
149	2025-07-06 04:08:12.650985+00	2025-07-06 04:08:12.650985+00	\N	/knowledge/entry/:id	删除知识条目	知识库	DELETE
150	2025-07-06 04:08:12.653658+00	2025-07-06 04:08:12.653658+00	\N	/knowledge/entry/:id	获取知识条目	知识库	GET
151	2025-07-06 04:08:12.655306+00	2025-07-06 04:08:12.655306+00	\N	/knowledge/entries	列出知识条目	知识库	GET
152	2025-07-06 04:08:12.656809+00	2025-07-06 04:08:12.656809+00	\N	/knowledge/entries/search	搜索知识条目	知识库	POST
153	2025-07-06 04:08:12.658765+00	2025-07-06 04:08:12.658765+00	\N	/knowledge/rule	创建知识规则	知识库	POST
154	2025-07-06 04:08:12.659578+00	2025-07-06 04:08:12.659578+00	\N	/knowledge/rule/:id	更新知识规则	知识库	PUT
155	2025-07-06 04:08:12.661684+00	2025-07-06 04:08:12.661684+00	\N	/knowledge/rule/:id	删除知识规则	知识库	DELETE
156	2025-07-06 04:08:12.662803+00	2025-07-06 04:08:12.662803+00	\N	/knowledge/rule/:id	获取知识规则	知识库	GET
157	2025-07-06 04:08:12.664275+00	2025-07-06 04:08:12.664275+00	\N	/knowledge/rules	列出知识规则	知识库	GET
158	2025-07-06 04:08:12.665079+00	2025-07-06 04:08:12.665079+00	\N	/knowledge/case	创建案例研究	知识库	POST
159	2025-07-06 04:08:12.667012+00	2025-07-06 04:08:12.667012+00	\N	/knowledge/case/:id	更新案例研究	知识库	PUT
160	2025-07-06 04:08:12.668008+00	2025-07-06 04:08:12.668008+00	\N	/knowledge/case/:id	删除案例研究	知识库	DELETE
161	2025-07-06 04:08:12.669648+00	2025-07-06 04:08:12.669648+00	\N	/knowledge/case/:id	获取案例研究	知识库	GET
162	2025-07-06 04:08:12.670829+00	2025-07-06 04:08:12.670829+00	\N	/knowledge/cases	列出案例研究	知识库	GET
163	2025-07-06 04:08:12.67183+00	2025-07-06 04:08:12.67183+00	\N	/knowledge/cases/search	搜索案例研究	知识库	POST
164	2025-07-06 04:08:12.673425+00	2025-07-06 04:08:12.673425+00	\N	/knowledge/statistics	获取知识库统计	知识库	GET
165	2025-07-06 04:08:12.674715+00	2025-07-06 04:08:12.674715+00	\N	/knowledge/import/standard	导入NESMA标准知识	知识库	POST
166	2025-07-06 04:08:12.676232+00	2025-07-06 04:08:12.676232+00	\N	/knowledge/tags/popular	获取热门标签	知识库	GET
167	2025-07-06 04:08:12.677585+00	2025-07-06 04:08:12.677585+00	\N	/knowledge/recommended	获取推荐知识	知识库	GET
168	2025-07-06 05:18:30.10147+00	2025-07-06 05:18:30.10147+00	\N	/agent/register	注册Agent	Agent	POST
169	2025-07-06 05:18:30.105052+00	2025-07-06 05:18:30.105052+00	\N	/agent/:id	获取Agent信息	Agent	GET
170	2025-07-06 05:18:30.110092+00	2025-07-06 05:18:30.110092+00	\N	/agent/list	获取Agent列表	Agent	GET
171	2025-07-06 05:18:30.111601+00	2025-07-06 05:18:30.111601+00	\N	/agent/:id/status	更新Agent状态	Agent	PUT
172	2025-07-06 05:18:30.113114+00	2025-07-06 05:18:30.113114+00	\N	/agent/:id/heartbeat	Agent心跳	Agent	POST
173	2025-07-06 06:10:42.387389+00	2025-07-06 06:10:42.387389+00	\N	/agent/workflow	创建工作流	工作流	POST
174	2025-07-06 06:10:42.388392+00	2025-07-06 06:10:42.388392+00	\N	/agent/workflow/:id	获取工作流详情	工作流	GET
175	2025-07-06 06:10:42.388889+00	2025-07-06 06:10:42.388889+00	\N	/agent/workflows	工作流列表	工作流	GET
176	2025-07-06 06:10:42.389885+00	2025-07-06 06:10:42.389885+00	\N	/agent/workflow/:id/execute	执行工作流	工作流	POST
177	2025-07-06 06:10:42.391474+00	2025-07-06 06:10:42.391474+00	\N	/agent/execution/:id	获取执行状态	工作流	GET
178	2025-07-06 06:10:42.391979+00	2025-07-06 06:10:42.391979+00	\N	/agent/execution/:id/cancel	取消执行	工作流	POST
179	2025-07-06 06:10:42.393279+00	2025-07-06 06:10:42.393279+00	\N	/agent/statistics/workflow	工作流统计	工作流	GET
180	2025-07-06 15:16:50.340804+00	2025-07-06 15:16:50.340804+00	\N	/nesma/document	CreateDocument	文档管理	POST
181	2025-07-06 15:16:50.347598+00	2025-07-06 15:16:50.347598+00	\N	/nesma/document	UpdateDocument	文档管理	PUT
182	2025-07-06 15:16:50.353994+00	2025-07-06 15:16:50.353994+00	\N	/nesma/document/:id	DeleteDocument	文档管理	DELETE
183	2025-07-06 15:16:50.356744+00	2025-07-06 15:16:50.356744+00	\N	/nesma/document/:id	GetDocument	文档管理	GET
184	2025-07-06 15:16:50.358264+00	2025-07-06 15:16:50.358264+00	\N	/nesma/document/list	GetDocumentList	文档管理	GET
185	2025-07-06 15:16:50.361425+00	2025-07-06 15:16:50.361425+00	\N	/nesma/document/generate	GenerateDocument	文档管理	POST
186	2025-07-06 15:16:50.362731+00	2025-07-06 15:16:50.362731+00	\N	/nesma/document/batch-generate	BatchGenerateDocument	文档管理	POST
187	2025-07-06 15:16:50.364957+00	2025-07-06 15:16:50.364957+00	\N	/nesma/document/preview	PreviewDocument	文档管理	POST
188	2025-07-06 15:16:50.366951+00	2025-07-06 15:16:50.366951+00	\N	/nesma/document/download	DownloadDocument	文档管理	GET
189	2025-07-06 15:16:50.372037+00	2025-07-06 15:16:50.372037+00	\N	/nesma/document/progress	GetDocumentProgress	文档管理	GET
190	2025-07-06 15:16:50.379139+00	2025-07-06 15:16:50.379139+00	\N	/nesma/document/stats	GetDocumentStats	文档管理	GET
191	2025-07-06 15:16:50.379641+00	2025-07-06 15:16:50.379641+00	\N	/nesma/document/batch-delete	BatchDeleteDocuments	文档管理	POST
192	2025-07-06 15:16:50.379641+00	2025-07-06 15:16:50.379641+00	\N	/nesma/template	CreateTemplate	模板管理	POST
193	2025-07-06 15:16:50.381867+00	2025-07-06 15:16:50.381867+00	\N	/nesma/template	UpdateTemplate	模板管理	PUT
194	2025-07-06 15:16:50.383868+00	2025-07-06 15:16:50.383868+00	\N	/nesma/template/:id	DeleteTemplate	模板管理	DELETE
195	2025-07-06 15:16:50.389638+00	2025-07-06 15:16:50.389638+00	\N	/nesma/template/:id	GetTemplate	模板管理	GET
196	2025-07-06 15:16:50.390651+00	2025-07-06 15:16:50.390651+00	\N	/nesma/template/list	GetTemplateList	模板管理	GET
197	2025-07-06 15:16:50.392137+00	2025-07-06 15:16:50.392137+00	\N	/nesma/template/options	GetTemplateOptions	模板管理	GET
198	2025-07-06 15:16:50.393131+00	2025-07-06 15:16:50.393131+00	\N	/nesma/template/variables	GetTemplateVariables	模板管理	GET
199	2025-07-06 15:16:50.394552+00	2025-07-06 15:16:50.394552+00	\N	/nesma/template/:id/default	SetDefaultTemplate	模板管理	POST
200	2025-07-06 15:16:50.40312+00	2025-07-06 15:16:50.40312+00	\N	/nesma/template/:id/activate	ActivateTemplate	模板管理	POST
201	2025-07-06 15:16:50.409236+00	2025-07-06 15:16:50.409236+00	\N	/nesma/template/upload	UploadTemplate	模板管理	POST
202	2025-07-06 15:16:50.411766+00	2025-07-06 15:16:50.411766+00	\N	/nesma/template/batch-delete	BatchDeleteTemplates	模板管理	POST
203	2025-07-06 19:31:32.524306+00	2025-07-06 19:31:32.524306+00	\N	/nesma/recommendation/items	获取推荐列表	智能推荐	GET
204	2025-07-06 19:31:32.530628+00	2025-07-06 19:31:32.530628+00	\N	/nesma/recommendation/behavior	记录用户行为	智能推荐	POST
205	2025-07-06 19:31:32.537074+00	2025-07-06 19:31:32.537074+00	\N	/nesma/recommendation/feedback	记录推荐反馈	智能推荐	POST
206	2025-07-06 19:31:32.539087+00	2025-07-06 19:31:32.539087+00	\N	/nesma/recommendation/stats	获取推荐统计	智能推荐	GET
207	2025-07-06 19:31:32.541365+00	2025-07-06 19:31:32.541365+00	\N	/nesma/recommendation/personalized	个性化推荐	智能推荐	POST
208	2025-07-06 19:31:32.543439+00	2025-07-06 19:31:32.543439+00	\N	/nesma/recommendation/config	获取推荐配置	智能推荐	GET
209	2025-07-06 19:31:32.545144+00	2025-07-06 19:31:32.545144+00	\N	/ai/services/switch	切换默认AI服务	AI服务管理	POST
210	2025-07-06 19:31:32.546894+00	2025-07-06 19:31:32.546894+00	\N	/ai/services/test	测试AI服务	AI服务管理	POST
211	2025-07-06 19:31:32.548458+00	2025-07-06 19:31:32.548458+00	\N	/ai/services/circuit-breaker/reset	重置熔断器	AI服务管理	POST
212	2025-07-06 19:31:32.550339+00	2025-07-06 19:31:32.550339+00	\N	/ai/services	获取AI服务列表	AI服务管理	GET
213	2025-07-06 19:31:32.551976+00	2025-07-06 19:31:32.551976+00	\N	/ai/services/health	获取AI服务健康状态	AI服务管理	GET
214	2025-07-06 19:31:32.553713+00	2025-07-06 19:31:32.553713+00	\N	/ai/services/config	获取AI服务配置	AI服务管理	GET
215	2025-07-06 19:31:32.555411+00	2025-07-06 19:31:32.555411+00	\N	/ai/services/circuit-breaker/stats	获取熔断器统计信息	AI服务管理	GET
216	2025-07-06 19:31:32.556507+00	2025-07-06 19:31:32.556507+00	\N	/chat/send	发送消息	聊天功能	POST
217	2025-07-06 19:31:32.557991+00	2025-07-06 19:31:32.557991+00	\N	/chat/session	创建会话	聊天功能	POST
218	2025-07-06 19:31:32.559491+00	2025-07-06 19:31:32.559491+00	\N	/chat/session/:sessionId	更新会话	聊天功能	PUT
219	2025-07-06 19:31:32.561082+00	2025-07-06 19:31:32.561082+00	\N	/chat/session/:sessionId	删除会话	聊天功能	DELETE
220	2025-07-06 19:31:32.562809+00	2025-07-06 19:31:32.562809+00	\N	/chat/session/:sessionId/clear	清空会话消息	聊天功能	POST
221	2025-07-06 19:31:32.564357+00	2025-07-06 19:31:32.564357+00	\N	/chat/sessions	获取会话列表	聊天功能	GET
222	2025-07-06 19:31:32.56586+00	2025-07-06 19:31:32.56586+00	\N	/chat/session/:sessionId	获取会话详情	聊天功能	GET
223	2025-07-06 19:31:32.567371+00	2025-07-06 19:31:32.567371+00	\N	/chat/session/:sessionId/messages	获取会话消息	聊天功能	GET
224	2025-07-06 19:31:32.568383+00	2025-07-06 19:31:32.568383+00	\N	/chat/stats	获取聊天统计信息	聊天功能	GET
225	2025-07-06 19:31:32.569871+00	2025-07-06 19:31:32.569871+00	\N	/chat/session/:sessionId/export	导出聊天记录	聊天功能	GET
226	2025-07-06 19:31:32.570949+00	2025-07-06 19:31:32.570949+00	\N	/requirement-analysis/analyze	分析需求	需求分析	POST
227	2025-07-06 19:31:32.572455+00	2025-07-06 19:31:32.572455+00	\N	/requirement-analysis/batch-analyze	批量分析需求	需求分析	POST
228	2025-07-06 19:31:32.573491+00	2025-07-06 19:31:32.573491+00	\N	/requirement-analysis/history	获取分析历史	需求分析	GET
229	2025-07-06 19:31:32.57449+00	2025-07-06 19:31:32.57449+00	\N	/requirement-analysis/result/:analysisId	获取分析结果	需求分析	GET
230	2025-07-06 19:31:32.57549+00	2025-07-06 19:31:32.57549+00	\N	/requirement-analysis/quick-analyze	快速分析需求	需求分析	POST
231	2025-07-06 19:31:32.576991+00	2025-07-06 19:31:32.576991+00	\N	/requirement-analysis/config	获取分析配置	需求分析	GET
232	2025-07-06 19:31:32.57799+00	2025-07-06 19:31:32.57799+00	\N	/requirement-analysis/stats	获取分析统计信息	需求分析	GET
233	2025-07-06 19:31:32.57899+00	2025-07-06 19:31:32.57899+00	\N	/requirement-analysis/compare	对比分析结果	需求分析	POST
234	2025-07-06 19:31:32.57999+00	2025-07-06 19:31:32.57999+00	\N	/nesma-evaluation/create	创建评估	NESMA评估	POST
235	2025-07-06 19:31:32.581774+00	2025-07-06 19:31:32.581774+00	\N	/nesma-evaluation/start	启动评估	NESMA评估	POST
236	2025-07-06 19:31:32.58382+00	2025-07-06 19:31:32.58382+00	\N	/nesma-evaluation/:id	获取评估详情	NESMA评估	GET
237	2025-07-06 19:31:32.584836+00	2025-07-06 19:31:32.584836+00	\N	/nesma-evaluation/list	获取评估列表	NESMA评估	GET
238	2025-07-06 19:31:32.586218+00	2025-07-06 19:31:32.586218+00	\N	/nesma-evaluation/function-points	获取功能点列表	NESMA评估	GET
239	2025-07-06 19:31:32.587345+00	2025-07-06 19:31:32.587345+00	\N	/nesma-evaluation/function-point	创建功能点	NESMA评估	POST
240	2025-07-06 19:31:32.587872+00	2025-07-06 19:31:32.587872+00	\N	/nesma-evaluation/function-point/:id	更新功能点	NESMA评估	PUT
241	2025-07-06 19:31:32.588372+00	2025-07-06 19:31:32.588372+00	\N	/nesma-evaluation/function-point/:id	删除功能点	NESMA评估	DELETE
242	2025-07-06 19:31:32.588867+00	2025-07-06 19:31:32.588867+00	\N	/nesma-evaluation/:id/stats	获取评估统计信息	NESMA评估	GET
243	2025-07-06 19:31:32.589874+00	2025-07-06 19:31:32.589874+00	\N	/nesma-evaluation/complexity-metrics	获取复杂度指标	NESMA评估	GET
244	2025-07-06 19:31:32.591466+00	2025-07-06 19:31:32.591466+00	\N	/nesma-evaluation/validation-items	获取验证项列表	NESMA评估	GET
245	2025-07-06 19:31:32.592462+00	2025-07-06 19:31:32.592462+00	\N	/nesma-evaluation/validation-item/:id	更新验证项	NESMA评估	PUT
246	2025-07-06 19:31:32.593466+00	2025-07-06 19:31:32.593466+00	\N	/nesma-evaluation/review	审核评估	NESMA评估	POST
247	2025-07-06 19:31:32.599918+00	2025-07-06 19:31:32.599918+00	\N	/nesma-evaluation/:id/export	导出评估报告	NESMA评估	GET
248	2025-07-06 19:31:32.601257+00	2025-07-06 19:31:32.601257+00	\N	/nesma-evaluation/project/:projectId/summary	获取项目评估摘要	NESMA评估	GET
249	2025-07-06 19:31:32.602387+00	2025-07-06 19:31:32.602387+00	\N	/nesma-evaluation/:id/recalculate	重新计算评估	NESMA评估	POST
250	2025-07-07 04:15:13.86106+00	2025-07-07 04:15:13.86106+00	\N	/monitoring/service/stats	获取服务性能统计	模型性能监控	GET
251	2025-07-07 04:15:13.868431+00	2025-07-07 04:15:13.868431+00	\N	/monitoring/error/analysis	获取错误分析	模型性能监控	GET
252	2025-07-07 04:15:13.875325+00	2025-07-07 04:15:13.875325+00	\N	/monitoring/token/stats	获取Token使用统计	模型性能监控	GET
253	2025-07-07 04:15:13.876824+00	2025-07-07 04:15:13.876824+00	\N	/monitoring/time-series	获取时间序列统计	模型性能监控	GET
254	2025-07-07 04:15:13.87846+00	2025-07-07 04:15:13.87846+00	\N	/monitoring/metrics	获取监控指标	模型性能监控	GET
255	2025-07-07 04:15:13.880344+00	2025-07-07 04:15:13.880344+00	\N	/monitoring/health	获取服务健康状态	模型性能监控	GET
256	2025-07-07 04:15:13.881818+00	2025-07-07 04:15:13.881818+00	\N	/monitoring/system/resources	获取系统资源使用情况	模型性能监控	GET
257	2025-07-07 04:15:13.884009+00	2025-07-07 04:15:13.884009+00	\N	/monitoring/model/comparison	获取AI模型性能对比	模型性能监控	GET
259	2025-07-07 04:15:13.887205+00	2025-07-07 04:15:13.887205+00	\N	/monitoring/cost/analysis	获取成本分析	模型性能监控	GET
260	2025-07-07 04:15:13.888201+00	2025-07-07 04:15:13.888201+00	\N	/monitoring/report/export	导出监控报告	模型性能监控	GET
258	2025-07-07 04:15:13.885514+00	2025-07-07 04:15:45.411667+00	\N	/monitoring/response-time/distribution	获取响应时间分布	模型性能监控	GET
261	2025-07-07 16:02:10.234598+00	2025-07-07 16:02:10.234598+00	\N	/agent/:id	正常显示Agent基本信息	Agent管理	GET
262	2025-07-07 16:02:10.24194+00	2025-07-07 16:02:10.24194+00	\N	/agent/:id/tasks	加载Agent任务列表	Agent管理	GET
263	2025-07-07 16:02:10.249304+00	2025-07-07 16:02:10.249304+00	\N	/agent/:id/logs	显示Agent运行日志	Agent管理	GET
264	2025-07-07 16:02:10.251754+00	2025-07-07 16:02:10.251754+00	\N	/monitoring/service/stats	获取服务性能统计	模型性能监控	GET
265	2025-07-07 16:55:28.772094+00	2025-07-07 16:55:28.772094+00	\N	/requirement/delete-by-condition	根据条件删除需求	需求	POST
266	2025-07-07 17:03:13.951073+00	2025-07-07 17:10:32.793594+00	\N	/nesma/requirement/max-version/2	获取项目最大版本号	需求管理	GET
267	2025-07-08 20:51:22.64941+00	2025-07-08 20:52:09.736133+00	\N	/nesma/recommendation/ai	大模型推荐	智能推荐	POST
268	2025-07-08 21:53:13.250543+00	2025-07-08 21:53:13.250543+00	\N	/nesma/recommendation/adopt	采纳推荐信息入库	智能推荐	POST
269	2025-07-09 00:22:19.866752+00	2025-07-09 00:22:19.866754+00	\N	/nesma/project-cycle	创建项目周期	项目周期管理	POST
270	2025-07-09 00:22:19.87395+00	2025-07-09 00:22:19.873952+00	\N	/nesma/project-cycle	更新项目周期	项目周期管理	PUT
271	2025-07-09 00:22:19.881385+00	2025-07-09 00:22:19.881387+00	\N	/nesma/project-cycle/:id	删除项目周期	项目周期管理	DELETE
272	2025-07-09 00:22:19.883229+00	2025-07-09 00:22:19.883231+00	\N	/nesma/project-cycle/status	更新周期状态	项目周期管理	PUT
273	2025-07-09 00:22:19.887013+00	2025-07-09 00:22:19.887015+00	\N	/nesma/project-cycle/:id	获取周期	项目周期管理	GET
274	2025-07-09 00:22:19.888902+00	2025-07-09 00:22:19.888904+00	\N	/nesma/project-cycle/list	获取周期列表	项目周期管理	GET
275	2025-07-09 00:22:19.890886+00	2025-07-09 00:22:19.890888+00	\N	/nesma/project-cycle/project/:projectId	获取项目的所有周期	项目周期管理	GET
276	2025-07-09 16:22:42.274217+00	2025-07-09 16:22:42.274217+00	\N	/nesma/project-cycle/set-active	设置为活跃周期	项目周期管理	POST
277	2025-07-09 16:33:13.696773+00	2025-07-09 16:33:13.696773+00	\N	/nesma/analysis/analyze	一键分析优化	智能分析	POST
278	2025-07-09 16:35:57.362559+00	2025-07-09 16:35:57.362559+00	\N	/nesma/analysis/progress/:taskId	获取分析进度	智能分析	GET
279	2025-07-09 16:37:10.135282+00	2025-07-09 16:37:10.135282+00	\N	/nesma/requirement/max-version/:projectId	获取项目最大版本号	项目管理	GET
280	2025-07-10 03:49:48.60123+00	2025-07-10 03:49:48.601231+00	\N	/nesma/generator/level4/generate	生成四级功能点	功能点生成	POST
281	2025-07-10 03:49:48.607856+00	2025-07-10 03:49:48.607857+00	\N	/nesma/generator/level4/create	创建四级功能点	功能点生成	POST
282	2025-07-10 03:49:48.614492+00	2025-07-10 03:49:48.614494+00	\N	/nesma/generator/level4/batch-create	批量创建	功能点生成	POST
283	2025-07-10 03:49:48.616444+00	2025-07-10 03:49:48.616445+00	\N	/nesma/generator/level4/stats/{cycleId}	生成统计	功能点生成	GET
284	2025-07-10 03:49:48.617954+00	2025-07-10 03:49:48.617956+00	\N	/nesma/generator/level4/validate	验证建议	功能点生成	POST
285	2025-07-10 03:49:48.619149+00	2025-07-10 03:49:48.61915+00	\N	/nesma/generator/description/generate	生成需求描述	描述生成	POST
286	2025-07-10 03:49:48.620317+00	2025-07-10 03:49:48.620318+00	\N	/nesma/generator/description/apply	应用生成的描述	描述生成	POST
287	2025-07-10 03:49:48.62177+00	2025-07-10 03:49:48.621772+00	\N	/nesma/generator/description/batch-generate	批量生成描述	描述生成	POST
288	2025-07-10 03:49:48.622722+00	2025-07-10 03:49:48.622724+00	\N	/nesma/generator/description/history/{cycleId}	获取生成历史	描述生成	GET
289	2025-07-10 03:49:48.623809+00	2025-07-10 03:49:48.62381+00	\N	/nesma/generator/description/stats/{cycleId}	获取生成统计	描述生成	GET
290	2025-07-10 03:49:48.625916+00	2025-07-10 03:49:48.625917+00	\N	/nesma/generator/description/preview	预览描述生成	描述生成	POST
291	2025-07-10 03:49:48.627325+00	2025-07-10 03:49:48.627326+00	\N	/nesma/generator/description/validate	验证描述质量	描述生成	POST
292	2025-07-10 03:49:48.629147+00	2025-07-10 03:49:48.629149+00	\N	/nesma/generator/mermaid/generate	生成Mermaid流程图	流程图生成	POST
293	2025-07-10 03:49:48.6306+00	2025-07-10 03:49:48.630601+00	\N	/nesma/generator/mermaid/apply	应用Mermaid流程图	流程图生成	POST
294	2025-07-10 03:49:48.631574+00	2025-07-10 03:49:48.631575+00	\N	/nesma/generator/mermaid/batch-generate	批量生成	流程图生成	POST
295	2025-07-10 03:49:48.633501+00	2025-07-10 03:49:48.633502+00	\N	/nesma/generator/mermaid/history/{cycleId}	获取生成历史	流程图生成	GET
296	2025-07-10 03:49:48.634646+00	2025-07-10 03:49:48.634647+00	\N	/nesma/generator/mermaid/stats/{cycleId}	获取生成统计	流程图生成	GET
297	2025-07-10 03:49:48.635927+00	2025-07-10 03:49:48.635928+00	\N	/nesma/generator/mermaid/validate	验证流程图	流程图生成	POST
298	2025-07-10 03:49:48.637072+00	2025-07-10 03:49:48.637074+00	\N	/nesma/generator/mermaid/preview	预览流程图	流程图生成	POST
299	2025-07-10 13:21:50.32071+00	2025-07-10 13:21:50.320711+00	\N	/monitoring/service/stats	获取服务统计信息	服务监控	GET
300	2025-07-10 13:21:50.327808+00	2025-07-10 13:21:50.32781+00	\N	/monitoring/error/analysis	获取错误分析	服务监控	GET
301	2025-07-10 13:21:50.333581+00	2025-07-10 13:21:50.333583+00	\N	/monitoring/token/stats	获取令牌使用统计	服务监控	GET
302	2025-07-10 13:21:50.335202+00	2025-07-10 13:21:50.335203+00	\N	/monitoring/time-series	获取时间序列统计	服务监控	GET
303	2025-07-10 13:21:50.336932+00	2025-07-10 13:21:50.336933+00	\N	/monitoring/metrics	获取监控指标	服务监控	GET
304	2025-07-10 13:21:50.338448+00	2025-07-10 13:21:50.33845+00	\N	/monitoring/health	获取服务健康状态	服务监控	GET
305	2025-07-10 13:21:50.34011+00	2025-07-10 13:21:50.34011+00	\N	/monitoring/system/resources	获取系统资源	服务监控	GET
306	2025-07-10 13:21:50.341648+00	2025-07-10 13:21:50.341649+00	\N	/monitoring/model/comparison	获取模型性能对比	服务监控	GET
307	2025-07-10 13:21:50.343037+00	2025-07-10 13:21:50.343038+00	\N	/monitoring/response-time/distribution	获取响应时间分布	服务监控	GET
308	2025-07-10 13:21:50.344373+00	2025-07-10 13:21:50.344374+00	\N	/monitoring/cost/analysis	获取成本分析	服务监控	GET
309	2025-07-10 13:21:50.346167+00	2025-07-10 13:21:50.346168+00	\N	/monitoring/report/export	导出监控报告	服务监控	GET
310	2025-07-10 13:21:50.348184+00	2025-07-10 13:21:50.348184+00	\N	/nesma/analysis/level3/analyze	分析三级需求	三级需求分析	POST
311	2025-07-10 13:21:50.349325+00	2025-07-10 13:21:50.349326+00	\N	/nesma/analysis/level3/apply-optimization	应用优化建议	三级需求分析	POST
312	2025-07-10 13:21:50.350389+00	2025-07-10 13:21:50.35039+00	\N	/nesma/analysis/level3/create-expansion	创建扩展需求	三级需求分析	POST
313	2025-07-10 13:21:50.351644+00	2025-07-10 13:21:50.351645+00	\N	/nesma/analysis/level3/batch-apply-optimizations	批量应用优化	三级需求分析	POST
314	2025-07-10 13:21:50.352901+00	2025-07-10 13:21:50.352902+00	\N	/nesma/analysis/level3/batch-create-expansions	批量创建扩展	三级需求分析	POST
315	2025-07-10 13:21:50.354062+00	2025-07-10 13:21:50.354063+00	\N	/nesma/analysis/level3/history/:cycleId	获取分析历史	三级需求分析	GET
316	2025-07-10 13:21:50.35522+00	2025-07-10 13:21:50.35522+00	\N	/nesma/analysis/level3/stats/:cycleId	获取分析统计	三级需求分析	GET
317	2025-07-10 13:21:50.356567+00	2025-07-10 13:21:50.356567+00	\N	/nesma/generator/level4/generate	生成四级功能点	功能点生成	POST
318	2025-07-10 13:21:50.357893+00	2025-07-10 13:21:50.357893+00	\N	/nesma/generator/level4/create	创建四级需求	功能点生成	POST
319	2025-07-10 13:21:50.35879+00	2025-07-10 13:21:50.358791+00	\N	/nesma/generator/level4/batch-create	批量创建四级需求	功能点生成	POST
320	2025-07-10 13:21:50.359563+00	2025-07-10 13:21:50.359564+00	\N	/nesma/generator/level4/history/:cycleId	获取四级生成历史	功能点生成	GET
321	2025-07-10 13:21:50.360634+00	2025-07-10 13:21:50.360635+00	\N	/nesma/generator/level4/stats/:cycleId	获取四级生成统计	功能点生成	GET
322	2025-07-10 13:21:50.361519+00	2025-07-10 13:21:50.36152+00	\N	/nesma/generator/level4/validate	验证四级建议	功能点生成	POST
323	2025-07-11 11:45:16.127786+00	2025-07-11 11:45:16.127788+00	\N	/:id/ai-analysis	单个需求的AI分析	AI分析	GET
324	2025-07-11 11:45:16.135048+00	2025-07-11 11:45:16.13505+00	\N	/ai-analysis-stats	项目AI分析统计	AI分析	GET
325	2025-07-11 11:45:16.139257+00	2025-07-11 11:45:16.139259+00	\N	/compare-versions	版本对比	AI分析	POST
326	2025-07-11 11:45:16.141275+00	2025-07-11 11:45:16.141275+00	\N	/ai-status	更新AI状态	AI分析	PUT
327	2025-07-11 11:45:16.14249+00	2025-07-11 11:45:16.142491+00	\N	/ai-optimized	AI优化后的需求列表	AI分析	GET
\.


--
-- Data for Name: sys_authorities; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_authorities (created_at, updated_at, deleted_at, authority_id, authority_name, parent_id, default_router) FROM stdin;
2025-07-05 11:15:59.091074+00	2025-07-13 07:02:10.078374+00	\N	999	管理员	0	workspace
\.


--
-- Data for Name: sys_authority_btns; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_authority_btns (authority_id, sys_menu_id, sys_base_menu_btn_id) FROM stdin;
\.


--
-- Data for Name: sys_authority_menus; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_authority_menus (sys_base_menu_id, sys_authority_authority_id) FROM stdin;
65	999
64	999
66	999
3	999
10	999
11	999
12	999
13	999
14	999
15	999
16	999
4	999
6	999
24	999
25	999
20	999
23	999
21	999
22	999
26	999
27	999
28	999
29	999
8	999
40	999
62	999
54	999
60	999
59	999
58	999
56	999
57	999
53	999
\.


--
-- Data for Name: sys_auto_code_histories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_auto_code_histories (id, created_at, updated_at, deleted_at, table_name, package, request, struct_name, abbreviation, business_db, description, templates, "Injections", flag, api_ids, menu_id, export_template_id, package_id) FROM stdin;
\.


--
-- Data for Name: sys_auto_code_packages; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_auto_code_packages (id, created_at, updated_at, deleted_at, "desc", label, template, package_name, module) FROM stdin;
\.


--
-- Data for Name: sys_base_menu_btns; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_base_menu_btns (id, created_at, updated_at, deleted_at, name, "desc", sys_base_menu_id) FROM stdin;
\.


--
-- Data for Name: sys_base_menu_parameters; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_base_menu_parameters (id, created_at, updated_at, deleted_at, sys_base_menu_id, type, key, value) FROM stdin;
\.


--
-- Data for Name: sys_base_menus; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_base_menus (id, created_at, updated_at, deleted_at, menu_level, parent_id, path, name, hidden, component, sort, active_name, keep_alive, default_menu, title, icon, close_tab, transition_type) FROM stdin;
4	2025-07-05 10:16:11.624647+00	2025-07-05 10:16:11.624647+00	\N	0	0	person	person	t	view/person/person.vue	4		f	f	个人信息	message	f	
8	2025-07-05 10:16:11.624647+00	2025-07-05 10:16:11.624647+00	\N	0	0	state	state	f	view/system/state.vue	8		f	f	服务器状态	cloudy	f	
10	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	3	authority	authority	f	view/superAdmin/authority/authority.vue	1		f	f	角色管理	avatar	f	
11	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	3	menu	menu	f	view/superAdmin/menu/menu.vue	2		t	f	菜单管理	tickets	f	
12	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	3	api	api	f	view/superAdmin/api/api.vue	3		t	f	api管理	platform	f	
13	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	3	user	user	f	view/superAdmin/user/user.vue	4		f	f	用户管理	coordinate	f	
14	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	3	dictionary	dictionary	f	view/superAdmin/dictionary/sysDictionary.vue	5		f	f	字典管理	notebook	f	
15	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	3	operation	operation	f	view/superAdmin/operation/sysOperationRecord.vue	6		f	f	操作历史	pie-chart	f	
16	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	3	sysParams	sysParams	f	view/superAdmin/params/sysParams.vue	7		f	f	参数管理	compass	f	
22	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	6	system	system	f	view/systemTools/system/system.vue	4		f	f	系统配置	operation	f	
24	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	\N	1	6	autoCodeEdit/:id	autoCodeEdit	t	view/systemTools/autoCode/index.vue	0		f	f	自动化代码-${id}	magic-stick	f	
7	2025-07-05 10:16:11.624647+00	2025-07-05 10:16:11.624647+00	2025-07-05 10:16:31.039106+00	0	0	https://www.gin-vue-admin.com	https://www.gin-vue-admin.com	f	/	0		f	f	官方网站	customer-gva	f	
2	2025-07-05 10:16:11.624647+00	2025-07-05 10:16:11.624647+00	2025-07-05 10:16:33.862005+00	0	0	about	about	f	view/about/index.vue	9		f	f	关于我们	info-filled	f	
25	2025-07-05 10:16:11.626299+00	2025-07-05 11:56:49.774912+00	\N	1	6	autoPkg	autoPkg	t	view/systemTools/autoPkg/autoPkg.vue	0		f	f	模板配置	folder	f	
20	2025-07-05 10:16:11.626299+00	2025-07-05 11:56:54.390908+00	\N	1	6	autoCode	autoCode	t	view/systemTools/autoCode/index.vue	1		t	f	代码生成器	cpu	f	
37	2025-07-05 12:02:59.957644+00	2025-07-06 04:00:18.24506+00	2025-07-11 17:24:38.859167+00	0	35	requirement	requirement	f	view/nesma/requirement/index.vue	2		f	f	需求管理	soccer	f	
3	2025-07-05 10:16:11.624647+00	2025-07-10 19:50:55.598665+00	\N	0	0	admin	superAdmin	f	view/superAdmin/index.vue	99		f	f	系统设置	user	f	
40	2025-07-06 05:19:02.328984+00	2025-07-16 20:17:01.43851+00	\N	0	54	agent	agent	f	view/nesma/agent/index.vue	7		f	f	Agent管理	magnet	f	
23	2025-07-05 10:16:11.626299+00	2025-07-05 11:56:58.335549+00	\N	1	6	autoCodeAdmin	autoCodeAdmin	t	view/systemTools/autoCodeAdmin/index.vue	2		f	f	自动化代码管理	magic-stick	f	
21	2025-07-05 10:16:11.626299+00	2025-07-05 11:57:02.411176+00	\N	1	6	formCreate	formCreate	t	view/systemTools/formCreate/index.vue	3		t	f	表单生成器	magic-stick	f	
26	2025-07-05 10:16:11.626299+00	2025-07-05 11:57:07.144187+00	\N	1	6	exportTemplate	exportTemplate	t	view/systemTools/exportTemplate/exportTemplate.vue	5		f	f	导出模板	reading	f	
27	2025-07-05 10:16:11.626299+00	2025-07-05 11:57:11.323018+00	\N	1	6	picture	picture	t	view/systemTools/autoCode/picture.vue	6		f	f	AI页面绘制	picture-filled	f	
28	2025-07-05 10:16:11.626299+00	2025-07-05 11:57:15.543283+00	\N	1	6	mcpTool	mcpTool	t	view/systemTools/autoCode/mcp.vue	7		f	f	Mcp Tools模板	magnet	f	
29	2025-07-05 10:16:11.626299+00	2025-07-05 11:57:22.323712+00	\N	1	6	mcpTest	mcpTest	t	view/systemTools/autoCode/mcpTest.vue	7		f	f	Mcp Tools测试	partly-cloudy	f	
1	2025-07-05 10:16:11.624647+00	2025-07-10 19:49:08.686929+00	2025-07-11 14:50:33.532237+00	0	0	dashboard	dashboard	f	view/dashboard/index.vue	0		f	f	仪表盘	odometer	f	
6	2025-07-05 10:16:11.624647+00	2025-07-06 03:26:23.748299+00	\N	0	3	systemTools	systemTools	f	view/systemTools/index.vue	5		f	f	系统工具	tools	f	
44	2025-07-06 14:50:32.120418+00	2025-07-06 19:22:56.992142+00	2025-07-16 20:12:55.933502+00	0	35	document	document	f	view/nesma/document/index.vue	6		f	f	文档生成	document	f	
39	2025-07-06 03:58:51.660589+00	2025-07-06 03:58:51.660589+00	2025-07-06 04:00:01.254891+00	0	38	knowledgeIndex	knowledgeIndex	f	view/nesma/knowledge/index.vue	0		f	f	知识库管理		f	
38	2025-07-06 03:58:07.567941+00	2025-07-06 04:00:23.604071+00	2025-07-11 17:24:40.819038+00	0	35	knowledge	knowledge	f	view/nesma/knowledge/index.vue	3		f	f	知识库	knife-fork	f	
46	2025-07-06 19:23:35.496083+00	2025-07-06 19:23:35.496083+00	2025-07-11 20:15:29.119893+00	0	35	recommendation	recommendation	f	view/nesma/recommendation/index.vue	0		f	f	智能推荐	document-remove	f	
42	2025-07-06 06:20:30.612877+00	2025-07-06 06:20:30.612877+00	2025-07-06 10:51:53.400383+00	0	41	workflow	workflow	f	view/nesma/agent/components/WorkflowManagement.vue	0		f	f	可视化编辑路由		f	
45	2025-07-06 18:12:05.482138+00	2025-07-16 20:13:24.495174+00	2025-07-16 20:18:03.87138+00	0	54	evaluation	evaluation	f	view/nesma/evaluation/index.vue	7		f	f	评估引擎	bottom-left	f	
43	2025-07-06 10:49:30.878621+00	2025-07-06 12:51:39.079288+00	2025-07-06 12:52:36.603595+00	0	41	designer	designer	f	view/nesma/workflow/designer.vue	1		f	f	可视化流程设计	add-location	f	
41	2025-07-06 06:12:46.332016+00	2025-07-06 13:38:03.243226+00	2025-07-16 20:12:59.538792+00	0	35	workflow	workflow	f	view/nesma/workflow/index.vue	5		f	f	流程编排	bowl	f	
17	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:33.684278+00	1	5	upload	upload	f	view/example/upload/upload.vue	5		f	f	媒体库（上传下载）	upload	f	
36	2025-07-05 11:33:22.794357+00	2025-07-05 11:52:30.863815+00	2025-07-11 17:24:36.62225+00	0	35	project	project	f	view/nesma/project/index.vue	1		f	f	项目管理	apple	f	
32	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:46.841277+00	1	9	pubPlug	pubPlug	f	view/systemTools/pubPlug/pubPlug.vue	3		f	f	打包插件	files	f	
18	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:36.376693+00	1	5	breakpoint	breakpoint	f	view/example/breakpoint/breakpoint.vue	6		f	f	断点续传	upload-filled	f	
19	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:39.313227+00	1	5	customer	customer	f	view/example/customer/customer.vue	7		f	f	客户列表（资源示例）	avatar	f	
5	2025-07-05 10:16:11.624647+00	2025-07-16 16:43:56.938953+00	2025-07-16 20:13:42.01724+00	0	0	example	example	f	view/example/index.vue	7		f	f	示例文件	management	f	
33	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:48.449097+00	1	9	plugin-email	plugin-email	f	plugin/email/view/index.vue	4		f	f	邮件插件	message	f	
34	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:50.865209+00	1	9	anInfo	anInfo	f	plugin/announcement/view/info.vue	5		f	f	公告管理[示例]	scaleToOriginal	f	
31	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:53.161283+00	1	9	installPlugin	installPlugin	f	view/systemTools/installPlugin/index.vue	1		f	f	插件安装	box	f	
9	2025-07-05 10:16:11.624647+00	2025-07-05 11:56:06.749096+00	2025-07-16 20:13:57.083765+00	0	0	plugin	plugin	t	view/routerHolder.vue	6		f	f	插件系统	cherry	f	
50	2025-07-10 13:26:03.968591+00	2025-07-10 13:26:03.968591+00	2025-07-10 19:34:55.269757+00	0	49	ai-dashboard	ai-dashboard	f	view/nesma/ai/dashboard.vue	0		f	f	AI服务监控	monitor	f	
51	2025-07-10 13:26:28.177497+00	2025-07-10 13:26:28.177497+00	2025-07-10 19:34:58.757279+00	0	49	knowledge-graph	knowledge-graph	f	view/nesma/ai/knowledge-graph.vue	0		f	f	知识图谱	connection	f	
52	2025-07-10 13:27:05.456555+00	2025-07-10 13:27:05.456555+00	2025-07-10 19:35:01.456888+00	0	49	intelligent-analysis	intelligent-analysis	f	view/nesma/requirement/intelligent-analysis.vue	0		f	f	智能需求分析	magic-stick	f	
49	2025-07-10 13:25:24.497134+00	2025-07-10 13:25:24.497134+00	2025-07-10 19:35:03.452976+00	0	0	routerHolder	routerHolder	f	view/routerHolder.vue	0		f	f	AI智能服务	cpu	f	
48	2025-07-07 04:11:07.237925+00	2025-07-07 04:11:07.237925+00	2025-07-11 17:24:43.87786+00	0	35	monitoring	monitoring	f	view/nesma/monitoring/index.vue	0		f	f	性能监控	monitor	f	
55	2025-07-10 19:36:20.697717+00	2025-07-11 20:20:28.525755+00	2025-07-16 20:14:28.596414+00	0	0	routerHolder	routerHolder	f	view/routerHolder.vue	3		f	f	AI服务	magic-stick	f	
66	2025-07-13 07:01:07.919268+00	2025-07-16 20:14:40.740031+00	\N	0	54	level3	level3	t	view/nesma/analyzer/level3.vue	0		f	f	L3需求分析器	clock	f	
64	2025-07-13 07:00:08.69152+00	2025-07-16 20:14:45.74493+00	\N	0	54	mermaid	mermaid	t	view/nesma/generator/mermaid.vue	0		f	f	Mermaid流程图生成器	add-location	f	
65	2025-07-13 07:00:40.152518+00	2025-07-16 20:14:53.027526+00	\N	0	54	level4	level4	t	view/nesma/generator/level4.vue	0		f	f	L4功能点生成器	camera-filled	f	
62	2025-07-10 19:42:59.10245+00	2025-07-16 20:16:31.565733+00	\N	0	54	chat	chat	f	view/nesma/chat/index.vue	3		f	f	AI对话助手	chat-square	f	
56	2025-07-10 19:38:52.647966+00	2025-07-11 20:07:45.377206+00	\N	0	54	project	project	f	view/nesma/project/index.vue	1		f	f	项目管理	folder-opened	f	
58	2025-07-10 19:40:37.511003+00	2025-07-16 20:16:39.083308+00	\N	0	54	nesmaEvaluation	nesmaEvaluation	f	view/nesma/evaluation/index.vue	4		f	f	NESMA评估	checked	f	
59	2025-07-10 19:41:08.80755+00	2025-07-16 20:16:50.407515+00	\N	0	54	document	nesmaDocument	f	view/nesma/document/index.vue	5		f	f	文档生成	files	f	
60	2025-07-10 19:41:50.351806+00	2025-07-16 20:17:19.452791+00	\N	0	54	knowledge	nesmaKnowledge	f	view/nesma/knowledge/index.vue	6		f	f	知识库	reading	f	
57	2025-07-10 19:39:48.800193+00	2025-07-11 20:12:40.560054+00	\N	0	54	requirement	requirement	f	view/nesma/requirement/index.vue	2		f	f	需求管理	document	f	
47	2025-07-07 04:10:39.015932+00	2025-07-07 04:10:39.015932+00	2025-07-11 20:18:50.7893+00	0	35	chat	chat	f	view/nesma/chat/index.vue	0		f	f	AI对话	ai-gva	f	
67	2025-07-17 00:44:57.859262+00	2025-07-17 00:44:57.859262+00	\N	0	24	anInfo	anInfo	f	plugin/announcement/view/info.vue	5		f	f	公告管理	box	f	
54	2025-07-10 19:35:30.526475+00	2025-07-11 15:12:18.110942+00	\N	0	0	nesma	nesma	f	view/routerHolder.vue	2		f	f	NESMA功能	management	f	
53	2025-07-10 19:34:18.836363+00	2025-07-11 15:31:00.990577+00	\N	0	0	workspace	workspace	f	view/nesma/analysis/workspace.vue	1		f	f	分析工作台	house	f	
61	2025-07-10 19:42:22.437756+00	2025-07-10 19:42:22.437756+00	2025-07-11 16:39:50.21498+00	0	54	monitor	nesmaMonitor	f	view/nesma/monitor/dashboard.vue	0		f	f	系统监控	monitor	f	
63	2025-07-10 19:43:32.813029+00	2025-07-11 20:28:56.951464+00	2025-07-12 06:50:25.626858+00	0	55	analysis	analysis	f	view/nesma/analysis/workspace.vue	2		f	f	智能分析	data-analysis	f	
35	2025-07-05 11:15:13.780732+00	2025-07-10 19:51:01.268737+00	2025-07-16 20:13:27.484018+00	0	0	index	index	f	view/nesma/index.vue	50		f	f	NESMA系统-旧	trend-charts	f	
30	2025-07-05 10:16:11.626299+00	2025-07-05 10:16:11.626299+00	2025-07-16 20:13:55.00391+00	1	9	https://plugin.gin-vue-admin.com/	https://plugin.gin-vue-admin.com/	f	https://plugin.gin-vue-admin.com/	0		f	f	插件市场	shop	f	
\.


--
-- Data for Name: sys_data_authority_id; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_data_authority_id (sys_authority_authority_id, data_authority_id_authority_id) FROM stdin;
888	8881
9528	8881
999	888
999	8881
999	9528
999	999
\.


--
-- Data for Name: sys_dictionaries; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_dictionaries (id, created_at, updated_at, deleted_at, name, type, status, "desc") FROM stdin;
1	2025-07-05 10:16:11.603248+00	2025-07-05 10:16:11.607032+00	\N	性别	gender	t	性别字典
2	2025-07-05 10:16:11.603248+00	2025-07-05 10:16:11.609206+00	\N	数据库int类型	int	t	int类型对应的数据库类型
3	2025-07-05 10:16:11.603248+00	2025-07-05 10:16:11.611743+00	\N	数据库时间日期类型	time.Time	t	数据库时间日期类型
4	2025-07-05 10:16:11.603248+00	2025-07-05 10:16:11.614693+00	\N	数据库浮点型	float64	t	数据库浮点型
5	2025-07-05 10:16:11.603248+00	2025-07-05 10:16:11.618271+00	\N	数据库字符串	string	t	数据库字符串
6	2025-07-05 10:16:11.603248+00	2025-07-05 10:16:11.620943+00	\N	数据库bool类型	bool	t	数据库bool类型
\.


--
-- Data for Name: sys_dictionary_details; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_dictionary_details (id, created_at, updated_at, deleted_at, label, value, extend, status, sort, sys_dictionary_id) FROM stdin;
1	2025-07-05 10:16:11.607532+00	2025-07-05 10:16:11.607532+00	\N	男	1		t	1	1
2	2025-07-05 10:16:11.607532+00	2025-07-05 10:16:11.607532+00	\N	女	2		t	2	1
3	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	smallint	1	mysql	t	1	2
4	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	mediumint	2	mysql	t	2	2
5	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	int	3	mysql	t	3	2
6	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	bigint	4	mysql	t	4	2
7	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	int2	5	pgsql	t	5	2
8	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	int4	6	pgsql	t	6	2
9	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	int6	7	pgsql	t	7	2
10	2025-07-05 10:16:11.609707+00	2025-07-05 10:16:11.609707+00	\N	int8	8	pgsql	t	8	2
11	2025-07-05 10:16:11.612318+00	2025-07-05 10:16:11.612318+00	\N	date			t	0	3
12	2025-07-05 10:16:11.612318+00	2025-07-05 10:16:11.612318+00	\N	time	1	mysql	t	1	3
13	2025-07-05 10:16:11.612318+00	2025-07-05 10:16:11.612318+00	\N	year	2	mysql	t	2	3
14	2025-07-05 10:16:11.612318+00	2025-07-05 10:16:11.612318+00	\N	datetime	3	mysql	t	3	3
15	2025-07-05 10:16:11.612318+00	2025-07-05 10:16:11.612318+00	\N	timestamp	5	mysql	t	5	3
16	2025-07-05 10:16:11.612318+00	2025-07-05 10:16:11.612318+00	\N	timestamptz	6	pgsql	t	5	3
17	2025-07-05 10:16:11.6154+00	2025-07-05 10:16:11.6154+00	\N	float			t	0	4
18	2025-07-05 10:16:11.6154+00	2025-07-05 10:16:11.6154+00	\N	double	1	mysql	t	1	4
19	2025-07-05 10:16:11.6154+00	2025-07-05 10:16:11.6154+00	\N	decimal	2	mysql	t	2	4
20	2025-07-05 10:16:11.6154+00	2025-07-05 10:16:11.6154+00	\N	numeric	3	pgsql	t	3	4
21	2025-07-05 10:16:11.6154+00	2025-07-05 10:16:11.6154+00	\N	smallserial	4	pgsql	t	4	4
22	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	char			t	0	5
23	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	varchar	1	mysql	t	1	5
24	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	tinyblob	2	mysql	t	2	5
25	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	tinytext	3	mysql	t	3	5
26	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	text	4	mysql	t	4	5
27	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	blob	5	mysql	t	5	5
28	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	mediumblob	6	mysql	t	6	5
29	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	mediumtext	7	mysql	t	7	5
30	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	longblob	8	mysql	t	8	5
31	2025-07-05 10:16:11.618795+00	2025-07-05 10:16:11.618795+00	\N	longtext	9	mysql	t	9	5
32	2025-07-05 10:16:11.620943+00	2025-07-05 10:16:11.620943+00	\N	tinyint	1	mysql	t	0	6
33	2025-07-05 10:16:11.620943+00	2025-07-05 10:16:11.620943+00	\N	bool	2	pgsql	t	0	6
\.


--
-- Data for Name: sys_export_template_condition; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_export_template_condition (id, created_at, updated_at, deleted_at, template_id, "from", "column", operator) FROM stdin;
\.


--
-- Data for Name: sys_export_template_join; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_export_template_join (id, created_at, updated_at, deleted_at, template_id, joins, "table", "on") FROM stdin;
\.


--
-- Data for Name: sys_export_templates; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_export_templates (id, created_at, updated_at, deleted_at, db_name, name, table_name, template_id, template_info, "limit", "order") FROM stdin;
1	2025-07-05 10:16:11.756915+00	2025-07-05 10:16:11.756915+00	\N		api	sys_apis	api	{\n"path":"路径",\n"method":"方法（大写）",\n"description":"方法介绍",\n"api_group":"方法分组"\n}	\N	
\.


--
-- Data for Name: sys_ignore_apis; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_ignore_apis (id, created_at, updated_at, deleted_at, path, method) FROM stdin;
1	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/swagger/*any	GET
2	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/api/freshCasbin	GET
3	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/uploads/file/*filepath	GET
4	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/health	GET
5	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/uploads/file/*filepath	HEAD
6	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/autoCode/llmAuto	POST
7	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/system/reloadSystem	POST
8	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/base/login	POST
9	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/base/captcha	POST
10	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/init/initdb	POST
11	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/init/checkdb	POST
12	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/info/getInfoDataSource	GET
13	2025-07-05 10:16:11.564381+00	2025-07-05 10:16:11.564381+00	\N	/info/getInfoPublic	GET
\.


--
-- Data for Name: sys_operation_records; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_operation_records (id, created_at, updated_at, deleted_at, ip, method, path, status, latency, agent, error_message, body, resp, user_id) FROM stdin;
\.


--
-- Data for Name: sys_params; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_params (id, created_at, updated_at, deleted_at, name, key, value, "desc") FROM stdin;
\.


--
-- Data for Name: sys_user_authority; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_user_authority (sys_user_id, sys_authority_authority_id) FROM stdin;
1	999
\.


--
-- Data for Name: sys_users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sys_users (id, created_at, updated_at, deleted_at, uuid, username, password, nick_name, header_img, authority_id, phone, email, enable, origin_setting) FROM stdin;
1	2025-07-05 10:16:11.742874+00	2025-07-06 14:51:53.011563+00	\N	00000000-0000-4000-8000-000000000001	admin	$2a$10$Vayz0uxmUcYT5PSVIwqEjurqmrIF2x5rxWfb4JZ3sLxOyQ.pwkWfK	蜂巢工作室	/logo.png	999			1	{}
\.


--
-- Data for Name: user_behaviors; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_behaviors (id, created_at, updated_at, deleted_at, user_id, action, item_type, item_id, session_id, context, duration, device_type, user_agent, ip_address) FROM stdin;
\.


--
-- Name: ai_business_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_business_analyses_id_seq', 1, false);


--
-- Name: ai_compliance_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_compliance_analyses_id_seq', 1, false);


--
-- Name: ai_functional_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_functional_analyses_id_seq', 1, false);


--
-- Name: ai_project_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_project_analyses_id_seq', 1, false);


--
-- Name: ai_quality_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_quality_analyses_id_seq', 1, false);


--
-- Name: ai_recommendation_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_recommendation_analyses_id_seq', 1, false);


--
-- Name: ai_risk_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_risk_analyses_id_seq', 1, false);


--
-- Name: ai_service_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_service_metrics_id_seq', 1, false);


--
-- Name: ai_technical_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.ai_technical_analyses_id_seq', 1, false);


--
-- Name: casbin_rule_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.casbin_rule_id_seq', 6571, true);


--
-- Name: chat_messages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.chat_messages_id_seq', 1, false);


--
-- Name: chat_sessions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.chat_sessions_id_seq', 1, true);


--
-- Name: exa_attachment_category_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.exa_attachment_category_id_seq', 1, false);


--
-- Name: exa_customers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.exa_customers_id_seq', 1, false);


--
-- Name: exa_file_chunks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.exa_file_chunks_id_seq', 1, false);


--
-- Name: exa_file_upload_and_downloads_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.exa_file_upload_and_downloads_id_seq', 1, false);


--
-- Name: exa_files_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.exa_files_id_seq', 1, false);


--
-- Name: gva_announcements_info_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.gva_announcements_info_id_seq', 1, false);


--
-- Name: jwt_blacklists_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.jwt_blacklists_id_seq', 1, false);


--
-- Name: nesma_agent_interactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_agent_interactions_id_seq', 1, false);


--
-- Name: nesma_agent_messages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_agent_messages_id_seq', 1, false);


--
-- Name: nesma_agent_tasks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_agent_tasks_id_seq', 1, false);


--
-- Name: nesma_agents_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_agents_id_seq', 5, true);


--
-- Name: nesma_analysis_histories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_analysis_histories_id_seq', 1, false);


--
-- Name: nesma_case_studies_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_case_studies_id_seq', 1, false);


--
-- Name: nesma_compact_memories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_compact_memories_id_seq', 1, false);


--
-- Name: nesma_complexity_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_complexity_metrics_id_seq', 1, false);


--
-- Name: nesma_doc_batches_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_doc_batches_id_seq', 1, false);


--
-- Name: nesma_doc_templates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_doc_templates_id_seq', 3, true);


--
-- Name: nesma_documents_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_documents_id_seq', 1, false);


--
-- Name: nesma_embeddings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_embeddings_id_seq', 1, false);


--
-- Name: nesma_evaluation_factors_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_evaluation_factors_id_seq', 1, false);


--
-- Name: nesma_evaluation_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_evaluation_history_id_seq', 1, false);


--
-- Name: nesma_evaluations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_evaluations_id_seq', 1, true);


--
-- Name: nesma_function_points_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_function_points_id_seq', 1, false);


--
-- Name: nesma_knowledge_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_knowledge_entries_id_seq', 3, true);


--
-- Name: nesma_knowledge_rules_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_knowledge_rules_id_seq', 1, false);


--
-- Name: nesma_nesma_evaluations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_nesma_evaluations_id_seq', 1, false);


--
-- Name: nesma_optimization_suggestions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_optimization_suggestions_id_seq', 1, false);


--
-- Name: nesma_project_analyses_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_project_analyses_id_seq', 1, false);


--
-- Name: nesma_project_cycles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_project_cycles_id_seq', 2, true);


--
-- Name: nesma_projects_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_projects_id_seq', 1, true);


--
-- Name: nesma_requirement_analysis_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_requirement_analysis_id_seq', 1, false);


--
-- Name: nesma_requirement_analysis_tasks_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_requirement_analysis_tasks_id_seq', 1, false);


--
-- Name: nesma_requirement_optimizations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_requirement_optimizations_id_seq', 1, false);


--
-- Name: nesma_requirement_versions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_requirement_versions_id_seq', 2, true);


--
-- Name: nesma_requirements_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_requirements_id_seq', 13, true);


--
-- Name: nesma_session_memories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_session_memories_id_seq', 1, false);


--
-- Name: nesma_system_activities_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_system_activities_id_seq', 1, false);


--
-- Name: nesma_validation_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_validation_items_id_seq', 1, false);


--
-- Name: nesma_vector_indexes_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_vector_indexes_id_seq', 1, false);


--
-- Name: nesma_vector_memories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_vector_memories_id_seq', 1, false);


--
-- Name: nesma_vector_queries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_vector_queries_id_seq', 1, false);


--
-- Name: nesma_vector_stores_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_vector_stores_id_seq', 1, false);


--
-- Name: nesma_workflow_executions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_workflow_executions_id_seq', 1, false);


--
-- Name: nesma_workflows_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.nesma_workflows_id_seq', 1, false);


--
-- Name: recommendation_adoptions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.recommendation_adoptions_id_seq', 1, false);


--
-- Name: recommendation_histories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.recommendation_histories_id_seq', 1, false);


--
-- Name: requirement_analysis_results_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.requirement_analysis_results_id_seq', 1, false);


--
-- Name: sys_apis_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_apis_id_seq', 327, true);


--
-- Name: sys_authorities_authority_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_authorities_authority_id_seq', 1, false);


--
-- Name: sys_auto_code_histories_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_auto_code_histories_id_seq', 1, false);


--
-- Name: sys_auto_code_packages_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_auto_code_packages_id_seq', 4, true);


--
-- Name: sys_base_menu_btns_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_base_menu_btns_id_seq', 1, false);


--
-- Name: sys_base_menu_parameters_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_base_menu_parameters_id_seq', 1, false);


--
-- Name: sys_base_menus_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_base_menus_id_seq', 67, true);


--
-- Name: sys_dictionaries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_dictionaries_id_seq', 6, true);


--
-- Name: sys_dictionary_details_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_dictionary_details_id_seq', 33, true);


--
-- Name: sys_export_template_condition_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_export_template_condition_id_seq', 1, false);


--
-- Name: sys_export_template_join_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_export_template_join_id_seq', 1, false);


--
-- Name: sys_export_templates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_export_templates_id_seq', 1, true);


--
-- Name: sys_ignore_apis_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_ignore_apis_id_seq', 13, true);


--
-- Name: sys_operation_records_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_operation_records_id_seq', 21178, true);


--
-- Name: sys_params_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_params_id_seq', 1, false);


--
-- Name: sys_users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.sys_users_id_seq', 2, true);


--
-- Name: user_behaviors_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.user_behaviors_id_seq', 1, false);


--
-- Name: ai_business_analyses ai_business_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_business_analyses
    ADD CONSTRAINT ai_business_analyses_pkey PRIMARY KEY (id);


--
-- Name: ai_compliance_analyses ai_compliance_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_compliance_analyses
    ADD CONSTRAINT ai_compliance_analyses_pkey PRIMARY KEY (id);


--
-- Name: ai_functional_analyses ai_functional_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_functional_analyses
    ADD CONSTRAINT ai_functional_analyses_pkey PRIMARY KEY (id);


--
-- Name: ai_project_analyses ai_project_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_project_analyses
    ADD CONSTRAINT ai_project_analyses_pkey PRIMARY KEY (id);


--
-- Name: ai_quality_analyses ai_quality_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_quality_analyses
    ADD CONSTRAINT ai_quality_analyses_pkey PRIMARY KEY (id);


--
-- Name: ai_recommendation_analyses ai_recommendation_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_recommendation_analyses
    ADD CONSTRAINT ai_recommendation_analyses_pkey PRIMARY KEY (id);


--
-- Name: ai_risk_analyses ai_risk_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_risk_analyses
    ADD CONSTRAINT ai_risk_analyses_pkey PRIMARY KEY (id);


--
-- Name: ai_service_metrics ai_service_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_service_metrics
    ADD CONSTRAINT ai_service_metrics_pkey PRIMARY KEY (id);


--
-- Name: ai_technical_analyses ai_technical_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ai_technical_analyses
    ADD CONSTRAINT ai_technical_analyses_pkey PRIMARY KEY (id);


--
-- Name: casbin_rule casbin_rule_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.casbin_rule
    ADD CONSTRAINT casbin_rule_pkey PRIMARY KEY (id);


--
-- Name: chat_messages chat_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_pkey PRIMARY KEY (id);


--
-- Name: chat_sessions chat_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_pkey PRIMARY KEY (id);


--
-- Name: exa_attachment_category exa_attachment_category_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_attachment_category
    ADD CONSTRAINT exa_attachment_category_pkey PRIMARY KEY (id);


--
-- Name: exa_customers exa_customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_customers
    ADD CONSTRAINT exa_customers_pkey PRIMARY KEY (id);


--
-- Name: exa_file_chunks exa_file_chunks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_chunks
    ADD CONSTRAINT exa_file_chunks_pkey PRIMARY KEY (id);


--
-- Name: exa_file_upload_and_downloads exa_file_upload_and_downloads_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_file_upload_and_downloads
    ADD CONSTRAINT exa_file_upload_and_downloads_pkey PRIMARY KEY (id);


--
-- Name: exa_files exa_files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.exa_files
    ADD CONSTRAINT exa_files_pkey PRIMARY KEY (id);


--
-- Name: gva_announcements_info gva_announcements_info_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.gva_announcements_info
    ADD CONSTRAINT gva_announcements_info_pkey PRIMARY KEY (id);


--
-- Name: jwt_blacklists jwt_blacklists_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jwt_blacklists
    ADD CONSTRAINT jwt_blacklists_pkey PRIMARY KEY (id);


--
-- Name: nesma_agent_interactions nesma_agent_interactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_interactions
    ADD CONSTRAINT nesma_agent_interactions_pkey PRIMARY KEY (id);


--
-- Name: nesma_agent_messages nesma_agent_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_messages
    ADD CONSTRAINT nesma_agent_messages_pkey PRIMARY KEY (id);


--
-- Name: nesma_agent_tasks nesma_agent_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_tasks
    ADD CONSTRAINT nesma_agent_tasks_pkey PRIMARY KEY (id);


--
-- Name: nesma_agents nesma_agents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agents
    ADD CONSTRAINT nesma_agents_pkey PRIMARY KEY (id);


--
-- Name: nesma_analysis_histories nesma_analysis_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_analysis_histories
    ADD CONSTRAINT nesma_analysis_histories_pkey PRIMARY KEY (id);


--
-- Name: nesma_case_studies nesma_case_studies_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_case_studies
    ADD CONSTRAINT nesma_case_studies_pkey PRIMARY KEY (id);


--
-- Name: nesma_compact_memories nesma_compact_memories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_compact_memories
    ADD CONSTRAINT nesma_compact_memories_pkey PRIMARY KEY (id);


--
-- Name: nesma_complexity_metrics nesma_complexity_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_complexity_metrics
    ADD CONSTRAINT nesma_complexity_metrics_pkey PRIMARY KEY (id);


--
-- Name: nesma_doc_batches nesma_doc_batches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_doc_batches
    ADD CONSTRAINT nesma_doc_batches_pkey PRIMARY KEY (id);


--
-- Name: nesma_doc_templates nesma_doc_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_doc_templates
    ADD CONSTRAINT nesma_doc_templates_pkey PRIMARY KEY (id);


--
-- Name: nesma_documents nesma_documents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_documents
    ADD CONSTRAINT nesma_documents_pkey PRIMARY KEY (id);


--
-- Name: nesma_embeddings nesma_embeddings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_embeddings
    ADD CONSTRAINT nesma_embeddings_pkey PRIMARY KEY (id);


--
-- Name: nesma_evaluation_factors nesma_evaluation_factors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_evaluation_factors
    ADD CONSTRAINT nesma_evaluation_factors_pkey PRIMARY KEY (id);


--
-- Name: nesma_evaluation_history nesma_evaluation_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_evaluation_history
    ADD CONSTRAINT nesma_evaluation_history_pkey PRIMARY KEY (id);


--
-- Name: nesma_evaluations nesma_evaluations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_evaluations
    ADD CONSTRAINT nesma_evaluations_pkey PRIMARY KEY (id);


--
-- Name: nesma_function_points nesma_function_points_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_function_points
    ADD CONSTRAINT nesma_function_points_pkey PRIMARY KEY (id);


--
-- Name: nesma_knowledge_entries nesma_knowledge_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_knowledge_entries
    ADD CONSTRAINT nesma_knowledge_entries_pkey PRIMARY KEY (id);


--
-- Name: nesma_knowledge_rules nesma_knowledge_rules_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_knowledge_rules
    ADD CONSTRAINT nesma_knowledge_rules_pkey PRIMARY KEY (id);


--
-- Name: nesma_nesma_evaluations nesma_nesma_evaluations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_nesma_evaluations
    ADD CONSTRAINT nesma_nesma_evaluations_pkey PRIMARY KEY (id);


--
-- Name: nesma_optimization_suggestions nesma_optimization_suggestions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_optimization_suggestions
    ADD CONSTRAINT nesma_optimization_suggestions_pkey PRIMARY KEY (id);


--
-- Name: nesma_project_analyses nesma_project_analyses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_project_analyses
    ADD CONSTRAINT nesma_project_analyses_pkey PRIMARY KEY (id);


--
-- Name: nesma_project_cycles nesma_project_cycles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_project_cycles
    ADD CONSTRAINT nesma_project_cycles_pkey PRIMARY KEY (id);


--
-- Name: nesma_projects nesma_projects_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_projects
    ADD CONSTRAINT nesma_projects_pkey PRIMARY KEY (id);


--
-- Name: nesma_requirement_analysis nesma_requirement_analysis_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_analysis
    ADD CONSTRAINT nesma_requirement_analysis_pkey PRIMARY KEY (id);


--
-- Name: nesma_requirement_analysis_tasks nesma_requirement_analysis_tasks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_analysis_tasks
    ADD CONSTRAINT nesma_requirement_analysis_tasks_pkey PRIMARY KEY (id);


--
-- Name: nesma_requirement_optimizations nesma_requirement_optimizations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_optimizations
    ADD CONSTRAINT nesma_requirement_optimizations_pkey PRIMARY KEY (id);


--
-- Name: nesma_requirement_versions nesma_requirement_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirement_versions
    ADD CONSTRAINT nesma_requirement_versions_pkey PRIMARY KEY (id);


--
-- Name: nesma_requirements nesma_requirements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_requirements
    ADD CONSTRAINT nesma_requirements_pkey PRIMARY KEY (id);


--
-- Name: nesma_session_memories nesma_session_memories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_session_memories
    ADD CONSTRAINT nesma_session_memories_pkey PRIMARY KEY (id);


--
-- Name: nesma_system_activities nesma_system_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_system_activities
    ADD CONSTRAINT nesma_system_activities_pkey PRIMARY KEY (id);


--
-- Name: nesma_validation_items nesma_validation_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_validation_items
    ADD CONSTRAINT nesma_validation_items_pkey PRIMARY KEY (id);


--
-- Name: nesma_vector_indexes nesma_vector_indexes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_indexes
    ADD CONSTRAINT nesma_vector_indexes_pkey PRIMARY KEY (id);


--
-- Name: nesma_vector_memories nesma_vector_memories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_memories
    ADD CONSTRAINT nesma_vector_memories_pkey PRIMARY KEY (id);


--
-- Name: nesma_vector_queries nesma_vector_queries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_queries
    ADD CONSTRAINT nesma_vector_queries_pkey PRIMARY KEY (id);


--
-- Name: nesma_vector_stores nesma_vector_stores_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_stores
    ADD CONSTRAINT nesma_vector_stores_pkey PRIMARY KEY (id);


--
-- Name: nesma_workflow_executions nesma_workflow_executions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_workflow_executions
    ADD CONSTRAINT nesma_workflow_executions_pkey PRIMARY KEY (id);


--
-- Name: nesma_workflows nesma_workflows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_workflows
    ADD CONSTRAINT nesma_workflows_pkey PRIMARY KEY (id);


--
-- Name: recommendation_adoptions recommendation_adoptions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recommendation_adoptions
    ADD CONSTRAINT recommendation_adoptions_pkey PRIMARY KEY (id);


--
-- Name: recommendation_histories recommendation_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recommendation_histories
    ADD CONSTRAINT recommendation_histories_pkey PRIMARY KEY (id);


--
-- Name: requirement_analysis_results requirement_analysis_results_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.requirement_analysis_results
    ADD CONSTRAINT requirement_analysis_results_pkey PRIMARY KEY (id);


--
-- Name: sys_apis sys_apis_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_apis
    ADD CONSTRAINT sys_apis_pkey PRIMARY KEY (id);


--
-- Name: sys_authority_menus sys_authority_menus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authority_menus
    ADD CONSTRAINT sys_authority_menus_pkey PRIMARY KEY (sys_base_menu_id, sys_authority_authority_id);


--
-- Name: sys_auto_code_histories sys_auto_code_histories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_histories
    ADD CONSTRAINT sys_auto_code_histories_pkey PRIMARY KEY (id);


--
-- Name: sys_auto_code_packages sys_auto_code_packages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_auto_code_packages
    ADD CONSTRAINT sys_auto_code_packages_pkey PRIMARY KEY (id);


--
-- Name: sys_base_menu_btns sys_base_menu_btns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_btns
    ADD CONSTRAINT sys_base_menu_btns_pkey PRIMARY KEY (id);


--
-- Name: sys_base_menu_parameters sys_base_menu_parameters_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menu_parameters
    ADD CONSTRAINT sys_base_menu_parameters_pkey PRIMARY KEY (id);


--
-- Name: sys_base_menus sys_base_menus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_base_menus
    ADD CONSTRAINT sys_base_menus_pkey PRIMARY KEY (id);


--
-- Name: sys_data_authority_id sys_data_authority_id_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_data_authority_id
    ADD CONSTRAINT sys_data_authority_id_pkey PRIMARY KEY (sys_authority_authority_id, data_authority_id_authority_id);


--
-- Name: sys_dictionaries sys_dictionaries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionaries
    ADD CONSTRAINT sys_dictionaries_pkey PRIMARY KEY (id);


--
-- Name: sys_dictionary_details sys_dictionary_details_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_dictionary_details
    ADD CONSTRAINT sys_dictionary_details_pkey PRIMARY KEY (id);


--
-- Name: sys_export_template_condition sys_export_template_condition_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_condition
    ADD CONSTRAINT sys_export_template_condition_pkey PRIMARY KEY (id);


--
-- Name: sys_export_template_join sys_export_template_join_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_template_join
    ADD CONSTRAINT sys_export_template_join_pkey PRIMARY KEY (id);


--
-- Name: sys_export_templates sys_export_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_export_templates
    ADD CONSTRAINT sys_export_templates_pkey PRIMARY KEY (id);


--
-- Name: sys_ignore_apis sys_ignore_apis_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_ignore_apis
    ADD CONSTRAINT sys_ignore_apis_pkey PRIMARY KEY (id);


--
-- Name: sys_operation_records sys_operation_records_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_operation_records
    ADD CONSTRAINT sys_operation_records_pkey PRIMARY KEY (id);


--
-- Name: sys_params sys_params_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_params
    ADD CONSTRAINT sys_params_pkey PRIMARY KEY (id);


--
-- Name: sys_user_authority sys_user_authority_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_user_authority
    ADD CONSTRAINT sys_user_authority_pkey PRIMARY KEY (sys_user_id, sys_authority_authority_id);


--
-- Name: sys_users sys_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_users
    ADD CONSTRAINT sys_users_pkey PRIMARY KEY (id);


--
-- Name: nesma_agent_messages uni_nesma_agent_messages_message_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_messages
    ADD CONSTRAINT uni_nesma_agent_messages_message_id UNIQUE (message_id);


--
-- Name: nesma_agent_tasks uni_nesma_agent_tasks_task_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agent_tasks
    ADD CONSTRAINT uni_nesma_agent_tasks_task_id UNIQUE (task_id);


--
-- Name: nesma_agents uni_nesma_agents_agent_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_agents
    ADD CONSTRAINT uni_nesma_agents_agent_id UNIQUE (agent_id);


--
-- Name: nesma_embeddings uni_nesma_embeddings_embedding_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_embeddings
    ADD CONSTRAINT uni_nesma_embeddings_embedding_id UNIQUE (embedding_id);


--
-- Name: nesma_vector_indexes uni_nesma_vector_indexes_index_name; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_indexes
    ADD CONSTRAINT uni_nesma_vector_indexes_index_name UNIQUE (index_name);


--
-- Name: nesma_vector_queries uni_nesma_vector_queries_query_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_queries
    ADD CONSTRAINT uni_nesma_vector_queries_query_id UNIQUE (query_id);


--
-- Name: nesma_vector_stores uni_nesma_vector_stores_vector_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_vector_stores
    ADD CONSTRAINT uni_nesma_vector_stores_vector_id UNIQUE (vector_id);


--
-- Name: nesma_workflow_executions uni_nesma_workflow_executions_execution_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_workflow_executions
    ADD CONSTRAINT uni_nesma_workflow_executions_execution_id UNIQUE (execution_id);


--
-- Name: nesma_workflows uni_nesma_workflows_workflow_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.nesma_workflows
    ADD CONSTRAINT uni_nesma_workflows_workflow_id UNIQUE (workflow_id);


--
-- Name: sys_authorities uni_sys_authorities_authority_id; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sys_authorities
    ADD CONSTRAINT uni_sys_authorities_authority_id PRIMARY KEY (authority_id);


--
-- Name: user_behaviors user_behaviors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_behaviors
    ADD CONSTRAINT user_behaviors_pkey PRIMARY KEY (id);


--
-- Name: idx_ai_business_analyses_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_business_analyses_analysis_id ON public.ai_business_analyses USING btree (analysis_id);


--
-- Name: idx_ai_business_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_business_analyses_deleted_at ON public.ai_business_analyses USING btree (deleted_at);


--
-- Name: idx_ai_compliance_analyses_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_compliance_analyses_analysis_id ON public.ai_compliance_analyses USING btree (analysis_id);


--
-- Name: idx_ai_compliance_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_compliance_analyses_deleted_at ON public.ai_compliance_analyses USING btree (deleted_at);


--
-- Name: idx_ai_functional_analyses_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_functional_analyses_analysis_id ON public.ai_functional_analyses USING btree (analysis_id);


--
-- Name: idx_ai_functional_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_functional_analyses_deleted_at ON public.ai_functional_analyses USING btree (deleted_at);


--
-- Name: idx_ai_project_analyses_analysis_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_analysis_type ON public.ai_project_analyses USING btree (analysis_type);


--
-- Name: idx_ai_project_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_deleted_at ON public.ai_project_analyses USING btree (deleted_at);


--
-- Name: idx_ai_project_analyses_evaluation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_evaluation_id ON public.ai_project_analyses USING btree (evaluation_id);


--
-- Name: idx_ai_project_analyses_overall_score; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_overall_score ON public.ai_project_analyses USING btree (overall_score);


--
-- Name: idx_ai_project_analyses_project_evaluation; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_project_evaluation ON public.ai_project_analyses USING btree (project_id, evaluation_id);


--
-- Name: idx_ai_project_analyses_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_project_id ON public.ai_project_analyses USING btree (project_id);


--
-- Name: idx_ai_project_analyses_start_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_start_time ON public.ai_project_analyses USING btree (start_time);


--
-- Name: idx_ai_project_analyses_status_progress; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_project_analyses_status_progress ON public.ai_project_analyses USING btree (status, progress);


--
-- Name: idx_ai_quality_analyses_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_quality_analyses_analysis_id ON public.ai_quality_analyses USING btree (analysis_id);


--
-- Name: idx_ai_quality_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_quality_analyses_deleted_at ON public.ai_quality_analyses USING btree (deleted_at);


--
-- Name: idx_ai_recommendation_analyses_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_recommendation_analyses_analysis_id ON public.ai_recommendation_analyses USING btree (analysis_id);


--
-- Name: idx_ai_recommendation_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_recommendation_analyses_deleted_at ON public.ai_recommendation_analyses USING btree (deleted_at);


--
-- Name: idx_ai_recommendation_analyses_impact_urgency; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_recommendation_analyses_impact_urgency ON public.ai_recommendation_analyses USING btree (impact_level, urgency);


--
-- Name: idx_ai_recommendation_analyses_type_priority; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_recommendation_analyses_type_priority ON public.ai_recommendation_analyses USING btree (recommendation_type, priority);


--
-- Name: idx_ai_risk_analyses_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_risk_analyses_analysis_id ON public.ai_risk_analyses USING btree (analysis_id);


--
-- Name: idx_ai_risk_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_risk_analyses_deleted_at ON public.ai_risk_analyses USING btree (deleted_at);


--
-- Name: idx_ai_service_metrics_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_service_metrics_created_at ON public.ai_service_metrics USING btree (created_at);


--
-- Name: idx_ai_service_metrics_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_service_metrics_deleted_at ON public.ai_service_metrics USING btree (deleted_at);


--
-- Name: idx_ai_service_metrics_recorded_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_service_metrics_recorded_at ON public.ai_service_metrics USING btree (recorded_at);


--
-- Name: idx_ai_service_metrics_service_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_service_metrics_service_name ON public.ai_service_metrics USING btree (service_name);


--
-- Name: idx_ai_service_metrics_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_service_metrics_user_id ON public.ai_service_metrics USING btree (user_id);


--
-- Name: idx_ai_technical_analyses_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_technical_analyses_analysis_id ON public.ai_technical_analyses USING btree (analysis_id);


--
-- Name: idx_ai_technical_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_ai_technical_analyses_deleted_at ON public.ai_technical_analyses USING btree (deleted_at);


--
-- Name: idx_casbin_rule; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_casbin_rule ON public.casbin_rule USING btree (ptype, v0, v1, v2, v3, v4, v5);


--
-- Name: idx_chat_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_messages_deleted_at ON public.chat_messages USING btree (deleted_at);


--
-- Name: idx_chat_messages_session_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_messages_session_id ON public.chat_messages USING btree (session_id);


--
-- Name: idx_chat_sessions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_sessions_deleted_at ON public.chat_sessions USING btree (deleted_at);


--
-- Name: idx_chat_sessions_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_sessions_project_id ON public.chat_sessions USING btree (project_id);


--
-- Name: idx_chat_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_chat_sessions_user_id ON public.chat_sessions USING btree (user_id);


--
-- Name: idx_evaluation_factors_evaluation_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_evaluation_factors_evaluation_id ON public.nesma_evaluation_factors USING btree (evaluation_id);


--
-- Name: idx_evaluation_factors_nesma_version; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_evaluation_factors_nesma_version ON public.nesma_evaluation_factors USING btree (nesma_version);


--
-- Name: idx_exa_attachment_category_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_attachment_category_deleted_at ON public.exa_attachment_category USING btree (deleted_at);


--
-- Name: idx_exa_customers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_customers_deleted_at ON public.exa_customers USING btree (deleted_at);


--
-- Name: idx_exa_file_chunks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_file_chunks_deleted_at ON public.exa_file_chunks USING btree (deleted_at);


--
-- Name: idx_exa_file_upload_and_downloads_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_file_upload_and_downloads_deleted_at ON public.exa_file_upload_and_downloads USING btree (deleted_at);


--
-- Name: idx_exa_files_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_exa_files_deleted_at ON public.exa_files USING btree (deleted_at);


--
-- Name: idx_gva_announcements_info_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_gva_announcements_info_deleted_at ON public.gva_announcements_info USING btree (deleted_at);


--
-- Name: idx_jwt_blacklists_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jwt_blacklists_deleted_at ON public.jwt_blacklists USING btree (deleted_at);


--
-- Name: idx_nesma_agent_interactions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_agent_interactions_deleted_at ON public.nesma_agent_interactions USING btree (deleted_at);


--
-- Name: idx_nesma_agent_messages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_agent_messages_deleted_at ON public.nesma_agent_messages USING btree (deleted_at);


--
-- Name: idx_nesma_agent_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_agent_tasks_deleted_at ON public.nesma_agent_tasks USING btree (deleted_at);


--
-- Name: idx_nesma_agents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_agents_deleted_at ON public.nesma_agents USING btree (deleted_at);


--
-- Name: idx_nesma_analysis_histories_analysis_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_analysis_histories_analysis_date ON public.nesma_analysis_histories USING btree (analysis_date);


--
-- Name: idx_nesma_analysis_histories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_analysis_histories_deleted_at ON public.nesma_analysis_histories USING btree (deleted_at);


--
-- Name: idx_nesma_analysis_histories_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_analysis_histories_project_id ON public.nesma_analysis_histories USING btree (project_id);


--
-- Name: idx_nesma_analysis_histories_project_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_analysis_histories_project_type ON public.nesma_analysis_histories USING btree (project_id, analysis_type);


--
-- Name: idx_nesma_analysis_tasks_priority; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_analysis_tasks_priority ON public.nesma_requirement_analysis_tasks USING btree (priority);


--
-- Name: idx_nesma_analysis_tasks_type_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_analysis_tasks_type_status ON public.nesma_requirement_analysis_tasks USING btree (task_type, status);


--
-- Name: idx_nesma_case_studies_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_case_studies_deleted_at ON public.nesma_case_studies USING btree (deleted_at);


--
-- Name: idx_nesma_compact_memories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_compact_memories_deleted_at ON public.nesma_compact_memories USING btree (deleted_at);


--
-- Name: idx_nesma_complexity_metrics_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_complexity_metrics_deleted_at ON public.nesma_complexity_metrics USING btree (deleted_at);


--
-- Name: idx_nesma_doc_batches_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_doc_batches_deleted_at ON public.nesma_doc_batches USING btree (deleted_at);


--
-- Name: idx_nesma_doc_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_doc_templates_deleted_at ON public.nesma_doc_templates USING btree (deleted_at);


--
-- Name: idx_nesma_documents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_documents_deleted_at ON public.nesma_documents USING btree (deleted_at);


--
-- Name: idx_nesma_embeddings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_embeddings_deleted_at ON public.nesma_embeddings USING btree (deleted_at);


--
-- Name: idx_nesma_evaluation_factors_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_evaluation_factors_deleted_at ON public.nesma_evaluation_factors USING btree (deleted_at);


--
-- Name: idx_nesma_evaluation_history_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_evaluation_history_deleted_at ON public.nesma_evaluation_history USING btree (deleted_at);


--
-- Name: idx_nesma_evaluations_ai_analysis_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_evaluations_ai_analysis_id ON public.nesma_evaluations USING btree (ai_analysis_id);


--
-- Name: idx_nesma_evaluations_cycle_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_evaluations_cycle_id ON public.nesma_evaluations USING btree (cycle_id);


--
-- Name: idx_nesma_evaluations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_evaluations_deleted_at ON public.nesma_evaluations USING btree (deleted_at);


--
-- Name: idx_nesma_evaluations_requirement_version_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_evaluations_requirement_version_id ON public.nesma_evaluations USING btree (requirement_version_id);


--
-- Name: idx_nesma_function_points_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_function_points_deleted_at ON public.nesma_function_points USING btree (deleted_at);


--
-- Name: idx_nesma_knowledge_entries_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_knowledge_entries_deleted_at ON public.nesma_knowledge_entries USING btree (deleted_at);


--
-- Name: idx_nesma_knowledge_rules_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_knowledge_rules_deleted_at ON public.nesma_knowledge_rules USING btree (deleted_at);


--
-- Name: idx_nesma_nesma_evaluations_cycle_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_nesma_evaluations_cycle_id ON public.nesma_nesma_evaluations USING btree (cycle_id);


--
-- Name: idx_nesma_nesma_evaluations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_nesma_evaluations_deleted_at ON public.nesma_nesma_evaluations USING btree (deleted_at);


--
-- Name: idx_nesma_nesma_evaluations_evaluation_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_nesma_evaluations_evaluation_date ON public.nesma_nesma_evaluations USING btree (evaluation_date);


--
-- Name: idx_nesma_nesma_evaluations_evaluation_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_nesma_evaluations_evaluation_type ON public.nesma_nesma_evaluations USING btree (evaluation_type);


--
-- Name: idx_nesma_nesma_evaluations_overall_grade; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_nesma_evaluations_overall_grade ON public.nesma_nesma_evaluations USING btree (overall_grade);


--
-- Name: idx_nesma_nesma_evaluations_project_cycle; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_nesma_evaluations_project_cycle ON public.nesma_nesma_evaluations USING btree (project_id, cycle_id);


--
-- Name: idx_nesma_nesma_evaluations_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_nesma_evaluations_project_id ON public.nesma_nesma_evaluations USING btree (project_id);


--
-- Name: idx_nesma_optimization_suggestions_assigned_to; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_optimization_suggestions_assigned_to ON public.nesma_optimization_suggestions USING btree (assigned_to);


--
-- Name: idx_nesma_optimization_suggestions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_optimization_suggestions_deleted_at ON public.nesma_optimization_suggestions USING btree (deleted_at);


--
-- Name: idx_nesma_optimization_suggestions_priority_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_optimization_suggestions_priority_status ON public.nesma_optimization_suggestions USING btree (priority, status);


--
-- Name: idx_nesma_optimization_suggestions_project; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_optimization_suggestions_project ON public.nesma_optimization_suggestions USING btree (project_id);


--
-- Name: idx_nesma_optimization_suggestions_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_optimization_suggestions_project_id ON public.nesma_optimization_suggestions USING btree (project_id);


--
-- Name: idx_nesma_optimization_suggestions_requirement_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_optimization_suggestions_requirement_id ON public.nesma_optimization_suggestions USING btree (requirement_id);


--
-- Name: idx_nesma_optimization_suggestions_type_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_optimization_suggestions_type_category ON public.nesma_optimization_suggestions USING btree (suggestion_type, category);


--
-- Name: idx_nesma_project_analyses_analysis_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_analysis_date ON public.nesma_project_analyses USING btree (analysis_date);


--
-- Name: idx_nesma_project_analyses_analysis_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_analysis_type ON public.nesma_project_analyses USING btree (analysis_type);


--
-- Name: idx_nesma_project_analyses_cycle_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_cycle_id ON public.nesma_project_analyses USING btree (cycle_id);


--
-- Name: idx_nesma_project_analyses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_deleted_at ON public.nesma_project_analyses USING btree (deleted_at);


--
-- Name: idx_nesma_project_analyses_overall_grade; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_overall_grade ON public.nesma_project_analyses USING btree (overall_grade);


--
-- Name: idx_nesma_project_analyses_project_cycle; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_project_cycle ON public.nesma_project_analyses USING btree (project_id, cycle_id);


--
-- Name: idx_nesma_project_analyses_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_project_id ON public.nesma_project_analyses USING btree (project_id);


--
-- Name: idx_nesma_project_analyses_version_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_analyses_version_id ON public.nesma_project_analyses USING btree (version_id);


--
-- Name: idx_nesma_project_cycles_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_cycles_deleted_at ON public.nesma_project_cycles USING btree (deleted_at);


--
-- Name: idx_nesma_project_cycles_phase; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_cycles_phase ON public.nesma_project_cycles USING btree (phase);


--
-- Name: idx_nesma_project_cycles_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_cycles_project_id ON public.nesma_project_cycles USING btree (project_id);


--
-- Name: idx_nesma_project_cycles_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_project_cycles_status ON public.nesma_project_cycles USING btree (status);


--
-- Name: idx_nesma_projects_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_projects_deleted_at ON public.nesma_projects USING btree (deleted_at);


--
-- Name: idx_nesma_requirement_analysis_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_analysis_deleted_at ON public.nesma_requirement_analysis USING btree (deleted_at);


--
-- Name: idx_nesma_requirement_analysis_tasks_cycle_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_analysis_tasks_cycle_id ON public.nesma_requirement_analysis_tasks USING btree (cycle_id);


--
-- Name: idx_nesma_requirement_analysis_tasks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_analysis_tasks_deleted_at ON public.nesma_requirement_analysis_tasks USING btree (deleted_at);


--
-- Name: idx_nesma_requirement_analysis_tasks_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_analysis_tasks_project_id ON public.nesma_requirement_analysis_tasks USING btree (project_id);


--
-- Name: idx_nesma_requirement_optimizations_applied; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_optimizations_applied ON public.nesma_requirement_optimizations USING btree (applied_to_requirement);


--
-- Name: idx_nesma_requirement_optimizations_decision; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_optimizations_decision ON public.nesma_requirement_optimizations USING btree (user_decision);


--
-- Name: idx_nesma_requirement_optimizations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_optimizations_deleted_at ON public.nesma_requirement_optimizations USING btree (deleted_at);


--
-- Name: idx_nesma_requirement_optimizations_generated_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_optimizations_generated_at ON public.nesma_requirement_optimizations USING btree (generated_at);


--
-- Name: idx_nesma_requirement_optimizations_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_optimizations_project_id ON public.nesma_requirement_optimizations USING btree (project_id);


--
-- Name: idx_nesma_requirement_optimizations_requirement; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_optimizations_requirement ON public.nesma_requirement_optimizations USING btree (requirement_id);


--
-- Name: idx_nesma_requirement_optimizations_requirement_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_optimizations_requirement_id ON public.nesma_requirement_optimizations USING btree (requirement_id);


--
-- Name: idx_nesma_requirement_versions_cycle_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_versions_cycle_id ON public.nesma_requirement_versions USING btree (cycle_id);


--
-- Name: idx_nesma_requirement_versions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_versions_deleted_at ON public.nesma_requirement_versions USING btree (deleted_at);


--
-- Name: idx_nesma_requirement_versions_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_versions_status ON public.nesma_requirement_versions USING btree (status);


--
-- Name: idx_nesma_requirement_versions_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirement_versions_type ON public.nesma_requirement_versions USING btree (version_type);


--
-- Name: idx_nesma_requirements_ai_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirements_ai_status ON public.nesma_requirements USING btree (ai_analysis_status);


--
-- Name: idx_nesma_requirements_cycle_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirements_cycle_id ON public.nesma_requirements USING btree (cycle_id);


--
-- Name: idx_nesma_requirements_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirements_deleted_at ON public.nesma_requirements USING btree (deleted_at);


--
-- Name: idx_nesma_requirements_project_cycle; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirements_project_cycle ON public.nesma_requirements USING btree (project_id, cycle_id);


--
-- Name: idx_nesma_requirements_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirements_project_id ON public.nesma_requirements USING btree (project_id);


--
-- Name: idx_nesma_requirements_version_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirements_version_id ON public.nesma_requirements USING btree (version_id);


--
-- Name: idx_nesma_requirements_version_level; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_requirements_version_level ON public.nesma_requirements USING btree (version_id, level);


--
-- Name: idx_nesma_session_memories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_session_memories_deleted_at ON public.nesma_session_memories USING btree (deleted_at);


--
-- Name: idx_nesma_system_activities_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_system_activities_deleted_at ON public.nesma_system_activities USING btree (deleted_at);


--
-- Name: idx_nesma_validation_items_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_validation_items_deleted_at ON public.nesma_validation_items USING btree (deleted_at);


--
-- Name: idx_nesma_vector_indexes_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_vector_indexes_deleted_at ON public.nesma_vector_indexes USING btree (deleted_at);


--
-- Name: idx_nesma_vector_memories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_vector_memories_deleted_at ON public.nesma_vector_memories USING btree (deleted_at);


--
-- Name: idx_nesma_vector_queries_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_vector_queries_deleted_at ON public.nesma_vector_queries USING btree (deleted_at);


--
-- Name: idx_nesma_vector_stores_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_vector_stores_deleted_at ON public.nesma_vector_stores USING btree (deleted_at);


--
-- Name: idx_nesma_workflow_executions_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_workflow_executions_deleted_at ON public.nesma_workflow_executions USING btree (deleted_at);


--
-- Name: idx_nesma_workflows_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_nesma_workflows_deleted_at ON public.nesma_workflows USING btree (deleted_at);


--
-- Name: idx_recommendation_adoptions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recommendation_adoptions_user_id ON public.recommendation_adoptions USING btree (user_id);


--
-- Name: idx_recommendation_histories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recommendation_histories_deleted_at ON public.recommendation_histories USING btree (deleted_at);


--
-- Name: idx_recommendation_histories_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_recommendation_histories_user_id ON public.recommendation_histories USING btree (user_id);


--
-- Name: idx_requirement_analysis_results_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_requirement_analysis_results_deleted_at ON public.requirement_analysis_results USING btree (deleted_at);


--
-- Name: idx_requirement_analysis_results_project_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_requirement_analysis_results_project_id ON public.requirement_analysis_results USING btree (project_id);


--
-- Name: idx_requirement_analysis_results_requirement_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_requirement_analysis_results_requirement_id ON public.requirement_analysis_results USING btree (requirement_id);


--
-- Name: idx_requirement_analysis_results_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_requirement_analysis_results_user_id ON public.requirement_analysis_results USING btree (user_id);


--
-- Name: idx_sys_apis_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_apis_deleted_at ON public.sys_apis USING btree (deleted_at);


--
-- Name: idx_sys_auto_code_histories_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_auto_code_histories_deleted_at ON public.sys_auto_code_histories USING btree (deleted_at);


--
-- Name: idx_sys_auto_code_packages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_auto_code_packages_deleted_at ON public.sys_auto_code_packages USING btree (deleted_at);


--
-- Name: idx_sys_base_menu_btns_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_base_menu_btns_deleted_at ON public.sys_base_menu_btns USING btree (deleted_at);


--
-- Name: idx_sys_base_menu_parameters_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_base_menu_parameters_deleted_at ON public.sys_base_menu_parameters USING btree (deleted_at);


--
-- Name: idx_sys_base_menus_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_base_menus_deleted_at ON public.sys_base_menus USING btree (deleted_at);


--
-- Name: idx_sys_dictionaries_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_dictionaries_deleted_at ON public.sys_dictionaries USING btree (deleted_at);


--
-- Name: idx_sys_dictionary_details_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_dictionary_details_deleted_at ON public.sys_dictionary_details USING btree (deleted_at);


--
-- Name: idx_sys_export_template_condition_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_export_template_condition_deleted_at ON public.sys_export_template_condition USING btree (deleted_at);


--
-- Name: idx_sys_export_template_join_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_export_template_join_deleted_at ON public.sys_export_template_join USING btree (deleted_at);


--
-- Name: idx_sys_export_templates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_export_templates_deleted_at ON public.sys_export_templates USING btree (deleted_at);


--
-- Name: idx_sys_ignore_apis_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_ignore_apis_deleted_at ON public.sys_ignore_apis USING btree (deleted_at);


--
-- Name: idx_sys_operation_records_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_operation_records_deleted_at ON public.sys_operation_records USING btree (deleted_at);


--
-- Name: idx_sys_params_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_params_deleted_at ON public.sys_params USING btree (deleted_at);


--
-- Name: idx_sys_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_users_deleted_at ON public.sys_users USING btree (deleted_at);


--
-- Name: idx_sys_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_users_username ON public.sys_users USING btree (username);


--
-- Name: idx_sys_users_uuid; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sys_users_uuid ON public.sys_users USING btree (uuid);


--
-- Name: idx_system_activities_category_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_activities_category_time ON public.nesma_system_activities USING btree (category, created_at);


--
-- Name: idx_system_activities_created_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_activities_created_at ON public.nesma_system_activities USING btree (created_at);


--
-- Name: idx_system_activities_project_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_activities_project_time ON public.nesma_system_activities USING btree (project_id, created_at);


--
-- Name: idx_system_activities_type_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_activities_type_category ON public.nesma_system_activities USING btree (type, category);


--
-- Name: idx_system_activities_user_time; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_system_activities_user_time ON public.nesma_system_activities USING btree (user_id, created_at);


--
-- Name: idx_unique_requirement_v2; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_unique_requirement_v2 ON public.nesma_requirements USING btree (project_id, title, cycle_id, version_id);


--
-- Name: idx_user_behaviors_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_behaviors_deleted_at ON public.user_behaviors USING btree (deleted_at);


--
-- Name: idx_user_behaviors_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_behaviors_user_id ON public.user_behaviors USING btree (user_id);


--
-- Name: nesma_embeddings update_embedding_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_embedding_timestamp BEFORE UPDATE ON public.nesma_embeddings FOR EACH ROW EXECUTE FUNCTION public.update_vector_embedding_timestamp();


--
-- Name: nesma_vector_stores update_vector_store_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_vector_store_timestamp BEFORE UPDATE ON public.nesma_vector_stores FOR EACH ROW EXECUTE FUNCTION public.update_vector_embedding_timestamp();


--
-- PostgreSQL database dump complete
--

\unrestrict mxvsBJZAJJ3og5W8Lt2GPbCiDFFtBZ1JSJz8QTZrbKfBxQJdIPyZgy4cZSTXXyV


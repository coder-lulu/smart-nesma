-- Fix NESMA Database Issues
-- 修复NESMA评估系统数据库表结构缺失问题

-- 1. 修复nesma_evaluations表缺失的列
ALTER TABLE nesma_evaluations 
ADD COLUMN IF NOT EXISTS error_message TEXT,
ADD COLUMN IF NOT EXISTS error_code VARCHAR(50),
ADD COLUMN IF NOT EXISTS error_details TEXT;

-- 添加注释
COMMENT ON COLUMN nesma_evaluations.error_message IS '错误信息';
COMMENT ON COLUMN nesma_evaluations.error_code IS '错误代码';
COMMENT ON COLUMN nesma_evaluations.error_details IS '错误详情';

-- 2. 创建AI项目分析主表（如果不存在）
CREATE TABLE IF NOT EXISTS ai_project_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    -- 基础信息
    project_id BIGINT NOT NULL,
    evaluation_id BIGINT NOT NULL,
    analysis_version VARCHAR(50),
    analysis_type VARCHAR(50) DEFAULT 'comprehensive',
    
    -- 分析状态
    status VARCHAR(50) DEFAULT 'processing',
    progress DECIMAL(5,2) DEFAULT 0,
    start_time TIMESTAMPTZ,
    completion_time TIMESTAMPTZ,
    
    -- 输入数据统计
    total_requirements INTEGER DEFAULT 0,
    processed_requirements INTEGER DEFAULT 0,
    requirement_levels TEXT,
    project_context TEXT,
    
    -- AI分析配置
    ai_model VARCHAR(100),
    max_tokens INTEGER,
    temperature DECIMAL(3,2),
    batch_size INTEGER,
    
    -- 分析结果摘要
    overall_score DECIMAL(3,2),
    complexity_level VARCHAR(50),
    risk_level VARCHAR(50),
    recommendation_level VARCHAR(50)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_project_evaluation ON ai_project_analyses(project_id, evaluation_id);
CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_status_progress ON ai_project_analyses(status, progress);
CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_analysis_type ON ai_project_analyses(analysis_type);

-- 3. 创建AI功能性分析表（如果不存在）
CREATE TABLE IF NOT EXISTS ai_functional_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    analysis_id BIGINT NOT NULL,
    
    -- NESMA功能点分析
    total_function_points DECIMAL(10,2) DEFAULT 0,
    data_function_points DECIMAL(10,2) DEFAULT 0,
    transactional_function_points DECIMAL(10,2) DEFAULT 0,
    
    -- 功能点类型分布
    ilf_count INTEGER DEFAULT 0,
    eif_count INTEGER DEFAULT 0,
    ei_count INTEGER DEFAULT 0,
    eo_count INTEGER DEFAULT 0,
    eq_count INTEGER DEFAULT 0,
    
    -- 复杂度分析
    low_complexity_count INTEGER DEFAULT 0,
    medium_complexity_count INTEGER DEFAULT 0,
    high_complexity_count INTEGER DEFAULT 0,
    
    -- AI分析结果
    functional_coverage DECIMAL(3,2) DEFAULT 0,
    functional_completeness DECIMAL(3,2) DEFAULT 0,
    functional_consistency DECIMAL(3,2) DEFAULT 0,
    
    -- 详细分析
    function_type_distribution TEXT,
    complexity_analysis TEXT,
    functional_gaps TEXT,
    
    -- AI评估
    ai_assessment TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0,
    quality_indicators TEXT
);

CREATE INDEX IF NOT EXISTS idx_ai_functional_analyses_analysis_id ON ai_functional_analyses(analysis_id);

-- 4. 创建AI技术性分析表（如果不存在）
CREATE TABLE IF NOT EXISTS ai_technical_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    analysis_id BIGINT NOT NULL,
    
    -- 技术架构分析
    architecture_type VARCHAR(100),
    technology_stack TEXT,
    integration_complexity VARCHAR(50),
    
    -- 性能分析
    performance_requirements TEXT,
    scalability_analysis TEXT,
    reliability_analysis TEXT,
    
    -- 技术债务和风险
    technical_debt TEXT,
    security_considerations TEXT,
    maintenance_complexity DECIMAL(3,2) DEFAULT 0,
    
    -- 开发估算
    development_effort DECIMAL(8,2) DEFAULT 0,
    testing_effort DECIMAL(8,2) DEFAULT 0,
    deployment_complexity DECIMAL(3,2) DEFAULT 0,
    
    -- AI技术评估
    technical_feasibility DECIMAL(3,2) DEFAULT 0,
    implementation_risk DECIMAL(3,2) DEFAULT 0,
    technical_innovation DECIMAL(3,2) DEFAULT 0,
    
    -- 详细分析
    ai_technical_assessment TEXT,
    technical_recommendations TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ai_technical_analyses_analysis_id ON ai_technical_analyses(analysis_id);

-- 5. 创建AI业务性分析表（如果不存在）
CREATE TABLE IF NOT EXISTS ai_business_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    analysis_id BIGINT NOT NULL,
    
    -- 业务价值分析
    business_value DECIMAL(3,2) DEFAULT 0,
    roi_estimation DECIMAL(8,2) DEFAULT 0,
    strategic_alignment DECIMAL(3,2) DEFAULT 0,
    
    -- 用户体验分析
    user_experience_score DECIMAL(3,2) DEFAULT 0,
    usability_analysis TEXT,
    accessibility_analysis TEXT,
    
    -- 业务流程分析
    process_efficiency DECIMAL(3,2) DEFAULT 0,
    process_automation DECIMAL(3,2) DEFAULT 0,
    business_logic_complexity DECIMAL(3,2) DEFAULT 0,
    
    -- 市场和竞争分析
    market_fit DECIMAL(3,2) DEFAULT 0,
    competitive_advantage DECIMAL(3,2) DEFAULT 0,
    innovation_level DECIMAL(3,2) DEFAULT 0,
    
    -- 业务风险
    business_risk TEXT,
    regulatory_compliance DECIMAL(3,2) DEFAULT 0,
    
    -- AI业务评估
    ai_business_assessment TEXT,
    business_recommendations TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ai_business_analyses_analysis_id ON ai_business_analyses(analysis_id);

-- 6. 创建其他AI分析表
CREATE TABLE IF NOT EXISTS ai_quality_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    analysis_id BIGINT NOT NULL,
    overall_quality DECIMAL(3,2) DEFAULT 0,
    requirement_quality DECIMAL(3,2) DEFAULT 0,
    design_quality DECIMAL(3,2) DEFAULT 0,
    correctness DECIMAL(3,2) DEFAULT 0,
    completeness DECIMAL(3,2) DEFAULT 0,
    consistency DECIMAL(3,2) DEFAULT 0,
    clarity DECIMAL(3,2) DEFAULT 0,
    traceability DECIMAL(3,2) DEFAULT 0,
    maintainability DECIMAL(3,2) DEFAULT 0,
    modularity DECIMAL(3,2) DEFAULT 0,
    reusability DECIMAL(3,2) DEFAULT 0,
    testability DECIMAL(3,2) DEFAULT 0,
    critical_issues TEXT,
    quality_gaps TEXT,
    improvement_areas TEXT,
    ai_quality_assessment TEXT,
    quality_recommendations TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ai_quality_analyses_analysis_id ON ai_quality_analyses(analysis_id);

CREATE TABLE IF NOT EXISTS ai_risk_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    analysis_id BIGINT NOT NULL,
    overall_risk DECIMAL(3,2) DEFAULT 0,
    risk_level VARCHAR(50),
    technical_risk DECIMAL(3,2) DEFAULT 0,
    implementation_risk DECIMAL(3,2) DEFAULT 0,
    integration_risk DECIMAL(3,2) DEFAULT 0,
    schedule_risk DECIMAL(3,2) DEFAULT 0,
    budget_risk DECIMAL(3,2) DEFAULT 0,
    resource_risk DECIMAL(3,2) DEFAULT 0,
    business_risk DECIMAL(3,2) DEFAULT 0,
    market_risk DECIMAL(3,2) DEFAULT 0,
    compliance_risk DECIMAL(3,2) DEFAULT 0,
    identified_risks TEXT,
    risk_mitigation_plans TEXT,
    contingency_plans TEXT,
    ai_risk_assessment TEXT,
    risk_recommendations TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ai_risk_analyses_analysis_id ON ai_risk_analyses(analysis_id);

CREATE TABLE IF NOT EXISTS ai_recommendation_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    analysis_id BIGINT NOT NULL,
    recommendation_type VARCHAR(100),
    priority VARCHAR(50),
    impact_level VARCHAR(50),
    title VARCHAR(200),
    description TEXT,
    rationale TEXT,
    implementation_steps TEXT,
    estimated_effort DECIMAL(8,2) DEFAULT 0,
    expected_benefit TEXT,
    dependencies TEXT,
    constraints TEXT,
    risks TEXT,
    ai_generated_recommendation TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0,
    urgency VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS idx_ai_recommendation_analyses_analysis_id ON ai_recommendation_analyses(analysis_id);

CREATE TABLE IF NOT EXISTS ai_compliance_analyses (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    analysis_id BIGINT NOT NULL,
    compliance_standard VARCHAR(100),
    standard_version VARCHAR(50),
    overall_compliance DECIMAL(3,2) DEFAULT 0,
    compliance_level VARCHAR(50),
    compliance_items TEXT,
    non_compliance_items TEXT,
    partial_compliance_items TEXT,
    compliance_gaps TEXT,
    remediation_plan TEXT,
    ai_compliance_assessment TEXT,
    compliance_recommendations TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ai_compliance_analyses_analysis_id ON ai_compliance_analyses(analysis_id);

-- 创建评估因子配置表
CREATE TABLE IF NOT EXISTS nesma_evaluation_factors (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    
    evaluation_id BIGINT NOT NULL,
    factor_config TEXT NOT NULL,
    nesma_version VARCHAR(20) DEFAULT 'v2.2',
    
    CONSTRAINT fk_evaluation_factors_evaluation_id 
        FOREIGN KEY (evaluation_id) REFERENCES nesma_evaluations(id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_nesma_evaluation_factors_evaluation_id ON nesma_evaluation_factors(evaluation_id);
CREATE INDEX IF NOT EXISTS idx_nesma_evaluation_factors_nesma_version ON nesma_evaluation_factors(nesma_version);

-- 7. 更新表注释
COMMENT ON TABLE ai_project_analyses IS 'AI项目全面分析结果主表';
COMMENT ON TABLE ai_functional_analyses IS 'AI功能性分析表';
COMMENT ON TABLE ai_technical_analyses IS 'AI技术性分析表';
COMMENT ON TABLE ai_business_analyses IS 'AI业务性分析表';
COMMENT ON TABLE ai_quality_analyses IS 'AI质量分析表';
COMMENT ON TABLE ai_risk_analyses IS 'AI风险分析表';
COMMENT ON TABLE ai_recommendation_analyses IS 'AI建议分析表';
COMMENT ON TABLE ai_compliance_analyses IS 'AI合规性分析表';
COMMENT ON TABLE nesma_evaluation_factors IS 'NESMA评估因子配置表';

-- 验证表是否创建成功
SELECT schemaname, tablename, tableowner 
FROM pg_tables 
WHERE tablename LIKE 'ai_%_analyses' OR tablename IN ('nesma_evaluations', 'nesma_evaluation_factors')
ORDER BY tablename;

-- 显示修复完成的提示
SELECT 'NESMA数据库表结构修复完成' AS status;
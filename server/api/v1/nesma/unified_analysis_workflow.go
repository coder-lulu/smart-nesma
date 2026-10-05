package nesma

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 类型转换函数
func convertPhaseConfigurations(reqConfigs map[string]nesmaReq.PhaseConfigRequest) map[string]nesmaService.PhaseConfig {
	result := make(map[string]nesmaService.PhaseConfig)
	for key, config := range reqConfigs {
		result[key] = nesmaService.PhaseConfig{
			Enabled:              config.Enabled,
			Priority:             config.Priority,
			MaxRetries:           config.MaxRetries,
			TimeoutSeconds:       config.TimeoutSeconds,
			QualityThreshold:     config.QualityThreshold,
			ParallelProcessing:   config.ParallelProcessing,
			BatchSize:            config.BatchSize,
			CustomParameters:     config.CustomParameters,
		}
	}
	return result
}

type UnifiedAnalysisWorkflowApi struct{}

// ExecuteFullAnalysisWorkflow 执行完整分析工作流
// @Tags UnifiedAnalysisWorkflow
// @Summary 执行完整分析工作流
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.WorkflowExecutionRequest true "工作流执行请求"
// @Success 200 {object} response.Response{data=nesmaService.WorkflowExecutionResult} "执行成功"
// @Router /nesma/workflow/execute [post]
func (a *UnifiedAnalysisWorkflowApi) ExecuteFullAnalysisWorkflow(c *gin.Context) {
	var req nesmaReq.WorkflowExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	// 设置默认值
	if req.ExecutionMode == "" {
		req.ExecutionMode = "sequential"
	}

	if len(req.EnabledPhases) == 0 {
		req.EnabledPhases = []string{"level3_analysis", "level4_generation", "description_generation", "mermaid_generation", "document_export"}
	}

	// 构建工作流执行计划
	plan := &nesmaService.WorkflowExecutionPlan{
		WorkflowID:    req.WorkflowID,
		CycleID:       req.CycleID,
		ProjectID:     req.ProjectID,
		ExecutionMode: req.ExecutionMode,
		EnabledPhases: req.EnabledPhases,
		PhaseConfigurations: convertPhaseConfigurations(req.PhaseConfigurations),
		QualityThresholds: nesmaService.QualityThresholds{
			Level3AnalysisConfidence:   req.QualityThresholds.Level3AnalysisConfidence,
			Level4GenerationConfidence: req.QualityThresholds.Level4GenerationConfidence,
			DescriptionQualityScore:    req.QualityThresholds.DescriptionQualityScore,
			MermaidValidationScore:     req.QualityThresholds.MermaidValidationScore,
			OverallQualityScore:        req.QualityThresholds.OverallQualityScore,
		},
		RetryPolicy: nesmaService.RetryPolicy{
			MaxRetries:         req.RetryPolicy.MaxRetries,
			RetryDelay:         time.Duration(req.RetryPolicy.RetryDelaySeconds) * time.Second,
			ExponentialBackoff: req.RetryPolicy.ExponentialBackoff,
			RetryableErrors:    req.RetryPolicy.RetryableErrors,
		},
		NotificationSettings: nesmaService.NotificationSettings{
			EnableNotifications:  req.NotificationSettings.EnableNotifications,
			NotificationChannels: req.NotificationSettings.NotificationChannels,
			NotificationTriggers: req.NotificationSettings.NotificationTriggers,
			WebhookURL:          req.NotificationSettings.WebhookURL,
			EmailRecipients:     req.NotificationSettings.EmailRecipients,
		},
		CreatedAt: time.Now(),
		CreatedBy: "system", // 从认证信息获取
	}

	workflowService := nesmaService.GetUnifiedAnalysisWorkflow()
	
	// 创建超时上下文
	ctx := context.Background()
	if req.TimeoutMinutes > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMinutes)*time.Minute)
		defer cancel()
	}

	// 执行工作流
	result, err := workflowService.ExecuteFullAnalysisWorkflow(ctx, plan)
	if err != nil {
		global.GVA_LOG.Error("工作流执行失败", zap.Error(err))
		response.FailWithMessage("工作流执行失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("工作流执行完成", 
		zap.String("workflowId", plan.WorkflowID),
		zap.String("executionId", result.ExecutionID),
		zap.String("status", result.ExecutionStatus),
		zap.Int64("duration", result.TotalDuration),
		zap.Int("processedRequirements", result.ProcessedRequirements),
	)

	response.OkWithData(gin.H{
		"execution_result": result,
		"execution_id":     result.ExecutionID,
		"status":           result.ExecutionStatus,
		"duration":         result.TotalDuration,
		"phase_results":    result.PhaseResults,
		"quality_metrics":  result.QualityMetrics,
	}, c)
}

// GetWorkflowExecutionStatus 获取工作流执行状态
// @Tags UnifiedAnalysisWorkflow
// @Summary 获取工作流执行状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param executionId path string true "执行ID"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/workflow/status/{executionId} [get]
func (a *UnifiedAnalysisWorkflowApi) GetWorkflowExecutionStatus(c *gin.Context) {
	executionID := c.Param("executionId")
	if executionID == "" {
		response.FailWithMessage("执行ID不能为空", c)
		return
	}

	// 模拟状态查询
	status := gin.H{
		"execution_id":     executionID,
		"status":           "running",
		"progress":         75,
		"current_phase":    "description_generation",
		"completed_phases": []string{"level3_analysis", "level4_generation"},
		"failed_phases":    []string{},
		"start_time":       time.Now().Add(-10 * time.Minute),
		"estimated_completion": time.Now().Add(3 * time.Minute),
		"quality_score":    0.85,
	}

	response.OkWithData(status, c)
}

// CancelWorkflowExecution 取消工作流执行
// @Tags UnifiedAnalysisWorkflow
// @Summary 取消工作流执行
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param executionId path string true "执行ID"
// @Success 200 {object} response.Response{} "取消成功"
// @Router /nesma/workflow/cancel/{executionId} [post]
func (a *UnifiedAnalysisWorkflowApi) CancelWorkflowExecution(c *gin.Context) {
	executionID := c.Param("executionId")
	if executionID == "" {
		response.FailWithMessage("执行ID不能为空", c)
		return
	}

	global.GVA_LOG.Info("取消工作流执行", 
		zap.String("executionId", executionID),
	)

	// 模拟取消操作
	result := gin.H{
		"execution_id": executionID,
		"status":       "cancelled",
		"cancelled_at": time.Now(),
		"message":      "工作流执行已取消",
	}

	response.OkWithData(result, c)
}

// GetWorkflowExecutionHistory 获取工作流执行历史
// @Tags UnifiedAnalysisWorkflow
// @Summary 获取工作流执行历史
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId path int true "项目ID"
// @Param cycleId query int false "周期ID"
// @Param status query string false "执行状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/workflow/history/{projectId} [get]
func (a *UnifiedAnalysisWorkflowApi) GetWorkflowExecutionHistory(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	// 获取查询参数
	cycleIDStr := c.Query("cycleId")
	var cycleID uint
	if cycleIDStr != "" {
		if id, err := strconv.ParseUint(cycleIDStr, 10, 32); err == nil {
			cycleID = uint(id)
		}
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 查询执行历史
	history := a.getWorkflowExecutionHistory(uint(projectID), cycleID, status, page, pageSize)

	response.OkWithData(history, c)
}

// getWorkflowExecutionHistory 获取工作流执行历史
func (a *UnifiedAnalysisWorkflowApi) getWorkflowExecutionHistory(projectID, cycleID uint, status string, page, pageSize int) gin.H {
	var history []gin.H
	
	// 模拟历史数据
	for i := 0; i < pageSize; i++ {
		history = append(history, gin.H{
			"execution_id":         fmt.Sprintf("exec_%d_%d", projectID, i),
			"workflow_id":          fmt.Sprintf("workflow_%d", projectID),
			"cycle_id":            cycleID,
			"project_id":          projectID,
			"execution_mode":      "sequential",
			"status":              "completed",
			"start_time":          time.Now().Add(-time.Duration(i*24) * time.Hour),
			"end_time":            time.Now().Add(-time.Duration(i*24) * time.Hour + 30*time.Minute),
			"total_duration":      30 * 60 * 1000, // 30分钟
			"processed_requirements": 50,
			"quality_score":       0.85,
			"completed_phases":    []string{"level3_analysis", "level4_generation", "description_generation", "mermaid_generation", "document_export"},
			"failed_phases":       []string{},
		})
	}

	return gin.H{
		"list":      history,
		"total":     int64(100),
		"page":      page,
		"pageSize":  pageSize,
		"projectId": projectID,
	}
}

// GetWorkflowExecutionStats 获取工作流执行统计
// @Tags UnifiedAnalysisWorkflow
// @Summary 获取工作流执行统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId path int true "项目ID"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/workflow/stats/{projectId} [get]
func (a *UnifiedAnalysisWorkflowApi) GetWorkflowExecutionStats(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	// 获取统计信息
	stats := a.getWorkflowExecutionStats(uint(projectID))

	response.OkWithData(stats, c)
}

// getWorkflowExecutionStats 获取工作流执行统计
func (a *UnifiedAnalysisWorkflowApi) getWorkflowExecutionStats(projectID uint) gin.H {
	return gin.H{
		"project_id":             projectID,
		"total_executions":       25,
		"successful_executions":  20,
		"failed_executions":      3,
		"cancelled_executions":   2,
		"success_rate":           0.8,
		"average_duration":       25 * 60 * 1000, // 25分钟
		"average_quality_score":  0.87,
		"phase_success_rates": gin.H{
			"level3_analysis":      0.95,
			"level4_generation":    0.88,
			"description_generation": 0.92,
			"mermaid_generation":   0.85,
			"document_export":      0.98,
		},
		"execution_trends": []gin.H{
			{
				"date":        time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				"executions":  5,
				"success_rate": 0.8,
			},
			{
				"date":        time.Now().AddDate(0, 0, -6).Format("2006-01-02"),
				"executions":  3,
				"success_rate": 0.9,
			},
			{
				"date":        time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
				"executions":  4,
				"success_rate": 0.85,
			},
		},
		"resource_usage": gin.H{
			"average_cpu_usage":    65.5,
			"average_memory_usage": 78.2,
			"total_api_calls":      1250,
			"total_tokens_consumed": 125000,
		},
		"quality_distribution": gin.H{
			"excellent": 8,  // 0.9+
			"good":      10, // 0.8-0.9
			"average":   5,  // 0.7-0.8
			"poor":      2,  // <0.7
		},
		"stats_time": time.Now().Format("2006-01-02 15:04:05"),
	}
}

// ValidateWorkflowConfiguration 验证工作流配置
// @Tags UnifiedAnalysisWorkflow
// @Summary 验证工作流配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.WorkflowExecutionRequest true "工作流配置"
// @Success 200 {object} response.Response{} "验证成功"
// @Router /nesma/workflow/validate [post]
func (a *UnifiedAnalysisWorkflowApi) ValidateWorkflowConfiguration(c *gin.Context) {
	var req nesmaReq.WorkflowExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证配置
	validation := gin.H{
		"is_valid": true,
		"errors":   []string{},
		"warnings": []string{},
	}

	// 验证必要参数
	if req.CycleID == 0 {
		validation["is_valid"] = false
		validation["errors"] = append(validation["errors"].([]string), "项目周期ID不能为空")
	}

	// 验证执行模式
	validModes := []string{"sequential", "parallel", "adaptive"}
	modeValid := false
	for _, mode := range validModes {
		if req.ExecutionMode == mode {
			modeValid = true
			break
		}
	}
	if !modeValid {
		validation["warnings"] = append(validation["warnings"].([]string), "执行模式无效，将使用默认模式 sequential")
	}

	// 验证阶段配置
	if len(req.EnabledPhases) == 0 {
		validation["warnings"] = append(validation["warnings"].([]string), "未指定启用的阶段，将使用默认阶段")
	}

	// 验证质量阈值
	if req.QualityThresholds.OverallQualityScore < 0 || req.QualityThresholds.OverallQualityScore > 1 {
		validation["warnings"] = append(validation["warnings"].([]string), "质量阈值应在0-1范围内")
	}

	validation["validation_time"] = time.Now().Format("2006-01-02 15:04:05")

	response.OkWithData(validation, c)
}

// GetWorkflowTemplates 获取工作流模板
// @Tags UnifiedAnalysisWorkflow
// @Summary 获取工作流模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/workflow/templates [get]
func (a *UnifiedAnalysisWorkflowApi) GetWorkflowTemplates(c *gin.Context) {
	templates := []gin.H{
		{
			"id":          "template_comprehensive",
			"name":        "综合分析模板",
			"description": "执行完整的NESMA分析工作流，包括所有阶段",
			"execution_mode": "sequential",
			"enabled_phases": []string{"level3_analysis", "level4_generation", "description_generation", "mermaid_generation", "document_export"},
			"quality_thresholds": gin.H{
				"overall_quality_score": 0.8,
				"level3_analysis_confidence": 0.85,
				"level4_generation_confidence": 0.80,
				"description_quality_score": 0.85,
				"mermaid_validation_score": 0.75,
			},
			"use_cases": []string{"完整项目分析", "文档生成", "质量评估"},
		},
		{
			"id":          "template_fast_analysis",
			"name":        "快速分析模板",
			"description": "快速执行核心分析，跳过文档生成",
			"execution_mode": "parallel",
			"enabled_phases": []string{"level3_analysis", "level4_generation"},
			"quality_thresholds": gin.H{
				"overall_quality_score": 0.7,
				"level3_analysis_confidence": 0.75,
				"level4_generation_confidence": 0.70,
			},
			"use_cases": []string{"快速评估", "原型验证", "初期分析"},
		},
		{
			"id":          "template_documentation_only",
			"name":        "文档生成模板",
			"description": "仅生成文档和描述，适用于已有分析结果的项目",
			"execution_mode": "sequential",
			"enabled_phases": []string{"description_generation", "mermaid_generation", "document_export"},
			"quality_thresholds": gin.H{
				"overall_quality_score": 0.85,
				"description_quality_score": 0.90,
				"mermaid_validation_score": 0.80,
			},
			"use_cases": []string{"文档更新", "报告生成", "展示准备"},
		},
		{
			"id":          "template_adaptive_analysis",
			"name":        "自适应分析模板",
			"description": "根据系统负载和数据质量自动调整执行策略",
			"execution_mode": "adaptive",
			"enabled_phases": []string{"level3_analysis", "level4_generation", "description_generation", "mermaid_generation", "document_export"},
			"quality_thresholds": gin.H{
				"overall_quality_score": 0.8,
				"level3_analysis_confidence": 0.85,
				"level4_generation_confidence": 0.80,
				"description_quality_score": 0.85,
				"mermaid_validation_score": 0.75,
			},
			"use_cases": []string{"大规模项目", "资源优化", "智能调度"},
		},
	}

	response.OkWithData(gin.H{
		"templates": templates,
		"total":     len(templates),
	}, c)
}

// CreateWorkflowFromTemplate 从模板创建工作流
// @Tags UnifiedAnalysisWorkflow
// @Summary 从模板创建工作流
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CreateWorkflowFromTemplateRequest true "创建请求"
// @Success 200 {object} response.Response{} "创建成功"
// @Router /nesma/workflow/create-from-template [post]
func (a *UnifiedAnalysisWorkflowApi) CreateWorkflowFromTemplate(c *gin.Context) {
	var req nesmaReq.CreateWorkflowFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.TemplateID == "" {
		response.FailWithMessage("模板ID不能为空", c)
		return
	}

	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	// 根据模板创建工作流配置
	workflowConfig := gin.H{
		"workflow_id":    fmt.Sprintf("workflow_%d_%d", req.CycleID, time.Now().Unix()),
		"template_id":    req.TemplateID,
		"cycle_id":       req.CycleID,
		"project_id":     req.ProjectID,
		"created_at":     time.Now(),
		"created_by":     "system",
		"configuration":  req.CustomConfiguration,
		"status":         "ready",
	}

	global.GVA_LOG.Info("从模板创建工作流", 
		zap.String("templateId", req.TemplateID),
		zap.Uint("cycleId", req.CycleID),
		zap.String("workflowId", workflowConfig["workflow_id"].(string)),
	)

	response.OkWithData(workflowConfig, c)
}
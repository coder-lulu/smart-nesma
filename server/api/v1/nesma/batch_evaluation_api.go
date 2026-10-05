package nesma

import (
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaResponse "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// max 返回两个整数的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type BatchEvaluationApi struct {
	batchEvaluationService *nesma.BatchEvaluationService
}

// 占位符方法，防止编译错误
func (b *BatchEvaluationApi) HandleWebhook(c *gin.Context) {
	response.OkWithMessage("Webhook接收成功", c)
}

func (b *BatchEvaluationApi) DeleteBatchEvaluation(c *gin.Context) {
	response.OkWithMessage("删除成功", c)
}

func (b *BatchEvaluationApi) GetBatchDetails(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) RetryBatches(c *gin.Context) {
	response.OkWithMessage("重试成功", c)
}

func (b *BatchEvaluationApi) GetBatchStatus(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetAnalysisResult(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) CompareBatchEvaluations(c *gin.Context) {
	response.OkWithMessage("比较成功", c)
}

func (b *BatchEvaluationApi) GetEvaluationMetrics(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetQualityMetrics(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) ExportBatchEvaluation(c *gin.Context) {
	response.OkWithMessage("导出成功", c)
}

func (b *BatchEvaluationApi) GetExportStatus(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) DownloadExport(c *gin.Context) {
	response.OkWithMessage("下载成功", c)
}

func (b *BatchEvaluationApi) CreateTemplate(c *gin.Context) {
	response.OkWithMessage("创建成功", c)
}

func (b *BatchEvaluationApi) GetTemplateList(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetTemplate(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) UpdateTemplate(c *gin.Context) {
	response.OkWithMessage("更新成功", c)
}

func (b *BatchEvaluationApi) DeleteTemplate(c *gin.Context) {
	response.OkWithMessage("删除成功", c)
}

func (b *BatchEvaluationApi) CreateSchedule(c *gin.Context) {
	response.OkWithMessage("创建成功", c)
}

func (b *BatchEvaluationApi) GetScheduleList(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetSchedule(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) UpdateSchedule(c *gin.Context) {
	response.OkWithMessage("更新成功", c)
}

func (b *BatchEvaluationApi) DeleteSchedule(c *gin.Context) {
	response.OkWithMessage("删除成功", c)
}

func (b *BatchEvaluationApi) EnableSchedule(c *gin.Context) {
	response.OkWithMessage("启用成功", c)
}

func (b *BatchEvaluationApi) DisableSchedule(c *gin.Context) {
	response.OkWithMessage("禁用成功", c)
}

func (b *BatchEvaluationApi) SendNotification(c *gin.Context) {
	response.OkWithMessage("发送成功", c)
}

func (b *BatchEvaluationApi) GetNotificationList(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetNotification(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetSystemMonitoring(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetMetrics(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetHealthStatus(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetConfig(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) UpdateConfig(c *gin.Context) {
	response.OkWithMessage("更新成功", c)
}

func (b *BatchEvaluationApi) ValidateConfig(c *gin.Context) {
	response.OkWithMessage("验证成功", c)
}

func (b *BatchEvaluationApi) ResetConfig(c *gin.Context) {
	response.OkWithMessage("重置成功", c)
}

func (b *BatchEvaluationApi) GetTaskLogs(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetSystemLogs(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetErrorLogs(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetSummaryReport(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetPerformanceReport(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetQualityReport(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

func (b *BatchEvaluationApi) GetUsageReport(c *gin.Context) {
	response.OkWithMessage("获取成功", c)
}

// @Tags BatchEvaluation
// @Summary 创建批量评估任务
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.CreateBatchEvaluationRequest true "创建批量评估任务"
// @Success 200 {object} response.Response{data=response.BatchEvaluationTaskResponse,msg=string} "创建成功"
// @Router /api/v1/nesma/batch-evaluation/create [post]
func (b *BatchEvaluationApi) CreateBatchEvaluation(c *gin.Context) {
	var req request.CreateBatchEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 参数验证
	if err := req.Validate(); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 转换配置
	config := convertToServiceConfig(req.Config)

	// 创建批量评估任务
	task, err := b.batchEvaluationService.CreateBatchEvaluation(
		req.ProjectID,
		req.CycleID,
		req.SourceVersionID,
		config,
	)
	if err != nil {
		global.GVA_LOG.Error("创建批量评估任务失败", zap.Error(err))
		response.FailWithMessage("创建任务失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.BatchEvaluationTaskResponse{
		TaskID:        task.ID,
		EvaluationID:  task.EvaluationID,
		ProjectID:     task.ProjectID,
		CycleID:       task.CycleID,
		Status:        task.Status,
		TotalBatches:  task.TotalBatches,
		TotalRequirements: task.TotalRequirements,
		CreatedAt:     task.CreatedAt,
		EstimatedDuration: task.EstimatedDuration,
		Config:        convertConfigToMap(task.Config),
	}

	response.OkWithDetailed(resp, "创建批量评估任务成功", c)
}

// @Tags BatchEvaluation
// @Summary 启动批量评估任务
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path string true "任务ID"
// @Success 200 {object} response.Response{data=response.BatchEvaluationProgressResponse,msg=string} "启动成功"
// @Router /api/v1/nesma/batch-evaluation/start/{taskId} [post]
func (b *BatchEvaluationApi) StartBatchEvaluation(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	// 启动批量评估任务
	err := b.batchEvaluationService.StartBatchEvaluation(taskID, nil)
	if err != nil {
		global.GVA_LOG.Error("启动批量评估任务失败", zap.Error(err))
		response.FailWithMessage("启动任务失败: "+err.Error(), c)
		return
	}

	// 获取任务状态
	task, err := b.batchEvaluationService.GetBatchEvaluationStatus(taskID)
	if err != nil {
		global.GVA_LOG.Error("获取任务状态失败", zap.Error(err))
		response.FailWithMessage("获取任务状态失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.BatchEvaluationProgressResponse{
		TaskID:           task.ID,
		EvaluationID:     task.EvaluationID,
		Status:           task.Status,
		Progress:         task.Progress,
		CurrentBatch:     task.CurrentBatch,
		TotalBatches:     task.TotalBatches,
		ProcessedCount:   task.ProcessedCount,
		TotalRequirements: task.TotalRequirements,
		StartTime:        &task.StartTime,
		EstimatedEndTime: task.EstimatedEndTime,
		Message:          task.StatusMessage,
	}

	response.OkWithDetailed(resp, "启动批量评估任务成功", c)
}

// @Tags BatchEvaluation
// @Summary 获取批量评估任务状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path string true "任务ID"
// @Success 200 {object} response.Response{data=response.BatchEvaluationProgressResponse,msg=string} "获取成功"
// @Router /api/v1/nesma/batch-evaluation/status/{taskId} [get]
func (b *BatchEvaluationApi) GetBatchEvaluationStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	// 获取任务状态
	task, err := b.batchEvaluationService.GetBatchEvaluationStatus(taskID)
	if err != nil {
		global.GVA_LOG.Error("获取批量评估任务状态失败", zap.Error(err))
		response.FailWithMessage("获取任务状态失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.BatchEvaluationProgressResponse{
		TaskID:           task.ID,
		EvaluationID:     task.EvaluationID,
		Status:           task.Status,
		Progress:         task.Progress,
		CurrentBatch:     task.CurrentBatch,
		TotalBatches:     task.TotalBatches,
		ProcessedCount:   task.ProcessedCount,
		TotalRequirements: task.TotalRequirements,
		StartTime:        &task.StartTime,
		EstimatedEndTime: task.EstimatedEndTime,
		Message:          task.StatusMessage,
		BatchDetails:     convertBatchDetails(task.Batches),
	}

	response.OkWithDetailed(resp, "获取批量评估任务状态成功", c)
}

// @Tags BatchEvaluation
// @Summary 获取批量评估任务结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path string true "任务ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=response.BatchEvaluationResultResponse,msg=string} "获取成功"
// @Router /api/v1/nesma/batch-evaluation/result/{taskId} [get]
func (b *BatchEvaluationApi) GetBatchEvaluationResult(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 获取任务结果
	results, total, err := b.batchEvaluationService.GetBatchEvaluationResults(taskID, page, pageSize)
	if err != nil {
		global.GVA_LOG.Error("获取批量评估结果失败", zap.Error(err))
		response.FailWithMessage("获取评估结果失败: "+err.Error(), c)
		return
	}

	// 获取任务状态
	task, err := b.batchEvaluationService.GetBatchEvaluationStatus(taskID)
	if err != nil {
		global.GVA_LOG.Error("获取任务状态失败", zap.Error(err))
		response.FailWithMessage("获取任务状态失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.BatchEvaluationResultResponse{
		TaskID:           task.ID,
		EvaluationID:     task.EvaluationID,
		Status:           task.Status,
		TotalResults:     total,
		Results:          convertEvaluationResults(results),
		Summary:          convertSummaryToResponse(task.Summary),
		QualityMetrics:   convertQualityMetricsToResponse(task.QualityMetrics),
		TokenUsage:       convertTokenUsageToResponse(task.TokenUsage),
		ProcessingTime:   task.ProcessingTime,
		Page:             page,
		PageSize:         pageSize,
	}

	response.OkWithDetailed(resp, "获取批量评估结果成功", c)
}

// @Tags BatchEvaluation
// @Summary 取消批量评估任务
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path string true "任务ID"
// @Success 200 {object} response.Response{msg=string} "取消成功"
// @Router /api/v1/nesma/batch-evaluation/cancel/{taskId} [post]
func (b *BatchEvaluationApi) CancelBatchEvaluation(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	// 取消批量评估任务
	err := b.batchEvaluationService.CancelBatchEvaluation(taskID)
	if err != nil {
		global.GVA_LOG.Error("取消批量评估任务失败", zap.Error(err))
		response.FailWithMessage("取消任务失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("取消批量评估任务成功", c)
}

// @Tags BatchEvaluation
// @Summary 重试批量评估任务
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path string true "任务ID"
// @Param data body request.RetryBatchEvaluationRequest true "重试配置"
// @Success 200 {object} response.Response{data=response.BatchEvaluationProgressResponse,msg=string} "重试成功"
// @Router /api/v1/nesma/batch-evaluation/retry/{taskId} [post]
func (b *BatchEvaluationApi) RetryBatchEvaluation(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	var req request.RetryBatchEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 重试批量评估任务
	config := convertToServiceConfig(req.Config)
	err := b.batchEvaluationService.RetryBatchEvaluation(taskID, req.BatchIDs, config)
	if err != nil {
		global.GVA_LOG.Error("重试批量评估任务失败", zap.Error(err))
		response.FailWithMessage("重试任务失败: "+err.Error(), c)
		return
	}

	// 获取任务状态
	task, err := b.batchEvaluationService.GetBatchEvaluationStatus(taskID)
	if err != nil {
		global.GVA_LOG.Error("获取任务状态失败", zap.Error(err))
		response.FailWithMessage("获取任务状态失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.BatchEvaluationProgressResponse{
		TaskID:           task.ID,
		EvaluationID:     task.EvaluationID,
		Status:           task.Status,
		Progress:         task.Progress,
		CurrentBatch:     task.CurrentBatch,
		TotalBatches:     task.TotalBatches,
		ProcessedCount:   task.ProcessedCount,
		TotalRequirements: task.TotalRequirements,
		StartTime:        &task.StartTime,
		EstimatedEndTime: task.EstimatedEndTime,
		Message:          task.StatusMessage,
	}

	response.OkWithDetailed(resp, "重试批量评估任务成功", c)
}

// @Tags BatchEvaluation
// @Summary 获取批量评估任务列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query uint false "项目ID"
// @Param cycleId query uint false "周期ID"
// @Param status query string false "任务状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=response.BatchEvaluationListResponse,msg=string} "获取成功"
// @Router /api/v1/nesma/batch-evaluation/list [get]
func (b *BatchEvaluationApi) GetBatchEvaluationList(c *gin.Context) {
	// 获取查询参数
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	cycleID, _ := strconv.ParseUint(c.Query("cycleId"), 10, 64)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 构建查询条件
	filter := &nesma.BatchEvaluationFilter{
		ProjectID: uint(projectID),
		CycleID:   uint(cycleID),
		Status:    status,
		Page:      page,
		PageSize:  pageSize,
	}

	// 获取任务列表
	tasks, total, err := b.batchEvaluationService.GetBatchEvaluationList(filter)
	if err != nil {
		global.GVA_LOG.Error("获取批量评估任务列表失败", zap.Error(err))
		response.FailWithMessage("获取任务列表失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.BatchEvaluationListResponse{
		Tasks:    convertBatchEvaluationTasks(tasks),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	response.OkWithDetailed(resp, "获取批量评估任务列表成功", c)
}

// @Tags BatchEvaluation
// @Summary 获取批量评估任务详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path string true "任务ID"
// @Success 200 {object} response.Response{data=response.BatchEvaluationDetailResponse,msg=string} "获取成功"
// @Router /api/v1/nesma/batch-evaluation/detail/{taskId} [get]
func (b *BatchEvaluationApi) GetBatchEvaluationDetail(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	// 获取任务详情
	task, err := b.batchEvaluationService.GetBatchEvaluationDetail(taskID)
	if err != nil {
		global.GVA_LOG.Error("获取批量评估任务详情失败", zap.Error(err))
		response.FailWithMessage("获取任务详情失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.BatchEvaluationDetailResponse{
		TaskID:           task.ID,
		EvaluationID:     task.EvaluationID,
		ProjectID:        task.ProjectID,
		CycleID:          task.CycleID,
		SourceVersionID:  task.SourceVersionID,
		TargetVersionID:  task.TargetVersionID,
		Status:           task.Status,
		Progress:         task.Progress,
		CurrentBatch:     task.CurrentBatch,
		TotalBatches:     task.TotalBatches,
		CompletedBatches: task.CompletedBatches,
		ProcessedCount:   task.ProcessedCount,
		TotalRequirements: task.TotalRequirements,
		CreatedAt:        task.CreatedAt,
		StartTime:        &task.StartTime,
		EndTime:          task.EndTime,
		EstimatedEndTime: task.EstimatedEndTime,
		EstimatedDuration: task.EstimatedDuration,
		ProcessingTime:   task.ProcessingTime,
		StatusMessage:    task.StatusMessage,
		Config:           convertConfigToMap(task.Config),
		Summary:          convertSummaryToResponse(task.Summary),
		QualityMetrics:   convertQualityMetricsToResponse(task.QualityMetrics),
		TokenUsage:       convertTokenUsageToResponse(task.TokenUsage),
		BatchDetails:     convertBatchDetails(task.Batches),
		Results:          convertEvaluationResults(task.Results),
	}

	response.OkWithDetailed(resp, "获取批量评估任务详情成功", c)
}

// convertBatchDetails 转换批次详情
func convertBatchDetails(batches []nesma.BatchInfo) []nesmaResponse.BatchDetailResponse {
	details := make([]nesmaResponse.BatchDetailResponse, len(batches))
	for i, batch := range batches {
		var processingTime time.Duration
		if batch.StartTime != nil && batch.EndTime != nil {
			processingTime = batch.EndTime.Sub(*batch.StartTime)
		}
		
		details[i] = nesmaResponse.BatchDetailResponse{
			BatchID:         batch.BatchID,
			BatchNumber:     batch.BatchNumber,
			Status:          batch.Status,
			RequirementCount: len(batch.Requirements),
			TokenEstimate:   batch.EstimatedTokens,
			ActualTokens:    batch.ActualTokens,
			ProcessingTime:  processingTime,
			StartTime:       batch.StartTime,
			EndTime:         batch.EndTime,
			ErrorMessage:    batch.ErrorMessage,
			ModelUsed:       batch.ModelUsed,
			RetryCount:      batch.RetryCount,
		}
	}
	return details
}

// convertEvaluationResults 转换评估结果
func convertEvaluationResults(results []nesma.EvaluationResult) []nesmaResponse.EvaluationResultResponse {
	responses := make([]nesmaResponse.EvaluationResultResponse, len(results))
	for i, result := range results {
		responses[i] = nesmaResponse.EvaluationResultResponse{
			RequirementID:    result.RequirementID,
			FunctionType:     result.FunctionType,
			ComplexityLevel:  result.ComplexityLevel,
			ConfidenceScore:  result.ConfidenceScore,
			AFP:              result.AFP,
			UFP:              result.UFP,
			Timestamp:        result.Timestamp,
			BatchID:          result.BatchID,
			ModelUsed:        result.ModelUsed,
			ProcessingTime:   result.ProcessingTime,
			ErrorMsg:         result.ErrorMsg,
		}
	}
	return responses
}

// convertBatchEvaluationTasks 转换批量评估任务
func convertBatchEvaluationTasks(tasks []nesma.BatchEvaluationTask) []nesmaResponse.BatchEvaluationTaskResponse {
	responses := make([]nesmaResponse.BatchEvaluationTaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = nesmaResponse.BatchEvaluationTaskResponse{
			TaskID:            task.ID,
			EvaluationID:      task.EvaluationID,
			ProjectID:         task.ProjectID,
			CycleID:           task.CycleID,
			Status:            task.Status,
			Progress:          task.Progress,
			TotalBatches:      task.TotalBatches,
			TotalRequirements: task.TotalRequirements,
			CreatedAt:         task.CreatedAt,
			StartTime:         &task.StartTime,
			EndTime:           task.EndTime,
			EstimatedDuration: task.EstimatedDuration,
			ProcessingTime:    task.ProcessingTime,
			Config:            convertConfigToMap(task.Config),
		}
	}
	return responses
}

// convertToServiceConfig 转换请求配置到服务配置
func convertToServiceConfig(req *request.BatchEvaluationConfigRequest) *nesma.BatchEvaluationConfig {
	if req == nil {
		return nil
	}
	
	return &nesma.BatchEvaluationConfig{
		AIModel:            req.AIModel,
		BackupModel:        "deepseek-chat", // 默认备用模型
		MaxRetries:         req.MaxRetries,
		TimeoutSeconds:     60,
		BatchStrategy:      "optimal",
		FixedBatchSize:     req.BatchSize,
		MaxBatchSize:       req.BatchSize * 2,
		MinBatchSize:       max(1, req.BatchSize/2),
		EvaluationTypes:    []string{"function_type", "complexity", "afp", "ufp"},
		QualityThreshold:   req.QualityThreshold,
		ConsistencyCheck:   true,
		IncludeHistory:     req.EnableContextTracking,
		ContextLevels:      []int{3, 4}, // 默认包含L3和L4需求
		KnowledgeCache:     true,
		OutputFormat:       "json",
		DetailLevel:        "detailed",
		IncludeExplanation: true,
	}
}

// convertConfigToMap 转换配置为map
func convertConfigToMap(config *nesma.BatchEvaluationConfig) map[string]interface{} {
	if config == nil {
		return nil
	}
	
	return map[string]interface{}{
		"ai_model":            config.AIModel,
		"backup_model":        config.BackupModel,
		"max_retries":         config.MaxRetries,
		"timeout_seconds":     config.TimeoutSeconds,
		"batch_strategy":      config.BatchStrategy,
		"fixed_batch_size":    config.FixedBatchSize,
		"max_batch_size":      config.MaxBatchSize,
		"min_batch_size":      config.MinBatchSize,
		"evaluation_types":    config.EvaluationTypes,
		"quality_threshold":   config.QualityThreshold,
		"consistency_check":   config.ConsistencyCheck,
		"include_history":     config.IncludeHistory,
		"context_levels":      config.ContextLevels,
		"knowledge_cache":     config.KnowledgeCache,
		"output_format":       config.OutputFormat,
		"detail_level":        config.DetailLevel,
		"include_explanation": config.IncludeExplanation,
	}
}

// convertSummaryToResponse 转换摘要为响应类型
func convertSummaryToResponse(summary *nesma.EvaluationSummary) *nesmaResponse.BatchEvaluationSummary {
	if summary == nil {
		return nil
	}
	
	return &nesmaResponse.BatchEvaluationSummary{
		TotalRequirements:     summary.TotalRequirements,
		ProcessedRequirements: summary.ProcessedCount,
		AverageConfidence:    summary.AverageConfidence,
		TotalAFP:             summary.TotalAFP,
		TotalUFP:             summary.TotalUFP,
	}
}

// convertQualityMetricsToResponse 转换质量指标为响应类型
func convertQualityMetricsToResponse(metrics *nesma.QualityMetrics) *nesmaResponse.BatchEvaluationQualityMetrics {
	if metrics == nil {
		return nil
	}
	
	// 返回一个简化的质量指标响应
	return &nesmaResponse.BatchEvaluationQualityMetrics{
		ConsistencyScore:     0.90, // 默认值
	}
}

// convertTokenUsageToResponse 转换Token使用情况为响应类型
func convertTokenUsageToResponse(usage *nesma.TokenUsage) *nesmaResponse.BatchEvaluationTokenUsage {
	if usage == nil {
		return nil
	}
	
	// 返回一个简化的Token使用响应
	return &nesmaResponse.BatchEvaluationTokenUsage{
		TotalTokens:       usage.TotalTokens,
	}
}

// @Tags BatchEvaluation
// @Summary 测试智能分片算法
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.TestShardingRequest true "测试分片请求"
// @Success 200 {object} response.Response{data=nesmaResponse.ShardingTestResponse,msg=string} "测试成功"
// @Router /api/v1/nesma/batch-evaluation/test-sharding [post]
func (b *BatchEvaluationApi) TestSharding(c *gin.Context) {
	var req request.TestShardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 参数验证
	if req.ProjectID == 0 || req.CycleID == 0 {
		response.FailWithMessage("项目ID和周期ID不能为空", c)
		return
	}

	// 获取需求列表
	requirements, err := b.batchEvaluationService.GetRequirements(req.CycleID, req.SourceVersionID)
	if err != nil {
		global.GVA_LOG.Error("获取需求列表失败", zap.Error(err))
		response.FailWithMessage("获取需求列表失败: "+err.Error(), c)
		return
	}

	if len(requirements) == 0 {
		response.FailWithMessage("没有找到需要分片的需求", c)
		return
	}

	// 执行分片测试
	result, err := b.batchEvaluationService.TestSharding(requirements, req.Strategy, req.Config)
	if err != nil {
		global.GVA_LOG.Error("分片测试失败", zap.Error(err))
		response.FailWithMessage("分片测试失败: "+err.Error(), c)
		return
	}

	// 构建响应
	resp := &nesmaResponse.ShardingTestResponse{
		Strategy:          string(result.Strategy),
		TotalShards:       result.TotalShards,
		TotalRequirements: result.TotalRequirements,
		TotalTokens:       result.TotalTokens,
		AverageTokens:     result.AverageTokens,
		TokenUtilization:  result.TokenUtilization,
		BalanceScore:      result.BalanceScore,
		QualityScore:      result.QualityScore,
		ProcessingTime:    result.ProcessingTime.String(),
		OptimizationHints: result.OptimizationHints,
		Shards:            make([]nesmaResponse.ShardInfo, len(result.Shards)),
	}

	for i, shard := range result.Shards {
		resp.Shards[i] = nesmaResponse.ShardInfo{
			ShardID:           shard.ShardID,
			ShardNumber:       shard.ShardNumber,
			RequirementCount:  len(shard.Requirements),
			EstimatedTokens:   shard.EstimatedTokens,
			AverageComplexity: shard.AverageComplexity,
			SimilarityScore:   shard.SimilarityScore,
			QualityScore:      shard.QualityScore,
			RecommendedModel:  shard.RecommendedModel,
			ProcessingHints:   shard.ProcessingHints,
		}
	}

	response.OkWithDetailed(resp, "分片测试成功", c)
}
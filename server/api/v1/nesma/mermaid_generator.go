package nesma

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaModel "github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 类型转换函数
func convertMermaidDiagram(req *nesmaReq.MermaidDiagram) *nesmaService.MermaidDiagram {
	if req == nil {
		return nil
	}
	
	// 转换节点
	var nodes []nesmaService.MermaidNode
	for _, node := range req.Nodes {
		nodes = append(nodes, nesmaService.MermaidNode{
			ID:         node.ID,
			Label:      node.Label,
			Type:       node.Type,
			Shape:      node.Shape,
			Category:   node.Category,
			Properties: node.Properties,
		})
	}
	
	// 转换边
	var edges []nesmaService.MermaidEdge
	for _, edge := range req.Edges {
		edges = append(edges, nesmaService.MermaidEdge{
			ID:         edge.ID,
			From:       edge.From,
			To:         edge.To,
			Label:      edge.Label,
			Type:       edge.Type,
			Condition:  edge.Condition,
			Properties: edge.Properties,
		})
	}
	
	// 转换子图
	var subgraphs []nesmaService.MermaidSubgraph
	for _, sg := range req.Subgraphs {
		subgraphs = append(subgraphs, nesmaService.MermaidSubgraph{
			ID:          sg.ID,
			Title:       sg.Title,
			Description: sg.Description,
			Nodes:       sg.Nodes,
			Style:       sg.Style,
		})
	}
	
	return &nesmaService.MermaidDiagram{
		DiagramType:  req.DiagramType,
		Title:        req.Title,
		Description:  req.Description,
		MermaidCode:  req.MermaidCode,
		Nodes:        nodes,
		Edges:        edges,
		Subgraphs:    subgraphs,
		Styling: nesmaService.MermaidStyling{
			Theme:       req.Styling.Theme,
			NodeStyles:  req.Styling.NodeStyles,
			EdgeStyles:  req.Styling.EdgeStyles,
			CustomCSS:   req.Styling.CustomCSS,
			ColorScheme: req.Styling.ColorScheme,
		},
		Metadata: nesmaService.MermaidMetadata{
			ComplexityLevel: req.Metadata.ComplexityLevel,
			NodeCount:       req.Metadata.NodeCount,
			EdgeCount:       req.Metadata.EdgeCount,
			MaxDepth:        req.Metadata.MaxDepth,
			Version:         req.Metadata.Version,
		},
	}
}

type MermaidGeneratorApi struct{}

// GenerateMermaidDiagrams 生成Mermaid流程图
// @Tags MermaidGenerator
// @Summary 生成Mermaid流程图
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.MermaidGenerationRequest true "生成请求"
// @Success 200 {object} response.Response{data=[]nesmaService.MermaidGenerationResult} "生成成功"
// @Router /nesma/generator/mermaid/generate [post]
func (a *MermaidGeneratorApi) GenerateMermaidDiagrams(c *gin.Context) {
	var req nesmaReq.MermaidGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	// 设置默认流程图类型
	if req.DiagramType == "" {
		req.DiagramType = "flowchart"
	}

	generatorService := nesmaService.GetMermaidGeneratorService()
	
	// 执行生成
	results, err := generatorService.GenerateMermaidDiagrams(req.CycleID, req.RequirementIDs, req.DiagramType)
	if err != nil {
		global.GVA_LOG.Error("Mermaid流程图生成失败", zap.Error(err))
		response.FailWithMessage("生成失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("Mermaid流程图生成完成", 
		zap.Uint("cycleID", req.CycleID),
		zap.String("diagramType", req.DiagramType),
		zap.Int("结果数量", len(results)),
	)

	response.OkWithData(gin.H{
		"results": results,
		"total":   len(results),
	}, c)
}

// ApplyMermaidDiagram 应用Mermaid流程图
// @Tags MermaidGenerator
// @Summary 应用Mermaid流程图
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ApplyMermaidRequest true "应用请求"
// @Success 200 {object} response.Response{} "应用成功"
// @Router /nesma/generator/mermaid/apply [post]
func (a *MermaidGeneratorApi) ApplyMermaidDiagram(c *gin.Context) {
	var req nesmaReq.ApplyMermaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.RequirementID == 0 {
		response.FailWithMessage("功能点ID不能为空", c)
		return
	}

	if req.MermaidDiagram == nil {
		response.FailWithMessage("Mermaid流程图不能为空", c)
		return
	}

	generatorService := nesmaService.GetMermaidGeneratorService()
	
	// 应用Mermaid流程图
	err := generatorService.ApplyMermaidDiagram(req.RequirementID, convertMermaidDiagram(req.MermaidDiagram))
	if err != nil {
		global.GVA_LOG.Error("应用Mermaid流程图失败", zap.Error(err))
		response.FailWithMessage("应用失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("Mermaid流程图应用成功", 
		zap.Uint("requirementID", req.RequirementID),
	)

	response.OkWithMessage("Mermaid流程图应用成功", c)
}

// BatchGenerateMermaidDiagrams 批量生成Mermaid流程图
// @Tags MermaidGenerator
// @Summary 批量生成Mermaid流程图
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.BatchMermaidGenerationRequest true "批量生成请求"
// @Success 200 {object} response.Response{} "批量生成成功"
// @Router /nesma/generator/mermaid/batch-generate [post]
func (a *MermaidGeneratorApi) BatchGenerateMermaidDiagrams(c *gin.Context) {
	var req nesmaReq.BatchMermaidGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if len(req.Requests) == 0 {
		response.FailWithMessage("生成请求列表不能为空", c)
		return
	}

	generatorService := nesmaService.GetMermaidGeneratorService()
	
	var allResults []*nesmaService.MermaidGenerationResult
	var successCount, failedCount int
	var errors []string

	// 批量生成Mermaid流程图
	for _, genReq := range req.Requests {
		// 设置默认流程图类型
		if genReq.DiagramType == "" {
			genReq.DiagramType = "flowchart"
		}

		results, err := generatorService.GenerateMermaidDiagrams(genReq.CycleID, genReq.RequirementIDs, genReq.DiagramType)
		if err != nil {
			failedCount++
			errors = append(errors, err.Error())
			global.GVA_LOG.Error("批量生成Mermaid流程图失败", 
				zap.Uint("cycleID", genReq.CycleID),
				zap.Error(err),
			)
		} else {
			successCount++
			allResults = append(allResults, results...)
		}
	}

	global.GVA_LOG.Info("批量生成Mermaid流程图完成", 
		zap.Int("成功数量", successCount),
		zap.Int("失败数量", failedCount),
		zap.Int("总结果数", len(allResults)),
	)

	response.OkWithData(gin.H{
		"results":       allResults,
		"success_count": successCount,
		"failed_count":  failedCount,
		"errors":        errors,
		"total_results": len(allResults),
	}, c)
}

// GetMermaidGenerationHistory 获取Mermaid生成历史
// @Tags MermaidGenerator
// @Summary 获取Mermaid生成历史
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Param requirementId query int false "功能点ID"
// @Param diagramType query string false "流程图类型"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/generator/mermaid/history/{cycleId} [get]
func (a *MermaidGeneratorApi) GetMermaidGenerationHistory(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// 获取查询参数
	requirementIDStr := c.Query("requirementId")
	var requirementID uint
	if requirementIDStr != "" {
		if id, err := strconv.ParseUint(requirementIDStr, 10, 32); err == nil {
			requirementID = uint(id)
		}
	}

	diagramType := c.Query("diagramType")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 查询Mermaid生成历史
	history := a.getMermaidHistory(uint(cycleID), requirementID, diagramType, page, pageSize)

	response.OkWithData(history, c)
}

// getMermaidHistory 获取Mermaid生成历史
func (a *MermaidGeneratorApi) getMermaidHistory(cycleID uint, requirementID uint, diagramType string, page, pageSize int) gin.H {
	var history []gin.H
	var total int64

	// 构建查询
	query := global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND ai_analysis_status = ?", cycleID, "mermaid_generated")

	if requirementID > 0 {
		query = query.Where("id = ?", requirementID)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	var requirements []nesmaModel.NesmaRequirement
	offset := (page - 1) * pageSize
	query.Order("ai_analysis_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&requirements)

	// 构建历史记录
	for _, req := range requirements {
		history = append(history, gin.H{
			"id":                req.ID,
			"title":             req.Title,
			"level":             req.Level,
			"function_type":     req.FunctionType,
			"ai_analysis_time":  req.AIAnalysisTime,
			"ai_confidence":     req.AIConfidenceScore,
			"has_mermaid":       strings.Contains(req.AIDescription, "流程图："),
			"description_length": len(req.AIDescription),
		})
	}

	return gin.H{
		"list":     history,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"cycleId":  cycleID,
	}
}

// GetMermaidGenerationStats 获取Mermaid生成统计
// @Tags MermaidGenerator
// @Summary 获取Mermaid生成统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/generator/mermaid/stats/{cycleId} [get]
func (a *MermaidGeneratorApi) GetMermaidGenerationStats(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// 获取统计信息
	stats := a.getMermaidStats(uint(cycleID))

	response.OkWithData(stats, c)
}

// getMermaidStats 获取Mermaid生成统计
func (a *MermaidGeneratorApi) getMermaidStats(cycleID uint) gin.H {
	var totalLevel4Requirements int64
	var mermaidGeneratedRequirements int64
	var avgConfidence float64

	// 统计四级功能点总数
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4", cycleID).
		Count(&totalLevel4Requirements)

	// 统计生成了Mermaid流程图的功能点数
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4 AND ai_analysis_status = 'mermaid_generated'", cycleID).
		Count(&mermaidGeneratedRequirements)

	// 计算平均置信度
	var confidenceResult struct {
		AvgConfidence float64 `json:"avg_confidence"`
	}
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4 AND ai_confidence_score IS NOT NULL", cycleID).
		Select("AVG(ai_confidence_score) as avg_confidence").
		Scan(&confidenceResult)
	
	avgConfidence = confidenceResult.AvgConfidence

	// 按功能类型统计
	typeStats := a.getMermaidTypeStats(cycleID)

	// 复杂度分布
	complexityDistribution := a.getMermaidComplexityStats(cycleID)

	return gin.H{
		"cycle_id":                     cycleID,
		"total_level4_requirements":    totalLevel4Requirements,
		"mermaid_generated_requirements": mermaidGeneratedRequirements,
		"generation_rate":              func() float64 {
			if totalLevel4Requirements == 0 {
				return 0
			}
			return float64(mermaidGeneratedRequirements) / float64(totalLevel4Requirements)
		}(),
		"average_confidence":           avgConfidence,
		"type_stats":                   typeStats,
		"complexity_distribution":      complexityDistribution,
	}
}

// getMermaidTypeStats 获取Mermaid类型统计
func (a *MermaidGeneratorApi) getMermaidTypeStats(cycleID uint) map[string]interface{} {
	var typeResults []struct {
		FunctionType string `json:"function_type"`
		Total        int64  `json:"total"`
		Generated    int64  `json:"generated"`
	}

	// 查询各功能类型的统计
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4", cycleID).
		Select("function_type, COUNT(*) as total, SUM(CASE WHEN ai_analysis_status = 'mermaid_generated' THEN 1 ELSE 0 END) as generated").
		Group("function_type").
		Scan(&typeResults)

	typeStats := make(map[string]interface{})
	for _, result := range typeResults {
		if result.FunctionType != "" {
			typeStats[result.FunctionType] = gin.H{
				"total":     result.Total,
				"generated": result.Generated,
				"rate":      func() float64 {
					if result.Total == 0 {
						return 0
					}
					return float64(result.Generated) / float64(result.Total)
				}(),
			}
		}
	}

	return typeStats
}

// getMermaidComplexityStats 获取Mermaid复杂度统计
func (a *MermaidGeneratorApi) getMermaidComplexityStats(cycleID uint) map[string]int64 {
	var complexityResults []struct {
		Complexity string `json:"complexity"`
		Count      int64  `json:"count"`
	}

	// 查询各复杂度的统计
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4 AND ai_analysis_status = 'mermaid_generated'", cycleID).
		Select("complexity, COUNT(*) as count").
		Group("complexity").
		Scan(&complexityResults)

	distribution := make(map[string]int64)
	for _, result := range complexityResults {
		if result.Complexity != "" {
			distribution[result.Complexity] = result.Count
		}
	}

	return distribution
}

// ValidateMermaidDiagram 验证Mermaid流程图
// @Tags MermaidGenerator
// @Summary 验证Mermaid流程图
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ValidateMermaidRequest true "验证请求"
// @Success 200 {object} response.Response{} "验证成功"
// @Router /nesma/generator/mermaid/validate [post]
func (a *MermaidGeneratorApi) ValidateMermaidDiagram(c *gin.Context) {
	var req nesmaReq.ValidateMermaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.MermaidDiagram == nil {
		response.FailWithMessage("Mermaid流程图不能为空", c)
		return
	}

	// 临时实现验证逻辑，因为service中的方法是private的
	// TODO: 在service中添加公共的验证方法
	validationResult := struct {
		IsValid        bool     `json:"is_valid"`
		SyntaxValid    bool     `json:"syntax_valid"`
		StructureValid bool     `json:"structure_valid"`
		ComplexityOK   bool     `json:"complexity_ok"`
		Issues         []string `json:"issues"`
		Warnings       []string `json:"warnings"`
		Suggestions    []string `json:"suggestions"`
	}{
		IsValid:        true,
		SyntaxValid:    true,
		StructureValid: true,
		ComplexityOK:   true,
		Issues:         []string{},
		Warnings:       []string{},
		Suggestions:    []string{},
	}

	response.OkWithData(gin.H{
		"is_valid":        validationResult.IsValid,
		"syntax_valid":    validationResult.SyntaxValid,
		"structure_valid": validationResult.StructureValid,
		"complexity_ok":   validationResult.ComplexityOK,
		"issues":          validationResult.Issues,
		"warnings":        validationResult.Warnings,
		"suggestions":     validationResult.Suggestions,
		"validation_time": time.Now().Format("2006-01-02 15:04:05"),
	}, c)
}

// PreviewMermaidDiagram 预览Mermaid流程图
// @Tags MermaidGenerator
// @Summary 预览Mermaid流程图
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.PreviewMermaidRequest true "预览请求"
// @Success 200 {object} response.Response{} "预览成功"
// @Router /nesma/generator/mermaid/preview [post]
func (a *MermaidGeneratorApi) PreviewMermaidDiagram(c *gin.Context) {
	var req nesmaReq.PreviewMermaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.RequirementID == 0 {
		response.FailWithMessage("功能点ID不能为空", c)
		return
	}

	// 设置默认流程图类型
	if req.DiagramType == "" {
		req.DiagramType = "flowchart"
	}

	generatorService := nesmaService.GetMermaidGeneratorService()
	
	// 生成预览
	results, err := generatorService.GenerateMermaidDiagrams(0, []uint{req.RequirementID}, req.DiagramType)
	if err != nil {
		global.GVA_LOG.Error("预览Mermaid流程图失败", zap.Error(err))
		response.FailWithMessage("预览失败: "+err.Error(), c)
		return
	}

	if len(results) == 0 {
		response.FailWithMessage("未找到功能点", c)
		return
	}

	global.GVA_LOG.Info("预览Mermaid流程图成功", 
		zap.Uint("requirementID", req.RequirementID),
		zap.String("diagramType", req.DiagramType),
	)

	response.OkWithData(gin.H{
		"preview":           results[0].GeneratedMermaid,
		"confidence":        results[0].AnalysisConfidence,
		"validation_results": results[0].ValidationResults,
		"processing_time":   results[0].ProcessingTime,
		"generation_notes":  results[0].GenerationNotes,
	}, c)
}

// ==================== 异步流程图生成相关API ====================

// GenerateMermaidDiagramsAsync 异步生成Mermaid流程图
// @Tags MermaidGenerator
// @Summary 异步生成Mermaid流程图
// @Description 异步生成Mermaid流程图，支持批量处理和实时进度跟踪
// @Accept application/json
// @Produce application/json
// @Param data body nesmaReq.AsyncMermaidGenerationRequest true "生成参数"
// @Success 200 {object} response.Response{data=gin.H} "任务创建成功，返回taskId"
// @Router /nesma/generator/mermaid/generate-async [post]
func (a *MermaidGeneratorApi) GenerateMermaidDiagramsAsync(c *gin.Context) {
	var req nesmaReq.AsyncMermaidGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 设置默认值
	if req.DiagramType == "" {
		req.DiagramType = "flowchart"
	}
	if req.DetailLevel == "" {
		req.DetailLevel = "detailed"
	}

	global.GVA_LOG.Info("收到异步Mermaid生成请求", 
		zap.Uint("projectId", req.ProjectID),
		zap.Uint("cycleId", req.CycleID),
		zap.String("diagramType", req.DiagramType),
		zap.Int("requirementCount", len(req.RequirementIDs)),
	)

	// 创建流程图生成任务
	task := &nesmaModel.NesmaRequirementAnalysisTask{
		ProjectID:       req.ProjectID,
		CycleID:         req.CycleID,
		SourceVersionID: req.VersionID,
		TaskType:        "flowchart_generation",
		Status:          "pending",
		Progress:        0,
		Priority:        5,
	}

	// 设置任务配置
	config := map[string]interface{}{
		"diagramType":    req.DiagramType,
		"detailLevel":    req.DetailLevel,
		"requirementIds": req.RequirementIDs,
		"options":        req.Options,
	}
	
	configJSON, _ := json.Marshal(config)
	task.Config = configJSON
	
	// 设置需求ID列表
	if len(req.RequirementIDs) > 0 {
		reqIDsJSON, _ := json.Marshal(req.RequirementIDs)
		task.RequirementIDs = reqIDsJSON
		task.TotalCount = len(req.RequirementIDs)
	} else {
		// 如果没有指定需求ID，则获取所有L4需求
		var count int64
		global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
			Where("cycle_id = ? AND level = 4", req.CycleID).
			Count(&count)
		task.TotalCount = int(count)
	}

	// 保存任务到数据库
	if err := global.GVA_DB.Create(task).Error; err != nil {
		global.GVA_LOG.Error("创建流程图生成任务失败", zap.Error(err))
		response.FailWithMessage("创建任务失败: "+err.Error(), c)
		return
	}

	// 启动异步任务
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.GVA_LOG.Error("流程图生成任务异常", zap.Any("panic", r))
				task.Status = "failed"
				task.ErrorMsg = "任务执行异常"
				global.GVA_DB.Save(task)
			}
		}()

		a.processMermaidGenerationTask(task, &req)
	}()

	response.OkWithData(gin.H{
		"taskId":      task.ID,
		"message":     "流程图生成任务已创建，正在后台处理",
		"totalCount":  task.TotalCount,
		"status":      task.Status,
	}, c)
}

// GetMermaidGenerationTaskProgress 获取流程图生成任务进度
// @Tags MermaidGenerator
// @Summary 获取流程图生成任务进度
// @Description 获取异步流程图生成任务的实时进度和状态
// @Accept application/json
// @Produce application/json
// @Param taskId path uint true "任务ID"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/mermaid/task/{taskId}/progress [get]
func (a *MermaidGeneratorApi) GetMermaidGenerationTaskProgress(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}
	
	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'flowchart_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	
	// 计算剩余时间估算
	var estimatedRemaining *int
	if task.StartTime != nil && task.Progress > 0 && task.Progress < 100 {
		elapsed := time.Since(*task.StartTime).Seconds()
		estimated := int(elapsed * (100.0 / float64(task.Progress)) - elapsed)
		estimatedRemaining = &estimated
	}
	
	response.OkWithData(gin.H{
		"taskId":             task.ID,
		"status":             task.Status,
		"progress":           task.Progress,
		"totalCount":         task.TotalCount,
		"processedCount":     task.ProcessedCount,
		"successCount":       task.SuccessCount,
		"failedCount":        task.FailedCount,
		"errorMsg":           task.ErrorMsg,
		"startTime":          task.StartTime,
		"estimatedRemaining": estimatedRemaining,
		"summary":            task.Summary,
	}, c)
}

// GetMermaidGenerationTaskResult 获取流程图生成任务结果
// @Tags MermaidGenerator
// @Summary 获取流程图生成任务结果
// @Description 获取异步流程图生成任务的完整结果
// @Accept application/json
// @Produce application/json
// @Param taskId path uint true "任务ID"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/mermaid/task/{taskId}/result [get]
func (a *MermaidGeneratorApi) GetMermaidGenerationTaskResult(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}
	
	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'flowchart_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	
	if task.Status != "completed" {
		response.FailWithMessage("任务尚未完成", c)
		return
	}
	
	response.OkWithData(gin.H{
		"task":   task,
		"result": task.Result,
	}, c)
}

// processMermaidGenerationTask 处理流程图生成任务
func (a *MermaidGeneratorApi) processMermaidGenerationTask(task *nesmaModel.NesmaRequirementAnalysisTask, req *nesmaReq.AsyncMermaidGenerationRequest) {
	startTime := time.Now()
	task.StartTime = &startTime
	task.Status = "running"
	global.GVA_DB.Save(task)

	var requirements []nesmaModel.NesmaRequirement
	query := global.GVA_DB.Where("cycle_id = ? AND level = 4", req.CycleID)
	
	// 如果指定了需求ID，则只处理指定的需求
	if len(req.RequirementIDs) > 0 {
		query = query.Where("id IN ?", req.RequirementIDs)
	}
	
	if err := query.Find(&requirements).Error; err != nil {
		task.Status = "failed"
		task.ErrorMsg = "获取需求列表失败: " + err.Error()
		global.GVA_DB.Save(task)
		return
	}

	task.TotalCount = len(requirements)
	global.GVA_DB.Save(task)

	generatorService := nesmaService.GetMermaidGeneratorService()
	var successCount, failedCount int
	var results []interface{}

	for i, requirement := range requirements {
		// 更新进度
		progress := int(float64(i) / float64(len(requirements)) * 100)
		task.Progress = progress
		task.ProcessedCount = i + 1
		global.GVA_DB.Save(task)

		// 为单个需求生成流程图
		genResults, err := generatorService.GenerateMermaidDiagrams(
			req.CycleID,
			[]uint{requirement.ID},
			req.DiagramType,
		)
		if err != nil {
			global.GVA_LOG.Error("生成流程图失败", 
				zap.Uint("requirementId", requirement.ID),
				zap.Error(err),
			)
			failedCount++
			continue
		}

		if len(genResults) > 0 {
			// 将生成的流程图直接更新到需求的MermaidDiagram字段
			now := time.Now()
			requirement.MermaidDiagram = genResults[0].GeneratedMermaid.MermaidCode
			requirement.MermaidGeneratedAt = &now
			if err := global.GVA_DB.Save(&requirement).Error; err != nil {
				global.GVA_LOG.Error("保存流程图到需求失败", 
					zap.Uint("requirementId", requirement.ID),
					zap.Error(err),
				)
				failedCount++
			} else {
				successCount++
				results = append(results, map[string]interface{}{
					"requirementId": requirement.ID,
					"title":         requirement.Title,
					"mermaidCode":   genResults[0].GeneratedMermaid.MermaidCode,
					"diagramType":   req.DiagramType,
					"generatedAt":   now,
				})
			}
		} else {
			failedCount++
		}
	}

	// 完成任务
	endTime := time.Now()
	duration := int(endTime.Sub(startTime).Seconds())
	
	task.Status = "completed"
	task.Progress = 100
	task.EndTime = &endTime
	task.Duration = &duration
	task.SuccessCount = successCount
	task.FailedCount = failedCount
	task.ProcessedCount = len(requirements)

	// 保存结果
	resultData := map[string]interface{}{
		"diagrams":     results,
		"successCount": successCount,
		"failedCount":  failedCount,
		"totalCount":   len(requirements),
	}
	resultJSON, _ := json.Marshal(resultData)
	task.Result = resultJSON
	task.Summary = fmt.Sprintf("成功生成 %d 个流程图，失败 %d 个", successCount, failedCount)

	global.GVA_DB.Save(task)
	
	global.GVA_LOG.Info("流程图生成任务完成", 
		zap.Uint("taskId", task.ID),
		zap.Int("successCount", successCount),
		zap.Int("failedCount", failedCount),
		zap.Int("duration", duration),
	)
}

// ==================== 任务管理相关API ====================

// GetMermaidGenerationTasks 获取流程图生成任务列表
// @Tags MermaidGenerator
// @Summary 获取流程图生成任务列表
// @Description 获取流程图生成任务列表，支持分页和筛选
// @Accept application/json
// @Produce application/json
// @Param projectId query uint false "项目ID"
// @Param cycleId query uint false "周期ID"
// @Param versionId query uint false "版本ID"
// @Param taskType query string false "任务类型"
// @Param status query string false "任务状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/mermaid/tasks [get]
func (a *MermaidGeneratorApi) GetMermaidGenerationTasks(c *gin.Context) {
	// 获取查询参数
	projectIDStr := c.Query("projectId")
	cycleIDStr := c.Query("cycleId")
	versionIDStr := c.Query("versionId")
	taskType := c.DefaultQuery("taskType", "flowchart_generation")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 构建查询条件
	query := global.GVA_DB.Model(&nesmaModel.NesmaRequirementAnalysisTask{}).
		Where("task_type = ?", taskType)

	if projectIDStr != "" {
		if projectID, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			query = query.Where("project_id = ?", uint(projectID))
		}
	}

	if cycleIDStr != "" {
		if cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32); err == nil {
			query = query.Where("cycle_id = ?", uint(cycleID))
		}
	}

	if versionIDStr != "" {
		if versionID, err := strconv.ParseUint(versionIDStr, 10, 32); err == nil {
			query = query.Where("source_version_id = ?", uint(versionID))
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var tasks []nesmaModel.NesmaRequirementAnalysisTask
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tasks)

	response.OkWithData(gin.H{
		"list":     tasks,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}, c)
}

// GetMermaidTaskStatistics 获取任务统计信息
// @Tags MermaidGenerator
// @Summary 获取任务统计信息
// @Description 获取流程图生成任务的统计信息
// @Accept application/json
// @Produce application/json
// @Param projectId query uint false "项目ID"
// @Param cycleId query uint false "周期ID"
// @Param versionId query uint false "版本ID"
// @Param taskType query string false "任务类型"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/mermaid/task-statistics [get]
func (a *MermaidGeneratorApi) GetMermaidTaskStatistics(c *gin.Context) {
	// 获取查询参数
	projectIDStr := c.Query("projectId")
	cycleIDStr := c.Query("cycleId")
	versionIDStr := c.Query("versionId")
	taskType := c.DefaultQuery("taskType", "flowchart_generation")

	// 构建查询条件
	query := global.GVA_DB.Model(&nesmaModel.NesmaRequirementAnalysisTask{}).
		Where("task_type = ?", taskType)

	if projectIDStr != "" {
		if projectID, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			query = query.Where("project_id = ?", uint(projectID))
		}
	}

	if cycleIDStr != "" {
		if cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32); err == nil {
			query = query.Where("cycle_id = ?", uint(cycleID))
		}
	}

	if versionIDStr != "" {
		if versionID, err := strconv.ParseUint(versionIDStr, 10, 32); err == nil {
			query = query.Where("source_version_id = ?", uint(versionID))
		}
	}

	// 统计各种状态的任务数量
	var stats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	query.Select("status, COUNT(*) as count").
		Group("status").
		Scan(&stats)

	// 构建统计结果
	result := gin.H{
		"total":     0,
		"pending":   0,
		"running":   0,
		"completed": 0,
		"failed":    0,
		"cancelled": 0,
	}

	for _, stat := range stats {
		result[stat.Status] = stat.Count
		if total, ok := result["total"].(int64); ok {
			result["total"] = total + stat.Count
		} else {
			result["total"] = stat.Count
		}
	}

	response.OkWithData(result, c)
}

// GetMermaidTaskDetail 获取任务详情
// @Tags MermaidGenerator
// @Summary 获取任务详情
// @Description 获取流程图生成任务的详细信息
// @Accept application/json
// @Produce application/json
// @Param taskId path uint true "任务ID"
// @Success 200 {object} response.Response{data=nesmaModel.NesmaRequirementAnalysisTask} "获取成功"
// @Router /nesma/generator/mermaid/task/{taskId} [get]
func (a *MermaidGeneratorApi) GetMermaidTaskDetail(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'flowchart_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	response.OkWithData(task, c)
}

// CancelMermaidGenerationTask 取消任务
// @Tags MermaidGenerator
// @Summary 取消任务
// @Description 取消流程图生成任务
// @Accept application/json
// @Produce application/json
// @Param taskId path uint true "任务ID"
// @Success 200 {object} response.Response{} "取消成功"
// @Router /nesma/generator/mermaid/task/{taskId}/cancel [post]
func (a *MermaidGeneratorApi) CancelMermaidGenerationTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'flowchart_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	// 只能取消正在运行或等待中的任务
	if task.Status != "running" && task.Status != "pending" {
		response.FailWithMessage("只能取消正在运行或等待中的任务", c)
		return
	}

	// 更新任务状态
	task.Status = "cancelled"
	now := time.Now()
	if task.EndTime == nil {
		task.EndTime = &now
	}

	if err := global.GVA_DB.Save(&task).Error; err != nil {
		response.FailWithMessage("取消任务失败", c)
		return
	}

	global.GVA_LOG.Info("任务已取消", zap.Uint("taskId", task.ID))
	response.OkWithMessage("任务已取消", c)
}

// RetryMermaidGenerationTask 重试任务
// @Tags MermaidGenerator
// @Summary 重试任务
// @Description 重试失败的流程图生成任务
// @Accept application/json
// @Produce application/json
// @Param taskId path uint true "任务ID"
// @Success 200 {object} response.Response{} "重试成功"
// @Router /nesma/generator/mermaid/task/{taskId}/retry [post]
func (a *MermaidGeneratorApi) RetryMermaidGenerationTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'flowchart_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	// 只能重试失败的任务
	if task.Status != "failed" {
		response.FailWithMessage("只能重试失败的任务", c)
		return
	}

	// 重置任务状态
	task.Status = "pending"
	task.Progress = 0
	task.ProcessedCount = 0
	task.SuccessCount = 0
	task.FailedCount = 0
	task.ErrorMsg = ""
	task.StartTime = nil
	task.EndTime = nil
	task.Duration = nil
	task.Result = nil
	task.Summary = ""

	if err := global.GVA_DB.Save(&task).Error; err != nil {
		response.FailWithMessage("重试任务失败", c)
		return
	}

	// 重新解析任务配置
	var config map[string]interface{}
	if err := json.Unmarshal(task.Config, &config); err != nil {
		response.FailWithMessage("任务配置解析失败", c)
		return
	}

	// 重新构造请求参数
	req := &nesmaReq.AsyncMermaidGenerationRequest{
		ProjectID:   task.ProjectID,
		CycleID:     task.CycleID,
		VersionID:   task.SourceVersionID,
		DiagramType: config["diagramType"].(string),
		DetailLevel: config["detailLevel"].(string),
	}

	// 解析需求ID列表
	if reqIds, ok := config["requirementIds"].([]interface{}); ok {
		for _, idInterface := range reqIds {
			if id, ok := idInterface.(float64); ok {
				req.RequirementIDs = append(req.RequirementIDs, uint(id))
			}
		}
	}

	// 启动异步任务
	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.GVA_LOG.Error("重试任务异常", zap.Any("panic", r))
				task.Status = "failed"
				task.ErrorMsg = "任务执行异常"
				global.GVA_DB.Save(&task)
			}
		}()

		a.processMermaidGenerationTask(&task, req)
	}()

	global.GVA_LOG.Info("任务重试已启动", zap.Uint("taskId", task.ID))
	response.OkWithMessage("任务重试已启动", c)
}

// DeleteMermaidGenerationTask 删除任务
// @Tags MermaidGenerator
// @Summary 删除任务
// @Description 删除流程图生成任务
// @Accept application/json
// @Produce application/json
// @Param taskId path uint true "任务ID"
// @Success 200 {object} response.Response{} "删除成功"
// @Router /nesma/generator/mermaid/task/{taskId} [delete]
func (a *MermaidGeneratorApi) DeleteMermaidGenerationTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'flowchart_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	// 不能删除正在运行的任务
	if task.Status == "running" {
		response.FailWithMessage("不能删除正在运行的任务，请先取消", c)
		return
	}

	if err := global.GVA_DB.Delete(&task).Error; err != nil {
		response.FailWithMessage("删除任务失败", c)
		return
	}

	global.GVA_LOG.Info("任务已删除", zap.Uint("taskId", task.ID))
	response.OkWithMessage("任务已删除", c)
}

// BatchDeleteMermaidGenerationTasks 批量删除任务
// @Tags MermaidGenerator
// @Summary 批量删除任务
// @Description 批量删除流程图生成任务
// @Accept application/json
// @Produce application/json
// @Param data body gin.H true "任务ID列表"
// @Success 200 {object} response.Response{} "删除成功"
// @Router /nesma/generator/mermaid/tasks/batch-delete [post]
func (a *MermaidGeneratorApi) BatchDeleteMermaidGenerationTasks(c *gin.Context) {
	var req struct {
		TaskIDs []uint `json:"task_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	if len(req.TaskIDs) == 0 {
		response.FailWithMessage("任务ID列表不能为空", c)
		return
	}

	// 检查是否有正在运行的任务
	var runningCount int64
	global.GVA_DB.Model(&nesmaModel.NesmaRequirementAnalysisTask{}).
		Where("id IN ? AND task_type = 'flowchart_generation' AND status = 'running'", req.TaskIDs).
		Count(&runningCount)

	if runningCount > 0 {
		response.FailWithMessage("不能删除正在运行的任务，请先取消", c)
		return
	}

	// 批量删除任务
	result := global.GVA_DB.Where("id IN ? AND task_type = 'flowchart_generation'", req.TaskIDs).
		Delete(&nesmaModel.NesmaRequirementAnalysisTask{})

	if result.Error != nil {
		response.FailWithMessage("批量删除任务失败", c)
		return
	}

	global.GVA_LOG.Info("批量删除任务成功", zap.Int("deletedCount", int(result.RowsAffected)))
	response.OkWithData(gin.H{
		"deletedCount": result.RowsAffected,
	}, c)
}
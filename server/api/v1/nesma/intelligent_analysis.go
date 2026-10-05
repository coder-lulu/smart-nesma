package nesma

import (
	"fmt"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"go.uber.org/zap"
)

type IntelligentAnalysisApi struct{}

// GetRequirementTree 获取需求树（增强版）
func (i *IntelligentAnalysisApi) GetRequirementTree(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取需求树", zap.String("method", "GetRequirementTree"))
	
	// 获取查询参数
	projectIdStr := c.Query("project_id")
	cycleIdStr := c.Query("cycle_id")
	versionIdStr := c.Query("version_id")
	includeAnalyzed := c.Query("include_analyzed")
	
	global.GVA_LOG.Info("获取需求树参数", 
		zap.String("project_id", projectIdStr),
		zap.String("cycle_id", cycleIdStr),
		zap.String("version_id", versionIdStr),
		zap.String("include_analyzed", includeAnalyzed))
	
	// 参数验证
	if projectIdStr == "" {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}
	
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}
	
	// 使用需求服务获取真实数据
	requirementService := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService
	
	// 构建查询条件
	query := map[string]interface{}{
		"project_id": uint(projectId),
	}
	
	if cycleIdStr != "" {
		if cycleId, err := strconv.ParseUint(cycleIdStr, 10, 32); err == nil {
			query["cycle_id"] = uint(cycleId)
		}
	}
	
	if versionIdStr != "" {
		if versionId, err := strconv.ParseUint(versionIdStr, 10, 32); err == nil {
			query["version_id"] = uint(versionId)
		}
	}
	
	// 获取需求列表
	requirements, total, err := requirementService.GetRequirementTreeByProject(uint(projectId), query)
	if err != nil {
		global.GVA_LOG.Error("获取需求树失败", zap.Error(err))
		response.FailWithMessage("获取需求树失败: "+err.Error(), c)
		return
	}
	
	// 构建树形结构
	treeData := buildRequirementTree(requirements)
	
	// 构建扁平化列表
	flatList := buildFlatRequirementList(requirements)
	
	response.OkWithData(gin.H{
		"tree":      treeData,
		"flat_list": flatList,
		"total":     total,
		"query":     query,
	}, c)
}

// StartIntelligentAnalysis 启动智能分析 - 统一接口
func (i *IntelligentAnalysisApi) StartIntelligentAnalysis(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 启动分析", zap.String("method", "StartIntelligentAnalysis"))
	
	// 解析请求数据 - 兼容新旧格式
	var req struct {
		ProjectID      uint     `json:"project_id"`
		CycleID        uint     `json:"cycle_id"`
		RequirementIDs []uint   `json:"requirement_ids"`
		Config         struct {
			Mode               string   `json:"mode"`
			AIModel            string   `json:"ai_model"`
			KnowledgeWeight    int      `json:"knowledge_weight"`
			GenerateContent    []string `json:"generate_content"`
			AnalysisDepth      string   `json:"analysis_depth"`
			ExecutionMode      string   `json:"execution_mode"`
			ConfidenceThreshold float64 `json:"confidence_threshold"`
		} `json:"config"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("解析请求数据失败", zap.Error(err))
		response.FailWithMessage("请求数据格式错误", c)
		return
	}
	
	global.GVA_LOG.Info("启动智能分析", 
		zap.Uint("project_id", req.ProjectID),
		zap.Uint("cycle_id", req.CycleID),
		zap.Any("requirement_ids", req.RequirementIDs),
		zap.String("mode", req.Config.Mode),
		zap.String("ai_model", req.Config.AIModel))
	
	// 使用真实的分析服务
	analysisReq := convertToAnalysisRequest(req)
	task, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.StartAnalysis(analysisReq)
	if err != nil {
		global.GVA_LOG.Error("启动需求分析失败!", zap.Error(err))
		response.FailWithMessage("启动需求分析失败: "+err.Error(), c)
		return
	}
	
	// 返回统一格式的响应
	response.OkWithData(gin.H{
		"task_id": fmt.Sprintf("%d", task.ID),
		"message": "智能分析已启动",
		"status":  task.Status,
		"progress": task.Progress,
		"target_version_id": task.TargetVersionID,
	}, c)
}

// GetAnalysisProgress 获取分析进度 - 统一接口
func (i *IntelligentAnalysisApi) GetAnalysisProgress(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取进度", zap.String("method", "GetAnalysisProgress"))
	
	taskId := c.Param("taskId")
	global.GVA_LOG.Info("获取分析进度", zap.String("task_id", taskId))
	
	// 尝试解析为数字ID（新的分析系统）
	if taskIdInt, err := strconv.ParseUint(taskId, 10, 32); err == nil {
		// 使用真实的分析服务
		task, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.GetAnalysisProgress(uint(taskIdInt))
		if err != nil {
			global.GVA_LOG.Error("获取分析进度失败!", zap.Error(err))
			response.FailWithMessage("获取分析进度失败: "+err.Error(), c)
			return
		}
		
		response.OkWithData(gin.H{
			"task_id":        taskId,
			"status":         task.Status,
			"percentage":     task.Progress,
			"current_step":   getStepDescription(task.Status, task.Progress),
			"detail":         task.Summary,
			"processed_count": task.ProcessedCount,
			"total_count":     task.TotalCount,
			"success_count":   task.SuccessCount,
			"failed_count":    task.FailedCount,
			"start_time":      task.StartTime,
			"end_time":        task.EndTime,
			"error_message":   task.ErrorMsg,
			"logs": generateProgressLogs(task),
		}, c)
		return
	}
	
	// 兼容旧的字符串格式taskId（模拟数据）
	progressData := gin.H{
		"task_id":      taskId,
		"status":       "completed",
		"percentage":   100,
		"current_step": "分析完成",
		"detail":       "成功分析了9个需求项",
		"logs": []map[string]interface{}{
			{
				"timestamp": time.Now().Add(-5*time.Minute).Format("2006-01-02 15:04:05"),
				"message":   "开始分析用户管理模块",
			},
			{
				"timestamp": time.Now().Add(-3*time.Minute).Format("2006-01-02 15:04:05"),
				"message":   "正在进行AI优化分析...",
			},
			{
				"timestamp": time.Now().Add(-1*time.Minute).Format("2006-01-02 15:04:05"),
				"message":   "知识库匹配完成",
			},
			{
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
				"message":   "分析完成，共处理9个需求",
			},
		},
		"processed_count": 9,
		"total_count":     9,
		"start_time":      time.Now().Add(-6*time.Minute).Format("2006-01-02 15:04:05"),
		"end_time":        time.Now().Format("2006-01-02 15:04:05"),
	}
	
	response.OkWithData(progressData, c)
}

// GetAnalysisResult 获取分析结果详情
func (i *IntelligentAnalysisApi) GetAnalysisResult(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取结果", zap.String("method", "GetAnalysisResult"))
	
	taskId := c.Param("taskId")
	global.GVA_LOG.Info("获取分析结果", zap.String("task_id", taskId))
	
	// 尝试解析为数字ID（新的分析系统）
	if taskIdInt, err := strconv.ParseUint(taskId, 10, 32); err == nil {
		// 使用真实的分析服务获取结果
		task, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.GetAnalysisProgress(uint(taskIdInt))
		if err != nil {
			global.GVA_LOG.Error("获取分析任务失败", zap.Error(err))
			response.FailWithMessage("获取分析任务失败: "+err.Error(), c)
			return
		}
		
		if task.Status != "completed" {
			response.FailWithMessage("分析任务尚未完成", c)
			return
		}
		
		// 获取分析结果数据
		resultData, err := i.getAnalysisResultFromTask(task)
		if err != nil {
			global.GVA_LOG.Error("获取分析结果失败", zap.Error(err))
			response.FailWithMessage("获取分析结果失败: "+err.Error(), c)
			return
		}
		
		response.OkWithData(resultData, c)
		return
	}
	
	// 兼容旧的字符串格式taskId（返回模拟数据）
	resultData := gin.H{
		"summary": gin.H{
			"total_requirements": 9,
			"optimized_count":    7,
			"confidence_score":   85,
			"function_points":    42,
			"function_type_distribution": []gin.H{
				{"name": "EI", "value": 3},
				{"name": "EO", "value": 2},
				{"name": "EQ", "value": 2},
				{"name": "ILF", "value": 2},
				{"name": "EIF", "value": 0},
			},
			"complexity_distribution": []gin.H{
				{"name": "Low", "value": 4},
				{"name": "Average", "value": 3},
				{"name": "High", "value": 2},
			},
		},
		"details": []gin.H{
			{
				"id":                   1,
				"title":                "用户管理模块",
				"original_description": "用户管理相关功能",
				"optimized_description": "用户管理模块包含用户注册、登录、权限管理等核心功能，支持用户信息的全生命周期管理",
				"function_type":        "ILF",
				"complexity":           "High",
				"confidence":           90,
				"optimization_level":   "medium",
				"knowledge_references": []gin.H{
					{"id": 1, "title": "NESMA用户管理最佳实践"},
					{"id": 2, "title": "内部逻辑文件复杂度评估标准"},
				},
			},
			{
				"id":                   2,
				"title":                "用户注册功能",
				"original_description": "新用户注册",
				"optimized_description": "用户注册功能包括信息输入验证、重复性检查、数据持久化存储，确保用户信息的完整性和安全性",
				"function_type":        "EI",
				"complexity":           "Average",
				"confidence":           85,
				"optimization_level":   "high",
				"knowledge_references": []gin.H{
					{"id": 3, "title": "外部输入功能标准"},
				},
			},
			{
				"id":                   3,
				"title":                "验证用户信息",
				"original_description": "检查用户输入",
				"optimized_description": "验证用户信息功能负责检查用户输入的合法性、格式正确性以及业务规则符合性",
				"function_type":        "EQ",
				"complexity":           "Low",
				"confidence":           80,
				"optimization_level":   "medium",
				"knowledge_references": []gin.H{
					{"id": 4, "title": "外部查询功能评估"},
				},
			},
		},
	}
	
	response.OkWithData(resultData, c)
}

// GetAnalysisRecommendations 获取分析建议
func (i *IntelligentAnalysisApi) GetAnalysisRecommendations(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取建议", zap.String("method", "GetAnalysisRecommendations"))
	
	taskId := c.Param("taskId")
	global.GVA_LOG.Info("获取分析建议", zap.String("task_id", taskId))
	
	// 模拟分析建议数据
	recommendationsData := gin.H{
		"recommendations": []gin.H{
			{
				"type":        "success",
				"title":       "需求描述优化建议",
				"description": "建议将\"用户管理相关功能\"细化为更具体的功能描述，包括用户注册、登录、权限管理等具体功能点",
				"category":    "需求完善",
				"priority":    "High",
				"effort":      "1-2小时",
				"target_requirements": []int{1, 7},
			},
			{
				"type":        "warning",
				"title":       "功能点计算建议",
				"description": "部分需求的复杂度评估可能偏低，建议重新评估数据元素类型和文件引用数量",
				"category":    "NESMA评估",
				"priority":    "Medium",
				"effort":      "30分钟",
				"target_requirements": []int{3, 6},
			},
			{
				"type":        "info",
				"title":       "知识库扩展建议",
				"description": "建议添加更多产品管理领域的NESMA案例，以提高相关需求的分析准确性",
				"category":    "知识库",
				"priority":    "Low",
				"effort":      "2-4小时",
				"target_requirements": []int{8, 9},
			},
			{
				"type":        "success",
				"title":       "测试用例生成建议",
				"description": "建议为高复杂度的需求生成详细的测试用例，确保功能实现的完整性",
				"category":    "质量保证",
				"priority":    "Medium",
				"effort":      "1-3小时",
				"target_requirements": []int{1, 7},
			},
		},
		"summary": gin.H{
			"total_recommendations": 4,
			"high_priority":         1,
			"medium_priority":       2,
			"low_priority":          1,
			"categories": gin.H{
				"需求完善": 1,
				"NESMA评估": 1,
				"知识库":   1,
				"质量保证": 1,
			},
		},
	}
	
	response.OkWithData(recommendationsData, c)
}

// ApplyAnalysisRecommendation 应用分析建议
func (i *IntelligentAnalysisApi) ApplyAnalysisRecommendation(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 应用建议", zap.String("method", "ApplyAnalysisRecommendation"))
	response.FailWithMessage("功能开发中，敬请期待", c)
}

// GetAIServiceStatus 获取AI服务状态
func (i *IntelligentAnalysisApi) GetAIServiceStatus(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - AI服务状态", zap.String("method", "GetAIServiceStatus"))
	
	// 返回模拟状态
	response.OkWithData(gin.H{
		"status":           "active",
		"available_models": []string{"deepseek-chat", "gpt-4", "claude-3"},
		"current_model":    "deepseek-chat",
		"health_check":     true,
		"last_check":       "2024-01-08T10:00:00Z",
	}, c)
}

// GetAvailableAIModels 获取可用AI模型列表
func (i *IntelligentAnalysisApi) GetAvailableAIModels(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取AI模型", zap.String("method", "GetAvailableAIModels"))
	
	models := []gin.H{
		{
			"id":          "deepseek-chat",
			"name":        "DeepSeek Chat",
			"description": "DeepSeek的对话模型，适用于NESMA需求分析",
			"status":      "active",
			"cost":        "低",
			"speed":       "快",
			"quality":     "高",
		},
		{
			"id":          "gpt-4",
			"name":        "GPT-4",
			"description": "OpenAI的GPT-4模型，具有优秀的分析能力",
			"status":      "active",
			"cost":        "高",
			"speed":       "中",
			"quality":     "很高",
		},
	}

	response.OkWithData(gin.H{
		"models": models,
	}, c)
}

// TestAIModel 测试AI模型连接
func (i *IntelligentAnalysisApi) TestAIModel(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 测试AI模型", zap.String("method", "TestAIModel"))
	response.FailWithMessage("功能开发中，敬请期待", c)
}

// SearchKnowledgeBase 搜索知识库
func (i *IntelligentAnalysisApi) SearchKnowledgeBase(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 搜索知识库", zap.String("method", "SearchKnowledgeBase"))
	response.FailWithMessage("功能开发中，敬请期待", c)
}

// GetKnowledgeGraph 获取知识图谱数据
func (i *IntelligentAnalysisApi) GetKnowledgeGraph(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取知识图谱", zap.String("method", "GetKnowledgeGraph"))
	
	// 模拟知识图谱数据
	graphData := gin.H{
		"nodes": []gin.H{
			{
				"id":       "entity_1",
				"name":     "用户管理",
				"type":     "module", 
				"category": "business",
				"value":    10,
				"x":        100,
				"y":        100,
				"properties": gin.H{
					"complexity": "High",
					"domain":     "user_management",
				},
			},
			{
				"id":       "entity_2", 
				"name":     "数据验证",
				"type":     "function",
				"category": "technical",
				"value":    8,
				"x":        200,
				"y":        150,
				"properties": gin.H{
					"complexity": "Medium",
					"domain":     "validation",
				},
			},
			{
				"id":       "entity_3",
				"name":     "NESMA标准",
				"type":     "standard",
				"category": "knowledge",
				"value":    15,
				"x":        150,
				"y":        50,
				"properties": gin.H{
					"version": "2.1",
					"domain":  "measurement",
				},
			},
		},
		"links": []gin.H{
			{
				"source": "entity_1",
				"target": "entity_2",
				"type":   "includes",
				"weight": 0.8,
				"properties": gin.H{
					"relationship": "composition",
				},
			},
			{
				"source": "entity_1",
				"target": "entity_3", 
				"type":   "follows",
				"weight": 0.9,
				"properties": gin.H{
					"relationship": "compliance",
				},
			},
		},
		"categories": []gin.H{
			{"name": "business", "color": "#ff7f0e"},
			{"name": "technical", "color": "#2ca02c"},
			{"name": "knowledge", "color": "#1f77b4"},
		},
	}
	
	response.OkWithData(graphData, c)
}

// GetKnowledgeEntities 获取知识实体列表
func (i *IntelligentAnalysisApi) GetKnowledgeEntities(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取知识实体", zap.String("method", "GetKnowledgeEntities"))
	
	// 模拟实体数据
	entities := []gin.H{
		{
			"id":          "entity_1",
			"name":        "用户管理",
			"type":        "module",
			"category":    "business",
			"description": "用户管理模块，包含用户注册、登录、权限等功能",
			"properties": gin.H{
				"complexity":    "High",
				"domain":        "user_management",
				"function_type": "ILF",
			},
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			"id":          "entity_2",
			"name":        "数据验证",
			"type":        "function",
			"category":    "technical", 
			"description": "数据输入验证功能，确保数据完整性和合法性",
			"properties": gin.H{
				"complexity":    "Medium",
				"domain":        "validation",
				"function_type": "EQ",
			},
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			"id":          "entity_3",
			"name":        "NESMA标准",
			"type":        "standard",
			"category":    "knowledge",
			"description": "NESMA功能点分析标准和最佳实践",
			"properties": gin.H{
				"version": "2.1",
				"domain":  "measurement",
			},
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
			"updated_at": time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	
	response.OkWithData(gin.H{
		"entities": entities,
		"total":    len(entities),
	}, c)
}

// GetKnowledgeRelations 获取知识关系列表
func (i *IntelligentAnalysisApi) GetKnowledgeRelations(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 获取知识关系", zap.String("method", "GetKnowledgeRelations"))
	
	// 模拟关系数据
	relations := []gin.H{
		{
			"id":          "relation_1",
			"source":      "entity_1",
			"target":      "entity_2", 
			"type":        "includes",
			"description": "用户管理模块包含数据验证功能",
			"weight":      0.8,
			"properties": gin.H{
				"relationship": "composition",
				"strength":     "strong",
			},
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
		{
			"id":          "relation_2",
			"source":      "entity_1",
			"target":      "entity_3",
			"type":        "follows",
			"description": "用户管理模块遵循NESMA标准",
			"weight":      0.9,
			"properties": gin.H{
				"relationship": "compliance",
				"strength":     "strong",
			},
			"created_at": time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	
	response.OkWithData(gin.H{
		"relations": relations,
		"total":     len(relations),
	}, c)
}

// VectorSearch 向量搜索
func (i *IntelligentAnalysisApi) VectorSearch(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 向量搜索", zap.String("method", "VectorSearch"))
	response.FailWithMessage("功能开发中，敬请期待", c)
}

// ExportIntelligentAnalysisReport 导出智能分析报告
func (i *IntelligentAnalysisApi) ExportIntelligentAnalysisReport(c *gin.Context) {
	global.GVA_LOG.Info("智能分析 - 导出报告", zap.String("method", "ExportIntelligentAnalysisReport"))
	response.FailWithMessage("功能开发中，敬请期待", c)
}

// 统一接口的辅助函数

// convertToAnalysisRequest 转换为分析请求格式
func convertToAnalysisRequest(req interface{}) *nesmaReq.NesmaAnalyzeRequest {
	// 将统一接口的请求转换为MVP分析服务所需的格式
	switch r := req.(type) {
	case map[string]interface{}:
		// 处理通用map格式
		projectID, _ := r["project_id"].(uint)
		cycleID, _ := r["cycle_id"].(uint)
		
		return &nesmaReq.NesmaAnalyzeRequest{
			ProjectID: projectID,
			CycleID:   cycleID,
			Config: struct {
				AnalysisType    string   `json:"analysisType"`
				TargetLevels    []int    `json:"targetLevels"`
				AIModel         string   `json:"aiModel"`
				EnableMermaid   bool     `json:"enableMermaid"`
				BatchSize       int      `json:"batchSize"`
				MaxRetries      int      `json:"maxRetries"`
				SkipCompleted   bool     `json:"skipCompleted"`
			}{
				AnalysisType:  "requirement_analysis",
				TargetLevels:  []int{3, 4},
				AIModel:       "deepseek",
				EnableMermaid: true,
				BatchSize:     10,
				MaxRetries:    3,
				SkipCompleted: true,
			},
		}
	default:
		// 处理结构体格式
		if reqStruct, ok := req.(struct {
			ProjectID      uint     `json:"project_id"`
			CycleID        uint     `json:"cycle_id"`
			RequirementIDs []uint   `json:"requirement_ids"`
			Config         struct {
				Mode               string   `json:"mode"`
				AIModel            string   `json:"ai_model"`
				KnowledgeWeight    int      `json:"knowledge_weight"`
				GenerateContent    []string `json:"generate_content"`
				AnalysisDepth      string   `json:"analysis_depth"`
				ExecutionMode      string   `json:"execution_mode"`
				ConfidenceThreshold float64 `json:"confidence_threshold"`
			} `json:"config"`
		}); ok {
			
			aiModel := reqStruct.Config.AIModel
			if aiModel == "" {
				aiModel = "deepseek"
			}
			
			return &nesmaReq.NesmaAnalyzeRequest{
				ProjectID:      reqStruct.ProjectID,
				CycleID:        reqStruct.CycleID,
				RequirementIDs: reqStruct.RequirementIDs, // 传递需求ID列表
				Config: struct {
					AnalysisType    string   `json:"analysisType"`
					TargetLevels    []int    `json:"targetLevels"`
					AIModel         string   `json:"aiModel"`
					EnableMermaid   bool     `json:"enableMermaid"`
					BatchSize       int      `json:"batchSize"`
					MaxRetries      int      `json:"maxRetries"`
					SkipCompleted   bool     `json:"skipCompleted"`
				}{
					AnalysisType:  "requirement_analysis",
					TargetLevels:  []int{3, 4},
					AIModel:       aiModel,
					EnableMermaid: true,
					BatchSize:     10,
					MaxRetries:    3,
					SkipCompleted: true,
				},
			}
		}
	}
	
	// 默认返回基本配置
	return &nesmaReq.NesmaAnalyzeRequest{
		ProjectID: 1,
		CycleID:   1,
		Config: struct {
			AnalysisType    string   `json:"analysisType"`
			TargetLevels    []int    `json:"targetLevels"`
			AIModel         string   `json:"aiModel"`
			EnableMermaid   bool     `json:"enableMermaid"`
			BatchSize       int      `json:"batchSize"`
			MaxRetries      int      `json:"maxRetries"`
			SkipCompleted   bool     `json:"skipCompleted"`
		}{
			AnalysisType:  "requirement_analysis",
			TargetLevels:  []int{3, 4},
			AIModel:       "deepseek",
			EnableMermaid: true,
			BatchSize:     10,
			MaxRetries:    3,
			SkipCompleted: true,
		},
	}
}

// getStepDescription 根据状态和进度获取步骤描述
func getStepDescription(status string, progress int) string {
	switch status {
	case "pending":
		return "等待开始分析..."
	case "running":
		if progress < 20 {
			return "正在加载需求数据..."
		} else if progress < 40 {
			return "正在进行AI分析..."
		} else if progress < 60 {
			return "正在搜索知识库..."
		} else if progress < 80 {
			return "正在生成优化建议..."
		} else {
			return "正在完成分析..."
		}
	case "completed":
		return "分析完成"
	case "failed":
		return "分析失败"
	default:
		return "未知状态"
	}
}

// generateProgressLogs 生成进度日志
func generateProgressLogs(task interface{}) []map[string]interface{} {
	// 这里可以根据task的实际结构生成更详细的日志
	// 暂时返回一个通用的日志格式
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-5*time.Minute).Format("2006-01-02 15:04:05"),
			"message":   "开始智能分析任务",
		},
		{
			"timestamp": time.Now().Add(-3*time.Minute).Format("2006-01-02 15:04:05"),
			"message":   "正在进行AI驱动的需求分析...",
		},
		{
			"timestamp": time.Now().Add(-1*time.Minute).Format("2006-01-02 15:04:05"),
			"message":   "知识库匹配和优化建议生成中...",
		},
		{
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			"message":   "分析任务完成",
		},
	}
	return logs
}

// buildRequirementTree 构建需求树形结构
func buildRequirementTree(requirements []nesma.NesmaRequirement) []map[string]interface{} {
	var parentMap = make(map[uint][]map[string]interface{})
	var topLevel []map[string]interface{}
	
	// 先处理所有需求，转换为通用格式
	for _, req := range requirements {
		item := convertRequirementToMap(req)
		if item == nil {
			continue
		}
		
		// 按父级分组
		if parentId, exists := item["parent_id"]; exists && parentId != nil {
			if pid, ok := parentId.(*uint); ok && pid != nil && *pid > 0 {
				parentMap[*pid] = append(parentMap[*pid], item)
			} else {
				topLevel = append(topLevel, item)
			}
		} else {
			topLevel = append(topLevel, item)
		}
	}
	
	// 递归构建树形结构
	var buildChildren func(parentId uint) []map[string]interface{}
	buildChildren = func(parentId uint) []map[string]interface{} {
		children, exists := parentMap[parentId]
		if !exists {
			return []map[string]interface{}{}
		}
		
		for i := range children {
			if childId, exists := children[i]["id"]; exists {
				if cid, ok := childId.(uint); ok {
					children[i]["children"] = buildChildren(cid)
				}
			}
		}
		return children
	}
	
	// 为顶级需求添加子需求
	for i := range topLevel {
		if reqId, exists := topLevel[i]["id"]; exists {
			if rid, ok := reqId.(uint); ok {
				topLevel[i]["children"] = buildChildren(rid)
			}
		}
	}
	
	return topLevel
}

// buildFlatRequirementList 构建需求扁平化列表
func buildFlatRequirementList(requirements []nesma.NesmaRequirement) []map[string]interface{} {
	var flatList []map[string]interface{}
	
	for _, req := range requirements {
		item := convertRequirementToMap(req)
		if item != nil {
			// 只保留需要的字段
			flatItem := map[string]interface{}{
				"id":            item["id"],
				"title":         item["title"],
				"level":         item["level"],
				"complexity":    item["complexity"],
				"function_type": item["function_type"],
				"parent_id":     item["parent_id"],
				"status":        item["status"],
				"ai_analysis_status": item["ai_analysis_status"],
			}
			flatList = append(flatList, flatItem)
		}
	}
	
	return flatList
}

// convertRequirementToMap 将需求结构体转换为map
func convertRequirementToMap(req interface{}) map[string]interface{} {
	// 根据实际的需求模型进行转换
	switch r := req.(type) {
	case nesma.NesmaRequirement:
		return map[string]interface{}{
			"id":                  r.ID,
			"title":               r.Title,
			"description":         r.Description,
			"level":               r.Level,
			"complexity":          r.Complexity,
			"function_type":       r.FunctionType,
			"parent_id":           r.ParentID,
			"status":              r.Status,
			"ai_analysis_status":  r.AIAnalysisStatus,
			"project_id":          r.ProjectID,
			"cycle_id":            r.CycleID,
			"version_id":          r.VersionID,
			"priority":            r.Priority,
			"order_index":         r.OrderIndex,
			"afp":                 r.AFP,
			"ufp":                 r.UFP,
			"reuse_level":         r.ReuseLevel,
			"modification_type":   r.ModificationType,
			"ai_description":      r.AIDescription,
			"ai_generated_title":  r.AIGeneratedTitle,
			"ai_complexity_score": r.AIComplexityScore,
			"ai_confidence_score": r.AIConfidenceScore,
			"created_at":          r.CreatedAt,
			"updated_at":          r.UpdatedAt,
		}
	case *nesma.NesmaRequirement:
		if r == nil {
			return nil
		}
		return map[string]interface{}{
			"id":                  r.ID,
			"title":               r.Title,
			"description":         r.Description,
			"level":               r.Level,
			"complexity":          r.Complexity,
			"function_type":       r.FunctionType,
			"parent_id":           r.ParentID,
			"status":              r.Status,
			"ai_analysis_status":  r.AIAnalysisStatus,
			"project_id":          r.ProjectID,
			"cycle_id":            r.CycleID,
			"version_id":          r.VersionID,
			"priority":            r.Priority,
			"order_index":         r.OrderIndex,
			"afp":                 r.AFP,
			"ufp":                 r.UFP,
			"reuse_level":         r.ReuseLevel,
			"modification_type":   r.ModificationType,
			"ai_description":      r.AIDescription,
			"ai_generated_title":  r.AIGeneratedTitle,
			"ai_complexity_score": r.AIComplexityScore,
			"ai_confidence_score": r.AIConfidenceScore,
			"created_at":          r.CreatedAt,
			"updated_at":          r.UpdatedAt,
		}
	default:
		// 对于其他类型，返回nil
		return nil
	}
}

// getAnalysisResultFromTask 从分析任务获取分析结果
func (i *IntelligentAnalysisApi) getAnalysisResultFromTask(task interface{}) (gin.H, error) {
	// 检查task的类型
	var analysisTask *nesma.NesmaRequirementAnalysisTask
	
	// 尝试从不同类型中获取任务信息
	switch t := task.(type) {
	case *nesma.NesmaRequirementAnalysisTask:
		analysisTask = t
	case nesma.NesmaRequirementAnalysisTask:
		analysisTask = &t
	default:
		global.GVA_LOG.Error("任务数据类型不匹配", zap.Any("taskType", fmt.Sprintf("%T", task)))
		return nil, fmt.Errorf("任务数据类型不匹配")
	}

	// 验证必要的字段
	if analysisTask.ProjectID == 0 || analysisTask.CycleID == 0 {
		global.GVA_LOG.Error("任务缺少必要信息", 
			zap.Uint("projectID", analysisTask.ProjectID),
			zap.Uint("cycleID", analysisTask.CycleID))
		return nil, fmt.Errorf("任务缺少必要信息")
	}

	// 获取目标版本ID
	var targetVersionID uint
	if analysisTask.TargetVersionID != nil {
		targetVersionID = *analysisTask.TargetVersionID
	} else {
		// 如果没有目标版本ID，使用源版本ID
		targetVersionID = analysisTask.SourceVersionID
	}

	// 获取需求服务
	requirementService := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService
	
	// 构建查询条件获取分析后的需求
	queryConditions := map[string]interface{}{
		"project_id": analysisTask.ProjectID,
		"cycle_id":   analysisTask.CycleID,
		"version_id": targetVersionID,
	}
	
	requirements, total, err := requirementService.GetRequirementTreeByProject(analysisTask.ProjectID, queryConditions)
	if err != nil {
		global.GVA_LOG.Error("获取需求数据失败", zap.Error(err))
		return nil, fmt.Errorf("获取需求数据失败: %v", err)
	}

	// 统计功能点类型分布
	typeDistribution := make(map[string]int)
	complexityDistribution := make(map[string]int)
	var totalFunctionPoints float64
	var optimizedCount int
	var totalConfidence float64
	var confidenceCount int
	
	// 构建详细的需求列表
	var details []gin.H
	
	for _, req := range requirements {
		// 只处理四级需求（功能点）
		if req.Level == 4 {
			// 统计功能点类型
			if req.FunctionType != "" {
				typeDistribution[req.FunctionType]++
				totalFunctionPoints += req.AFP
			}
			
			// 统计复杂度分布
			if req.Complexity != "" {
				complexityDistribution[req.Complexity]++
			}
			
			// 统计AI优化情况
			if req.AIDescription != "" || req.AIGeneratedTitle != "" {
				optimizedCount++
			}
			
			// 统计置信度
			if req.AIConfidenceScore != nil && *req.AIConfidenceScore > 0 {
				totalConfidence += *req.AIConfidenceScore
				confidenceCount++
			}
			
			// 构建详细信息
			detail := gin.H{
				"id":                   req.ID,
				"title":                req.Title,
				"original_description": req.Description,
				"optimized_description": req.AIDescription,
				"function_type":        req.FunctionType,
				"complexity":           req.Complexity,
				"confidence":           i.getConfidenceScore(req.AIConfidenceScore),
				"afp":                  req.AFP,
				"ufp":                  req.UFP,
				"optimization_level":   i.getOptimizationLevel(req.AIConfidenceScore),
				"knowledge_references": i.buildKnowledgeReferences(req),
			}
			details = append(details, detail)
		}
	}
	
	// 计算平均置信度
	avgConfidence := 0.0
	if confidenceCount > 0 {
		avgConfidence = totalConfidence / float64(confidenceCount)
	}
	
	// 如果从数据库没有获取到足够的数据，使用任务统计信息
	if total == 0 && analysisTask.TotalCount > 0 {
		total = int64(analysisTask.TotalCount)
	}
	if optimizedCount == 0 && analysisTask.SuccessCount > 0 {
		optimizedCount = analysisTask.SuccessCount
	}
	if avgConfidence == 0 && analysisTask.QualityScore != nil {
		avgConfidence = *analysisTask.QualityScore / 100 // 转换为0-1范围
	}
	
	// 构建功能点类型分布数组
	var functionTypeDistribution []gin.H
	functionTypes := []string{"EI", "EO", "EQ", "ILF", "EIF"}
	for _, funcType := range functionTypes {
		functionTypeDistribution = append(functionTypeDistribution, gin.H{
			"name":  funcType,
			"value": typeDistribution[funcType],
		})
	}
	
	// 构建复杂度分布数组
	var complexityDistributionArray []gin.H
	complexityLevels := []string{"Low", "Average", "High"}
	for _, level := range complexityLevels {
		complexityDistributionArray = append(complexityDistributionArray, gin.H{
			"name":  level,
			"value": complexityDistribution[level],
		})
	}
	
	// 构建分析结果
	resultData := gin.H{
		"summary": gin.H{
			"total_requirements":         total,
			"optimized_count":            optimizedCount,
			"confidence_score":           avgConfidence,
			"function_points":            totalFunctionPoints,
			"function_type_distribution": functionTypeDistribution,
			"complexity_distribution":    complexityDistributionArray,
		},
		"details": details,
		"analysis_metadata": gin.H{
			"task_id":           analysisTask.ID,
			"project_id":        analysisTask.ProjectID,
			"cycle_id":          analysisTask.CycleID,
			"target_version_id": targetVersionID,
			"analysis_time":     analysisTask.CreatedAt,
			"task_status":       analysisTask.Status,
			"task_progress":     analysisTask.Progress,
			"success_rate":      analysisTask.GetSuccessRate(),
		},
	}
	
	return resultData, nil
}

// getOptimizationLevel 根据置信度分数获取优化级别
func (i *IntelligentAnalysisApi) getOptimizationLevel(confidenceScore *float64) string {
	if confidenceScore == nil {
		return "unknown"
	}
	if *confidenceScore >= 0.8 {
		return "high"
	} else if *confidenceScore >= 0.6 {
		return "medium"
	} else {
		return "low"
	}
}

// getConfidenceScore 获取置信度分数
func (i *IntelligentAnalysisApi) getConfidenceScore(confidenceScore *float64) float64 {
	if confidenceScore == nil {
		return 0.0
	}
	return *confidenceScore
}

// buildKnowledgeReferences 构建知识库引用信息
func (i *IntelligentAnalysisApi) buildKnowledgeReferences(req nesma.NesmaRequirement) []gin.H {
	// 这里可以根据需求的AI分析结果或其他字段构建知识库引用
	// 当前返回示例数据，后续可以集成真实的知识库服务
	var references []gin.H
	
	if req.FunctionType != "" {
		references = append(references, gin.H{
			"id":    1,
			"title": fmt.Sprintf("NESMA %s 功能点标准", req.FunctionType),
		})
	}
	
	if req.Complexity != "" {
		references = append(references, gin.H{
			"id":    2,
			"title": fmt.Sprintf("%s复杂度评估指南", req.Complexity),
		})
	}
	
	return references
}
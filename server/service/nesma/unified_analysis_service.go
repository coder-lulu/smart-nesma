package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"slices"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"go.uber.org/zap"
)

// AnalysisConfig 分析配置
type AnalysisConfig struct {
	AnalysisScope       string   `json:"analysis_scope"`        // 分析范围: all, specific_levels, incremental
	IncludeLevels      []int    `json:"include_levels"`        // 包含的级别: [3,4] 或 [3] 或 [4]
	OptimizeTitle      bool     `json:"optimize_title"`        // 是否优化标题
	OptimizeDescription bool    `json:"optimize_description"`  // 是否优化描述
	GenerateMermaid    bool     `json:"generate_mermaid"`      // 是否生成Mermaid流程图
	AIModel           string   `json:"ai_model"`              // AI模型选择: deepseek-chat, deepseek-reasoner
	MaxRetries        int      `json:"max_retries"`           // 最大重试次数
	BatchSize         int      `json:"batch_size"`            // 批处理大小
	MaxConcurrency    int      `json:"max_concurrency"`       // 最大并发数
}

// DefaultAnalysisConfig 默认分析配置
var DefaultAnalysisConfig = AnalysisConfig{
	AnalysisScope:       "all",
	IncludeLevels:      []int{3, 4},
	OptimizeTitle:      true,
	OptimizeDescription: true,
	GenerateMermaid:    false,
	AIModel:           "deepseek-chat",
	MaxRetries:        3,
	BatchSize:         10,
	MaxConcurrency:    5, // 简单的并发控制
}

// UnifiedAnalysisService 统一分析服务 - 简化版本
type UnifiedAnalysisService struct {
	analysisService *RequirementAnalysisService
}

// NewUnifiedAnalysisService 创建统一分析服务
func NewUnifiedAnalysisService() *UnifiedAnalysisService {
	return &UnifiedAnalysisService{
		analysisService: &RequirementAnalysisService{},
	}
}

// ProjectAnalysisRequest 项目分析请求
type ProjectAnalysisRequest struct {
	ProjectID uint `json:"project_id" binding:"required"`
	CycleID   uint `json:"cycle_id" binding:"required"`
}

// ProjectAnalysisResponse 项目分析响应
type ProjectAnalysisResponse struct {
	TaskID    uint   `json:"task_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	ProjectID uint   `json:"project_id"`
	CycleID   uint   `json:"cycle_id"`
}

// AnalysisProgressResponse 分析进度响应
type AnalysisProgressResponse struct {
	TaskID         uint    `json:"task_id"`
	Status         string  `json:"status"`         // pending/running/completed/failed
	Progress       int     `json:"progress"`       // 0-100
	ProcessedCount int     `json:"processed_count"`
	TotalCount     int     `json:"total_count"`
	Message        string  `json:"message"`
	ErrorMsg       string  `json:"error_msg,omitempty"`
}

// ExecuteProjectAnalysis 执行项目分析 - 简化版本
func (s *UnifiedAnalysisService) ExecuteProjectAnalysis(ctx context.Context, req *nesmaReq.UnifiedAnalysisRequest) (*ProjectAnalysisResponse, error) {
	// 处理可选的CycleID
	var cycleID uint
	if req.CycleID != nil {
		cycleID = *req.CycleID
	} else {
		return nil, fmt.Errorf("cycle_id是必需的")
	}

	// 1. 解析用户参数
	analysisConfig := s.parseAnalysisConfig(req)
	configJson, _ := json.Marshal(analysisConfig)

	global.GVA_LOG.Info("开始执行项目分析", 
		zap.Uint("projectId", req.ProjectID), 
		zap.Uint("cycleId", cycleID),
		zap.String("analysisScope", analysisConfig.AnalysisScope),
		zap.Any("includeLevels", analysisConfig.IncludeLevels))

	// 2. 获取当前活跃周期的最新版本
	sourceVersion, err := s.getActiveVersionForCycle(cycleID)
	if err != nil {
		return nil, fmt.Errorf("获取活跃版本失败: %w", err)
	}

	// 3. 创建分析任务
	task := &nesma.NesmaRequirementAnalysisTask{
		ProjectID:       req.ProjectID,
		CycleID:         cycleID,
		TaskType:        "project_analysis",
		Status:          "pending",
		Progress:        0,
		SourceVersionID: sourceVersion.ID,
		Config:          configJson, // 保存分析配置
	}

	if err := global.GVA_DB.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建分析任务失败: %w", err)
	}

	// 4. 异步执行分析
	go s.executeAnalysisAsync(task)

	return &ProjectAnalysisResponse{
		TaskID:    task.ID,
		Status:    "pending",
		Message:   "分析任务已创建，正在后台处理",
		ProjectID: req.ProjectID,
		CycleID:   cycleID,
	}, nil
}

// executeAnalysisAsync 异步执行分析 - 简化版本
func (s *UnifiedAnalysisService) executeAnalysisAsync(task *nesma.NesmaRequirementAnalysisTask) {
	// 更新任务状态为运行中
	task.Status = "running"
	task.Progress = 0
	now := time.Now()
	task.StartTime = &now
	global.GVA_DB.Save(task)

	global.GVA_LOG.Info("开始后台分析任务", zap.Uint("taskId", task.ID))

	// 执行分析
	err := s.performSimpleAnalysis(task)
	
	// 更新任务完成状态
	endTime := time.Now()
	task.EndTime = &endTime
	
	if err != nil {
		task.Status = "failed"
		task.ErrorMsg = err.Error()
		global.GVA_LOG.Error("分析任务失败", zap.Uint("taskId", task.ID), zap.Error(err))
	} else {
		task.Status = "completed"
		task.Progress = 100
		global.GVA_LOG.Info("分析任务完成", zap.Uint("taskId", task.ID))
	}
	
	global.GVA_DB.Save(task)
}

// performSimpleAnalysis 执行简单分析 - 使用简单多线程
func (s *UnifiedAnalysisService) performSimpleAnalysis(task *nesma.NesmaRequirementAnalysisTask) error {
	// 1. 解析分析配置
	var analysisConfig AnalysisConfig
	if task.Config != nil {
		err := json.Unmarshal(task.Config, &analysisConfig)
	if err != nil {
			global.GVA_LOG.Warn("解析分析配置失败，使用默认配置", zap.Error(err))
			analysisConfig = DefaultAnalysisConfig
		}
	} else {
		analysisConfig = DefaultAnalysisConfig
	}

	global.GVA_LOG.Info("使用分析配置", 
		zap.Uint("taskId", task.ID),
		zap.String("analysisScope", analysisConfig.AnalysisScope),
		zap.Any("includeLevels", analysisConfig.IncludeLevels),
		zap.String("aiModel", analysisConfig.AIModel))

	// 2. 创建新版本用于存储分析结果（会自动复制源版本的需求）
	newVersion, err := s.createAnalysisVersion(task.CycleID, task.SourceVersionID)
		if err != nil {
		return fmt.Errorf("创建分析版本失败: %w", err)
	}

	task.TargetVersionID = &newVersion.ID
	global.GVA_DB.Save(task)

	// 3. 根据配置获取需要分析的需求
	requirements, err := s.getRequirementsForAnalysis(newVersion.ID, analysisConfig)
		if err != nil {
		return fmt.Errorf("获取分析需求失败: %w", err)
	}

	if len(requirements) == 0 {
		return fmt.Errorf("根据配置没有找到需要分析的需求")
	}

	task.TotalCount = len(requirements)
	global.GVA_DB.Save(task)

	global.GVA_LOG.Info("开始分析需求", 
		zap.Uint("taskId", task.ID),
		zap.Int("totalCount", len(requirements)),
		zap.Uint("newVersionId", newVersion.ID))

	// 4. 使用简单的多线程处理需求
	err = s.processRequirementsWithSimpleThreads(requirements, task, analysisConfig)
	if err != nil {
		return fmt.Errorf("处理需求失败: %w", err)
	}

	global.GVA_LOG.Info("简单分析完成", 
		zap.Uint("taskId", task.ID),
		zap.Int("successCount", task.SuccessCount),
		zap.Int("failedCount", task.FailedCount))

	return nil
}

// processRequirementsWithSimpleThreads 使用简单多线程处理需求
func (s *UnifiedAnalysisService) processRequirementsWithSimpleThreads(requirements []nesma.NesmaRequirement, task *nesma.NesmaRequirementAnalysisTask, config AnalysisConfig) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	// 使用channel控制并发数量
	semaphore := make(chan struct{}, config.MaxConcurrency)
	
	// 初始化统计变量
	processedCount := 0
	successCount := 0
	failedCount := 0
	
	// 计算总任务数
	totalTasks := 0
	if config.OptimizeDescription {
		totalTasks += len(requirements)
	}
	
	// 计算L4生成任务数
	l4TaskCount := 0
	if slices.Contains(config.IncludeLevels, 4) {
	for _, req := range requirements {
			if req.Level == 3 && slices.Contains(config.IncludeLevels, 4) {
				l4TaskCount++
			}
		}
		totalTasks += l4TaskCount
	}
	
	// 计算Mermaid生成任务数
	mermaidTaskCount := 0
	if config.GenerateMermaid {
		var l4Requirements []nesma.NesmaRequirement
		err := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
			Where("version_id = ? AND level = 4", *task.TargetVersionID).
			Find(&l4Requirements).Error
		if err != nil {
			return fmt.Errorf("获取L4需求失败: %w", err)
		}
		mermaidTaskCount = len(l4Requirements)
		totalTasks += mermaidTaskCount
	}
	
	// 设置总任务数
	task.TotalCount = totalTasks
	global.GVA_DB.Save(task)
	
	global.GVA_LOG.Info("任务统计初始化", 
		zap.Uint("taskId", task.ID), 
		zap.Int("totalTasks", totalTasks),
		zap.Int("optimizeTasks", len(requirements)),
		zap.Int("l4Tasks", l4TaskCount),
		zap.Int("mermaidTasks", mermaidTaskCount))

	// 启动进度更新协程
	progressTicker := time.NewTicker(3 * time.Second)
	defer progressTicker.Stop()
	
	go func() {
		for range progressTicker.C {
			mu.Lock()
			currentProcessed := processedCount
			currentSuccess := successCount
			currentFailed := failedCount
			mu.Unlock()
			
			if totalTasks > 0 {
				progress := int(float64(currentProcessed) / float64(totalTasks) * 100)
				if progress > 100 {
					progress = 100
				}
				
				task.Progress = progress
				task.ProcessedCount = currentProcessed
				task.SuccessCount = currentSuccess
				task.FailedCount = currentFailed
	global.GVA_DB.Save(task)

				global.GVA_LOG.Info("分析进度更新", 
					zap.Uint("taskId", task.ID),
					zap.Int("processed", currentProcessed),
					zap.Int("total", totalTasks),
					zap.Int("progress", progress),
					zap.Int("success", currentSuccess),
					zap.Int("failed", currentFailed))
			}
		}
	}()

	// 处理需求优化
	if config.OptimizeDescription {
		global.GVA_LOG.Info("开始需求优化阶段", 
			zap.Uint("taskId", task.ID),
			zap.Int("requirementCount", len(requirements)))
		
		for i, req := range requirements {
			wg.Add(1)
			
			go func(index int, requirement nesma.NesmaRequirement) {
	defer wg.Done()
	
				// 获取信号量
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				
				global.GVA_LOG.Info("开始分析需求", 
					zap.Uint("requirementId", requirement.ID), 
					zap.String("title", requirement.Title))
				
				// 分析单个需求
				err := s.analyzeSimpleRequirement(&requirement, *task.TargetVersionID, config)
				
				// 更新统计
				mu.Lock()
				processedCount++
				if err != nil {
					failedCount++
					global.GVA_LOG.Warn("需求分析失败", 
						zap.Uint("requirementId", requirement.ID), 
						zap.String("title", requirement.Title),
				zap.Error(err))
		} else {
					successCount++
					global.GVA_LOG.Info("需求分析成功", 
						zap.Uint("requirementId", requirement.ID), 
						zap.String("title", requirement.Title))
				}
				mu.Unlock()
				
			}(i, req)
		}
		
		// 等待需求优化完成
		wg.Wait()
		
		global.GVA_LOG.Info("需求优化阶段完成", 
		zap.Uint("taskId", task.ID),
			zap.Int("processed", processedCount),
			zap.Int("success", successCount),
			zap.Int("failed", failedCount))
	}

	// 生成L4需求
	if slices.Contains(config.IncludeLevels, 4) {
		global.GVA_LOG.Info("开始L4生成阶段", 
			zap.Uint("taskId", task.ID),
			zap.Int("l4TaskCount", l4TaskCount))
		
		for _, requirement := range requirements {	
			if requirement.Level == 3 && slices.Contains(config.IncludeLevels, 4) {
		wg.Add(1)
				go func(requirement nesma.NesmaRequirement) {
	defer wg.Done()
	
					level4GeneratorService := GetLevel4GeneratorService()
					if level4GeneratorService == nil {
						global.GVA_LOG.Error("无法获取Level4GeneratorService实例", 
							zap.Uint("requirementId", requirement.ID))
						
						mu.Lock()
						processedCount++
						failedCount++
						mu.Unlock()
						return
					}
					
					var cycle nesma.NesmaProjectCycle
					err := global.GVA_DB.First(&cycle, *requirement.CycleID).Error
					if err != nil {
						global.GVA_LOG.Error("获取Cycle失败", 
							zap.Uint("cycleId", *requirement.CycleID), 
				zap.Error(err))
						
						mu.Lock()
						processedCount++
						failedCount++
						mu.Unlock()
						return
					}
					
					result, err := level4GeneratorService.generateLevel4ForRequirement(&requirement, &cycle)
	if err != nil {
						global.GVA_LOG.Error("生成L4需求失败", 
							zap.Uint("requirementId", requirement.ID),
							zap.Error(err))
						
						mu.Lock()
						processedCount++
						failedCount++
						mu.Unlock()
						return
					}
					
					_, err = level4GeneratorService.saveLevel4RequirementsToDatabase(result, *task.TargetVersionID)
	if err != nil {
						global.GVA_LOG.Error("保存L4需求到数据库失败", 
							zap.Uint("requirementId", requirement.ID),
							zap.Error(err))
						
						mu.Lock()
						processedCount++
						failedCount++
						mu.Unlock()
						return
					}
					
					mu.Lock()
					processedCount++
					successCount++
					mu.Unlock()
					
					global.GVA_LOG.Info("L4需求生成成功", 
						zap.Uint("requirementId", requirement.ID))
					
				}(requirement)
			}
		}
		
		// 等待L4生成完成
		wg.Wait()
		
		global.GVA_LOG.Info("L4生成阶段完成", 
			zap.Uint("taskId", task.ID),
			zap.Int("processed", processedCount),
			zap.Int("success", successCount),
			zap.Int("failed", failedCount))
	}

	// 生成Mermaid流程图
	if config.GenerateMermaid {
		global.GVA_LOG.Info("开始Mermaid生成阶段", 
			zap.Uint("taskId", task.ID),
			zap.Int("mermaidTaskCount", mermaidTaskCount))
		
		var l4Requirements []nesma.NesmaRequirement
		err := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
			Where("version_id = ? AND level = 4", *task.TargetVersionID).
			Find(&l4Requirements).Error
		if err != nil {
			return fmt.Errorf("获取L4需求失败: %w", err)
		}
		
		for _, l4Requirement := range l4Requirements {
			wg.Add(1)
			go func(l4Requirement nesma.NesmaRequirement) {
				defer wg.Done()
				
				err := s.generateMermaidForRequirement([]nesma.NesmaRequirement{l4Requirement}, task)
				
				mu.Lock()
				processedCount++
				if err != nil {
			failedCount++
					global.GVA_LOG.Error("生成流程图失败", 
						zap.Uint("requirementId", l4Requirement.ID),
						zap.Error(err))
		} else {
			successCount++
					global.GVA_LOG.Info("流程图生成成功", 
						zap.Uint("requirementId", l4Requirement.ID))
				}
				mu.Unlock()
				
			}(l4Requirement)
		}
		
		// 等待Mermaid生成完成
		wg.Wait()
		
		global.GVA_LOG.Info("Mermaid生成阶段完成", 
			zap.Uint("taskId", task.ID),
			zap.Int("processed", processedCount),
			zap.Int("success", successCount),
			zap.Int("failed", failedCount))
	}

	// 最终更新任务状态
	task.ProcessedCount = processedCount
	task.SuccessCount = successCount
	task.FailedCount = failedCount
	task.Progress = 100
	task.Summary = fmt.Sprintf("分析完成，总任务数: %d，成功: %d，失败: %d", totalTasks, successCount, failedCount)
	global.GVA_DB.Save(task)
	
	global.GVA_LOG.Info("所有任务完成", 
		zap.Uint("taskId", task.ID),
		zap.Int("totalTasks", totalTasks),
		zap.Int("processed", processedCount),
		zap.Int("success", successCount),
		zap.Int("failed", failedCount))

	return nil
}

// generateMermaidForRequirement 生成流程图
func (s *UnifiedAnalysisService) generateMermaidForRequirement(requirements []nesma.NesmaRequirement, task *nesma.NesmaRequirementAnalysisTask) error {
	generatorService := GetMermaidGeneratorService()
	if generatorService == nil {
		global.GVA_LOG.Error("无法获取MermaidGeneratorService实例")
		return fmt.Errorf("无法获取MermaidGeneratorService实例")
	}
	requirement := requirements[0]
	// 为单个需求生成流程图
	genResults, err := generatorService.GenerateMermaidDiagrams(
		task.CycleID,
		[]uint{requirement.ID},
		"flowchart",
	)
	if err != nil {
		global.GVA_LOG.Error("生成流程图失败", 
			zap.Uint("requirementId", requirement.ID),
			zap.Error(err),
		)
	}
	// 检查是否有生成结果，避免数组越界
	if len(genResults) > 0 && genResults[0].GeneratedMermaid.MermaidCode != "" {
		// 将生成的流程图直接更新到需求的MermaidDiagram字段
		now := time.Now()
		requirement.MermaidDiagram = genResults[0].GeneratedMermaid.MermaidCode
		requirement.MermaidGeneratedAt = &now
		if err := global.GVA_DB.Save(&requirement).Error; err != nil {
			global.GVA_LOG.Error("保存流程图到需求失败", 
				zap.Uint("requirementId", requirement.ID),
				zap.Error(err),
			)
		}
		global.GVA_LOG.Info("流程图生成并保存成功", 
			zap.Uint("requirementId", requirement.ID))
	} else {
		global.GVA_LOG.Warn("流程图生成失败，跳过保存", 
			zap.Uint("requirementId", requirement.ID))
	}

	return nil
}

// analyzeSimpleRequirement 分析单个需求 - 使用已有的AI服务
func (s *UnifiedAnalysisService) analyzeSimpleRequirement(req *nesma.NesmaRequirement, newVersionID uint, config AnalysisConfig) error {
	// 获取AI服务
	singleRequirementAnalysisService := SingleRequirementAnalysisService{}
	response, err := singleRequirementAnalysisService.executeSingleAnalysis(&SingleAnalysisRequest{
		RequirementID: req.ID,
		ProjectID:     req.ProjectID,
		CycleID:       *req.CycleID,
	})
	if err != nil {
		return err
	}

	err = s.parseAndSaveOptimizedRequirement(response, req.ID, newVersionID)
	if err != nil {
		return err
	}
	return nil
}

// applyDefaultOptimization 应用默认优化
func (s *UnifiedAnalysisService) applyDefaultOptimization(req *nesma.NesmaRequirement, newVersionID uint) error {
	// 更新现有需求记录
	updates := map[string]interface{}{
		"title":                "[优化] " + req.Title,
		"description":          req.Description + "\n\n[系统优化：使用默认NESMA标准进行了格式规范化]",
		"function_type":        "EQ", // 默认为查询类型
		"complexity_level":     "Medium",
		"ai_confidence_score":  0.5,
		"ai_analysis_status":   "completed",
		"ai_analysis_time":     time.Now(),
	}

	err := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("id = ? AND version_id = ?", req.ID, newVersionID).
		Updates(updates).Error

	return err
}

// parseAndSaveOptimizedRequirement 解析并保存优化后的需求
func (s *UnifiedAnalysisService) parseAndSaveOptimizedRequirement(response *SingleAnalysisResult, requirementID uint, newVersionID uint) error {

	// 更新现有需求记录
	updates := map[string]interface{}{
		"title":                response.OptimizedInfo.Title,
		"description":          response.OptimizedInfo.Description,
		"function_type":        response.OptimizedInfo.FunctionType,
		"complexity_level":     response.OptimizedInfo.Complexity,
		"ai_confidence_score":  response.Confidence,
		"ai_analysis_status":   "completed",
		"status":               "completed",
		"ai_analysis_time":     time.Now(),
		"business_value":       response.OptimizedInfo.BusinessValue,
		"afp":                  response.OptimizedInfo.RecommendedAFP,
		"ufp":                  response.OptimizedInfo.RecommendedUFP,
		"notes":                response.AnalysisNote,
	}

	err := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("id = ? AND version_id = ?", requirementID, newVersionID).
		Updates(updates).Error

	return err
}

// GetAnalysisProgress 获取分析进度
func (s *UnifiedAnalysisService) GetAnalysisProgress(taskID uint) (*AnalysisProgressResponse, error) {
	var task nesma.NesmaRequirementAnalysisTask
	err := global.GVA_DB.Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, fmt.Errorf("找不到分析任务: %w", err)
	}

	response := &AnalysisProgressResponse{
		TaskID:         task.ID,
		Status:         task.Status,
		Progress:       task.Progress,
		ProcessedCount: task.ProcessedCount,
		TotalCount:     task.TotalCount,
		ErrorMsg:       task.ErrorMsg,
	}

	// 根据状态设置消息
	switch task.Status {
	case "pending":
		response.Message = "分析任务等待中..."
	case "running":
		response.Message = fmt.Sprintf("正在分析需求 (%d/%d)", task.ProcessedCount, task.TotalCount)
	case "completed":
		response.Message = fmt.Sprintf("分析完成！成功处理 %d 个需求", task.SuccessCount)
	case "failed":
		response.Message = "分析失败：" + task.ErrorMsg
	default:
		response.Message = "未知状态"
	}

	return response, nil
}

// createAnalysisVersion 创建分析版本并复制源版本内容
func (s *UnifiedAnalysisService) createAnalysisVersion(cycleID, sourceVersionID uint) (*nesma.NesmaRequirementVersion, error) {
	// 1. 获取源版本信息
	var sourceVersion nesma.NesmaRequirementVersion
	err := global.GVA_DB.Where("id = ?", sourceVersionID).First(&sourceVersion).Error
	if err != nil {
		return nil, fmt.Errorf("源版本不存在 (ID: %d): %w", sourceVersionID, err)
	}

	// 2. 生成新版本号
	newVersionNum := s.generateNextVersionNumber(cycleID)

	// 3. 创建新版本
	newVersion := &nesma.NesmaRequirementVersion{
		CycleID:     cycleID,
		Version:     newVersionNum,
		VersionType: "ai_analyzed",
		CreatedBy:   "system",
		Summary:     fmt.Sprintf("基于%s的AI分析版本", sourceVersion.Version),
	}

	if err := global.GVA_DB.Create(newVersion).Error; err != nil {
		return nil, fmt.Errorf("创建新版本失败: %w", err)
	}

	global.GVA_LOG.Info("创建新版本成功", 
		zap.Uint("cycleId", cycleID),
		zap.String("sourceVersion", sourceVersion.Version),
		zap.String("newVersion", newVersion.Version),
		zap.Uint("newVersionId", newVersion.ID))

	// 4. 复制源版本的需求到新版本
	err = s.copyRequirementsToNewVersion(sourceVersionID, newVersion.ID)
	if err != nil {
		// 如果复制失败，删除已创建的版本
		global.GVA_DB.Delete(newVersion)
		return nil, fmt.Errorf("复制需求到新版本失败: %w", err)
	}

	return newVersion, nil
}

// generateNextVersionNumber 生成下一个版本号
func (s *UnifiedAnalysisService) generateNextVersionNumber(cycleID uint) string {
	var lastVersion nesma.NesmaRequirementVersion
	err := global.GVA_DB.Where("cycle_id = ?", cycleID).
		Order("created_at DESC").
		First(&lastVersion).Error

	if err != nil {
		return "v1.0" // 如果没有版本，从v1.0开始
	}

	// 解析当前版本号
	currentVersion := lastVersion.Version
	
	// 检查版本号格式是否为vx.x
	if len(currentVersion) < 3 || currentVersion[0] != 'v' {
		global.GVA_LOG.Warn("版本号格式异常，重置为v1.0", 
			zap.String("currentVersion", currentVersion))
		return "v1.0"
	}
	
	// 提取主版本号和次版本号
	versionParts := strings.Split(currentVersion[1:], ".")
	if len(versionParts) != 2 {
		global.GVA_LOG.Warn("版本号格式异常，重置为v1.0", 
			zap.String("currentVersion", currentVersion))
		return "v1.0"
	}
	
	// 解析版本号
	majorStr := versionParts[0]
	minorStr := versionParts[1]
	
	major, err := strconv.Atoi(majorStr)
	if err != nil {
		global.GVA_LOG.Warn("主版本号解析失败，重置为v1.0", 
			zap.String("currentVersion", currentVersion),
			zap.Error(err))
		return "v1.0"
	}
	
	minor, err := strconv.Atoi(minorStr)
	if err != nil {
		global.GVA_LOG.Warn("次版本号解析失败，重置为v1.0", 
			zap.String("currentVersion", currentVersion),
			zap.Error(err))
		return "v1.0"
	}
	
	// 生成下一个版本号
	minor++
	if minor > 9 {
		major++
		minor = 0
	}
	
	nextVersion := fmt.Sprintf("v%d.%d", major, minor)
	
	global.GVA_LOG.Info("生成下一个版本号", 
		zap.String("currentVersion", currentVersion),
		zap.String("nextVersion", nextVersion))
	
	return nextVersion
}

// copyRequirementsToNewVersion 复制需求到新版本
func (s *UnifiedAnalysisService) copyRequirementsToNewVersion(sourceVersionID, targetVersionID uint) error {
	// 1. 获取源版本的所有需求
	var sourceRequirements []nesma.NesmaRequirement
	err := global.GVA_DB.Where("version_id = ?", sourceVersionID).Find(&sourceRequirements).Error
	if err != nil {
		return fmt.Errorf("获取源版本需求失败: %w", err)
	}

	if len(sourceRequirements) == 0 {
		global.GVA_LOG.Warn("源版本没有需求，跳过复制", zap.Uint("sourceVersionId", sourceVersionID))
		return nil
	}

	global.GVA_LOG.Info("开始复制需求", 
		zap.Uint("sourceVersionId", sourceVersionID),
		zap.Uint("targetVersionId", targetVersionID),
		zap.Int("requirementCount", len(sourceRequirements)))

	// 2. 逐个复制需求
	copiedCount := 0
	for _, req := range sourceRequirements {
		// 创建需求的副本
		copiedReq := req
		copiedReq.ID = 0 // 清空ID，让数据库生成新ID
		copiedReq.VersionID = &targetVersionID // 设置新版本ID
		
		// 重置AI分析相关字段
		copiedReq.AIAnalysisStatus = "pending"
		copiedReq.AIDescription = ""
		copiedReq.AIGeneratedTitle = ""
		copiedReq.AIComplexityScore = nil
		copiedReq.AIConfidenceScore = nil
		copiedReq.AIAnalysisTime = nil
		copiedReq.AIOptimizationApplied = false
		copiedReq.AIOptimizationAppliedAt = nil

		// 保存复制的需求
		if err := global.GVA_DB.Create(&copiedReq).Error; err != nil {
			global.GVA_LOG.Error("复制需求失败", 
				zap.Uint("originalId", req.ID),
				zap.String("title", req.Title),
				zap.Error(err))
			continue // 继续复制其他需求
		}
		copiedCount++
	}

	global.GVA_LOG.Info("需求复制完成", 
		zap.Uint("sourceVersionId", sourceVersionID),
		zap.Uint("targetVersionId", targetVersionID),
		zap.Int("copiedCount", copiedCount),
		zap.Int("totalCount", len(sourceRequirements)))

	return nil
}

// getRequirementsForAnalysis 根据配置获取需要分析的需求
func (s *UnifiedAnalysisService) getRequirementsForAnalysis(versionID uint, config AnalysisConfig) ([]nesma.NesmaRequirement, error) {
	var requirements []nesma.NesmaRequirement
	
	// 构建基础查询
	query := global.GVA_DB.Where("version_id = ?", versionID)

	// 根据分析范围过滤
	switch config.AnalysisScope {
	case "all":
		// 分析所有配置中指定级别的需求
		if len(config.IncludeLevels) > 0 {
			query = query.Where("level IN (?)", config.IncludeLevels)
		}
	case "specific_levels":
		// 只分析特定级别
		if len(config.IncludeLevels) > 0 {
			query = query.Where("level IN (?)", config.IncludeLevels)
	} else {
			// 如果没有指定级别，默认分析L3和L4
			query = query.Where("level IN (3, 4)")
		}
	case "incremental":
		// 增量分析：只分析那些尚未进行AI分析的需求
		if len(config.IncludeLevels) > 0 {
			query = query.Where("level IN (?) AND (ai_analysis_status = 'pending' OR ai_analysis_status = '')", config.IncludeLevels)
		} else {
			query = query.Where("level IN (3, 4) AND (ai_analysis_status = 'pending' OR ai_analysis_status = '')")
		}
	default:
		// 默认分析L3和L4级别
		query = query.Where("level IN (3, 4)")
	}

	// 执行查询
	err := query.Order("level ASC, order_index ASC").Find(&requirements).Error
	if err != nil {
		return nil, fmt.Errorf("查询需求失败: %w", err)
	}
	
	global.GVA_LOG.Info("获取到分析需求", 
		zap.Uint("versionId", versionID), 
		zap.String("analysisScope", config.AnalysisScope),
		zap.Any("includeLevels", config.IncludeLevels),
		zap.Int("requirementCount", len(requirements)))
	
	return requirements, nil
}

// parseAnalysisConfig 解析分析配置
func (s *UnifiedAnalysisService) parseAnalysisConfig(req *nesmaReq.UnifiedAnalysisRequest) AnalysisConfig {
	config := DefaultAnalysisConfig // 从默认配置开始
	modelSelect := "deepseek-chat"
	// 从请求参数中解析配置
	if req.Parameters != nil {
		global.GVA_LOG.Info("解析用户提交的参数", zap.Any("parameters", req.Parameters))

		// 解析是否包含描述优化
		if includeDescription, ok := req.Parameters["includeDescription"].(bool); ok {
			config.OptimizeDescription = includeDescription
			global.GVA_LOG.Info("设置描述优化", zap.Bool("includeDescription", includeDescription))
		}

		// 解析是否包含L4需求
		if includeLevel4, ok := req.Parameters["includeLevel4"].(bool); ok {
			if includeLevel4 {
				// 如果用户选择了L4，确保包含L4级别
				config.IncludeLevels = []int{3, 4}
			} else {
				// 如果用户没选择L4，只分析L3
				config.IncludeLevels = []int{3}
			}
			global.GVA_LOG.Info("设置级别包含", zap.Any("includeLevels", config.IncludeLevels))
		}

		// 解析是否包含Mermaid流程图
		if includeMermaid, ok := req.Parameters["includeMermaid"].(bool); ok {
			config.GenerateMermaid = includeMermaid
			global.GVA_LOG.Info("设置Mermaid生成", zap.Bool("includeMermaid", includeMermaid))
		}

		// 解析是否自动保存L4需求
		if autoSaveL4, ok := req.Parameters["autoSaveL4"].(bool); ok {
			// 这个参数暂时记录，后续可能会用到
			global.GVA_LOG.Info("自动保存L4设置", zap.Bool("autoSaveL4", autoSaveL4))
		}

		// 解析L4处理策略
		if l4Strategy, ok := req.Parameters["l4Strategy"].(string); ok && l4Strategy != "" {
			// 根据策略调整处理方式
			switch l4Strategy {
			case "comprehensive":
				config.MaxConcurrency = 3 // 综合分析时降低并发，提高质量
			case "focused":
				modelSelect = "deepseek-reasoner"
				config.MaxConcurrency = 5 // 简单分析时可以提高并发
			}
			global.GVA_LOG.Info("设置L4策略", zap.String("l4Strategy", l4Strategy))
		}

		// 解析L4最大数量
		if maxL4Count, ok := req.Parameters["maxL4Count"].(float64); ok && maxL4Count > 0 {
			// 这个参数可以用于限制分析的需求数量
			maxCount := int(maxL4Count)
			global.GVA_LOG.Info("设置L4最大数量", zap.Int("maxL4Count", maxCount))
		}

		// 解析流程图类型
		if diagramType, ok := req.Parameters["diagramType"].(string); ok && diagramType != "" {
			// 这个参数记录，用于Mermaid生成时使用
			global.GVA_LOG.Info("设置流程图类型", zap.String("diagramType", diagramType))
		}

		// 解析详细级别
		if detailLevel, ok := req.Parameters["detailLevel"].(string); ok && detailLevel != "" {
			// 根据详细级别调整分析深度
			switch detailLevel {
			case "detailed":
				config.OptimizeTitle = true
				config.OptimizeDescription = true
			case "simple":
				config.OptimizeTitle = true
				config.OptimizeDescription = false
			}
			global.GVA_LOG.Info("设置详细级别", zap.String("detailLevel", detailLevel))
		}

		// 解析选中的选项数组
		if selectedOptions, ok := req.Parameters["selectedOptions"].([]interface{}); ok && len(selectedOptions) > 0 {
			global.GVA_LOG.Info("用户选择的选项", zap.Any("selectedOptions", selectedOptions))
			
			// 重置所有选项为false，然后根据用户选择设置
			config.OptimizeDescription = false
			config.GenerateMermaid = false
			includeL4 := false

			for _, option := range selectedOptions {
				if optStr, ok := option.(string); ok {
					switch optStr {
					case "description":
						config.OptimizeDescription = true
						global.GVA_LOG.Info("启用描述优化")
					case "level4":
						includeL4 = true
						global.GVA_LOG.Info("启用L4分析")
					case "mermaid":
						config.GenerateMermaid = true
						global.GVA_LOG.Info("启用Mermaid生成")
					}
				}
			}

			// 根据是否包含L4调整级别
			if includeL4 {
				config.IncludeLevels = []int{3, 4}
				} else {
				config.IncludeLevels = []int{3}
			}
		}

		// 解析其他可能的参数（保持向后兼容）
		if scope, ok := req.Parameters["analysis_scope"].(string); ok && scope != "" {
			config.AnalysisScope = scope
		}

		if aiModel, ok := req.Parameters["ai_model"].(string); ok && aiModel != "" {
			config.AIModel = modelSelect
		}

		if batchSize, ok := req.Parameters["batch_size"].(float64); ok && batchSize > 0 {
			config.BatchSize = int(batchSize)
		}

		if maxRetries, ok := req.Parameters["max_retries"].(float64); ok && maxRetries > 0 {
			config.MaxRetries = int(maxRetries)
		}

		if maxConcurrency, ok := req.Parameters["max_concurrency"].(float64); ok && maxConcurrency > 0 {
			config.MaxConcurrency = int(maxConcurrency)
		}
	}

	global.GVA_LOG.Info("最终解析的分析配置", 
		zap.String("analysisScope", config.AnalysisScope),
		zap.Any("includeLevels", config.IncludeLevels),
		zap.Bool("optimizeTitle", config.OptimizeTitle),
		zap.Bool("optimizeDescription", config.OptimizeDescription),
		zap.Bool("generateMermaid", config.GenerateMermaid),
		zap.String("aiModel", config.AIModel),
		zap.Int("maxConcurrency", config.MaxConcurrency))

	return config
}

// getActiveVersionForCycle 获取周期的活跃版本（最新版本）
func (s *UnifiedAnalysisService) getActiveVersionForCycle(cycleID uint) (*nesma.NesmaRequirementVersion, error) {
	var activeVersion nesma.NesmaRequirementVersion
	
	// 查找该周期的最新版本
	err := global.GVA_DB.Where("cycle_id = ?", cycleID).
		Order("created_at DESC").
		First(&activeVersion).Error
	
	if err != nil {
		global.GVA_LOG.Warn("未找到活跃版本，将创建初始版本", 
			zap.Uint("cycleId", cycleID), 
			zap.Error(err))
		
		// 如果没有找到版本，创建一个初始版本
		return s.createInitialVersion(cycleID)
	}

	global.GVA_LOG.Info("找到活跃版本", 
		zap.Uint("cycleId", cycleID),
		zap.Uint("versionId", activeVersion.ID),
		zap.String("version", activeVersion.Version))

	return &activeVersion, nil
}

// createInitialVersion 创建初始版本
func (s *UnifiedAnalysisService) createInitialVersion(cycleID uint) (*nesma.NesmaRequirementVersion, error) {
	initialVersion := &nesma.NesmaRequirementVersion{
		CycleID:     cycleID,
		Version:     "v1.0",
		VersionType: "initial",
		CreatedBy:   "system",
		Summary:     "初始版本",
	}

	if err := global.GVA_DB.Create(initialVersion).Error; err != nil {
		return nil, fmt.Errorf("创建初始版本失败: %w", err)
	}

	global.GVA_LOG.Info("创建初始版本成功", 
		zap.Uint("cycleId", cycleID),
		zap.Uint("versionId", initialVersion.ID))

	return initialVersion, nil
}

// DetailedAnalysisResult 详细分析结果
type DetailedAnalysisResult struct {
	TaskInfo         *nesma.NesmaRequirementAnalysisTask `json:"task_info"`
	SourceVersion    *nesma.NesmaRequirementVersion      `json:"source_version"`
	TargetVersion    *nesma.NesmaRequirementVersion      `json:"target_version"`
	AnalyzedRequirements []RequirementComparisonItem     `json:"analyzed_requirements"`
	Statistics       AnalysisStatistics                  `json:"statistics"`
}

// RequirementComparisonItem 需求对比项
type RequirementComparisonItem struct {
	OriginalRequirement *nesma.NesmaRequirement `json:"original_requirement"`
	OptimizedRequirement *nesma.NesmaRequirement `json:"optimized_requirement"`
	Changes              []string                `json:"changes"`
	ImprovementScore     float64                 `json:"improvement_score"`
}

// AnalysisStatistics 分析统计
type AnalysisStatistics struct {
	TotalRequirements    int     `json:"total_requirements"`
	ProcessedCount       int     `json:"processed_count"`
	SuccessCount         int     `json:"success_count"`
	FailedCount          int     `json:"failed_count"`
	SuccessRate          float64 `json:"success_rate"`
	AverageConfidence    float64 `json:"average_confidence"`
	ProcessingTime       string  `json:"processing_time"`
}

// GetDetailedAnalysisResult 获取详细分析结果
func (s *UnifiedAnalysisService) GetDetailedAnalysisResult(taskID uint, detailLevel string) (*DetailedAnalysisResult, error) {
	// 1. 获取任务信息
	var task nesma.NesmaRequirementAnalysisTask
	err := global.GVA_DB.Where("id = ?", taskID).First(&task).Error
		if err != nil {
		return nil, fmt.Errorf("找不到分析任务: %w", err)
	}

	if task.Status != "completed" {
		return nil, fmt.Errorf("分析任务尚未完成，当前状态: %s", task.Status)
	}

	// 2. 获取源版本和目标版本信息
	var sourceVersion, targetVersion nesma.NesmaRequirementVersion
	
	if err := global.GVA_DB.Where("id = ?", task.SourceVersionID).First(&sourceVersion).Error; err != nil {
		return nil, fmt.Errorf("获取源版本失败: %w", err)
	}

	if task.TargetVersionID != nil {
		if err := global.GVA_DB.Where("id = ?", *task.TargetVersionID).First(&targetVersion).Error; err != nil {
			return nil, fmt.Errorf("获取目标版本失败: %w", err)
		}
	}

	// 3. 构建详细结果
	result := &DetailedAnalysisResult{
		TaskInfo:      &task,
		SourceVersion: &sourceVersion,
		TargetVersion: &targetVersion,
		Statistics: AnalysisStatistics{
			TotalRequirements: task.TotalCount,
			ProcessedCount:    task.ProcessedCount,
			SuccessCount:      task.SuccessCount,
			FailedCount:       task.FailedCount,
		},
	}

	// 计算成功率
	if task.TotalCount > 0 {
		result.Statistics.SuccessRate = float64(task.SuccessCount) / float64(task.TotalCount) * 100
	}

	// 计算处理时间
	if task.StartTime != nil && task.EndTime != nil {
		duration := task.EndTime.Sub(*task.StartTime)
		result.Statistics.ProcessingTime = duration.String()
	}

	// 4. 根据detail_level决定是否加载需求对比
	if detailLevel == "detailed" && task.TargetVersionID != nil {
		comparisons, avgConfidence, err := s.buildRequirementComparisons(task.SourceVersionID, *task.TargetVersionID)
		if err != nil {
			global.GVA_LOG.Warn("构建需求对比失败", zap.Error(err))
		} else {
			result.AnalyzedRequirements = comparisons
			result.Statistics.AverageConfidence = avgConfidence
		}
	}

	global.GVA_LOG.Info("获取详细分析结果成功", 
		zap.Uint("taskId", taskID),
		zap.String("detailLevel", detailLevel),
		zap.Int("requirementCount", len(result.AnalyzedRequirements)))

	return result, nil
}

// buildRequirementComparisons 构建需求对比
func (s *UnifiedAnalysisService) buildRequirementComparisons(sourceVersionID, targetVersionID uint) ([]RequirementComparisonItem, float64, error) {
	// 获取源版本需求
	var sourceRequirements []nesma.NesmaRequirement
	err := global.GVA_DB.Where("version_id = ?", sourceVersionID).Find(&sourceRequirements).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取源版本需求失败: %w", err)
	}

	// 获取目标版本需求
	var targetRequirements []nesma.NesmaRequirement
	err = global.GVA_DB.Where("version_id = ?", targetVersionID).Find(&targetRequirements).Error
	if err != nil {
		return nil, 0, fmt.Errorf("获取目标版本需求失败: %w", err)
	}

	// 构建需求映射（按Code或Title匹配）
	targetMap := make(map[string]*nesma.NesmaRequirement)
	for i := range targetRequirements {
		target := &targetRequirements[i]
		key := target.Code
		if key == "" {
			key = target.Title
		}
		targetMap[key] = target
	}

	var comparisons []RequirementComparisonItem
	var totalConfidence float64
	var confidenceCount int

	// 构建对比项
	for i := range sourceRequirements {
		source := &sourceRequirements[i]
		key := source.Code
		if key == "" {
			key = source.Title
		}

		if target, exists := targetMap[key]; exists {
			comparison := RequirementComparisonItem{
				OriginalRequirement:  source,
				OptimizedRequirement: target,
				Changes:              s.detectChanges(source, target),
				ImprovementScore:     s.calculateImprovementScore(source, target),
			}
			comparisons = append(comparisons, comparison)

			// 累计置信度
			if target.AIConfidenceScore != nil {
				totalConfidence += *target.AIConfidenceScore
				confidenceCount++
			}
		}
	}

	// 计算平均置信度
	var avgConfidence float64
	if confidenceCount > 0 {
		avgConfidence = totalConfidence / float64(confidenceCount)
	}

	return comparisons, avgConfidence, nil
}

// detectChanges 检测变更
func (s *UnifiedAnalysisService) detectChanges(original, optimized *nesma.NesmaRequirement) []string {
	var changes []string

	if original.Title != optimized.Title {
		changes = append(changes, "标题已优化")
	}

	if original.Description != optimized.Description {
		changes = append(changes, "描述已优化")
	}

	if original.FunctionType != optimized.FunctionType {
		changes = append(changes, fmt.Sprintf("功能类型: %s → %s", original.FunctionType, optimized.FunctionType))
	}

	if original.ComplexityLevel != optimized.ComplexityLevel {
		changes = append(changes, fmt.Sprintf("复杂度: %s → %s", original.ComplexityLevel, optimized.ComplexityLevel))
	}

	if optimized.AIConfidenceScore != nil {
		changes = append(changes, fmt.Sprintf("AI置信度: %.2f", *optimized.AIConfidenceScore))
	}

	return changes
}

// calculateImprovementScore 计算改进得分
func (s *UnifiedAnalysisService) calculateImprovementScore(original, optimized *nesma.NesmaRequirement) float64 {
	score := 0.0
	
	// 基础改进分数
	if original.Title != optimized.Title {
		score += 0.3
	}
	if original.Description != optimized.Description {
		score += 0.4
	}
	if original.FunctionType != optimized.FunctionType {
		score += 0.2
	}
	if original.ComplexityLevel != optimized.ComplexityLevel {
		score += 0.1
	}

	// AI置信度加权
	if optimized.AIConfidenceScore != nil {
		score *= *optimized.AIConfidenceScore
	}

	return score
}

// ParseAnalysisConfig 导出方法用于测试
func (s *UnifiedAnalysisService) ParseAnalysisConfig(req *nesmaReq.UnifiedAnalysisRequest) AnalysisConfig {
	return s.parseAnalysisConfig(req)
}
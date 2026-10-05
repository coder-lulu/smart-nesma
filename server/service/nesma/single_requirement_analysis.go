package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

type SingleRequirementAnalysisService struct{}

// HierarchicalAnalysisRequest 层次化分析请求
type HierarchicalAnalysisRequest struct {
	RequirementID         uint   `json:"requirementId" binding:"required"`
	ProjectID             uint   `json:"projectId" binding:"required"`
	CycleID               uint   `json:"cycleId" binding:"required"`
	AnalysisType          string `json:"analysisType"`          // single, generate_l4, optimize_l4
	AutoGenerateL4        bool   `json:"autoGenerateL4"`        // 是否自动生成L4
	IncludeKnowledgeBase  bool   `json:"includeKnowledgeBase"`  // 是否使用知识库
	MaxL4Count            int    `json:"maxL4Count"`            // 最大L4生成数量
	L4GenerationStrategy  string `json:"l4GenerationStrategy"` // expand, create_new, optimize_existing
}

// HierarchicalAnalysisResult 层次化分析结果
type HierarchicalAnalysisResult struct {
	RequirementID        uint                    `json:"requirementId"`
	AnalysisType         string                  `json:"analysisType"`
	OriginalInfo         *AnalysisRequirementInfo        `json:"originalInfo"`
	OptimizedInfo        *AnalysisRequirementInfo        `json:"optimizedInfo"`
	AnalysisNote         string                  `json:"analysisNote"`
	Confidence           float64                 `json:"confidence"`
	Timestamp            time.Time               `json:"timestamp"`
	
	// L4生成相关
	L4GenerationResult   *L4GenerationSummary    `json:"l4GenerationResult,omitempty"`
	KnowledgeReferences  []KnowledgeReference    `json:"knowledgeReferences,omitempty"`
	ProcessingSteps      []ProcessingStep        `json:"processingSteps"`
}

// L4GenerationSummary L4生成摘要
type L4GenerationSummary struct {
	Strategy             string              `json:"strategy"`             // expand, create_new, optimize_existing
	ExistingL4Count      int                 `json:"existingL4Count"`
	GeneratedL4Count     int                 `json:"generatedL4Count"`
	OptimizedL4Count     int                 `json:"optimizedL4Count"`
	L4Suggestions        []L4Suggestion      `json:"l4Suggestions"`
	GenerationNotes      string              `json:"generationNotes"`
	RecommendedActions   []string            `json:"recommendedActions"`
}

// ProcessingStep 处理步骤
type ProcessingStep struct {
	Step        string    `json:"step"`
	Description string    `json:"description"`
	Status      string    `json:"status"`     // pending, running, completed, failed
	StartTime   time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Details     string    `json:"details"`
	ErrorMsg    string    `json:"errorMsg,omitempty"`
}

// HierarchicalAnalysisTask 层次化分析任务
type HierarchicalAnalysisTask struct {
	ID                   uint                        `json:"id"`
	RequirementID        uint                        `json:"requirementId"`
	ProjectID            uint                        `json:"projectId"`
	CycleID              uint                        `json:"cycleId"`
	AnalysisType         string                      `json:"analysisType"`
	Status               string                      `json:"status"`        // pending, running, completed, failed
	Progress             int                         `json:"progress"`      // 0-100
	CurrentStep          string                      `json:"currentStep"`
	StepDesc             string                      `json:"stepDesc"`
	AnimationType        string                      `json:"animationType"`
	Result               *HierarchicalAnalysisResult `json:"result"`
	ErrorMsg             string                      `json:"errorMsg"`
	StartTime            time.Time                   `json:"startTime"`
	EndTime              *time.Time                  `json:"endTime"`
	CreatedAt            time.Time                   `json:"createdAt"`
	ProcessingSteps      []ProcessingStep            `json:"processingSteps"`
	Config               map[string]interface{}      `json:"config"`
}

// L4Suggestion L4功能点建议（重新定义以避免冲突）
type L4Suggestion struct {
	SuggestedTitle       string  `json:"suggestedTitle"`
	SuggestedDescription string  `json:"suggestedDescription"`
	SuggestedCode        string  `json:"suggestedCode"`
	FunctionType         string  `json:"functionType"`         // EI/EO/EQ/ILF/EIF
	BusinessValue        string  `json:"businessValue"`
	AcceptanceCriteria   string  `json:"acceptanceCriteria"`
	EstimatedComplexity  string  `json:"estimatedComplexity"`   // 简单/中等/复杂
	RecommendedAFP       float64 `json:"recommendedAFP"`
	RecommendedUFP       float64 `json:"recommendedUFP"`
	Priority             int     `json:"priority"`               // 1-5
	Confidence           float64 `json:"confidence"`             // 0-1
	GenerationReason     string  `json:"generationReason"`
	RelatedKnowledge     string  `json:"relatedKnowledge"`
	ExistingL4ID         *uint   `json:"existingL4Id,omitempty"` // 如果是优化现有L4
	ActionType           string  `json:"actionType"`             // create_new, optimize_existing, expand
}

// SingleAnalysisRequest 单需求分析请求
type SingleAnalysisRequest struct {
	RequirementID uint `json:"requirementId" binding:"required"`
	ProjectID     uint `json:"projectId" binding:"required"`
	CycleID       uint `json:"cycleId" binding:"required"`
}

// SingleAnalysisResult 单需求分析结果
type SingleAnalysisResult struct {
	RequirementID uint   `json:"requirementId"`
	OriginalInfo  *AnalysisRequirementInfo `json:"originalInfo"`
	OptimizedInfo *AnalysisRequirementInfo `json:"optimizedInfo"`
	AnalysisNote  string `json:"analysisNote"`
	Confidence    float64 `json:"confidence"`
	Timestamp     time.Time `json:"timestamp"`
}

// AnalysisRequirementInfo 分析用需求信息结构（重命名避免冲突）
type AnalysisRequirementInfo struct {
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	FunctionType    string  `json:"functionType"`
	Complexity      string  `json:"complexity"`
	BusinessValue   string  `json:"businessValue"`
	RecommendedAFP  *float64 `json:"recommendedAfp"`
	RecommendedUFP  *float64 `json:"recommendedUfp"`
}

// AnalysisContext 分析上下文
type AnalysisContext struct {
	Project        *nesma.NesmaProject      `json:"project"`
	Cycle          *nesma.NesmaProjectCycle `json:"cycle"`
	ParentRequirements []nesma.NesmaRequirement `json:"parentRequirements"`
	SiblingRequirements []nesma.NesmaRequirement `json:"siblingRequirements"`
	KnowledgeRefs  []string `json:"knowledgeRefs"`
}

// SingleAnalysisTask 单需求分析任务
type SingleAnalysisTask struct {
	ID            uint      `json:"id"`
	RequirementID uint      `json:"requirementId"`
	ProjectID     uint      `json:"projectId"`
	CycleID       uint      `json:"cycleId"`
	Status        string    `json:"status"`        // pending, running, completed, failed
	Progress      int       `json:"progress"`      // 0-100
	CurrentStage  string    `json:"currentStage"`  // 当前阶段
	StageDesc     string    `json:"stageDesc"`     // 阶段描述
	AnimationType string    `json:"animationType"` // 动画类型提示
	Result        *SingleAnalysisResult `json:"result"`
	ErrorMsg      string    `json:"errorMsg"`
	StartTime     time.Time `json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
	CreatedAt     time.Time `json:"createdAt"`
	
	// 阶段历史记录
	StageHistory  []AnalysisStage `json:"stageHistory"`
}

// AnalysisStage 分析阶段记录
type AnalysisStage struct {
	Stage       string    `json:"stage"`
	Description string    `json:"description"`
	Progress    int       `json:"progress"`
	StartTime   time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Duration    int64     `json:"duration"` // 毫秒
	Status      string    `json:"status"`   // running, completed, failed
}

// 内存中的任务存储（生产环境建议使用Redis或数据库）
var taskStore = make(map[uint]*SingleAnalysisTask)
var taskIDCounter uint = 1000

// AnalyzeSingleRequirementAsync 异步分析单个需求
func (s *SingleRequirementAnalysisService) AnalyzeSingleRequirementAsync(req *SingleAnalysisRequest) (*SingleAnalysisTask, error) {
	global.GVA_LOG.Info("开始创建单需求分析任务", zap.Uint("requirementId", req.RequirementID))

	// 创建分析任务
	taskIDCounter++
	task := &SingleAnalysisTask{
		ID:            taskIDCounter,
		RequirementID: req.RequirementID,
		ProjectID:     req.ProjectID,
		CycleID:       req.CycleID,
		Status:        "pending",
		Progress:      0,
		StartTime:     time.Now(),
		CreatedAt:     time.Now(),
	}

	// 存储任务
	taskStore[task.ID] = task

	// 启动异步分析
	go s.runSingleAnalysisTask(task)

	global.GVA_LOG.Info("单需求分析任务已创建", zap.Uint("taskId", task.ID), zap.Uint("requirementId", req.RequirementID))
	return task, nil
}

// runSingleAnalysisTask 执行单需求分析任务
func (s *SingleRequirementAnalysisService) runSingleAnalysisTask(task *SingleAnalysisTask) {
	global.GVA_LOG.Info("开始执行单需求分析", zap.Uint("taskId", task.ID), zap.Uint("requirementId", task.RequirementID))

	// 阶段1：初始化
	s.updateTaskStage(task, "initializing", "正在初始化分析环境...", 5, "pulse")
	time.Sleep(500 * time.Millisecond) // 模拟初始化时间
	
	// 阶段2：需求验证
	s.updateTaskStage(task, "validating", "正在验证需求信息...", 15, "scan")
	
	// 执行分析
	req := &SingleAnalysisRequest{
		RequirementID: task.RequirementID,
		ProjectID:     task.ProjectID,
		CycleID:       task.CycleID,
	}

	// 阶段3：获取需求信息
	s.updateTaskStage(task, "fetching", "正在获取需求详情...", 25, "loading")
	requirement, err := s.getRequirementWithValidation(req.RequirementID)
	if err != nil {
		s.setTaskFailed(task, fmt.Sprintf("获取需求信息失败: %v", err))
		return
	}

	// 阶段4：构建分析上下文
	s.updateTaskStage(task, "context_building", "正在收集项目背景和上下文信息...", 35, "network")
	context, err := s.buildAnalysisContext(requirement, req.ProjectID, req.CycleID)
	if err != nil {
		global.GVA_LOG.Error("构建分析上下文失败", zap.Error(err))
		// 非致命错误，继续分析
		context = &AnalysisContext{}
	}

	// 阶段5：智能分析
	s.updateTaskStage(task, "ai_analyzing", "AI正在进行深度需求分析...", 50, "brain")
	
	// 构建prompt
	prompt := s.buildContextualAnalysisPrompt(requirement, context)
	
	// 阶段6：AI调用
	s.updateTaskStage(task, "ai_calling", "正在调用AI分析服务...", 70, "api")
	aiResult, err := s.callAIForSingleRequirement(prompt)
	if err != nil {
		s.setTaskFailed(task, fmt.Sprintf("AI分析失败: %v", err))
		return
	}

	// 阶段7：结果解析
	s.updateTaskStage(task, "parsing", "正在解析AI分析结果...", 85, "parse")
	result := s.buildAnalysisResult(requirement, aiResult)
	
	// 阶段8：完成
	s.updateTaskStage(task, "finalizing", "正在整理分析结果...", 95, "check")
	time.Sleep(300 * time.Millisecond) // 短暂延迟让用户看到完成阶段
	
	// 任务完成
	task.Status = "completed"
	task.Progress = 100
	task.CurrentStage = "completed"
	task.StageDesc = "分析完成！AI已为您优化需求"
	task.AnimationType = "success"
	task.Result = result
	endTime := time.Now()
	task.EndTime = &endTime
	
	// 完成最后一个阶段
	s.completeCurrentStage(task)

	global.GVA_LOG.Info("单需求分析完成", zap.Uint("taskId", task.ID), zap.Uint("requirementId", task.RequirementID))
}

// updateTaskStage 更新任务阶段
func (s *SingleRequirementAnalysisService) updateTaskStage(task *SingleAnalysisTask, stage, description string, progress int, animationType string) {
	// 完成当前阶段
	s.completeCurrentStage(task)
	
	// 开始新阶段
	now := time.Now()
	newStage := AnalysisStage{
		Stage:       stage,
		Description: description,
		Progress:    progress,
		StartTime:   now,
		Status:      "running",
	}
	
	task.CurrentStage = stage
	task.StageDesc = description
	task.Progress = progress
	task.AnimationType = animationType
	task.Status = "running"
	task.StageHistory = append(task.StageHistory, newStage)
	
	global.GVA_LOG.Info("任务阶段更新", 
		zap.Uint("taskId", task.ID),
		zap.String("stage", stage),
		zap.String("description", description),
		zap.Int("progress", progress))
}

// completeCurrentStage 完成当前阶段
func (s *SingleRequirementAnalysisService) completeCurrentStage(task *SingleAnalysisTask) {
	if len(task.StageHistory) > 0 {
		lastStage := &task.StageHistory[len(task.StageHistory)-1]
		if lastStage.Status == "running" {
			endTime := time.Now()
			lastStage.EndTime = &endTime
			lastStage.Status = "completed"
			lastStage.Duration = endTime.Sub(lastStage.StartTime).Milliseconds()
		}
	}
}

// setTaskFailed 设置任务失败
func (s *SingleRequirementAnalysisService) setTaskFailed(task *SingleAnalysisTask, errorMsg string) {
	task.Status = "failed"
	task.ErrorMsg = errorMsg
	task.CurrentStage = "failed"
	task.StageDesc = "分析失败: " + errorMsg
	task.AnimationType = "error"
	task.Progress = 0
	endTime := time.Now()
	task.EndTime = &endTime
	
	// 完成当前阶段为失败
	if len(task.StageHistory) > 0 {
		lastStage := &task.StageHistory[len(task.StageHistory)-1]
		if lastStage.Status == "running" {
			endTime := time.Now()
			lastStage.EndTime = &endTime
			lastStage.Status = "failed"
			lastStage.Duration = endTime.Sub(lastStage.StartTime).Milliseconds()
		}
	}
	
	global.GVA_LOG.Error("单需求分析失败", zap.Error(fmt.Errorf(errorMsg)), zap.Uint("taskId", task.ID))
}

// GetSingleAnalysisTask 获取单需求分析任务
func (s *SingleRequirementAnalysisService) GetSingleAnalysisTask(taskID uint) (*SingleAnalysisTask, error) {
	task, exists := taskStore[taskID]
	if !exists {
		return nil, fmt.Errorf("任务不存在")
	}
	return task, nil
}

// executeSingleAnalysis 执行单需求分析（原来的逻辑）
func (s *SingleRequirementAnalysisService) executeSingleAnalysis(req *SingleAnalysisRequest) (*SingleAnalysisResult, error) {
	global.GVA_LOG.Info("开始单需求分析", zap.Uint("requirementId", req.RequirementID))

	// 1. 获取需求信息
	requirement, err := s.getRequirementWithValidation(req.RequirementID)
	if err != nil {
		return nil, err
	}

	// 2. 构建分析上下文
	context, err := s.buildAnalysisContext(requirement, req.ProjectID, req.CycleID)
	if err != nil {
		global.GVA_LOG.Error("构建分析上下文失败", zap.Error(err))
		return nil, fmt.Errorf("构建分析上下文失败: %v", err)
	}

	// 3. 构建智能分析prompt
	prompt := s.buildContextualAnalysisPrompt(requirement, context)

	// 4. 调用AI进行分析
	aiResult, err := s.callAIForSingleRequirement(prompt)
	if err != nil {
		global.GVA_LOG.Error("AI分析失败", zap.Error(err))
		return nil, fmt.Errorf("AI分析失败: %v", err)
	}

	// 5. 解析和构建结果
	result := s.buildAnalysisResult(requirement, aiResult)
	
	global.GVA_LOG.Info("单需求分析完成", zap.Uint("requirementId", req.RequirementID))
	return result, nil
}

// getRequirementWithValidation 获取并验证需求信息
func (s *SingleRequirementAnalysisService) getRequirementWithValidation(requirementID uint) (*nesma.NesmaRequirement, error) {
	var requirement nesma.NesmaRequirement
	
	err := global.GVA_DB.First(&requirement, requirementID).Error
	if err != nil {
		return nil, fmt.Errorf("需求不存在: %v", err)
	}

	return &requirement, nil
}

// buildAnalysisContext 构建分析上下文
func (s *SingleRequirementAnalysisService) buildAnalysisContext(requirement *nesma.NesmaRequirement, projectID, cycleID uint) (*AnalysisContext, error) {
	context := &AnalysisContext{}

	// 获取项目信息
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, projectID).Error; err == nil {
		context.Project = &project
	}

	// 获取周期信息
	var cycle nesma.NesmaProjectCycle
	if err := global.GVA_DB.First(&cycle, cycleID).Error; err == nil {
		context.Cycle = &cycle
	}

	// 获取父级需求（1-2级）
	var parentRequirements []nesma.NesmaRequirement
	if requirement.ParentID != nil {
		// 获取直接父级需求
		var directParent nesma.NesmaRequirement
		if err := global.GVA_DB.First(&directParent, *requirement.ParentID).Error; err == nil {
			parentRequirements = append(parentRequirements, directParent)
			
			// 获取更上级的需求
			if directParent.ParentID != nil {
				var grandParent nesma.NesmaRequirement
				if err := global.GVA_DB.First(&grandParent, *directParent.ParentID).Error; err == nil {
					parentRequirements = append([]nesma.NesmaRequirement{grandParent}, parentRequirements...)
				}
			}
		}
	} else {
		// 如果没有直接父级，查找同项目的1-2级需求
		global.GVA_DB.Where("project_id = ? AND level IN (1, 2)", projectID).
			Order("level ASC").
			Find(&parentRequirements)
	}
	context.ParentRequirements = parentRequirements

	// 获取同级需求（相同父级的需求）
	var siblingRequirements []nesma.NesmaRequirement
	if requirement.ParentID != nil {
		global.GVA_DB.Where("parent_id = ? AND id != ?", *requirement.ParentID, requirement.ID).
			Limit(5).
			Find(&siblingRequirements)
	} else {
		// 获取同级别的需求
		global.GVA_DB.Where("project_id = ? AND level = ? AND id != ?", projectID, requirement.Level, requirement.ID).
			Limit(5).
			Find(&siblingRequirements)
	}
	context.SiblingRequirements = siblingRequirements

	// TODO: 集成知识库相关内容
	context.KnowledgeRefs = []string{"NESMA标准", "功能点分析最佳实践"}

	return context, nil
}

// buildContextualAnalysisPrompt 构建包含上下文的分析prompt
func (s *SingleRequirementAnalysisService) buildContextualAnalysisPrompt(requirement *nesma.NesmaRequirement, context *AnalysisContext) string {
	// 构建项目背景信息
	projectInfo := ""
	if context.Project != nil {
		projectInfo = fmt.Sprintf("项目名称：%s\n项目描述：%s\n业务领域：%s\n", 
			context.Project.Name, context.Project.Description, context.Project.Domain)
	}

	// 构建周期信息
	cycleInfo := ""
	if context.Cycle != nil {
		cycleInfo = fmt.Sprintf("建设周期：%s\n周期描述：%s\n", 
			context.Cycle.Name, context.Cycle.Description)
	}

	// 构建父级需求信息
	parentInfo := ""
	if len(context.ParentRequirements) > 0 {
		parentInfo = "上级需求背景：\n"
		for _, parent := range context.ParentRequirements {
			parentInfo += fmt.Sprintf("- L%d: %s - %s\n", parent.Level, parent.Title, parent.Description)
		}
	}

	// 构建同级需求信息
	siblingInfo := ""
	if len(context.SiblingRequirements) > 0 {
		siblingInfo = "相关同级需求：\n"
		for _, sibling := range context.SiblingRequirements {
			siblingInfo += fmt.Sprintf("- %s: %s\n", sibling.Title, sibling.Description)
		}
	}

	// 构建完整的分析prompt
	prompt := fmt.Sprintf(`你是一位国际认证的NESMA功能点分析专家，拥有20年以上的软件度量经验。你的专业资质包括：

## 核心专业能力
- **NESMA 2.2国际标准专家**：精通NESMA功能点分析的所有细节、规则和最佳实践
- **ISO/IEC 14143标准认证**：具备国际软件度量标准认证资质
- **软件度量专家**：在功能点测量、软件度量、质量评估方面有深厚造诣
- **需求工程专家**：擅长需求获取、分析、建模、验证和管理的全流程
- **多行业项目经验**：在金融、制造、医疗、政府、电商等15+个行业有丰富的项目实践

## 专业技能矩阵
- **精确功能点识别**：准确识别EI/EO/EQ/ILF/EIF五种功能类型，准确率>95%%
- **科学复杂度评估**：基于DET/RET/FTR的科学评估方法，考虑技术、业务、数据三个维度
- **功能点精确计算**：严格按照NESMA标准计算AFP和UFP值
- **质量风险评估**：从完整性、一致性、可追溯性、可测试性等多维度评估需求质量
- **优化建议提供**：基于行业最佳实践，提供具体可操作的改进方案

## 分析方法论
- **标准合规性**：严格按照NESMA 2.2标准进行分析和计算
- **系统性思维**：从项目整体角度分析需求的价值和影响
- **数据驱动决策**：基于历史数据和行业基准进行科学评估
- **价值导向思维**：始终关注业务价值和投资回报率

请基于以下完整的项目背景对指定需求进行智能分析优化。

## 项目背景信息
%s

## 建设周期信息
%s

## 需求层级背景
%s

## 相关需求参考
%s

## 待分析需求
需求标题：%s
需求描述：%s
当前层级：%d级
当前功能类型：%s
当前复杂度：%s

## 分析要求
请基于上述完整背景，对该需求进行专业的NESMA分析优化。

**重要：请严格按照以下JSON格式回复，不要添加任何markdown标记或额外文本：**

{
  "optimized_title": "优化后的需求标题",
  "optimized_description": "详细功能描述，描述信息要符合人类自然语言，涵盖到功能点所涉及业务场景，并描述清楚功能点所涉及的业务逻辑，要详细但不冗长，不要写输入、处理和输出，将其作为自然语言描述",
  "function_type": "EI/EO/EQ/ILF/EIF中最合适的类型",
  "complexity_level": "Low/Average/High",
  "business_value": "业务价值说明",
  "recommended_afp": 推荐的AFP值,
  "recommended_ufp": 推荐的UFP值,
  "analysis_notes": "分析说明和改进建议",
  "confidence_score": 0-1的置信度评分,
  "improvement_points": ["改进点1", "改进点2", "改进点3"]
}

## 分析重点
1. 充分考虑项目背景和业务领域特征
2. 参考上级需求的整体架构设计
3. 与同级需求保持一致性和完整性
4. 确保功能类型分类的准确性
5. 提供具体可行的优化建议`,
		projectInfo, cycleInfo, parentInfo, siblingInfo,
		requirement.Title, requirement.Description, requirement.Level, 
		requirement.FunctionType, requirement.Complexity)

	return prompt
}

// callAIForSingleRequirement 调用AI分析单个需求
func (s *SingleRequirementAnalysisService) callAIForSingleRequirement(prompt string) (*AIResponse, error) {
	// 初始化AI服务
	aiService := NewDeepSeekService(
		global.GVA_CONFIG.AI.DeepSeek.APIKey,
		global.GVA_CONFIG.AI.DeepSeek.BaseURL,
		global.GVA_CONFIG.AI.DeepSeek.Model,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Second)
	defer cancel()

	config := &AIConfig{
		MaxTokens:   32000,
		Temperature: 0.7,
		TopP:        0.95,
		Model:       "deepseek-reasoner",
		Stream:      false,
	}

	return aiService.GenerateText(ctx, prompt, config)
}

// buildAnalysisResult 构建分析结果
func (s *SingleRequirementAnalysisService) buildAnalysisResult(requirement *nesma.NesmaRequirement, aiResponse *AIResponse) *SingleAnalysisResult {
	result := &SingleAnalysisResult{
		RequirementID: requirement.ID,
		OriginalInfo: &AnalysisRequirementInfo{
			Title:          requirement.Title,
			Description:    requirement.Description,
			FunctionType:   requirement.FunctionType,
			Complexity:     requirement.Complexity,
			BusinessValue:  requirement.BusinessValue,
			RecommendedAFP: &requirement.AFP,
			RecommendedUFP: &requirement.UFP,
		},
		Timestamp: time.Now(),
	}

	// 解析AI响应
	if len(aiResponse.Choices) > 0 {
		responseText := aiResponse.Choices[0].Message.Content
		
		// 尝试从markdown中提取JSON
		jsonData := s.extractJSONFromMarkdown(responseText)
		
		var analysisData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonData), &analysisData); err == nil {
			// 解析优化信息
			optimizedInfo := &AnalysisRequirementInfo{
				Title:         getStringValue(analysisData, "optimized_title", requirement.Title),
				Description:   getStringValue(analysisData, "optimized_description", requirement.Description),
				FunctionType:  getStringValue(analysisData, "function_type", requirement.FunctionType),
				Complexity:    getStringValue(analysisData, "complexity_level", requirement.Complexity),
				BusinessValue: getStringValue(analysisData, "business_value", requirement.BusinessValue),
			}

			if afp, ok := analysisData["recommended_afp"].(float64); ok {
				optimizedInfo.RecommendedAFP = &afp
			}
			if ufp, ok := analysisData["recommended_ufp"].(float64); ok {
				optimizedInfo.RecommendedUFP = &ufp
			}

			result.OptimizedInfo = optimizedInfo
			result.AnalysisNote = getStringValue(analysisData, "analysis_notes", "AI分析完成")
			if confidence, ok := analysisData["confidence_score"].(float64); ok {
				result.Confidence = confidence
			} else {
				result.Confidence = 0.8
			}
			
			global.GVA_LOG.Info("AI响应解析成功", 
				zap.String("title", optimizedInfo.Title),
				zap.String("functionType", optimizedInfo.FunctionType),
				zap.Float64("confidence", result.Confidence))
		} else {
			// 解析失败时的降级处理
			global.GVA_LOG.Warn("AI响应JSON解析失败", zap.Error(err), zap.String("extractedJSON", jsonData))
			result.OptimizedInfo = result.OriginalInfo
			result.AnalysisNote = fmt.Sprintf("AI响应解析失败，原始响应：%s", responseText)
			result.Confidence = 0.5
		}
	} else {
		result.OptimizedInfo = result.OriginalInfo
		result.AnalysisNote = "AI响应为空"
		result.Confidence = 0.3
	}

	return result
}

// extractJSONFromMarkdown 从markdown格式中提取JSON
func (s *SingleRequirementAnalysisService) extractJSONFromMarkdown(text string) string {
	// 匹配 ```json 和 ``` 之间的内容
	re := regexp.MustCompile("(?s)```json\\s*(.+?)\\s*```")
	matches := re.FindStringSubmatch(text)
	
	if len(matches) > 1 {
		jsonText := strings.TrimSpace(matches[1])
		global.GVA_LOG.Info("从markdown中提取JSON成功", zap.String("jsonLength", fmt.Sprintf("%d", len(jsonText))))
		return jsonText
	}
	
	// 如果没有找到markdown格式，尝试匹配纯JSON（以{开头}结尾）
	re2 := regexp.MustCompile("(?s)\\{.+\\}")
	matches2 := re2.FindStringSubmatch(text)
	
	if len(matches2) > 0 {
		jsonText := strings.TrimSpace(matches2[0])
		global.GVA_LOG.Info("提取纯JSON成功", zap.String("jsonLength", fmt.Sprintf("%d", len(jsonText))))
		return jsonText
	}
	
	global.GVA_LOG.Warn("未能从响应中提取JSON", zap.String("originalText", text))
	return text
}

// getStringValue 安全获取字符串值
func getStringValue(data map[string]interface{}, key, defaultValue string) string {
	if value, ok := data[key].(string); ok && value != "" {
		return value
	}
	return defaultValue
}

// AnalyzeHierarchicalRequirementAsync 异步层次化需求分析
func (s *SingleRequirementAnalysisService) AnalyzeHierarchicalRequirementAsync(req *HierarchicalAnalysisRequest) (*HierarchicalAnalysisTask, error) {
	global.GVA_LOG.Info("开始创建层次化需求分析任务", 
		zap.Uint("requirementId", req.RequirementID),
		zap.String("analysisType", req.AnalysisType))

	// 验证需求存在性和层级
	requirement, err := s.validateRequirementForHierarchicalAnalysis(req.RequirementID)
	if err != nil {
		return nil, err
	}

	// 创建层次化分析任务
	taskIDCounter++
	task := &HierarchicalAnalysisTask{
		ID:            taskIDCounter,
		RequirementID: req.RequirementID,
		ProjectID:     req.ProjectID,
		CycleID:       req.CycleID,
		AnalysisType:  req.AnalysisType,
		Status:        "pending",
		Progress:      0,
		StartTime:     time.Now(),
		CreatedAt:     time.Now(),
		Config: map[string]interface{}{
			"autoGenerateL4":       req.AutoGenerateL4,
			"includeKnowledgeBase": req.IncludeKnowledgeBase,
			"maxL4Count":           req.MaxL4Count,
			"l4GenerationStrategy": req.L4GenerationStrategy,
		},
	}

	// 存储任务
	hierarchicalTaskStore[task.ID] = task

	// 启动异步分析
	go s.runHierarchicalAnalysisTask(task, req, requirement)

	global.GVA_LOG.Info("层次化需求分析任务已创建", 
		zap.Uint("taskId", task.ID), 
		zap.Uint("requirementId", req.RequirementID))
	return task, nil
}

// runHierarchicalAnalysisTask 执行层次化分析任务
func (s *SingleRequirementAnalysisService) runHierarchicalAnalysisTask(task *HierarchicalAnalysisTask, req *HierarchicalAnalysisRequest, requirement *nesma.NesmaRequirement) {
	global.GVA_LOG.Info("开始执行层次化需求分析", 
		zap.Uint("taskId", task.ID), 
		zap.Uint("requirementId", task.RequirementID),
		zap.String("analysisType", req.AnalysisType))

	// 根据分析类型选择不同的处理流程
	switch req.AnalysisType {
	case "generate_l4":
		s.runL4GenerationAnalysis(task, req, requirement)
	case "optimize_l4":
		s.runL4OptimizationAnalysis(task, req, requirement)
	default:
		s.runEnhancedSingleAnalysis(task, req, requirement)
	}
}

// runL4GenerationAnalysis 运行L4生成分析
func (s *SingleRequirementAnalysisService) runL4GenerationAnalysis(task *HierarchicalAnalysisTask, req *HierarchicalAnalysisRequest, requirement *nesma.NesmaRequirement) {
	// 阶段1：验证L3需求
	s.updateHierarchicalTaskStep(task, "validating_l3", "正在验证L3需求信息...", 10, "scan")
	
	if requirement.Level != 3 {
		s.setHierarchicalTaskFailed(task, "只有L3需求才能生成L4功能点")
		return
	}

	// 阶段2：分析现有L4结构
	s.updateHierarchicalTaskStep(task, "analyzing_existing_l4", "正在分析现有L4功能点结构...", 20, "search")
	
	existingL4s, err := s.getExistingL4Requirements(requirement.ID)
	if err != nil {
		global.GVA_LOG.Error("获取现有L4失败", zap.Error(err))
		existingL4s = []nesma.NesmaRequirement{}
	}

	// 确定生成策略
	strategy := s.determineL4GenerationStrategy(existingL4s, req.L4GenerationStrategy)
	
	// 阶段3：构建项目上下文
	s.updateHierarchicalTaskStep(task, "building_context", "正在构建项目背景和需求上下文...", 30, "connection")
	
	context, err := s.buildEnhancedAnalysisContext(requirement, req.ProjectID, req.CycleID)
	if err != nil {
		global.GVA_LOG.Error("构建分析上下文失败", zap.Error(err))
		context = &AnalysisContext{}
	}

	// 阶段4：知识库搜索
	var knowledgeRefs []KnowledgeReference
	if req.IncludeKnowledgeBase {
		s.updateHierarchicalTaskStep(task, "searching_knowledge", "正在搜索相关知识库内容...", 40, "search")
		knowledgeRefs, err = s.searchL4KnowledgeBase(requirement, context)
		if err != nil {
			global.GVA_LOG.Error("知识库搜索失败", zap.Error(err))
			knowledgeRefs = []KnowledgeReference{}
		}
	}

	// 阶段5：AI生成L4功能点
	s.updateHierarchicalTaskStep(task, "generating_l4", "AI正在生成L4功能点...", 60, "cpu")
	
	l4Result, err := s.generateL4WithAI(requirement, existingL4s, context, knowledgeRefs, strategy, req.MaxL4Count)
	if err != nil {
		s.setHierarchicalTaskFailed(task, fmt.Sprintf("L4生成失败: %v", err))
		return
	}

	// 阶段6：分析L3需求本身（如果需要）
	s.updateHierarchicalTaskStep(task, "analyzing_l3", "正在优化L3需求本身...", 80, "edit")
	
	l3Analysis, err := s.analyzeL3Requirement(requirement, context, knowledgeRefs)
	if err != nil {
		global.GVA_LOG.Error("L3分析失败", zap.Error(err))
		// 非致命错误，继续处理
	}

	// 阶段7：整合结果
	s.updateHierarchicalTaskStep(task, "integrating_results", "正在整合分析结果...", 90, "check")
	
	result := s.buildHierarchicalAnalysisResult(requirement, l3Analysis, l4Result, knowledgeRefs, strategy)

	// 完成任务
	task.Status = "completed"
	task.Progress = 100
	task.CurrentStep = "completed"
	task.StepDesc = "层次化分析完成！已生成L4功能点建议"
	task.AnimationType = "success"
	task.Result = result
	endTime := time.Now()
	task.EndTime = &endTime

	s.completeCurrentHierarchicalStep(task)
	
	global.GVA_LOG.Info("L4生成分析完成", 
		zap.Uint("taskId", task.ID), 
		zap.Int("generatedL4Count", len(l4Result.L4Suggestions)))
}

// runL4OptimizationAnalysis 运行L4优化分析
func (s *SingleRequirementAnalysisService) runL4OptimizationAnalysis(task *HierarchicalAnalysisTask, req *HierarchicalAnalysisRequest, requirement *nesma.NesmaRequirement) {
	// 阶段1：获取现有L4
	s.updateHierarchicalTaskStep(task, "fetching_existing_l4", "正在获取现有L4功能点...", 15, "download")
	
	existingL4s, err := s.getExistingL4Requirements(requirement.ID)
	if err != nil || len(existingL4s) == 0 {
		s.setHierarchicalTaskFailed(task, "没有找到可优化的L4功能点")
		return
	}

	// 阶段2：构建上下文
	s.updateHierarchicalTaskStep(task, "building_context", "正在构建优化上下文...", 25, "connection")
	
	analysisContext, err := s.buildEnhancedAnalysisContext(requirement, req.ProjectID, req.CycleID)
	if err != nil {
		global.GVA_LOG.Error("构建分析上下文失败", zap.Error(err))
		analysisContext = &AnalysisContext{}
	}

	// 阶段3：逐个优化L4
	s.updateHierarchicalTaskStep(task, "optimizing_l4s", "正在逐个优化L4功能点...", 40, "edit")
	
	optimizedL4s := []L4Suggestion{}
	for i, l4 := range existingL4s {
		progress := 40 + int(float64(i)/float64(len(existingL4s))*40)
		s.updateHierarchicalTaskStep(task, "optimizing_l4s", 
			fmt.Sprintf("正在优化L4: %s (%d/%d)", l4.Title, i+1, len(existingL4s)), 
			progress, "edit")
		
		optimized, err := s.optimizeSingleL4(l4, requirement, analysisContext)
		if err != nil {
			global.GVA_LOG.Error("L4优化失败", zap.Error(err), zap.String("l4Title", l4.Title))
			continue
		}
		optimizedL4s = append(optimizedL4s, *optimized)
	}

	// 阶段4：整合结果
	s.updateHierarchicalTaskStep(task, "integrating_results", "正在整合优化结果...", 90, "check")
	
	l4Result := &L4GenerationSummary{
		Strategy:         "optimize_existing",
		ExistingL4Count:  len(existingL4s),
		OptimizedL4Count: len(optimizedL4s),
		L4Suggestions:    optimizedL4s,
		GenerationNotes:  fmt.Sprintf("已优化 %d 个现有L4功能点", len(optimizedL4s)),
		RecommendedActions: []string{
			"建议逐个审核优化结果",
			"可根据需要调整优化内容",
			"确认功能点分类和复杂度评估",
		},
	}

	result := s.buildHierarchicalAnalysisResult(requirement, nil, l4Result, []KnowledgeReference{}, "optimize_existing")

	// 完成任务
	task.Status = "completed"
	task.Progress = 100
	task.CurrentStep = "completed"
	task.StepDesc = "L4优化完成！已优化现有功能点"
	task.AnimationType = "success"
	task.Result = result
	endTime := time.Now()
	task.EndTime = &endTime

	s.completeCurrentHierarchicalStep(task)
	
	global.GVA_LOG.Info("L4优化分析完成", 
		zap.Uint("taskId", task.ID), 
		zap.Int("optimizedL4Count", len(optimizedL4s)))
}

// runEnhancedSingleAnalysis 运行增强的单需求分析
func (s *SingleRequirementAnalysisService) runEnhancedSingleAnalysis(task *HierarchicalAnalysisTask, req *HierarchicalAnalysisRequest, requirement *nesma.NesmaRequirement) {
	// 使用现有的单需求分析逻辑，但支持L4生成
	s.updateHierarchicalTaskStep(task, "initializing", "正在初始化增强分析...", 5, "pulse")
	time.Sleep(500 * time.Millisecond)
	
	s.updateHierarchicalTaskStep(task, "validating", "正在验证需求信息...", 15, "scan")
	
	s.updateHierarchicalTaskStep(task, "context_building", "正在收集项目背景和上下文信息...", 35, "network")
	context, err := s.buildEnhancedAnalysisContext(requirement, req.ProjectID, req.CycleID)
	if err != nil {
		global.GVA_LOG.Error("构建分析上下文失败", zap.Error(err))
		context = &AnalysisContext{}
	}

	s.updateHierarchicalTaskStep(task, "ai_analyzing", "AI正在进行深度需求分析...", 50, "brain")
	
	prompt := s.buildEnhancedAnalysisPrompt(requirement, context, req.AutoGenerateL4)
	
	s.updateHierarchicalTaskStep(task, "ai_calling", "正在调用AI分析服务...", 70, "api")
	aiResult, err := s.callAIForSingleRequirement(prompt)
	if err != nil {
		s.setHierarchicalTaskFailed(task, fmt.Sprintf("AI分析失败: %v", err))
		return
	}

	s.updateHierarchicalTaskStep(task, "parsing", "正在解析AI分析结果...", 85, "parse")
	singleResult := s.buildAnalysisResult(requirement, aiResult)
	
	// 转换为层次化结果格式
	result := &HierarchicalAnalysisResult{
		RequirementID: requirement.ID,
		AnalysisType:  "single",
		OriginalInfo:  singleResult.OriginalInfo,
		OptimizedInfo: singleResult.OptimizedInfo,
		AnalysisNote:  singleResult.AnalysisNote,
		Confidence:    singleResult.Confidence,
		Timestamp:     singleResult.Timestamp,
	}

	s.updateHierarchicalTaskStep(task, "finalizing", "正在整理分析结果...", 95, "check")
	time.Sleep(300 * time.Millisecond)
	
	// 任务完成
	task.Status = "completed"
	task.Progress = 100
	task.CurrentStep = "completed"
	task.StepDesc = "增强分析完成！AI已为您优化需求"
	task.AnimationType = "success"
	task.Result = result
	endTime := time.Now()
	task.EndTime = &endTime
	
	s.completeCurrentHierarchicalStep(task)

	global.GVA_LOG.Info("增强单需求分析完成", zap.Uint("taskId", task.ID))
}

// 层次化任务存储
var hierarchicalTaskStore = make(map[uint]*HierarchicalAnalysisTask)

// GetHierarchicalAnalysisTask 获取层次化分析任务
func (s *SingleRequirementAnalysisService) GetHierarchicalAnalysisTask(taskID uint) (*HierarchicalAnalysisTask, error) {
	task, exists := hierarchicalTaskStore[taskID]
	if !exists {
		return nil, fmt.Errorf("层次化分析任务不存在")
	}
	return task, nil
}

// validateRequirementForHierarchicalAnalysis 验证需求是否可进行层次化分析
func (s *SingleRequirementAnalysisService) validateRequirementForHierarchicalAnalysis(requirementID uint) (*nesma.NesmaRequirement, error) {
	var requirement nesma.NesmaRequirement
	err := global.GVA_DB.First(&requirement, requirementID).Error
	if err != nil {
		return nil, fmt.Errorf("需求不存在: %v", err)
	}
	return &requirement, nil
}

// updateHierarchicalTaskStep 更新层次化任务步骤
func (s *SingleRequirementAnalysisService) updateHierarchicalTaskStep(task *HierarchicalAnalysisTask, step, description string, progress int, animationType string) {
	// 完成当前步骤
	s.completeCurrentHierarchicalStep(task)
	
	// 开始新步骤
	now := time.Now()
	newStep := ProcessingStep{
		Step:        step,
		Description: description,
		Status:      "running",
		StartTime:   now,
		Details:     description,
	}
	
	task.CurrentStep = step
	task.StepDesc = description
	task.Progress = progress
	task.AnimationType = animationType
	task.Status = "running"
	task.ProcessingSteps = append(task.ProcessingSteps, newStep)
	
	global.GVA_LOG.Info("层次化任务步骤更新", 
		zap.Uint("taskId", task.ID),
		zap.String("step", step),
		zap.String("description", description),
		zap.Int("progress", progress))
}

// completeCurrentHierarchicalStep 完成当前层次化步骤
func (s *SingleRequirementAnalysisService) completeCurrentHierarchicalStep(task *HierarchicalAnalysisTask) {
	if len(task.ProcessingSteps) > 0 {
		lastStep := &task.ProcessingSteps[len(task.ProcessingSteps)-1]
		if lastStep.Status == "running" {
			endTime := time.Now()
			lastStep.EndTime = &endTime
			lastStep.Status = "completed"
		}
	}
}

// setHierarchicalTaskFailed 设置层次化任务失败
func (s *SingleRequirementAnalysisService) setHierarchicalTaskFailed(task *HierarchicalAnalysisTask, errorMsg string) {
	task.Status = "failed"
	task.ErrorMsg = errorMsg
	task.CurrentStep = "failed"
	task.StepDesc = "分析失败: " + errorMsg
	task.AnimationType = "error"
	task.Progress = 0
	endTime := time.Now()
	task.EndTime = &endTime
	
	// 完成当前步骤为失败
	if len(task.ProcessingSteps) > 0 {
		lastStep := &task.ProcessingSteps[len(task.ProcessingSteps)-1]
		if lastStep.Status == "running" {
			endTime := time.Now()
			lastStep.EndTime = &endTime
			lastStep.Status = "failed"
			lastStep.ErrorMsg = errorMsg
		}
	}
	
	global.GVA_LOG.Error("层次化分析失败", zap.Error(fmt.Errorf(errorMsg)), zap.Uint("taskId", task.ID))
}

// getExistingL4Requirements 获取现有的L4需求
func (s *SingleRequirementAnalysisService) getExistingL4Requirements(parentID uint) ([]nesma.NesmaRequirement, error) {
	var l4s []nesma.NesmaRequirement
	err := global.GVA_DB.Where("parent_id = ? AND level = 4", parentID).Find(&l4s).Error
	return l4s, err
}

// determineL4GenerationStrategy 确定L4生成策略
func (s *SingleRequirementAnalysisService) determineL4GenerationStrategy(existingL4s []nesma.NesmaRequirement, requestedStrategy string) string {
	if requestedStrategy != "" {
		return requestedStrategy
	}
	
	if len(existingL4s) == 0 {
		return "create_new"
	} else if len(existingL4s) < 5 {
		return "expand"
	} else {
		return "optimize_existing"
	}
}

// buildEnhancedAnalysisContext 构建增强的分析上下文
func (s *SingleRequirementAnalysisService) buildEnhancedAnalysisContext(requirement *nesma.NesmaRequirement, projectID, cycleID uint) (*AnalysisContext, error) {
	// 复用现有的上下文构建逻辑
	return s.buildAnalysisContext(requirement, projectID, cycleID)
}

// searchL4KnowledgeBase 搜索L4相关知识库
func (s *SingleRequirementAnalysisService) searchL4KnowledgeBase(requirement *nesma.NesmaRequirement, context *AnalysisContext) ([]KnowledgeReference, error) {
	// 构建搜索查询
	searchQuery := fmt.Sprintf("%s %s L4功能点 四级", requirement.Title, requirement.Description)
	if context.Project != nil {
		searchQuery += " " + context.Project.Domain
	}
	
	// 模拟知识库搜索（实际应该集成真实的知识库服务）
	knowledgeRefs := []KnowledgeReference{
		{
			ID:        1,
			Title:     "NESMA L4功能点识别指南",
			Category:  "NESMA_STANDARD",
			Relevance: 0.9,
		},
		{
			ID:        2,
			Title:     "四级功能点分解最佳实践",
			Category:  "BEST_PRACTICE",
			Relevance: 0.8,
		},
	}
	
	return knowledgeRefs, nil
}

// generateL4WithAI 使用AI生成L4功能点
func (s *SingleRequirementAnalysisService) generateL4WithAI(requirement *nesma.NesmaRequirement, existingL4s []nesma.NesmaRequirement, analysisContext *AnalysisContext, knowledgeRefs []KnowledgeReference, strategy string, maxCount int) (*L4GenerationSummary, error) {
	// 构建L4生成的prompt
	prompt := s.buildL4GenerationPrompt(requirement, existingL4s, analysisContext, knowledgeRefs, strategy, maxCount)
	
	// 调用AI服务
	aiService := NewDeepSeekService(
		global.GVA_CONFIG.AI.DeepSeek.APIKey,
		global.GVA_CONFIG.AI.DeepSeek.BaseURL,
		global.GVA_CONFIG.AI.DeepSeek.Model,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second) // 增加超时时间
	defer cancel()

	config := &AIConfig{
		MaxTokens:   3000,
		Temperature: 0.8,
		TopP:        0.9,
		Model:       "deepseek-chat",
		Stream:      false,
	}

	response, err := aiService.GenerateText(ctx, prompt, config)
	if err != nil {
		return nil, fmt.Errorf("AI L4生成失败: %w", err)
	}

	// 解析L4建议
	l4Suggestions, err := s.parseL4SuggestionsFromAI(response.Text)
	if err != nil {
		global.GVA_LOG.Error("解析L4建议失败", zap.Error(err))
		l4Suggestions = s.createFallbackL4Suggestions(strategy)
	}

	// 构建生成摘要
	result := &L4GenerationSummary{
		Strategy:         strategy,
		ExistingL4Count:  len(existingL4s),
		GeneratedL4Count: len(l4Suggestions),
		L4Suggestions:    l4Suggestions,
		GenerationNotes:  fmt.Sprintf("基于%s策略生成了%d个L4功能点", strategy, len(l4Suggestions)),
		RecommendedActions: []string{
			"建议逐个审核生成的L4功能点",
			"可根据实际需要调整功能描述",
			"确认功能类型分类的准确性",
			"验证功能点数估算的合理性",
		},
	}

	return result, nil
}

// buildL4GenerationPrompt 构建L4生成的prompt
func (s *SingleRequirementAnalysisService) buildL4GenerationPrompt(requirement *nesma.NesmaRequirement, existingL4s []nesma.NesmaRequirement, analysisContext *AnalysisContext, knowledgeRefs []KnowledgeReference, strategy string, maxCount int) string {
	var knowledgeContext strings.Builder
	for _, ref := range knowledgeRefs {
		knowledgeContext.WriteString(fmt.Sprintf("- %s: %s\n", ref.Title, ref.Category))
	}

	var existingContext strings.Builder
	for _, existing := range existingL4s {
		existingContext.WriteString(fmt.Sprintf("- %s: %s\n", existing.Title, existing.Description))
	}

	var projectContext string
	if analysisContext.Project != nil {
		projectContext = fmt.Sprintf("项目：%s\n业务领域：%s\n", analysisContext.Project.Name, analysisContext.Project.Domain)
	}

	strategyDesc := map[string]string{
		"create_new":        "基于L3需求全新创建L4功能点",
		"expand":           "在现有L4基础上扩展补充新的功能点",
		"optimize_existing": "优化现有L4功能点的描述和分类",
	}

	return fmt.Sprintf(`你是NESMA功能点分析专家，请基于L3需求生成详细的L4功能点建议。

%s

【L3需求信息】
标题: %s
描述: %s
层级: L%d
功能类型: %s

【现有L4功能点】
%s

【生成策略】
%s: %s

【相关知识库】
%s

【生成要求】
1. 生成%d个以内的L4功能点
2. 每个L4功能点应该是独立的、可计数的功能单元
3. 确保L4功能点覆盖L3需求的核心业务场景
4. 按照NESMA标准进行功能类型分类（EI/EO/EQ/ILF/EIF）
5. 提供合理的复杂度评估和功能点数推荐

【输出格式】
请以JSON格式返回：
{
  "l4_suggestions": [
    {
      "suggestedTitle": "L4功能点标题",
      "suggestedDescription": "详细功能描述",
      "suggestedCode": "功能编码",
      "functionType": "EI/EO/EQ/ILF/EIF",
      "businessValue": "业务价值说明",
      "acceptanceCriteria": "验收标准",
      "estimatedComplexity": "简单/中等/复杂",
      "recommendedAFP": AFP值,
      "recommendedUFP": UFP值,
      "priority": 1-5,
      "confidence": 0.0-1.0,
      "generationReason": "生成理由",
      "relatedKnowledge": "相关知识",
      "actionType": "%s"
    }
  ]
}`, 
	projectContext,
	requirement.Title, requirement.Description, requirement.Level, requirement.FunctionType,
	existingContext.String(),
	strategy, strategyDesc[strategy],
	knowledgeContext.String(),
	maxCount,
	strategy)
}

// parseL4SuggestionsFromAI 从AI响应中解析L4建议
func (s *SingleRequirementAnalysisService) parseL4SuggestionsFromAI(response string) ([]L4Suggestion, error) {
	var result struct {
		L4Suggestions []L4Suggestion `json:"l4_suggestions"`
	}

	// 尝试直接解析JSON
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果解析失败，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonContent := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
				return nil, fmt.Errorf("解析L4建议JSON失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("无法从响应中提取JSON")
		}
	}

	return result.L4Suggestions, nil
}

// createFallbackL4Suggestions 创建备用L4建议
func (s *SingleRequirementAnalysisService) createFallbackL4Suggestions(strategy string) []L4Suggestion {
	return []L4Suggestion{
		{
			SuggestedTitle:       "基础数据操作功能",
			SuggestedDescription: "提供基础的数据增删改查操作",
			SuggestedCode:        "L4_001",
			FunctionType:         "EI",
			BusinessValue:        "满足基本的数据管理需求",
			AcceptanceCriteria:   "能够完成基本的数据操作",
			EstimatedComplexity:  "简单",
			RecommendedAFP:       3.0,
			RecommendedUFP:       3.0,
			Priority:             3,
			Confidence:           0.6,
			GenerationReason:     "AI响应解析失败，使用默认功能点",
			RelatedKnowledge:     "通用数据操作模板",
			ActionType:           strategy,
		},
	}
}

// analyzeL3Requirement 分析L3需求本身
func (s *SingleRequirementAnalysisService) analyzeL3Requirement(requirement *nesma.NesmaRequirement, context *AnalysisContext, knowledgeRefs []KnowledgeReference) (*SingleAnalysisResult, error) {
	// 构建L3分析的prompt
	prompt := s.buildContextualAnalysisPrompt(requirement, context)
	
	// 调用AI分析
	aiResult, err := s.callAIForSingleRequirement(prompt)
	if err != nil {
		return nil, err
	}
	
	// 构建分析结果
	result := s.buildAnalysisResult(requirement, aiResult)
	return result, nil
}

// buildHierarchicalAnalysisResult 构建层次化分析结果
func (s *SingleRequirementAnalysisService) buildHierarchicalAnalysisResult(requirement *nesma.NesmaRequirement, l3Analysis *SingleAnalysisResult, l4Result *L4GenerationSummary, knowledgeRefs []KnowledgeReference, strategy string) *HierarchicalAnalysisResult {
	result := &HierarchicalAnalysisResult{
		RequirementID:       requirement.ID,
		AnalysisType:        "generate_l4",
		Timestamp:           time.Now(),
		L4GenerationResult:  l4Result,
		KnowledgeReferences: knowledgeRefs,
	}

	// 如果有L3分析结果，使用它
	if l3Analysis != nil {
		result.OriginalInfo = l3Analysis.OriginalInfo
		result.OptimizedInfo = l3Analysis.OptimizedInfo
		result.AnalysisNote = l3Analysis.AnalysisNote
		result.Confidence = l3Analysis.Confidence
	} else {
		// 否则使用原始需求信息
		result.OriginalInfo = &AnalysisRequirementInfo{
			Title:          requirement.Title,
			Description:    requirement.Description,
			FunctionType:   requirement.FunctionType,
			Complexity:     requirement.Complexity,
			BusinessValue:  requirement.BusinessValue,
			RecommendedAFP: &requirement.AFP,
			RecommendedUFP: &requirement.UFP,
		}
		result.OptimizedInfo = result.OriginalInfo
		result.AnalysisNote = fmt.Sprintf("基于%s策略生成了L4功能点建议", strategy)
		result.Confidence = 0.8
	}

	return result
}

// optimizeSingleL4 优化单个L4功能点
func (s *SingleRequirementAnalysisService) optimizeSingleL4(l4 nesma.NesmaRequirement, parentL3 *nesma.NesmaRequirement, analysisContext *AnalysisContext) (*L4Suggestion, error) {
	prompt := fmt.Sprintf(`作为NESMA专家，请优化以下L4功能点：

【父级L3需求】
标题: %s
描述: %s

【当前L4功能点】
标题: %s
描述: %s
功能类型: %s
复杂度: %s

【优化要求】
1. 保持L4的独立性和可计数性
2. 完善功能描述和业务价值
3. 准确的NESMA功能类型分类
4. 合理的复杂度和功能点数估算

请以JSON格式返回优化建议：
{
  "suggestedTitle": "优化后的标题",
  "suggestedDescription": "优化后的详细描述",
  "functionType": "EI/EO/EQ/ILF/EIF",
  "businessValue": "业务价值说明",
  "estimatedComplexity": "简单/中等/复杂",
  "recommendedAFP": AFP值,
  "recommendedUFP": UFP值,
  "confidence": 0.0-1.0,
  "generationReason": "优化理由"
}`, 
	parentL3.Title, parentL3.Description,
	l4.Title, l4.Description, l4.FunctionType, l4.Complexity)

	// 调用AI
	aiService := NewDeepSeekService(
		global.GVA_CONFIG.AI.DeepSeek.APIKey,
		global.GVA_CONFIG.AI.DeepSeek.BaseURL,
		global.GVA_CONFIG.AI.DeepSeek.Model,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	config := &AIConfig{
		MaxTokens:   1500,
		Temperature: 0.7,
		TopP:        0.9,
		Model:       "deepseek-chat",
		Stream:      false,
	}

	response, err := aiService.GenerateText(ctx, prompt, config)
	if err != nil {
		return nil, fmt.Errorf("L4优化AI调用失败: %w", err)
	}

	// 解析响应
	var optimization L4Suggestion
	if err := json.Unmarshal([]byte(response.Text), &optimization); err != nil {
		// 解析失败时的降级处理
		optimization = L4Suggestion{
			SuggestedTitle:       l4.Title,
			SuggestedDescription: l4.Description,
			FunctionType:         l4.FunctionType,
			EstimatedComplexity:  l4.Complexity,
			BusinessValue:        "优化中...",
			RecommendedAFP:       l4.AFP,
			RecommendedUFP:       l4.UFP,
			Confidence:           0.6,
			GenerationReason:     "AI响应解析失败，保持原有内容",
			ExistingL4ID:         &l4.ID,
			ActionType:           "optimize_existing",
		}
	}

	optimization.ExistingL4ID = &l4.ID
	optimization.ActionType = "optimize_existing"

	return &optimization, nil
}

// buildEnhancedAnalysisPrompt 构建增强的分析prompt
func (s *SingleRequirementAnalysisService) buildEnhancedAnalysisPrompt(requirement *nesma.NesmaRequirement, context *AnalysisContext, includeL4Generation bool) string {
	basePrompt := s.buildContextualAnalysisPrompt(requirement, context)
	
	if !includeL4Generation {
		return basePrompt
	}

	// 为支持L4生成的增强提示
	enhancedPrompt := basePrompt + `

【L4生成支持】
如果当前需求是L3级别，请额外提供L4功能点生成建议：

{
  "optimized_title": "优化后的需求标题",
  "optimized_description": "详细的功能描述",
  "function_type": "EI/EO/EQ/ILF/EIF",
  "complexity_level": "Low/Average/High",
  "business_value": "业务价值说明",
  "recommended_afp": 推荐的AFP值,
  "recommended_ufp": 推荐的UFP值,
  "analysis_notes": "分析说明",
  "confidence_score": 0-1的置信度,
  "l4_generation_hints": [
    "L4功能点生成提示1",
    "L4功能点生成提示2"
  ]
}`

	return enhancedPrompt
}
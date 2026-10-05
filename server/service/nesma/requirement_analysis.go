package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type RequirementAnalysisService struct{}

// StartAnalysis 启动需求分析
func (s *RequirementAnalysisService) StartAnalysis(req *nesmaReq.NesmaAnalyzeRequest) (*nesma.NesmaRequirementAnalysisTask, error) {
	// 1. 验证项目和周期
	var cycle nesma.NesmaProjectCycle
	if err := global.GVA_DB.Preload("Project").First(&cycle, req.CycleID).Error; err != nil {
		return nil, errors.New("项目周期不存在")
	}
	
	if cycle.ProjectID != req.ProjectID {
		return nil, errors.New("项目周期不匹配")
	}

	// 2. 获取当前周期的激活版本或创建新版本
	sourceVersion, err := s.getOrCreateActiveVersion(req.CycleID)
	if err != nil {
		return nil, fmt.Errorf("获取源版本失败: %v", err)
	}

	// 3. 创建分析任务
	configJson, _ := json.Marshal(s.buildTaskConfig(req))
	task := &nesma.NesmaRequirementAnalysisTask{
		ProjectID:       req.ProjectID,
		CycleID:         req.CycleID,
		SourceVersionID: sourceVersion.ID,
		TaskType:        "requirement_analysis",
		Status:          "pending",
		Priority:        5,
		Config:          configJson,
	}

	if len(req.RequirementIDs) > 0 {
		requirementIDsJson, _ := json.Marshal(req.RequirementIDs)
		task.RequirementIDs = requirementIDsJson
	}

	if err := global.GVA_DB.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建分析任务失败: %v", err)
	}

	// 4. 启动异步分析任务
	go s.runAnalysisTask(task)

	return task, nil
}

// runAnalysisTask 执行分析任务
func (s *RequirementAnalysisService) runAnalysisTask(task *nesma.NesmaRequirementAnalysisTask) {
	// 设置任务开始
	task.SetRunning()
	global.GVA_DB.Save(task)

	// 获取要分析的需求列表
	requirements, err := s.getRequirementsForAnalysis(task.CycleID, task.SourceVersionID, task.RequirementIDs)
	if err != nil {
		task.SetFailed(fmt.Sprintf("获取需求列表失败: %v", err))
		global.GVA_DB.Save(task)
		return
	}

	task.TotalCount = len(requirements)
	global.GVA_DB.Save(task)

	// 使用真实AI分析过程
	s.realAnalysisProcess(task, requirements)
}

// realAnalysisProcess 真实AI分析过程 - 多线程版本
func (s *RequirementAnalysisService) realAnalysisProcess(task *nesma.NesmaRequirementAnalysisTask, requirements []nesma.NesmaRequirement) {
	global.GVA_LOG.Info("开始多线程AI分析", zap.Uint("taskId", task.ID), zap.Int("requirementCount", len(requirements)))

	// 获取并发配置
	config := GetConcurrentConfig()
	maxWorkers := config.MaxAnalysisWorkers
	batchSize := config.AnalysisBatchSize
	
	global.GVA_LOG.Info("并发配置", 
		zap.Int("maxWorkers", maxWorkers), 
		zap.Int("batchSize", batchSize),
		zap.Int("totalRequirements", len(requirements)))

	// 创建任务通道和结果通道
	taskChan := make(chan nesma.NesmaRequirement, len(requirements))
	resultChan := make(chan *AnalysisResult, len(requirements))
	
	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go s.analysisWorker(i, taskChan, resultChan, &wg)
	}
	
	// 发送任务到通道
	for _, req := range requirements {
		taskChan <- req
	}
	close(taskChan)
	
	// 等待所有工作协程完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 收集结果并批量保存
	var results []*AnalysisResult
	batchCount := 0
	
	for result := range resultChan {
		results = append(results, result)
		batchCount++
		
		// 批量更新任务状态和保存需求
		if batchCount%batchSize == 0 {
			s.updateAnalysisProgress(task, results)
			results = results[:0] // 清空切片但保留容量
		}
	}
	
	// 处理剩余结果
	if len(results) > 0 {
		s.updateAnalysisProgress(task, results)
	}
	
	// 设置任务完成
	task.Summary = fmt.Sprintf("多线程AI分析完成。共处理%d个需求，成功%d个，失败%d个，AI调用%d次", 
		task.TotalCount, task.SuccessCount, task.FailedCount, task.AICallCount)
	task.SetCompleted()
	global.GVA_DB.Save(task)

	global.GVA_LOG.Info("多线程AI分析任务完成", zap.Uint("taskId", task.ID), 
		zap.Int("workers", maxWorkers), zap.Int("processedCount", task.ProcessedCount))
}

// buildNESMAAnalysisPrompt 构建NESMA分析提示词
func (s *RequirementAnalysisService) buildNESMAAnalysisPrompt(req nesma.NesmaRequirement, knowledgeService *KnowledgeService) string {
	// 1. 获取项目背景信息
	var cycleID uint
	if req.CycleID != nil {
		cycleID = *req.CycleID
	}
	projectContext := s.getProjectContext(req.ProjectID, cycleID)
	
	// 2. 获取父级需求上下文
	parentContext := s.getParentRequirementContext(req.ParentID)
	
	// 3. 获取相关知识库内容
	knowledgeContext := s.getRelevantKnowledge(req, knowledgeService)
	
	// 4. 获取同级需求参考
	siblingContext := s.getSiblingRequirementsContext(req.ParentID, req.ID)

	// 构建增强的AI分析提示词，充分利用deepseek-reasoner的推理能力
	prompt := fmt.Sprintf(`# 🧠 深度推理任务：NESMA功能点分析

<thinking>
作为一位资深NESMA专家，我需要对以下需求进行深度分析。让我系统性地思考这个问题：

1. **需求理解阶段**：
   - 首先理解需求的核心意图和业务价值
   - 分析需求在整个系统中的位置和作用
   - 识别需求的输入、处理逻辑、输出

2. **功能点分析阶段**：
   - 根据NESMA 2.2标准，判断需求属于哪种功能类型（EI/EO/EQ/ILF/EIF）
   - 分析数据元素复杂度（DET）和记录元素类型（RET）
   - 评估功能事务关系（FTR）的复杂程度

3. **复杂度评估阶段**：
   - 考虑技术实现复杂度（架构、算法、性能）
   - 评估业务逻辑复杂度（规则、流程、异常处理）
   - 分析数据复杂度（结构、关系、一致性）

4. **质量评估阶段**：
   - 检查需求的完整性、清晰性、可测试性
   - 识别潜在的风险点和改进机会
   - 提供具体的优化建议

5. **价值分析阶段**：
   - 评估需求对业务目标的贡献度
   - 分析实现成本与预期收益的比例
   - 判断需求的优先级建议
</thinking>

## 🎯 专业角色定位
你是一位拥有20年经验的NESMA国际认证专家，精通：
- **NESMA 2.2国际标准**：功能点分析的权威方法论
- **ISO/IEC 14143标准**：软件度量的国际基准
- **跨行业最佳实践**：15+个行业的项目实战经验
- **深度推理能力**：系统性分析复杂需求的能力
- **质量工程专业**：全面的软件质量保证体系

## 🔍 分析目标与要求
请运用你的深度推理能力，对以下需求进行全面而精确的NESMA功能点分析。你需要：

1. **系统性思考**：从多个维度深入理解需求
2. **科学化评估**：基于NESMA标准进行精确量化
3. **预判性分析**：识别潜在风险和改进机会
4. **价值导向优化**：提供业务价值最大化的建议
5. **具体化输出**：给出明确可执行的结论

# 📋 项目背景信息
%s

# 🔗 需求层级上下文
%s

# 🎯 当前分析需求
**需求标题：** %s
**需求描述：** %s
**需求层级：** L%d级需求
**当前功能类型：** %s
**需求编码：** %s

# 📚 NESMA知识库参考
%s

# 👥 同级需求参考
%s

# 🎯 分析任务要求

## 分析目标
请对上述需求进行全面的NESMA功能点分析和需求优化，你需要：

1. **深度理解需求本质**：透过表面描述理解真正的业务价值和功能意图
2. **精准功能分类**：基于NESMA标准准确识别功能类型
3. **智能需求优化**：提供专业的需求描述优化建议
4. **复杂度科学评估**：综合考虑数据、逻辑、接口复杂度
5. **价值导向分析**：确保分析结果对项目团队具有实际指导价值

## 分析维度
- **功能特征分析**：数据处理类型、业务流程复杂度、用户交互方式
- **技术实现考量**：接口复杂度、数据存储需求、计算逻辑难度
- **业务价值评估**：用户价值、业务重要性、实现优先级
- **质量属性要求**：性能、安全、可用性、扩展性需求

## 输出要求
严格按照以下JSON格式返回分析结果，不要包含任何格式标记、解释文字或其他内容：

{
  "optimized_title": "经过专业优化的需求标题，要求简洁明确，体现核心功能价值",
  "optimized_description": "详细优化的需求描述，包含：业务场景、功能细节、用户价值、接受条件、边界条件等完整信息，描述要专业且易懂,描述要详细但不冗长，要符合人类自然语言，要详细但不冗长,不要包含输入、处理和输出，使用自然语言描述，输出长度在100-200字左右",
  "function_type": "基于NESMA标准的精确分类：EI(外部输入)/EO(外部输出)/EQ(外部查询)/ILF(内部逻辑文件)/EIF(外部接口文件)",
  "complexity_score": "0-100数值，综合评估技术实现难度、业务逻辑复杂度、数据处理复杂度",
  "confidence_score": "0.0-1.0小数，表示分析结果的置信度",
  "recommended_afp": "调整功能点数值，考虑项目特性和复杂度",
  "recommended_ufp": "未调整功能点数值，严格按NESMA标准计算",
  "analysis_notes": "详细的专业分析说明，包括分类依据、复杂度评估理由、优化建议说明、实施要点提醒等，分析要详细但不冗长，要符合人类自然语言，要详细但不冗长,不要包含输入、处理和输出，使用自然语言描述，输出长度在100-200字左右",
  "business_value": "业务价值分析，说明该需求对整体项目的重要性和价值贡献，分析要详细但不冗长，要符合人类自然语言，要详细但不冗长,不要包含输入、处理和输出，使用自然语言描述，输出长度在100-200字左右",
  "reuse_level": "重用程度评估：高/中/低，基于该功能在其他项目或模块中的可复用性",
  "modification_type": "修改类型：新增/优化/修改/删除，表示该需求在项目中的实施性质",
  "acceptance_criteria": "验收标准建议，明确的、可测试的验收条件和标准"
}

## 🧠 深度推理指导

在分析过程中，请遵循以下推理链条：

### 第一步：理解与建模
- 仔细理解需求的业务背景和技术上下文
- 建立需求的概念模型（输入-处理-输出模型）
- 识别关键的业务实体和业务规则

### 第二步：分类与映射
- 将需求映射到NESMA功能类型的判断决策树
- 考虑数据持久化特征（内部/外部数据）
- 考虑用户交互特征（输入/输出/查询）

### 第三步：复杂度量化
- 数据元素复杂度（DET）：计算输入/输出字段数量和复杂程度
- 记录元素类型（RET）：分析涉及的数据实体和关系复杂度
- 功能事务关系（FTR）：评估外部引用和内部逻辑复杂度

### 第四步：质量与风险评估
- 识别需求描述中的模糊点和歧义
- 评估实现风险（技术风险、依赖风险、时间风险）
- 提出具体的风险缓解措施

### 第五步：优化与价值提升
- 基于最佳实践优化需求描述
- 提供能提升需求质量的具体建议
- 确保分析结果对项目团队具有实际指导价值

请严格按照上述推理流程进行深度分析，确保每个结论都有坚实的逻辑基础！`, 
		projectContext, 
		parentContext, 
		req.Title, 
		req.Description, 
		req.Level, 
		req.FunctionType, 
		req.Code, 
		knowledgeContext, 
		siblingContext)

	return prompt
}

// getProjectContext 获取项目背景信息
func (s *RequirementAnalysisService) getProjectContext(projectID, cycleID uint) string {
	var project nesma.NesmaProject
	var cycle nesma.NesmaProjectCycle
	
	// 获取项目信息
	if err := global.GVA_DB.First(&project, projectID).Error; err != nil {
		return "项目信息：暂无法获取项目详细信息"
	}
	
	// 获取周期信息
	if err := global.GVA_DB.First(&cycle, cycleID).Error; err != nil {
		return fmt.Sprintf("**项目名称：** %s\n**项目描述：** %s\n**业务领域：** %s\n**周期信息：** 暂无法获取周期详情", 
			project.Name, project.Description, project.Domain)
	}
	
	return fmt.Sprintf(`**项目名称：** %s
**项目描述：** %s  
**业务领域：** %s
**当前周期：** %s
**周期描述：** %s
**项目状态：** %s`, 
		project.Name, 
		project.Description, 
		project.Domain, 
		cycle.Name, 
		cycle.Description, 
		cycle.Status)
}

// getParentRequirementContext 获取父级需求上下文
func (s *RequirementAnalysisService) getParentRequirementContext(parentID *uint) string {
	if parentID == nil {
		return "**层级位置：** 顶级需求，无上级需求"
	}
	
	var parentReq nesma.NesmaRequirement
	if err := global.GVA_DB.First(&parentReq, *parentID).Error; err != nil {
		return "**上级需求：** 暂无法获取上级需求信息"
	}
	
	// 递归获取更上级的需求信息
	var grandParentInfo string
	if parentReq.ParentID != nil {
		var grandParentReq nesma.NesmaRequirement
		if err := global.GVA_DB.First(&grandParentReq, *parentReq.ParentID).Error; err == nil {
			grandParentInfo = fmt.Sprintf("\n**上上级需求：** %s (%s)", grandParentReq.Title, grandParentReq.Code)
		}
	}
	
	return fmt.Sprintf(`**直接上级需求：** %s (%s)
**上级需求描述：** %s
**上级需求类型：** L%d级 - %s%s`, 
		parentReq.Title, 
		parentReq.Code, 
		parentReq.Description, 
		parentReq.Level, 
		parentReq.FunctionType,
		grandParentInfo)
}

// getRelevantKnowledge 获取相关知识库内容
func (s *RequirementAnalysisService) getRelevantKnowledge(req nesma.NesmaRequirement, knowledgeService *KnowledgeService) string {
	if knowledgeService == nil {
		return s.getDefaultNESMAKnowledge()
	}
	
	// 构建搜索关键词（保留但不使用，避免编译警告）
	_ = fmt.Sprintf("%s %s %s 功能点分析", req.Title, req.Description, req.FunctionType)
	
	// 尝试从知识库搜索相关内容（简化实现）
	// TODO: 实现真正的知识库搜索
	knowledge := s.getDefaultNESMAKnowledge()
	
	// 添加针对性的知识内容
	specificKnowledge := s.getSpecificKnowledgeByLevel(req.Level, req.FunctionType)
	
	return fmt.Sprintf("%s\n\n## 针对性指导\n%s", knowledge, specificKnowledge)
}

// getSiblingRequirementsContext 获取同级需求参考
func (s *RequirementAnalysisService) getSiblingRequirementsContext(parentID *uint, currentID uint) string {
	var siblings []nesma.NesmaRequirement
	var query *gorm.DB
	
	if parentID != nil {
		query = global.GVA_DB.Where("parent_id = ? AND id != ?", *parentID, currentID)
	} else {
		query = global.GVA_DB.Where("parent_id IS NULL AND id != ?", currentID)
	}
	
	if err := query.Limit(3).Find(&siblings).Error; err != nil || len(siblings) == 0 {
		return "**同级需求：** 暂无同级需求参考"
	}
	
	siblingInfo := "**同级需求参考：**\n"
	for i, sibling := range siblings {
		siblingInfo += fmt.Sprintf("%d. %s (%s) - %s\n", 
			i+1, sibling.Title, sibling.Code, sibling.FunctionType)
	}
	
	return siblingInfo
}

// getDefaultNESMAKnowledge 获取默认NESMA知识
func (s *RequirementAnalysisService) getDefaultNESMAKnowledge() string {
	return `## NESMA 2.1 功能点分析标准

### 功能类型定义
- **EI (External Input)**: 处理来自应用边界外的数据或控制信息的基本过程
- **EO (External Output)**: 向应用边界外发送数据或控制信息的基本过程  
- **EQ (External Inquiry)**: 从应用边界外发送输入并接收输出的基本过程
- **ILF (Internal Logical File)**: 由应用维护的用户可识别的逻辑相关数据组
- **EIF (External Interface File)**: 由其他应用维护但本应用引用的用户可识别的逻辑相关数据组

### 复杂度评估标准
- **数据元素类型(DET)**: 用户可识别的非重复字段
- **记录元素类型(RET)**: 用户可识别的数据子组
- **文件类型引用(FTR)**: 基本过程维护或引用的逻辑文件

### 功能点计算矩阵
| 类型 | 简单 | 中等 | 复杂 |
|------|------|------|------|
| EI   | 3    | 4    | 6    |
| EO   | 4    | 5    | 7    |
| EQ   | 3    | 4    | 6    |
| ILF  | 7    | 10   | 15   |
| EIF  | 5    | 7    | 10   |`
}

// getSpecificKnowledgeByLevel 根据层级获取特定知识
func (s *RequirementAnalysisService) getSpecificKnowledgeByLevel(level int, functionType string) string {
	levelGuidance := map[int]string{
		1: "L1级需求通常是业务目标和愿景，关注整体价值和战略目标",
		2: "L2级需求是功能模块划分，要明确模块边界和主要职责",
		3: "L3级需求是具体功能描述，需要详细的输入输出和处理逻辑",
		4: "L4级需求是详细规格说明，要求完整的技术实现细节",
	}
	
	typeGuidance := map[string]string{
		"EI": "外部输入功能重点关注数据验证、业务规则应用、数据转换存储",
		"EO": "外部输出功能重点关注数据查询、格式化、报表生成逻辑",
		"EQ": "外部查询功能重点关注查询条件、数据检索、结果展示",
		"ILF": "内部逻辑文件重点关注数据结构、存储方式、维护操作",
		"EIF": "外部接口文件重点关注接口定义、数据交换、集成方式",
	}
	
	guidance := levelGuidance[level]
	if guidance == "" {
		guidance = "请根据需求特点进行专业分析"
	}
	
	typeGuide := typeGuidance[functionType]
	if typeGuide == "" {
		typeGuide = "请基于NESMA标准进行功能分类"
	}
	
	return fmt.Sprintf("**层级指导：** %s\n**类型指导：** %s", guidance, typeGuide)
}

// callAIForRequirement 调用AI分析单个需求 - 升级为使用AI服务管理器
func (s *RequirementAnalysisService) callAIForRequirement(aiService *DeepSeekService, prompt string) (*AIResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 增加到5分钟
	defer cancel()

	// 使用新的AI服务管理器
	aiManager := GetAIServiceManager()
	
	// 构建需求上下文信息
	requirementContext := map[string]interface{}{
		"analysis_type": "nesma_requirement",
		"language":      "chinese",
		"standard":      "nesma_v2.2",
	}
	
	// 使用专门的NESMA分析接口
	response, err := aiManager.AnalyzeNESMARequirement(ctx, prompt, requirementContext)
	if err != nil {
		// 降级处理：使用传统方式
		global.GVA_LOG.Warn("AI服务管理器分析失败，降级到传统方式", zap.Error(err))
		config := &AIConfig{
			MaxTokens:   8192,
			Temperature: 0.3,
			TopP:        0.9,
			Model:       "deepseek-chat",
			Stream:      false,
		}
		return aiService.GenerateText(ctx, prompt, config)
	}
	
	// 转换响应格式
	aiResponse := &AIResponse{
		Choices: []Choice{
			{
				Message: APIMessage{
					Role:    "assistant",
					Content: response.Content,
				},
			},
		},
		Usage: Usage{
			TotalTokens:      response.TokensUsed,
			PromptTokens:     response.TokensUsed / 2, // 估算
			CompletionTokens: response.TokensUsed / 2, // 估算
		},
	}
	
	global.GVA_LOG.Info("AI分析完成", 
		zap.String("model", response.Model),
		zap.Duration("processing_time", response.ProcessingTime),
		zap.Int("tokens_used", response.TokensUsed),
		zap.Float64("confidence", response.Confidence))
	
	return aiResponse, nil
}

// parseAIResponseToRequirement 解析AI响应并直接应用优化结果到需求字段
func (s *RequirementAnalysisService) parseAIResponseToRequirement(req *nesma.NesmaRequirement, aiResponse *AIResponse) {
	if len(aiResponse.Choices) == 0 {
		s.setDefaultAIValues(req)
		return
	}

	responseText := aiResponse.Choices[0].Message.Content
	
	// 更智能的JSON提取逻辑
	cleanedText := s.extractJSONFromResponse(responseText)
	
	global.GVA_LOG.Debug("AI响应处理", 
		zap.String("originalResponse", responseText),
		zap.String("cleanedJSON", cleanedText),
		zap.Uint("requirementID", req.ID))
	
	// 尝试解析JSON响应
	var analysisResult map[string]interface{}
	if err := json.Unmarshal([]byte(cleanedText), &analysisResult); err != nil {
		global.GVA_LOG.Warn("AI响应解析失败，使用默认值", 
			zap.Error(err), 
			zap.String("cleanedText", cleanedText), 
			zap.String("originalResponse", responseText))
		s.setDefaultAIValues(req)
		req.AIDescription = fmt.Sprintf("AI原始响应：\n%s", responseText)
		return
	}

	// 保持原有的基本信息不变
	originalLevel := req.Level
	originalCode := req.Code
	originalParentID := req.ParentID
	
	// 解析各个字段并直接应用到主字段
	req.AIAnalysisStatus = "completed"
	req.Status = "completed"
	
	// 配置：是否直接应用AI优化结果（默认开启）
	autoApplyOptimization := true // 可以从配置文件或数据库读取
	
	// 1. 标题字段映射 - 直接应用优化结果
	if title, ok := analysisResult["optimized_title"].(string); ok && title != "" && len(title) < 500 {
		// 保存原始标题到AI字段作为备份
		req.AIGeneratedTitle = title
		
		// 直接应用优化后的标题（如果配置启用）
		if autoApplyOptimization {
			originalTitle := req.Title
			req.Title = title
			global.GVA_LOG.Info("应用AI优化标题", 
				zap.Uint("requirementID", req.ID),
				zap.String("originalTitle", originalTitle),
				zap.String("optimizedTitle", title))
		}
	} else {
		req.AIGeneratedTitle = fmt.Sprintf("【AI优化】%s", req.Title)
	}

	// 2. 描述字段映射 - 直接应用优化结果
	if desc, ok := analysisResult["optimized_description"].(string); ok && desc != "" && len(desc) < 2000 {
		// 保存AI优化描述到备份字段
		req.AIDescription = desc
		
		// 直接应用优化后的描述（如果配置启用）
		if autoApplyOptimization {
			originalDesc := req.Description
			req.Description = desc
			global.GVA_LOG.Info("应用AI优化描述", 
				zap.Uint("requirementID", req.ID),
				zap.String("originalDesc", s.truncateString(originalDesc, 50)),
				zap.Int("optimizedDescLength", len(desc)))
		}
	} else {
		req.AIDescription = req.Description
	}

	// 3. 功能类型字段映射 - 直接应用
	if funcType, ok := analysisResult["function_type"].(string); ok && funcType != "" {
		validTypes := map[string]bool{"EI": true, "EO": true, "EQ": true, "ILF": true, "EIF": true}
		if validTypes[funcType] {
			oldFunctionType := req.FunctionType
			req.FunctionType = funcType
			global.GVA_LOG.Info("应用AI功能类型分析", 
				zap.Uint("requirementID", req.ID),
				zap.String("oldFunctionType", oldFunctionType),
				zap.String("newFunctionType", funcType))
		}
	}

	// 4. 复杂度分数 - 直接应用到复杂度字段
	if complexity, ok := analysisResult["complexity_score"].(float64); ok && complexity >= 0 && complexity <= 100 {
		req.AIComplexityScore = &complexity
		
		// 直接更新复杂度级别
		if autoApplyOptimization {
			oldComplexity := req.Complexity
			if complexity >= 80 {
				req.Complexity = "复杂"
				req.ComplexityLevel = "High"
			} else if complexity >= 50 {
				req.Complexity = "中等"
				req.ComplexityLevel = "Average"
			} else {
				req.Complexity = "简单"
				req.ComplexityLevel = "Low"
			}
			
			global.GVA_LOG.Info("应用AI复杂度分析", 
				zap.Uint("requirementID", req.ID),
				zap.String("oldComplexity", oldComplexity),
				zap.String("newComplexity", req.Complexity),
				zap.Float64("complexityScore", complexity))
		}
	}

	// 5. 置信度分数
	if confidence, ok := analysisResult["confidence_score"].(float64); ok && confidence >= 0 && confidence <= 1 {
		req.AIConfidenceScore = &confidence
	}

	// 6. AFP/UFP 推荐值映射 - 直接应用
	if afp, ok := analysisResult["recommended_afp"].(float64); ok && afp >= 0 && afp <= 50 {
		req.RecommendedAFP = &afp
		
		if autoApplyOptimization {
			oldAFP := req.AFP
			req.AFP = afp
			global.GVA_LOG.Info("应用AI AFP推荐值", 
				zap.Uint("requirementID", req.ID),
				zap.Float64("oldAFP", oldAFP),
				zap.Float64("newAFP", afp))
		}
	}

	if ufp, ok := analysisResult["recommended_ufp"].(float64); ok && ufp >= 0 && ufp <= 50 {
		req.RecommendedUFP = &ufp
		
		if autoApplyOptimization {
			oldUFP := req.UFP
			req.UFP = ufp
			global.GVA_LOG.Info("应用AI UFP推荐值", 
				zap.Uint("requirementID", req.ID),
				zap.Float64("oldUFP", oldUFP),
				zap.Float64("newUFP", ufp))
		}
	}

	// 7. 分析说明 - 直接应用到备注字段
	if notes, ok := analysisResult["analysis_notes"].(string); ok && notes != "" && len(notes) < 1000 {
		if autoApplyOptimization {
			// 将AI分析说明追加到现有备注
			if req.Notes != "" {
				req.Notes = req.Notes + "\n\n【AI分析】\n" + notes
			} else {
				req.Notes = "【AI分析】\n" + notes
			}
		}
	}

	// 8. 重用程度推荐 - 新增功能
	if reuseLevel, ok := analysisResult["reuse_level"].(string); ok && reuseLevel != "" {
		validReuseLevels := map[string]bool{"高": true, "中": true, "低": true, "High": true, "Medium": true, "Low": true}
		if validReuseLevels[reuseLevel] && autoApplyOptimization {
			// 标准化重用程度
			switch reuseLevel {
			case "High":
				req.ReuseLevel = "高"
			case "Medium":
				req.ReuseLevel = "中"
			case "Low":
				req.ReuseLevel = "低"
			default:
				req.ReuseLevel = reuseLevel
			}
			global.GVA_LOG.Info("应用AI重用程度分析", 
				zap.Uint("requirementID", req.ID),
				zap.String("reuseLevel", req.ReuseLevel))
		}
	}

	// 9. 修改类型推荐 - 新增功能
	if modType, ok := analysisResult["modification_type"].(string); ok && modType != "" {
		validModTypes := map[string]bool{"新增": true, "优化": true, "删除": true, "修改": true}
		if validModTypes[modType] && autoApplyOptimization {
			req.ModificationType = modType
			global.GVA_LOG.Info("应用AI修改类型分析", 
				zap.Uint("requirementID", req.ID),
				zap.String("modificationType", modType))
		}
	}

	// 10. 业务价值分析 - 直接应用到业务价值字段
	if businessValue, ok := analysisResult["business_value"].(string); ok && businessValue != "" && len(businessValue) < 1000 {
		if autoApplyOptimization {
			req.BusinessValue = businessValue
			global.GVA_LOG.Info("应用AI业务价值分析", 
				zap.Uint("requirementID", req.ID),
				zap.Int("businessValueLength", len(businessValue)))
		}
	}

	// 11. 验收标准建议 - 直接应用
	if acceptanceCriteria, ok := analysisResult["acceptance_criteria"].(string); ok && acceptanceCriteria != "" && len(acceptanceCriteria) < 1000 {
		if autoApplyOptimization {
			req.AcceptanceCriteria = acceptanceCriteria
			global.GVA_LOG.Info("应用AI验收标准建议", 
				zap.Uint("requirementID", req.ID),
				zap.Int("acceptanceCriteriaLength", len(acceptanceCriteria)))
		}
	}

	// 恢复不应更改的字段
	req.Level = originalLevel
	req.Code = originalCode
	req.ParentID = originalParentID
	
	// 标记AI优化已应用
	if autoApplyOptimization {
		req.AIOptimizationApplied = true
		now := time.Now()
		req.AIOptimizationAppliedAt = &now
	}
	
	global.GVA_LOG.Info("AI响应解析完成，优化结果已应用", 
		zap.Uint("requirementID", req.ID),
		zap.String("functionType", req.FunctionType),
		zap.String("complexity", req.Complexity),
		zap.Float64("afp", req.AFP),
		zap.Float64("ufp", req.UFP),
		zap.String("reuseLevel", req.ReuseLevel),
		zap.String("modificationType", req.ModificationType),
		zap.Bool("autoApplied", autoApplyOptimization))
}

// getFloatValue 安全获取浮点数值
func (s *RequirementAnalysisService) getFloatValue(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}

// truncateString 截断字符串到指定长度
func (s *RequirementAnalysisService) truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "..."
}

// extractJSONFromResponse 从AI响应中提取JSON
func (s *RequirementAnalysisService) extractJSONFromResponse(response string) string {
	// 查找JSON开始和结束标记
	startIdx := strings.Index(response, "{")
	if startIdx == -1 {
		return response
	}
	
	// 从后往前查找最后一个}
	endIdx := strings.LastIndex(response, "}")
	if endIdx == -1 || endIdx <= startIdx {
		return response
	}
	
	return response[startIdx : endIdx+1]
}

// setDefaultAIValues 设置默认AI分析值
func (s *RequirementAnalysisService) setDefaultAIValues(req *nesma.NesmaRequirement) {
	req.AIGeneratedTitle = fmt.Sprintf("【AI优化】%s", req.Title)
	req.AIDescription = fmt.Sprintf("AI自动优化的需求描述：%s", req.Description)
	
	defaultComplexity := 70.0
	req.AIComplexityScore = &defaultComplexity
	req.Complexity = "Average"
	
	defaultConfidence := 0.75
	req.AIConfidenceScore = &defaultConfidence
	
	defaultAFP := 3.0
	req.RecommendedAFP = &defaultAFP
	req.AFP = defaultAFP
	
	defaultUFP := 3.0
	req.RecommendedUFP = &defaultUFP
	req.UFP = defaultUFP
	
	if req.FunctionType == "" {
		req.FunctionType = "EI"
	}
	
	req.Notes = "AI分析失败，使用默认值"
}

// getRequirementsForAnalysis 获取需要分析的需求列表并创建新版本副本（支持特定需求ID过滤）
func (s *RequirementAnalysisService) getRequirementsForAnalysis(cycleID uint, sourceVersionID uint, requirementIDsBytes []byte) ([]nesma.NesmaRequirement, error) {
	global.GVA_LOG.Info("开始查询并准备需求分析", zap.Uint("cycleID", cycleID), zap.Uint("sourceVersionID", sourceVersionID))
	
	// 解析需求ID列表
	var specificRequirementIDs []uint
	if len(requirementIDsBytes) > 0 {
		if err := json.Unmarshal(requirementIDsBytes, &specificRequirementIDs); err != nil {
			global.GVA_LOG.Error("解析需求ID列表失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("解析到特定需求ID", zap.Any("ids", specificRequirementIDs))
		}
	}
	
	// 1. 查询源版本的需求（按层级排序）
	var sourceRequirements []nesma.NesmaRequirement
	query := global.GVA_DB.Where("cycle_id = ? AND version_id = ?", cycleID, sourceVersionID)
	
	// 如果有特定需求ID，添加ID过滤
	if len(specificRequirementIDs) > 0 {
		query = query.Where("id IN ?", specificRequirementIDs)
		global.GVA_LOG.Info("按特定需求ID查询", zap.Any("requirementIds", specificRequirementIDs))
	} else {
		// 只查询L1-L4级别的需求进行完整复制
		query = query.Where("level >= ? AND level <= ?", 1, 4)
		global.GVA_LOG.Info("查询整个周期的L1-L4级别需求")
	}
	
	// 按层级和代码排序，确保父级在子级之前
	err := query.Order("level ASC, code ASC").Find(&sourceRequirements).Error
	if err != nil {
		global.GVA_LOG.Error("查询源需求失败", zap.Error(err))
		return nil, err
	}
	
	global.GVA_LOG.Info("查询到源需求", zap.Int("count", len(sourceRequirements)))
	
	if len(sourceRequirements) == 0 {
		return []nesma.NesmaRequirement{}, nil
	}
	
	// 2. 创建新版本
	newVersion, err := s.createNewAnalysisVersion(cycleID)
	if err != nil {
		return nil, err
	}
	
	global.GVA_LOG.Info("创建新分析版本", zap.Uint("newVersionID", newVersion.ID), zap.String("version", newVersion.Version))
	
	// 3. 按层级顺序复制需求，维护父子关系
	newRequirements, err := s.createRequirementsWithHierarchy(sourceRequirements, newVersion.ID)
	if err != nil {
		return nil, err
	}
	
	global.GVA_LOG.Info("成功创建新版本需求", zap.Int("count", len(newRequirements)))
	
	// 4. 只返回需要分析的L2-L4级别需求
	var analysisRequirements []nesma.NesmaRequirement
	for _, req := range newRequirements {
		if req.Level >= 2 && req.Level <= 4 {
			analysisRequirements = append(analysisRequirements, req)
		}
	}
	
	global.GVA_LOG.Info("返回待分析需求", zap.Int("total", len(newRequirements)), zap.Int("forAnalysis", len(analysisRequirements)))
	
	return analysisRequirements, nil
}

// createNewAnalysisVersion 创建新的分析版本
func (s *RequirementAnalysisService) createNewAnalysisVersion(cycleID uint) (*nesma.NesmaRequirementVersion, error) {
	// 查询当前最新版本号
	var latestVersion nesma.NesmaRequirementVersion
	err := global.GVA_DB.Where("cycle_id = ?", cycleID).
		Order("created_at DESC").
		First(&latestVersion).Error
	
	var newVersionName string
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newVersionName = "v1.0"
		} else {
			return nil, err
		}
	} else {
		// 生成新版本号（简单递增）
		newVersionName = s.generateNextVersion(latestVersion.Version)
	}
	
	// 创建新版本
	newVersion := nesma.NesmaRequirementVersion{
		CycleID:     cycleID,
		Version:     newVersionName,
		VersionType: "ai_analysis",
		CreatedBy:   "ai_system",
		Summary:     fmt.Sprintf("AI分析优化版本 - %s", time.Now().Format("2006-01-02 15:04:05")),
		Status:      "active",
	}
	
	err = global.GVA_DB.Create(&newVersion).Error
	if err != nil {
		return nil, err
	}
	
	return &newVersion, nil
}

// createRequirementsWithHierarchy 按层级顺序创建需求，维护父子关系
func (s *RequirementAnalysisService) createRequirementsWithHierarchy(sourceRequirements []nesma.NesmaRequirement, newVersionID uint) ([]nesma.NesmaRequirement, error) {
	// 使用事务确保数据一致性
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// 旧ID到新ID的映射关系
	idMapping := make(map[uint]uint)
	var newRequirements []nesma.NesmaRequirement
	
	// 按层级分组
	levelGroups := make(map[int][]nesma.NesmaRequirement)
	for _, req := range sourceRequirements {
		levelGroups[req.Level] = append(levelGroups[req.Level], req)
	}
	
	// 按层级从1到4依次创建
	for level := 1; level <= 4; level++ {
		requirements := levelGroups[level]
		if len(requirements) == 0 {
			continue
		}
		
		global.GVA_LOG.Info("创建层级需求", zap.Int("level", level), zap.Int("count", len(requirements)))
		
		for _, sourceReq := range requirements {
			// 创建新需求副本
			newReq := nesma.NesmaRequirement{
				ProjectID:         sourceReq.ProjectID,
				CycleID:          sourceReq.CycleID,
				VersionID:        &newVersionID,
				Level:            sourceReq.Level,
				Code:             sourceReq.Code,
				Title:            sourceReq.Title,
				Description:      sourceReq.Description,
				FunctionType:     sourceReq.FunctionType,
				Complexity:       sourceReq.Complexity,
				AFP:              sourceReq.AFP,
				UFP:              sourceReq.UFP,
				Notes:            sourceReq.Notes,
				Status:           "pending_analysis",
				ParentID:         nil,
			}
			
			// 如果有父级，从映射中查找新的父级ID
			if sourceReq.ParentID != nil {
				if newParentID, exists := idMapping[*sourceReq.ParentID]; exists {
					newReq.ParentID = &newParentID
				} else {
					global.GVA_LOG.Warn("找不到父级需求的新ID", zap.Uint("sourceParentID", *sourceReq.ParentID), zap.Uint("sourceReqID", sourceReq.ID))
				}
			}
			
			// 保存新需求
			err := tx.Create(&newReq).Error
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("创建需求失败: %w", err)
			}
			
			// 记录ID映射关系
			idMapping[sourceReq.ID] = newReq.ID
			newRequirements = append(newRequirements, newReq)
			
			global.GVA_LOG.Debug("创建需求映射", 
				zap.Uint("sourceID", sourceReq.ID), 
				zap.Uint("newID", newReq.ID),
				zap.String("code", newReq.Code),
				zap.Int("level", newReq.Level))
		}
	}
	
	// 提交事务
	err := tx.Commit().Error
	if err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}
	
	global.GVA_LOG.Info("成功创建所有层级需求", 
		zap.Int("total", len(newRequirements)),
		zap.Int("idMappings", len(idMapping)))
	
	return newRequirements, nil
}

// generateNextVersion 生成下一个版本号
func (s *RequirementAnalysisService) generateNextVersion(currentVersion string) string {
	// 简单的版本号递增逻辑
	// v1.0 -> v1.1 -> v1.2 ... -> v2.0
	parts := strings.Split(strings.TrimPrefix(currentVersion, "v"), ".")
	if len(parts) != 2 {
		return "v1.1"
	}
	
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	
	minor++
	if minor >= 10 {
		major++
		minor = 0
	}
	
	return fmt.Sprintf("v%d.%d", major, minor)
}

// getOrCreateActiveVersion 获取或创建激活版本
func (s *RequirementAnalysisService) getOrCreateActiveVersion(cycleID uint) (*nesma.NesmaRequirementVersion, error) {
	var version nesma.NesmaRequirementVersion
	
	// 查找激活版本
	err := global.GVA_DB.Where("cycle_id = ?", cycleID).Order("created_at DESC").First(&version).Error
	if err == nil {
		return &version, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 如果没有激活版本，创建一个初始版本
	version = nesma.NesmaRequirementVersion{
		CycleID:     cycleID,
		Version:     "v1.0",
		VersionType: "initial",
		CreatedBy:   "system",
		Summary:     "系统自动创建的初始版本",
		Status:      "active",
	}

	if err := global.GVA_DB.Create(&version).Error; err != nil {
		return nil, err
	}

	return &version, nil
}

// createAnalysisVersion 创建分析版本
func (s *RequirementAnalysisService) createAnalysisVersion(cycleID uint, taskID uint) (*nesma.NesmaRequirementVersion, error) {
	// 查询最新版本号以生成新版本号
	var latestVersion nesma.NesmaRequirementVersion
	global.GVA_DB.Where("cycle_id = ?", cycleID).Order("created_at DESC").First(&latestVersion)
	
	newVersionNumber := s.generateNextVersion(latestVersion.Version)

	version := &nesma.NesmaRequirementVersion{
		CycleID:         cycleID,
		Version:         newVersionNumber,
		VersionType:     "analyzed",
		CreatedBy:       "ai_analysis",
		Summary:         "AI分析优化版本",
		Description:     "通过AI智能分析生成的需求优化版本",
		Status:          "draft",
		AnalysisTaskID:  &taskID,
		AnalysisStartTime: func() *time.Time { t := time.Now(); return &t }(),
	}

	if err := global.GVA_DB.Create(version).Error; err != nil {
		return nil, err
	}

	return version, nil
}

// buildTaskConfig 构建任务配置
func (s *RequirementAnalysisService) buildTaskConfig(req *nesmaReq.NesmaAnalyzeRequest) map[string]interface{} {
	config := map[string]interface{}{
		"analysisType":  "requirement_analysis",
		"targetLevels":  []int{3, 4},
		"aiModel":       "deepseek",
		"enableMermaid": false,
		"batchSize":     10,
		"maxRetries":    3,
		"skipCompleted": true,
	}

	// 如果请求中有配置，覆盖默认值
	if req.Config.AnalysisType != "" {
		config["analysisType"] = req.Config.AnalysisType
	}
	if len(req.Config.TargetLevels) > 0 {
		config["targetLevels"] = req.Config.TargetLevels
	}
	if req.Config.AIModel != "" {
		config["aiModel"] = req.Config.AIModel
	}

	return config
}

// GetAnalysisProgress 获取分析进度
func (s *RequirementAnalysisService) GetAnalysisProgress(taskID uint) (*nesma.NesmaRequirementAnalysisTask, error) {
	var task nesma.NesmaRequirementAnalysisTask
	err := global.GVA_DB.Preload("Project").
		Preload("Cycle").
		Preload("SourceVersion").
		Preload("TargetVersion").
		First(&task, taskID).Error
	return &task, err
}

// GetAnalysisTasks 获取分析任务列表（旧版本，保持兼容性）
func (s *RequirementAnalysisService) GetAnalysisTasks(projectID, cycleID uint) ([]nesma.NesmaRequirementAnalysisTask, error) {
	var tasks []nesma.NesmaRequirementAnalysisTask
	
	db := global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{})
	
	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}
	if cycleID > 0 {
		db = db.Where("cycle_id = ?", cycleID)
	}

	err := db.Preload("Project").
		Preload("Cycle").
		Order("created_at DESC").
		Find(&tasks).Error

	return tasks, err
}

// DeleteAnalysisTask 删除分析任务
func (s *RequirementAnalysisService) DeleteAnalysisTask(taskID uint) error {
	return global.GVA_DB.Delete(&nesma.NesmaRequirementAnalysisTask{}, taskID).Error
}

// GetAnalysisTasksWithPagination 获取分析任务列表（支持分页和筛选）
func (s *RequirementAnalysisService) GetAnalysisTasksWithPagination(req *nesmaReq.AnalysisTaskListRequest) ([]nesma.NesmaRequirementAnalysisTask, int64, error) {
	var tasks []nesma.NesmaRequirementAnalysisTask
	var total int64
	
	db := global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{})
	
	// 基础筛选条件
	if req.ProjectID != nil && *req.ProjectID > 0 {
		db = db.Where("project_id = ?", *req.ProjectID)
	}
	if req.CycleID != nil && *req.CycleID > 0 {
		db = db.Where("cycle_id = ?", *req.CycleID)
	}
	
	// 任务类型筛选
	if req.TaskType != "" {
		db = db.Where("task_type = ?", req.TaskType)
	}
	
	// 状态筛选
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	
	// 创建时间筛选
	if req.StartDate != "" {
		startTime, err := time.Parse("2006-01-02 15:04:05", req.StartDate)
		if err == nil {
			db = db.Where("created_at >= ?", startTime)
		} else {
			// 尝试只解析日期部分
			startTime, err = time.Parse("2006-01-02", req.StartDate)
			if err == nil {
				db = db.Where("created_at >= ?", startTime)
			}
		}
	}
	
	if req.EndDate != "" {
		endTime, err := time.Parse("2006-01-02 15:04:05", req.EndDate)
		if err == nil {
			db = db.Where("created_at <= ?", endTime)
		} else {
			// 尝试只解析日期部分，如果只是日期，则设置为当天结束时间
			endTime, err = time.Parse("2006-01-02", req.EndDate)
			if err == nil {
				endTime = endTime.Add(24*time.Hour - time.Second) // 设置为23:59:59
				db = db.Where("created_at <= ?", endTime)
			}
		}
	}
	
	// 统计总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 分页参数
	offset := (req.Page - 1) * req.PageSize
	if offset < 0 {
		offset = 0
	}
	
	// 查询数据
	err = db.Preload("Project").
		Preload("Cycle").
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&tasks).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	global.GVA_LOG.Info("查询分析任务列表", 
		zap.Int("page", req.Page),
		zap.Int("pageSize", req.PageSize),
		zap.String("taskType", req.TaskType),
		zap.String("status", req.Status),
		zap.Int64("total", total),
		zap.Int("results", len(tasks)))
	
	return tasks, total, nil
}

// GetTaskStatistics 获取任务统计信息
func (s *RequirementAnalysisService) GetTaskStatistics(projectID, cycleID uint) (map[string]interface{}, error) {
	var stats struct {
		Total     int64 `json:"total"`
		Running   int64 `json:"running"`
		Completed int64 `json:"completed"`
		Failed    int64 `json:"failed"`
		Pending   int64 `json:"pending"`
		Cancelled int64 `json:"cancelled"`
	}

	db := global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{})
	
	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}
	if cycleID > 0 {
		db = db.Where("cycle_id = ?", cycleID)
	}

	// 统计总数
	db.Count(&stats.Total)

	// 按状态统计 - 为每个状态创建独立的查询
	if projectID > 0 && cycleID > 0 {
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND cycle_id = ? AND status = ?", projectID, cycleID, "running").Count(&stats.Running)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND cycle_id = ? AND status = ?", projectID, cycleID, "completed").Count(&stats.Completed)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND cycle_id = ? AND status = ?", projectID, cycleID, "failed").Count(&stats.Failed)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND cycle_id = ? AND status = ?", projectID, cycleID, "pending").Count(&stats.Pending)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND cycle_id = ? AND status = ?", projectID, cycleID, "cancelled").Count(&stats.Cancelled)
	} else if projectID > 0 {
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND status = ?", projectID, "running").Count(&stats.Running)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND status = ?", projectID, "completed").Count(&stats.Completed)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND status = ?", projectID, "failed").Count(&stats.Failed)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND status = ?", projectID, "pending").Count(&stats.Pending)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("project_id = ? AND status = ?", projectID, "cancelled").Count(&stats.Cancelled)
	} else {
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("status = ?", "running").Count(&stats.Running)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("status = ?", "completed").Count(&stats.Completed)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("status = ?", "failed").Count(&stats.Failed)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("status = ?", "pending").Count(&stats.Pending)
		global.GVA_DB.Model(&nesma.NesmaRequirementAnalysisTask{}).
			Where("status = ?", "cancelled").Count(&stats.Cancelled)
	}

	result := map[string]interface{}{
		"total":     stats.Total,
		"running":   stats.Running,
		"completed": stats.Completed,
		"failed":    stats.Failed,
		"pending":   stats.Pending,
		"cancelled": stats.Cancelled,
	}

	return result, nil
}

// CancelAnalysis 取消分析任务
func (s *RequirementAnalysisService) CancelAnalysis(taskID uint) error {
	var task nesma.NesmaRequirementAnalysisTask
	err := global.GVA_DB.First(&task, taskID).Error
	if err != nil {
		return err
	}
	
	if task.Status == "completed" || task.Status == "cancelled" {
		return errors.New("任务已完成或已取消，无法取消")
	}
	
	task.Status = "cancelled"
	task.Summary = "任务已被用户取消"
	
	return global.GVA_DB.Save(&task).Error
}

// AnalysisResult 分析结果结构
type AnalysisResult struct {
	RequirementID uint
	Success       bool
	Error         error
	UpdatedReq    nesma.NesmaRequirement
}

// analysisWorker 分析工作协程
func (s *RequirementAnalysisService) analysisWorker(workerID int, taskChan <-chan nesma.NesmaRequirement, resultChan chan<- *AnalysisResult, wg *sync.WaitGroup) {
	defer wg.Done()
	
	// 每个工作协程创建独立的AI服务实例
	aiService := NewDeepSeekService(global.GVA_CONFIG.AI.DeepSeek.APIKey, global.GVA_CONFIG.AI.DeepSeek.BaseURL, "deepseek-chat")
	if aiService == nil {
		global.GVA_LOG.Error("工作协程AI服务初始化失败", zap.Int("workerID", workerID))
		return
	}
	
	// 获取知识库实例
	knowledgeService := &KnowledgeService{}
	
	global.GVA_LOG.Info("分析工作协程启动", zap.Int("workerID", workerID))
	
	for req := range taskChan {
		result := &AnalysisResult{
			RequirementID: req.ID,
			UpdatedReq:    req,
		}
		
		global.GVA_LOG.Debug("工作协程开始处理需求", 
			zap.Int("workerID", workerID),
			zap.Uint("requirementID", req.ID),
			zap.String("title", req.Title))
		
		// 构建NESMA分析提示词
		prompt := s.buildNESMAAnalysisPrompt(req, knowledgeService)
		
		// 调用AI进行分析
		aiResponse, err := s.callAIForRequirement(aiService, prompt)
		if err != nil {
			global.GVA_LOG.Error("AI分析失败", 
				zap.Int("workerID", workerID),
				zap.Error(err), 
				zap.String("title", req.Title))
			result.Success = false
			result.Error = err
			s.setDefaultAIValues(&result.UpdatedReq)
		} else {
			// 解析AI响应并更新需求
			s.parseAIResponseToRequirement(&result.UpdatedReq, aiResponse)
			result.Success = true
		}
		
		// 设置分析完成状态
		result.UpdatedReq.AIAnalysisStatus = "completed"
		result.UpdatedReq.Status = "completed"
		analysisTime := time.Now()
		result.UpdatedReq.AIAnalysisTime = &analysisTime
		
		global.GVA_LOG.Debug("工作协程完成需求处理", 
			zap.Int("workerID", workerID),
			zap.Uint("requirementID", req.ID),
			zap.Bool("success", result.Success))
		
		resultChan <- result
	}
	
	global.GVA_LOG.Info("分析工作协程结束", zap.Int("workerID", workerID))
}

// updateAnalysisProgress 批量更新分析任务进度
func (s *RequirementAnalysisService) updateAnalysisProgress(task *nesma.NesmaRequirementAnalysisTask, results []*AnalysisResult) {
	successCount := 0
	failedCount := 0
	aiCallCount := 0
	
	// 分别处理每个需求，避免事务中一个失败影响全部
	for _, result := range results {
		if result.Success {
			successCount++
			aiCallCount++
		} else {
			failedCount++
		}
		
		// 单独保存每个需求，避免事务级联失败
		if err := s.saveRequirementSafely(&result.UpdatedReq); err != nil {
			global.GVA_LOG.Error("保存需求失败", 
				zap.Error(err),
				zap.Uint("requirementID", result.RequirementID))
			failedCount++
		} else {
			global.GVA_LOG.Debug("需求保存成功",
				zap.Uint("requirementID", result.RequirementID),
				zap.String("title", result.UpdatedReq.Title))
		}
	}
	
	// 更新任务统计
	task.SuccessCount += successCount
	task.FailedCount += failedCount
	task.AICallCount += aiCallCount
	task.ProcessedCount += len(results)
	task.UpdateProgress()
	
	// 单独保存任务状态
	if err := global.GVA_DB.Save(task).Error; err != nil {
		global.GVA_LOG.Error("保存任务状态失败", zap.Error(err))
	}
	
	global.GVA_LOG.Info("批量更新分析任务进度完成", 
		zap.Uint("taskId", task.ID),
		zap.Int("batchSize", len(results)),
		zap.Int("successCount", successCount),
		zap.Int("failedCount", failedCount),
		zap.Int("processedCount", task.ProcessedCount),
		zap.Int("totalCount", task.TotalCount),
		zap.Int("progress", task.Progress))
}

// saveRequirementSafely 安全保存需求，避免字段错误
func (s *RequirementAnalysisService) saveRequirementSafely(req *nesma.NesmaRequirement) error {
	// 验证和清理字段
	if req.AIDescription == "" {
		req.AIDescription = req.Description
	}
	if req.AIGeneratedTitle == "" {
		req.AIGeneratedTitle = req.Title
	}
	if req.FunctionType == "" {
		req.FunctionType = "EQ"
	}
	
	// 确保数值字段有效
	if req.AIComplexityScore != nil && (*req.AIComplexityScore < 0 || *req.AIComplexityScore > 100) {
		defaultScore := 50.0
		req.AIComplexityScore = &defaultScore
	}
	if req.AIConfidenceScore != nil && (*req.AIConfidenceScore < 0 || *req.AIConfidenceScore > 1) {
		defaultConfidence := 0.8
		req.AIConfidenceScore = &defaultConfidence
	}
	
	// 使用独立事务保存
	return global.GVA_DB.Save(req).Error
}
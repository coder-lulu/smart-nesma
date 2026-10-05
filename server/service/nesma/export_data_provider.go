package nesma

import (
	"errors"
	"time"
	"strconv"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ExportDataProvider 导出数据提供器
type ExportDataProvider struct {
	db *gorm.DB
}

// NewExportDataProvider 创建数据提供器
func NewExportDataProvider() *ExportDataProvider {
	return &ExportDataProvider{
		db: global.GVA_DB,
	}
}

// GetProjectExportData 获取项目导出数据
func (p *ExportDataProvider) GetProjectExportData(req *ExportRequest) (*ProjectExportData, error) {
	global.GVA_LOG.Info("开始获取项目导出数据", 
		zap.Uint("projectId", req.ProjectID),
		zap.Any("cycleId", req.CycleID),
		zap.Any("versionId", req.VersionID),
		zap.String("exportType", string(req.ExportType)))

	// 1. 获取项目信息
	projectInfo, err := p.getProjectInfo(req.ProjectID)
	if err != nil {
		return nil, NewExportError(ErrProjectNotFound, "项目不存在", err.Error())
	}

	// 2. 获取周期信息（如果指定）
	var cycleInfo *CycleInfo
	if req.CycleID != nil {
		cycleInfo, err = p.getCycleInfo(*req.CycleID, req.ProjectID)
		if err != nil {
			return nil, NewExportError(ErrCycleNotFound, "项目周期不存在", err.Error())
		}
	}

	// 3. 获取版本信息（如果指定）
	var versionInfo *VersionInfo
	if req.VersionID != nil {
		versionInfo, err = p.getVersionInfo(*req.VersionID, req.CycleID)
		if err != nil {
			return nil, NewExportError(ErrVersionNotFound, "需求版本不存在", err.Error())
		}
	}

	// 4. 获取需求列表
	requirements, err := p.getRequirements(req.ProjectID, req.CycleID, req.VersionID)
	if err != nil {
		return nil, NewExportError(ErrDataGeneration, "获取需求数据失败", err.Error())
	}

	// 5. 计算统计信息
	statistics := p.calculateStatistics(requirements)

	// 6. 构建导出数据
	exportData := &ProjectExportData{
		Project:      *projectInfo,
		Cycle:        cycleInfo,
		Version:      versionInfo,
		Requirements: requirements,
		Statistics:   statistics,
		GeneratedAt:  time.Now(),
	}

	global.GVA_LOG.Info("项目导出数据获取完成",
		zap.Int("需求数量", len(requirements)),
		zap.Int("总UFP", int(statistics.TotalUFP)),
		zap.Int("总AFP", int(statistics.TotalAFP)))

	return exportData, nil
}

// getProjectInfo 获取项目信息
func (p *ExportDataProvider) getProjectInfo(projectID uint) (*ProjectInfo, error) {
	var project nesma.NesmaProject
	if err := p.db.First(&project, projectID).Error; err != nil {
		return nil, err
	}

	// TODO: 获取项目负责人名字，这里暂时使用ID
	ownerName := "用户" + strconv.Itoa(int(project.OwnerID))

	return &ProjectInfo{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Domain:      project.Domain,
		Status:      project.Status,
		Owner:       ownerName,
		StartDate:   project.StartDate,
		EndDate:     project.EndDate,
		Settings:    project.Settings,
	}, nil
}

// getCycleInfo 获取周期信息
func (p *ExportDataProvider) getCycleInfo(cycleID uint, projectID uint) (*CycleInfo, error) {
	var cycle nesma.NesmaProjectCycle
	if err := p.db.Where("id = ? AND project_id = ?", cycleID, projectID).First(&cycle).Error; err != nil {
		return nil, err
	}

	return &CycleInfo{
		ID:          cycle.ID,
		Name:        cycle.Name,
		Description: cycle.Description,
		Status:      cycle.Status,
		StartDate:   cycle.StartDate,
		EndDate:     cycle.EndDate,
	}, nil
}

// getVersionInfo 获取版本信息
func (p *ExportDataProvider) getVersionInfo(versionID uint, cycleID *uint) (*VersionInfo, error) {
	var version nesma.NesmaRequirementVersion
	query := p.db.Where("id = ?", versionID)
	if cycleID != nil {
		query = query.Where("cycle_id = ?", *cycleID)
	}
	
	if err := query.First(&version).Error; err != nil {
		return nil, err
	}

	// TODO: 获取用户名，这里暂时使用ID  
	createdBy := "用户" + version.CreatedBy

	return &VersionInfo{
		ID:          version.ID,
		Version:     version.Version,
		VersionType: version.VersionType,
		Summary:     version.Summary,
		CreatedBy:   createdBy,
		CreatedAt:   version.CreatedAt,
	}, nil
}

// getRequirements 获取需求列表
func (p *ExportDataProvider) getRequirements(projectID uint, cycleID, versionID *uint) ([]RequirementInfo, error) {
	var requirements []nesma.NesmaRequirement
	
	// 构建查询条件
	query := p.db.Where("project_id = ?", projectID)
	
	// 如果指定了周期和版本，则精确筛选
	if cycleID != nil && versionID != nil {
		query = query.Where("cycle_id = ? AND version_id = ?", *cycleID, *versionID)
	} else if cycleID != nil {
		// 只指定周期，获取该周期的最新版本需求
		query = query.Where("cycle_id = ?", *cycleID)
	}
	// 如果都没指定，获取项目的所有需求

	// 执行查询，按层级和排序索引排序
	if err := query.Order("level ASC, order_index ASC, id ASC").Find(&requirements).Error; err != nil {
		return nil, err
	}

	global.GVA_LOG.Info("查询到需求数据", 
		zap.Int("数量", len(requirements)),
		zap.String("查询条件", "project_id="+strconv.Itoa(int(projectID))))

	// 转换为导出格式，建立层级关系
	return p.convertToRequirementInfo(requirements), nil
}

// convertToRequirementInfo 转换为需求信息格式，处理层级关系
func (p *ExportDataProvider) convertToRequirementInfo(requirements []nesma.NesmaRequirement) []RequirementInfo {
	var result []RequirementInfo
	
	// 建立ID到需求的映射，便于查找父需求
	reqMap := make(map[uint]*nesma.NesmaRequirement)
	for i := range requirements {
		reqMap[requirements[i].ID] = &requirements[i]
	}

	for _, req := range requirements {
		// 构建层级标题路径
		level1Title, level2Title, level3Title, level4Title := p.buildLevelTitles(&req, reqMap)
		
		reqInfo := RequirementInfo{
			ID:                req.ID,
			Code:              req.Code,
			Title:             req.Title,
			Description:       req.Description,
			Level:             req.Level,
			ParentID:          req.ParentID,
			OrderIndex:        req.OrderIndex,
			
			// 层级信息
			Level1Title:       level1Title,
			Level2Title:       level2Title,
			Level3Title:       level3Title,
			Level4Title:       level4Title,
			
			// NESMA相关
			FunctionType:      req.FunctionType,
			Complexity:        req.Complexity,
			ComplexityLevel:   req.ComplexityLevel,
			UFP:               req.UFP,
			AFP:               req.AFP,
			ReuseLevel:        req.ReuseLevel,
			ModificationType:  req.ModificationType,
			
			// AI分析相关
			AIDescription:     req.AIDescription,
			AIGeneratedTitle:  req.AIGeneratedTitle,
			AIComplexityScore: req.AIComplexityScore,
			AIConfidenceScore: req.AIConfidenceScore,
			AIAnalysisTime:    req.AIAnalysisTime,
			MermaidDiagram:    req.MermaidDiagram,
			
			// 业务相关
			BusinessValue:     req.BusinessValue,
			AcceptanceCriteria: req.AcceptanceCriteria,
			Priority:          req.Priority,
			Status:            req.Status,
			Category:          req.Category,
			
			// 工时相关
			EstimateHours:     req.EstimateHours,
			ActualHours:       req.ActualHours,
			Notes:             req.Notes,
		}
		
		result = append(result, reqInfo)
	}

	return result
}

// buildLevelTitles 构建层级标题
func (p *ExportDataProvider) buildLevelTitles(req *nesma.NesmaRequirement, reqMap map[uint]*nesma.NesmaRequirement) (string, string, string, string) {
	var level1, level2, level3, level4 string
	
	// 向上遍历找到所有父级
	current := req
	levelTitles := make(map[int]string)
	
	for current != nil {
		levelTitles[current.Level] = current.Title
		
		if current.ParentID == nil {
			break
		}
		
		parent, exists := reqMap[*current.ParentID]
		if !exists {
			break
		}
		current = parent
	}
	
	// 提取各级标题
	if title, exists := levelTitles[1]; exists {
		level1 = title
	}
	if title, exists := levelTitles[2]; exists {
		level2 = title
	}
	if title, exists := levelTitles[3]; exists {
		level3 = title
	}
	if title, exists := levelTitles[4]; exists {
		level4 = title
	}
	
	return level1, level2, level3, level4
}

// calculateStatistics 计算统计信息
func (p *ExportDataProvider) calculateStatistics(requirements []RequirementInfo) StatisticsInfo {
	stats := StatisticsInfo{
		AdjustmentFactor: 1.0, // 默认调整因子
	}

	// 基础计数
	stats.TotalRequirements = len(requirements)
	
	var aiAnalyzedCount int
	var totalConfidence float64
	var confidenceCount int

	for _, req := range requirements {
		// 按层级统计
		switch req.Level {
		case 1:
			stats.Level1Count++
		case 2:
			stats.Level2Count++
		case 3:
			stats.Level3Count++
		case 4:
			stats.Level4Count++
		}

		// 按功能类型统计（只统计4级需求）
		if req.Level == 4 && req.FunctionType != "" {
			switch strings.ToUpper(req.FunctionType) {
			case "EI":
				stats.EICount++
				stats.EIUFP += req.UFP
			case "EO":
				stats.EOCount++
				stats.EOUFP += req.UFP
			case "EQ":
				stats.EQCount++
				stats.EQUFP += req.UFP
			case "ILF":
				stats.ILFCount++
				stats.ILFUFP += req.UFP
			case "EIF":
				stats.EIFCount++
				stats.EIFUFP += req.UFP
			}
		}

		// 复杂度统计
		switch strings.ToLower(req.Complexity) {
		case "简单", "low", "simple":
			stats.SimpleCount++
		case "中等", "average", "medium":
			stats.AverageCount++
		case "复杂", "high", "complex":
			stats.ComplexCount++
		}

		// UFP/AFP统计
		stats.TotalUFP += req.UFP
		stats.TotalAFP += req.AFP

		// 工时统计
		if req.EstimateHours != nil {
			stats.TotalEstimateHours += *req.EstimateHours
		}
		if req.ActualHours != nil {
			stats.TotalActualHours += *req.ActualHours
		}

		// AI分析统计
		if req.AIDescription != "" || req.AIAnalysisTime != nil {
			aiAnalyzedCount++
		}
		if req.AIConfidenceScore != nil {
			totalConfidence += *req.AIConfidenceScore
			confidenceCount++
		}
	}

	// 计算AI分析统计
	stats.AIAnalyzedCount = aiAnalyzedCount
	if confidenceCount > 0 {
		stats.AvgConfidenceScore = totalConfidence / float64(confidenceCount)
	}

	// 如果总UFP为0但总AFP不为0，计算调整因子
	if stats.TotalUFP > 0 && stats.TotalAFP > 0 {
		stats.AdjustmentFactor = stats.TotalAFP / stats.TotalUFP
	}

	global.GVA_LOG.Info("统计信息计算完成",
		zap.Int("总需求数", stats.TotalRequirements),
		zap.Int("功能点数", stats.Level4Count),
		zap.Float64("总UFP", stats.TotalUFP),
		zap.Float64("总AFP", stats.TotalAFP),
		zap.Int("AI分析数", stats.AIAnalyzedCount))

	return stats
}

// ValidateExportRequest 验证导出请求
func (p *ExportDataProvider) ValidateExportRequest(req *ExportRequest) error {
	// 1. 验证导出类型
	if !req.ExportType.IsValid() {
		return NewExportError(ErrInvalidExportType, "不支持的导出类型: "+string(req.ExportType))
	}

	// 2. 验证项目存在
	var project nesma.NesmaProject
	if err := p.db.First(&project, req.ProjectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NewExportError(ErrProjectNotFound, "项目不存在")
		}
		return NewExportError(ErrProjectNotFound, "查询项目失败", err.Error())
	}

	// 3. 验证周期存在（如果指定）
	if req.CycleID != nil {
		var cycle nesma.NesmaProjectCycle
		if err := p.db.Where("id = ? AND project_id = ?", *req.CycleID, req.ProjectID).First(&cycle).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewExportError(ErrCycleNotFound, "项目周期不存在")
			}
			return NewExportError(ErrCycleNotFound, "查询项目周期失败", err.Error())
		}
	}

	// 4. 验证版本存在（如果指定）
	if req.VersionID != nil {
		var version nesma.NesmaRequirementVersion
		query := p.db.Where("id = ?", *req.VersionID)
		if req.CycleID != nil {
			query = query.Where("cycle_id = ?", *req.CycleID)
		}
		
		if err := query.First(&version).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return NewExportError(ErrVersionNotFound, "需求版本不存在")
			}
			return NewExportError(ErrVersionNotFound, "查询需求版本失败", err.Error())
		}
	}

	return nil
}
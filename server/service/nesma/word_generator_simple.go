package nesma

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// WordRequirementSpecGeneratorSimple 简化的Word需求规格书生成器
type WordRequirementSpecGeneratorSimple struct {
	storePath string
}

// NewWordRequirementSpecGenerator 创建Word生成器
func NewWordRequirementSpecGenerator() *WordRequirementSpecGeneratorSimple {
	return &WordRequirementSpecGeneratorSimple{
		storePath: filepath.Join(global.GVA_CONFIG.Local.StorePath, "exports"),
	}
}

// GetSupportedType 获取支持的导出类型
func (g *WordRequirementSpecGeneratorSimple) GetSupportedType() ExportType {
	return ExportWordRequirementSpec
}

// Validate 验证数据
func (g *WordRequirementSpecGeneratorSimple) Validate(data *ProjectExportData) error {
	if data == nil {
		return NewExportError(ErrDataGeneration, "导出数据为空")
	}
	
	if data.Project.Name == "" {
		return NewExportError(ErrDataGeneration, "项目名称不能为空")
	}
	
	if len(data.Requirements) == 0 {
		return NewExportError(ErrDataGeneration, "项目没有需求数据")
	}
	
	return nil
}

// Generate 生成Word文档（暂时以文本格式实现）
func (g *WordRequirementSpecGeneratorSimple) Generate(data *ProjectExportData, config ExportConfig) (string, error) {
	global.GVA_LOG.Info("开始生成Word需求规格书", 
		zap.String("项目", data.Project.Name),
		zap.Int("需求数量", len(data.Requirements)))

	// 确保目录存在
	if err := os.MkdirAll(g.storePath, 0755); err != nil {
		return "", NewExportError(ErrFileGeneration, "创建目录失败", err.Error())
	}

	// 生成文件名
	fileName := fmt.Sprintf("%s_需求规格说明书_%s.txt", 
		g.sanitizeFileName(data.Project.Name),
		time.Now().Format("20060102150405"))
	
	filePath := filepath.Join(g.storePath, fileName)

	// 生成文本内容
	content := g.generateTextContent(data, config)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", NewExportError(ErrFileGeneration, "保存文档失败", err.Error())
	}

	global.GVA_LOG.Info("Word需求规格书生成完成（文本格式）", 
		zap.String("文件路径", filePath))

	return filePath, nil
}

// generateTextContent 生成文本内容
func (g *WordRequirementSpecGeneratorSimple) generateTextContent(data *ProjectExportData, config ExportConfig) string {
	var content strings.Builder

	// 文档标题
	content.WriteString("==========================================\n")
	content.WriteString(fmt.Sprintf("      %s 需求规格说明书\n", data.Project.Name))
	content.WriteString("==========================================\n\n")

	// 项目基本信息
	content.WriteString("项目基本信息:\n")
	content.WriteString("----------------------------------------\n")
	content.WriteString(fmt.Sprintf("项目名称: %s\n", data.Project.Name))
	content.WriteString(fmt.Sprintf("项目领域: %s\n", data.Project.Domain))
	content.WriteString(fmt.Sprintf("项目状态: %s\n", data.Project.Status))
	content.WriteString(fmt.Sprintf("项目负责人: %s\n", data.Project.Owner))
	content.WriteString(fmt.Sprintf("生成时间: %s\n", data.GeneratedAt.Format("2006-01-02 15:04:05")))
	
	if data.Cycle != nil {
		content.WriteString(fmt.Sprintf("建设周期: %s\n", data.Cycle.Name))
	}
	if data.Version != nil {
		content.WriteString(fmt.Sprintf("需求版本: %s\n", data.Version.Version))
	}
	content.WriteString("\n")

	// 统计信息
	if config.IncludeStatistics {
		content.WriteString("需求统计信息:\n")
		content.WriteString("----------------------------------------\n")
		stats := data.Statistics
		content.WriteString(fmt.Sprintf("总需求数: %d\n", stats.TotalRequirements))
		content.WriteString(fmt.Sprintf("一级模块: %d个\n", stats.Level1Count))
		content.WriteString(fmt.Sprintf("二级模块: %d个\n", stats.Level2Count))
		content.WriteString(fmt.Sprintf("三级模块: %d个\n", stats.Level3Count))
		content.WriteString(fmt.Sprintf("功能点: %d个\n", stats.Level4Count))
		content.WriteString("\n")

		content.WriteString("NESMA功能点统计:\n")
		content.WriteString(fmt.Sprintf("EI外部输入: %d个, UFP: %.1f\n", stats.EICount, stats.EIUFP))
		content.WriteString(fmt.Sprintf("EO外部输出: %d个, UFP: %.1f\n", stats.EOCount, stats.EOUFP))
		content.WriteString(fmt.Sprintf("EQ外部查询: %d个, UFP: %.1f\n", stats.EQCount, stats.EQUFP))
		content.WriteString(fmt.Sprintf("ILF内部文件: %d个, UFP: %.1f\n", stats.ILFCount, stats.ILFUFP))
		content.WriteString(fmt.Sprintf("EIF外部文件: %d个, UFP: %.1f\n", stats.EIFCount, stats.EIFUFP))
		content.WriteString(fmt.Sprintf("总计UFP: %.1f\n", stats.TotalUFP))
		content.WriteString(fmt.Sprintf("调整因子: %.2f\n", stats.AdjustmentFactor))
		content.WriteString(fmt.Sprintf("总计AFP: %.1f\n", stats.TotalAFP))
		content.WriteString("\n")
	}

	// 需求详细列表
	content.WriteString("需求详细列表:\n")
	content.WriteString("==========================================\n")

	// 按层级组织需求
	g.generateRequirementsByLevel(data.Requirements, &content, config, 1)

	// 附录
	content.WriteString("\n\n附录:\n")
	content.WriteString("----------------------------------------\n")
	content.WriteString("术语表:\n")
	content.WriteString("UFP - 未调整功能点 (Unadjusted Function Points)\n")
	content.WriteString("AFP - 调整后功能点 (Adjusted Function Points)\n")
	content.WriteString("EI - 外部输入 (External Input)\n")
	content.WriteString("EO - 外部输出 (External Output)\n")
	content.WriteString("EQ - 外部查询 (External Inquiry)\n")
	content.WriteString("ILF - 内部逻辑文件 (Internal Logical File)\n")
	content.WriteString("EIF - 外部接口文件 (External Interface File)\n")

	return content.String()
}

// generateRequirementsByLevel 按层级生成需求
func (g *WordRequirementSpecGeneratorSimple) generateRequirementsByLevel(requirements []RequirementInfo, content *strings.Builder, config ExportConfig, startIndex int) {
	// 获取一级需求
	level1Reqs := g.getRequirementsByLevel(requirements, 1)

	for i, level1 := range level1Reqs {
		// 一级需求
		content.WriteString(fmt.Sprintf("\n%d. %s\n", startIndex+i, level1.Title))
		if level1.Description != "" {
			content.WriteString(fmt.Sprintf("   描述: %s\n", level1.Description))
		}
		if config.IncludeAIAnalysis && level1.AIDescription != "" {
			content.WriteString(fmt.Sprintf("   AI优化描述: %s\n", level1.AIDescription))
		}

		// 二级需求
		level2Reqs := g.getChildRequirements(requirements, level1.ID, 2)
		for j, level2 := range level2Reqs {
			content.WriteString(fmt.Sprintf("   %d.%d %s\n", startIndex+i, j+1, level2.Title))
			if level2.Description != "" {
				content.WriteString(fmt.Sprintf("      描述: %s\n", level2.Description))
			}
			if config.IncludeAIAnalysis && level2.AIDescription != "" {
				content.WriteString(fmt.Sprintf("      AI优化描述: %s\n", level2.AIDescription))
			}

			// 三级需求
			level3Reqs := g.getChildRequirements(requirements, level2.ID, 3)
			for k, level3 := range level3Reqs {
				content.WriteString(fmt.Sprintf("      %d.%d.%d %s\n", startIndex+i, j+1, k+1, level3.Title))
				if level3.Description != "" {
					content.WriteString(fmt.Sprintf("         描述: %s\n", level3.Description))
				}
				if config.IncludeAIAnalysis && level3.AIDescription != "" {
					content.WriteString(fmt.Sprintf("         AI优化描述: %s\n", level3.AIDescription))
				}

				// 四级需求（功能点）
				level4Reqs := g.getChildRequirements(requirements, level3.ID, 4)
				for l, level4 := range level4Reqs {
					content.WriteString(fmt.Sprintf("         %d.%d.%d.%d %s\n", startIndex+i, j+1, k+1, l+1, level4.Title))
					if level4.Description != "" {
						content.WriteString(fmt.Sprintf("            描述: %s\n", level4.Description))
					}
					if config.IncludeAIAnalysis && level4.AIDescription != "" {
						content.WriteString(fmt.Sprintf("            AI优化描述: %s\n", level4.AIDescription))
					}
					
					// 功能点信息
					if level4.FunctionType != "" {
						content.WriteString(fmt.Sprintf("            功能类型: %s\n", level4.FunctionType))
					}
					if level4.Complexity != "" {
						content.WriteString(fmt.Sprintf("            复杂度: %s\n", level4.Complexity))
					}
					if level4.UFP > 0 {
						content.WriteString(fmt.Sprintf("            UFP: %.1f\n", level4.UFP))
					}
					if level4.AFP > 0 {
						content.WriteString(fmt.Sprintf("            AFP: %.1f\n", level4.AFP))
					}
					if level4.AIConfidenceScore != nil {
						content.WriteString(fmt.Sprintf("            AI置信度: %.1f%%\n", *level4.AIConfidenceScore*100))
					}

					// Mermaid流程图
					if config.IncludeMermaidDiag && level4.MermaidDiagram != "" {
						content.WriteString("            流程图:\n")
						content.WriteString(fmt.Sprintf("            %s\n", level4.MermaidDiagram))
					}
				}
			}
		}
	}
}

// getRequirementsByLevel 获取指定层级的需求
func (g *WordRequirementSpecGeneratorSimple) getRequirementsByLevel(requirements []RequirementInfo, level int) []RequirementInfo {
	var result []RequirementInfo
	for _, req := range requirements {
		if req.Level == level {
			result = append(result, req)
		}
	}
	return result
}

// getChildRequirements 获取子需求
func (g *WordRequirementSpecGeneratorSimple) getChildRequirements(requirements []RequirementInfo, parentID uint, level int) []RequirementInfo {
	var result []RequirementInfo
	for _, req := range requirements {
		if req.Level == level && req.ParentID != nil && *req.ParentID == parentID {
			result = append(result, req)
		}
	}
	return result
}

// sanitizeFileName 清理文件名
func (g *WordRequirementSpecGeneratorSimple) sanitizeFileName(name string) string {
	replacements := map[string]string{
		"/": "_", "\\": "_", ":": "_", "*": "_", "?": "_",
		"\"": "_", "<": "_", ">": "_", "|": "_",
	}
	
	result := name
	for old, new := range replacements {
		result = strings.Replace(result, old, new, -1)
	}
	
	return result
}
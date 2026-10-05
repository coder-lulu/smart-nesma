package nesma

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaRes "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"go.uber.org/zap"
)

type TemplateService struct{}

// CreateTemplate 创建文档模板
func (s *TemplateService) CreateTemplate(req *nesmaReq.CreateNesmaDocTemplateRequest, userID uint) (*nesma.NesmaDocTemplate, error) {
	// 检查模板名称是否重复
	var count int64
	global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).Where("name = ? AND type = ? AND format = ?",
		req.Name, req.Type, req.Format).Count(&count)
	if count > 0 {
		return nil, errors.New("相同类型和格式的模板名称已存在")
	}

	// 如果设置为默认模板，将其他同类型模板设为非默认
	if req.IsDefault {
		global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).
			Where("type = ? AND format = ? AND is_default = true", req.Type, req.Format).
			Update("is_default", false)
	}

	// 创建模板
	template := nesma.NesmaDocTemplate{
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Format:       req.Format,
		Category:     req.Category,
		TemplatePath: req.TemplatePath,
		PreviewPath:  req.PreviewPath,
		Config:       req.Config,
		Variables:    req.Variables,
		IsDefault:    req.IsDefault,
		IsActive:     req.IsActive,
		UsageCount:   0,
		CreatedBy:    userID,
	}

	if err := global.GVA_DB.Create(&template).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

// UpdateTemplate 更新文档模板
func (s *TemplateService) UpdateTemplate(req *nesmaReq.UpdateNesmaDocTemplateRequest) (*nesma.NesmaDocTemplate, error) {
	var template nesma.NesmaDocTemplate
	if err := global.GVA_DB.First(&template, req.ID).Error; err != nil {
		return nil, errors.New("模板不存在")
	}

	// 检查模板名称是否重复（排除自己）
	var count int64
	global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).
		Where("name = ? AND type = ? AND format = ? AND id != ?",
			req.Name, req.Type, req.Format, req.ID).Count(&count)
	if count > 0 {
		return nil, errors.New("相同类型和格式的模板名称已存在")
	}

	// 如果设置为默认模板，将其他同类型模板设为非默认
	if req.IsDefault && !template.IsDefault {
		global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).
			Where("type = ? AND format = ? AND is_default = true AND id != ?",
				req.Type, req.Format, req.ID).
			Update("is_default", false)
	}

	// 更新模板
	updates := map[string]interface{}{
		"name":          req.Name,
		"description":   req.Description,
		"type":          req.Type,
		"format":        req.Format,
		"category":      req.Category,
		"template_path": req.TemplatePath,
		"preview_path":  req.PreviewPath,
		"config":        req.Config,
		"variables":     req.Variables,
		"is_default":    req.IsDefault,
		"is_active":     req.IsActive,
	}

	if err := global.GVA_DB.Model(&template).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新查询获取更新后的数据
	if err := global.GVA_DB.First(&template, req.ID).Error; err != nil {
		return nil, err
	}

	return &template, nil
}

// DeleteTemplate 删除文档模板
func (s *TemplateService) DeleteTemplate(id uint) error {
	var template nesma.NesmaDocTemplate
	if err := global.GVA_DB.First(&template, id).Error; err != nil {
		return errors.New("模板不存在")
	}

	// 检查是否有文档在使用此模板
	var docCount int64
	global.GVA_DB.Model(&nesma.NesmaDocument{}).Where("template_id = ?", id).Count(&docCount)
	if docCount > 0 {
		return errors.New("模板正在被使用，无法删除")
	}

	// 删除模板文件
	if template.TemplatePath != "" {
		if err := os.Remove(template.TemplatePath); err != nil {
			global.GVA_LOG.Warn("删除模板文件失败", zap.String("filePath", template.TemplatePath), zap.Error(err))
		}
	}

	// 删除预览图
	if template.PreviewPath != "" {
		if err := os.Remove(template.PreviewPath); err != nil {
			global.GVA_LOG.Warn("删除预览图失败", zap.String("filePath", template.PreviewPath), zap.Error(err))
		}
	}

	return global.GVA_DB.Delete(&template).Error
}

// GetTemplate 获取模板详情
func (s *TemplateService) GetTemplate(id uint) (*nesmaRes.NesmaDocTemplateResponse, error) {
	var template nesma.NesmaDocTemplate
	if err := global.GVA_DB.First(&template, id).Error; err != nil {
		return nil, errors.New("模板不存在")
	}

	response := &nesmaRes.NesmaDocTemplateResponse{
		NesmaDocTemplate: template,
		TypeText:         s.getTypeText(template.Type),
		FormatText:       s.getFormatText(template.Format),
		CategoryText:     template.Category,
	}

	return response, nil
}

// GetTemplateList 获取模板列表
func (s *TemplateService) GetTemplateList(req *nesmaReq.NesmaDocTemplateSearch) (*nesmaRes.NesmaDocTemplateListResponse, error) {
	var templates []nesma.NesmaDocTemplate
	var total int64

	db := global.GVA_DB.Model(&nesma.NesmaDocTemplate{})

	// 条件筛选
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Format != "" {
		db = db.Where("format = ?", req.Format)
	}
	if req.Category != "" {
		db = db.Where("category = ?", req.Category)
	}
	if req.IsDefault != nil {
		db = db.Where("is_default = ?", *req.IsDefault)
	}
	if req.IsActive != nil {
		db = db.Where("is_active = ?", *req.IsActive)
	}
	if req.CreatedBy != nil {
		db = db.Where("created_by = ?", *req.CreatedBy)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 获取总数
	db.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := db.Order("is_default DESC, usage_count DESC, created_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&templates).Error; err != nil {
		return nil, err
	}

	// 转换响应格式
	var list []nesmaRes.NesmaDocTemplateResponse
	for _, template := range templates {
		list = append(list, nesmaRes.NesmaDocTemplateResponse{
			NesmaDocTemplate: template,
			TypeText:         s.getTypeText(template.Type),
			FormatText:       s.getFormatText(template.Format),
			CategoryText:     template.Category,
		})
	}

	return &nesmaRes.NesmaDocTemplateListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetTemplateOptions 获取模板选项
func (s *TemplateService) GetTemplateOptions(docType, format string) (*nesmaRes.TemplateOptionsResponse, error) {
	var templates []nesma.NesmaDocTemplate

	db := global.GVA_DB.Where("is_active = true")
	if docType != "" {
		db = db.Where("type = ?", docType)
	}
	if format != "" {
		db = db.Where("format = ?", format)
	}

	if err := db.Order("is_default DESC, usage_count DESC").Find(&templates).Error; err != nil {
		return nil, err
	}

	var options []nesmaRes.TemplateOption
	for _, template := range templates {
		options = append(options, nesmaRes.TemplateOption{
			ID:          template.ID,
			Name:        template.Name,
			Description: template.Description,
			Type:        template.Type,
			Format:      template.Format,
			Category:    template.Category,
			IsDefault:   template.IsDefault,
		})
	}

	return &nesmaRes.TemplateOptionsResponse{
		Options: options,
	}, nil
}

// GetTemplateVariables 获取模板变量
func (s *TemplateService) GetTemplateVariables(templateID uint) (*nesmaRes.TemplateVariablesResponse, error) {
	var template nesma.NesmaDocTemplate
	if err := global.GVA_DB.First(&template, templateID).Error; err != nil {
		return nil, errors.New("模板不存在")
	}

	var variables []nesmaRes.TemplateVariable
	if template.Variables != nil {
		if err := json.Unmarshal(template.Variables, &variables); err != nil {
			return nil, errors.New("模板变量格式错误")
		}
	}

	return &nesmaRes.TemplateVariablesResponse{
		Variables: variables,
	}, nil
}

// IncrementUsage 增加模板使用次数
func (s *TemplateService) IncrementUsage(templateID uint) error {
	return global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).
		Where("id = ?", templateID).
		UpdateColumn("usage_count", global.GVA_DB.Raw("usage_count + 1")).Error
}

// SetDefaultTemplate 设置默认模板
func (s *TemplateService) SetDefaultTemplate(templateID uint) error {
	var template nesma.NesmaDocTemplate
	if err := global.GVA_DB.First(&template, templateID).Error; err != nil {
		return errors.New("模板不存在")
	}

	// 将同类型其他模板设为非默认
	if err := global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).
		Where("type = ? AND format = ? AND id != ?", template.Type, template.Format, templateID).
		Update("is_default", false).Error; err != nil {
		return err
	}

	// 设置当前模板为默认
	return global.GVA_DB.Model(&template).Update("is_default", true).Error
}

// ActivateTemplate 激活/停用模板
func (s *TemplateService) ActivateTemplate(templateID uint, isActive bool) error {
	var template nesma.NesmaDocTemplate
	if err := global.GVA_DB.First(&template, templateID).Error; err != nil {
		return errors.New("模板不存在")
	}

	return global.GVA_DB.Model(&template).Update("is_active", isActive).Error
}

// UploadTemplate 上传模板文件
func (s *TemplateService) UploadTemplate(req *nesmaReq.UploadTemplateRequest, filePath string, userID uint) (*nesmaRes.UploadTemplateResponse, error) {
	// 移动文件到模板目录
	templateDir := filepath.Join(global.GVA_CONFIG.Local.StorePath, "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return nil, err
	}

	fileName := filepath.Base(filePath)
	destPath := filepath.Join(templateDir, fileName)
	if err := os.Rename(filePath, destPath); err != nil {
		return nil, err
	}

	// 提取文件内容
	content, err := s.extractFileContent(destPath, req.Type)
	if err != nil {
		global.GVA_LOG.Warn("提取文件内容失败", zap.Error(err))
		// 不因为内容提取失败而中断上传，使用默认内容
		content = s.getDefaultTemplateContent(req.Type, req.Format)
	}

	// 创建模板记录
	template := nesma.NesmaDocTemplate{
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Format:       req.Format,
		Category:     req.Category,
		Content:      content,
		TemplatePath: destPath,
		IsDefault:    req.IsDefault,
		IsActive:     true,
		UsageCount:   0,
		CreatedBy:    userID,
	}

	// 如果设置为默认模板，将其他同类型模板设为非默认
	if req.IsDefault {
		global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).
			Where("type = ? AND format = ? AND is_default = true", req.Type, req.Format).
			Update("is_default", false)
	}

	if err := global.GVA_DB.Create(&template).Error; err != nil {
		// 如果创建失败，删除已上传的文件
		os.Remove(destPath)
		return nil, err
	}

	return &nesmaRes.UploadTemplateResponse{
		TemplateID:   template.ID,
		TemplatePath: destPath,
		Message:      "模板上传成功",
	}, nil
}

// BatchDeleteTemplates 批量删除模板
func (s *TemplateService) BatchDeleteTemplates(req *nesmaReq.TemplateIdsRequest) error {
	// 检查模板是否存在且没有被使用
	for _, id := range req.IDs {
		var template nesma.NesmaDocTemplate
		if err := global.GVA_DB.First(&template, id).Error; err != nil {
			return errors.New(fmt.Sprintf("模板ID %d 不存在", id))
		}

		var docCount int64
		global.GVA_DB.Model(&nesma.NesmaDocument{}).Where("template_id = ?", id).Count(&docCount)
		if docCount > 0 {
			return errors.New(fmt.Sprintf("模板 %s 正在被使用，无法删除", template.Name))
		}
	}

	// 批量删除
	return global.GVA_DB.Where("id IN ?", req.IDs).Delete(&nesma.NesmaDocTemplate{}).Error
}

// 辅助方法
func (s *TemplateService) getTypeText(docType string) string {
	switch docType {
	case "word":
		return "Word文档"
	case "excel":
		return "Excel表格"
	case "pdf":
		return "PDF文档"
	default:
		return "未知类型"
	}
}

func (s *TemplateService) getFormatText(format string) string {
	switch format {
	case "requirement_spec":
		return "需求规格说明书"
	case "nesma_report":
		return "NESMA评估报告"
	case "business_summary":
		return "业务需求汇总表"
	default:
		return "未知格式"
	}
}

// extractFileContent 提取文件内容
func (s *TemplateService) extractFileContent(filePath, fileType string) (string, error) {
	// 暂时返回文件基本信息，后续可以实现具体的内容提取
	// 对于Word、Excel、PDF文件，需要专门的库来提取内容
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", err
	}

	// 根据文件类型返回基本信息
	switch fileType {
	case "word":
		return fmt.Sprintf("Word模板文档\n文件大小: %d 字节\n上传时间: %s\n\n此模板可用于生成Word格式的文档。",
			fileInfo.Size(), fileInfo.ModTime().Format("2006-01-02 15:04:05")), nil
	case "excel":
		return fmt.Sprintf("Excel模板表格\n文件大小: %d 字节\n上传时间: %s\n\n此模板可用于生成Excel格式的表格文档。",
			fileInfo.Size(), fileInfo.ModTime().Format("2006-01-02 15:04:05")), nil
	case "pdf":
		return fmt.Sprintf("PDF模板文档\n文件大小: %d 字节\n上传时间: %s\n\n此模板可用于生成PDF格式的文档。",
			fileInfo.Size(), fileInfo.ModTime().Format("2006-01-02 15:04:05")), nil
	default:
		return "未知格式的模板文件", nil
	}
}

// getDefaultTemplateContent 获取默认模板内容
func (s *TemplateService) getDefaultTemplateContent(fileType, format string) string {
	switch format {
	case "requirement_spec":
		return `需求规格说明书模板

1. 项目概述
   - 项目名称: {{project_name}}
   - 项目描述: {{project_description}}
   - 项目版本: {{project_version}}

2. 功能需求
   - 功能点列表
   - 业务规则
   - 用户界面要求

3. NESMA功能点分析
   - 数据功能点 (ILF/EIF)
   - 事务功能点 (EI/EO/EQ)
   - 复杂度评估

4. 技术要求
   - 技术架构
   - 性能要求
   - 安全要求

5. 验收标准
   - 功能验收
   - 性能验收
   - 安全验收`

	case "nesma_report":
		return `NESMA功能点评估报告

1. 评估概述
   - 评估日期: {{evaluation_date}}
   - 评估人员: {{evaluator}}
   - 项目名称: {{project_name}}

2. 功能点统计
   - 内部逻辑文件 (ILF): {{ilf_count}} 个
   - 外部接口文件 (EIF): {{eif_count}} 个
   - 外部输入 (EI): {{ei_count}} 个
   - 外部输出 (EO): {{eo_count}} 个
   - 外部查询 (EQ): {{eq_count}} 个

3. 复杂度分析
   - 低复杂度: {{low_complexity_count}} 个
   - 中等复杂度: {{medium_complexity_count}} 个
   - 高复杂度: {{high_complexity_count}} 个

4. 功能点总计
   - 未调整功能点: {{unadjusted_fp}} FP
   - 调整因子: {{adjustment_factor}}
   - 调整后功能点: {{adjusted_fp}} FP

5. 评估结论
   - 项目规模评估
   - 工作量估算
   - 建议和说明`

	case "business_summary":
		return `业务需求汇总表

项目信息:
- 项目名称: {{project_name}}
- 业务部门: {{business_department}}
- 项目经理: {{project_manager}}
- 创建日期: {{creation_date}}

需求汇总:
序号 | 需求名称 | 需求描述 | 优先级 | 功能点类型 | 复杂度 | 状态
1    | 用户管理 | 用户注册登录管理 | 高 | EI | 中等 | 待开发
2    | 数据查询 | 业务数据查询展示 | 高 | EO | 中等 | 待开发
3    | 报表生成 | 统计报表生成 | 中 | EO | 高 | 待开发

功能点统计:
- 总功能点数: {{total_function_points}} FP
- 数据功能点: {{data_function_points}} FP
- 事务功能点: {{transaction_function_points}} FP

项目估算:
- 预估工期: {{estimated_duration}} 天
- 预估工作量: {{estimated_effort}} 人天
- 预估成本: {{estimated_cost}} 元`

	default:
		return fmt.Sprintf("默认%s模板内容\n\n这是一个%s格式的模板文件。\n您可以编辑此模板以满足您的需求。",
			s.getTypeText(fileType), s.getFormatText(format))
	}
}

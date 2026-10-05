package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// InitNesmaSampleTemplates 初始化NESMA示例模板
func InitNesmaSampleTemplates() {
	var count int64
	global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).Count(&count)

	// 如果已经有模板，则不重复添加
	if count > 0 {
		global.GVA_LOG.Info("NESMA sample templates already exist, skipping initialization")
		return
	}

	global.GVA_LOG.Info("Initializing NESMA sample templates...")

	templates := []nesma.NesmaDocTemplate{
		{
			Name:        "标准需求规格说明书模板",
			Description: "基于NESMA标准的需求规格说明书模板，包含完整的功能点分析框架",
			Type:        "word",
			Format:      "requirement_spec",
			Category:    "标准模板",
			Content: `需求规格说明书模板

1. 项目概述
   - 项目名称: {{project_name}}
   - 项目描述: {{project_description}}
   - 项目版本: {{project_version}}
   - 项目经理: {{project_manager}}
   - 开发团队: {{development_team}}

2. 功能需求分析
   2.1 业务功能概述
   - 核心业务流程
   - 用户角色定义
   - 系统边界确定

   2.2 功能点识别
   - 数据功能点 (Data Functions)
     * 内部逻辑文件 (ILF): {{ilf_list}}
     * 外部接口文件 (EIF): {{eif_list}}
   
   - 事务功能点 (Transaction Functions)  
     * 外部输入 (EI): {{ei_list}}
     * 外部输出 (EO): {{eo_list}}
     * 外部查询 (EQ): {{eq_list}}

3. NESMA功能点评估
   3.1 复杂度评估标准
   - 低复杂度 (Low): 3-4个权重
   - 中等复杂度 (Average): 4-6个权重  
   - 高复杂度 (High): 6-7个权重

   3.2 功能点统计表
   | 类型 | 低复杂度 | 中等复杂度 | 高复杂度 | 小计 |
   |------|----------|------------|----------|------|
   | ILF  | {{ilf_low}} × 7 | {{ilf_avg}} × 10 | {{ilf_high}} × 15 | {{ilf_total}} |
   | EIF  | {{eif_low}} × 5 | {{eif_avg}} × 7 | {{eif_high}} × 10 | {{eif_total}} |
   | EI   | {{ei_low}} × 3 | {{ei_avg}} × 4 | {{ei_high}} × 6 | {{ei_total}} |
   | EO   | {{eo_low}} × 4 | {{eo_avg}} × 5 | {{eo_high}} × 7 | {{eo_total}} |
   | EQ   | {{eq_low}} × 3 | {{eq_avg}} × 4 | {{eq_high}} × 6 | {{eq_total}} |

   总计未调整功能点: {{total_unadjusted_fp}} FP

4. 技术要求
   - 系统架构: {{system_architecture}}
   - 技术栈: {{technology_stack}}
   - 性能要求: {{performance_requirements}}
   - 安全要求: {{security_requirements}}

5. 验收标准
   - 功能验收标准
   - 性能验收标准
   - 安全验收标准`,
			IsDefault:  true,
			IsActive:   true,
			UsageCount: 0,
			CreatedBy:  1,
		},
		{
			Name:        "NESMA功能点评估报告模板",
			Description: "专业的NESMA功能点评估报告模板，符合国际标准",
			Type:        "excel",
			Format:      "nesma_report",
			Category:    "评估报告",
			Content: `NESMA功能点评估报告

项目信息
========
项目名称: {{project_name}}
评估日期: {{evaluation_date}}
评估人员: {{evaluator_name}}
评估版本: {{evaluation_version}}
项目阶段: {{project_phase}}

评估摘要
========
评估方法: NESMA Function Point Analysis v2.2
评估范围: {{evaluation_scope}}
评估精度: {{evaluation_accuracy}}

功能点统计
==========

1. 数据功能点 (Data Functions)
   内部逻辑文件 (ILF):
   - 低复杂度: {{ilf_low_count}} 个 × 7 FP = {{ilf_low_fp}} FP
   - 中等复杂度: {{ilf_avg_count}} 个 × 10 FP = {{ilf_avg_fp}} FP  
   - 高复杂度: {{ilf_high_count}} 个 × 15 FP = {{ilf_high_fp}} FP
   小计: {{ilf_total_fp}} FP

   外部接口文件 (EIF):
   - 低复杂度: {{eif_low_count}} 个 × 5 FP = {{eif_low_fp}} FP
   - 中等复杂度: {{eif_avg_count}} 个 × 7 FP = {{eif_avg_fp}} FP
   - 高复杂度: {{eif_high_count}} 个 × 10 FP = {{eif_high_fp}} FP
   小计: {{eif_total_fp}} FP

2. 事务功能点 (Transaction Functions)
   外部输入 (EI):
   - 低复杂度: {{ei_low_count}} 个 × 3 FP = {{ei_low_fp}} FP
   - 中等复杂度: {{ei_avg_count}} 个 × 4 FP = {{ei_avg_fp}} FP
   - 高复杂度: {{ei_high_count}} 个 × 6 FP = {{ei_high_fp}} FP
   小计: {{ei_total_fp}} FP

   外部输出 (EO):
   - 低复杂度: {{eo_low_count}} 个 × 4 FP = {{eo_low_fp}} FP
   - 中等复杂度: {{eo_avg_count}} 个 × 5 FP = {{eo_avg_fp}} FP
   - 高复杂度: {{eo_high_count}} 个 × 7 FP = {{eo_high_fp}} FP
   小计: {{eo_total_fp}} FP

   外部查询 (EQ):
   - 低复杂度: {{eq_low_count}} 个 × 3 FP = {{eq_low_fp}} FP
   - 中等复杂度: {{eq_avg_count}} 个 × 4 FP = {{eq_avg_fp}} FP
   - 高复杂度: {{eq_high_count}} 个 × 6 FP = {{eq_high_fp}} FP
   小计: {{eq_total_fp}} FP

功能点汇总
==========
数据功能点总计: {{data_fp_total}} FP ({{data_fp_percentage}}%)
事务功能点总计: {{transaction_fp_total}} FP ({{transaction_fp_percentage}}%)
未调整功能点总计: {{unadjusted_fp_total}} FP

调整因子评估
============
技术复杂度调整因子: {{technical_complexity_factor}}
环境复杂度调整因子: {{environment_complexity_factor}}
综合调整因子: {{overall_adjustment_factor}}

最终结果
========
调整后功能点: {{adjusted_fp_total}} FP
工作量估算: {{effort_estimate}} 人天
开发周期估算: {{duration_estimate}} 天
成本估算: {{cost_estimate}} 元

质量指标
========
识别置信度: {{identification_confidence}}%
评估准确度: {{evaluation_accuracy}}%
合规性评分: {{compliance_score}}/100

建议和说明
==========
{{recommendations_and_notes}}`,
			IsDefault:  true,
			IsActive:   true,
			UsageCount: 0,
			CreatedBy:  1,
		},
		{
			Name:        "业务需求汇总表模板",
			Description: "便于项目管理的业务需求汇总表格模板",
			Type:        "excel",
			Format:      "business_summary",
			Category:    "管理表格",
			Content: `业务需求汇总表

项目基本信息
============
项目编号: {{project_code}}
项目名称: {{project_name}}
所属部门: {{department}}
项目经理: {{project_manager}}
业务负责人: {{business_owner}}
技术负责人: {{technical_lead}}
预计开始日期: {{start_date}}
预计结束日期: {{end_date}}
项目状态: {{project_status}}

需求清单
========
序号 | 需求ID | 需求名称 | 需求描述 | 优先级 | 业务价值 | 功能点类型 | 复杂度 | 预估工时 | 责任人 | 状态 | 备注
-----|--------|----------|----------|--------|----------|------------|--------|----------|--------|------|------
1    | REQ001 | 用户注册 | 系统用户注册功能 | 高 | 核心 | EI | 中 | 8h | 张三 | 开发中 | 包含邮箱验证
2    | REQ002 | 用户登录 | 系统用户登录功能 | 高 | 核心 | EI | 低 | 4h | 李四 | 已完成 | 支持多种登录方式
3    | REQ003 | 用户信息管理 | 用户个人信息维护 | 中 | 重要 | ILF | 中 | 12h | 王五 | 待开发 | 
4    | REQ004 | 数据查询 | 业务数据查询展示 | 高 | 核心 | EO | 中 | 16h | 赵六 | 开发中 | 支持多条件查询
5    | REQ005 | 报表生成 | 统计报表生成导出 | 中 | 重要 | EO | 高 | 24h | 孙七 | 待开发 | 支持多种格式

功能点统计
==========
功能点类型分布:
- 内部逻辑文件 (ILF): {{ilf_count}} 个
- 外部接口文件 (EIF): {{eif_count}} 个  
- 外部输入 (EI): {{ei_count}} 个
- 外部输出 (EO): {{eo_count}} 个
- 外部查询 (EQ): {{eq_count}} 个

复杂度分布:
- 低复杂度: {{low_complexity_count}} 个 ({{low_complexity_percentage}}%)
- 中等复杂度: {{avg_complexity_count}} 个 ({{avg_complexity_percentage}}%)
- 高复杂度: {{high_complexity_count}} 个 ({{high_complexity_percentage}}%)

总功能点数: {{total_function_points}} FP
数据功能点: {{data_function_points}} FP ({{data_percentage}}%)
事务功能点: {{transaction_function_points}} FP ({{transaction_percentage}}%)

项目估算
========
预估指标:
- 总工作量: {{total_effort}} 人天
- 开发周期: {{development_duration}} 天
- 测试周期: {{testing_duration}} 天
- 总项目周期: {{total_duration}} 天
- 项目成本: {{project_cost}} 元

人力资源需求:
- 项目经理: {{pm_count}} 人
- 系统分析师: {{analyst_count}} 人  
- 开发工程师: {{developer_count}} 人
- 测试工程师: {{tester_count}} 人
- UI/UX设计师: {{designer_count}} 人

风险评估
========
主要风险:
1. {{risk_1}}
2. {{risk_2}}
3. {{risk_3}}

缓解措施:
1. {{mitigation_1}}
2. {{mitigation_2}}
3. {{mitigation_3}}

里程碑计划
==========
里程碑 | 计划完成日期 | 主要交付物 | 责任人 | 状态
-------|--------------|------------|--------|------
需求分析完成 | {{milestone_1_date}} | 需求规格说明书 | {{milestone_1_owner}} | {{milestone_1_status}}
系统设计完成 | {{milestone_2_date}} | 系统设计文档 | {{milestone_2_owner}} | {{milestone_2_status}}
开发完成 | {{milestone_3_date}} | 系统代码 | {{milestone_3_owner}} | {{milestone_3_status}}
测试完成 | {{milestone_4_date}} | 测试报告 | {{milestone_4_owner}} | {{milestone_4_status}}
上线部署 | {{milestone_5_date}} | 生产系统 | {{milestone_5_owner}} | {{milestone_5_status}}

更新日志
========
版本 | 更新日期 | 更新内容 | 更新人
-----|----------|----------|--------
v1.0 | {{update_date_1}} | 初始版本创建 | {{updater_1}}
v1.1 | {{update_date_2}} | 添加新需求REQ006-REQ010 | {{updater_2}}
v1.2 | {{update_date_3}} | 修订功能点评估 | {{updater_3}}`,
			IsDefault:  true,
			IsActive:   true,
			UsageCount: 0,
			CreatedBy:  1,
		},
	}

	for _, template := range templates {
		var existingCount int64
		global.GVA_DB.Model(&nesma.NesmaDocTemplate{}).
			Where("name = ? AND type = ? AND format = ?", template.Name, template.Type, template.Format).
			Count(&existingCount)

		if existingCount == 0 {
			if err := global.GVA_DB.Create(&template).Error; err != nil {
				global.GVA_LOG.Error("Failed to create sample template",
					zap.String("name", template.Name),
					zap.Error(err))
			} else {
				global.GVA_LOG.Info("Created sample template",
					zap.String("name", template.Name))
			}
		}
	}

	global.GVA_LOG.Info("NESMA sample templates initialization completed")
}

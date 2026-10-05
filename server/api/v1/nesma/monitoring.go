package nesma

import (
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
)

type MonitoringApi struct{}

// @Tags 监控管理
// @Summary 获取服务性能统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param serviceName query string false "服务名称"
// @Param modelName query string false "模型名称"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=nesma.ServicePerformanceStats} "获取成功"
// @Router /monitoring/service/stats [get]
func (m *MonitoringApi) GetServiceStats(ctx *gin.Context) {
	query := nesma.MonitoringQuery{
		ServiceName: getStringPtr(ctx.Query("serviceName")),
		ModelName:   getStringPtr(ctx.Query("modelName")),
	}

	// 解析时间参数
	if startTimeStr := ctx.Query("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := ctx.Query("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			query.EndTime = &endTime
		}
	}

	monitoringService := nesma.GetMonitoringService()
	stats, err := monitoringService.GetServiceStats(ctx, query)
	if err != nil {
		global.GVA_LOG.Error("获取服务统计失败: " + err.Error())
		response.FailWithMessage("获取服务统计失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(stats, ctx)
}

// @Tags 监控管理
// @Summary 获取错误分析
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param serviceName query string false "服务名称"
// @Param modelName query string false "模型名称"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=[]nesma.ErrorAnalysis} "获取成功"
// @Router /monitoring/error/analysis [get]
func (m *MonitoringApi) GetErrorAnalysis(ctx *gin.Context) {
	query := nesma.MonitoringQuery{
		ServiceName: getStringPtr(ctx.Query("serviceName")),
		ModelName:   getStringPtr(ctx.Query("modelName")),
	}

	// 解析时间参数
	if startTimeStr := ctx.Query("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := ctx.Query("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			query.EndTime = &endTime
		}
	}

	monitoringService := nesma.GetMonitoringService()
	analysis, err := monitoringService.GetErrorAnalysis(ctx, query)
	if err != nil {
		global.GVA_LOG.Error("获取错误分析失败: " + err.Error())
		response.FailWithMessage("获取错误分析失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(analysis, ctx)
}

// @Tags 监控管理
// @Summary 获取Token使用统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param serviceName query string false "服务名称"
// @Param modelName query string false "模型名称"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=[]nesma.TokenUsageStats} "获取成功"
// @Router /monitoring/token/stats [get]
func (m *MonitoringApi) GetTokenUsageStats(ctx *gin.Context) {
	query := nesma.MonitoringQuery{
		ServiceName: getStringPtr(ctx.Query("serviceName")),
		ModelName:   getStringPtr(ctx.Query("modelName")),
	}

	// 解析时间参数
	if startTimeStr := ctx.Query("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := ctx.Query("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			query.EndTime = &endTime
		}
	}

	monitoringService := nesma.GetMonitoringService()
	stats, err := monitoringService.GetTokenUsageStats(ctx, query)
	if err != nil {
		global.GVA_LOG.Error("获取Token使用统计失败: " + err.Error())
		response.FailWithMessage("获取Token使用统计失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(stats, ctx)
}

// @Tags 监控管理
// @Summary 获取时间序列统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param serviceName query string false "服务名称"
// @Param interval query string false "时间间隔(hour/day/week)"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /monitoring/time-series [get]
func (m *MonitoringApi) GetTimeSeriesStats(ctx *gin.Context) {
	query := nesma.MonitoringQuery{
		ServiceName: getStringPtr(ctx.Query("serviceName")),
	}

	interval := ctx.DefaultQuery("interval", "hour")

	// 解析时间参数
	if startTimeStr := ctx.Query("startTime"); startTimeStr != "" {
		if startTime, err := time.Parse("2006-01-02 15:04:05", startTimeStr); err == nil {
			query.StartTime = &startTime
		}
	}

	if endTimeStr := ctx.Query("endTime"); endTimeStr != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", endTimeStr); err == nil {
			query.EndTime = &endTime
		}
	}

	monitoringService := nesma.GetMonitoringService()
	stats, err := monitoringService.GetTimeSeriesStats(ctx, query, interval)
	if err != nil {
		global.GVA_LOG.Error("获取时间序列统计失败: " + err.Error())
		response.FailWithMessage("获取时间序列统计失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(stats, ctx)
}

// @Tags 监控管理
// @Summary 获取监控指标
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param serviceName query string false "服务名称"
// @Param modelName query string false "模型名称"
// @Param status query string false "状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /monitoring/metrics [get]
func (m *MonitoringApi) GetMonitoringMetrics(ctx *gin.Context) {
	query := nesma.MonitoringQuery{
		ServiceName: getStringPtr(ctx.Query("serviceName")),
		ModelName:   getStringPtr(ctx.Query("modelName")),
		Status:      getStringPtr(ctx.Query("status")),
	}

	// 解析分页参数
	if pageStr := ctx.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			query.Page = page
		}
	} else {
		query.Page = 1
	}

	if pageSizeStr := ctx.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			query.PageSize = pageSize
		}
	} else {
		query.PageSize = 20
	}

	monitoringService := nesma.GetMonitoringService()
	metrics, total, err := monitoringService.GetMetrics(ctx, query)
	if err != nil {
		global.GVA_LOG.Error("获取监控指标失败: " + err.Error())
		response.FailWithMessage("获取监控指标失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(map[string]interface{}{
		"list":     metrics,
		"total":    total,
		"page":     query.Page,
		"pageSize": query.PageSize,
	}, ctx)
}

// @Tags 监控管理
// @Summary 获取服务健康状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /monitoring/health [get]
func (m *MonitoringApi) GetServiceHealth(ctx *gin.Context) {
	monitoringService := nesma.GetMonitoringService()
	health, err := monitoringService.GetServiceHealth(ctx)
	if err != nil {
		global.GVA_LOG.Error("获取服务健康状态失败: " + err.Error())
		response.FailWithMessage("获取服务健康状态失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(health, ctx)
}

// @Tags 监控管理
// @Summary 获取需求统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID，不传则获取全部项目统计"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /monitor/requirement-stats [get]
func (m *MonitoringApi) GetRequirementStats(ctx *gin.Context) {
	projectIDStr := ctx.Query("projectId")
	
	var projectID *uint
	if projectIDStr != "" {
		if projectIDUint, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			temp := uint(projectIDUint)
			projectID = &temp
		} else {
			response.FailWithMessage("项目ID格式错误", ctx)
			return
		}
	}

	// 调用需求服务获取统计
	stats, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetNesmaRequirementStats(projectID)
	if err != nil {
		global.GVA_LOG.Error("获取需求统计失败!", zap.Error(err))
		response.FailWithMessage("获取需求统计失败: "+err.Error(), ctx)
		return
	}

	// 转换为监控面板需要的格式
	monitorStats := map[string]interface{}{
		"totalCount":      stats.TotalCount,
		"levelStats":      stats.LevelStats,
		"statusStats":     stats.StatusStats,
		"priorityStats":   stats.PriorityStats,
		"complexityStats": stats.ComplexityStats,
	}

	response.OkWithData(monitorStats, ctx)
}

// @Tags 监控管理
// @Summary 获取系统资源使用情况
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /monitoring/system/resources [get]
func (m *MonitoringApi) GetSystemResources(ctx *gin.Context) {
	// 模拟系统资源数据
	resources := map[string]interface{}{
		"cpuUsage":    65.5,
		"memoryUsage": 72.3,
		"diskUsage":   45.8,
		"networkIn":   1024.5,
		"networkOut":  768.2,
		"timestamp":   time.Now(),
	}

	response.OkWithData(resources, ctx)
}

// @Tags 监控管理
// @Summary 获取模型性能对比
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param models query string false "模型列表(逗号分隔)"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /monitoring/model/comparison [get]
func (m *MonitoringApi) GetModelPerformanceComparison(ctx *gin.Context) {
	// 模拟模型性能对比数据
	comparison := []map[string]interface{}{
		{
			"modelName":       "deepseek",
			"avgResponseTime": 1250,
			"successRate":     98.5,
			"errorRate":       1.5,
			"tokenUsage":      125000,
			"cost":            15.75,
		},
		{
			"modelName":       "gpt-4",
			"avgResponseTime": 1850,
			"successRate":     99.2,
			"errorRate":       0.8,
			"tokenUsage":      95000,
			"cost":            28.50,
		},
		{
			"modelName":       "claude",
			"avgResponseTime": 1680,
			"successRate":     98.8,
			"errorRate":       1.2,
			"tokenUsage":      108000,
			"cost":            21.60,
		},
	}

	response.OkWithData(comparison, ctx)
}

// @Tags 监控管理
// @Summary 获取响应时间分布
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param serviceName query string false "服务名称"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /monitoring/response-time/distribution [get]
func (m *MonitoringApi) GetResponseTimeDistribution(ctx *gin.Context) {
	// 模拟响应时间分布数据
	distribution := map[string]interface{}{
		"ranges":      []string{"0-100ms", "100-500ms", "500-1s", "1-2s", "2-5s", ">5s"},
		"counts":      []int{120, 350, 280, 150, 80, 20},
		"percentages": []float64{12.0, 35.0, 28.0, 15.0, 8.0, 2.0},
	}

	response.OkWithData(distribution, ctx)
}

// @Tags 监控管理
// @Summary 获取成本分析
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param serviceName query string false "服务名称"
// @Param modelName query string false "模型名称"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /monitoring/cost/analysis [get]
func (m *MonitoringApi) GetCostAnalysis(ctx *gin.Context) {
	// 模拟成本分析数据
	analysis := map[string]interface{}{
		"totalCost": 125.75,
		"costByModel": map[string]float64{
			"deepseek": 45.25,
			"gpt-4":    52.30,
			"claude":   28.20,
		},
		"costByService": map[string]float64{
			"chat":       65.50,
			"analysis":   35.25,
			"generation": 25.00,
		},
		"dailyCosts": []map[string]interface{}{
			{"date": "2024-01-01", "cost": 15.25},
			{"date": "2024-01-02", "cost": 18.50},
			{"date": "2024-01-03", "cost": 22.30},
			{"date": "2024-01-04", "cost": 19.80},
			{"date": "2024-01-05", "cost": 24.90},
		},
		"projectedMonthlyCost": 1890.50,
	}

	response.OkWithData(analysis, ctx)
}

// @Tags 监控管理
// @Summary 导出监控报告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/octet-stream
// @Param format query string false "导出格式(pdf/excel)"
// @Param startTime query string false "开始时间"
// @Param endTime query string false "结束时间"
// @Success 200 {file} file "导出成功"
// @Router /monitoring/report/export [get]
func (m *MonitoringApi) ExportMonitoringReport(ctx *gin.Context) {
	format := ctx.DefaultQuery("format", "pdf")

	// 这里应该调用报告生成服务
	// 目前返回模拟数据

	var contentType string
	var filename string
	var data []byte

	switch format {
	case "excel":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "monitoring_report.xlsx"
		data = []byte("模拟Excel报告数据")
	default:
		contentType = "application/pdf"
		filename = "monitoring_report.pdf"
		data = []byte("模拟PDF报告数据")
	}

	ctx.Header("Content-Type", contentType)
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Data(200, contentType, data)
}

// 辅助函数
func getStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

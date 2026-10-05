package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MonitoringService 监控服务
type MonitoringService struct {
	db *gorm.DB
}

// AIServiceMetrics AI服务指标
type AIServiceMetrics struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	UserID           uint      `json:"userId" gorm:"not null;index"`
	ServiceName      string    `json:"serviceName" gorm:"type:varchar(50);not null;index"`
	ModelName        string    `json:"modelName" gorm:"type:varchar(100);not null"`
	RequestType      string    `json:"requestType" gorm:"type:varchar(50);not null"` // chat, completion, etc.
	Status           string    `json:"status" gorm:"type:varchar(20);not null"`      // success, error, timeout
	ResponseTime     int64     `json:"responseTime" gorm:"not null"`                 // 响应时间(毫秒)
	TokenUsage       int       `json:"tokenUsage" gorm:"default:0"`                  // Token使用量
	PromptTokens     int       `json:"promptTokens" gorm:"default:0"`                // 输入Token数
	CompletionTokens int       `json:"completionTokens" gorm:"default:0"`            // 输出Token数
	ErrorMessage     string    `json:"errorMessage" gorm:"type:text"`                // 错误信息
	RequestPayload   string    `json:"requestPayload" gorm:"type:json"`              // 请求载荷
	CreatedAt        time.Time `json:"createdAt" gorm:"index"`
}

// ServicePerformanceStats 服务性能统计
type ServicePerformanceStats struct {
	ServiceName         string    `json:"serviceName"`
	TotalRequests       int64     `json:"totalRequests"`
	SuccessRequests     int64     `json:"successRequests"`
	ErrorRequests       int64     `json:"errorRequests"`
	SuccessRate         float64   `json:"successRate"`
	ErrorRate           float64   `json:"errorRate"`
	AvgResponseTime     float64   `json:"avgResponseTime"`
	MinResponseTime     int64     `json:"minResponseTime"`
	MaxResponseTime     int64     `json:"maxResponseTime"`
	P95ResponseTime     int64     `json:"p95ResponseTime"`
	P99ResponseTime     int64     `json:"p99ResponseTime"`
	TotalTokens         int64     `json:"totalTokens"`
	AvgTokensPerRequest float64   `json:"avgTokensPerRequest"`
	StartTime           time.Time `json:"startTime"`
	EndTime             time.Time `json:"endTime"`
}

// ErrorAnalysis 错误分析
type ErrorAnalysis struct {
	ErrorType    string    `json:"errorType"`
	ErrorCount   int64     `json:"errorCount"`
	ErrorRate    float64   `json:"errorRate"`
	LastOccurred time.Time `json:"lastOccurred"`
	Examples     []string  `json:"examples"`
}

// TokenUsageStats Token使用统计
type TokenUsageStats struct {
	ServiceName         string  `json:"serviceName"`
	ModelName           string  `json:"modelName"`
	TotalTokens         int64   `json:"totalTokens"`
	PromptTokens        int64   `json:"promptTokens"`
	CompletionTokens    int64   `json:"completionTokens"`
	RequestCount        int64   `json:"requestCount"`
	AvgTokensPerRequest float64 `json:"avgTokensPerRequest"`
	CostEstimate        float64 `json:"costEstimate"`
}

// MonitoringQuery 监控查询参数
type MonitoringQuery struct {
	UserID      *uint      `json:"userId"`
	ServiceName *string    `json:"serviceName"`
	ModelName   *string    `json:"modelName"`
	Status      *string    `json:"status"`
	RequestType *string    `json:"requestType"`
	StartTime   *time.Time `json:"startTime"`
	EndTime     *time.Time `json:"endTime"`
	Page        int        `json:"page"`
	PageSize    int        `json:"pageSize"`
}

// NewMonitoringService 创建监控服务
func NewMonitoringService() *MonitoringService {
	return &MonitoringService{
		db: global.GVA_DB,
	}
}

// RecordMetrics 记录指标
func (ms *MonitoringService) RecordMetrics(ctx context.Context, userID uint, serviceName, modelName, requestType string, startTime time.Time, status string, tokenUsage *TokenUsage, errorMsg string, requestPayload interface{}) error {
	responseTime := time.Since(startTime).Milliseconds()

	// 序列化请求载荷
	var payloadStr string
	if requestPayload != nil {
		if data, err := json.Marshal(requestPayload); err == nil {
			payloadStr = string(data)
		}
	}

	metrics := &AIServiceMetrics{
		UserID:         userID,
		ServiceName:    serviceName,
		ModelName:      modelName,
		RequestType:    requestType,
		Status:         status,
		ResponseTime:   responseTime,
		ErrorMessage:   errorMsg,
		RequestPayload: payloadStr,
		CreatedAt:      time.Now(),
	}

	if tokenUsage != nil {
		metrics.TokenUsage = tokenUsage.TotalTokens
		metrics.PromptTokens = tokenUsage.PromptTokens
		metrics.CompletionTokens = tokenUsage.CompletionTokens
	}

	err := ms.db.Create(metrics).Error
	if err != nil {
		global.GVA_LOG.Error("记录监控指标失败", zap.Error(err))
		return err
	}

	return nil
}

// GetServiceStats 获取服务统计
func (ms *MonitoringService) GetServiceStats(ctx context.Context, query MonitoringQuery) (*ServicePerformanceStats, error) {
	dbQuery := ms.db.Model(&AIServiceMetrics{})

	// 应用过滤条件
	dbQuery = ms.applyFilters(dbQuery, query)

	var stats struct {
		TotalRequests   int64     `json:"total_requests"`
		SuccessRequests int64     `json:"success_requests"`
		ErrorRequests   int64     `json:"error_requests"`
		AvgResponseTime float64   `json:"avg_response_time"`
		MinResponseTime int64     `json:"min_response_time"`
		MaxResponseTime int64     `json:"max_response_time"`
		TotalTokens     int64     `json:"total_tokens"`
		MinCreatedAt    time.Time `json:"min_created_at"`
		MaxCreatedAt    time.Time `json:"max_created_at"`
	}

	err := dbQuery.Select(`
		COUNT(*) as total_requests,
		COUNT(CASE WHEN status = 'success' THEN 1 END) as success_requests,
		COUNT(CASE WHEN status != 'success' THEN 1 END) as error_requests,
		AVG(response_time) as avg_response_time,
		MIN(response_time) as min_response_time,
		MAX(response_time) as max_response_time,
		SUM(token_usage) as total_tokens,
		MIN(created_at) as min_created_at,
		MAX(created_at) as max_created_at
	`).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	// 计算百分位响应时间
	p95, p99, err := ms.calculatePercentileResponseTimes(query)
	if err != nil {
		global.GVA_LOG.Warn("计算百分位响应时间失败", zap.Error(err))
	}

	// 计算成功率和错误率
	successRate := float64(0)
	errorRate := float64(0)
	if stats.TotalRequests > 0 {
		successRate = float64(stats.SuccessRequests) / float64(stats.TotalRequests) * 100
		errorRate = float64(stats.ErrorRequests) / float64(stats.TotalRequests) * 100
	}

	avgTokensPerRequest := float64(0)
	if stats.TotalRequests > 0 {
		avgTokensPerRequest = float64(stats.TotalTokens) / float64(stats.TotalRequests)
	}

	serviceName := "All Services"
	if query.ServiceName != nil {
		serviceName = *query.ServiceName
	}

	return &ServicePerformanceStats{
		ServiceName:         serviceName,
		TotalRequests:       stats.TotalRequests,
		SuccessRequests:     stats.SuccessRequests,
		ErrorRequests:       stats.ErrorRequests,
		SuccessRate:         successRate,
		ErrorRate:           errorRate,
		AvgResponseTime:     stats.AvgResponseTime,
		MinResponseTime:     stats.MinResponseTime,
		MaxResponseTime:     stats.MaxResponseTime,
		P95ResponseTime:     p95,
		P99ResponseTime:     p99,
		TotalTokens:         stats.TotalTokens,
		AvgTokensPerRequest: avgTokensPerRequest,
		StartTime:           stats.MinCreatedAt,
		EndTime:             stats.MaxCreatedAt,
	}, nil
}

// GetErrorAnalysis 获取错误分析
func (ms *MonitoringService) GetErrorAnalysis(ctx context.Context, query MonitoringQuery) ([]ErrorAnalysis, error) {
	dbQuery := ms.db.Model(&AIServiceMetrics{}).Where("status != ?", "success")

	// 应用过滤条件
	dbQuery = ms.applyFilters(dbQuery, query)

	var errorStats []struct {
		ErrorType    string    `json:"error_type"`
		ErrorCount   int64     `json:"error_count"`
		LastOccurred time.Time `json:"last_occurred"`
	}

	// 简单的错误分类：基于错误消息的前50个字符
	err := dbQuery.Select(`
		SUBSTRING(error_message, 1, 50) as error_type,
		COUNT(*) as error_count,
		MAX(created_at) as last_occurred
	`).Group("error_type").Order("error_count DESC").Limit(20).Find(&errorStats).Error

	if err != nil {
		return nil, err
	}

	// 获取总错误数用于计算错误率
	var totalErrors int64
	dbQuery.Count(&totalErrors)

	var results []ErrorAnalysis
	for _, stat := range errorStats {
		errorRate := float64(0)
		if totalErrors > 0 {
			errorRate = float64(stat.ErrorCount) / float64(totalErrors) * 100
		}

		// 获取错误示例
		var examples []string
		var errorMessages []string
		ms.db.Model(&AIServiceMetrics{}).
			Where("SUBSTRING(error_message, 1, 50) = ?", stat.ErrorType).
			Where("status != ?", "success").
			Where("error_message != ''").
			Order("created_at DESC").
			Limit(3).
			Pluck("error_message", &errorMessages)

		for _, msg := range errorMessages {
			if len(msg) > 100 {
				msg = msg[:100] + "..."
			}
			examples = append(examples, msg)
		}

		results = append(results, ErrorAnalysis{
			ErrorType:    stat.ErrorType,
			ErrorCount:   stat.ErrorCount,
			ErrorRate:    errorRate,
			LastOccurred: stat.LastOccurred,
			Examples:     examples,
		})
	}

	return results, nil
}

// GetTokenUsageStats 获取Token使用统计
func (ms *MonitoringService) GetTokenUsageStats(ctx context.Context, query MonitoringQuery) ([]TokenUsageStats, error) {
	dbQuery := ms.db.Model(&AIServiceMetrics{})

	// 应用过滤条件
	dbQuery = ms.applyFilters(dbQuery, query)

	var tokenStats []struct {
		ServiceName      string `json:"service_name"`
		ModelName        string `json:"model_name"`
		TotalTokens      int64  `json:"total_tokens"`
		PromptTokens     int64  `json:"prompt_tokens"`
		CompletionTokens int64  `json:"completion_tokens"`
		RequestCount     int64  `json:"request_count"`
	}

	err := dbQuery.Select(`
		service_name,
		model_name,
		SUM(token_usage) as total_tokens,
		SUM(prompt_tokens) as prompt_tokens,
		SUM(completion_tokens) as completion_tokens,
		COUNT(*) as request_count
	`).Group("service_name, model_name").Order("total_tokens DESC").Find(&tokenStats).Error

	if err != nil {
		return nil, err
	}

	var results []TokenUsageStats
	for _, stat := range tokenStats {
		avgTokensPerRequest := float64(0)
		if stat.RequestCount > 0 {
			avgTokensPerRequest = float64(stat.TotalTokens) / float64(stat.RequestCount)
		}

		// 简单的成本估算（这里使用假设的价格）
		costEstimate := ms.calculateCostEstimate(stat.ServiceName, stat.ModelName, stat.TotalTokens)

		results = append(results, TokenUsageStats{
			ServiceName:         stat.ServiceName,
			ModelName:           stat.ModelName,
			TotalTokens:         stat.TotalTokens,
			PromptTokens:        stat.PromptTokens,
			CompletionTokens:    stat.CompletionTokens,
			RequestCount:        stat.RequestCount,
			AvgTokensPerRequest: avgTokensPerRequest,
			CostEstimate:        costEstimate,
		})
	}

	return results, nil
}

// GetTimeSeriesStats 获取时序统计数据
func (ms *MonitoringService) GetTimeSeriesStats(ctx context.Context, query MonitoringQuery, interval string) ([]map[string]interface{}, error) {
	dbQuery := ms.db.Model(&AIServiceMetrics{})

	// 应用过滤条件
	dbQuery = ms.applyFilters(dbQuery, query)

	var timeFormat string
	switch interval {
	case "hour":
		timeFormat = "DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')"
	case "day":
		timeFormat = "DATE_FORMAT(created_at, '%Y-%m-%d')"
	case "week":
		timeFormat = "DATE_FORMAT(created_at, '%Y-%u')"
	case "month":
		timeFormat = "DATE_FORMAT(created_at, '%Y-%m')"
	default:
		timeFormat = "DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')"
	}

	var results []map[string]interface{}

	rows, err := dbQuery.Select(fmt.Sprintf(`
		%s as time_period,
		COUNT(*) as total_requests,
		COUNT(CASE WHEN status = 'success' THEN 1 END) as success_requests,
		COUNT(CASE WHEN status != 'success' THEN 1 END) as error_requests,
		AVG(response_time) as avg_response_time,
		SUM(token_usage) as total_tokens
	`, timeFormat)).Group("time_period").Order("time_period ASC").Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var timePeriod string
		var totalRequests, successRequests, errorRequests int64
		var avgResponseTime, totalTokens float64

		err := rows.Scan(&timePeriod, &totalRequests, &successRequests, &errorRequests, &avgResponseTime, &totalTokens)
		if err != nil {
			continue
		}

		successRate := float64(0)
		if totalRequests > 0 {
			successRate = float64(successRequests) / float64(totalRequests) * 100
		}

		results = append(results, map[string]interface{}{
			"time":            timePeriod,
			"totalRequests":   totalRequests,
			"successRequests": successRequests,
			"errorRequests":   errorRequests,
			"successRate":     successRate,
			"avgResponseTime": avgResponseTime,
			"totalTokens":     int64(totalTokens),
		})
	}

	return results, nil
}

// GetMetrics 获取指标列表
func (ms *MonitoringService) GetMetrics(ctx context.Context, query MonitoringQuery) ([]AIServiceMetrics, int64, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 100
	}

	dbQuery := ms.db.Model(&AIServiceMetrics{})

	// 应用过滤条件
	dbQuery = ms.applyFilters(dbQuery, query)

	var total int64
	err := dbQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var metrics []AIServiceMetrics
	err = dbQuery.Order("created_at DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&metrics).Error

	if err != nil {
		return nil, 0, err
	}

	return metrics, total, nil
}

// applyFilters 应用查询过滤条件
func (ms *MonitoringService) applyFilters(query *gorm.DB, filter MonitoringQuery) *gorm.DB {
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.ServiceName != nil {
		query = query.Where("service_name = ?", *filter.ServiceName)
	}
	if filter.ModelName != nil {
		query = query.Where("model_name = ?", *filter.ModelName)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.RequestType != nil {
		query = query.Where("request_type = ?", *filter.RequestType)
	}
	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	return query
}

// calculatePercentileResponseTimes 计算百分位响应时间
func (ms *MonitoringService) calculatePercentileResponseTimes(query MonitoringQuery) (int64, int64, error) {
	dbQuery := ms.db.Model(&AIServiceMetrics{})
	dbQuery = ms.applyFilters(dbQuery, query)

	var responseTimes []int64
	err := dbQuery.Order("response_time ASC").Pluck("response_time", &responseTimes).Error
	if err != nil {
		return 0, 0, err
	}

	if len(responseTimes) == 0 {
		return 0, 0, nil
	}

	// 计算P95和P99
	p95Index := int(float64(len(responseTimes)) * 0.95)
	p99Index := int(float64(len(responseTimes)) * 0.99)

	if p95Index >= len(responseTimes) {
		p95Index = len(responseTimes) - 1
	}
	if p99Index >= len(responseTimes) {
		p99Index = len(responseTimes) - 1
	}

	return responseTimes[p95Index], responseTimes[p99Index], nil
}

// calculateCostEstimate 计算成本估算
func (ms *MonitoringService) calculateCostEstimate(serviceName, modelName string, totalTokens int64) float64 {
	// 这里使用简化的定价模型，实际应用中应该从配置文件或数据库读取
	pricePerToken := 0.0

	switch serviceName {
	case "openai":
		switch modelName {
		case "gpt-3.5-turbo":
			pricePerToken = 0.002 / 1000 // $0.002 per 1K tokens
		case "gpt-4":
			pricePerToken = 0.03 / 1000 // $0.03 per 1K tokens
		case "gpt-4-turbo":
			pricePerToken = 0.01 / 1000 // $0.01 per 1K tokens
		}
	case "claude":
		pricePerToken = 0.01 / 1000 // 假设价格
	case "deepseek":
		pricePerToken = 0.001 / 1000 // 假设价格
	}

	return float64(totalTokens) * pricePerToken
}

// CleanupOldMetrics 清理旧的监控指标
func (ms *MonitoringService) CleanupOldMetrics(ctx context.Context, retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 90 // 默认保留90天
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	result := ms.db.Where("created_at < ?", cutoffTime).Delete(&AIServiceMetrics{})
	if result.Error != nil {
		return result.Error
	}

	global.GVA_LOG.Info("清理旧监控指标", zap.Int64("deleted_count", result.RowsAffected), zap.Time("cutoff_time", cutoffTime))
	return nil
}

// GetServiceHealth 获取服务健康状况
func (ms *MonitoringService) GetServiceHealth(ctx context.Context) (map[string]interface{}, error) {
	// 获取最近1小时的数据
	oneHourAgo := time.Now().Add(-1 * time.Hour)

	var services []struct {
		ServiceName     string    `json:"service_name"`
		TotalRequests   int64     `json:"total_requests"`
		SuccessRequests int64     `json:"success_requests"`
		AvgResponseTime float64   `json:"avg_response_time"`
		LastRequest     time.Time `json:"last_request"`
	}

	err := ms.db.Model(&AIServiceMetrics{}).
		Where("created_at >= ?", oneHourAgo).
		Select(`
			service_name,
			COUNT(*) as total_requests,
			COUNT(CASE WHEN status = 'success' THEN 1 END) as success_requests,
			AVG(response_time) as avg_response_time,
			MAX(created_at) as last_request
		`).
		Group("service_name").
		Find(&services).Error

	if err != nil {
		return nil, err
	}

	healthMap := make(map[string]interface{})

	for _, service := range services {
		successRate := float64(0)
		if service.TotalRequests > 0 {
			successRate = float64(service.SuccessRequests) / float64(service.TotalRequests) * 100
		}

		// 判断健康状态
		status := "healthy"
		if successRate < 95 {
			status = "warning"
		}
		if successRate < 80 {
			status = "unhealthy"
		}

		// 检查是否有最近的请求
		timeSinceLastRequest := time.Since(service.LastRequest).Minutes()
		if timeSinceLastRequest > 30 {
			status = "inactive"
		}

		healthMap[service.ServiceName] = map[string]interface{}{
			"status":               status,
			"totalRequests":        service.TotalRequests,
			"successRate":          successRate,
			"avgResponseTime":      service.AvgResponseTime,
			"lastRequest":          service.LastRequest,
			"timeSinceLastRequest": timeSinceLastRequest,
		}
	}

	return healthMap, nil
}

// 全局监控服务实例
var GlobalMonitoringService *MonitoringService

// InitMonitoringService 初始化监控服务
func InitMonitoringService() {
	GlobalMonitoringService = NewMonitoringService()

	// 自动迁移数据库表
	err := global.GVA_DB.AutoMigrate(&AIServiceMetrics{})
	if err != nil {
		global.GVA_LOG.Error("Failed to migrate monitoring tables", zap.Error(err))
	} else {
		global.GVA_LOG.Info("Monitoring service initialized successfully")
	}

	// 启动后台清理任务
	go GlobalMonitoringService.startCleanupTask()
}

// startCleanupTask 启动清理任务
func (ms *MonitoringService) startCleanupTask() {
	// 每天凌晨2点执行清理
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			if now.Hour() == 2 { // 凌晨2点
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				err := ms.CleanupOldMetrics(ctx, 90) // 保留90天
				cancel()

				if err != nil {
					global.GVA_LOG.Error("清理旧监控指标失败", zap.Error(err))
				}
			}
		}
	}
}

// GetMonitoringService 获取监控服务
func GetMonitoringService() *MonitoringService {
	if GlobalMonitoringService == nil {
		InitMonitoringService()
	}
	return GlobalMonitoringService
}

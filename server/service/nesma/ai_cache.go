package nesma

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AICache AI缓存系统
type AICache struct {
	l1Cache *LRUCache                  // 内存缓存（热点数据）
	l2Cache redis.UniversalClient      // Redis缓存（中等频率）
	l3Cache *gorm.DB                   // 数据库缓存（长期存储）
	stats   *CacheStats                // 缓存统计
	mu      sync.RWMutex
}

// CacheStats 缓存统计
type CacheStats struct {
	L1Hits     int64 `json:"l1_hits"`
	L1Misses   int64 `json:"l1_misses"`
	L2Hits     int64 `json:"l2_hits"`
	L2Misses   int64 `json:"l2_misses"`
	L3Hits     int64 `json:"l3_hits"`
	L3Misses   int64 `json:"l3_misses"`
	TotalHits  int64 `json:"total_hits"`
	TotalMisses int64 `json:"total_misses"`
}

// CacheEntry 缓存条目
type AICacheEntry struct {
	Key        string      `json:"key" gorm:"primaryKey"`
	Value      string      `json:"value"`
	Metadata   string      `json:"metadata"`
	CreatedAt  time.Time   `json:"created_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	AccessCount int        `json:"access_count"`
	LastAccess time.Time   `json:"last_access"`
}

// LRUCache 内存LRU缓存
type LRUCache struct {
	capacity int
	items    map[string]*LRUItem
	head     *LRUItem
	tail     *LRUItem
	mu       sync.RWMutex
}

// LRUItem LRU缓存项
type LRUItem struct {
	key       string
	value     interface{}
	expiresAt time.Time
	prev      *LRUItem
	next      *LRUItem
}

// NewAICache 创建AI缓存系统
func NewAICache() *AICache {
	cache := &AICache{
		l1Cache: NewLRUCache(1000), // 内存缓存1000个条目
		l2Cache: global.GVA_REDIS,  // 使用全局Redis
		l3Cache: global.GVA_DB,     // 使用全局数据库
		stats:   &CacheStats{},
	}
	
	// 确保数据库表存在
	cache.l3Cache.AutoMigrate(&AICacheEntry{})
	
	return cache
}

// Get 获取缓存
func (c *AICache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	hashKey := c.hashKey(key)
	
	// L1缓存查找
	if value, ok := c.l1Cache.Get(hashKey); ok {
		c.stats.L1Hits++
		c.stats.TotalHits++
		return value
	}
	c.stats.L1Misses++
	
	// L2缓存查找
	if c.l2Cache != nil {
		ctx := context.Background()
		result, err := c.l2Cache.Get(ctx, hashKey).Result()
		if err == nil {
			c.stats.L2Hits++
			c.stats.TotalHits++
			
			// 反序列化
			var value interface{}
			if err := json.Unmarshal([]byte(result), &value); err == nil {
				// 回写到L1缓存
				c.l1Cache.Set(hashKey, value, time.Hour)
				return value
			}
		}
	}
	c.stats.L2Misses++
	
	// L3缓存查找
	var entry AICacheEntry
	if err := c.l3Cache.Where("key = ? AND expires_at > ?", hashKey, time.Now()).First(&entry).Error; err == nil {
		c.stats.L3Hits++
		c.stats.TotalHits++
		
		// 反序列化
		var value interface{}
		if err := json.Unmarshal([]byte(entry.Value), &value); err == nil {
			// 回写到上级缓存
			c.l1Cache.Set(hashKey, value, time.Hour)
			if c.l2Cache != nil {
				c.l2Cache.Set(context.Background(), hashKey, entry.Value, time.Hour)
			}
			
			// 更新访问记录
			c.l3Cache.Model(&entry).Updates(map[string]interface{}{
				"access_count": entry.AccessCount + 1,
				"last_access":  time.Now(),
			})
			
			return value
		}
	}
	c.stats.L3Misses++
	c.stats.TotalMisses++
	
	return nil
}

// Set 设置缓存
func (c *AICache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	hashKey := c.hashKey(key)
	
	// 序列化值
	jsonValue, err := json.Marshal(value)
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("Failed to marshal cache value", zap.Error(err))
		}
		return
	}
	
	// L1缓存
	c.l1Cache.Set(hashKey, value, ttl)
	
	// L2缓存
	if c.l2Cache != nil {
		ctx := context.Background()
		c.l2Cache.Set(ctx, hashKey, string(jsonValue), ttl)
	}
	
	// L3缓存
	entry := AICacheEntry{
		Key:        hashKey,
		Value:      string(jsonValue),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
		AccessCount: 0,
		LastAccess: time.Now(),
	}
	
	// 使用UPSERT
	c.l3Cache.Where("key = ?", hashKey).Assign(entry).FirstOrCreate(&entry)
}

// Delete 删除缓存
func (c *AICache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	hashKey := c.hashKey(key)
	
	// 从所有层删除
	c.l1Cache.Delete(hashKey)
	
	if c.l2Cache != nil {
		c.l2Cache.Del(context.Background(), hashKey)
	}
	
	c.l3Cache.Where("key = ?", hashKey).Delete(&AICacheEntry{})
}

// hashKey 生成哈希键
func (c *AICache) hashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

// GetStats 获取缓存统计
func (c *AICache) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats
}

// ClearExpired 清理过期缓存
func (c *AICache) ClearExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// 清理L1缓存
	c.l1Cache.ClearExpired()
	
	// 清理L3缓存
	c.l3Cache.Where("expires_at < ?", time.Now()).Delete(&AICacheEntry{})
	
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("AI cache expired entries cleared")
	}
}

// NewLRUCache 创建LRU缓存
func NewLRUCache(capacity int) *LRUCache {
	lru := &LRUCache{
		capacity: capacity,
		items:    make(map[string]*LRUItem),
	}
	
	// 创建头尾节点
	lru.head = &LRUItem{}
	lru.tail = &LRUItem{}
	lru.head.next = lru.tail
	lru.tail.prev = lru.head
	
	return lru
}

// Get 获取LRU缓存值
func (lru *LRUCache) Get(key string) (interface{}, bool) {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	
	if item, exists := lru.items[key]; exists {
		// 检查是否过期
		if time.Now().After(item.expiresAt) {
			lru.removeItem(item)
			delete(lru.items, key)
			return nil, false
		}
		
		// 移到头部
		lru.moveToHead(item)
		return item.value, true
	}
	
	return nil, false
}

// Set 设置LRU缓存值
func (lru *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	
	if item, exists := lru.items[key]; exists {
		// 更新已存在的项
		item.value = value
		item.expiresAt = time.Now().Add(ttl)
		lru.moveToHead(item)
	} else {
		// 创建新项
		item := &LRUItem{
			key:       key,
			value:     value,
			expiresAt: time.Now().Add(ttl),
		}
		
		lru.items[key] = item
		lru.addToHead(item)
		
		// 检查容量
		if len(lru.items) > lru.capacity {
			tail := lru.removeTail()
			delete(lru.items, tail.key)
		}
	}
}

// Delete 删除LRU缓存值
func (lru *LRUCache) Delete(key string) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	
	if item, exists := lru.items[key]; exists {
		lru.removeItem(item)
		delete(lru.items, key)
	}
}

// ClearExpired 清理过期项
func (lru *LRUCache) ClearExpired() {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	
	now := time.Now()
	var expiredKeys []string
	
	for key, item := range lru.items {
		if now.After(item.expiresAt) {
			lru.removeItem(item)
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	for _, key := range expiredKeys {
		delete(lru.items, key)
	}
}

// LRU缓存链表操作
func (lru *LRUCache) addToHead(item *LRUItem) {
	item.prev = lru.head
	item.next = lru.head.next
	lru.head.next.prev = item
	lru.head.next = item
}

func (lru *LRUCache) removeItem(item *LRUItem) {
	item.prev.next = item.next
	item.next.prev = item.prev
}

func (lru *LRUCache) moveToHead(item *LRUItem) {
	lru.removeItem(item)
	lru.addToHead(item)
}

func (lru *LRUCache) removeTail() *LRUItem {
	last := lru.tail.prev
	lru.removeItem(last)
	return last
}

// AIMonitor AI监控系统
type AIMonitor struct {
	metrics    map[string]*AIMetrics
	alerts     []*AIAlert
	mu         sync.RWMutex
	startTime  time.Time
}

// AIMetrics AI指标
type AIMetrics struct {
	ModelName       string        `json:"model_name"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	FailedRequests  int64         `json:"failed_requests"`
	AvgLatency      time.Duration `json:"avg_latency"`
	MaxLatency      time.Duration `json:"max_latency"`
	MinLatency      time.Duration `json:"min_latency"`
	TotalTokens     int64         `json:"total_tokens"`
	TotalCost       float64       `json:"total_cost"`
	ErrorRate       float64       `json:"error_rate"`
	LastError       string        `json:"last_error"`
	LastErrorTime   time.Time     `json:"last_error_time"`
	QPS             float64       `json:"qps"`
	Uptime          time.Duration `json:"uptime"`
}

// AIAlert AI告警
type AIAlert struct {
	ID          string    `json:"id"`
	ModelName   string    `json:"model_name"`
	Type        string    `json:"type"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  time.Time `json:"resolved_at"`
}

// NewAIMonitor 创建AI监控系统
func NewAIMonitor() *AIMonitor {
	return &AIMonitor{
		metrics:   make(map[string]*AIMetrics),
		alerts:    make([]*AIAlert, 0),
		startTime: time.Now(),
	}
}

// RecordRequest 记录请求
func (m *AIMonitor) RecordRequest(modelName string, latency time.Duration, tokens int, cost float64, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.metrics[modelName]; !exists {
		m.metrics[modelName] = &AIMetrics{
			ModelName:  modelName,
			MinLatency: latency,
			MaxLatency: latency,
		}
	}
	
	metrics := m.metrics[modelName]
	metrics.TotalRequests++
	metrics.TotalTokens += int64(tokens)
	metrics.TotalCost += cost
	
	if success {
		metrics.SuccessRequests++
	} else {
		metrics.FailedRequests++
	}
	
	// 更新延迟统计
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}
	if latency < metrics.MinLatency {
		metrics.MinLatency = latency
	}
	
	// 更新平均延迟
	if metrics.TotalRequests == 1 {
		metrics.AvgLatency = latency
	} else {
		metrics.AvgLatency = time.Duration(
			(int64(metrics.AvgLatency) + int64(latency)) / 2,
		)
	}
	
	// 更新错误率
	metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	
	// 更新QPS
	uptime := time.Since(m.startTime).Seconds()
	if uptime > 0 {
		metrics.QPS = float64(metrics.TotalRequests) / uptime
	}
	
	metrics.Uptime = time.Since(m.startTime)
	
	// 检查告警条件
	m.checkAlerts(modelName, metrics)
}

// RecordError 记录错误
func (m *AIMonitor) RecordError(modelName string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.metrics[modelName]; !exists {
		m.metrics[modelName] = &AIMetrics{
			ModelName: modelName,
		}
	}
	
	metrics := m.metrics[modelName]
	metrics.LastError = err.Error()
	metrics.LastErrorTime = time.Now()
	
	// 创建错误告警
	alert := &AIAlert{
		ID:        fmt.Sprintf("error_%s_%d", modelName, time.Now().Unix()),
		ModelName: modelName,
		Type:      "error",
		Level:     "warning",
		Message:   fmt.Sprintf("AI service error: %s", err.Error()),
		Timestamp: time.Now(),
		Resolved:  false,
	}
	
	m.alerts = append(m.alerts, alert)
	
	if global.GVA_LOG != nil {
		global.GVA_LOG.Warn("AI service error recorded",
			zap.String("model", modelName),
			zap.Error(err))
	}
}

// checkAlerts 检查告警条件
func (m *AIMonitor) checkAlerts(modelName string, metrics *AIMetrics) {
	// 错误率过高告警
	if metrics.ErrorRate > 0.1 && metrics.TotalRequests > 10 {
		alert := &AIAlert{
			ID:        fmt.Sprintf("high_error_rate_%s_%d", modelName, time.Now().Unix()),
			ModelName: modelName,
			Type:      "high_error_rate",
			Level:     "critical",
			Message:   fmt.Sprintf("High error rate detected: %.2f%%", metrics.ErrorRate*100),
			Timestamp: time.Now(),
			Resolved:  false,
		}
		m.alerts = append(m.alerts, alert)
	}
	
	// 延迟过高告警
	if metrics.AvgLatency > 30*time.Second {
		alert := &AIAlert{
			ID:        fmt.Sprintf("high_latency_%s_%d", modelName, time.Now().Unix()),
			ModelName: modelName,
			Type:      "high_latency",
			Level:     "warning",
			Message:   fmt.Sprintf("High latency detected: %v", metrics.AvgLatency),
			Timestamp: time.Now(),
			Resolved:  false,
		}
		m.alerts = append(m.alerts, alert)
	}
	
	// 成本过高告警
	if metrics.TotalCost > 100.0 && metrics.TotalRequests > 100 {
		avgCost := metrics.TotalCost / float64(metrics.TotalRequests)
		if avgCost > 1.0 {
			alert := &AIAlert{
				ID:        fmt.Sprintf("high_cost_%s_%d", modelName, time.Now().Unix()),
				ModelName: modelName,
				Type:      "high_cost",
				Level:     "warning",
				Message:   fmt.Sprintf("High cost per request: $%.4f", avgCost),
				Timestamp: time.Now(),
				Resolved:  false,
			}
			m.alerts = append(m.alerts, alert)
		}
	}
}

// GetMetrics 获取指标
func (m *AIMonitor) GetMetrics(modelName string) *AIMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics[modelName]
}

// GetAllMetrics 获取所有指标
func (m *AIMonitor) GetAllMetrics() map[string]*AIMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]*AIMetrics)
	for k, v := range m.metrics {
		result[k] = v
	}
	return result
}

// GetAlerts 获取告警
func (m *AIMonitor) GetAlerts() []*AIAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.alerts
}

// GetUnresolvedAlerts 获取未解决的告警
func (m *AIMonitor) GetUnresolvedAlerts() []*AIAlert {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var unresolved []*AIAlert
	for _, alert := range m.alerts {
		if !alert.Resolved {
			unresolved = append(unresolved, alert)
		}
	}
	return unresolved
}

// ResolveAlert 解决告警
func (m *AIMonitor) ResolveAlert(alertID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, alert := range m.alerts {
		if alert.ID == alertID {
			alert.Resolved = true
			alert.ResolvedAt = time.Now()
			break
		}
	}
}

// ClearOldAlerts 清理旧告警
func (m *AIMonitor) ClearOldAlerts(olderThan time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	cutoff := time.Now().Add(-olderThan)
	var newAlerts []*AIAlert
	
	for _, alert := range m.alerts {
		if alert.Timestamp.After(cutoff) {
			newAlerts = append(newAlerts, alert)
		}
	}
	
	m.alerts = newAlerts
}

// GetSystemHealth 获取系统健康状态
func (m *AIMonitor) GetSystemHealth() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	health := map[string]interface{}{
		"uptime":            time.Since(m.startTime),
		"total_models":      len(m.metrics),
		"total_requests":    int64(0),
		"total_errors":      int64(0),
		"avg_error_rate":    float64(0),
		"active_alerts":     len(m.GetUnresolvedAlerts()),
		"system_status":     "healthy",
	}
	
	var totalRequests, totalErrors int64
	for _, metrics := range m.metrics {
		totalRequests += metrics.TotalRequests
		totalErrors += metrics.FailedRequests
	}
	
	health["total_requests"] = totalRequests
	health["total_errors"] = totalErrors
	
	if totalRequests > 0 {
		health["avg_error_rate"] = float64(totalErrors) / float64(totalRequests)
	}
	
	// 判断系统状态
	if len(m.GetUnresolvedAlerts()) > 0 {
		health["system_status"] = "warning"
	}
	
	if health["avg_error_rate"].(float64) > 0.05 {
		health["system_status"] = "critical"
	}
	
	return health
}

// StartCacheCleanup 启动缓存清理定时任务
func (c *AICache) StartCacheCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			c.ClearExpired()
		}
	}()
}

// StartMonitoringCleanup 启动监控清理定时任务
func (m *AIMonitor) StartMonitoringCleanup() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			m.ClearOldAlerts(7 * 24 * time.Hour) // 清理7天前的告警
		}
	}()
}
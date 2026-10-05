package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/redis/go-redis/v9"
)

// RedisService Redis服务接口
type RedisService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	GetObject(ctx context.Context, key string, obj interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error
	GetSession(ctx context.Context, sessionID string, obj interface{}) error
	DeleteSession(ctx context.Context, sessionID string) error
	PublishMessage(ctx context.Context, channel string, message interface{}) error
	SubscribeChannel(ctx context.Context, channel string) *redis.PubSub
	SetAgentStatus(ctx context.Context, agentID string, status string) error
	GetAgentStatus(ctx context.Context, agentID string) (string, error)
	LockAgent(ctx context.Context, agentID string, timeout time.Duration) (bool, error)
	UnlockAgent(ctx context.Context, agentID string) error
	IncrementCounter(ctx context.Context, key string) (int64, error)
	SetWithLock(ctx context.Context, lockKey string, key string, value interface{}, expiration time.Duration) error
}

// RedisServiceImpl Redis服务实现
type RedisServiceImpl struct {
	client redis.UniversalClient
}

// NewRedisService 创建Redis服务实例
func NewRedisService() RedisService {
	return &RedisServiceImpl{
		client: global.GVA_REDIS,
	}
}

// Set 设置键值对
func (r *RedisServiceImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		data = string(jsonData)
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取字符串值
func (r *RedisServiceImpl) Get(ctx context.Context, key string) (string, error) {
	result := r.client.Get(ctx, key)
	if result.Err() != nil {
		return "", result.Err()
	}
	return result.Val(), nil
}

// GetObject 获取对象
func (r *RedisServiceImpl) GetObject(ctx context.Context, key string, obj interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), obj)
}

// Delete 删除键
func (r *RedisServiceImpl) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists 检查键是否存在
func (r *RedisServiceImpl) Exists(ctx context.Context, key string) (bool, error) {
	result := r.client.Exists(ctx, key)
	if result.Err() != nil {
		return false, result.Err()
	}
	return result.Val() > 0, nil
}

// SetSession 设置会话数据
func (r *RedisServiceImpl) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Set(ctx, key, data, expiration)
}

// GetSession 获取会话数据
func (r *RedisServiceImpl) GetSession(ctx context.Context, sessionID string, obj interface{}) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.GetObject(ctx, key, obj)
}

// DeleteSession 删除会话
func (r *RedisServiceImpl) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.Delete(ctx, key)
}

// PublishMessage 发布消息
func (r *RedisServiceImpl) PublishMessage(ctx context.Context, channel string, message interface{}) error {
	var data string
	switch v := message.(type) {
	case string:
		data = v
	default:
		jsonData, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		data = string(jsonData)
	}

	return r.client.Publish(ctx, channel, data).Err()
}

// SubscribeChannel 订阅频道
func (r *RedisServiceImpl) SubscribeChannel(ctx context.Context, channel string) *redis.PubSub {
	return r.client.Subscribe(ctx, channel)
}

// SetAgentStatus 设置Agent状态
func (r *RedisServiceImpl) SetAgentStatus(ctx context.Context, agentID string, status string) error {
	key := fmt.Sprintf("agent:status:%s", agentID)
	return r.Set(ctx, key, status, 5*time.Minute)
}

// GetAgentStatus 获取Agent状态
func (r *RedisServiceImpl) GetAgentStatus(ctx context.Context, agentID string) (string, error) {
	key := fmt.Sprintf("agent:status:%s", agentID)
	return r.Get(ctx, key)
}

// LockAgent Agent加锁
func (r *RedisServiceImpl) LockAgent(ctx context.Context, agentID string, timeout time.Duration) (bool, error) {
	key := fmt.Sprintf("agent:lock:%s", agentID)
	result := r.client.SetNX(ctx, key, "locked", timeout)
	if result.Err() != nil {
		return false, result.Err()
	}
	return result.Val(), nil
}

// UnlockAgent Agent解锁
func (r *RedisServiceImpl) UnlockAgent(ctx context.Context, agentID string) error {
	key := fmt.Sprintf("agent:lock:%s", agentID)
	return r.Delete(ctx, key)
}

// IncrementCounter 递增计数器
func (r *RedisServiceImpl) IncrementCounter(ctx context.Context, key string) (int64, error) {
	result := r.client.Incr(ctx, key)
	if result.Err() != nil {
		return 0, result.Err()
	}
	return result.Val(), nil
}

// SetWithLock 带锁的设置操作
func (r *RedisServiceImpl) SetWithLock(ctx context.Context, lockKey string, key string, value interface{}, expiration time.Duration) error {
	// 获取锁
	locked, err := r.LockAgent(ctx, lockKey, 10*time.Second)
	if err != nil {
		return err
	}
	if !locked {
		return errors.New("failed to acquire lock")
	}

	// 确保解锁
	defer r.UnlockAgent(ctx, lockKey)

	// 设置值
	return r.Set(ctx, key, value, expiration)
}

// 使用agent_communication.go中定义的AgentCommunicationMessage

// AgentCommunicationService Agent通信服务
type AgentCommunicationService struct {
	redis RedisService
}

// NewAgentCommunicationService 创建Agent通信服务
func NewAgentCommunicationService(redis RedisService) *AgentCommunicationService {
	return &AgentCommunicationService{
		redis: redis,
	}
}

// SendMessage 发送消息
func (a *AgentCommunicationService) SendMessage(ctx context.Context, agentID string, message interface{}) error {
	// 发布到目标Agent的频道
	channel := fmt.Sprintf("agent:message:%s", agentID)
	return a.redis.PublishMessage(ctx, channel, message)
}

// SubscribeMessages 订阅消息
func (a *AgentCommunicationService) SubscribeMessages(ctx context.Context, agentID string) *redis.PubSub {
	channel := fmt.Sprintf("agent:message:%s", agentID)
	return a.redis.SubscribeChannel(ctx, channel)
}

// BroadcastMessage 广播消息
func (a *AgentCommunicationService) BroadcastMessage(ctx context.Context, message interface{}) error {
	// 发布到广播频道
	channel := "agent:broadcast"
	return a.redis.PublishMessage(ctx, channel, message)
}

// SessionManager 会话管理器
type SessionManager struct {
	redis RedisService
}

// NewSessionManager 创建会话管理器
func NewSessionManager(redis RedisService) *SessionManager {
	return &SessionManager{
		redis: redis,
	}
}

// CreateSession 创建会话
func (s *SessionManager) CreateSession(ctx context.Context, sessionID string, data interface{}) error {
	expiration := 24 * time.Hour // 24小时过期
	return s.redis.SetSession(ctx, sessionID, data, expiration)
}

// GetSession 获取会话
func (s *SessionManager) GetSession(ctx context.Context, sessionID string, obj interface{}) error {
	return s.redis.GetSession(ctx, sessionID, obj)
}

// UpdateSession 更新会话
func (s *SessionManager) UpdateSession(ctx context.Context, sessionID string, data interface{}) error {
	expiration := 24 * time.Hour
	return s.redis.SetSession(ctx, sessionID, data, expiration)
}

// DeleteSession 删除会话
func (s *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	return s.redis.DeleteSession(ctx, sessionID)
}

// ExtendSession 延长会话
func (s *SessionManager) ExtendSession(ctx context.Context, sessionID string) error {
	// 获取当前会话数据
	var sessionData map[string]interface{}
	err := s.GetSession(ctx, sessionID, &sessionData)
	if err != nil {
		return err
	}

	// 重新设置会话，延长过期时间
	expiration := 24 * time.Hour
	return s.redis.SetSession(ctx, sessionID, sessionData, expiration)
}

// 全局服务实例
var (
	redisService              RedisService
	agentCommunicationService *AgentCommunicationService
	sessionManager            *SessionManager
)

// GetRedisService 获取Redis服务实例
func GetRedisService() RedisService {
	if redisService == nil {
		redisService = NewRedisService()
	}
	return redisService
}

// GetAgentCommunicationService 获取Agent通信服务实例
func GetAgentCommunicationService() *AgentCommunicationService {
	if agentCommunicationService == nil {
		agentCommunicationService = NewAgentCommunicationService(GetRedisService())
	}
	return agentCommunicationService
}

// GetSessionManager 获取会话管理器实例
func GetSessionManager() *SessionManager {
	if sessionManager == nil {
		sessionManager = NewSessionManager(GetRedisService())
	}
	return sessionManager
}

package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VectorService 向量服务接口
type VectorService interface {
	CreateVector(ctx context.Context, vector *nesma.NesmaVectorStore) error
	GetVector(ctx context.Context, vectorID string) (*nesma.NesmaVectorStore, error)
	UpdateVector(ctx context.Context, vector *nesma.NesmaVectorStore) error
	DeleteVector(ctx context.Context, vectorID string) error
	SearchSimilar(ctx context.Context, queryVector []float64, limit int, threshold float64) ([]*VectorSearchResult, error)
	SearchByContent(ctx context.Context, content string, limit int) ([]*VectorSearchResult, error)
	BatchCreateVectors(ctx context.Context, vectors []*nesma.NesmaVectorStore) error
	GetVectorsByProject(ctx context.Context, projectID uint) ([]*nesma.NesmaVectorStore, error)
	GetVectorsByUser(ctx context.Context, userID uint) ([]*nesma.NesmaVectorStore, error)
	CreateEmbedding(ctx context.Context, content string, provider string) (*nesma.NesmaEmbedding, error)
	GetEmbedding(ctx context.Context, embeddingID string) (*nesma.NesmaEmbedding, error)
	CalculateSimilarity(vector1, vector2 []float64) (float64, error)
}

// VectorSearchResult 向量搜索结果
type VectorSearchResult struct {
	Vector     *nesma.NesmaVectorStore `json:"vector"`
	Similarity float64                 `json:"similarity"`
	Score      float64                 `json:"score"`
}

// VectorServiceImpl 向量服务实现
type VectorServiceImpl struct {
	db *gorm.DB
}

// ensureDB 确保数据库连接可用，返回安全的数据库实例
func (v *VectorServiceImpl) ensureDB() *gorm.DB {
	if v.db == nil {
		// 尝试重新获取全局连接，但不记录日志（避免初始化时的空指针异常）
		if global.GVA_DB != nil {
			v.db = global.GVA_DB
		} else {
			return nil
		}
	}
	return v.db
}

// NewVectorService 创建向量服务实例
func NewVectorService() VectorService {
	// 安全检查：确保全局数据库连接可用
	// 注意：在服务初始化阶段，global.GVA_LOG可能还未初始化，所以暂时不记录日志
	return &VectorServiceImpl{
		db: global.GVA_DB,
	}
}

// CreateVector 创建向量
func (v *VectorServiceImpl) CreateVector(ctx context.Context, vector *nesma.NesmaVectorStore) error {
	if vector.VectorID == "" {
		vector.VectorID = uuid.New().String()
	}

	// 计算内容哈希
	if vector.Content != "" {
		vector.Hash = utils.MD5V([]byte(vector.Content))
	}

	return v.db.WithContext(ctx).Create(vector).Error
}

// GetVector 获取向量
func (v *VectorServiceImpl) GetVector(ctx context.Context, vectorID string) (*nesma.NesmaVectorStore, error) {
	var vector nesma.NesmaVectorStore
	err := v.db.WithContext(ctx).Where("vector_id = ?", vectorID).First(&vector).Error
	if err != nil {
		return nil, err
	}
	return &vector, nil
}

// UpdateVector 更新向量
func (v *VectorServiceImpl) UpdateVector(ctx context.Context, vector *nesma.NesmaVectorStore) error {
	// 重新计算内容哈希
	if vector.Content != "" {
		vector.Hash = utils.MD5V([]byte(vector.Content))
	}

	return v.db.WithContext(ctx).Where("vector_id = ?", vector.VectorID).Updates(vector).Error
}

// DeleteVector 删除向量
func (v *VectorServiceImpl) DeleteVector(ctx context.Context, vectorID string) error {
	return v.db.WithContext(ctx).Where("vector_id = ?", vectorID).Delete(&nesma.NesmaVectorStore{}).Error
}

// SearchSimilar 相似度搜索
func (v *VectorServiceImpl) SearchSimilar(ctx context.Context, queryVector []float64, limit int, threshold float64) ([]*VectorSearchResult, error) {
	var vectors []nesma.NesmaVectorStore
	err := v.db.WithContext(ctx).Where("is_active = ?", true).Find(&vectors).Error
	if err != nil {
		return nil, err
	}

	var results []*VectorSearchResult
	for _, vector := range vectors {
		if vector.Embedding == "" {
			continue
		}

		// 解析向量
		var embedding []float64
		err := json.Unmarshal([]byte(vector.Embedding), &embedding)
		if err != nil {
			continue
		}

		// 计算相似度
		similarity, err := v.CalculateSimilarity(queryVector, embedding)
		if err != nil {
			continue
		}

		// 过滤低相似度结果
		if similarity < threshold {
			continue
		}

		results = append(results, &VectorSearchResult{
			Vector:     &vector,
			Similarity: similarity,
			Score:      similarity,
		})
	}

	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// 限制结果数量
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// SearchByContent 根据内容搜索
func (v *VectorServiceImpl) SearchByContent(ctx context.Context, content string, limit int) ([]*VectorSearchResult, error) {
	// 安全检查：确保数据库连接可用
	db := v.ensureDB()
	if db == nil {
		// 数据库连接不可用时返回空结果，不记录日志（避免初始化时的空指针异常）
		return []*VectorSearchResult{}, fmt.Errorf("数据库连接未初始化")
	}
	
	var vectors []nesma.NesmaVectorStore
	err := db.WithContext(ctx).Where("content LIKE ? AND is_active = ?", "%"+content+"%", true).Limit(limit).Find(&vectors).Error
	if err != nil {
		return nil, err
	}

	var results []*VectorSearchResult
	for _, vector := range vectors {
		results = append(results, &VectorSearchResult{
			Vector:     &vector,
			Similarity: 1.0, // 内容匹配给予最高相似度
			Score:      1.0,
		})
	}

	return results, nil
}

// BatchCreateVectors 批量创建向量
func (v *VectorServiceImpl) BatchCreateVectors(ctx context.Context, vectors []*nesma.NesmaVectorStore) error {
	for _, vector := range vectors {
		if vector.VectorID == "" {
			vector.VectorID = uuid.New().String()
		}
		if vector.Content != "" {
			vector.Hash = utils.MD5V([]byte(vector.Content))
		}
	}

	return v.db.WithContext(ctx).CreateInBatches(vectors, 100).Error
}

// GetVectorsByProject 获取项目的向量
func (v *VectorServiceImpl) GetVectorsByProject(ctx context.Context, projectID uint) ([]*nesma.NesmaVectorStore, error) {
	var vectors []*nesma.NesmaVectorStore
	err := v.db.WithContext(ctx).Where("project_id = ? AND is_active = ?", projectID, true).Find(&vectors).Error
	return vectors, err
}

// GetVectorsByUser 获取用户的向量
func (v *VectorServiceImpl) GetVectorsByUser(ctx context.Context, userID uint) ([]*nesma.NesmaVectorStore, error) {
	var vectors []*nesma.NesmaVectorStore
	err := v.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&vectors).Error
	return vectors, err
}

// CreateEmbedding 创建嵌入向量
func (v *VectorServiceImpl) CreateEmbedding(ctx context.Context, content string, provider string) (*nesma.NesmaEmbedding, error) {
	// 这里简化实现，实际应该调用AI服务生成嵌入向量
	embedding := &nesma.NesmaEmbedding{
		EmbeddingID: uuid.New().String(),
		Content:     content,
		Vector:      "[]", // 临时占位，实际应该调用AI服务
		Dimensions:  1536,
		Provider:    provider,
		Hash:        utils.MD5V([]byte(content)),
		IsActive:    true,
	}

	err := v.db.WithContext(ctx).Create(embedding).Error
	return embedding, err
}

// GetEmbedding 获取嵌入向量
func (v *VectorServiceImpl) GetEmbedding(ctx context.Context, embeddingID string) (*nesma.NesmaEmbedding, error) {
	var embedding nesma.NesmaEmbedding
	err := v.db.WithContext(ctx).Where("embedding_id = ?", embeddingID).First(&embedding).Error
	if err != nil {
		return nil, err
	}
	return &embedding, nil
}

// CalculateSimilarity 计算两个向量的余弦相似度
func (v *VectorServiceImpl) CalculateSimilarity(vector1, vector2 []float64) (float64, error) {
	if len(vector1) != len(vector2) {
		return 0, errors.New("vectors must have the same dimensions")
	}

	var dotProduct, norm1, norm2 float64

	for i := range vector1 {
		dotProduct += vector1[i] * vector2[i]
		norm1 += vector1[i] * vector1[i]
		norm2 += vector2[i] * vector2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0, errors.New("zero vector")
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2)), nil
}

// VectorIndexService 向量索引服务
type VectorIndexService struct {
	db *gorm.DB
}

// NewVectorIndexService 创建向量索引服务
func NewVectorIndexService() *VectorIndexService {
	return &VectorIndexService{
		db: global.GVA_DB,
	}
}

// CreateIndex 创建向量索引
func (v *VectorIndexService) CreateIndex(ctx context.Context, index *nesma.NesmaVectorIndex) error {
	if index.IndexName == "" {
		return errors.New("index name is required")
	}

	// 检查索引是否已存在
	var existingIndex nesma.NesmaVectorIndex
	err := v.db.WithContext(ctx).Where("index_name = ?", index.IndexName).First(&existingIndex).Error
	if err == nil {
		return errors.New("index already exists")
	}

	// 创建索引记录
	index.Status = "pending"
	err = v.db.WithContext(ctx).Create(index).Error
	if err != nil {
		return err
	}

	// 异步构建索引
	go v.buildIndex(index)

	return nil
}

// buildIndex 构建向量索引
func (v *VectorIndexService) buildIndex(index *nesma.NesmaVectorIndex) {
	// 更新状态为构建中
	v.db.Model(index).Updates(map[string]interface{}{
		"status": "building",
	})

	// 这里应该实现实际的索引构建逻辑
	// 例如创建PostgreSQL的向量索引
	time.Sleep(5 * time.Second) // 模拟构建过程

	// 更新状态为已构建
	v.db.Model(index).Updates(map[string]interface{}{
		"status":     "built",
		"is_built":   true,
		"build_time": time.Now(),
	})
}

// GetIndexes 获取所有索引
func (v *VectorIndexService) GetIndexes(ctx context.Context) ([]*nesma.NesmaVectorIndex, error) {
	var indexes []*nesma.NesmaVectorIndex
	err := v.db.WithContext(ctx).Find(&indexes).Error
	return indexes, err
}

// DeleteIndex 删除索引
func (v *VectorIndexService) DeleteIndex(ctx context.Context, indexName string) error {
	return v.db.WithContext(ctx).Where("index_name = ?", indexName).Delete(&nesma.NesmaVectorIndex{}).Error
}

// 全局向量服务实例
var (
	vectorService      VectorService
	vectorIndexService *VectorIndexService
)

// GetVectorService 获取向量服务实例
func GetVectorService() VectorService {
	if vectorService == nil {
		vectorService = NewVectorService()
	}
	return vectorService
}

// GetVectorIndexService 获取向量索引服务实例
func GetVectorIndexService() *VectorIndexService {
	if vectorIndexService == nil {
		vectorIndexService = NewVectorIndexService()
	}
	return vectorIndexService
}

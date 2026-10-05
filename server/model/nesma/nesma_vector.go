package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaVectorStore 向量存储表
type NesmaVectorStore struct {
	global.GVA_MODEL
	VectorID    string         `json:"vectorId" gorm:"type:varchar(100);not null;unique;comment:向量唯一标识"`
	Content     string         `json:"content" gorm:"type:text;not null;comment:原始内容"`
	Embedding   string         `json:"embedding" gorm:"type:text;comment:向量表示(JSON格式)"`
	Metadata    datatypes.JSON `json:"metadata" gorm:"type:json;comment:元数据"`
	ContentType string         `json:"contentType" gorm:"type:varchar(50);comment:内容类型:text,document,code,knowledge"`
	Source      string         `json:"source" gorm:"type:varchar(200);comment:来源"`
	ProjectID   *uint          `json:"projectId" gorm:"comment:项目ID"`
	UserID      uint           `json:"userId" gorm:"not null;comment:用户ID"`
	Tags        datatypes.JSON `json:"tags" gorm:"type:json;comment:标签"`
	Hash        string         `json:"hash" gorm:"type:varchar(64);comment:内容哈希"`
	Version     int            `json:"version" gorm:"default:1;comment:版本号"`
	IsActive    bool           `json:"isActive" gorm:"default:true;comment:是否激活"`
	ExpiresAt   *time.Time     `json:"expiresAt" gorm:"comment:过期时间"`
}

// TableName 自定义表名
func (NesmaVectorStore) TableName() string {
	return "nesma_vector_stores"
}

// NesmaVectorIndex 向量索引表
type NesmaVectorIndex struct {
	global.GVA_MODEL
	IndexName     string         `json:"indexName" gorm:"type:varchar(100);not null;unique;comment:索引名称"`
	TargetTable   string         `json:"targetTable" gorm:"type:varchar(100);not null;comment:表名"`
	ColumnName    string         `json:"columnName" gorm:"type:varchar(100);not null;comment:列名"`
	Dimensions    int            `json:"dimensions" gorm:"not null;comment:向量维度"`
	IndexType     string         `json:"indexType" gorm:"type:varchar(50);comment:索引类型:cosine,euclidean,dot_product"`
	IndexMethod   string         `json:"indexMethod" gorm:"type:varchar(50);comment:索引方法:ivfflat,hnsw"`
	Configuration datatypes.JSON `json:"configuration" gorm:"type:json;comment:索引配置"`
	IsBuilt       bool           `json:"isBuilt" gorm:"default:false;comment:是否已构建"`
	BuildTime     *time.Time     `json:"buildTime" gorm:"comment:构建时间"`
	Status        string         `json:"status" gorm:"type:varchar(50);default:'pending';comment:状态:pending,building,built,failed"`
}

// TableName 自定义表名
func (NesmaVectorIndex) TableName() string {
	return "nesma_vector_indexes"
}

// NesmaVectorQuery 向量查询记录表
type NesmaVectorQuery struct {
	global.GVA_MODEL
	QueryID      string         `json:"queryId" gorm:"type:varchar(100);not null;unique;comment:查询唯一标识"`
	QueryVector  string         `json:"queryVector" gorm:"type:text;not null;comment:查询向量"`
	QueryContent string         `json:"queryContent" gorm:"type:text;comment:查询内容"`
	ResultCount  int            `json:"resultCount" gorm:"comment:结果数量"`
	Results      datatypes.JSON `json:"results" gorm:"type:json;comment:查询结果"`
	Threshold    float64        `json:"threshold" gorm:"comment:相似度阈值"`
	QueryTime    int            `json:"queryTime" gorm:"comment:查询时间(毫秒)"`
	UserID       uint           `json:"userId" gorm:"not null;comment:用户ID"`
	ProjectID    *uint          `json:"projectId" gorm:"comment:项目ID"`
	SessionID    string         `json:"sessionId" gorm:"type:varchar(255);comment:会话ID"`
	QueryType    string         `json:"queryType" gorm:"type:varchar(50);comment:查询类型:similarity,keyword,hybrid"`
	Metadata     datatypes.JSON `json:"metadata" gorm:"type:json;comment:查询元数据"`
}

// TableName 自定义表名
func (NesmaVectorQuery) TableName() string {
	return "nesma_vector_queries"
}

// NesmaEmbedding 嵌入向量表
type NesmaEmbedding struct {
	global.GVA_MODEL
	EmbeddingID string         `json:"embeddingId" gorm:"type:varchar(100);not null;unique;comment:嵌入向量唯一标识"`
	Content     string         `json:"content" gorm:"type:text;not null;comment:原始内容"`
	Vector      string         `json:"vector" gorm:"type:text;not null;comment:向量数据"`
	Dimensions  int            `json:"dimensions" gorm:"not null;comment:向量维度"`
	Model       string         `json:"model" gorm:"type:varchar(100);comment:使用的模型"`
	TokenCount  int            `json:"tokenCount" gorm:"comment:token数量"`
	Cost        float64        `json:"cost" gorm:"comment:成本"`
	Provider    string         `json:"provider" gorm:"type:varchar(50);comment:提供商"`
	Metadata    datatypes.JSON `json:"metadata" gorm:"type:json;comment:元数据"`
	Hash        string         `json:"hash" gorm:"type:varchar(64);comment:内容哈希"`
	IsActive    bool           `json:"isActive" gorm:"default:true;comment:是否激活"`
}

// TableName 自定义表名
func (NesmaEmbedding) TableName() string {
	return "nesma_embeddings"
}

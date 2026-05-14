package milvus

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"ai_interview/internal/log"
)

const (
	// FieldID 主键字段，存 bank_question.id（UUID 字符串）
	FieldID = "question_id"
	// FieldEmbedding 向量字段
	FieldEmbedding = "embedding"
	// EmbeddingDim 与 DashScope text-embedding-v3 1024 维对齐
	EmbeddingDim = 1024
	// IndexNlist IVF_FLAT 桶数
	IndexNlist = 128
	// IndexNprobe 查询时探测桶数（精度/速度权衡，可按需调大）
	IndexNprobe = 16
)

// Client 封装 Milvus 连接及集合操作。
type Client struct {
	c          client.Client
	collection string
}

// Options Milvus 连接配置。
type Options struct {
	Addr       string // host:port，默认 127.0.0.1:19530
	Collection string // 集合名，默认 bank_questions_vec
}

// New 建立 Milvus gRPC 连接，不初始化集合。
func New(ctx context.Context, opts Options) (*Client, error) {
	c, err := client.NewClient(ctx, client.Config{Address: opts.Addr})
	if err != nil {
		return nil, fmt.Errorf("connect milvus %s: %w", opts.Addr, err)
	}
	log.Infof("[Milvus] connected to %s", opts.Addr)
	return &Client{c: c, collection: opts.Collection}, nil
}

// EnsureCollection 检查集合是否存在，不存在则创建并建索引。
// 幂等：已存在时直接返回 nil。
func (mc *Client) EnsureCollection(ctx context.Context) error {
	exists, err := mc.c.HasCollection(ctx, mc.collection)
	if err != nil {
		return fmt.Errorf("has collection: %w", err)
	}
	if exists {
		log.Infof("[Milvus] collection %q already exists, skip create", mc.collection)
		return nil
	}

	schema := &entity.Schema{
		CollectionName: mc.collection,
		Description:    "bank questions vector index",
		Fields: []*entity.Field{
			{
				Name:       FieldID,
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{"max_length": "64"},
			},
			{
				Name:     FieldEmbedding,
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", EmbeddingDim),
				},
			},
		},
	}

	if err := mc.c.CreateCollection(ctx, schema, 1 /* shards */); err != nil {
		return fmt.Errorf("create collection %q: %w", mc.collection, err)
	}
	log.Infof("[Milvus] collection %q created", mc.collection)

	// IVF_FLAT + COSINE
	idx, err := entity.NewIndexIvfFlat(entity.COSINE, IndexNlist)
	if err != nil {
		return fmt.Errorf("build ivf_flat index param: %w", err)
	}
	if err := mc.c.CreateIndex(ctx, mc.collection, FieldEmbedding, idx, false); err != nil {
		return fmt.Errorf("create index on %q: %w", mc.collection, err)
	}
	log.Infof("[Milvus] IVF_FLAT/COSINE index created on %q.%s (nlist=%d)",
		mc.collection, FieldEmbedding, IndexNlist)

	// 加载到内存，使集合可查
	if err := mc.c.LoadCollection(ctx, mc.collection, false); err != nil {
		return fmt.Errorf("load collection %q: %w", mc.collection, err)
	}
	log.Infof("[Milvus] collection %q loaded into memory", mc.collection)

	return nil
}

// Close 关闭 gRPC 连接。
func (mc *Client) Close() error {
	return mc.c.Close()
}

// RawClient 返回底层 client.Client，供搜索/插入操作使用。
func (mc *Client) RawClient() client.Client {
	return mc.c
}

// Collection 返回配置的集合名。
func (mc *Client) Collection() string {
	return mc.collection
}

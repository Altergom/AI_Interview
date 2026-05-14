package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"

	"ai_interview/internal/log"
)

// bankQuestionsMapping 定义题库 ES 索引的字段类型。
// - question / standard_answer：ik_max_word 中文分词（无 ik 时退回 standard）
// - tags / difficulty：keyword，精确过滤
// - question_id：keyword，对应 PgSQL UUID
const bankQuestionsMapping = `{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0
  },
  "mappings": {
    "properties": {
      "question_id":      { "type": "keyword" },
      "question":         { "type": "text",    "analyzer": "standard" },
      "standard_answer":  { "type": "text",    "analyzer": "standard" },
      "tags":             { "type": "keyword" },
      "difficulty":       { "type": "keyword" }
    }
  }
}`

// Client 封装 Elasticsearch 连接及索引操作。
type Client struct {
	es    *elasticsearch.Client
	index string
}

// Options ES 连接配置。
type Options struct {
	Addrs    []string // 节点地址列表
	Username string
	Password string
	Index    string // 索引名
}

// New 建立 ES 客户端并验证连通性。
func New(_ context.Context, opts Options) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: opts.Addrs,
	}
	if opts.Username != "" {
		cfg.Username = opts.Username
		cfg.Password = opts.Password
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("new es client: %w", err)
	}

	// 连通性探测
	res, err := es.Info()
	if err != nil {
		return nil, fmt.Errorf("ping es: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, fmt.Errorf("ping es: status %s", res.Status())
	}

	log.Infof("[ES] connected to %s", strings.Join(opts.Addrs, ","))
	return &Client{es: es, index: opts.Index}, nil
}

// EnsureIndex 检查索引是否存在，不存在则创建。幂等。
func (c *Client) EnsureIndex(ctx context.Context) error {
	res, err := c.es.Indices.Exists([]string{c.index},
		c.es.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("check index %q: %w", c.index, err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		log.Infof("[ES] index %q already exists, skip create", c.index)
		return nil
	}

	// 创建索引
	createRes, err := c.es.Indices.Create(
		c.index,
		c.es.Indices.Create.WithContext(ctx),
		c.es.Indices.Create.WithBody(bytes.NewBufferString(bankQuestionsMapping)),
	)
	if err != nil {
		return fmt.Errorf("create index %q: %w", c.index, err)
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		return fmt.Errorf("create index %q: %s", c.index, createRes.String())
	}

	log.Infof("[ES] index %q created with mapping", c.index)
	return nil
}

// Index 返回配置的索引名。
func (c *Client) Index() string {
	return c.index
}

// RawClient 返回底层 *elasticsearch.Client，供搜索/写入操作使用。
func (c *Client) RawClient() *elasticsearch.Client {
	return c.es
}

// IndexDocument 写入单条题目文档（upsert）。
func (c *Client) IndexDocument(ctx context.Context, doc map[string]any) error {
	id, _ := doc["question_id"].(string)
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal doc: %w", err)
	}

	res, err := c.es.Index(
		c.index,
		bytes.NewReader(body),
		c.es.Index.WithContext(ctx),
		c.es.Index.WithDocumentID(id),
		c.es.Index.WithRefresh("false"),
	)
	if err != nil {
		return fmt.Errorf("es index doc %s: %w", id, err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es index doc %s: %s", id, res.String())
	}
	return nil
}

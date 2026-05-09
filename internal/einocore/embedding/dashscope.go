// Package embedding 提供文本向量化能力。
// v1 仅接入 DashScope text-embedding-v3（1024 维），由 LlmProviderRegistry 的 QWEN_API_KEY 驱动。
package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai_interview/internal/log"
)

const (
	// DefaultModel DashScope text-embedding-v3，1024 维。
	DefaultModel = "text-embedding-v3"
	// DefaultDim 与 Milvus 集合 EmbeddingDim 对齐。
	DefaultDim = 1024
	// defaultBaseURL DashScope OpenAI 兼容 embedding 端点。
	defaultBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings"
	// defaultTimeout 单次请求超时。
	defaultTimeout = 30 * time.Second
)

// Service 提供文本 → 向量的统一接口。
type Service struct {
	apiKey  string
	baseURL string
	model   string
	dim     int
	client  *http.Client
}

// Options 构造参数。
type Options struct {
	APIKey  string // QWEN_API_KEY 或独立的 DashScope key
	BaseURL string // 覆盖默认端点（可选）
	Model   string // 覆盖模型名（可选）
	Dim     int    // 向量维度（可选，默认 1024）
}

// New 创建 embedding Service。
func New(opts Options) (*Service, error) {
	if opts.APIKey == "" {
		return nil, fmt.Errorf("[Embedding] APIKey is required")
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	model := opts.Model
	if model == "" {
		model = DefaultModel
	}
	dim := opts.Dim
	if dim == 0 {
		dim = DefaultDim
	}
	return &Service{
		apiKey:  opts.APIKey,
		baseURL: baseURL,
		model:   model,
		dim:     dim,
		client:  &http.Client{Timeout: defaultTimeout},
	}, nil
}

// ------- 请求/响应结构 -------

type embedRequest struct {
	Model      string `json:"model"`
	Input      any    `json:"input"`       // string 或 []string
	Dimensions int    `json:"dimensions"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Embed 将单条文本转为向量。
func (s *Service) Embed(ctx context.Context, text string) ([]float32, error) {
	vecs, err := s.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return vecs[0], nil
}

// EmbedBatch 批量文本向量化，返回与输入等长的向量切片。
// DashScope 单次最多 25 条，调用方若超出需自行分批。
func (s *Service) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := embedRequest{
		Model:      s.model,
		Input:      texts,
		Dimensions: s.dim,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("[Embedding] marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("[Embedding] new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[Embedding] http do: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[Embedding] read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[Embedding] status %d: %s", resp.StatusCode, raw)
	}

	var er embedResponse
	if err := json.Unmarshal(raw, &er); err != nil {
		return nil, fmt.Errorf("[Embedding] unmarshal response: %w", err)
	}
	if er.Error != nil {
		return nil, fmt.Errorf("[Embedding] api error %s: %s", er.Error.Code, er.Error.Message)
	}

	// 按 index 排序组装（API 返回顺序不保证）
	result := make([][]float32, len(texts))
	for _, d := range er.Data {
		if d.Index < len(result) {
			result[d.Index] = d.Embedding
		}
	}
	for i, v := range result {
		if v == nil {
			return nil, fmt.Errorf("[Embedding] missing vector for index %d", i)
		}
	}

	log.Debugf("[Embedding] embedded %d texts with model=%s dim=%d", len(texts), s.model, s.dim)
	return result, nil
}

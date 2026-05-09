// Package rag 提供多路召回检索能力。
// 向量召回（Milvus ANN）+ 关键词/标签召回（ES bool query）→ RRF 融合排序。
package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore/embedding"
	"ai_interview/internal/log"
	"ai_interview/internal/storage/es"
	"ai_interview/internal/storage/milvus"
	"ai_interview/internal/storage/postgres"
)

const (
	// rrfK RRF 公式中的平滑常数，标准值 60。
	rrfK = 60
	// defaultANNTopK 单路向量召回数量。
	defaultANNTopK = 20
	// defaultESTopK 单路 ES 召回数量。
	defaultESTopK = 20
)

// Retriever 多路召回检索器。
type Retriever struct {
	repo      *postgres.BankQuestionRepo
	embedSvc  *embedding.Service
	milvusCli *milvus.Client
	esCli     *es.Client
}

// NewRetriever 构造 Retriever。
func NewRetriever(
	repo *postgres.BankQuestionRepo,
	embedSvc *embedding.Service,
	milvusCli *milvus.Client,
	esCli *es.Client,
) *Retriever {
	return &Retriever{
		repo:      repo,
		embedSvc:  embedSvc,
		milvusCli: milvusCli,
		esCli:     esCli,
	}
}

// RetrieveOptions 检索参数。
type RetrieveOptions struct {
	// Query 语义检索文本（通常是技能点描述或面试方向）。
	Query string
	// Tags 标签精确过滤（OR 关系：命中任一 tag 即纳入候选）。
	Tags []string
	// TopK 最终返回数量。
	TopK int
}

// Retrieve 执行多路召回并 RRF 融合，返回 TopK 题目。
func (r *Retriever) Retrieve(ctx context.Context, opts RetrieveOptions) ([]*domain.RetrievedQuestion, error) {
	if opts.TopK <= 0 {
		opts.TopK = 10
	}

	// 并行触发两路召回
	type annResult struct {
		ids []string
		err error
	}
	type esResult struct {
		ids []string
		err error
	}

	annCh := make(chan annResult, 1)
	esCh := make(chan esResult, 1)

	go func() {
		ids, err := r.annSearch(ctx, opts.Query, defaultANNTopK)
		annCh <- annResult{ids, err}
	}()
	go func() {
		ids, err := r.esSearch(ctx, opts.Tags, opts.Query, defaultESTopK)
		esCh <- esResult{ids, err}
	}()

	annRes := <-annCh
	esRes := <-esCh

	// 只要有一路成功，就继续融合
	if annRes.err != nil && esRes.err != nil {
		return nil, fmt.Errorf("both retrieval paths failed: ann=%v es=%v", annRes.err, esRes.err)
	}
	if annRes.err != nil {
		log.Warnf("[RAG] ann search failed, using es only: %v", annRes.err)
	}
	if esRes.err != nil {
		log.Warnf("[RAG] es search failed, using ann only: %v", esRes.err)
	}

	// RRF 融合
	fusedIDs := rrfFuse(annRes.ids, esRes.ids, opts.TopK)
	if len(fusedIDs) == 0 {
		return nil, nil
	}

	// 批量从 PG 拉取完整记录
	results, err := r.fetchByIDs(ctx, fusedIDs)
	if err != nil {
		return nil, fmt.Errorf("[RAG] fetch by ids: %w", err)
	}

	log.Infof("[RAG] retrieved %d questions (query=%q tags=%v)", len(results), opts.Query, opts.Tags)
	return results, nil
}

// ------- 向量召回 -------

// annSearch 执行 Milvus ANN 搜索，返回 question_id 列表（按距离升序）。
func (r *Retriever) annSearch(ctx context.Context, query string, topK int) ([]string, error) {
	if query == "" {
		return nil, nil
	}

	vec, err := r.embedSvc.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	sp, err := entity.NewIndexIvfFlatSearchParam(milvus.IndexNprobe)
	if err != nil {
		return nil, fmt.Errorf("build search param: %w", err)
	}

	results, err := r.milvusCli.RawClient().Search(
		ctx,
		r.milvusCli.Collection(),
		nil, // partitions
		"",  // expr（不过滤）
		[]string{milvus.FieldID}, // output fields
		[]entity.Vector{entity.FloatVector(vec)},
		milvus.FieldEmbedding,
		entity.COSINE,
		topK,
		sp,
		client.WithSearchQueryConsistencyLevel(entity.ClBounded),
	)
	if err != nil {
		return nil, fmt.Errorf("milvus search: %w", err)
	}

	var ids []string
	for _, res := range results {
		for i := 0; i < res.ResultCount; i++ {
			val, err := res.IDs.Get(i)
			if err != nil {
				continue
			}
			if id, ok := val.(string); ok {
				ids = append(ids, id)
			}
		}
	}
	return ids, nil
}

// ------- ES 召回 -------

// esSearchBody ES bool query 请求体结构。
type esSearchBody struct {
	Query esQuery `json:"query"`
	Size  int     `json:"size"`
	// 只返回 _id，不需要 _source 字段（减少网络传输）
	Source bool `json:"_source"`
}

type esQuery struct {
	Bool esBoolQuery `json:"bool"`
}

type esBoolQuery struct {
	Should  []any `json:"should,omitempty"`
	Filter  []any `json:"filter,omitempty"`
	MinShould int  `json:"minimum_should_match,omitempty"`
}

type esTermsClause struct {
	Terms map[string][]string `json:"terms"`
}

type esMatchClause struct {
	Match map[string]esMatchField `json:"match"`
}

type esMatchField struct {
	Query string `json:"query"`
}

// esSearchResponse ES 响应结构（只取需要的字段）。
type esSearchResponse struct {
	Hits struct {
		Hits []struct {
			ID string `json:"_id"` // 文档 ID = question_id
		} `json:"hits"`
	} `json:"hits"`
}

// esSearch 执行 ES bool query：tags terms filter + question match，返回 question_id 列表。
func (r *Retriever) esSearch(ctx context.Context, tags []string, query string, size int) ([]string, error) {
	body := esSearchBody{
		Size:   size,
		Source: false,
	}

	var shouldClauses []any
	if len(tags) > 0 {
		shouldClauses = append(shouldClauses, esTermsClause{
			Terms: map[string][]string{"tags": tags},
		})
	}
	if query != "" {
		shouldClauses = append(shouldClauses, esMatchClause{
			Match: map[string]esMatchField{"question": {Query: query}},
		})
	}

	if len(shouldClauses) == 0 {
		return nil, nil
	}

	body.Query = esQuery{
		Bool: esBoolQuery{
			Should:    shouldClauses,
			MinShould: 1,
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal es body: %w", err)
	}

	raw := r.esCli.RawClient()
	res, err := raw.Search(
		raw.Search.WithContext(ctx),
		raw.Search.WithIndex(r.esCli.Index()),
		raw.Search.WithBody(bytes.NewReader(bodyBytes)),
	)
	if err != nil {
		return nil, fmt.Errorf("es search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("es search status %s: %s", res.Status(), body)
	}

	var esResp esSearchResponse
	if err := json.NewDecoder(res.Body).Decode(&esResp); err != nil {
		return nil, fmt.Errorf("decode es response: %w", err)
	}

	ids := make([]string, 0, len(esResp.Hits.Hits))
	for _, h := range esResp.Hits.Hits {
		ids = append(ids, h.ID)
	}
	return ids, nil
}

// ------- RRF 融合 -------

// rrfFuse 对两路有序 ID 列表做 Reciprocal Rank Fusion，返回融合后 TopK 的 ID 列表。
// 公式：score(d) = Σ 1 / (k + rank_i)
func rrfFuse(annIDs, esIDs []string, topK int) []string {
	scores := make(map[string]float64)

	for rank, id := range annIDs {
		scores[id] += 1.0 / float64(rrfK+rank+1)
	}
	for rank, id := range esIDs {
		scores[id] += 1.0 / float64(rrfK+rank+1)
	}

	type scored struct {
		id    string
		score float64
	}
	list := make([]scored, 0, len(scores))
	for id, s := range scores {
		list = append(list, scored{id, s})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].score > list[j].score
	})

	if topK > len(list) {
		topK = len(list)
	}
	ids := make([]string, topK)
	for i := range ids {
		ids[i] = list[i].id
	}
	return ids
}

// ------- PG 批量拉取 -------

// fetchByIDs 保序批量拉取 PG 记录，附带 RRF 名次作为近似分数。
func (r *Retriever) fetchByIDs(ctx context.Context, ids []string) ([]*domain.RetrievedQuestion, error) {
	// 逐个查（TopK 通常 ≤ 20，不值得写 IN 查询的额外复杂度）
	results := make([]*domain.RetrievedQuestion, 0, len(ids))
	for rank, id := range ids {
		rec, err := r.repo.GetByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get question %s: %w", id, err)
		}
		if rec == nil {
			continue // 向量库中存在但 PG 已删除，跳过
		}
		results = append(results, &domain.RetrievedQuestion{
			BankQuestionRecord: *rec,
			Score:              1.0 / float64(rrfK+rank+1), // 近似分数（仅用于调试）
		})
	}
	return results, nil
}

// Package worker 提供题目向量化异步 Worker。
// 消费 RabbitMQ vectorize_task 队列：
//  1. 从 PG 读取题目完整记录
//  2. 调 DashScope embedding 生成 1024 维向量
//  3. 写入 Milvus（ANN 索引）
//  4. 写入 ES（关键词/标签检索）
//  5. 更新 PG vec_status = done；失败时标记 failed
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"

	"ai_interview/internal/domain"
	"ai_interview/internal/einocore/embedding"
	"ai_interview/internal/log"
	"ai_interview/internal/mq"
	"ai_interview/internal/storage/es"
	"ai_interview/internal/storage/milvus"
	"ai_interview/internal/storage/postgres"
)

// VectorizeWorker 消费 vectorize_task 队列并完成向量化写入。
type VectorizeWorker struct {
	repo      *postgres.BankQuestionRepo
	embedSvc  *embedding.Service
	milvusCli *milvus.Client
	esCli     *es.Client
	msgs      <-chan amqp.Delivery
}

// NewVectorizeWorker 构造 Worker，msgs 由调用方（app 层）通过 mqclient.Consume 提供。
func NewVectorizeWorker(
	repo *postgres.BankQuestionRepo,
	embedSvc *embedding.Service,
	milvusCli *milvus.Client,
	esCli *es.Client,
	msgs <-chan amqp.Delivery,
) *VectorizeWorker {
	return &VectorizeWorker{
		repo:      repo,
		embedSvc:  embedSvc,
		milvusCli: milvusCli,
		esCli:     esCli,
		msgs:      msgs,
	}
}

// Run 启动消费循环，ctx 取消时退出。
// 建议在 goroutine 中调用：go worker.Run(ctx)
func (w *VectorizeWorker) Run(ctx context.Context) {
	log.Infof("[VectorizeWorker] started, waiting for tasks on queue=%s", mq.TopicVectorizeTask)
	for {
		select {
		case <-ctx.Done():
			log.Infof("[VectorizeWorker] context done, shutting down")
			return
		case d, ok := <-w.msgs:
			if !ok {
				log.Warnf("[VectorizeWorker] delivery channel closed, stopping")
				return
			}
			w.handle(ctx, d)
		}
	}
}

// handle 处理单条消息，最多重试 3 次；仍失败则标 failed 并 ACK 丢弃（死信队列留 v2）。
func (w *VectorizeWorker) handle(ctx context.Context, d amqp.Delivery) {
	var task mq.VectorizeTask
	if err := json.Unmarshal(d.Body, &task); err != nil {
		log.Errorf("[VectorizeWorker] unmarshal task: %v, body=%s", err, d.Body)
		d.Ack(false) // 消息格式错误，直接丢弃
		return
	}

	log.Infof("[VectorizeWorker] processing question_id=%s", task.QuestionID)

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if err := w.process(ctx, task.QuestionID); err != nil {
			lastErr = err
			log.Warnf("[VectorizeWorker] attempt %d/3 failed question_id=%s: %v", attempt, task.QuestionID, err)
			time.Sleep(time.Duration(attempt) * 2 * time.Second) // 简化指数退避
			continue
		}
		d.Ack(false)
		log.Infof("[VectorizeWorker] done question_id=%s", task.QuestionID)
		return
	}

	// 3 次全败，标记 failed，ACK 丢弃（v2 补死信队列）
	log.Errorf("[VectorizeWorker] all retries failed question_id=%s: %v", task.QuestionID, lastErr)
	_ = w.repo.UpdateVecStatus(ctx, task.QuestionID, domain.VecStatusFailed)
	d.Ack(false)
}

// process 执行完整的向量化写入流程。
func (w *VectorizeWorker) process(ctx context.Context, questionID string) error {
	// 1. 从 PG 读取完整记录
	rec, err := w.repo.GetByID(ctx, questionID)
	if err != nil {
		return fmt.Errorf("get question from pg: %w", err)
	}
	if rec == nil {
		// 题目被删除，直接 skip
		log.Warnf("[VectorizeWorker] question not found in pg, skipping question_id=%s", questionID)
		return nil
	}
	if rec.VecStatus == domain.VecStatusDone {
		log.Debugf("[VectorizeWorker] question already vectorized, skip question_id=%s", questionID)
		return nil
	}

	// 2. 生成 embedding（题目+标答拼接，提升语义质量）
	textToEmbed := rec.Question
	if rec.StandardAnswer != "" {
		textToEmbed = rec.Question + "\n" + rec.StandardAnswer
	}
	vec, err := w.embedSvc.Embed(ctx, textToEmbed)
	if err != nil {
		return fmt.Errorf("embed question: %w", err)
	}

	// 3. 写 Milvus
	if err := w.upsertMilvus(ctx, rec, vec); err != nil {
		return fmt.Errorf("upsert milvus: %w", err)
	}

	// 4. 写 ES
	if err := w.upsertES(ctx, rec); err != nil {
		return fmt.Errorf("upsert es: %w", err)
	}

	// 5. 更新 PG vec_status = done
	if err := w.repo.UpdateVecStatus(ctx, questionID, domain.VecStatusDone); err != nil {
		return fmt.Errorf("update vec_status: %w", err)
	}
	return nil
}

// upsertMilvus 将向量写入 Milvus。
// SDK v2 不支持原生 upsert，使用 DeleteByPks + Insert 实现幂等写入。
func (w *VectorizeWorker) upsertMilvus(ctx context.Context, rec *domain.BankQuestionRecord, vec []float32) error {
	cli := w.milvusCli.RawClient()
	col := w.milvusCli.Collection()

	// 尝试先删（忽略不存在/空集合错误）
	pkCol := entity.NewColumnVarChar(milvus.FieldID, []string{rec.ID})
	_ = cli.DeleteByPks(ctx, col, "", pkCol)

	// 插入
	idCol := entity.NewColumnVarChar(milvus.FieldID, []string{rec.ID})
	vecCol := entity.NewColumnFloatVector(milvus.FieldEmbedding, milvus.EmbeddingDim, [][]float32{vec})
	_, err := cli.Insert(ctx, col, "", idCol, vecCol)
	if err != nil {
		return fmt.Errorf("milvus insert question_id=%s: %w", rec.ID, err)
	}
	return nil
}

// upsertES 将题目文档写入 ES（IndexDocument 内部使用 question_id 作为文档 ID，天然幂等）。
func (w *VectorizeWorker) upsertES(ctx context.Context, rec *domain.BankQuestionRecord) error {
	doc := map[string]any{
		"question_id":     rec.ID,
		"question":        rec.Question,
		"standard_answer": rec.StandardAnswer,
		"tags":            rec.Tags,
		"difficulty":      string(rec.Difficulty),
	}
	return w.esCli.IndexDocument(ctx, doc)
}

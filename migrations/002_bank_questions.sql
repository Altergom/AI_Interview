-- 题库元数据表（PgSQL 只存结构化字段，向量数据在 Milvus，全文索引在 ES）
-- vec_status 标记异步写入 Milvus+ES 的状态，由 Worker 维护。

CREATE TABLE IF NOT EXISTS bank_questions (
    id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    question             TEXT         NOT NULL,
    standard_answer      TEXT         NOT NULL DEFAULT '',
    -- 技术标签，如 ["golang", "goroutine", "channel"]
    tags                 JSONB        NOT NULL DEFAULT '[]',
    -- 关联概念，如 ["context", "select", "WaitGroup"]
    related_concepts     JSONB        NOT NULL DEFAULT '[]',
    -- 追问题目 ID 列表（引用本表其他题目）
    followup_question_ids JSONB       NOT NULL DEFAULT '[]',
    difficulty           VARCHAR(16)  NOT NULL DEFAULT 'medium'
                             CHECK (difficulty IN ('easy', 'medium', 'hard')),
    -- 异步向量化状态：pending / done / failed
    vec_status           VARCHAR(16)  NOT NULL DEFAULT 'pending'
                             CHECK (vec_status IN ('pending', 'done', 'failed')),
    created_at           TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- 按标签检索（GIN 索引支持 @> 包含查询）
CREATE INDEX IF NOT EXISTS idx_bank_questions_tags
    ON bank_questions USING GIN (tags);

-- 按方向+状态批量拉取待向量化任务
CREATE INDEX IF NOT EXISTS idx_bank_questions_vec_status
    ON bank_questions (vec_status, created_at);

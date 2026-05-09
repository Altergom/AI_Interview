-- 通用异步任务表，供 report_generate / vectorize 等 Worker 共用。
-- 状态机：pending → processing → completed | failed
-- retry_count 由 Worker 自增；v1 失败直接标 failed（死信队列推 v2）。

CREATE TABLE IF NOT EXISTS async_tasks (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    type          VARCHAR(64)  NOT NULL,
    ref_id        VARCHAR(64)  NOT NULL,   -- 关联业务 ID，如 interview_id / question_id
    status        VARCHAR(16)  NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    retry_count   INT          NOT NULL DEFAULT 0,
    error_message TEXT         NOT NULL DEFAULT '',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- 按业务实体快速查任务（如查某次面试的报告任务状态）
CREATE INDEX IF NOT EXISTS idx_async_tasks_ref_id  ON async_tasks (type, ref_id);
-- Worker 拉取待处理任务
CREATE INDEX IF NOT EXISTS idx_async_tasks_pending ON async_tasks (status, created_at)
    WHERE status = 'pending';

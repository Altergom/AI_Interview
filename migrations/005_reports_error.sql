-- reports 表补 error_message 列。
-- 失败时 Worker 写入原因，前端 GET /report 查到非空即知生成失败。
-- 正常情况由 WebSocket 推送通知，不需要轮询 status。

ALTER TABLE reports ADD COLUMN IF NOT EXISTS error_message TEXT NOT NULL DEFAULT '';

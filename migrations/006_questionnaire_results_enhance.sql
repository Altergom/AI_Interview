-- 006_questionnaire_results_enhance.sql
-- 目的：为 SFT 问卷标注 v1 增强 questionnaire_results 表
-- 约定：不使用数据库外键，引用完整性由应用层保证

-- 1) 新增字段
ALTER TABLE questionnaire_results
  ADD COLUMN IF NOT EXISTS user_id UUID,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
  ADD COLUMN IF NOT EXISTS exported_at TIMESTAMP NULL;

-- 2) 幂等：一个 (interview, turn) 只能有一条标注，重复提交走 UPSERT 更新
CREATE UNIQUE INDEX IF NOT EXISTS uq_questionnaire_interview_turn
  ON questionnaire_results(interview_id, turn_id);

-- 3) 按用户查询（个人标注历史 / 权限审计）
CREATE INDEX IF NOT EXISTS idx_questionnaire_user_id
  ON questionnaire_results(user_id);

-- 4) v2 增量导出 worker 用：只索引未导出的行，体积小
CREATE INDEX IF NOT EXISTS idx_questionnaire_exported_at
  ON questionnaire_results(exported_at)
  WHERE exported_at IS NULL;

-- 5) 时间维度统计 / 增量拉取
CREATE INDEX IF NOT EXISTS idx_questionnaire_created_at
  ON questionnaire_results(created_at);

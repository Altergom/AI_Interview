-- seed_questionnaire.sql
-- 用途：本地手工验证 SFT 问卷标注表结构与索引
-- 用法：psql ... -f migrations/seed_questionnaire.sql
-- 注意：本脚本只用于开发环境验证，禁止在生产执行

BEGIN;

-- ============================================================
-- 1) 准备一个测试用户 + 一场面试 + 5 个 turn
-- ============================================================

-- 固定 UUID 便于反复验证（删后重跑幂等）
DO $$
DECLARE
  v_user_id      UUID := '11111111-1111-1111-1111-111111111111';
  v_interview_id UUID := '22222222-2222-2222-2222-222222222222';
BEGIN
  -- 清理旧 seed 数据（按反向依赖顺序）
  DELETE FROM questionnaire_results WHERE interview_id = v_interview_id;
  DELETE FROM interview_turns       WHERE interview_id = v_interview_id;
  DELETE FROM interviews            WHERE id           = v_interview_id;
  DELETE FROM users                 WHERE id           = v_user_id;

  -- 用户
  INSERT INTO users (id, email, password_hash, created_at)
  VALUES (v_user_id, 'sft-seed@example.com', 'bcrypt$placeholder', NOW());

  -- 面试
  INSERT INTO interviews (id, user_id, started_at, ended_at, status)
  VALUES (v_interview_id, v_user_id, NOW() - INTERVAL '30 minutes', NOW(), 'finished');

  -- 5 个 turn：覆盖 intro / questioning / closing 三个阶段
  INSERT INTO interview_turns (id, interview_id, turn_id, stage, question, user_answer, asr_raw, created_at) VALUES
    (gen_random_uuid(), v_interview_id, 'T01', 'intro',       '请做一下自我介绍',                 '我叫张三，五年 Go 后端经验……',          '我叫 张三 五年 Go 后端经验',          NOW() - INTERVAL '25 min'),
    (gen_random_uuid(), v_interview_id, 'T02', 'questioning', 'GMP 模型中 P 的作用是什么？',       'P 是逻辑处理器，持有可运行 G 的队列……', 'P 是逻辑处理器 持有可运行 G 的队列', NOW() - INTERVAL '20 min'),
    (gen_random_uuid(), v_interview_id, 'T03', 'questioning', 'channel 关闭后再写会发生什么？',    '会 panic，关闭的 channel 不能再写入',   '会 panic 关闭的 channel 不能写入',    NOW() - INTERVAL '15 min'),
    (gen_random_uuid(), v_interview_id, 'T04', 'questioning', '你怎么排查 goroutine 泄漏？',       'pprof goroutine profile 看堆栈……',      'pprof goroutine profile 看堆栈',      NOW() - INTERVAL '10 min'),
    (gen_random_uuid(), v_interview_id, 'T05', 'closing',     '你还有什么想问公司的？',            '团队规模和技术栈演进方向',              '团队规模 和 技术栈 演进方向',         NOW() - INTERVAL '5  min');
END $$;

-- ============================================================
-- 2) 插入问卷标注：good / bad 混合，覆盖 DPO 正负样本场景
-- ============================================================
DO $$
DECLARE
  v_user_id      UUID := '11111111-1111-1111-1111-111111111111';
  v_interview_id UUID := '22222222-2222-2222-2222-222222222222';
BEGIN
  INSERT INTO questionnaire_results
    (id, interview_id, turn_id, quality, feedback, user_id, created_at, updated_at, exported_at)
  VALUES
    (gen_random_uuid(), v_interview_id, 'T01', 'good', '提问开放，节奏合适',                v_user_id, NOW(), NOW(), NULL),
    (gen_random_uuid(), v_interview_id, 'T02', 'good', '追问到位，覆盖了 M-P 绑定细节',     v_user_id, NOW(), NOW(), NULL),
    (gen_random_uuid(), v_interview_id, 'T03', 'bad',  '追问偏离主线，跑去问 select 语法',  v_user_id, NOW(), NOW(), NULL),
    (gen_random_uuid(), v_interview_id, 'T04', 'bad',  '没有要求结合实战案例，过于书面',    v_user_id, NOW(), NOW(), NULL);
  -- T05 故意不标，验证「允许部分提交」
END $$;

COMMIT;

-- ============================================================
-- 3) 验证查询：跑完后人工 check
-- ============================================================

-- 3.1 应返回 4 行（T01-T04），T05 无标注
SELECT turn_id, quality, LEFT(feedback, 20) AS feedback_preview, exported_at
FROM questionnaire_results
WHERE interview_id = '22222222-2222-2222-2222-222222222222'
ORDER BY turn_id;

-- 3.2 验证 UPSERT：再插同一个 (interview_id, turn_id=T01) 应触发更新，行数不变
INSERT INTO questionnaire_results
  (id, interview_id, turn_id, quality, feedback, user_id, created_at, updated_at)
VALUES
  (gen_random_uuid(), '22222222-2222-2222-2222-222222222222', 'T01', 'bad',
   '改主意了：开场太啰嗦', '11111111-1111-1111-1111-111111111111', NOW(), NOW())
ON CONFLICT (interview_id, turn_id)
DO UPDATE SET quality   = EXCLUDED.quality,
              feedback  = EXCLUDED.feedback,
              updated_at = NOW();

-- 3.3 验证完后 T01 应变为 'bad'，总行数仍为 4
SELECT COUNT(*) AS total,
       SUM(CASE WHEN quality='good' THEN 1 ELSE 0 END) AS good_count,
       SUM(CASE WHEN quality='bad'  THEN 1 ELSE 0 END) AS bad_count
FROM questionnaire_results
WHERE interview_id = '22222222-2222-2222-2222-222222222222';

-- 3.4 验证「未导出」索引：v2 增量导出 worker 拉取范围
SELECT turn_id, quality
FROM questionnaire_results
WHERE interview_id = '22222222-2222-2222-2222-222222222222'
  AND exported_at IS NULL
ORDER BY created_at;

-- 3.5 多轮 conversation 形态预览（导出 JSONL 时的真实 join）
SELECT t.turn_id,
       t.stage,
       t.question,
       t.user_answer,
       q.quality,
       q.feedback
FROM interview_turns t
LEFT JOIN questionnaire_results q
       ON q.interview_id = t.interview_id
      AND q.turn_id      = t.turn_id
WHERE t.interview_id = '22222222-2222-2222-2222-222222222222'
ORDER BY t.turn_id;

-- resumes 表：存储结构化简历，按 content_hash 去重。
-- content_hash: 对提取的 PDF 原始文本做 SHA-256，相同文本命中缓存直接返回。
-- parsed_data:  JSON 格式存储 StructuredResume（skills / projects / internships / education）。

CREATE TABLE IF NOT EXISTS resumes (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      VARCHAR(64) NOT NULL,
  content_hash CHAR(64)    NOT NULL,          -- SHA-256 hex，64 字符
  parsed_data  JSONB       NOT NULL,
  created_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- 去重索引：同一 content_hash 只保留一条，跨用户复用
CREATE UNIQUE INDEX IF NOT EXISTS idx_resumes_content_hash ON resumes(content_hash);

-- 按用户查询索引
CREATE INDEX IF NOT EXISTS idx_resumes_user_id ON resumes(user_id);

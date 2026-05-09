-- resumes 表补充 s3_key 列，记录原始 PDF 在 S3 的对象路径。
-- s3_key 用于追溯原始文件，不作唯一约束（同一文件可由不同用户上传，路径不同）。

ALTER TABLE resumes ADD COLUMN IF NOT EXISTS s3_key TEXT NOT NULL DEFAULT '';

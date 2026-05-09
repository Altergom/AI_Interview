-- 补充 users 表：用户名、游客标记
-- username: 注册时填写的展示名称；游客用 guest_{uuid 前8位} 生成
-- is_guest: true 表示游客账号，JWT 有效期 24h

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS username  VARCHAR(100) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS is_guest  BOOLEAN      NOT NULL DEFAULT FALSE;

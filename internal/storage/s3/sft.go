package s3

import (
	"context"
	"io"
	"time"
)

// UploadSFTJSONL 上传 SFT 训练数据 JSONL 文件到 S3。
// 路径: /sft/{date}/{filename}.jsonl
func (c *Client) UploadSFTJSONL(ctx context.Context, t time.Time, filename string, body io.Reader) error {
	// TODO: 实现 JSONL 上传
	// 1. 使用 paths.SFTJSONLPrefix() 生成前缀
	// 2. 拼接 filename
	// 3. 调用 Upload()，contentType: application/jsonl
	return nil
}

// ListSFTJSONL 列出指定日期的 JSONL 文件。
func (c *Client) ListSFTJSONL(ctx context.Context, t time.Time) ([]string, error) {
	// TODO: 实现列表查询
	// 1. 使用 paths.SFTJSONLPrefix() 生成前缀
	// 2. ListObjectsV2 with prefix
	return nil, nil
}

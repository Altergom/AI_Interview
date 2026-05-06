package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadSFTJSONL 上传 SFT 训练数据 JSONL 文件到 S3。
// 路径: sft/{date}/{filename}.jsonl
func (c *Client) UploadSFTJSONL(ctx context.Context, t time.Time, filename string, body io.Reader) error {
	key := SFTJSONLPrefix(t) + filename + ".jsonl"
	return c.Upload(ctx, key, body, "application/jsonl")
}

// ListSFTJSONL 列出指定日期的 JSONL 文件 key 列表。
func (c *Client) ListSFTJSONL(ctx context.Context, t time.Time) ([]string, error) {
	prefix := SFTJSONLPrefix(t)
	out, err := c.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects %q: %w", prefix, err)
	}

	keys := make([]string, 0, len(out.Contents))
	for _, obj := range out.Contents {
		keys = append(keys, aws.ToString(obj.Key))
	}
	return keys, nil
}

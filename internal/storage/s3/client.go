package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client 封装 S3 客户端（兼容 MinIO）。
type Client struct {
	s3Client *s3.Client
	bucket   string
}

// Options S3 连接配置。
type Options struct {
	Endpoint  string // MinIO: http://localhost:9000
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

// New 初始化 S3 客户端。
func New(ctx context.Context, opts Options) (*Client, error) {
	// TODO: 实现 S3 客户端初始化
	// 1. 配置 aws-sdk-go-v2
	// 2. 支持自定义 endpoint（MinIO）
	// 3. 配置 credentials
	return nil, nil
}

// Close 关闭 S3 客户端（如有需要）。
func (c *Client) Close() error {
	// S3 客户端通常无需显式关闭，预留接口
	return nil
}

// Upload 通用上传方法。
func (c *Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	// TODO: 实现上传逻辑
	// 1. PutObject
	// 2. 设置 ContentType
	return nil
}

// Download 通用下载方法。
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// TODO: 实现下载逻辑
	return nil, nil
}

// Delete 删除对象。
func (c *Client) Delete(ctx context.Context, key string) error {
	// TODO: 实现删除逻辑
	return nil
}

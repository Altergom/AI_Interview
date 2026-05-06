package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client 封装 S3 客户端（兼容阿里云 OSS / MinIO）。
type Client struct {
	s3Client *s3.Client
	bucket   string
}

// Options S3 连接配置。
type Options struct {
	Endpoint  string // 阿里云: https://oss-cn-hangzhou.aliyuncs.com  MinIO: http://localhost:9000
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

// New 初始化 S3 客户端。
func New(ctx context.Context, opts Options) (*Client, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               opts.Endpoint,
				HostnameImmutable: true,
			}, nil
		},
	)

	cfg := aws.Config{
		Region:                      opts.Region,
		Credentials:                 credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.SecretKey, ""),
		EndpointResolverWithOptions: resolver,
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // MinIO 必须开启 path-style，OSS 兼容
	})

	// 验证 bucket 可访问
	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(opts.Bucket),
	}); err != nil {
		return nil, fmt.Errorf("head bucket %q: %w", opts.Bucket, err)
	}

	return &Client{s3Client: client, bucket: opts.Bucket}, nil
}

// Upload 通用上传方法。
func (c *Client) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

// Download 通用下载方法。
func (c *Client) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	return out.Body, nil
}

// Delete 删除对象。
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}

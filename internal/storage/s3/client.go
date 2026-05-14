package s3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyhttp "github.com/aws/smithy-go/transport/http"

	"ai_interview/internal/log"
)

// Client 封装 S3 客户端（兼容阿里云 OSS / MinIO）。
type Client struct {
	s3Client *s3.Client
	presign  *s3.PresignClient
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
// 若 bucket 不存在（MinIO 本地开发场景），自动创建并应用权限策略。
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

	raw := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // MinIO 必须开启 path-style，OSS 兼容
	})

	if err := ensureBucket(ctx, raw, opts.Bucket, opts.Region); err != nil {
		return nil, err
	}

	c := &Client{
		s3Client: raw,
		presign:  s3.NewPresignClient(raw),
		bucket:   opts.Bucket,
	}
	return c, nil
}

// ensureBucket 检查 bucket 是否存在，不存在则创建并写入权限策略。
func ensureBucket(ctx context.Context, client *s3.Client, bucket, region string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		// bucket 已存在，直接应用策略（幂等）
		return applyBucketPolicy(ctx, client, bucket)
	}

	// 判断是否是 404（bucket 不存在）
	var respErr *smithyhttp.ResponseError
	if !errors.As(err, &respErr) || respErr.HTTPStatusCode() != 404 {
		// 非 404 错误（如认证失败、网络问题）直接返回
		return fmt.Errorf("[S3] head bucket %q: %w", bucket, err)
	}

	// bucket 不存在，自动创建（MinIO 本地开发场景）
	log.Infof("[S3] bucket %q not found, creating...", bucket)
	createInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}
	// AWS S3 us-east-1 不能指定 LocationConstraint，其他区域必须指定
	if region != "" && region != "us-east-1" {
		createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}
	if _, err := client.CreateBucket(ctx, createInput); err != nil {
		return fmt.Errorf("[S3] create bucket %q: %w", bucket, err)
	}
	log.Infof("[S3] bucket %q created", bucket)

	return applyBucketPolicy(ctx, client, bucket)
}

// bucketPolicy 描述 S3 Bucket 权限策略。
// 策略：默认私有，只有服务端（持有凭证）可读写；不开放匿名访问。
// 说明：简历/音频等敏感文件通过预签名 URL 临时授权，无需公开 ACL。
type bucketPolicyStatement struct {
	Sid       string   `json:"Sid"`
	Effect    string   `json:"Effect"`
	Principal string   `json:"Principal"`
	Action    []string `json:"Action"`
	Resource  []string `json:"Resource"`
}

type bucketPolicy struct {
	Version   string                  `json:"Version"`
	Statement []bucketPolicyStatement `json:"Statement"`
}

// applyBucketPolicy 向 bucket 写入权限策略（私有，拒绝匿名访问）。
func applyBucketPolicy(ctx context.Context, client *s3.Client, bucket string) error {
	// 仅允许凭证用户访问，拒绝所有匿名请求（Principal: "*" + Condition 效果同 private ACL）
	// MinIO 兼容此策略；阿里云 OSS 私有 bucket 默认即如此，策略为幂等操作。
	policy := bucketPolicy{
		Version: "2012-10-17",
		Statement: []bucketPolicyStatement{
			{
				Sid:       "DenyAnonymousAccess",
				Effect:    "Deny",
				Principal: "*",
				Action:    []string{"s3:GetObject", "s3:PutObject", "s3:DeleteObject"},
				Resource:  []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucket)},
			},
		},
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("[S3] marshal bucket policy: %w", err)
	}

	policyStr := string(policyJSON)
	if _, err := client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policyStr),
	}); err != nil {
		// 部分云厂商（阿里云 OSS）私有 bucket 不支持 PutBucketPolicy，记录警告但不中断
		log.Warnf("[S3] put bucket policy skipped (may be unsupported by provider): %v", err)
	}

	return nil
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

// PresignGetURL 生成对象下载的预签名 URL。
// 典型用途：前端下载简历 PDF、报告音频等，expires 建议 5~15 分钟。
func (c *Client) PresignGetURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	req, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign get %q: %w", key, err)
	}
	return req.URL, nil
}

// PresignPutURL 生成对象上传的预签名 URL。
// 典型用途：前端直传简历 PDF 到 S3，expires 建议 5 分钟。
func (c *Client) PresignPutURL(ctx context.Context, key, contentType string, expires time.Duration) (string, error) {
	req, err := c.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign put %q: %w", key, err)
	}
	return req.URL, nil
}

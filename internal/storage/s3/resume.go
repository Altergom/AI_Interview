package s3

import (
	"context"
	"io"
)

// UploadResume 上传简历文件到 S3。
// 路径: /resumes/{user_id}/{filename}
func (c *Client) UploadResume(ctx context.Context, userID, filename string, body io.Reader) error {
	// TODO: 实现简历上传
	// 1. 使用 paths.ResumeObjectKey() 生成 key
	// 2. 调用 Upload()，contentType: application/pdf
	return nil
}

// DownloadResume 下载简历文件。
func (c *Client) DownloadResume(ctx context.Context, userID, filename string) (io.ReadCloser, error) {
	// TODO: 实现简历下载
	return nil, nil
}

package s3

import (
	"context"
	"io"
)

// UploadResume 上传简历 PDF 到 S3。
// 路径: resumes/{user_id}/{filename}
func (c *Client) UploadResume(ctx context.Context, userID, filename string, body io.Reader) error {
	return c.Upload(ctx, ResumeObjectKey(userID, filename), body, "application/pdf")
}

// DownloadResume 下载简历文件。
func (c *Client) DownloadResume(ctx context.Context, userID, filename string) (io.ReadCloser, error) {
	return c.Download(ctx, ResumeObjectKey(userID, filename))
}

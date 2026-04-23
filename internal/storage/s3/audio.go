package s3

import (
	"context"
	"io"
)

// UploadAudio 上传音频文件到 S3。
// 路径: /audio/{interview_id}/{turn_id}.wav
func (c *Client) UploadAudio(ctx context.Context, interviewID, turnID string, body io.Reader) error {
	// TODO: 实现音频上传
	// 1. 使用 paths.AudioObjectKey() 生成 key
	// 2. 调用 Upload()，contentType: audio/wav
	return nil
}

// DownloadAudio 下载音频文件。
func (c *Client) DownloadAudio(ctx context.Context, interviewID, turnID string) (io.ReadCloser, error) {
	// TODO: 实现音频下载
	return nil, nil
}

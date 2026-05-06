package s3

import (
	"context"
	"io"
)

// UploadAudio 上传音频文件到 S3。
// 路径: audio/{interview_id}/{turn_id}.wav
func (c *Client) UploadAudio(ctx context.Context, interviewID, turnID string, body io.Reader) error {
	return c.Upload(ctx, AudioObjectKey(interviewID, turnID), body, "audio/wav")
}

// DownloadAudio 下载音频文件。
func (c *Client) DownloadAudio(ctx context.Context, interviewID, turnID string) (io.ReadCloser, error) {
	return c.Download(ctx, AudioObjectKey(interviewID, turnID))
}

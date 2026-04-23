package s3

import (
	"fmt"
	"time"
)

// ResumeObjectKey /resumes/{user_id}/filename
func ResumeObjectKey(userID, filename string) string {
	return fmt.Sprintf("resumes/%s/%s", userID, filename)
}

// AudioObjectKey /audio/{interview_id}/{turn_id}.wav
func AudioObjectKey(interviewID, turnID string) string {
	return fmt.Sprintf("audio/%s/%s.wav", interviewID, turnID)
}

// SFTJSONLPrefix /sft/{date}/
func SFTJSONLPrefix(t time.Time) string {
	return fmt.Sprintf("sft/%s/", t.UTC().Format("2006-01-02"))
}

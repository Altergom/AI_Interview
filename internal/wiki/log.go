package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// appendLog 在 log.md 末尾追加一条 ingest 记录。
func appendLog(wikiDir, rawFile, questionSlug string) error {
	date := time.Now().Format("2006-01-02")
	entry := fmt.Sprintf("%s ingest %s → %s\n", date, rawFile, questionSlug)

	f, err := os.OpenFile(filepath.Join(wikiDir, "log.md"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("write log: %w", err)
	}
	return nil
}

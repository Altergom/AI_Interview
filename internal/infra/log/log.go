package log

import (
	"log/slog"
	"os"
	"strings"
)

// Init 将默认 slog 输出到 stdout：text 便于本地阅读，json 便于容器与日志采集（Loki、ELK、云厂商采集器）。
func Init(level, format, env string) {
	lvl := parseLevel(level)
	opts := &slog.HandlerOptions{Level: lvl}
	var h slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, opts)
	default:
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(h).With("env", env)
	slog.SetDefault(logger)
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

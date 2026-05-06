package log

import (
	"strings"

	hertzzap "github.com/hertz-contrib/logger/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// Init 以 zap 为后端初始化全局 hlog。
// LOG_LEVEL:  debug | info | warn | error
// LOG_FORMAT: text（本地开发）| json（生产/容器）
func Init(level, format, env string) {
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "time"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var enc zapcore.Encoder
	if strings.ToLower(strings.TrimSpace(format)) == "json" {
		enc = zapcore.NewJSONEncoder(encCfg)
	} else {
		encCfg.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString("[" + l.CapitalString() + "]")
		}
		enc = zapcore.NewConsoleEncoder(encCfg)
	}

	logger := hertzzap.NewLogger(
		hertzzap.WithZapOptions(
			zap.WithCaller(true),
			zap.AddCallerSkip(3),
			zap.Fields(zap.String("env", env)),
		),
		hertzzap.WithCoreEnc(enc),
	)

	hlog.SetLogger(logger)
	hlog.SetLevel(parseLevel(level))
}

func Debugf(format string, v ...any) { hlog.Debugf(format, v...) }
func Infof(format string, v ...any)  { hlog.Infof(format, v...) }
func Warnf(format string, v ...any)  { hlog.Warnf(format, v...) }
func Errorf(format string, v ...any) { hlog.Errorf(format, v...) }
func Info(v ...any)                  { hlog.Info(v...) }
func Warn(v ...any)                  { hlog.Warn(v...) }
func Error(v ...any)                 { hlog.Error(v...) }

func parseLevel(s string) hlog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return hlog.LevelDebug
	case "warn", "warning":
		return hlog.LevelWarn
	case "error":
		return hlog.LevelError
	default:
		return hlog.LevelInfo
	}
}

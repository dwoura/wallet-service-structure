package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log *zap.Logger
)

func init() {
	// 默认初始化一个 Nop Logger，防止未 Init 就调用导致 panic
	Log = zap.NewNop()
}

// Init initializes the global logger
func Init(env string) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	Log, err = config.Build(zap.AddCallerSkip(1)) // Skip 1 caller so logs show where logger.Info was called, not wrapper
	if err != nil {
		panic(err)
	}

	// 替换全局标准库 log (这样所有通过 log.Printf 打印的也会被重定向到 Zap)
	zap.ReplaceGlobals(Log)
}

// Sync flushes any buffered log entries
func Sync() {
	_ = Log.Sync()
}

// Helper functions for direct usage
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

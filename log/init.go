package log

import (
	"github.com/vearne/chat/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var DefaultLogger *zap.Logger

func InitLogger(logConfig *config.LogConfig) {
	alevel := zap.NewAtomicLevel()
	hook := lumberjack.Logger{
		Filename:   logConfig.FilePath,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
	}
	w := zapcore.AddSync(&hook)

	switch logConfig.Level {
	case "debug":
		alevel.SetLevel(zap.DebugLevel)
	case "info":
		alevel.SetLevel(zap.InfoLevel)
	case "error":
		alevel.SetLevel(zap.ErrorLevel)
	default:
		alevel.SetLevel(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoderConfig.ConsoleSeparator = " | "

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		alevel,
	)

	DefaultLogger = zap.New(core)
	DefaultLogger = DefaultLogger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
	DefaultLogger.Info("DefaultLogger init success")
}

func Debug(msg string, fields ...zapcore.Field) {
	DefaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	DefaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	DefaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	DefaultLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	DefaultLogger.Fatal(msg, fields...)
}

func Sync() error {
	return DefaultLogger.Sync()
}

package log

import (
	"github.com/vearne/chat/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
)

var DefaultLogger *zap.Logger

func InitLogger() {
	// 动态调整日志级别
	// curl -XPUT --data '{"level":"info"}' http://localhost:19090/handle/level
	alevel := zap.NewAtomicLevel()
	http.HandleFunc("/handle/level", alevel.ServeHTTP)
	go func() {
		if err := http.ListenAndServe(config.GetOpts().Logger.ListenAddress,
			nil); err != nil {
			panic(err)
		}
	}()

	hook := lumberjack.Logger{
		Filename:   config.GetOpts().Logger.FilePath,
		MaxSize:    1024, // megabytes
		MaxBackups: 3,
		MaxAge:     7,    //days
		Compress:   true, // disabled by default
	}
	w := zapcore.AddSync(&hook)

	switch config.GetOpts().Logger.Level {
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
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		alevel,
	)

	DefaultLogger = zap.New(core)
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

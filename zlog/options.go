package zlog

import (
	"io"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DefaultLevel      = zapcore.InfoLevel
	DefaultTimeLayout = time.RFC3339
)

type Level int8

const (
	DebugLevel Level = iota - 1
	// InfoLevel is the default logging priority.
	InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel
)

type Option func(*option)

type option struct {
	level           zapcore.Level
	fields          []zap.Field
	file            io.Writer
	timeLayout      string
	outputInConsole bool
	outputErrorLog  bool
	errorFile       io.Writer
	writers         []LogWriter
	consoleWriter   *consoleWriter
}

// WithLevel 设置日志等级
func WithLevel(level Level) Option {
	return func(o *option) {
		o.level = zapcore.Level(level)
	}
}

// WithFields 添加附加字段
func WithFields(fields ...zap.Field) Option {
	return func(o *option) {
		o.fields = fields
	}
}

// WithFile 设置日志文件
func WithFileRotation(filePath string, maxSize, maxBackups, maxAge int) Option {
	return func(o *option) {
		o.file = &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
		}
	}
}

// WithSeparateErrorFile 单独输出 错误日志
func WithSeparateErrorFile(filePath string, maxSize, maxBackups, maxAge int) Option {
	return func(o *option) {
		o.outputErrorLog = true

		o.errorFile = &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
		}
	}
}

// WithTimeLayout 设置时间格式
func WithTimeLayout(timeLayout string) Option {
	return func(o *option) {
		o.timeLayout = timeLayout
	}
}

func WithOutput(w ...LogWriter) Option {
	return func(o *option) {
		o.writers = w
	}
}

// WithOutputInConsole 输出到控制台
func WithOutputInConsole() Option {
	return func(o *option) {
		o.outputInConsole = true
		o.consoleWriter = newConsoleWriter(o.level)
	}
}

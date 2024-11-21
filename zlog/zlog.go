package zlog

import (
	"os"
	"time"

	"github.com/zR-Zr/goin/interfaces"
	"github.com/zR-Zr/goin/ztools"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	sugger *zap.SugaredLogger
)

var _ interfaces.Logger = (*zlogger)(nil)

type zlogger struct {
	sugger *zap.SugaredLogger
}

func (l *zlogger) Sync() error {
	return l.sugger.Sync()
}

func (l *zlogger) Info(msg string, kAvs ...interface{}) {
	l.sugger.Infow(msg, kAvs...)
}

func (l *zlogger) Debug(msg string, kAvs ...interface{}) {
	l.sugger.Debugw(msg, kAvs...)
}

func (l *zlogger) Warn(msg string, kAvs ...interface{}) {
	l.sugger.Warnw(msg, kAvs...)
}

func (l *zlogger) Error(msg string, kAvs ...interface{}) {
	l.sugger.Errorw(msg, kAvs...)
}

func (l *zlogger) Panic(msg string, kAvs ...interface{}) {
	l.sugger.Panicw(msg, kAvs...)
}

func CreateJsonLogger(opts ...Option) (interfaces.Logger, error) {
	opt := &option{level: zapcore.InfoLevel, fields: make(map[string]string)}
	for _, opFunc := range opts {
		opFunc(opt)
	}

	timeLayout := ztools.GetOrDefault(opt.timeLayout, "2006-01-02 15:04:05")

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewTee()
	if opt.outputInConsole {
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= opt.level && lvl < zapcore.ErrorLevel
		})
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})

		stdout := zapcore.Lock(os.Stdout)
		stderr := zapcore.Lock(os.Stderr)

		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder := zapcore.NewConsoleEncoder(encoderConfig)
		core = zapcore.NewTee(

			zapcore.NewCore(encoder, stdout, lowPriority),
			zapcore.NewCore(encoder, stderr, highPriority),
		)
	}

	if opt.file != nil {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

		if !opt.outputErrorLog {
			core = zapcore.NewTee(
				core,
				zapcore.NewCore(jsonEncoder, zapcore.AddSync(opt.file), zap.LevelEnablerFunc(func(l zapcore.Level) bool {
					return l >= opt.level
				})),
			)
		} else {
			core = zapcore.NewTee(
				core,
				zapcore.NewCore(jsonEncoder,
					zapcore.AddSync(opt.file),
					zap.LevelEnablerFunc(func(l zapcore.Level) bool {
						return l >= opt.level && l < zapcore.ErrorLevel
					}),
				),
				zapcore.NewCore(jsonEncoder,
					zapcore.AddSync(opt.errorFile),
					zap.LevelEnablerFunc(func(l zapcore.Level) bool {
						return l >= zapcore.ErrorLevel
					}),
				),
			)
		}

	}

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	for k, v := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: k, Type: zapcore.StringType, String: v}))
	}

	return &zlogger{
		sugger: logger.Sugar(),
	}, nil
}

package zlog

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/zR-Zr/goin/interfaces"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *zap.Logger
	sugger *zap.SugaredLogger
)

var _ interfaces.Logger = (*zlogger)(nil)

type zlogger struct {
	sugger *zap.SugaredLogger
	lock   sync.RWMutex
}

// CreateLogger 创建Logger 实例
func CreateLogger(opts ...Option) (interfaces.Logger, error) {
	opt := &option{level: DefaultLevel, timeLayout: DefaultTimeLayout} // 使用默认值
	for _, opFunc := range opts {
		opFunc(opt)
	}

	// timeLayout := ztools.GetOrDefault(opt.timeLayout, "2006-01-02 15:04:05")

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
			enc.AppendString(t.Format(opt.timeLayout))
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewTee()

	if opt.outputInConsole {
		stdout := zapcore.Lock(os.Stdout)
		stderr := zapcore.Lock(os.Stderr)

		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 设置控制台输出的日志级别颜色

		encoder := zapcore.NewConsoleEncoder(encoderConfig)

		// 将低级别日志输出到 stdout, 高级别日志输出到 stderr
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, stdout, opt.consoleWriter.logPriority),
			zapcore.NewCore(encoder, stderr, opt.consoleWriter.heighPriority),
		)
	}

	if opt.file != nil {
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		jsonEncoder := zapcore.NewJSONEncoder(encoderConfig) // 创建 json 编码器

		// 如果没有配置单独的错误日志文件, 则所有级别的日志都输出到同一个文件
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

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.Fields(opt.fields...)) // 创建 zap.logger 实例

	// for k, v := range opt.fields {
	// 	logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: k, Type: zapcore.StringType, String: v}))
	// }

	return &zlogger{
		sugger: logger.Sugar(),
	}, nil
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

// LogWriter 日志写入口
type LogWriter interface {
	Write(p []byte) (n int, err error)
	Sync() error
}

// consoleWriter 控制台写入器
type consoleWriter struct {
	logPriority   zap.LevelEnablerFunc
	heighPriority zap.LevelEnablerFunc
}

func newConsoleWriter(level zapcore.Level) *consoleWriter {
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= level && lvl < zapcore.ErrorLevel
	})

	heighPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	return &consoleWriter{
		logPriority:   lowPriority,
		heighPriority: heighPriority,
	}
}

// fileWriter 文件写入器
type fileWriter struct {
	file *os.File   // 使用 os.File 指针
	mu   sync.Mutex // 添加互斥锁,保证并发安全.
}

// 创建文件写入器, 在文件写入器中, lumberjack 不再负责缓冲区和写入文件,只负责文件的切割.
func NewFileWriter(filePath string, maxSize, maxBackups, maxAge int) (*fileWriter, error) {
	// 1. 创建 lumberjack.Logger 实例,用于日志文件轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
	}

	// 2. 打开日志文件
	// 使用 os.OpenFile 打开文件
	// os.O_WRONLY: 只写模式
	// os.O_CREATE: 如果文件不存在,创建文件
	// os.O_APPEND: 追加模式写入
	// 0644: 文件权限 (rw-r--r--)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// 3. 初始化日志文件
	// 这段代码的目的是确保 lumberjack.Logger 可以正常工作.因为 lumberjack.Logger 在第一次写入时才会创建文件
	// 所以我们先写入一个空字节, 出发文件的创建和 lumberjack 的初始化
	if _, err := file.Write([]byte{}); err != nil {
		return nil, fmt.Errorf("failed to initialize log file: %w", err)
	}

	// 手动触发日志轮转, 确保日志文件按配置进行轮转
	// 这通常在应用启动时执行一次即可.
	if err := lumberjackLogger.Rotate(); err != nil {
		return nil, fmt.Errorf("failed to rotate log file: %w", err)
	}

	// 4. 返回 filterWriter 实例
	return &fileWriter{file: file}, nil
}

// networkWriter 网络写入器
type networkWriter struct {
	conn net.Conn
}

func NewNetworkWriter(address string) (*networkWriter, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote log server: %w", err)
	}
	return &networkWriter{conn: conn}, nil
}

// 为不同的类型写入器实现 LogWriter 接口
func (w *consoleWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (w *consoleWriter) Sync() error {
	return nil
}

// 写入日志
func (w *fileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock() // 加锁,保证并发安全
	defer w.mu.Unlock()
	return w.file.Write(p)
}

func (w *fileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Sync() // 手动刷新缓冲区
}

func (w *networkWriter) Write(p []byte) (n int, err error) {
	return w.conn.Write(p)
}

func (w *networkWriter) Sync() error {
	return nil
}

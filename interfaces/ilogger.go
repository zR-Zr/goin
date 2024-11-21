package interfaces

type Logger interface {
	Sync() error
	Info(msg string, kAvs ...interface{})
	Debug(msg string, kAvs ...interface{})
	Warn(msg string, kAvs ...interface{})
	Error(msg string, kAvs ...interface{})
	Panic(msg string, kAvs ...interface{})
}

package logger

import (
	"fmt"

	"github.com/zR-Zr/goin/interfaces"
	"github.com/zR-Zr/goin/pkg/config"
	"github.com/zR-Zr/goin/zlog"
)

func InitLogger(cfg *config.Config) (logger interfaces.Logger, err error) {
	var opts []zlog.Option

	// 设置日志级别
	switch cfg.Logger.Level {
	case "debug":
		opts = append(opts, zlog.WithLevel(zlog.DebugLevel))
	case "info":
		opts = append(opts, zlog.WithLevel(zlog.InfoLevel))
	case "warn":
		opts = append(opts, zlog.WithLevel(zlog.WarnLevel))
	case "error":
		opts = append(opts, zlog.WithLevel(zlog.ErrorLevel))
	case "dpanic":
		opts = append(opts, zlog.WithLevel(zlog.DPanicLevel))
	case "panic":
		opts = append(opts, zlog.WithLevel(zlog.PanicLevel))
	case "fatal":
	}

	// 设置日志输出
	for _, output := range cfg.Logger.Output {
		switch output {
		case "console":
			opts = append(opts, zlog.WithOutputInConsole())
		case "file":
			writer, err := zlog.NewFileWriter(cfg.Logger.FielPath, cfg.Logger.MaxSize, cfg.Logger.MaxBackups, cfg.Logger.MaxAge)
			if err != nil {
				return nil, fmt.Errorf("logger: failed to create log file: %w", err)
			}

			opts = append(opts, zlog.WithOutput(writer))
			if cfg.Logger.ErrorFilePath != "" {
				errorWriter, err := zlog.NewFileWriter(cfg.Logger.ErrorFilePath, cfg.Logger.MaxSize, cfg.Logger.MaxBackups, cfg.Logger.MaxAge)
				if err != nil {
					return nil, fmt.Errorf("logger: failed to create error log file: %w", err)
				}

				opts = append(opts, zlog.WithOutput(errorWriter))
			}
		case "network":
			writer, err := zlog.NewNetworkWriter(cfg.Logger.RemoteAddr)
			if err != nil {
				return nil, fmt.Errorf("logger: failed to connect to remote log server: %w", err)
			}

			opts = append(opts, zlog.WithOutput(writer))
		}
	}

	logger, err = zlog.CreateLogger(opts...)
	if err != nil {
		return nil, fmt.Errorf("logger: failed to create logger: %w", err)
	}

	return logger, nil
}

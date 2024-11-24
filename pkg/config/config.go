package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 框架配置结构体
type Config struct {
	AppName string `mapstructure:"app_name`
	HTTP    struct {
		Addr         string        `mapstructure:"addr"`
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`
		WriteTimeout time.Duration `mapstructure:"write_timeout"`
	} `mapstructure:"http"`

	Database struct {
		Host            string        `mapstructure:"host"`
		Port            int           `mapstructure:"port"`
		User            string        `mapstructure:"user"`
		Password        string        `mapstructure:"password"`
		DBName          string        `mapstructure:"dbname"`
		Charset         string        `mapstructure:"charset"`
		ParseTime       bool          `mapstructure:"parseTime"`
		Loc             string        `mapstructure:"loc"`
		ReadHosts       []string      `mapstructure:"read_hosts"`
		WriteHost       string        `mapstructure:"write_host"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	} `mapstructure:"database"`

	Logger struct {
		Level         string   `mapstructure:"level"`
		Output        []string `mapstructure:"output"` // 支持多个输出目标
		FielPath      string   `mapstructure:"filepath"`
		ErrorFilePath string   `mapstructure:"error_filepath"` // 单独的错误日志
		MaxSize       int      `mapstructure:"max_size"`
		MaxBackups    int      `mapstructure:"max_backups"`
		MaxAge        int      `mapstructure:"max_age"`
		RemoteAddr    string   `mapstructure:"remote_addr"` // 远程日志服务器地址
	} `mapstructure:"logger"`

	Cache struct {
		Type     string `mapstructure:"type"`
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"cache"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认配置
	viper.SetDefault("app_name", "goin-app") // 默认应用名
	viper.SetDefault("http.addr", ":8080")
	viper.SetDefault("http.read_timeout", "60s")
	viper.SetDefault("http.write_timeout", "60s")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.dbname", "mydatabase")
	viper.SetDefault("database.charset", "utf8mb4")
	viper.SetDefault("database.parseTime", true)
	viper.SetDefault("database.loc", "Local")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", "30m")
	viper.SetDefault("logger.level", "debug")
	viper.SetDefault("logger.output", "console")
	viper.SetDefault("cache.type", "none") // 默认不使用缓存

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 获取 Appname 配置，并将其转换为环境变量前缀
	appName := viper.GetString("app_name")
	envPrefix := strings.ToUpper(strings.ReplaceAll(appName, "-", "_"))

	// 绑定环境变量,允许环境变量覆盖配置
	viper.SetEnvPrefix(envPrefix)
	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

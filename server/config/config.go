package config

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

var Config configModel
var Ip string

type LogConfig struct {
	Filename   string `yaml:"filename"`   // log file path, default: ./log/app.log
	MaxSize    int    `yaml:"maxSize"`    // MB per file, default: 20
	MaxBackups int    `yaml:"maxBackups"` // number of backups, default: 3
	MaxAge     int    `yaml:"maxAge"`     // days to retain, default: 1
	Compress   bool   `yaml:"compress"`   // gzip old files, default: true
}

type DatabaseConfig struct {
	Path            string `yaml:"path"`            // SQLite file path, default: ./data.db
	MaxOpenConns    int    `yaml:"maxOpenConns"`    // default: 10
	MaxIdleConns    int    `yaml:"maxIdleConns"`    // default: 5
	ConnMaxLifetime int    `yaml:"connMaxLifetime"` // seconds, default: 300
}

type PipelineConfig struct {
	BatchSize   int `yaml:"batchSize"`   // default: 1000
	ChannelSize int `yaml:"channelSize"` // default: 10000
}

type configModel struct {
	Username    string         `yaml:"username"`
	Password    string         `yaml:"password"`
	JwtSecret   string         `yaml:"jwtSecret"`
	ApiSecret   string         `yaml:"apiSecret"`
	AesKey      string         `yaml:"aesKey"`
	InitDb      bool           `yaml:"initDb"`
	LogLevel    string         `yaml:"logLevel"` // dev or prod
	Log         LogConfig      `yaml:"log"`
	Database    DatabaseConfig `yaml:"database"`
	Pipeline    PipelineConfig `yaml:"pipeline"`
	ServerUrl   string         `yaml:"serverUrl"`
	RunWeb      bool           `yaml:"runWeb"`
	WebUrl      string         `yaml:"webUrl"`
	CorsOrigins []string       `yaml:"corsOrigins"` // Allowed CORS origins, e.g. ["http://localhost:8081"]
}

func init() {
	if err := LoadConfig(); err != nil {
		// During tests or when config.yaml is not available, use defaults silently.
		// The main function should call LoadConfig() explicitly and handle the error.
		fmt.Printf("Warning: %v (using defaults)\n", err)
	}
}

// LoadConfig reads config.yaml and populates the global Config.
// Environment variables ETL_USERNAME, ETL_PASSWORD, ETL_JWT_SECRET, ETL_API_SECRET, ETL_AES_KEY
// take precedence over values in the config file.
func LoadConfig() error {
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		return fmt.Errorf("Failed To Read Config File: %w", err)
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("Failed To Parse Config File: %w", err)
	}

	// Environment variable overrides
	if v := os.Getenv("ETL_USERNAME"); v != "" {
		Config.Username = v
	}
	if v := os.Getenv("ETL_PASSWORD"); v != "" {
		Config.Password = v
	}
	if v := os.Getenv("ETL_JWT_SECRET"); v != "" {
		Config.JwtSecret = v
	}
	if v := os.Getenv("ETL_API_SECRET"); v != "" {
		Config.ApiSecret = v
	}
	if v := os.Getenv("ETL_AES_KEY"); v != "" {
		Config.AesKey = v
	}
	if v := os.Getenv("ETL_SERVER_URL"); v != "" {
		Config.ServerUrl = v
	}
	if v := os.Getenv("ETL_LOG_LEVEL"); v != "" {
		Config.LogLevel = v
	}
	if Config.Pipeline.BatchSize <= 0 {
		Config.Pipeline.BatchSize = 1000
	}
	if Config.Pipeline.ChannelSize <= 0 {
		Config.Pipeline.ChannelSize = 10000
	}

	Ip = GetLocalIP()
	return nil
}

func SaveConfig() error {
	data, err := yaml.Marshal(Config)
	if err != nil {
		return err
	}
	return os.WriteFile("./config.yaml", data, 0644)
}

// GetLocalIP 获取本地IP地址
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

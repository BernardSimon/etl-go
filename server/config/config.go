package config

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

var Config configModel
var Ip string

const configFileName = "config.yaml"

type LogConfig struct {
	Filename   string `yaml:"filename"`   // log file path, default: ./log/app.log
	MaxSize    int    `yaml:"maxSize"`    // MB per file, default: 20
	MaxBackups int    `yaml:"maxBackups"` // number of backups, default: 3
	MaxAge     int    `yaml:"maxAge"`     // days to retain, default: 1
	Compress   bool   `yaml:"compress"`   // gzip old files, default: true
}

type DatabaseConfig struct {
	Driver          string `yaml:"driver"`          // sqlite (default) | mysql | postgres
	Path            string `yaml:"path"`            // SQLite file path, default: ./data.db (only used when driver=sqlite)
	DSN             string `yaml:"dsn"`             // MySQL/PostgreSQL DSN, e.g. "user:pass@tcp(host:3306)/db?parseTime=True"
	MaxOpenConns    int    `yaml:"maxOpenConns"`    // default: 10
	MaxIdleConns    int    `yaml:"maxIdleConns"`    // default: 5
	ConnMaxLifetime int    `yaml:"connMaxLifetime"` // seconds, default: 300
}

type PipelineConfig struct {
	BatchSize   int `yaml:"batchSize"`   // default: 1000
	ChannelSize int `yaml:"channelSize"` // default: 10000
}

type DatasourcePoolConfig struct {
	MaxOpenConns    int `yaml:"maxOpenConns"`    // default: 5
	MaxIdleConns    int `yaml:"maxIdleConns"`    // default: 2
	ConnMaxLifetime int `yaml:"connMaxLifetime"` // seconds, default: 300
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
	DatasourcePool DatasourcePoolConfig `yaml:"datasourcePool"`
	ServerUrl   string         `yaml:"serverUrl"`
	RunWeb      bool           `yaml:"runWeb"`
	WebUrl      string         `yaml:"webUrl"`
	CorsOrigins []string       `yaml:"corsOrigins"` // Allowed CORS origins, e.g. ["http://localhost:8081"]
	TotpEnabled bool           `yaml:"totpEnabled"` // Enable TOTP two-factor authentication
	TotpSecret  string         `yaml:"totpSecret"`  // Base32-encoded TOTP secret (configure in authenticator app)
}

func init() {
	if err := LoadConfig(); err != nil {
		// During tests or when config.yaml is not available, use defaults silently.
		// The main function should call LoadConfig() explicitly and handle the error.
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Printf("Warning: %v (using defaults)\n", err)
		}
	}
}

func defaultConfig() configModel {
	return configModel{
		Username:  "admin",
		Password:  "password123",
		JwtSecret: mustRandomHex(32),
		ApiSecret: mustRandomHex(32),
		AesKey:    mustRandomHex(32),
		InitDb:    true,
		LogLevel:  "dev",
		Log: LogConfig{
			Filename:   "./log/app.log",
			MaxSize:    20,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		Database: DatabaseConfig{
			Driver:          "sqlite",
			Path:            "./data.db",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		},
		Pipeline: PipelineConfig{
			BatchSize:   1000,
			ChannelSize: 10000,
		},
		DatasourcePool: DatasourcePoolConfig{
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 300,
		},
		ServerUrl: "0.0.0.0:8080",
		RunWeb:    true,
		WebUrl:    "0.0.0.0:8081",
		CorsOrigins: []string{
			"http://localhost:8081",
			"http://localhost:5173",
		},
		TotpEnabled: false,
		TotpSecret:  mustRandomTotpSecret(),
	}
}

func mustRandomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to generate random bytes: %w", err))
	}
	return hex.EncodeToString(buf)
}

// mustRandomTotpSecret generates a random base32-encoded TOTP secret (20 bytes / 160 bits).
func mustRandomTotpSecret() string {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Errorf("failed to generate random TOTP secret: %w", err))
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buf)
}

func configPath() string {
	execPath := executableConfigPath()
	cwdPath := workingDirectoryConfigPath()

	switch {
	case fileExists(execPath):
		return execPath
	case fileExists(cwdPath):
		return cwdPath
	case execPath != "":
		return execPath
	default:
		return cwdPath
	}
}

func executableConfigPath() string {
	execFile, err := os.Executable()
	if err != nil || execFile == "" {
		return ""
	}
	resolvedExecFile, err := filepath.EvalSymlinks(execFile)
	if err == nil && resolvedExecFile != "" {
		execFile = resolvedExecFile
	}
	return filepath.Join(filepath.Dir(execFile), configFileName)
}

func workingDirectoryConfigPath() string {
	return filepath.Join(".", configFileName)
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func applyEnvOverrides() {
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
}

func applyRuntimeDefaults() {
	if Config.Username == "" {
		Config.Username = "admin"
	}
	if Config.Password == "" {
		Config.Password = "password123"
	}
	if Config.JwtSecret == "" {
		Config.JwtSecret = mustRandomHex(32)
	}
	if Config.AesKey == "" {
		Config.AesKey = mustRandomHex(32)
	}
	if Config.LogLevel == "" {
		Config.LogLevel = "dev"
	}
	if Config.Log.Filename == "" {
		Config.Log.Filename = "./log/app.log"
	}
	if Config.Log.MaxSize <= 0 {
		Config.Log.MaxSize = 20
	}
	if Config.Log.MaxBackups <= 0 {
		Config.Log.MaxBackups = 3
	}
	if Config.Log.MaxAge <= 0 {
		Config.Log.MaxAge = 7
	}
	if Config.Database.Driver == "" {
		Config.Database.Driver = "sqlite"
	}
	if Config.Database.Driver == "sqlite" && Config.Database.Path == "" {
		Config.Database.Path = "./data.db"
	}
	if Config.Database.MaxOpenConns <= 0 {
		Config.Database.MaxOpenConns = 10
	}
	if Config.Database.MaxIdleConns <= 0 {
		Config.Database.MaxIdleConns = 5
	}
	if Config.Database.ConnMaxLifetime <= 0 {
		Config.Database.ConnMaxLifetime = 300
	}
	if Config.Pipeline.BatchSize <= 0 {
		Config.Pipeline.BatchSize = 1000
	}
	if Config.Pipeline.ChannelSize <= 0 {
		Config.Pipeline.ChannelSize = 10000
	}
	if Config.DatasourcePool.MaxOpenConns <= 0 {
		Config.DatasourcePool.MaxOpenConns = 5
	}
	if Config.DatasourcePool.MaxIdleConns <= 0 {
		Config.DatasourcePool.MaxIdleConns = 2
	}
	if Config.DatasourcePool.ConnMaxLifetime <= 0 {
		Config.DatasourcePool.ConnMaxLifetime = 300
	}
	if Config.ServerUrl == "" {
		Config.ServerUrl = "0.0.0.0:8080"
	}
	if Config.WebUrl == "" {
		Config.WebUrl = "0.0.0.0:8081"
	}
	if len(Config.CorsOrigins) == 0 {
		Config.CorsOrigins = []string{
			"http://localhost:8081",
			"http://localhost:5173",
		}
	}
}

func EnsureConfig() (bool, error) {
	path := configPath()
	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return false, fmt.Errorf("failed to stat config file: %w", err)
		}
		Config = defaultConfig()
		applyEnvOverrides()
		Ip = GetLocalIP()
		if err := SaveConfig(); err != nil {
			return false, fmt.Errorf("failed to initialize config file: %w", err)
		}
		return true, nil
	}

	if err := LoadConfig(); err != nil {
		return false, err
	}

	// Repair critical missing fields for older or partial configs.
	original := Config
	applyRuntimeDefaults()
	if !reflect.DeepEqual(original, Config) {
		if err := SaveConfig(); err != nil {
			return false, fmt.Errorf("failed to persist normalized config: %w", err)
		}
	}
	return false, nil
}

func EnsureRuntimePaths() error {
	paths := []string{Config.Log.Filename}
	// 只有 sqlite 需要确保数据库文件所在目录存在
	if Config.Database.Driver == "" || Config.Database.Driver == "sqlite" {
		paths = append(paths, Config.Database.Path)
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		dir := filepath.Dir(path)
		if dir == "." || dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create runtime directory for %s: %w", path, err)
		}
	}
	return nil
}

// LoadConfig reads config.yaml and populates the global Config.
// Environment variables ETL_USERNAME, ETL_PASSWORD, ETL_JWT_SECRET, ETL_API_SECRET, ETL_AES_KEY
// take precedence over values in the config file.
func LoadConfig() error {
	data, err := os.ReadFile(configPath())
	if err != nil {
		return fmt.Errorf("Failed To Read Config File: %w", err)
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		return fmt.Errorf("Failed To Parse Config File: %w", err)
	}
	applyEnvOverrides()
	applyRuntimeDefaults()

	Ip = GetLocalIP()
	return nil
}

func SaveConfig() error {
	if err := os.MkdirAll(filepath.Dir(configPath()), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(Config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}

func ApplyDatasourcePoolConfig(target map[string]string) {
	if target == nil {
		return
	}
	target["__pool_max_open_conns"] = fmt.Sprintf("%d", Config.DatasourcePool.MaxOpenConns)
	target["__pool_max_idle_conns"] = fmt.Sprintf("%d", Config.DatasourcePool.MaxIdleConns)
	target["__pool_conn_max_lifetime"] = fmt.Sprintf("%d", Config.DatasourcePool.ConnMaxLifetime)
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

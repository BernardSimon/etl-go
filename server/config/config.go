package config

import (
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

var Config configModel
var Ip string

type configModel struct {
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	JwtSecret   string   `yaml:"jwtSecret"`
	AesKey      string   `yaml:"aesKey"`
	InitDb      bool     `yaml:"initDb"`
	LogLevel    string   `yaml:"logLevel"` // dev or prod
	ServerUrl   string   `yaml:"serverUrl"`
	RunWeb      bool     `yaml:"runWeb"`
	WebUrl      string   `yaml:"webUrl"`
	CorsOrigins []string `yaml:"corsOrigins"` // Allowed CORS origins, e.g. ["http://localhost:8081"]
}

func init() {
	if err := LoadConfig(); err != nil {
		// During tests or when config.yaml is not available, use defaults silently.
		// The main function should call LoadConfig() explicitly and handle the error.
		fmt.Printf("Warning: %v (using defaults)\n", err)
	}
}

// LoadConfig reads config.yaml and populates the global Config.
// Environment variables ETL_USERNAME, ETL_PASSWORD, ETL_JWT_SECRET, ETL_AES_KEY
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
	if v := os.Getenv("ETL_AES_KEY"); v != "" {
		Config.AesKey = v
	}
	if v := os.Getenv("ETL_SERVER_URL"); v != "" {
		Config.ServerUrl = v
	}
	if v := os.Getenv("ETL_LOG_LEVEL"); v != "" {
		Config.LogLevel = v
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

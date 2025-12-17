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
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	JwtSecret string `yaml:"jwtSecret"`
	AesKey    string `yaml:"aesKey"`
	Language  string `yaml:"language"`
	InitDb    bool   `yaml:"initDb"`
	LogLevel  string `yaml:"logLevel"` // dev or prod
	ServerUrl string `yaml:"serverUrl"`
	RunWeb    bool   `yaml:"runWeb"`
	WebUrl    string `yaml:"webUrl"`
}

func init() {
	data, err := os.ReadFile("./config.yaml")
	if err != nil {
		fmt.Printf("Failed To Read Config File: %v", err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		fmt.Printf("Failed To Parse Config File: %v", err)
		os.Exit(1)
	}
	Ip = GetLocalIP()
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

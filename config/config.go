package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Token stores the oauth tokens
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int64  `json:"created_at"`
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	if t == nil {
		return true
	}
	return time.Now().Unix() > t.CreatedAt+int64(t.ExpiresIn)
}

// Config 存储应用程序的配置
type Config struct {
	Token *Token `json:"token"`
}

// 配置文件路径变量（导出供测试使用）
var (
	DefaultConfigFile = "~/.config/trakt/config.json"
)

// Save 保存配置到默认文件
func (c *Config) Save() error {
	// 处理~符号
	configFile := os.ExpandEnv(DefaultConfigFile)

	// 创建目录
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 序列化配置
	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(configFile, data, 0644)
}

// LoadConfig 从默认文件加载配置
func LoadConfig() (*Config, error) {
	// 处理~符号
	configFile := os.ExpandEnv(DefaultConfigFile)

	// 读取文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// 解析配置
	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
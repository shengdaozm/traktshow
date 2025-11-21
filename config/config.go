package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// 配置文件路径（用户主目录下的隐藏文件）
const configFileName = ".trakt-client-config.json"

// Config Trakt客户端配置结构体
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
}

// 全局配置实例
var globalConfig *Config

// Init 初始化配置：优先加载已保存的配置，无则引导用户输入并保存
func Init() error {
	// 尝试加载已存在的配置
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("读取配置文件失败：%v", err)
		}
		if err := json.Unmarshal(data, &globalConfig); err != nil {
			return fmt.Errorf("解析配置文件失败：%v", err)
		}
		log.Println("配置加载成功（来自：", configPath, "）")
		return nil
	}

	// 无配置文件，引导用户输入
	log.Println("未找到配置文件，开始初始化配置...")
	var clientID, clientSecret string

	fmt.Print("请输入你的 Trakt Client ID：")
	if _, err := fmt.Scanln(&clientID); err != nil {
		return fmt.Errorf("输入Client ID失败：%v", err)
	}

	fmt.Print("请输入你的 Trakt Client Secret：")
	if _, err := fmt.Scanln(&clientSecret); err != nil {
		return fmt.Errorf("输入Client Secret失败：%v", err)
	}

	// 生成默认配置
	globalConfig = &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  "http://localhost:8081/callback", // 固定本地回调地址（无需用户输入）
	}

	// 保存配置到文件
	if err := saveConfig(); err != nil {
		return fmt.Errorf("保存配置失败：%v", err)
	}
	log.Println("配置初始化完成，已保存到：", configPath)
	return nil
}

// Get 获取全局配置（需先调用Init）
func Get() *Config {
	if globalConfig == nil {
		log.Panic("配置未初始化，请先调用config.Init()")
	}
	return globalConfig
}

// 获取配置文件路径（跨平台兼容：Windows/Linux/Mac）
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("获取用户主目录失败：%v", err)
	}
	return filepath.Join(homeDir, configFileName)
}

// 保存配置到文件
func saveConfig() error {
	data, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0600) // 0600：仅当前用户可读写（安全）
}
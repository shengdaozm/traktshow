package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigCRUD(t *testing.T) {
	// 创建临时目录替代真实路径
	tempDir := t.TempDir()
	testConfigPath := filepath.Join(tempDir, "config.json")

	// 保存原始默认配置路径
	origDefaultConfig := DefaultConfigFile
	// 修改默认配置路径为临时目录
	DefaultConfigFile = testConfigPath

	// 测试数据
	testAPIKey := "test_api_key"

	// 1. 创建并保存配置
	config := Config{APIKey: testAPIKey}
	err := config.Save()
	if err != nil {
		t.Errorf("保存配置失败: %v", err)
	}

	// 2. 读取配置
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Errorf("加载配置失败: %v", err)
	}

	// 3. 验证数据一致性
	if loadedConfig.APIKey != testAPIKey {
		t.Errorf("期望APIKey %s, 得到 %s", testAPIKey, loadedConfig.APIKey)
	}

	// 4. 验证文件实际位置
	if _, err := os.Stat(testConfigPath); os.IsNotExist(err) {
		t.Errorf("期望配置文件存在于 %s", testConfigPath)
	}

	// 恢复原始配置路径
	DefaultConfigFile = origDefaultConfig
}
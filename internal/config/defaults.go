package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDBDir 确保 DB 文件所在目录存在。
func EnsureDBDir(path string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建数据库目录: %w", err)
		}
	}
	return nil
}

// SmokeCheck 冒烟测试：验证配置基本可用。
func (c *AppConfig) SmokeCheck() error {
	if c.ServerPort == "" {
		return fmt.Errorf("端口未配置")
	}
	if c.DBPath == "" {
		return fmt.Errorf("数据库路径未配置")
	}
	if c.MaxTreatmentCycleSize <= 0 {
		return fmt.Errorf("批量上限非法")
	}
	return nil
}
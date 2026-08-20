package config

import (
	"os"
	"strconv"
	"time"
)

// Load 从环境变量加载配置，未设置则用默认值。
func Load() *AppConfig {
	c := New()
	if v := os.Getenv("CR_PORT"); v != "" {
		c.ServerPort = v
	}
	if v := os.Getenv("CR_DB_PATH"); v != "" {
		c.DBPath = v
	}
	if v := os.Getenv("CR_WEB_DIR"); v != "" {
		c.WebDir = v
	}
	if v := os.Getenv("CR_COLLECTOR_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.CollectorInterval = d
		}
	}
	if v := os.Getenv("CR_OFFLINE_AFTER"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.OfflineAfter = d
		}
	}
	if v := os.Getenv("CR_EVAL_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.EvaluatorWorkers = n
		}
	}
	if v := os.Getenv("CR_MAX_BATCH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.MaxTreatmentCycleSize = n
		}
	}
	if v := os.Getenv("CR_SIMULATOR"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.EnableSimulator = b
		}
	}
	if v := os.Getenv("CR_SIM_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.SimulatorInterval = d
		}
	}
	if v := os.Getenv("CR_SNAPSHOT_REFRESH"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.SnapshotRefresh = d
		}
	}
	return c
}
package config

import "time"

// AppConfig 应用配置。
type AppConfig struct {
	ServerPort     string
	DBPath         string
	WebDir         string
	CollectorInterval time.Duration
	OfflineAfter   time.Duration
	EvaluatorWorkers int
	MaxTreatmentCycleSize   int
	EnableSimulator bool
	SimulatorInterval time.Duration
	SnapshotRefresh time.Duration
}

// New 返回默认配置。
func New() *AppConfig {
	return &AppConfig{
		ServerPort:        "8080",
		DBPath:            "vessel.db",
		WebDir:            "web",
		CollectorInterval: 5 * time.Second,
		OfflineAfter:      90 * time.Second,
		EvaluatorWorkers:  4,
		MaxTreatmentCycleSize:      500,
		EnableSimulator:   true,
		SimulatorInterval: 3 * time.Second,
		SnapshotRefresh:   5 * time.Second,
	}
}
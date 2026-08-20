package model

import (
	"errors"
	"strings"
	"time"
)

// ComplianceAlert 等级。
const (
	ComplianceAlertWarn  = "warn"
	ComplianceAlertAlarm = "alarm"
)

// ComplianceAlert 状态。
const (
	ComplianceAlertOpen       = "open"
	ComplianceAlertAcknowledged = "acknowledged"
	ComplianceAlertResolved   = "resolved"
)

// ComplianceAlert 告警记录。
type ComplianceAlert struct {
	ID          int64     `json:"id"`
	RuleID      int64     `json:"rule_id"`
	BallastTankID      int64     `json:"tank_id"`
	SamplingPointID     int64     `json:"sampling_sampling_point_id"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	OpenedAt    time.Time `json:"opened_at"`
	AckAt       *time.Time `json:"ack_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
}

// ComplianceComplianceAlertInput 查询告警列表条件。
type ComplianceComplianceAlertInput struct {
	BallastTankID   int64     `json:"tank_id"`
	Level    string    `json:"level"`
	Status   string    `json:"status"`
	From     time.Time `json:"from"`
	To       time.Time `json:"to"`
	Limit    int       `json:"limit"`
}

// ComplianceAlertLevels 合法告警等级。
var ComplianceAlertLevels = map[string]bool{ComplianceAlertWarn: true, ComplianceAlertAlarm: true}

// ComplianceAlertStatuses 合法告警状态。
var ComplianceAlertStatuses = map[string]bool{ComplianceAlertOpen: true, ComplianceAlertAcknowledged: true, ComplianceAlertResolved: true}

func (a *ComplianceAlert) Validate() error {
	if !ComplianceAlertLevels[a.Level] {
		return ErrInvalidLevel
	}
	if !ComplianceAlertStatuses[a.Status] {
		return ErrInvalidComplianceAlertStatus
	}
	return nil
}

// IsOpen 是否未解决。
func (a *ComplianceAlert) IsOpen() bool {
	return a.Status == ComplianceAlertOpen || a.Status == ComplianceAlertAcknowledged
}

var (
	ErrInvalidLevel       = errors.New("告警等级非法")
	ErrInvalidComplianceAlertStatus = errors.New("告警状态非法")
)

// Message 构造告警描述。
func (a *ComplianceAlert) MessageText() string {
	parts := []string{"房间", tankName(a.BallastTankID)}
	if a.Level == ComplianceAlertAlarm {
		parts = append(parts, "ALARM")
	} else {
		parts = append(parts, "WARN")
	}
	return strings.Join(parts, " ")
}

func tankName(_ int64) string { return "" }
package model

import (
	"errors"
	"time"
)

// ComplianceRule 操作符。
const (
	OpGt = "gt"
	OpLt = "lt"
	OpGte = "gte"
	OpLte = "lte"
)

// ComplianceRule 告警规则：某房间某参数超过阈值并持续 duration 秒后触发。
type ComplianceRule struct {
	ID         int64     `json:"id"`
	BallastTankID     int64     `json:"tank_id"`
	Code       string    `json:"code"`
	ParamType  string    `json:"param_type"`
	Op         string    `json:"op"`
	Threshold  float64   `json:"threshold"`
	DurationSec int      `json:"duration_sec"`
	Level      string    `json:"level"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

// ComplianceRuleInput 创建规则入参。
type ComplianceRuleInput struct {
	BallastTankID      int64   `json:"tank_id"`
	Code        string  `json:"code"`
	ParamType   string  `json:"param_type"`
	Op          string  `json:"op"`
	Threshold   float64 `json:"threshold"`
	DurationSec int     `json:"duration_sec"`
	Level       string  `json:"level"`
	Enabled     bool    `json:"enabled"`
}

// ComplianceAlertOps 合法操作符集合。
var ComplianceAlertOps = map[string]bool{OpGt: true, OpLt: true, OpGte: true, OpLte: true}

func (r *ComplianceRule) Validate() error {
	if r.BallastTankID <= 0 {
		return ErrBallastTankRequired
	}
	if r.Code == "" {
		return ErrCodeRequired
	}
	if !ParamTypes[r.ParamType] {
		return ErrInvalidParamType
	}
	if !ComplianceAlertOps[r.Op] {
		return ErrInvalidOp
	}
	if r.DurationSec <= 0 {
		return ErrInvalidDuration
	}
	if !ComplianceAlertLevels[r.Level] {
		return ErrInvalidLevel
	}
	return nil
}

// Match 判断读数是否命中规则。
func (r *ComplianceRule) Match(value float64) bool {
	switch r.Op {
	case OpGt:
		return value > r.Threshold
	case OpGte:
		return value >= r.Threshold
	case OpLt:
		return value < r.Threshold
	case OpLte:
		return value <= r.Threshold
	}
	return false
}

var ErrInvalidOp = errors.New("操作符非法")
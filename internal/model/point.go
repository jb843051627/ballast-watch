package model

import (
	"errors"
	"time"
)

// ParamType 环境参数类型。
const (
	ParamTemp        = "temp"        // 温度 ℃
	ParamHumidity    = "humidity"    // 相对湿度 %
	ParamPressure    = "pressure"    // 压差 Pa
	ParamParticle05  = "particle_05" // ≥0.5μm 粒子计数 个/m³
	ParamParticle50  = "particle_50" // ≥5.0μm 粒子计数 个/m³
)

// SamplingPoint 监测点：定义房间某参数的采集点位与阈值。
type SamplingPoint struct {
	ID               int64     `json:"id"`
	BallastTankID           int64     `json:"tank_id"`
	Code             string    `json:"code"`
	ParamType        string    `json:"param_type"`
	ThresholdMin     float64   `json:"threshold_min"`
	ThresholdMax     float64   `json:"threshold_max"`
	AlarmDurationSec int       `json:"alarm_duration_sec"`
	Enabled          bool      `json:"enabled"`
	CreatedAt        time.Time `json:"created_at"`
}

// SamplingPointInput 创建监测点入参。
type SamplingPointInput struct {
	BallastTankID           int64   `json:"tank_id"`
	Code             string  `json:"code"`
	ParamType        string  `json:"param_type"`
	ThresholdMin     float64 `json:"threshold_min"`
	ThresholdMax     float64 `json:"threshold_max"`
	AlarmDurationSec int     `json:"alarm_duration_sec"`
	Enabled          bool    `json:"enabled"`
}

// ParamTypes 合法参数类型集合。
var ParamTypes = map[string]bool{
	ParamTemp:       true,
	ParamHumidity:   true,
	ParamPressure:   true,
	ParamParticle05: true,
	ParamParticle50: true,
}

// ParamUnits 参数单位映射。
var ParamUnits = map[string]string{
	ParamTemp:       "℃",
	ParamHumidity:   "%",
	ParamPressure:   "Pa",
	ParamParticle05: "个/m³",
	ParamParticle50: "个/m³",
}

func (p *SamplingPoint) Validate() error {
	if p.BallastTankID <= 0 {
		return ErrBallastTankRequired
	}
	if p.Code == "" {
		return ErrCodeRequired
	}
	if !ParamTypes[p.ParamType] {
		return ErrInvalidParamType
	}
	if p.ThresholdMax < p.ThresholdMin {
		return ErrThresholdInverted
	}
	if p.AlarmDurationSec <= 0 {
		return ErrInvalidDuration
	}
	return nil
}

var (
	ErrBallastTankRequired       = errors.New("tank 必填")
	ErrInvalidParamType   = errors.New("参数类型非法")
	ErrThresholdInverted  = errors.New("阈值上限小于下限")
	ErrInvalidDuration    = errors.New("持续时长必须大于 0")
	ErrSamplingPointDisabled = errors.New("采样点已停用")
)
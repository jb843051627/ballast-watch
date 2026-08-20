package model

import (
	"errors"
	"time"
)

// BallastTank 房间：洁净区下的物理房间，绑定环境目标参数。
type BallastTank struct {
	ID                int64     `json:"id"`
	VesselID       int64     `json:"vessel_id"`
	Name              string    `json:"name"`
	Code              string    `json:"code"`
	Kind              string    `json:"kind"` // classify / atelier / locker
	AreaSqm           float64   `json:"area_sqm"`
	TargetPressure    float64   `json:"target_pressure"` // 目标压差 Pa
	PressureTolerance float64   `json:"pressure_tolerance"`
	TargetTemp        float64   `json:"target_temp"` // 目标温度 ℃
	TargetHumidity    float64   `json:"target_humidity"`
	Status            string    `json:"status"` // 当前洁净状态
	CreatedAt         time.Time `json:"created_at"`
}

// TankInput 创建/更新房间入参。
type TankInput struct {
	VesselID       int64   `json:"vessel_id"`
	Name              string  `json:"name"`
	Code              string  `json:"code"`
	Kind              string  `json:"kind"`
	AreaSqm           float64 `json:"area_sqm"`
	TargetPressure    float64 `json:"target_pressure"`
	PressureTolerance float64 `json:"pressure_tolerance"`
	TargetTemp        float64 `json:"target_temp"`
	TargetHumidity    float64 `json:"target_humidity"`
}

func (r *BallastTank) Validate() error {
	if r.VesselID <= 0 {
		return ErrVesselRequired
	}
	if r.Name == "" {
		return ErrNameRequired
	}
	if r.Code == "" {
		return ErrCodeRequired
	}
	switch r.Kind {
	case "classify", "atelier", "locker":
	default:
		return ErrInvalidKind
	}
	if r.TargetPressure <= 0 || r.PressureTolerance <= 0 {
		return ErrInvalidPressureTarget
	}
	return nil
}

// BallastTankKindNames 房间类型中文名（用于报表展示）。
func BallastTankKindNames() map[string]string {
	return map[string]string{
		"classify": "分类间",
		"atelier":  "操作间",
		"locker":   "更衣室",
	}
}

// NewBallastTankStatus 默认房间初始状态。
const NewBallastTankStatus = "at_rest"

// sentinel errors
var (
	ErrVesselRequired = errors.New("vessel 必填")
	ErrInvalidKind       = errors.New("房间类型非法")
	ErrInvalidPressureTarget = errors.New("压差目标与容差必须大于 0")
)
package model

import (
	"errors"
	"time"
)

// Calibration 传感器校准记录。
type Calibration struct {
	ID          int64     `json:"id"`
	SensorID    int64     `json:"sensor_id"`
	PerformedAt time.Time `json:"performed_at"`
	DueAt       time.Time `json:"due_at"`
	Standard    string    `json:"standard"`
	Result      string    `json:"result"` // pass / fail
	Operator    string    `json:"operator"`
	CreatedAt   time.Time `json:"created_at"`
}

// CalibrationInput 记录校准入参。
type CalibrationInput struct {
	SensorID    int64     `json:"sensor_id"`
	PerformedAt time.Time `json:"performed_at"`
	DueAt       time.Time `json:"due_at"`
	Standard    string    `json:"standard"`
	Result      string    `json:"result"`
	Operator    string    `json:"operator"`
}

func (c *Calibration) Validate() error {
	if c.SensorID <= 0 {
		return ErrSensorRequired
	}
	if c.Standard == "" {
		return ErrStandardRequired
	}
	if c.Result != "pass" && c.Result != "fail" {
		return ErrInvalidResult
	}
	if c.Operator == "" {
		return ErrOperatorRequired
	}
	if !c.DueAt.After(c.PerformedAt) {
		return ErrInvalidDueDate
	}
	return nil
}

// UpdateSensorDueDate 根据校准记录更新传感器下次校准时间。
func (c *Calibration) UpdateSensorDueDate(s *Sensor) {
	s.CalibrationDueAt = c.DueAt
	if c.Result == "fail" {
		s.Status = SensorFault
	} else if s.Status == SensorFault {
		s.Status = SensorActive
	}
}

var (
	ErrSensorRequired  = errors.New("sensor 必填")
	ErrStandardRequired = errors.New("校准标准必填")
	ErrInvalidResult   = errors.New("校准结果必须是 pass/fail")
	ErrOperatorRequired = errors.New("操作人必填")
	ErrInvalidDueDate  = errors.New("下次校准时间必须晚于校准时间")
)
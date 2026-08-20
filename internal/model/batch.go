package model

import (
	"errors"
	"time"
)

// 批次状态。
const (
	TreatmentCyclePlanning  = "planning"
	TreatmentCycleInProgress = "in_progress"
	TreatmentCycleCompleted = "completed"
	TreatmentCycleAborted   = "aborted"
)

// TreatmentCycle 生产批次：关联房间，驱动洁净状态机。
type TreatmentCycle struct {
	ID        int64     `json:"id"`
	BallastTankID    int64     `json:"tank_id"`
	Name      string    `json:"name"`
	Product   string    `json:"product"`
	Phase     string    `json:"phase"`
	StartAt   time.Time `json:"start_at"`
	EndAt     *time.Time `json:"end_at"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// TreatmentCycleInput 启动批次入参。
type TreatmentCycleInput struct {
	BallastTankID  int64  `json:"tank_id"`
	Name    string `json:"name"`
	Product string `json:"product"`
	Phase   string `json:"phase"`
}

// TreatmentCycleStatuses 合法批次状态。
var TreatmentCycleStatuses = map[string]bool{
	TreatmentCyclePlanning: true, TreatmentCycleInProgress: true, TreatmentCycleCompleted: true, TreatmentCycleAborted: true,
}

func (b *TreatmentCycle) Validate() error {
	if b.BallastTankID <= 0 {
		return ErrBallastTankRequired
	}
	if b.Name == "" {
		return ErrNameRequired
	}
	if b.Product == "" {
		return ErrProductRequired
	}
	if !TreatmentCycleStatuses[b.Status] {
		return ErrInvalidTreatmentCycleStatus
	}
	return nil
}

var (
	ErrProductRequired   = errors.New("产品必填")
	ErrInvalidTreatmentCycleStatus = errors.New("批次状态非法")
)
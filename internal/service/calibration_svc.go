package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// CalibrationService 校准管理。
type CalibrationService struct {
	cals    store.CalibrationStore
	sensors store.SensorStore
}

func NewCalibrationService(cals store.CalibrationStore, sensors store.SensorStore) *CalibrationService {
	return &CalibrationService{cals: cals, sensors: sensors}
}

// Record 记录校准并刷新传感器校准时间与状态。
func (s *CalibrationService) Record(ctx context.Context, in *model.CalibrationInput) (*model.Calibration, error) {
	se, _ := s.sensors.GetByID(ctx, in.SensorID)
	c := &model.Calibration{
		SensorID:    in.SensorID,
		PerformedAt: in.PerformedAt,
		DueAt:       in.DueAt,
		Standard:    in.Standard,
		Result:      in.Result,
		Operator:    in.Operator,
		CreatedAt:   time.Now(),
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.cals.Create(ctx, c); err != nil {
		return nil, err
	}
	c.UpdateSensorDueDate(se)
	if err := s.sensors.Update(ctx, se); err != nil {
		return nil, err
	}
	return c, nil
}

// List 列出某传感器校准记录。
func (s *CalibrationService) List(ctx context.Context, sensorID int64, limit int) ([]*model.Calibration, error) {
	return s.cals.ListBySensor(ctx, sensorID, limit)
}

// ListDueSensors 统计已到期未校准的传感器数。
func (s *CalibrationService) ListDueSensors(ctx context.Context, now time.Time) (int, error) {
	return s.cals.CountDueSensors(ctx, now)
}
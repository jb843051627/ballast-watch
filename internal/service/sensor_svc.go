package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
	"ballast-watch/internal/util"
)

// SensorService 传感器管理。
type SensorService struct {
	sensors store.SensorStore
	sampling_points  store.SamplingPointStore
}

func NewSensorService(sensors store.SensorStore, sampling_points store.SamplingPointStore) *SensorService {
	return &SensorService{sensors: sensors, sampling_points: sampling_points}
}

// Register 注册传感器（校验监测点存在）。
func (s *SensorService) Register(ctx context.Context, in *model.SensorInput) (*model.Sensor, error) {
	if _, err := s.sampling_points.GetByID(ctx, in.SamplingPointID); err != nil {
		return nil, err
	}
	se := &model.Sensor{
		SamplingPointID:          in.SamplingPointID,
		Serial:           in.Serial,
		Vendor:           in.Vendor,
		Battery:          in.Battery,
		Status:           model.SensorActive,
		CalibrationDueAt: in.CalibrationDueAt,
		CreatedAt:        time.Now(),
	}
	if err := se.Validate(); err != nil {
		return nil, err
	}
	if !util.ValidateSerial(se.Serial) {
		return nil, model.ErrInvalidInput
	}
	if err := s.sensors.Create(ctx, se); err != nil {
		return nil, err
	}
	return se, nil
}

// List 列出传感器。
func (s *SensorService) List(ctx context.Context, page model.Page) ([]*model.Sensor, error) {
	page.Normalize()
	return s.sensors.List(ctx, page.Limit, page.Offset)
}

// Get 获取传感器。
func (s *SensorService) Get(ctx context.Context, id int64) (*model.Sensor, error) {
	return s.sensors.GetByID(ctx, id)
}

// ListByPoint 按监测点列出传感器。
func (s *SensorService) ListByPoint(ctx context.Context, sampling_pointID int64) ([]*model.Sensor, error) {
	return s.sensors.ListByPoint(ctx, sampling_pointID)
}

// MarkOffline 刷新传感器在线状态。
func (s *SensorService) MarkOffline(ctx context.Context, now time.Time, offlineAfter time.Duration) error {
	page := model.Page{Limit: 500}
	sensors, err := s.sensors.List(ctx, page.Limit, page.Offset)
	if err != nil {
		return err
	}
	for _, se := range sensors {
		if se.Status == model.SensorFault {
			continue
		}
		se.MarkStatusBySeen(now, offlineAfter)
		if se.Status == model.SensorOffline {
			_ = s.sensors.Update(ctx, se)
		}
	}
	return nil
}
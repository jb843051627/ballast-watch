package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// SamplingPointService 监测点管理。
type SamplingPointService struct {
	sampling_points store.SamplingPointStore
	tanks  store.TankStore
}

func NewSamplingPointService(sampling_points store.SamplingPointStore, tanks store.TankStore) *SamplingPointService {
	return &SamplingPointService{sampling_points: sampling_points, tanks: tanks}
}

// Create 创建监测点（校验房间存在）。
func (s *SamplingPointService) Create(ctx context.Context, in *model.SamplingPointInput) (*model.SamplingPoint, error) {
	if _, err := s.tanks.GetByID(ctx, in.BallastTankID); err != nil {
		return nil, err
	}
	p := &model.SamplingPoint{
		BallastTankID:           in.BallastTankID,
		Code:             in.Code,
		ParamType:        in.ParamType,
		ThresholdMin:     in.ThresholdMin,
		ThresholdMax:     in.ThresholdMax,
		AlarmDurationSec: in.AlarmDurationSec,
		Enabled:          in.Enabled,
		CreatedAt:        time.Now(),
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.sampling_points.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// List 列出监测点。
func (s *SamplingPointService) List(ctx context.Context, tankID int64) ([]*model.SamplingPoint, error) {
	if tankID > 0 {
		return s.sampling_points.ListByBallastTank(ctx, tankID)
	}
	return s.sampling_points.ListAll(ctx)
}

// Get 获取监测点。
func (s *SamplingPointService) Get(ctx context.Context, id int64) (*model.SamplingPoint, error) {
	return s.sampling_points.GetByID(ctx, id)
}

// Toggle 启停监测点。
func (s *SamplingPointService) Toggle(ctx context.Context, id int64, enabled bool) (*model.SamplingPoint, error) {
	if err := s.sampling_points.SetEnabled(ctx, id, enabled); err != nil {
		return nil, err
	}
	return s.sampling_points.GetByID(ctx, id)
}
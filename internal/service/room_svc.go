package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// TankService 房间管理。
type TankService struct {
	tanks       store.TankStore
	vessels  store.VesselStore
}

func NewTankService(tanks store.TankStore, vessels store.VesselStore) *TankService {
	return &TankService{tanks: tanks, vessels: vessels}
}

// Create 创建房间（校验洁净区存在）。
func (s *TankService) Create(ctx context.Context, in *model.TankInput) (*model.BallastTank, error) {
	if _, err := s.vessels.GetByID(ctx, in.VesselID); err != nil {
		return nil, err
	}
	r := &model.BallastTank{
		VesselID:       in.VesselID,
		Name:              in.Name,
		Code:              in.Code,
		Kind:              in.Kind,
		AreaSqm:           in.AreaSqm,
		TargetPressure:    in.TargetPressure,
		PressureTolerance: in.PressureTolerance,
		TargetTemp:        in.TargetTemp,
		TargetHumidity:    in.TargetHumidity,
		Status:            model.NewBallastTankStatus,
		CreatedAt:         time.Now(),
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := s.tanks.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// List 列出房间。
func (s *TankService) List(ctx context.Context, page model.Page) ([]*model.BallastTank, error) {
	page.Normalize()
	return s.tanks.List(ctx, page.Limit, page.Offset)
}

// ListByVessel 按洁净区列出房间。
func (s *TankService) ListByVessel(ctx context.Context, vesselID int64) ([]*model.BallastTank, error) {
	return s.tanks.ListByVessel(ctx, vesselID)
}

// Get 获取房间。
func (s *TankService) Get(ctx context.Context, id int64) (*model.BallastTank, error) {
	return s.tanks.GetByID(ctx, id)
}

// Update 更新房间。
func (s *TankService) Update(ctx context.Context, id int64, in *model.TankInput) (*model.BallastTank, error) {
	r, err := s.tanks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r.Name = in.Name
	r.Code = in.Code
	r.Kind = in.Kind
	r.AreaSqm = in.AreaSqm
	r.TargetPressure = in.TargetPressure
	r.PressureTolerance = in.PressureTolerance
	r.TargetTemp = in.TargetTemp
	r.TargetHumidity = in.TargetHumidity
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := s.tanks.Update(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// GetStatus 获取房间当前状态。
func (s *TankService) GetStatus(ctx context.Context, id int64) (*model.ComplianceStatus, error) {
	r, err := s.tanks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.ComplianceStatus{
		BallastTankID:  r.ID,
		State:   r.Status,
		ChangedAt: time.Now(),
	}, nil
}
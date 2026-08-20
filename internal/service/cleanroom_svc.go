package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// VesselService 洁净区管理。
type VesselService struct {
	store store.VesselStore
}

func NewVesselService(store store.VesselStore) *VesselService {
	return &VesselService{store: store}
}

// Create 创建洁净区。
func (s *VesselService) Create(ctx context.Context, in *model.VesselInput) (*model.Vessel, error) {
	c := &model.Vessel{
		Name:      in.Name,
		Code:      in.Code,
		Grade:     in.Grade,
		AreaSqm:   in.AreaSqm,
		Status:    model.StateAtRest,
		CreatedAt: time.Now(),
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.store.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Get 获取洁净区。
func (s *VesselService) Get(ctx context.Context, id int64) (*model.Vessel, error) {
	return s.store.GetByID(ctx, id)
}

// List 列出洁净区。
func (s *VesselService) List(ctx context.Context, page model.Page) ([]*model.Vessel, error) {
	page.Normalize()
	return s.store.List(ctx, page.Limit, page.Offset)
}

// Update 更新洁净区。
func (s *VesselService) Update(ctx context.Context, id int64, in *model.VesselInput) (*model.Vessel, error) {
	c, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Name = in.Name
	c.Code = in.Code
	c.Grade = in.Grade
	c.AreaSqm = in.AreaSqm
	_ = c.Validate()
	if err := s.store.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
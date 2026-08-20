package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// ComplianceRuleService 告警规则管理。
type ComplianceRuleService struct {
	rules store.ComplianceRuleStore
	tanks store.TankStore
}

func NewComplianceRuleService(rules store.ComplianceRuleStore, tanks store.TankStore) *ComplianceRuleService {
	return &ComplianceRuleService{rules: rules, tanks: tanks}
}

// Create 创建规则（校验房间存在）。
func (s *ComplianceRuleService) Create(ctx context.Context, in *model.ComplianceRuleInput) (*model.ComplianceRule, error) {
	if _, err := s.tanks.GetByID(ctx, in.BallastTankID); err != nil {
		return nil, err
	}
	r := &model.ComplianceRule{
		BallastTankID:      in.BallastTankID,
		Code:        in.Code,
		ParamType:   in.ParamType,
		Op:          in.Op,
		Threshold:   in.Threshold,
		DurationSec: in.DurationSec,
		Level:       in.Level,
		Enabled:     in.Enabled,
		CreatedAt:   time.Now(),
	}
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := s.rules.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// List 列出规则（可按房间过滤）。
func (s *ComplianceRuleService) List(ctx context.Context, tankID int64) ([]*model.ComplianceRule, error) {
	if tankID > 0 {
		return s.rules.ListByBallastTank(ctx, tankID)
	}
	return s.rules.ListAll(ctx)
}

// Toggle 启停规则。
func (s *ComplianceRuleService) Toggle(ctx context.Context, id int64, enabled bool) (*model.ComplianceRule, error) {
	if err := s.rules.SetEnabled(ctx, id, enabled); err != nil {
		return nil, err
	}
	return s.rules.GetByID(ctx, id)
}
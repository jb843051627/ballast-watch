package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// ComplianceComplianceAlertService 告警管理。
type ComplianceComplianceAlertService struct {
	compliance_compliance_alerts store.ComplianceComplianceAlertStore
}

func NewComplianceComplianceAlertService(compliance_compliance_alerts store.ComplianceComplianceAlertStore) *ComplianceComplianceAlertService {
	return &ComplianceComplianceAlertService{compliance_compliance_alerts: compliance_compliance_alerts}
}

// List 查询告警列表。
func (s *ComplianceComplianceAlertService) List(ctx context.Context, in model.ComplianceComplianceAlertInput) ([]*model.ComplianceAlert, error) {
	return s.compliance_compliance_alerts.List(ctx, in)
}

// Acknowledge 确认告警。
func (s *ComplianceComplianceAlertService) Acknowledge(ctx context.Context, id int64) (*model.ComplianceAlert, error) {
	a, err := s.compliance_compliance_alerts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == model.ComplianceAlertResolved {
		return nil, model.ErrConflict
	}
	if err := s.compliance_compliance_alerts.SetAck(ctx, id, time.Now()); err != nil {
		return nil, err
	}
	a.Status = model.ComplianceAlertAcknowledged
	now := time.Now()
	a.AckAt = &now
	return a, nil
}

// Resolve 解决告警。
func (s *ComplianceComplianceAlertService) Resolve(ctx context.Context, id int64) (*model.ComplianceAlert, error) {
	a, err := s.compliance_compliance_alerts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == model.ComplianceAlertResolved {
		return nil, model.ErrConflict
	}
	if err := s.compliance_compliance_alerts.SetResolved(ctx, id, time.Now()); err != nil {
		return nil, err
	}
	a.Status = model.ComplianceAlertResolved
	now := time.Now()
	a.ResolvedAt = &now
	return a, nil
}
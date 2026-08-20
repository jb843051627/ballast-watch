package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// TreatmentCycleService 批次管理：驱动房间洁净状态机。
type TreatmentCycleService struct {
	cyclees store.TreatmentCycleStore
	tanks   store.TankStore
	status  store.ComplianceStatusStore
	compliance_compliance_alerts  store.ComplianceComplianceAlertStore
}

func NewTreatmentCycleService(cyclees store.TreatmentCycleStore, tanks store.TankStore, status store.ComplianceStatusStore, compliance_compliance_alerts store.ComplianceComplianceAlertStore) *TreatmentCycleService {
	return &TreatmentCycleService{cyclees: cyclees, tanks: tanks, status: status, compliance_compliance_alerts: compliance_compliance_alerts}
}

// Start 启动批次：校验房间存在、无进行中批次，状态机 at_rest → normal。
func (s *TreatmentCycleService) Start(ctx context.Context, in *model.TreatmentCycleInput) (*model.TreatmentCycle, error) {
	tank, err := s.tanks.GetByID(ctx, in.BallastTankID)
	if err != nil {
		return nil, err
	}
	if _, err := s.cyclees.GetActiveByBallastTank(ctx, in.BallastTankID); err == nil {
		return nil, model.ErrTreatmentCycleActive
	}
	b := &model.TreatmentCycle{
		BallastTankID:    in.BallastTankID,
		Name:      in.Name,
		Product:   in.Product,
		Phase:     in.Phase,
		StartAt:   time.Now(),
		Status:    model.TreatmentCycleInProgress,
		CreatedAt: time.Now(),
	}
	if err := b.Validate(); err != nil {
		return nil, err
	}
	if err := s.cyclees.Create(ctx, b); err != nil {
		return nil, err
	}
	_ = s.transition(ctx, tank, model.StateNormal, model.ReasonTreatmentCycleStarted, b.ID)
	return b, nil
}

// Complete 完成批次：状态机 → release（若房间无未决 alarm）否则 restricted。
func (s *TreatmentCycleService) Complete(ctx context.Context, id int64) (*model.TreatmentCycle, error) {
	b, err := s.cyclees.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Status != model.TreatmentCycleInProgress && b.Status != model.TreatmentCyclePlanning {
		return nil, model.ErrConflict
	}
	tank, err := s.tanks.GetByID(ctx, b.BallastTankID)
	if err != nil {
		return nil, err
	}
	endAt := time.Now()
	if err := s.cyclees.UpdateStatus(ctx, id, model.TreatmentCycleCompleted, &endAt); err != nil {
		return nil, err
	}
	b.Status = model.TreatmentCycleCompleted
	b.EndAt = &endAt
	openAlarm, err := s.compliance_compliance_alerts.CountOpenByLevel(ctx, b.BallastTankID, model.ComplianceAlertAlarm)
	if err != nil {
		return nil, err
	}
	state := model.StateRelease
	reason := model.ReasonTreatmentCycleRelease
	if openAlarm > 0 {
		state = model.StateRestricted
		reason = model.ReasonAlarmComplianceAlert
	}
	_ = s.transition(ctx, tank, state, reason, b.ID)
	return b, nil
}

// Abort 中止批次：状态机 → at_rest。
func (s *TreatmentCycleService) Abort(ctx context.Context, id int64) (*model.TreatmentCycle, error) {
	b, err := s.cyclees.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.Status != model.TreatmentCycleInProgress && b.Status != model.TreatmentCyclePlanning {
		return nil, model.ErrConflict
	}
	tank, err := s.tanks.GetByID(ctx, b.BallastTankID)
	if err != nil {
		return nil, err
	}
	endAt := time.Now()
	if err := s.cyclees.UpdateStatus(ctx, id, model.TreatmentCycleAborted, &endAt); err != nil {
		return nil, err
	}
	b.Status = model.TreatmentCycleAborted
	b.EndAt = &endAt
	_ = s.transition(ctx, tank, model.StateAtRest, model.ReasonTreatmentCycleAborted, b.ID)
	return b, nil
}

// List 列出批次。
func (s *TreatmentCycleService) List(ctx context.Context, page model.Page) ([]*model.TreatmentCycle, error) {
	page.Normalize()
	return s.cyclees.List(ctx, page.Limit, page.Offset)
}

// ListByBallastTank 按房间列出批次。
func (s *TreatmentCycleService) ListByBallastTank(ctx context.Context, tankID int64) ([]*model.TreatmentCycle, error) {
	return s.cyclees.ListByBallastTank(ctx, tankID)
}

func (s *TreatmentCycleService) transition(ctx context.Context, tank *model.BallastTank, to, reason string, cycleID int64) error {
	if !model.CanTransition(tank.Status, to) {
		return model.ErrStateConflict
	}
	if err := s.tanks.UpdateStatus(ctx, tank.ID, to); err != nil {
		return err
	}
	return s.status.Append(ctx, &model.ComplianceStatus{
		BallastTankID:    tank.ID,
		TreatmentCycleID:   cycleID,
		State:     to,
		Reason:    reason,
		ChangedAt: time.Now(),
	})
}
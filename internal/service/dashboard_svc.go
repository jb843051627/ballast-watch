package service

import (
	"context"
	"errors"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// DashboardSnapshot 看板快照。
type DashboardSnapshot struct {
	BallastTanks      []*store.TankSnapshot `json:"tanks"`
	OpenComplianceAlerts int                   `json:"open_compliance_compliance_alerts"`
	UpdatedAt  time.Time             `json:"updated_at"`
}

// DashboardService 看板服务：缓存快照刷新与读取。
type DashboardService struct {
	tanks    store.TankStore
	sampling_points   store.SamplingPointStore
	cache    *store.Cache
	compliance_compliance_alerts   store.ComplianceComplianceAlertStore
	water_readings store.WaterWaterReadingStore
}

func NewDashboardService(tanks store.TankStore, sampling_points store.SamplingPointStore, cache *store.Cache,
	compliance_compliance_alerts store.ComplianceComplianceAlertStore, water_readings store.WaterWaterReadingStore) *DashboardService {
	return &DashboardService{tanks: tanks, sampling_points: sampling_points, cache: cache, compliance_compliance_alerts: compliance_compliance_alerts, water_readings: water_readings}
}

// Snapshot 读取看板快照（缓存优先）。
func (s *DashboardService) Snapshot(ctx context.Context) (*DashboardSnapshot, error) {
	snaps := s.cache.GetAll()
	openComplianceAlerts := 0
	for _, snap := range snaps {
		openComplianceAlerts += snap.OpenComplianceAlerts
	}
	return &DashboardSnapshot{
		BallastTanks:      snaps,
		OpenComplianceAlerts: openComplianceAlerts,
		UpdatedAt:  time.Now(),
	}, nil
}

// Refresh 刷新全部房间快照（后台定期调用）。
func (s *DashboardService) Refresh(ctx context.Context) error {
	tanks, err := s.tanks.List(ctx, 1000, 0)
	if err != nil {
		return err
	}
	for _, r := range tanks {
		// 大屏切换船舶/停机时 ctx 被取消，立即收尾，不再继续查所有实时快照
		if err := ctx.Err(); err != nil {
			return err
		}
		snap, err := s.buildTankSnapshot(ctx, r)
		if err != nil {
			return err
		}
		s.cache.Set(r.ID, snap)
	}
	return nil
}

func (s *DashboardService) buildTankSnapshot(ctx context.Context, tank *model.BallastTank) (*store.TankSnapshot, error) {
	snap := &store.TankSnapshot{
		BallastTankID:   tank.ID,
		BallastTankCode: tank.Code,
		Status:   tank.Status,
	}
	openComplianceAlerts, err := s.compliance_compliance_alerts.CountOpenByBallastTank(ctx, tank.ID)
	if err != nil {
		return nil, err
	}
	snap.OpenComplianceAlerts = openComplianceAlerts
	sampling_points, err := s.sampling_points.ListByBallastTank(ctx, tank.ID)
	if err != nil {
		return nil, err
	}
	var realtime []model.RealtimeWaterWaterReading
	for _, p := range sampling_points {
		latest, err := s.water_readings.LatestByPoint(ctx, p.ID)
		if err != nil {
			// 无读数时跳过该点；其余错误（含 ctx 取消）必须上抛，避免吞掉取消信号导致刷新迟迟不收尾
			if errors.Is(err, model.ErrNotFound) {
				continue
			}
			return nil, err
		}
		realtime = append(realtime, model.RealtimeWaterWaterReading{
			SamplingPointID:    p.ID,
			BallastTankID:     p.BallastTankID,
			ParamType:  p.ParamType,
			Value:      latest.Value,
			MeasuredAt: latest.MeasuredAt,
			WithinRange: latest.Value >= p.ThresholdMin && latest.Value <= p.ThresholdMax,
		})
	}
	snap.Realtime = realtime
	return snap, nil
}
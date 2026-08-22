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
	var refreshErr error
	for _, r := range tanks {
		snap, err := s.buildTankSnapshot(ctx, r)
		if err != nil {
			// 采样点查询失败时静默吞错会让大屏显示空舱并报成功，值班员会误判没有数据。
			// 记录首个失败原因并向上抛出，由后台任务记录日志而非上报成功。
			if refreshErr == nil {
				refreshErr = err
			}
			continue
		}
		s.cache.Set(r.ID, snap)
	}
	return refreshErr
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
			// ErrNotFound 表示该点确实无读数，跳过即可；其它错误（DB 故障等）必须上抛，
			// 否则会被当成"空舱"静默写入缓存，值班员误判为没有数据。
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
package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
	"ballast-watch/internal/util"
)

// WaterWaterReadingService 读数服务：入库、去重、评估触发、查询聚合。
type WaterWaterReadingService struct {
	water_readings  store.WaterWaterReadingStore
	sampling_points    store.SamplingPointStore
	sensors   store.SensorStore
	compliance_compliance_alerts    store.ComplianceComplianceAlertStore
	cache     *store.Cache
	engine    Engine
	mu        sync.Mutex
	realtimeStamp time.Time
}

// Engine 告警评估引擎接口（由 compliance_alerter.Engine 实现，避免循环依赖）。
type Engine interface {
	Evaluate(ctx context.Context, water_readings []*model.WaterReading) error
}

// readingDupKey 批内按秒去重的组合键：同一采样点同一秒只保留一条。
type readingDupKey struct {
	SamplingPointID int64
	sec             int64
}

// NewWaterWaterReadingService 创建读数服务。
func NewWaterWaterReadingService(water_readings store.WaterWaterReadingStore, sampling_points store.SamplingPointStore, sensors store.SensorStore,
	compliance_compliance_alerts store.ComplianceComplianceAlertStore, cache *store.Cache, engine Engine) *WaterWaterReadingService {
	return &WaterWaterReadingService{
		water_readings: water_readings,
		sampling_points:   sampling_points,
		sensors:  sensors,
		compliance_compliance_alerts:   compliance_compliance_alerts,
		cache:    cache,
		engine:   engine,
	}
}

// Ingest 批量上报读数：整批校验 + 去重 + 事务入库 + 评估。
// 加锁串行化判重-入库临界区，避免并发重发请求双双通过 ExistsDup 后各自落库。
func (s *WaterWaterReadingService) Ingest(ctx context.Context, cycle *model.WaterWaterReadingTreatmentCycle) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(cycle.WaterReadings) == 0 {
		return 0, model.ErrInvalidInput
	}
	now := time.Now()
	var toInsert []*model.WaterReading
	seen := make(map[readingDupKey]bool)
	for _, in := range cycle.WaterReadings {
		if !model.ParamTypes[in.ParamType] {
			return 0, model.ErrInvalidParamType
		}
		measuredAt, err := util.ParseTime(in.MeasuredAt)
		if err != nil {
			return 0, err
		}
		if measuredAt.IsZero() {
			measuredAt = now
		}
		// 统一截断到秒：同一采样点同一秒视为重复采样点，仅保留一条。
		measuredAt = util.TruncateSecond(measuredAt)
		key := readingDupKey{SamplingPointID: in.SamplingPointID, sec: measuredAt.Unix()}
		if seen[key] {
			continue
		}
		seen[key] = true
		dup, err := s.water_readings.ExistsDup(ctx, in.SamplingPointID, measuredAt)
		if err != nil {
			return 0, err
		}
		if dup {
			continue
		}
		toInsert = append(toInsert, &model.WaterReading{
			SamplingPointID:    in.SamplingPointID,
			SensorID:   in.SensorID,
			ParamType:  in.ParamType,
			Value:      in.Value,
			MeasuredAt: measuredAt,
			Raw:        in.Raw,
			CreatedAt:  now,
		})
	}
	if len(toInsert) == 0 {
		return 0, nil
	}
	if err := s.water_readings.InsertTreatmentCycle(ctx, toInsert); err != nil {
		return 0, err
	}
	for _, r := range toInsert {
		if r.SensorID > 0 {
			_ = s.sensors.UpdateLastSeen(ctx, r.SensorID, now)
		}
	}
	if s.engine != nil {
		if err := s.engine.Evaluate(ctx, toInsert); err != nil {
			return len(toInsert), err
		}
	}
	return len(toInsert), nil
}

// Query 查询读数。
func (s *WaterWaterReadingService) Query(ctx context.Context, q store.WaterReadingQuery) ([]*model.WaterReading, error) {
	return s.water_readings.Query(ctx, q)
}

// Stats 按参数聚合统计。
func (s *WaterWaterReadingService) Stats(ctx context.Context, paramType string, from, to time.Time) (*model.WaterWaterReadingStats, error) {
	return s.water_readings.Stats(ctx, paramType, from, to)
}

// RealtimeSnapshot 实时读数快照（按房间）。
func (s *WaterWaterReadingService) RealtimeSnapshot(ctx context.Context, tankID int64) (*store.TankSnapshot, error) {
	if snap, ok := s.cache.Get(tankID); ok {
		return snap, nil
	}
	sampling_points, err := s.sampling_points.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	var realtime []model.RealtimeWaterWaterReading
	for _, p := range sampling_points {
		if tankID > 0 && p.BallastTankID != tankID {
			continue
		}
		latest, err := s.water_readings.LatestByPoint(ctx, p.ID)
		if err != nil {
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
	snap := &store.TankSnapshot{
		BallastTankID: tankID,
		Realtime: realtime,
		UpdatedAt: time.Now(),
	}
	return snap, nil
}
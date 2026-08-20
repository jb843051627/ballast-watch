package service

import (
	"context"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
	"ballast-watch/internal/util"
)

// ReportService 报表服务：趋势、汇总、达标率。
type ReportService struct {
	water_readings    store.WaterWaterReadingStore
	sampling_points      store.SamplingPointStore
	tanks       store.TankStore
	vessels  store.VesselStore
	compliance_compliance_alerts      store.ComplianceComplianceAlertStore
	status      store.ComplianceStatusStore
}

func NewReportService(water_readings store.WaterWaterReadingStore, sampling_points store.SamplingPointStore, tanks store.TankStore,
	vessels store.VesselStore, compliance_compliance_alerts store.ComplianceComplianceAlertStore, status store.ComplianceStatusStore) *ReportService {
	return &ReportService{water_readings: water_readings, sampling_points: sampling_points, tanks: tanks, vessels: vessels, compliance_compliance_alerts: compliance_compliance_alerts, status: status}
}

// TrendPoint 趋势点。
type TrendPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
	Min   float64   `json:"min"`
	Max   float64   `json:"max"`
}

// TrendResult 趋势报表结果。
type TrendResult struct {
	BallastTankCode  string       `json:"tank_code"`
	PointCode string       `json:"sampling_point_code"`
	ParamType string       `json:"param_type"`
	Unit      string       `json:"unit"`
	Points    []TrendPoint `json:"sampling_points"`
}

// Trend 房间某参数趋势。
func (s *ReportService) Trend(ctx context.Context, tankID int64, paramType string, from, to time.Time, limit int) (*TrendResult, error) {
	tank, err := s.tanks.GetByID(ctx, tankID)
	if err != nil {
		return nil, err
	}
	if !model.ParamTypes[paramType] {
		return nil, model.ErrInvalidParamType
	}
	pts, err := s.sampling_points.ListByBallastTank(ctx, tankID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	res := &TrendResult{
		BallastTankCode:  tank.Code,
		ParamType: paramType,
		Unit:      model.ParamUnits[paramType],
	}
	for _, p := range pts {
		if p.ParamType != paramType {
			continue
		}
		water_readings, err := s.water_readings.QueryValuesWithTime(ctx, p.ID, paramType, from, to)
		if err != nil {
			return nil, err
		}
		if len(water_readings) == 0 {
			continue
		}
		if len(res.PointCode) == 0 {
			res.PointCode = p.Code
		}
		for _, r := range water_readings {
			res.Points = append(res.Points, TrendPoint{
				Time:  r.MeasuredAt,
				Value: r.Value,
				Min:   p.ThresholdMin,
				Max:   p.ThresholdMax,
			})
		}
	}
	if len(res.Points) > limit {
		res.Points = res.Points[len(res.Points)-limit:]
	}
	return res, nil
}

// BallastTankCompliance 房间达标率。
type BallastTankCompliance struct {
	BallastTankID     int64   `json:"tank_id"`
	BallastTankCode   string  `json:"tank_code"`
	ParamType  string  `json:"param_type"`
	WaterReadings   int     `json:"water_readings"`
	OutOfRange int     `json:"out_of_range"`
	Rate       float64 `json:"rate"` // 0-100
}

// SummaryResult 汇总报表。
type SummaryResult struct {
	GeneratedAt time.Time         `json:"generated_at"`
	BallastTankCount   int               `json:"tank_count"`
	Vessels  int               `json:"vessels"`
	Compliance  []BallastTankCompliance  `json:"compliance"`
	OpenComplianceAlerts  int               `json:"open_compliance_compliance_alerts"`
	DueSensors  int               `json:"due_sensors"`
}

// Summary 汇总各房间各参数达标率。
func (s *ReportService) Summary(ctx context.Context, from, to time.Time) (*SummaryResult, error) {
	res := &SummaryResult{GeneratedAt: time.Now()}
	tanks, err := s.tanks.List(ctx, 1000, 0)
	if err != nil {
		return nil, err
	}
	res.BallastTankCount = len(tanks)
	cc, err := s.vessels.Count(ctx)
	if err != nil {
		return nil, err
	}
	res.Vessels = cc
	openComplianceAlerts, err := s.compliance_compliance_alerts.List(ctx, model.ComplianceComplianceAlertInput{Status: model.ComplianceAlertOpen})
	if err != nil {
		return nil, err
	}
	res.OpenComplianceAlerts = len(openComplianceAlerts)
	for _, r := range tanks {
		pts, err := s.sampling_points.ListByBallastTank(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		for _, p := range pts {
			if seen[p.ParamType] {
				continue
			}
			seen[p.ParamType] = true
			vals, err := s.water_readings.QueryValues(ctx, p.ID, p.ParamType, from, to)
			if err != nil {
				return nil, err
			}
			oor := 0
			for _, v := range vals {
				if v < p.ThresholdMin || v > p.ThresholdMax {
					oor++
				}
			}
			rate := 0.0
			if len(vals) > 0 {
				rate = float64(oor) / float64(len(vals)) * 100
			}
			res.Compliance = append(res.Compliance, BallastTankCompliance{
				BallastTankID:     r.ID,
				BallastTankCode:   r.Code,
				ParamType:  p.ParamType,
				WaterReadings:   len(vals),
				OutOfRange: oor,
				Rate:       util.Round(rate, 1),
			})
		}
	}
	return res, nil
}

// StateDistribution 房间状态分布（看板用）。
type StateDistribution struct {
	State string `json:"state"`
	Count int    `json:"count"`
}

// StateDistribution 统计各状态房间数。
func (s *ReportService) StateDistribution(ctx context.Context) ([]StateDistribution, error) {
	var out []StateDistribution
	for state := range model.States {
		n, err := s.status.CountByState(ctx, state)
		if err != nil {
			return nil, err
		}
		out = append(out, StateDistribution{State: state, Count: n})
	}
	return out, nil
}
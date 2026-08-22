package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"ballast-watch/internal/store"
	"ballast-watch/internal/util"
)

// ExportService 导出服务：CSV 生成。
type ExportService struct {
	water_readings store.WaterWaterReadingStore
	sampling_points   store.SamplingPointStore
	tanks    store.TankStore
}

func NewExportService(water_readings store.WaterWaterReadingStore, sampling_points store.SamplingPointStore, tanks store.TankStore) *ExportService {
	return &ExportService{water_readings: water_readings, sampling_points: sampling_points, tanks: tanks}
}

// ExportWaterReadingsCSV 导出房间读数 CSV。
func (s *ExportService) ExportWaterReadingsCSV(ctx context.Context, tankID int64, from, to time.Time) ([]byte, error) {
	tank, err := s.tanks.GetByID(ctx, tankID)
	if err != nil {
		return nil, err
	}
	pts, err := s.sampling_points.ListByBallastTank(ctx, tankID)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"tank_code", "sampling_point_code", "param_type", "value", "measured_at", "within_range"})
	for _, p := range pts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		water_readings, err := s.water_readings.QueryValuesWithTime(ctx, p.ID, p.ParamType, from, to)
		if err != nil {
			return nil, err
		}
		for _, r := range water_readings {
			within := "yes"
			if r.Value < p.ThresholdMin || r.Value > p.ThresholdMax {
				within = "no"
			}
			_ = w.Write([]string{
				tank.Code,
				p.Code,
				p.ParamType,
				strconv.FormatFloat(r.Value, 'f', 2, 64),
				util.FormatTime(r.MeasuredAt),
				within,
			})
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BuildFilename 生成导出文件名。
func (s *ExportService) BuildFilename(tankCode string, day time.Time) string {
	return fmt.Sprintf("water_readings_%s_%s.csv", strings.ToLower(tankCode), util.FormatDate(day))
}
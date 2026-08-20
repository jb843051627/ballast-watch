package report

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"ballast-watch/internal/model"
)

// ExportRow CSV 导出行。
type ExportRow struct {
	BallastTankCode    string
	PointCode   string
	ParamType   string
	Value       float64
	MeasuredAt  time.Time
	WithinRange bool
}

// WriteCSV 将导出行写入 CSV 字节流。
func WriteCSV(rows []ExportRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{"tank_code", "sampling_point_code", "param_type", "value", "measured_at", "within_range"}); err != nil {
		return nil, err
	}
	for _, r := range rows {
		within := "yes"
		if !r.WithinRange {
			within = "no"
		}
		rec := []string{
			r.BallastTankCode,
			r.PointCode,
			r.ParamType,
			strconv.FormatFloat(r.Value, 'f', 2, 64),
			r.MeasuredAt.Format(time.RFC3339),
			within,
		}
		if err := w.Write(rec); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Filename 生成 CSV 文件名。
func Filename(tankCode string, day time.Time) string {
	return fmt.Sprintf("water_readings_%s_%s.csv", tankCode, day.Format("20060102"))
}

// FormatValue 格式化数值。
func FormatValue(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// ParamUnit 参数单位。
func ParamUnit(t string) string {
	return model.ParamUnits[t]
}
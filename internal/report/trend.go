package report

import (
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/util"
)

// TrendItem 趋势序列单点。
type TrendItem struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// TrendSeries 趋势序列（按点位）。
type TrendSeries struct {
	PointCode string       `json:"sampling_point_code"`
	ParamType string       `json:"param_type"`
	Unit      string       `json:"unit"`
	Items     []TrendItem  `json:"items"`
}

// BuildTrendSeries 构造趋势序列。
func BuildTrendSeries(sampling_pointCode, paramType string, water_readings []model.WaterReading) TrendSeries {
	items := make([]TrendItem, 0, len(water_readings))
	for _, r := range water_readings {
		items = append(items, TrendItem{Timestamp: r.MeasuredAt, Value: r.Value})
	}
	return TrendSeries{
		PointCode: sampling_pointCode,
		ParamType: paramType,
		Unit:      model.ParamUnits[paramType],
		Items:     items,
	}
}

// Downsample 降采样：超过 max 点则均匀抽稀。
func Downsample(series TrendSeries, max int) TrendSeries {
	if max <= 0 || len(series.Items) <= max {
		return series
	}
	step := float64(len(series.Items)) / float64(max)
	out := make([]TrendItem, 0, max)
	for i := 0; i < max; i++ {
		idx := int(float64(i) * step)
		if idx >= len(series.Items) {
			idx = len(series.Items) - 1
		}
		out = append(out, series.Items[idx])
	}
	series.Items = out
	return series
}

// Bucket 时间桶。
type Bucket struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Avg   float64   `json:"avg"`
	Max   float64   `json:"max"`
	Count int       `json:"count"`
}

// Bucketize 按时长分桶聚合。
func Bucketize(water_readings []model.WaterReading, bucketDur time.Duration) []Bucket {
	if len(water_readings) == 0 {
		return nil
	}
	start := water_readings[0].MeasuredAt
	var buckets []Bucket
	var cur *Bucket
	for _, r := range water_readings {
		idx := int(r.MeasuredAt.Sub(start) / bucketDur)
		for len(buckets) <= idx {
			bStart := start.Add(time.Duration(len(buckets)) * bucketDur)
			buckets = append(buckets, Bucket{
				Start: bStart,
				End:   bStart.Add(bucketDur),
			})
		}
		cur = &buckets[idx]
		cur.Count++
		cur.Avg += r.Value
		if r.Value > cur.Max {
			cur.Max = r.Value
		}
	}
	for i := range buckets {
		if buckets[i].Count > 0 {
			buckets[i].Avg = util.Round(buckets[i].Avg/float64(buckets[i].Count), 2)
		}
	}
	return buckets
}
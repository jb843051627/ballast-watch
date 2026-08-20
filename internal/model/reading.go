package model

import "time"

// WaterReading 单条环境读数。
type WaterReading struct {
	ID         int64     `json:"id"`
	SamplingPointID    int64     `json:"sampling_sampling_point_id"`
	SensorID   int64     `json:"sensor_id"`
	ParamType  string    `json:"param_type"`
	Value      float64   `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	Raw        string    `json:"raw"`
	CreatedAt  time.Time `json:"created_at"`
}

// WaterWaterReadingTreatmentCycle 批量上报的读数（网关/模拟器）。
type WaterWaterReadingTreatmentCycle struct {
	WaterReadings []WaterWaterReadingInput `json:"water_readings"`
}

// WaterWaterReadingInput 单条读数入参。
type WaterWaterReadingInput struct {
	SamplingPointID    int64   `json:"sampling_sampling_point_id"`
	SensorID   int64   `json:"sensor_id"`
	ParamType  string  `json:"param_type"`
	Value      float64 `json:"value"`
	MeasuredAt string  `json:"measured_at"` // RFC3339
	Raw        string  `json:"raw"`
}

// WaterWaterReadingStats 按参数聚合统计结果。
type WaterWaterReadingStats struct {
	ParamType  string  `json:"param_type"`
	Count      int     `json:"count"`
	Min        float64 `json:"min"`
	Max        float64 `json:"max"`
	Avg        float64 `json:"avg"`
	OutOfRange int     `json:"out_of_range"` // 超阈值条数
	Unit       string  `json:"unit"`
}

// RealtimeWaterWaterReading 看板实时读数（合并点位与房间信息）。
type RealtimeWaterWaterReading struct {
	SamplingPointID   int64     `json:"sampling_sampling_point_id"`
	BallastTankID    int64     `json:"tank_id"`
	BallastTankCode  string    `json:"tank_code"`
	ParamType string    `json:"param_type"`
	Value     float64   `json:"value"`
	MeasuredAt time.Time `json:"measured_at"`
	WithinRange bool    `json:"within_range"`
}
package collector

import (
	"context"
	"log"
	"math/rand"
	"time"

	"ballast-watch/internal/config"
	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
	"ballast-watch/internal/util"
)

// Simulator 模拟器：周期为各监测点生成模拟读数并投递到采集器。
type Simulator struct {
	cfg      *config.AppConfig
	sampling_points   store.SamplingPointStore
	sensors  store.SensorStore
	collect  *Collector
	stopCh   chan struct{}
	rand     *rand.Rand
}

// NewSimulator 创建模拟器。
func NewSimulator(cfg *config.AppConfig, sampling_points store.SamplingPointStore, sensors store.SensorStore, collect *Collector) *Simulator {
	return &Simulator{
		cfg:     cfg,
		sampling_points:  sampling_points,
		sensors: sensors,
		collect: collect,
		stopCh:  make(chan struct{}),
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Start 启动模拟器。
func (s *Simulator) Start() {
	go s.run()
}

func (s *Simulator) run() {
	ticker := time.NewTicker(s.cfg.SimulatorInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

// tick 生成一批读数。
func (s *Simulator) tick() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	sampling_points, err := s.sampling_points.ListAll(ctx)
	if err != nil {
		log.Printf("[simulator] 读取点位失败: %v", err)
		return
	}
	var cycle model.WaterWaterReadingTreatmentCycle
	now := time.Now()
	for _, p := range sampling_points {
		if !p.Enabled {
			continue
		}
		sensors, err := s.sensors.ListByPoint(ctx, p.ID)
		if err != nil || len(sensors) == 0 {
			continue
		}
		se := sensors[s.rand.Intn(len(sensors))]
		value := s.sample(p)
		cycle.WaterReadings = append(cycle.WaterReadings, model.WaterWaterReadingInput{
			SamplingPointID:    p.ID,
			SensorID:   se.ID,
			ParamType:  p.ParamType,
			Value:      value,
			MeasuredAt: now.Format(time.RFC3339),
		})
	}
	if len(cycle.WaterReadings) > 0 {
		s.collect.Enqueue(&cycle)
	}
}

// sample 依据点位阈值生成采样值：多数落在阈值内，偶发越界。
func (s *Simulator) sample(p *model.SamplingPoint) float64 {
	mid := (p.ThresholdMin + p.ThresholdMax) / 2
	span := p.ThresholdMax - p.ThresholdMin
	if span <= 0 {
		return mid
	}
	// 90% 正常波动，10% 越界（触发告警）
	if s.rand.Intn(100) < 10 {
		offset := span * (0.5 + s.rand.Float64())
		if s.rand.Intn(2) == 0 {
			return p.ThresholdMin - offset
		}
		return p.ThresholdMax + offset
	}
	// 正常波动在阈值的 ±30% 内
	jitter := span * 0.3 * (s.rand.Float64()*2 - 1)
	return clamp(mid+jitter, p.ThresholdMin, p.ThresholdMax)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Stop 停止模拟器。
func (s *Simulator) Stop() {
	close(s.stopCh)
}

// NowAlias 供测试对齐时间。
var NowAlias = util.Now
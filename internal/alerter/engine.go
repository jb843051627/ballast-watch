package compliance_alerter

import (
	"context"
	"errors"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
	"ballast-watch/internal/util"
)

// Engine 告警评估引擎：对入库读数匹配规则并维护告警状态。
type Engine struct {
	rules  store.ComplianceRuleStore
	compliance_compliance_alerts store.ComplianceComplianceAlertStore
	sampling_points store.SamplingPointStore
	reads  store.WaterWaterReadingStore
}

// NewEngine 创建引擎。
func NewEngine(rules store.ComplianceRuleStore, compliance_compliance_alerts store.ComplianceComplianceAlertStore, sampling_points store.SamplingPointStore) *Engine {
	return &Engine{rules: rules, compliance_compliance_alerts: compliance_compliance_alerts, sampling_points: sampling_points}
}

// WithWaterWaterReadingStore 注入读数存储（评估窗口查询用）。
func (e *Engine) WithWaterWaterReadingStore(reads store.WaterWaterReadingStore) *Engine {
	e.reads = reads
	return e
}

// Evaluate 评估一批新入库读数。
func (e *Engine) Evaluate(ctx context.Context, water_readings []*model.WaterReading) error {
	rules, err := e.rules.ListEnabled(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, r := range water_readings {
		sampling_point, err := e.sampling_points.GetByID(ctx, r.SamplingPointID)
		if err != nil {
			// 采样点已注销或查询失败：跳过该读数，不阻断同批次其余读数的评估。
			// 旧处理单元切换期间，网关缓存包仍会携带已注销点位的读数到达，此处必须容错。
			if errors.Is(err, model.ErrNotFound) {
				continue
			}
			return err
		}
		// 已禁用（注销）的采样点不再触发合规评估。
		if !sampling_point.Enabled {
			continue
		}
		for _, rule := range rules {
			if rule.BallastTankID != sampling_point.BallastTankID || rule.ParamType != r.ParamType {
				continue
			}
			if !rule.Match(r.Value) {
				continue
			}
			triggered, err := e.sustained(ctx, sampling_point, rule, now)
			if err != nil {
				return err
			}
			if !triggered {
				continue
			}
			if err := e.openOrUpdate(ctx, rule, sampling_point, r, now); err != nil {
				return err
			}
		}
	}
	return nil
}

// sustained 判断超标是否持续满 duration_sec。
// 规则：取最近 duration 秒窗口内读数，若超标占比 ≥ 0.8 视为持续。
func (e *Engine) sustained(ctx context.Context, sampling_point *model.SamplingPoint, rule *model.ComplianceRule, now time.Time) (bool, error) {
	if e.reads == nil {
		return true, nil
	}
	from := now.Add(-time.Duration(rule.DurationSec) * time.Second)
	vals, err := e.reads.QueryValues(ctx, sampling_point.ID, rule.ParamType, from, now)
	if err != nil {
		return false, err
	}
	if len(vals) == 0 {
		return false, nil
	}
	bad := 0
	for _, v := range vals {
		if rule.Match(v) {
			bad++
		}
	}
	return float64(bad)/float64(len(vals)) >= 0.8, nil
}

// openOrUpdate 开新告警或复用已有告警（避免重复创建）。
func (e *Engine) openOrUpdate(ctx context.Context, rule *model.ComplianceRule, sampling_point *model.SamplingPoint, r *model.WaterReading, now time.Time) error {
	existing, err := e.compliance_compliance_alerts.OpenByRulePoint(ctx, rule.ID, sampling_point.ID)
	if err != nil {
		if err != model.ErrNotFound {
			return err
		}
		msg := buildMessage(rule, sampling_point, r)
		a := &model.ComplianceAlert{
			RuleID:   rule.ID,
			BallastTankID:   rule.BallastTankID,
			SamplingPointID:  sampling_point.ID,
			Level:    rule.Level,
			Message:  msg,
			Status:   model.ComplianceAlertOpen,
			OpenedAt: now,
		}
		return e.compliance_compliance_alerts.Create(ctx, a)
	}
	// 已有告警：warn 升级为 alarm。
	if existing.Level == model.ComplianceAlertWarn && rule.Level == model.ComplianceAlertAlarm {
		if err := e.compliance_compliance_alerts.SetResolved(ctx, existing.ID, now); err != nil {
			return err
		}
		a := &model.ComplianceAlert{
			RuleID:   rule.ID,
			BallastTankID:   rule.BallastTankID,
			SamplingPointID:  sampling_point.ID,
			Level:    model.ComplianceAlertAlarm,
			Message:  buildMessage(rule, sampling_point, r),
			Status:   model.ComplianceAlertOpen,
			OpenedAt: now,
		}
		return e.compliance_compliance_alerts.Create(ctx, a)
	}
	return nil
}

// buildMessage 构造告警消息。
func buildMessage(rule *model.ComplianceRule, sampling_point *model.SamplingPoint, r *model.WaterReading) string {
	unit := model.ParamUnits[rule.ParamType]
	return "参数 " + rule.ParamType + " 读数 " + util.Truncate(formatValue(r.Value), 16) +
		unit + " 命中规则 " + rule.Code
}

func formatValue(v float64) string {
	return trimZero(v)
}

func trimZero(v float64) string {
	return util.TrimZero(v)
}
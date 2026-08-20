package compliance_alerter

import (
	"context"
	"log"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
)

// Notifier 告警通知：当前实现为结构化日志输出，可扩展为推送/短信。
type Notifier struct {
	compliance_compliance_alerts store.ComplianceComplianceAlertStore
	debug  bool
}

// NewNotifier 创建通知器。
func NewNotifier(compliance_compliance_alerts store.ComplianceComplianceAlertStore, debug bool) *Notifier {
	return &Notifier{compliance_compliance_alerts: compliance_compliance_alerts, debug: debug}
}

// NotifyNew 通知新告警。
func (n *Notifier) NotifyNew(ctx context.Context, a *model.ComplianceAlert) {
	log.Printf("[ALERT][%s] tank=%d sampling_point=%d %s", a.Level, a.BallastTankID, a.SamplingPointID, a.Message)
}

// NotifyResolved 通知告警解除。
func (n *Notifier) NotifyResolved(ctx context.Context, a *model.ComplianceAlert) {
	log.Printf("[ALERT-RESOLVED] tank=%d sampling_point=%d %s", a.BallastTankID, a.SamplingPointID, a.Message)
}

// Digest 生成告警日报摘要文本。
func (n *Notifier) Digest(ctx context.Context, day time.Time) (string, error) {
	from := day.Truncate(24 * time.Hour)
	to := from.Add(24 * time.Hour)
	list, err := n.compliance_compliance_alerts.List(ctx, model.ComplianceComplianceAlertInput{From: from, To: to, Limit: 200})
	if err != nil {
		return "", err
	}
	byLevel := map[string]int{}
	for _, a := range list {
		byLevel[a.Level]++
	}
	return formatDigest(day, len(list), byLevel), nil
}

func formatDigest(day time.Time, total int, byLevel map[string]int) string {
	text := "告警日报 " + day.Format("2006-01-02") + " 共 " + itoa(total) + " 条"
	for _, lv := range []string{model.ComplianceAlertAlarm, model.ComplianceAlertWarn} {
		if n, ok := byLevel[lv]; ok {
			text += "；" + lv + " " + itoa(n)
		}
	}
	return text
}

func itoa(n int) string {
	return intToStr(n)
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
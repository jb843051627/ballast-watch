package report

import (
	"time"

	"ballast-watch/internal/util"
)

// DailySummary 单日汇总。
type DailySummary struct {
	Date          string  `json:"date"`
	WaterReadingCount  int     `json:"water_reading_count"`
	OutOfRange    int     `json:"out_of_range"`
	CompliancePct float64 `json:"compliance_pct"`
	OpenComplianceAlerts    int     `json:"open_compliance_compliance_alerts"`
}

// BuildDailySummary 由统计值构造单日汇总。
func BuildDailySummary(day time.Time, water_readingCount, outOfRange, openComplianceAlerts int) DailySummary {
	pct := 0.0
	if water_readingCount > 0 {
		pct = float64(water_readingCount-outOfRange) / float64(water_readingCount) * 100
	}
	return DailySummary{
		Date:          util.FormatDate(day),
		WaterReadingCount:  water_readingCount,
		OutOfRange:    outOfRange,
		CompliancePct: util.Round(pct, 1),
		OpenComplianceAlerts:    openComplianceAlerts,
	}
}

// TrendSummary 阶段趋势汇总。
type TrendSummary struct {
	From          string  `json:"from"`
	To            string  `json:"to"`
	Days          int     `json:"days"`
	AvgCompliance float64 `json:"avg_compliance"`
}

// BuildTrendSummary 汇总多日达标率。
func BuildTrendSummary(from, to time.Time, daily []DailySummary) TrendSummary {
	if len(daily) == 0 {
		return TrendSummary{
			From: util.FormatDate(from),
			To:   util.FormatDate(to),
			Days: 0,
		}
	}
	sum := 0.0
	for _, d := range daily {
		sum += d.CompliancePct
	}
	return TrendSummary{
		From:          util.FormatDate(from),
		To:            util.FormatDate(to),
		Days:          len(daily),
		AvgCompliance: util.Round(sum/float64(len(daily)), 1),
	}
}
package util

import "time"

// Loc 统一业务时区（Asia/Shanghai, UTC+8）。
var Loc = time.FixedZone("CST", 8*3600)

// Now 当前业务时间。
func Now() time.Time {
	return time.Now().In(Loc)
}

// ParseTime 解析 RFC3339 时间字符串；空串返回零值。
func ParseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(Loc), nil
}

// FormatTime 格式化时间为业务时区。
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(Loc).Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期（报表用）。
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(Loc).Format("2006-01-02")
}

// StartOfDay 当天零点。
func StartOfDay(t time.Time) time.Time {
	tt := t.In(Loc)
	return time.Date(tt.Year(), tt.Month(), tt.Day(), 0, 0, 0, 0, Loc)
}

// EndOfDay 当天最后一刻。
func EndOfDay(t time.Time) time.Time {
	return StartOfDay(t).AddDate(0, 0, 1).Add(-time.Second)
}

// DayWindow 返回 [startOfDay, endOfDay] 窗口。
func DayWindow(t time.Time) (time.Time, time.Time) {
	start := StartOfDay(t)
	return start, start.AddDate(0, 0, 1).Add(-time.Second)
}

// MonthWindow 返回当月窗口。
func MonthWindow(t time.Time) (time.Time, time.Time) {
	tt := t.In(Loc)
	start := time.Date(tt.Year(), tt.Month(), 1, 0, 0, 0, 0, Loc)
	return start, start.AddDate(0, 1, 0).Add(-time.Second)
}

// RollingWindow 最近 N 分钟窗口 [now-N, now]。
func RollingWindow(now time.Time, minutes int) (time.Time, time.Time) {
	return now.Add(-time.Duration(minutes) * time.Minute), now
}

// TruncateSecond 截断到秒（统一入库时间精度）。
func TruncateSecond(t time.Time) time.Time {
	return t.Truncate(time.Second)
}
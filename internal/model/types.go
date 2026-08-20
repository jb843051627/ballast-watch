package model

import "errors"

// 全局 sentinel 错误（service/store 层共用）。
var (
	ErrNotFound      = errors.New("记录不存在")
	ErrConflict      = errors.New("冲突")
	ErrInvalidInput  = errors.New("输入非法")
	ErrNameRequired  = errors.New("名称必填")
	ErrCodeRequired  = errors.New("编码必填")
	ErrInvalidGrade  = errors.New("洁净等级必须是 ISO5/ISO7/ISO8")
	ErrInvalidArea   = errors.New("面积必须大于 0")
	ErrStateConflict = errors.New("当前状态不允许该转移")
	ErrTreatmentCycleActive   = errors.New("房间已有进行中的批次")
	ErrDuplicateCode = errors.New("编码已存在")
)

// Page 分页参数。
type Page struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// Normalize 归一化分页参数。
func (p *Page) Normalize() {
	if p.Limit <= 0 || p.Limit > 500 {
		p.Limit = 100
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}

// TimeRange 时间范围查询参数。
type TimeRange struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

// OK 时间范围是否合法。
func (t *TimeRange) OK() bool {
	return t.From >= 0 && t.To >= t.From
}

// Sum 数值求和辅助。
func Sum(vals []float64) float64 {
	var s float64
	for _, v := range vals {
		s += v
	}
	return s
}

// Mean 均值。
func Mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	return Sum(vals) / float64(len(vals))
}
package util

import (
	"math"
	"sort"
	"strconv"
)

// MinMax 返回切片最小/最大值；空切片返回 0,0。
func MinMax(vals []float64) (min, max float64) {
	if len(vals) == 0 {
		return 0, 0
	}
	min, max = vals[0], vals[0]
	for _, v := range vals[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

// Mean 均值。
func Mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var s float64
	for _, v := range vals {
		s += v
	}
	return s / float64(len(vals))
}

// Stddev 总体标准差。
func Stddev(vals []float64) float64 {
	n := len(vals)
	if n == 0 {
		return 0
	}
	m := Mean(vals)
	var sum float64
	for _, v := range vals {
		d := v - m
		sum += d * d
	}
	return math.Sqrt(sum / float64(n))
}

// Percentile 百分位数（Nearest-Rank）。
func Percentile(vals []float64, p float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sorted := make([]float64, len(vals))
	copy(sorted, vals)
	sort.Float64s(sorted)
	rank := int(math.Ceil(p/100*float64(len(sorted)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(sorted) {
		rank = len(sorted) - 1
	}
	return sorted[rank]
}

// CountAbove 统计大于阈值的个数。
func CountAbove(vals []float64, threshold float64) int {
	n := 0
	for _, v := range vals {
		if v > threshold {
			n++
		}
	}
	return n
}

// Rate 达标率：在 [min,max] 区间内的占比（0-100）。
func Rate(vals []float64, min, max float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	ok := 0
	for _, v := range vals {
		if v >= min && v <= max {
			ok++
		}
	}
	return float64(ok) / float64(len(vals)) * 100
}

// Round 保留小数点后 n 位。
func Round(v float64, n int) float64 {
	p := math.Pow10(n)
	return math.Round(v*p) / p
}

// TrimZero 去除浮点字符串尾部多余 0。
func TrimZero(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
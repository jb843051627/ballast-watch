package util

import (
	"regexp"
	"strings"
)

// CodeRe 编码格式：字母开头，含字母数字下划线中划线。
var CodeRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{1,31}$`)

// SerialRe 传感器序列号格式：字母数字与中划线，8-32 位。
var SerialRe = regexp.MustCompile(`^[A-Za-z0-9-]{8,32}$`)

// ValidateCode 校验业务编码。
func ValidateCode(code string) bool {
	return CodeRe.MatchString(code)
}

// ValidateSerial 校验传感器序列号。
func ValidateSerial(serial string) bool {
	return SerialRe.MatchString(serial)
}

// Clean 清洗字符串（去首尾空白）。
func Clean(s string) string {
	return strings.TrimSpace(s)
}

// Truncate 截断字符串到 n 个 rune。
func Truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// IsEmpty 判断空白字符串。
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ContainsAny 判断是否包含任一子串。
func ContainsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
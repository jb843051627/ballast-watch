package store

import "strings"

// isUniqueViolation 判断是否 SQLite UNIQUE 约束冲突。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "constraint failed: UNIQUE")
}

// boolToInt 转换 bool 到 SQLite int。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
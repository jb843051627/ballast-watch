package handler

import (
	"net/http"
)

// web 静态资源由 router 的 FileServer 直接服务，无需额外 handler。
// 该文件保留占位说明，避免包内无文件时 Go 编译器告警。

var _ = http.MethodGet
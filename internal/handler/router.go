package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"ballast-watch/internal/model"
	"ballast-watch/internal/service"
)

// Handler 聚合全部 HTTP handler。
type Handler struct {
	services *service.Services
	webDir   string
}

// NewHandler 创建 HTTP handler 聚合。
func NewHandler(services *service.Services, webDir string) *Handler {
	return &Handler{services: services, webDir: webDir}
}

// Router 构建路由。
func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// 读数
	mux.HandleFunc("POST /api/v1/water_readings", h.gatewayIngest)
	mux.HandleFunc("GET /api/v1/water_readings/query", h.water_readingsQuery)
	mux.HandleFunc("GET /api/v1/water_readings/stats", h.water_readingsStats)
	mux.HandleFunc("GET /api/v1/water_readings/realtime", h.water_readingsRealtime)

	// 洁净区 / 房间 / 点位
	mux.HandleFunc("POST /api/v1/vessels", h.vesselCreate)
	mux.HandleFunc("GET /api/v1/vessels", h.vesselList)
	mux.HandleFunc("GET /api/v1/vessels/{id}", h.vesselGet)
	mux.HandleFunc("PUT /api/v1/vessels/{id}", h.vesselUpdate)
	mux.HandleFunc("GET /api/v1/tanks", h.tankList)
	mux.HandleFunc("POST /api/v1/tanks", h.tankCreate)
	mux.HandleFunc("GET /api/v1/tanks/{id}", h.tankGet)
	mux.HandleFunc("GET /api/v1/tanks/{id}/status", h.tankStatus)
	mux.HandleFunc("POST /api/v1/sampling_points", h.sampling_pointCreate)
	mux.HandleFunc("GET /api/v1/sampling_points", h.sampling_pointList)
	mux.HandleFunc("PUT /api/v1/sampling_points/{id}/toggle", h.sampling_pointToggle)

	// 传感器 / 校准
	mux.HandleFunc("POST /api/v1/sensors", h.sensorRegister)
	mux.HandleFunc("GET /api/v1/sensors", h.sensorList)
	mux.HandleFunc("GET /api/v1/sensors/{id}", h.sensorGet)
	mux.HandleFunc("POST /api/v1/sensors/{id}/calibrate", h.calibrationRecord)
	mux.HandleFunc("GET /api/v1/sensors/{id}/calibrations", h.calibrationList)

	// 告警规则 / 告警
	mux.HandleFunc("POST /api/v1/compliance_alert-rules", h.ruleCreate)
	mux.HandleFunc("GET /api/v1/compliance_alert-rules", h.ruleList)
	mux.HandleFunc("PUT /api/v1/compliance_alert-rules/{id}/toggle", h.ruleToggle)
	mux.HandleFunc("GET /api/v1/compliance_compliance_alerts", h.compliance_alertList)
	mux.HandleFunc("PUT /api/v1/compliance_compliance_alerts/{id}/ack", h.compliance_alertAck)
	mux.HandleFunc("PUT /api/v1/compliance_compliance_alerts/{id}/resolve", h.compliance_alertResolve)

	// 批次
	mux.HandleFunc("POST /api/v1/cyclees", h.cycleStart)
	mux.HandleFunc("PUT /api/v1/cyclees/{id}/complete", h.cycleComplete)
	mux.HandleFunc("PUT /api/v1/cyclees/{id}/abort", h.cycleAbort)
	mux.HandleFunc("GET /api/v1/cyclees", h.cycleList)

	// 报表 / 看板
	mux.HandleFunc("GET /api/v1/reports/trend", h.reportTrend)
	mux.HandleFunc("GET /api/v1/reports/summary", h.reportSummary)
	mux.HandleFunc("GET /api/v1/reports/export", h.reportExport)
	mux.HandleFunc("GET /api/v1/dashboard", h.dashboard)

	// 健康检查
	mux.HandleFunc("GET /healthz", h.health)

	// 前端
	fs := http.FileServer(http.Dir(h.webDir))
	mux.HandleFunc("GET /web/", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/web/", fs).ServeHTTP(w, r)
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/index.html", http.StatusFound)
	})

	return logMiddleware(mux)
}

// logMiddleware 请求日志。
func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[http] %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// writeJSON 写 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeErr 写错误响应。
func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

// pathID 解析路径参数为 int64。
func pathID(r *http.Request, name string) (int64, error) {
	v := r.PathValue(name)
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, model.ErrInvalidInput
	}
	return id, nil
}

// statusFromError 错误码映射。用 errors.Is 识别被包装的 sentinel，
// 避免 fmt.Errorf 包装后的业务错误丢失语义、错误地回落到 500。
func statusFromError(err error) int {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, model.ErrDuplicateCode) ||
		errors.Is(err, model.ErrConflict) ||
		errors.Is(err, model.ErrStateConflict) ||
		errors.Is(err, model.ErrTreatmentCycleActive):
		return http.StatusConflict
	case errors.Is(err, model.ErrInvalidInput) ||
		errors.Is(err, model.ErrNameRequired) ||
		errors.Is(err, model.ErrCodeRequired) ||
		errors.Is(err, model.ErrInvalidGrade) ||
		errors.Is(err, model.ErrInvalidArea) ||
		errors.Is(err, model.ErrBallastTankRequired) ||
		errors.Is(err, model.ErrVesselRequired) ||
		errors.Is(err, model.ErrInvalidKind) ||
		errors.Is(err, model.ErrInvalidPressureTarget) ||
		errors.Is(err, model.ErrPointRequired) ||
		errors.Is(err, model.ErrInvalidParamType) ||
		errors.Is(err, model.ErrThresholdInverted) ||
		errors.Is(err, model.ErrInvalidDuration) ||
		errors.Is(err, model.ErrSerialRequired) ||
		errors.Is(err, model.ErrVendorRequired) ||
		errors.Is(err, model.ErrInvalidBattery) ||
		errors.Is(err, model.ErrInvalidLevel) ||
		errors.Is(err, model.ErrInvalidComplianceAlertStatus) ||
		errors.Is(err, model.ErrInvalidOp) ||
		errors.Is(err, model.ErrSensorRequired) ||
		errors.Is(err, model.ErrStandardRequired) ||
		errors.Is(err, model.ErrInvalidResult) ||
		errors.Is(err, model.ErrOperatorRequired) ||
		errors.Is(err, model.ErrInvalidDueDate) ||
		errors.Is(err, model.ErrProductRequired) ||
		errors.Is(err, model.ErrInvalidTreatmentCycleStatus) ||
		errors.Is(err, model.ErrInvalidState):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
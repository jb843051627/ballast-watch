package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/store"
	"ballast-watch/internal/util"
)

// gatewayIngest 转发到采集网关（由 router 挂载 gateway 处理）。
func (h *Handler) gatewayIngest(w http.ResponseWriter, r *http.Request) {
	var cycle model.WaterWaterReadingTreatmentCycle
	if err := json.NewDecoder(r.Body).Decode(&cycle); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	n, err := h.services.WaterReadings.Ingest(r.Context(), &cycle)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accepted": n})
}

func (h *Handler) water_readingsQuery(w http.ResponseWriter, r *http.Request) {
	q := store.WaterReadingQuery{Limit: 200}
	if v := r.URL.Query().Get("sampling_sampling_point_id"); v != "" {
		q.SamplingPointID, _ = strconv.ParseInt(v, 10, 64)
	}
	if v := r.URL.Query().Get("tank_id"); v != "" {
		q.BallastTankID, _ = strconv.ParseInt(v, 10, 64)
	}
	q.ParamType = r.URL.Query().Get("param_type")
	if v := r.URL.Query().Get("from"); v != "" {
		q.From, _ = util.ParseTime(v)
	}
	if v := r.URL.Query().Get("to"); v != "" {
		q.To, _ = util.ParseTime(v)
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			q.Limit = n
		}
	}
	list, err := h.services.WaterReadings.Query(r.Context(), q)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) water_readingsStats(w http.ResponseWriter, r *http.Request) {
	paramType := r.URL.Query().Get("param_type")
	if !model.ParamTypes[paramType] {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidParamType)
		return
	}
	from, to := parseRange(r)
	st, err := h.services.WaterReadings.Stats(r.Context(), paramType, from, to)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (h *Handler) water_readingsRealtime(w http.ResponseWriter, r *http.Request) {
	var tankID int64
	if v := r.URL.Query().Get("tank_id"); v != "" {
		tankID, _ = strconv.ParseInt(v, 10, 64)
	}
	snap, err := h.services.WaterReadings.RealtimeSnapshot(r.Context(), tankID)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

// parseRange 解析 from/to 时间范围（默认最近 24h）。
func parseRange(r *http.Request) (time.Time, time.Time) {
	now := time.Now()
	from := now.Add(-24 * time.Hour)
	to := now
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := util.ParseTime(v); err == nil && !t.IsZero() {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := util.ParseTime(v); err == nil && !t.IsZero() {
			to = t
		}
	}
	return from, to
}
package handler

import (
	"net/http"
	"strconv"
	"time"

	"ballast-watch/internal/model"
	"ballast-watch/internal/util"
)

func (h *Handler) reportTrend(w http.ResponseWriter, r *http.Request) {
	tankID, err := strconv.ParseInt(r.URL.Query().Get("tank_id"), 10, 64)
	if err != nil || tankID <= 0 {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	paramType := r.URL.Query().Get("param_type")
	from, to := parseRange(r)
	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	res, err := h.services.Reports.Trend(r.Context(), tankID, paramType, from, to, limit)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) reportSummary(w http.ResponseWriter, r *http.Request) {
	from, to := parseRange(r)
	res, err := h.services.Reports.Summary(r.Context(), from, to)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *Handler) reportExport(w http.ResponseWriter, r *http.Request) {
	tankID, err := strconv.ParseInt(r.URL.Query().Get("tank_id"), 10, 64)
	if err != nil || tankID <= 0 {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	from, to := parseRange(r)
	data, err := h.services.Export.ExportWaterReadingsCSV(r.Context(), tankID, from, to)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=water_readings.csv")
	_, _ = w.Write(data)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	snap, err := h.services.Dashboard.Snapshot(r.Context())
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
		"tz":     util.Loc.String(),
	})
}
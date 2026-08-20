package handler

import (
	"net/http"
	"strconv"

	"ballast-watch/internal/model"
)

func (h *Handler) compliance_alertList(w http.ResponseWriter, r *http.Request) {
	in := model.ComplianceComplianceAlertInput{Limit: 100}
	if v := r.URL.Query().Get("tank_id"); v != "" {
		in.BallastTankID, _ = strconv.ParseInt(v, 10, 64)
	}
	in.Level = r.URL.Query().Get("level")
	in.Status = r.URL.Query().Get("status")
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			in.Limit = n
		}
	}
	list, err := h.services.ComplianceAlerts.List(r.Context(), in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) compliance_alertAck(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	a, err := h.services.ComplianceAlerts.Acknowledge(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (h *Handler) compliance_alertResolve(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	a, err := h.services.ComplianceAlerts.Resolve(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}
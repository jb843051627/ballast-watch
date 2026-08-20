package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ballast-watch/internal/model"
)

func (h *Handler) sensorRegister(w http.ResponseWriter, r *http.Request) {
	var in model.SensorInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	s, err := h.services.Sensors.Register(r.Context(), &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) sensorList(w http.ResponseWriter, r *http.Request) {
	page := model.Page{Limit: 100, Offset: 0}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Limit = n
		}
	}
	if v := r.URL.Query().Get("sampling_sampling_point_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			list, err := h.services.Sensors.ListByPoint(r.Context(), n)
			if err != nil {
				writeErr(w, statusFromError(err), err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": list})
			return
		}
	}
	list, err := h.services.Sensors.List(r.Context(), page)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) sensorGet(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	s, err := h.services.Sensors.Get(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}
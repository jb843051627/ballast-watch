package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"ballast-watch/internal/model"
)

func (h *Handler) calibrationRecord(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	var in struct {
		PerformedAt time.Time `json:"performed_at"`
		DueAt       time.Time `json:"due_at"`
		Standard    string    `json:"standard"`
		Result      string    `json:"result"`
		Operator    string    `json:"operator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	cal, err := h.services.Calibrations.Record(r.Context(), &model.CalibrationInput{
		SensorID:    id,
		PerformedAt: in.PerformedAt,
		DueAt:       in.DueAt,
		Standard:    in.Standard,
		Result:      in.Result,
		Operator:    in.Operator,
	})
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, cal)
}

func (h *Handler) calibrationList(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.services.Calibrations.List(r.Context(), id, limit)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}
package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ballast-watch/internal/model"
)

func (h *Handler) ruleCreate(w http.ResponseWriter, r *http.Request) {
	var in model.ComplianceRuleInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	rule, err := h.services.Rules.Create(r.Context(), &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (h *Handler) ruleList(w http.ResponseWriter, r *http.Request) {
	var tankID int64
	if v := r.URL.Query().Get("tank_id"); v != "" {
		tankID, _ = strconv.ParseInt(v, 10, 64)
	}
	list, err := h.services.Rules.List(r.Context(), tankID)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) ruleToggle(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	var in struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	rule, err := h.services.Rules.Toggle(r.Context(), id, in.Enabled)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, rule)
}
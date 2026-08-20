package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ballast-watch/internal/model"
)

func (h *Handler) vesselCreate(w http.ResponseWriter, r *http.Request) {
	var in model.VesselInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	c, err := h.services.Vessels.Create(r.Context(), &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) vesselList(w http.ResponseWriter, r *http.Request) {
	page := model.Page{Limit: 100, Offset: 0}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Offset = n
		}
	}
	list, err := h.services.Vessels.List(r.Context(), page)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) vesselGet(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	c, err := h.services.Vessels.Get(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) vesselUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	var in model.VesselInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	c, err := h.services.Vessels.Update(r.Context(), id, &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}
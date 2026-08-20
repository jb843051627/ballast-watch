package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ballast-watch/internal/model"
)

func (h *Handler) tankCreate(w http.ResponseWriter, r *http.Request) {
	var in model.TankInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeErr(w, http.StatusBadRequest, model.ErrInvalidInput)
		return
	}
	tank, err := h.services.BallastTanks.Create(r.Context(), &in)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, tank)
}

func (h *Handler) tankList(w http.ResponseWriter, r *http.Request) {
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
	if v := r.URL.Query().Get("vessel_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			list, err := h.services.BallastTanks.ListByVessel(r.Context(), n)
			if err != nil {
				writeErr(w, statusFromError(err), err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": list})
			return
		}
	}
	list, err := h.services.BallastTanks.List(r.Context(), page)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func (h *Handler) tankGet(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	tank, err := h.services.BallastTanks.Get(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, tank)
}

func (h *Handler) tankStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	st, err := h.services.BallastTanks.GetStatus(r.Context(), id)
	if err != nil {
		writeErr(w, statusFromError(err), err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}
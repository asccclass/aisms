package handlers

import (
	"isms-privilege/internal/models"
	"net/http"
	"strings"
)

func (h *Handler) ListApplicationChangeRequests(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.ListApplicationChangeRequestsByCreator(GetUserEmail(r))
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, rows)
}

func (h *Handler) GetApplicationChangeRequest(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/application-change-requests/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	row, err := h.DB.GetApplicationChangeRequestByCreator(id, GetUserEmail(r))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, 200, row)
}

func (h *Handler) CreateApplicationChangeRequest(w http.ResponseWriter, r *http.Request) {
	var req models.ApplicationChangeRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	req.Creator = GetUserEmail(r)
	if strings.TrimSpace(req.Status) == "" {
		req.Status = "active"
	}
	id, err := h.DB.CreateApplicationChangeRequest(&req)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	req.ID = int(id)
	writeJSON(w, 201, req)
}

func (h *Handler) UpdateApplicationChangeRequest(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/application-change-requests/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	var req models.ApplicationChangeRequest
	if err := readJSON(r, &req); err != nil {
		writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	req.ID = id
	existing, _ := h.DB.GetApplicationChangeRequestByCreator(id, GetUserEmail(r))
	if existing == nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	req.Creator = existing.Creator
	if err := h.DB.UpdateApplicationChangeRequest(&req); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, req)
}

func (h *Handler) DeleteApplicationChangeRequest(w http.ResponseWriter, r *http.Request) {
	id, err := idFromPath(r, "/api/application-change-requests/")
	if err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return
	}
	existing, _ := h.DB.GetApplicationChangeRequestByCreator(id, GetUserEmail(r))
	if existing == nil {
		writeJSON(w, 404, map[string]string{"error": "not found"})
		return
	}
	if err := h.DB.DeleteApplicationChangeRequestByCreator(id, GetUserEmail(r)); err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]string{"message": "deleted"})
}

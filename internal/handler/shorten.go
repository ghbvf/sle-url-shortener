package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/user/url-shortener/internal/service"
)

type ShortenHandler struct {
	svc *service.ShortenerService
}

func NewShortenHandler(svc *service.ShortenerService) *ShortenHandler {
	return &ShortenHandler{svc: svc}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

func (h *ShortenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url is required"})
		return
	}

	result, err := h.svc.CreateShortLink(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid URL format"})
			return
		}
		slog.Error("create_short_link", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{
		Code:     result.Code,
		ShortURL: result.ShortURL,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

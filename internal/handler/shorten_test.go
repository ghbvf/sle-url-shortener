package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/url-shortener/internal/repository"
	"github.com/user/url-shortener/internal/service"
)

type mockRepo struct {
	createFn func(ctx context.Context, link *repository.Link) error
}

func (m *mockRepo) Create(ctx context.Context, link *repository.Link) error {
	return m.createFn(ctx, link)
}

func newTestHandler() *ShortenHandler {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *repository.Link) error {
			return nil
		},
	}
	svc := service.NewShortenerService(repo, "http://localhost:8080")
	return NewShortenHandler(svc)
}

func TestCreate_Success(t *testing.T) {
	h := newTestHandler()
	body := strings.NewReader(`{"url":"https://example.com"}`)
	req := httptest.NewRequest("POST", "/api/shorten", body)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
	resp := w.Body.String()
	if !strings.Contains(resp, `"code"`) || !strings.Contains(resp, `"short_url"`) {
		t.Errorf("response missing expected fields: %s", resp)
	}
}

func TestCreate_EmptyURL(t *testing.T) {
	h := newTestHandler()
	body := strings.NewReader(`{"url":""}`)
	req := httptest.NewRequest("POST", "/api/shorten", body)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreate_InvalidJSON(t *testing.T) {
	h := newTestHandler()
	body := strings.NewReader(`not json`)
	req := httptest.NewRequest("POST", "/api/shorten", body)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestCreate_InvalidURLFormat(t *testing.T) {
	h := newTestHandler()
	body := strings.NewReader(`{"url":"not-a-url"}`)
	req := httptest.NewRequest("POST", "/api/shorten", body)
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

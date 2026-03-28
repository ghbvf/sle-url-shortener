package service

import (
	"context"
	"testing"

	"github.com/user/url-shortener/internal/repository"
)

type mockRepo struct {
	createFn func(ctx context.Context, link *repository.Link) error
}

func (m *mockRepo) Create(ctx context.Context, link *repository.Link) error {
	return m.createFn(ctx, link)
}

func TestCreateShortLink_Success(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *repository.Link) error {
			return nil
		},
	}
	svc := NewShortenerService(repo, "http://localhost:8080")

	result, err := svc.CreateShortLink(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Code) != codeLength {
		t.Errorf("code length = %d, want %d", len(result.Code), codeLength)
	}
	if result.ShortURL != "http://localhost:8080/"+result.Code {
		t.Errorf("short_url = %q, want prefix http://localhost:8080/", result.ShortURL)
	}
}

func TestCreateShortLink_InvalidURL(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *repository.Link) error {
			return nil
		},
	}
	svc := NewShortenerService(repo, "http://localhost:8080")

	cases := []string{
		"not-a-url",
		"ftp://example.com",
		"",
	}
	for _, tc := range cases {
		_, err := svc.CreateShortLink(context.Background(), tc)
		if err == nil {
			t.Errorf("expected error for URL %q, got nil", tc)
		}
	}
}

func TestCreateShortLink_RetryOnDuplicate(t *testing.T) {
	calls := 0
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *repository.Link) error {
			calls++
			if calls <= 2 {
				return repository.ErrCodeExists
			}
			return nil
		},
	}
	svc := NewShortenerService(repo, "http://localhost:8080")

	result, err := svc.CreateShortLink(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("repo.Create called %d times, want 3", calls)
	}
	if len(result.Code) != codeLength {
		t.Errorf("code length = %d, want %d", len(result.Code), codeLength)
	}
}

func TestCreateShortLink_ExhaustedRetries(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *repository.Link) error {
			return repository.ErrCodeExists
		},
	}
	svc := NewShortenerService(repo, "http://localhost:8080")

	_, err := svc.CreateShortLink(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected error after exhausted retries, got nil")
	}
}

func TestValidateURL(t *testing.T) {
	valid := []string{
		"https://example.com",
		"http://example.com/path?q=1",
		"https://sub.domain.com:8080/path",
	}
	for _, u := range valid {
		if err := validateURL(u); err != nil {
			t.Errorf("validateURL(%q) = %v, want nil", u, err)
		}
	}

	invalid := []string{
		"not-a-url",
		"ftp://example.com",
		"://missing-scheme",
	}
	for _, u := range invalid {
		if err := validateURL(u); err == nil {
			t.Errorf("validateURL(%q) = nil, want error", u)
		}
	}
}

func TestGenerateCode(t *testing.T) {
	code, err := generateCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != codeLength {
		t.Errorf("code length = %d, want %d", len(code), codeLength)
	}
	for _, c := range code {
		found := false
		for _, ch := range charset {
			if c == ch {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("code contains invalid char %q", c)
		}
	}
}

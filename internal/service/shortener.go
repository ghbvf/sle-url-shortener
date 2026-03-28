package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/user/url-shortener/internal/repository"
)

const (
	codeLength = 6
	charset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxRetries = 3
)

var ErrInvalidURL = errors.New("invalid URL")

type ShortenerService struct {
	repo    repository.LinkRepository
	baseURL string
}

func NewShortenerService(repo repository.LinkRepository, baseURL string) *ShortenerService {
	return &ShortenerService{repo: repo, baseURL: baseURL}
}

type CreateResult struct {
	Code     string
	ShortURL string
}

func (s *ShortenerService) CreateShortLink(ctx context.Context, rawURL string) (*CreateResult, error) {
	if err := validateURL(rawURL); err != nil {
		return nil, ErrInvalidURL
	}

	for i := 0; i < maxRetries; i++ {
		code, err := generateCode()
		if err != nil {
			return nil, fmt.Errorf("generate code: %w", err)
		}

		link := &repository.Link{
			Code: code,
			URL:  rawURL,
		}

		err = s.repo.Create(ctx, link)
		if err == nil {
			return &CreateResult{
				Code:     code,
				ShortURL: s.baseURL + "/" + code,
			}, nil
		}
		if !errors.Is(err, repository.ErrCodeExists) {
			return nil, fmt.Errorf("create link: %w", err)
		}
	}

	return nil, fmt.Errorf("failed to generate unique code after %d retries", maxRetries)
}

func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("empty host")
	}
	return nil
}

func generateCode() (string, error) {
	b := make([]byte, codeLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

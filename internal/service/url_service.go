package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"URL-Shotener/internal/cache"
	"URL-Shotener/internal/config"
	"URL-Shotener/internal/model"
	"URL-Shotener/internal/repository"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type URLService interface {
	Shorten(ctx context.Context, originalURL string, expireAt *time.Time) (string, error)
	Resolve(ctx context.Context, shortCode string) (string, error)
}

type urlService struct {
	repo   repository.URLRepository
	cache  *cache.RedisCache
	config *config.Config
}

func NewURLService(repo repository.URLRepository, cache *cache.RedisCache, cfg *config.Config) URLService {
	return &urlService{
		repo:   repo,
		cache:  cache,
		config: cfg,
	}
}

// generateShortCode creates a random base62 string of the given length.
func generateShortCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	result := make([]byte, length)
	max := big.NewInt(int64(len(base62Chars)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = base62Chars[idx.Int64()]
	}
	return string(result), nil
}

func isValidURL(rawURL string) bool {
	return strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")
}

func (s *urlService) Shorten(ctx context.Context, originalURL string, expireAt *time.Time) (string, error) {
	// ۱. اعتبارسنجی URL
	if !isValidURL(originalURL) {
		return "", errors.New("invalid URL: must start with http:// or https://")
	}

	for attempt := 0; attempt < 5; attempt++ {
		shortCode, err := generateShortCode(6)
		if err != nil {
			return "", fmt.Errorf("failed to generate short code: %w", err)
		}

		urlModel := &model.URL{
			OriginalURL: originalURL,
			ShortCode:   shortCode,
			ExpiresAt:   expireAt,
		}

		err = s.repo.Save(ctx, urlModel)
		if err == nil {

			_ = s.cache.Set(ctx, shortCode, originalURL)
			fullShortURL := fmt.Sprintf("%s/%s", s.config.BaseURL, shortCode)
			return fullShortURL, nil
		}

		if strings.Contains(err.Error(), "duplicate key") {
			continue
		}

		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	return "", errors.New("could not generate a unique short code after multiple attempts")
}

func (s *urlService) Resolve(ctx context.Context, shortCode string) (string, error) {

	if cached, err := s.cache.Get(ctx, shortCode); err == nil && cached != "" {
		return cached, nil
	}

	urlModel, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if urlModel == nil {
		return "", errors.New("short code not found")
	}

	if urlModel.ExpiresAt != nil && urlModel.ExpiresAt.Before(time.Now()) {
		return "", errors.New("shortened link has expired")
	}

	_ = s.cache.Set(ctx, shortCode, urlModel.OriginalURL)

	return urlModel.OriginalURL, nil
}

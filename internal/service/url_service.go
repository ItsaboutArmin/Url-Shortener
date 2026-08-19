package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"URL-Shotener/internal/cache"
	"URL-Shotener/internal/config"
	"URL-Shotener/internal/model"
	"URL-Shotener/internal/repository"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var urlIDCounter uint64 = 1000000

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

func toBase62(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}
	var result []byte
	for num > 0 {
		remainder := num % 62
		result = append([]byte{base62Chars[remainder]}, result...)
		num /= 62
	}
	return string(result)
}

func isValidURL(rawURL string) bool {
	return strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://")
}

func (s *urlService) Shorten(ctx context.Context, originalURL string, expireAt *time.Time) (string, error) {

	if !isValidURL(originalURL) {
		return "", errors.New("invalid URL: must start with http:// or https://")
	}

	id := atomic.AddUint64(&urlIDCounter, 1)
	shortCode := toBase62(id)

	urlModel := &model.URL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		ExpiresAt:   expireAt,
	}

	err := s.repo.Save(ctx, urlModel)
	if err != nil {
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	_ = s.cache.Set(ctx, shortCode, originalURL)

	fullShortURL := fmt.Sprintf("%s/%s", s.config.BaseURL, shortCode)
	return fullShortURL, nil
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

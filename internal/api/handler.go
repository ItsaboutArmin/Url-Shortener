package api

import (
	"net/http"

	"time"

	"URL-Shotener/internal/service"

	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	service service.URLService
}

func NewURLHandler(service service.URLService) *URLHandler {
	return &URLHandler{service: service}
}

type ShortenRequest struct {
	URL       string `json:"url" binding:"required,url"`
	ExpiresAt string `json:"expires_at,omitempty"` // فرمت RFC3339
}

func (h *URLHandler) Shorten(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expires_at format, use RFC3339"})
			return
		}
		expiresAt = &t
	}

	shortURL, err := h.service.Shorten(c.Request.Context(), req.URL, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})
}

func (h *URLHandler) Redirect(c *gin.Context) {
	shortCode := c.Param("code")
	originalURL, err := h.service.Resolve(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusMovedPermanently, originalURL)
}

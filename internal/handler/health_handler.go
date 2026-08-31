package handler

import (
	"context"
	"net/http"
	"time"
	"toll_plaza/internal/cache"
	"toll_plaza/internal/repository/postgres"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db    *postgres.DB
	cache cache.CacheService
}

func NewHealthHandler(db *postgres.DB, cache cache.CacheService) *HealthHandler {
	return &HealthHandler{db: db, cache: cache}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "healthy"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "unreachable"
	}

	redisStatus := "healthy"
	if h.cache != nil {
		if err := h.cache.Ping(ctx); err != nil {
			redisStatus = "degraded/unreachable"
		}
	} else {
		redisStatus = "disabled"
	}

	status := http.StatusOK
	if dbStatus != "healthy" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":    "up",
		"timestamp": time.Now().Format(time.RFC3339),
		"database":  dbStatus,
		"redis":     redisStatus,
	})
}

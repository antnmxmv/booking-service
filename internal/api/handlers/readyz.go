package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReadinessMonitor interface {
	Load() bool // is ready
}

func NewReadyzHandler(isReady ReadinessMonitor) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if isReady == nil || !isReady.Load() {
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

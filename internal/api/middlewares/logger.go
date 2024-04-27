package middlewares

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(ctx *gin.Context) {
	startTime := time.Now()

	ctx.Next()

	r := ctx.Request

	level := "INF"
	if ctx.Writer.Status() >= 500 {
		level = "ERR"
	}
	log.Printf("[%s] method=%s path=%s code=%d duration=%s", level, r.Method, r.URL.Path, ctx.Writer.Status(), time.Since(startTime).String())
}

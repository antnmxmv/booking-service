package middlewares

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func Logger() func(http.Handler) http.Handler {
	f := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			h.ServeHTTP(rw, r)

			level := "INF"
			if rw.Status() >= 500 {
				level = "ERR"
			}
			log.Printf("[%s] method=%s path=%s code=%d duration_ms=%0.6f", level, r.Method, r.URL.Path, rw.Status(), float64(time.Since(startTime).Nanoseconds())/100000)

		}
		return http.HandlerFunc(fn)
	}
	return f
}

package handlers

import (
	"net/http"
)

type ReadinessMonitor interface {
	Load() bool // is ready
}

func NewReadyzHandler(isReady ReadinessMonitor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isReady == nil || !isReady.Load() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

package handlers

import (
	"net/http"
	"sync/atomic"
)

func NewReadyzHandler(isReady *atomic.Bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isReady == nil || !isReady.Load() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

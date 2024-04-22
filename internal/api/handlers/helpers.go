package handlers

import (
	"fmt"
	"net/http"
	"time"
)

type httpError struct {
	code int
	text string
}

func (err httpError) Error() string {
	return fmt.Sprintf("code: %d, text: %s", err.code, err.text)
}

const TimeLayout = time.RFC3339

func errorJSON(w http.ResponseWriter, error string, code int) {
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, error)
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func toDay(timestamp time.Time) time.Time {
	dayRounded := timestamp.UTC().Round(time.Hour * 24).Day()
	return time.Date(timestamp.Year(), timestamp.Month(), dayRounded, 0, 0, 0, 0, time.UTC)
}

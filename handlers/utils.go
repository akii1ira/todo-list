package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

func JSONWrite(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func ParseDate(s, layout string) (time.Time, error) {
	return time.Parse(layout, s)
}

func DateOnly(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

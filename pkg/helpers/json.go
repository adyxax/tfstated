package helpers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func Decode(r *http.Request, data any) error {
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}
	return nil
}

func Encode(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode json", "err", err)
		return fmt.Errorf("failed to encode json: %w", err)
	}
	return nil
}

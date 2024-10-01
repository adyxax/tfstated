package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func decode(r *http.Request, data any) error {
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}
	return nil
}

func encode(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode json", "err", err)
		return fmt.Errorf("failed to encode json: %w", err)
	}
	return nil
}

func errorResponse(w http.ResponseWriter, status int, err error) error {
	type errorResponse struct {
		Msg    string `json:"msg"`
		Status int    `json:"status"`
	}
	return encode(w, status, &errorResponse{
		Msg:    fmt.Sprintf("%+v", err),
		Status: status,
	})
}

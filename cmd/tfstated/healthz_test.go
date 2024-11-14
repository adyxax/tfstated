package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestHealthz(t *testing.T) {
	runHTTPRequest("GET", false, &url.URL{Path: "/healthz"}, nil, func(r *http.Response, err error) {
		if err != nil {
			t.Fatalf("failed healthcheck with error: %+v", err)
		} else if r.StatusCode != http.StatusOK {
			t.Fatalf("healthcheck should succeed, got %s", http.StatusText(r.StatusCode))
		}
	})
}

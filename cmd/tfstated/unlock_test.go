package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestUnlock(t *testing.T) {
	tests := []struct {
		method string
		auth   bool
		uri    url.URL
		body   io.Reader
		expect string
		status int
		msg    string
	}{
		{"UNLOCK", false, url.URL{Path: "/"}, nil, "", http.StatusUnauthorized, "/"},
		{"UNLOCK", true, url.URL{Path: "/"}, nil, "", http.StatusBadRequest, "/"},
		{"UNLOCK", true, url.URL{Path: "/non_existent_lock"}, nil, "", http.StatusBadRequest, "no lock data on non existent state"},
		{"UNLOCK", true, url.URL{Path: "/non_existent_lock"}, strings.NewReader("{\"ID\":\"00000000-0000-0000-0000-000000000000\"}"), "", http.StatusConflict, "valid lock data on non existent state"},
		{"LOCK", true, url.URL{Path: "/test_unlock"}, strings.NewReader("{\"ID\":\"00000000-0000-0000-0000-000000000000\"}"), "", http.StatusOK, "valid lock data on non existent state should create it empty"},
		{"UNLOCK", true, url.URL{Path: "/test_unlock"}, strings.NewReader("{\"ID\":\"FFFFFFFF-FFFF-FFFF-FFFF-FFFFFFFFFFFF\"}"), "", http.StatusConflict, "valid but wrong lock data on a locked state"},
		{"UNLOCK", true, url.URL{Path: "/test_unlock"}, strings.NewReader("{\"ID\":\"00000000-0000-0000-0000-000000000000\"}"), "", http.StatusOK, "valid and correct lock data on a locked state"},
		{"UNLOCK", true, url.URL{Path: "/test_unlock"}, strings.NewReader("{\"ID\":\"00000000-0000-0000-0000-000000000000\"}"), "", http.StatusConflict, "valid and correct lock data on a now unlocked state"},
	}
	for _, tt := range tests {
		runHTTPRequest(tt.method, tt.auth, &tt.uri, tt.body, func(r *http.Response, err error) {
			if err != nil {
				t.Fatalf("failed %s with error: %+v", tt.method, err)
			} else if r.StatusCode != tt.status {
				t.Fatalf("%s %s should %s, got %s", tt.method, tt.msg, http.StatusText(tt.status), http.StatusText(r.StatusCode))
			} else if tt.expect != "" {
				if body, err := io.ReadAll(r.Body); err != nil {
					t.Fatalf("failed to read body with error: %+v", err)
				} else if string(body) != tt.expect {
					t.Fatalf("%s should have returned \"%s\", got %s", tt.method, tt.expect, string(body))
				}
			}
		})
	}
}

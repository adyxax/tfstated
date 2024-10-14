package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestLock(t *testing.T) {
	tests := []struct {
		method string
		uri    url.URL
		body   io.Reader
		expect string
		status int
		msg    string
	}{
		{"LOCK", url.URL{Path: "/"}, nil, "", http.StatusBadRequest, "/"},
		{"LOCK", url.URL{Path: "/non_existent_lock"}, nil, "", http.StatusBadRequest, "no lock data on non existent state"},
		{"LOCK", url.URL{Path: "/non_existent_lock"}, strings.NewReader("{}"), "", http.StatusBadRequest, "invalid lock data on non existent state"},
		{"LOCK", url.URL{Path: "/test_lock"}, strings.NewReader("{\"ID\":\"00000000-0000-0000-0000-000000000000\"}"), "", http.StatusOK, "valid lock data on non existent state should create it empty"},
		{"GET", url.URL{Path: "/test_lock"}, nil, "", http.StatusOK, "/test_lock"},
		{"LOCK", url.URL{Path: "/test_lock"}, strings.NewReader("{\"ID\":\"\"}"), "", http.StatusBadRequest, "invalid lock data on already locked state"},
		{"LOCK", url.URL{Path: "/test_lock"}, strings.NewReader("{\"ID\":\"00000000-0000-0000-0000-000000000000\"}"), "", http.StatusConflict, "valid lock data on already locked state"},
		{"POST", url.URL{Path: "/test_lock", RawQuery: "ID=00000000-0000-0000-0000-000000000000"}, strings.NewReader("the_test_lock"), "", http.StatusOK, "/test_lock"},
		{"GET", url.URL{Path: "/test_lock"}, nil, "the_test_lock", http.StatusOK, "/test_lock"},
	}
	for _, tt := range tests {
		runHTTPRequest(tt.method, &tt.uri, tt.body, func(r *http.Response, err error) {
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

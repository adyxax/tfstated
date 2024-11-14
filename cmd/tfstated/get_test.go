package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		method string
		auth   bool
		uri    url.URL
		body   io.Reader
		expect string
		status int
		msg    string
	}{
		{"GET", false, url.URL{Path: "/"}, nil, "", http.StatusUnauthorized, "/"},
		{"GET", true, url.URL{Path: "/"}, nil, "", http.StatusBadRequest, "/"},
		{"GET", true, url.URL{Path: "/non_existent_get"}, strings.NewReader(""), "", http.StatusOK, "non existent"},
		{"POST", true, url.URL{Path: "/test_get"}, strings.NewReader("the_test_get"), "", http.StatusOK, "/test_get"},
		{"GET", true, url.URL{Path: "/test_get"}, nil, "the_test_get", http.StatusOK, "/test_get"},
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

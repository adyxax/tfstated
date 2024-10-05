package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestDelete(t *testing.T) {
	tests := []struct {
		method string
		uri    *url.URL
		body   io.Reader
		expect string
		status int
		msg    string
	}{
		{"DELETE", &url.URL{Path: "/"}, nil, "", http.StatusBadRequest, "/"},
		{"DELETE", &url.URL{Path: "/non_existent_delete"}, nil, "", http.StatusNotFound, "non existent"},
		{"POST", &url.URL{Path: "/test_delete"}, strings.NewReader("the_test_delete"), "", http.StatusOK, "/test_delete"},
		{"DELETE", &url.URL{Path: "/test_delete"}, nil, "", http.StatusOK, "/test_delete"},
		{"DELETE", &url.URL{Path: "/test_delete"}, nil, "", http.StatusNotFound, "/test_delete"},
	}
	for _, tt := range tests {
		runHTTPRequest(tt.method, tt.uri, tt.body, func(r *http.Response, err error) {
			if err != nil {
				t.Fatalf("failed %s with error: %+v", tt.method, err)
			} else if r.StatusCode != tt.status {
				t.Fatalf("%s %s should %s, got %s", tt.method, tt.msg, http.StatusText(tt.status), http.StatusText(r.StatusCode))
			} else if tt.expect != "" {
				if body, err := io.ReadAll(r.Body); err != nil {
					t.Fatalf("failed to read body with error: %+v", err)
				} else if strings.Compare(string(body), tt.expect) != 0 {
					t.Fatalf("%s should have returned \"%s\", got %s", tt.method, tt.expect, string(body))
				}
			}
		})
	}
}

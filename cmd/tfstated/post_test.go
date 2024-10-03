package main

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestPost(t *testing.T) {
	tests := []struct {
		method string
		uri    *url.URL
		body   io.Reader
		expect string
		status int
		msg    string
	}{
		{"POST", &url.URL{Path: "/"}, nil, "", http.StatusBadRequest, "/"},
		{"POST", &url.URL{Path: "/test_post"}, nil, "", http.StatusBadRequest, "without a body"},
		{"POST", &url.URL{Path: "/test_post"}, strings.NewReader("the_test_post"), "", http.StatusOK, "without lock ID in query string"},
		{"GET", &url.URL{Path: "/test_post"}, nil, "the_test_post", http.StatusOK, "/test_post"},
		{"POST", &url.URL{Path: "/test_post", RawQuery: "ID=test_post"}, strings.NewReader("the_test_post2"), "", http.StatusConflict, "with a lock ID on an unlocked state"},
		{"GET", &url.URL{Path: "/test_post"}, nil, "the_test_post", http.StatusOK, "/test_post"},
		{"LOCK", &url.URL{Path: "/test_post"}, strings.NewReader("{\"ID\":\"test_post_lock\"}"), "", http.StatusOK, "/test_post"},
		{"POST", &url.URL{Path: "/test_post", RawQuery: "ID=test_post_invalid"}, strings.NewReader("the_test_post3"), "", http.StatusConflict, "with a wrong lock ID on a locked state"},
		{"GET", &url.URL{Path: "/test_post"}, nil, "the_test_post", http.StatusOK, "/test_post"},
		{"POST", &url.URL{Path: "/test_post", RawQuery: "ID=test_post_lock"}, strings.NewReader("the_test_post4"), "", http.StatusOK, "with a correct lock ID on a locked state"},
		{"GET", &url.URL{Path: "/test_post"}, nil, "the_test_post4", http.StatusOK, "/test_post"},
		{"POST", &url.URL{Path: "/test_post"}, strings.NewReader("the_test_post5"), "", http.StatusOK, "without lock ID in query string on a locked state"},
		{"GET", &url.URL{Path: "/test_post"}, nil, "the_test_post5", http.StatusOK, "/test_post"},
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

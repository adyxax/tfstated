package main

import "net/http"

func addRoutes(mux *http.ServeMux) {
	mux.Handle("GET /healthz", handleHealthz())
}

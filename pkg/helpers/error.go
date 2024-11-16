package helpers

import (
	"fmt"
	"net/http"
)

func ErrorResponse(w http.ResponseWriter, status int, err error) {
	type errorResponse struct {
		Msg    string `json:"msg"`
		Status int    `json:"status"`
	}
	_ = Encode(w, status, &errorResponse{
		Msg:    fmt.Sprintf("%+v", err),
		Status: status,
	})
}

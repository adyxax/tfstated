package logger

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

type bodyWriter struct {
	http.ResponseWriter
	status int
}

func newBodyWriter(writer http.ResponseWriter) *bodyWriter {
	return &bodyWriter{
		ResponseWriter: writer,
		status:         http.StatusNotImplemented,
	}
}

// implements http.ResponseWriter
func (w *bodyWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

// implements http.ResponseWriter
func (w *bodyWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// implements http.Flusher
func (w *bodyWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// implements http.Hijacker
func (w *bodyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hi, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hi.Hijack()
	}
	return nil, nil, errors.New("Hijack not supported")
}

package logger

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
)

type bodyWriter struct {
	http.ResponseWriter
	body    *bytes.Buffer
	maxSize int
	bytes   int
	status  int
}

func newBodyWriter(writer http.ResponseWriter, recordBody bool) *bodyWriter {
	var body *bytes.Buffer
	if recordBody {
		body = bytes.NewBufferString("")
	}
	return &bodyWriter{
		body:           body,
		maxSize:        64000,
		ResponseWriter: writer,
		status:         http.StatusNotImplemented,
	}
}

// implements http.ResponseWriter
func (w *bodyWriter) Write(b []byte) (int, error) {
	if w.body != nil {
		if w.body.Len()+len(b) > w.maxSize {
			w.body.Write(b[:w.maxSize-w.body.Len()])
		} else {
			w.body.Write(b)
		}
	}

	w.bytes += len(b)
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

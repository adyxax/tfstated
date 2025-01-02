package logger

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

func Middleware(next http.Handler, recordBody bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				slog.Error(
					"panic",
					"err", err,
					"trace", string(debug.Stack()),
				)
			}
		}()
		start := time.Now()
		path := r.URL.Path
		query := r.URL.RawQuery

		bw := newBodyWriter(w, recordBody)

		next.ServeHTTP(bw, r)

		end := time.Now()
		requestAttributes := []slog.Attr{
			slog.Time("time", start.UTC()),
			slog.String("method", r.Method),
			slog.String("host", r.Host),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", r.RemoteAddr),
		}
		responseAttributes := []slog.Attr{
			slog.Time("time", end.UTC()),
			slog.Duration("latency", end.Sub(start)),
			slog.Int("length", bw.bytes),
			slog.Int("status", bw.status),
		}
		if recordBody {
			responseAttributes = append(responseAttributes, slog.String("body", bw.body.String()))
		}
		attributes := []slog.Attr{
			{
				Key:   "request",
				Value: slog.GroupValue(requestAttributes...),
			},
			{
				Key:   "response",
				Value: slog.GroupValue(responseAttributes...),
			},
		}
		level := slog.LevelInfo
		if bw.status >= http.StatusInternalServerError {
			level = slog.LevelError
		} else if bw.status >= http.StatusBadRequest && bw.status < http.StatusInternalServerError {
			level = slog.LevelWarn
		}
		slog.LogAttrs(r.Context(), level, strconv.Itoa(bw.status)+": "+http.StatusText(bw.status), attributes...)
	})
}

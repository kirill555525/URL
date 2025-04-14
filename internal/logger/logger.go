package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Log *zap.Logger = zap.NewNop()

type responseData struct {
	code int
	size int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(data []byte) (int, error) {
	size, err := r.ResponseWriter.Write(data)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.responseData.code = code
}

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()

	if err != nil {
		return err
	}

	Log = zl
	return nil

}

func RequestLogger(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		Log.Info("got incoming HTTP request", zap.String("method", r.Method), zap.String("URI", r.RequestURI))
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)

		Log.Info("completed HTTP request", zap.Duration("duration", duration))
	}

	return http.HandlerFunc(fn)
}

func ResponseLogger(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingResponseWriter{ResponseWriter: w, responseData: &responseData{}}
		h.ServeHTTP(lw, r)

		Log.Info("Response", zap.Int("code", lw.responseData.code), zap.Int("size", lw.responseData.size))
	}

	return http.HandlerFunc(fn)
}

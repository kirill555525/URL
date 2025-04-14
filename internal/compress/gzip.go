package compress

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	return cw.zw.Write(b)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cw.Header().Set(`Content-Encoding`, `gzip`)
	}
	cw.ResponseWriter.WriteHeader(statusCode)
}

func (cw *compressWriter) Close() error {
	return cw.zw.Close()
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{ResponseWriter: w, zw: gzip.NewWriter(w)}
}

type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func (cr *compressReader) Read(b []byte) (int, error) {
	return cr.zr.Read(b)
}

func (cr *compressReader) Close() error {
	err1 := cr.ReadCloser.Close()
	err2 := cr.zr.Close()

	return errors.Join(err1, err2)

}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{ReadCloser: r, zr: zr}, nil
}

func GzipMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ow := w

		contentTypes := map[string]struct{}{
			`text/html`:        {},
			`application/json`: {},
		}

		contentType := r.Header.Get("Content-Type")

		if _, ok := contentTypes[contentType]; ok {
			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				body, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = body
				defer body.Close()
			}
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		next.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(fn)
}

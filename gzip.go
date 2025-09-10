package sqlite

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.gzipWriter.Write(b)
}

func gzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !supportsGzip(r.Header.Get("Accept-Encoding")) {
			h.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gzipWriter := gzip.NewWriter(w)
		defer gzipWriter.Close()
		h.ServeHTTP(&gzipResponseWriter{ResponseWriter: w, gzipWriter: gzipWriter}, r)
	})
}

func supportsGzip(encoding string) bool {
	return strings.Contains(encoding, "gzip")
}

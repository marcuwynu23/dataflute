package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// Compress gzip-encodes responses for clients that advertise support
// through the Accept-Encoding header.
func Compress() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gz := gzip.NewWriter(w)
			gw := &gzipResponseWriter{
				ResponseWriter: w,
				writer:         gz,
			}

			next.ServeHTTP(gw, r)

			// Finalize the gzip stream only when the response was
			// actually encoded, so bodyless responses (204, 304, 1xx)
			// stay free of a stray gzip trailer.
			if gw.headerSet {
				_ = gz.Close()
			}
		})
	}
}

// gzipResponseWriter gzip-encodes writes to an underlying response.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer    *gzip.Writer
	skip      bool
	headerSet bool
	done      bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.done {
		return
	}
	w.done = true
	if code == http.StatusNoContent || code == http.StatusNotModified || code/100 == 1 {
		w.skip = true
	}
	if !w.skip {
		w.headerSet = true
		h := w.Header()
		h.Del("Content-Length")
		h.Set("Content-Encoding", "gzip")
		h.Add("Vary", "Accept-Encoding")
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.done {
		w.WriteHeader(http.StatusOK)
	}
	if w.skip {
		return w.ResponseWriter.Write(b)
	}
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) Flush() {
	if !w.skip {
		_ = w.writer.Flush()
	}
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap lets http.NewResponseController access the underlying writer.
func (w *gzipResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

var _ http.Flusher = (*gzipResponseWriter)(nil)

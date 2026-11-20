package middleware

import (
	"bytes"
	"net/http"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	if rec.status == 0 {
		rec.status = code
	}
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.WriteHeader(http.StatusOK)
	}
	return rec.ResponseWriter.Write(b)
}

type bufferedResponseWriter struct {
	header      http.Header
	buf         bytes.Buffer
	status      int
	wroteHeader bool
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		header: make(http.Header),
	}
}

func (bw *bufferedResponseWriter) Header() http.Header {
	return bw.header
}

func (bw *bufferedResponseWriter) WriteHeader(code int) {
	if !bw.wroteHeader {
		bw.status = code
		bw.wroteHeader = true
	}
}

func (bw *bufferedResponseWriter) Write(b []byte) (int, error) {
	if !bw.wroteHeader {
		bw.WriteHeader(http.StatusOK)
	}
	return bw.buf.Write(b)
}

func (bw *bufferedResponseWriter) flushTo(w http.ResponseWriter) {
	dst := w.Header()
	for k, vals := range bw.header {
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
	if bw.wroteHeader {
		w.WriteHeader(bw.status)
	}
	if bw.buf.Len() > 0 {
		_, _ = w.Write(bw.buf.Bytes())
	}
}

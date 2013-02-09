package tigertonic

import (
	"bytes"
	"net/http"
)

type testResponseWriter struct {
	Status int
	header http.Header
	Body   bytes.Buffer
}

func (w *testResponseWriter) Header() http.Header {
	if nil == w.header {
		w.header = make(map[string][]string)
	}
	return w.header
}

func (w *testResponseWriter) Write(body []byte) (int, error) {
	return w.Body.Write(body)
}

func (w *testResponseWriter) WriteHeader(status int) {
	w.Status = status
}

package forwarder

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxy(t *testing.T) {
	body := []byte("test-body")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "original-header-value", r.Header.Get("Original-Header"))
		assert.Equal(t, "extra-header-value", r.Header.Get("Extra-Header"))
		assert.Equal(t, "1", r.URL.Query().Get("original-query"))
		assert.Equal(t, "2", r.URL.Query().Get("extra-query"))
		assert.Empty(t, r.Header.Get("Keep-Alive"))

		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, "test-body", string(body))
		assert.Equal(t, r.ContentLength, int64(len(body)))

		w.WriteHeader(http.StatusOK)
	}))
	serverURL, _ := url.Parse(server.URL)

	in := httptest.NewRequest(http.MethodGet, "https://example.com?original-query=1", bytes.NewReader(body))
	in.Header.Set("Original-Header", "original-header-value")
	in.Header.Set("Keep-Alive", "300")

	extraHeaders := http.Header{}
	extraHeaders.Set("Extra-Header", "extra-header-value")

	forwarder := New(http.DefaultClient)

	res, err := forwarder.Forward(in, &Opts{
		TargetURL: *serverURL,
		Headers:   extraHeaders,
		Query:     map[string][]string{"extra-query": {"2"}},
	})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

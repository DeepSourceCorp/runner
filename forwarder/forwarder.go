package forwarder

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

type Opts struct {
	TargetURL url.URL
	Headers   http.Header
	Query     url.Values
}

type Forwarder struct {
	client *http.Client
}

func New(client *http.Client) *Forwarder {
	return &Forwarder{client: client}
}

func (f *Forwarder) Forward(req *http.Request, opts *Opts) (*http.Response, error) {
	defer req.Body.Close()
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	out, err := http.NewRequestWithContext(
		ctx,
		req.Method,
		opts.TargetURL.String(),
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create target request: %w", err)
	}

	copyHeader(out.Header, req.Header)
	appendHeaders(out, opts.Headers)

	copyQueryParams(out, req)
	appendQueryParams(out, opts.Query)

	removeHopHeaders(out.Header)
	removeCloudflareHeaders(out.Header)

	res, err := f.client.Do(out)
	if err != nil {
		return nil, fmt.Errorf("failed to make target request: %w", err)
	}
	removeHopHeaders(res.Header)
	return res, nil
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func removeHopHeaders(h http.Header) {
	hopHeaders := []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				h.Del(sf)
			}
		}
	}

	for _, k := range hopHeaders {
		h.Del(k)
	}
}

func removeCloudflareHeaders(h http.Header) {
	cloudflareHeaders := []string{
		"CF-Connecting-IP",
		"CF-IPCountry",
		"CF-RAY",
		"CF-Visitor",
		"CF-Request-ID",
		"CF-Worker",
	}

	for _, k := range cloudflareHeaders {
		h.Del(k)
	}
}

func copyQueryParams(dst, src *http.Request) {
	q := dst.URL.Query()
	for k, v := range src.URL.Query() {
		q[k] = v
	}
	dst.URL.RawQuery = q.Encode()
}

func appendQueryParams(req *http.Request, query url.Values) {
	q := req.URL.Query()
	for k, v := range query {
		q[k] = v
	}
	req.URL.RawQuery = q.Encode()
}

func appendHeaders(req *http.Request, headers http.Header) {
	for k, v := range headers {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
}

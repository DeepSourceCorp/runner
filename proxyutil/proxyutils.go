package proxyutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func RemoveHopHeaders(h http.Header) {
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

func RemoveCloudflareHeaders(h http.Header) {
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

func CopyQueryParams(dst, src *http.Request) {
	q := dst.URL.Query()
	for k, v := range src.URL.Query() {
		q[k] = v
	}
	dst.URL.RawQuery = q.Encode()
}

func AppendQueryParams(req *http.Request, query url.Values) {
	q := req.URL.Query()
	for k, v := range query {
		q[k] = v
	}
	req.URL.RawQuery = q.Encode()
}

func AppendHeaders(req *http.Request, headers http.Header) {
	for k, v := range headers {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}
}

type ForwarderOpts struct {
	TargetURL url.URL
	Headers   http.Header
	Query     url.Values
}

type Forwarder struct {
	client *http.Client
}

func NewForwarder(client *http.Client) *Forwarder {
	return &Forwarder{client: client}
}

func (f *Forwarder) Forward(req *http.Request, extras *ForwarderOpts) (*http.Response, error) {
	defer req.Body.Close()
	ctx := req.Context()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	out, err := http.NewRequestWithContext(
		ctx,
		req.Method,
		extras.TargetURL.String(),
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create target request: %w", err)
	}

	AppendHeaders(out, extras.Headers)

	CopyQueryParams(out, req)
	AppendQueryParams(out, extras.Query)

	RemoveHopHeaders(out.Header)
	RemoveCloudflareHeaders(out.Header)

	x, _ := json.Marshal(out.Header)
	fmt.Println("REQUEST URL", out.URL.String())
	fmt.Println("REQUEST HEADERS: ", string(x))

	return f.client.Do(out)
}

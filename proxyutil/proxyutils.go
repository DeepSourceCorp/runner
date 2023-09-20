package proxyutil

import "net/http"

func CopyHeader(dst, src http.Header) {
	for k, v := range src {
		dst[k] = v
	}
}

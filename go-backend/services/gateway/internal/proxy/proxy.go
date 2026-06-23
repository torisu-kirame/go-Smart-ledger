package proxy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Handler forwards requests to an upstream base URL, preserving path, query, status and headers.
// Uses explicit forwarding instead of httputil.ReverseProxy so go-zero does not downgrade 4xx/5xx to 200.
func Handler(target string) http.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		_ = r.Body.Close()

		outURL := *remote
		outURL.Path = joinURLPath(remote.Path, r.URL.Path)
		outURL.RawQuery = r.URL.RawQuery

		outReq, err := http.NewRequestWithContext(r.Context(), r.Method, outURL.String(), bytes.NewReader(body))
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		copyHeader(outReq.Header, r.Header)
		outReq.Header.Del("Connection")
		outReq.Header.Del("Content-Length")
		outReq.ContentLength = int64(len(body))
		outReq.Host = remote.Host
		if outReq.Header.Get("X-Forwarded-Host") == "" {
			outReq.Header.Set("X-Forwarded-Host", r.Host)
		}
		if outReq.Header.Get("X-Forwarded-Proto") == "" {
			if r.TLS != nil {
				outReq.Header.Set("X-Forwarded-Proto", "https")
			} else {
				outReq.Header.Set("X-Forwarded-Proto", "http")
			}
		}

		resp, err := client.Do(outReq)
		if err != nil {
			http.Error(w, "bad gateway", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		// Do not Flush() here: go-zero timeoutWriter flushes early and locks HTTP 200.
		_, _ = io.Copy(w, resp.Body)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func joinURLPath(a, b string) string {
	if a == "" || a == "/" {
		return b
	}
	if b == "" || b == "/" {
		return a
	}
	return strings.TrimSuffix(a, "/") + "/" + strings.TrimPrefix(b, "/")
}

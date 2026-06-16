package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Handler forwards requests to an upstream base URL, preserving path and query.
func Handler(target string) http.HandlerFunc {
	remote, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	p := httputil.NewSingleHostReverseProxy(remote)
	p.FlushInterval = -1
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = remote.Scheme
		r.URL.Host = remote.Host
		r.Host = remote.Host
		if !strings.HasPrefix(r.URL.Path, "/") {
			r.URL.Path = "/" + r.URL.Path
		}
		p.ServeHTTP(w, r)
	}
}

package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ReverseProxy :nodoc:
type ReverseProxy interface {
	AddServer(serverURL string) error
	Proxy(proxy *httputil.ReverseProxy, target *url.URL, wr http.ResponseWriter, req *http.Request) error
	Handler(w http.ResponseWriter, r *http.Request)
	Start()
}

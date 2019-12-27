package proxy

import (
	"net/http"
	"net/url"
)

// ReverseProxy :nodoc:
type ReverseProxy interface {
	AddServer(serverURL string) error
	Proxy(target *url.URL, wr http.ResponseWriter, req *http.Request) error
	Handler(w http.ResponseWriter, r *http.Request)
	Start()
}

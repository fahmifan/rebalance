package proxy

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// RoundRobin :nodoc:
type RoundRobin struct {
	urls    []*url.URL
	mapURL  map[string]string
	nextURL int
}

// NewRoundRobin :nodoc:
func NewRoundRobin() ReverseProxy {
	return &RoundRobin{
		mapURL: make(map[string]string),
	}
}

// Start round robin server :nodoc:
func (rr *RoundRobin) Start() {
	if len(rr.urls) == 0 {
		log.Fatal(errors.New("no urls"))
	}

	http.HandleFunc("/", rr.Handler)
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal(err)
	}
}

// AddServer :nodoc:
func (rr *RoundRobin) AddServer(serverURL string) error {
	if _, ok := rr.mapURL[serverURL]; ok {
		return errors.New("server url already added")
	}

	parsed, err := url.Parse(serverURL)
	if err != nil {
		return err
	}

	rr.urls = append(rr.urls, parsed)
	rr.mapURL[serverURL] = serverURL

	return nil
}

// Proxy :nodoc:
func (rr *RoundRobin) Proxy(target *url.URL, wr http.ResponseWriter, req *http.Request) error {
	proxy := httputil.NewSingleHostReverseProxy(target)
	req.URL.Host = target.Host
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.URL.Scheme = target.Scheme

	proxy.ServeHTTP(wr, req)
	return nil
}

// Handler :nodoc:
func (rr *RoundRobin) Handler(w http.ResponseWriter, r *http.Request) {
	rr.Proxy(rr.urls[rr.findNextURL()], w, r)
}

func (rr *RoundRobin) findNextURL() int {
	rr.nextURL = (rr.nextURL + 1) % len(rr.urls)
	return rr.nextURL
}

package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
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
	http.HandleFunc("/", rr.Handler)
	http.HandleFunc("/rebalance/join", rr.HandleJoin)

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

// FromRequest extracts the user IP address from req, if present.
func FromRequest(req *http.Request) (net.IP, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	return userIP, nil
}

// HandleJoin :nodoc:
func (rr *RoundRobin) HandleJoin(w http.ResponseWriter, r *http.Request) {
	ip, err := FromRequest(r)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("requst join from host ", ip.String())
	if err := rr.AddServer("http://" + ip.String()); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		resp, err := json.Marshal(map[string]interface{}{"error": err.Error()})
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Handler :nodoc:
func (rr *RoundRobin) Handler(w http.ResponseWriter, r *http.Request) {
	if len(rr.urls) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rr.Proxy(rr.urls[rr.findNextURL()], w, r)
}

func (rr *RoundRobin) findNextURL() int {
	rr.nextURL = (rr.nextURL + 1) % len(rr.urls)
	return rr.nextURL
}

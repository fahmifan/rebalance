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
	"sync/atomic"
)

// RoundRobin :nodoc:
type RoundRobin struct {
	urls    []*url.URL
	proxies []*httputil.ReverseProxy
	mapURL  map[string]string
	nextURL uint64
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

	taretURL, err := url.Parse(serverURL)
	if err != nil {
		return err
	}

	rr.urls = append(rr.urls, taretURL)
	rr.mapURL[serverURL] = serverURL
	rr.proxies = append(rr.proxies, httputil.NewSingleHostReverseProxy(taretURL))

	return nil
}

// Proxy :nodoc:
func (rr *RoundRobin) Proxy(proxy *httputil.ReverseProxy, target *url.URL, wr http.ResponseWriter, req *http.Request) error {
	req.URL.Host = target.Host
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.URL.Scheme = target.Scheme

	proxy.ServeHTTP(wr, req)
	return nil
}

// getClientIP extracts the user IP address from req, if present.
func getClientIP(req *http.Request) (net.IP, error) {
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
	ip, err := getClientIP(r)
	if err != nil {
		log.Fatal(err)
	}

	port := ":" + r.URL.Query().Get("port")

	log.Println("requst join from host ", ip.String()+port)
	if err := rr.AddServer("http://" + ip.String() + port); err != nil {
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
	w.Write([]byte("success join"))
}

// Handler :nodoc:
func (rr *RoundRobin) Handler(w http.ResponseWriter, r *http.Request) {
	if len(rr.urls) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	nextURL := rr.findNextURL()
	rr.Proxy(rr.proxies[nextURL], rr.urls[nextURL], w, r)
}

func (rr *RoundRobin) findNextURL() uint64 {
	atomic.StoreUint64(&rr.nextURL, atomic.AddUint64(&rr.nextURL, uint64(1))%uint64(len(rr.urls)))
	return rr.nextURL
}

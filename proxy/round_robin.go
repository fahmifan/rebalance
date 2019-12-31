package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	_Attempts   int = 0
	_Retry          = 1
	_MaxRetries     = 3
	_MaxAttemp      = 3
)

// Service :nodoc:
type Service struct {
	Proxy   *httputil.ReverseProxy
	URL     *url.URL
	isAlive bool
	mutex   sync.RWMutex
}

// NewService :nodoc:
func NewService(p *httputil.ReverseProxy, u *url.URL) *Service {
	return &Service{
		Proxy:   p,
		URL:     u,
		isAlive: true,
	}
}

// SetAlive :nodoc:
func (s *Service) SetAlive(alive bool) {
	s.mutex.Lock()
	s.isAlive = alive
	s.mutex.Unlock()
}

// IsAlive :nodoc:
func (s *Service) IsAlive() (alive bool) {
	s.mutex.Lock()
	alive = s.isAlive
	s.mutex.Unlock()
	return
}

// RoundRobin :nodoc:
type RoundRobin struct {
	urls     []*url.URL
	proxies  []*httputil.ReverseProxy
	mapURL   map[string]string
	nextURL  uint64
	services []*Service
	current  uint64
}

// NewRoundRobin :nodoc:
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		mapURL:   make(map[string]string),
		services: make([]*Service, 0),
	}
}

// NextIndex :nodoc:
func (rr *RoundRobin) NextIndex() int {
	return int(atomic.AddUint64(&rr.nextURL, uint64(1)) % uint64(len(rr.services)))
}

// Start round robin server :nodoc:
func (rr *RoundRobin) Start() {
	m := &http.ServeMux{}
	m.HandleFunc("/", rr.Handler)
	m.HandleFunc("/rebalance/join", rr.HandleJoin)

	if err := http.ListenAndServe(":9000", m); err != nil {
		log.Fatal(err)
	}
}

// AddServer :nodoc:
func (rr *RoundRobin) AddServer(targetURL string) error {
	if _, ok := rr.mapURL[targetURL]; ok {
		return errors.New("server url already added")
	}

	serviceURL, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	if ok := isServiceAlive(serviceURL); !ok {
		return errors.New("cannot dial service")
	}

	rr.urls = append(rr.urls, serviceURL)
	rr.mapURL[targetURL] = targetURL

	proxy := httputil.NewSingleHostReverseProxy(serviceURL)
	proxy.FlushInterval = -1
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          10000,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	proxy.Transport = transport

	rr.services = append(rr.services, NewService(proxy, serviceURL))

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

// FindNextService find next alive service
func (rr *RoundRobin) FindNextService() *Service {
	if len(rr.services) == 0 {
		return nil
	}

	next := rr.NextIndex()
	n := len(rr.services) + next
	nservice := len(rr.services)

	for i := next; i < n; i++ {
		idx := i % nservice
		if rr.services[idx].IsAlive() {
			isSameService := i == next
			if !isSameService {
				atomic.StoreUint64(&rr.current, uint64(idx))
			}

			return rr.services[idx]

		}
	}

	return nil
}

// Handler :nodoc:
func (rr *RoundRobin) Handler(w http.ResponseWriter, r *http.Request) {
	// if the same request routing for few attempts with different backends, increase the count
	attempts := getRetryAttemptsFromCtx(r, _Attempts)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	service := rr.FindNextService()
	if service == nil {
		http.Error(w, "service not available", http.StatusServiceUnavailable)
		return
	}

	service.Proxy.ServeHTTP(w, r)
	service.Proxy.ErrorHandler = rr.ProxyHandler(service)
}

// ProxyHandler :nodoc:
func (rr *RoundRobin) ProxyHandler(service *Service) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		if err != nil {
			log.Println(err)
		}

		retries := getRetryAttemptsFromCtx(r, _Retry)
		if retries < _MaxRetries {
			select {
			case <-time.After(10 * time.Millisecond):
				ctx := context.WithValue(r.Context(), _Retry, retries+1)
				service.Proxy.ServeHTTP(w, r.WithContext(ctx))
			}

			return
		}

		service.SetAlive(false)

		// if the same request routing for few attempts with different backends, increase the count
		attempts := getRetryAttemptsFromCtx(r, _Attempts)
		log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		service := rr.FindNextService()
		if service == nil {
			http.Error(w, "service not available", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), _Attempts, attempts+1)
		rr.Handler(w, r.WithContext(ctx))
	}
}

// HealthCheck check services health status
// mark service as alive if helathy
func (rr *RoundRobin) HealthCheck() {
	for _, s := range rr.services {
		status := "up"
		alive := isServiceAlive(s.URL)
		s.SetAlive(alive)
		if !alive {
			status = "down"
		}

		log.Printf("%s [%s]\n", s.URL, status)
	}
}

// RunHealthCheck run HealthCheck every 20 second
func (rr *RoundRobin) RunHealthCheck() {
	t := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			rr.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

func (rr *RoundRobin) findNextURL() uint64 {
	next := atomic.AddUint64(&rr.nextURL, uint64(1)) % uint64(len(rr.urls))
	atomic.StoreUint64(&rr.nextURL, next)
	return next
}

func getRetryAttemptsFromCtx(r *http.Request, retyAttempKey int) int {
	if val, ok := r.Context().Value(retyAttempKey).(int); ok {
		return val
	}

	return 0
}

func isServiceAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	_ = conn.Close()

	return true
}

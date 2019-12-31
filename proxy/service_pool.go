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
	"sync/atomic"
	"time"
)

type key int

const (
	_Attempts key = 0
	_Retry    key = 1

	_MaxRetries int = 3
	_MaxAttempt int = 3
)

// ServiceProxy :nodoc:
type ServiceProxy struct {
	mapURL         map[string]string
	services       []*Service
	currentService uint64
}

// NewServiceProxy :nodoc:
func NewServiceProxy() *ServiceProxy {
	return &ServiceProxy{
		mapURL:   make(map[string]string),
		services: make([]*Service, 0),
	}
}

// NextIndex :nodoc:
func (sp *ServiceProxy) NextIndex() int {
	return int(atomic.AddUint64(&sp.currentService, uint64(1)) % uint64(len(sp.services)))
}

// Start round robin server :nodoc:
func (sp *ServiceProxy) Start() {
	m := &http.ServeMux{}
	m.HandleFunc("/", sp.Handler)
	m.HandleFunc("/rebalance/join", sp.HandleJoin)

	if err := http.ListenAndServe(":9000", m); err != nil {
		log.Fatal(err)
	}
}

// AddServer :nodoc:
func (sp *ServiceProxy) AddServer(targetURL string) error {
	if _, ok := sp.mapURL[targetURL]; ok {
		return errors.New("server url already added")
	}

	serviceURL, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	if ok := isServiceAlive(serviceURL); !ok {
		return errors.New("cannot dial service")
	}

	sp.mapURL[targetURL] = targetURL

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

	sp.services = append(sp.services, NewService(proxy, serviceURL))

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
func (sp *ServiceProxy) HandleJoin(w http.ResponseWriter, r *http.Request) {
	ip, err := getClientIP(r)
	if err != nil {
		log.Fatal(err)
	}

	port := ":" + r.URL.Query().Get("port")

	log.Println("requst join from host ", ip.String()+port)
	if err := sp.AddServer("http://" + ip.String() + port); err != nil {
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
func (sp *ServiceProxy) FindNextService() *Service {
	if len(sp.services) == 0 {
		return nil
	}

	next := sp.NextIndex()
	n := len(sp.services) + next
	nservice := len(sp.services)

	for i := next; i < n; i++ {
		idx := i % nservice
		if sp.services[idx].IsAlive() {
			isSameService := i == next
			if !isSameService {
				atomic.StoreUint64(&sp.currentService, uint64(idx))
			}

			return sp.services[idx]

		}
	}

	return nil
}

// Handler :nodoc:
func (sp *ServiceProxy) Handler(w http.ResponseWriter, r *http.Request) {
	// if the same request routing for few attempts with different backends, increase the count
	attempts := getRetryAttemptsFromCtx(r, _Attempts)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	service := sp.FindNextService()
	if service == nil {
		http.Error(w, "service not available", http.StatusServiceUnavailable)
		return
	}

	service.Proxy.ServeHTTP(w, r)
	service.Proxy.ErrorHandler = sp.ProxyHandler(service)
}

// ProxyHandler :nodoc:
func (sp *ServiceProxy) ProxyHandler(service *Service) func(w http.ResponseWriter, r *http.Request, err error) {
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
		log.Printf("%s(%s) attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		service := sp.FindNextService()
		if service == nil {
			http.Error(w, "service not available", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), _Attempts, attempts+1)
		sp.Handler(w, r.WithContext(ctx))
	}
}

// HealthCheck check services health status
// mark service as alive if helathy
func (sp *ServiceProxy) HealthCheck() {
	for _, s := range sp.services {
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
func (sp *ServiceProxy) RunHealthCheck() {
	t := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			sp.HealthCheck()
			log.Println("Health check completed")
		}
	}
}

func (sp *ServiceProxy) findNextURL() uint64 {
	next := atomic.AddUint64(&sp.currentService, uint64(1)) % uint64(len(sp.services))
	atomic.StoreUint64(&sp.currentService, next)
	return next
}

func getRetryAttemptsFromCtx(r *http.Request, retyAttempKey key) int {
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

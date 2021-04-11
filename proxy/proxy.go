package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type key int

const (
	_attemptsKey key = 0
	_retryKey    key = 1

	_maxAttempt int = 3
)

// Proxy :nodoc:
type Proxy struct {
	server *http.Server

	servicesMut *sync.RWMutex
	services    []*Service

	mapURLMut *sync.RWMutex
	mapURL    map[string]string

	currentServiceMut *sync.RWMutex
	currentService    int
}

// NewProxy ServiceProxy factory
func NewProxy() *Proxy {
	return &Proxy{
		services:          make([]*Service, 0),
		mapURL:            make(map[string]string),
		mapURLMut:         &sync.RWMutex{},
		currentServiceMut: &sync.RWMutex{},
		servicesMut:       &sync.RWMutex{},
	}
}

// Start round robin server :nodoc:
func (sp *Proxy) Start() {
	m := &http.ServeMux{}
	m.HandleFunc("/", sp.handleProxy)
	m.HandleFunc("/rebalance/join", sp.handleJoin)
	m.HandleFunc("/rebalance/local-join", sp.handleLocalJoin)

	sp.server = &http.Server{Addr: ":9000", Handler: m}
	if err := sp.server.ListenAndServe(); err != nil {
		log.Error(err)
	}
}

// Stop stop loadbalancer
func (sp *Proxy) Stop(ctx context.Context) {
	if err := sp.server.Shutdown(ctx); err != nil {
		log.Error(err)
	}
}

var errExists = errors.New("error already exists")

// AddService :nodoc:
func (sp *Proxy) AddService(targetURL string) error {
	if _, ok := sp.mapURL[targetURL]; ok {
		return errExists
	}

	serviceURL, err := url.Parse(targetURL)
	if err != nil {
		return err
	}

	if ok := dial(serviceURL); !ok {
		return errors.New("cannot dial service")
	}

	sp.addTargetURL(targetURL)
	sp.addService(serviceURL)
	return nil
}

// RunHealthCheck run HealthCheck every 20 second
func (sp *Proxy) RunHealthCheck(sigInterrupt chan os.Signal) {
	t := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-t.C:
			log.Println("Starting health check...")
			sp.checkHealth()
			log.Println("Health check completed")
		case <-sigInterrupt:
			t.Stop()
			return
		}
	}
}

// handleJoin :nodoc:
func (sp *Proxy) handleJoin(w http.ResponseWriter, r *http.Request) {
	ip, err := extractIP(r)
	if err != nil {
		log.Fatal(err)
	}

	port := ":" + r.URL.Query().Get("port")

	log.Println("requst join from host ", ip.String()+port)
	err = sp.AddService("http://" + ip.String() + port)
	if err == errExists {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("already exists"))
		return
	}
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		resp, err := json.Marshal(map[string]interface{}{"error": err.Error()})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success join"))
}

// handleLocalJoin :nodoc:
func (sp *Proxy) handleLocalJoin(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	log.Println("requst join from host ", host)
	err := sp.AddService(host)
	if err == errExists {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("already exists"))
		return
	}
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		resp, err := json.Marshal(map[string]interface{}{"error": err.Error()})
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success join"))
}

// handleProxy :nodoc:
func (sp *Proxy) handleProxy(w http.ResponseWriter, r *http.Request) {
	// if the same request routed for few attempts with different service, increase the count
	attempts := retryAttemptsFromCtx(r, _attemptsKey)
	if attempts > _maxAttempt {
		log.Infof("%s(%s) max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "service not available", http.StatusServiceUnavailable)
		return
	}

	service := sp.findNextService()
	if service == nil {
		http.Error(w, "service not available", http.StatusServiceUnavailable)
		return
	}

	service.Proxy.ErrorHandler = sp.proxyErrorHandler(service)
	service.Proxy.ServeHTTP(w, r)
}

func (sp *Proxy) proxyErrorHandler(service *Service) func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		if err != nil {
			log.Error(err)
		}

		attempt := retryAttemptsFromCtx(r, _retryKey)
		if attempt < _maxAttempt {
			time.After(10 * time.Millisecond)
			ctx := context.WithValue(r.Context(), _retryKey, attempt+1)
			service.Proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		service.SetAlive(false)

		// if the same request routing for few attempts with different service,
		// increase the count
		attempts := retryAttemptsFromCtx(r, _attemptsKey)
		log.Infof("%s(%s) attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		service := sp.findNextService()
		if service == nil {
			http.Error(w, "service not available", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), _attemptsKey, attempts+1)
		sp.handleProxy(w, r.WithContext(ctx))
	}
}

// findNextService find next alive service
func (sp *Proxy) findNextService() *Service {
	if len(sp.services) == 0 {
		return nil
	}

	next := sp.currentService % len(sp.services)
	nservice := len(sp.services)
	// make itter to nservice+1 so when it reach `idx = nservice-1`,
	// but the serivce of `idx` is not alive, it will go back to `idx = 0`
	for i := next; i < nservice+1; i++ {
		idx := i % nservice
		if !sp.services[idx].IsAlive() {
			continue
		}

		currentService := (idx + 1) % nservice
		sp.setCurrentService(currentService)

		return sp.services[idx]
	}

	// no service found
	return nil
}

// checkHealth check services health status
// mark service as alive if helathy
func (sp *Proxy) checkHealth() {
	for i := range sp.services {
		alive := dial(sp.services[i].URL)
		sp.services[i].SetAlive(alive)
		status := "up"
		if !alive {
			status = "down"
		}

		log.Infof("%s [%s]\n", sp.services[i].URL, status)
	}
}

func (sp *Proxy) setCurrentService(val int) {
	sp.currentServiceMut.Lock()
	sp.currentService = val
	sp.currentServiceMut.Unlock()
}

func (sp *Proxy) addTargetURL(targetURL string) {
	sp.mapURLMut.Lock()
	sp.mapURL[targetURL] = targetURL
	sp.mapURLMut.Unlock()
}

func (sp *Proxy) addService(serviceURL *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(serviceURL)
	proxy.Transport = &http.Transport{
		DisableCompression:  true,
		DisableKeepAlives:   false,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	sp.servicesMut.Lock()
	sp.services = append(sp.services, NewService(proxy, serviceURL))
	sp.servicesMut.Unlock()
}

// extracts the user IP address from req, if present.
func extractIP(req *http.Request) (net.IP, error) {
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

func retryAttemptsFromCtx(r *http.Request, retyAttempKey key) int {
	if val, ok := r.Context().Value(retyAttempKey).(int); ok {
		return val
	}
	return 0
}

// success dial means the service is alive
func dial(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}

	_ = conn.Close()
	return true
}

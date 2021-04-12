package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// Service :nodoc:
type Service struct {
	Proxy   *httputil.ReverseProxy
	URL     *url.URL
	isAlive bool
	mutex   *sync.RWMutex
}

type serviceOpt func(s *Service)

func WithTransport(t http.RoundTripper) serviceOpt {
	return func(s *Service) {
		s.Proxy.Transport = t
	}
}

// NewService :nodoc:
func NewService(u *url.URL, opts ...serviceOpt) *Service {
	p := httputil.NewSingleHostReverseProxy(u)
	s := &Service{
		Proxy: p,
		URL:   u,
		mutex: &sync.RWMutex{},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SetAlive :nodoc:
func (s *Service) SetAlive(alive bool) {
	s.mutex.Lock()
	s.isAlive = alive
	s.mutex.Unlock()
}

// IsAlive :nodoc:
func (s *Service) IsAlive() (alive bool) {
	return s.isAlive
}

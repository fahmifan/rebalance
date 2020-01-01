package proxy

import (
	"net/http/httputil"
	"net/url"
	"sync"
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

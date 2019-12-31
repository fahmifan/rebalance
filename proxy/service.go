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

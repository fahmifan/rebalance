package config

import (
	"os"
)

const (
	// DefaultServicePort :nodoc:
	DefaultServicePort = "80"
)

// ClientServicePort :nodoc:
func ClientServicePort() string {
	if val, ok := os.LookupEnv("SERVICE_PORT"); ok {
		return val
	}

	return DefaultServicePort
}

package registry

import (
	"context"
	"net/http"
)

// Registry defines the core interface that udistribution needs from any registry implementation
type Registry interface {
	// ServeHTTP serves registry HTTP requests
	ServeHTTP(w http.ResponseWriter, r *http.Request)

	// Health checks if the registry is healthy
	Health(ctx context.Context) error

	// Shutdown gracefully shuts down the registry
	Shutdown(ctx context.Context) error
}

// RegistryConfig represents udistribution's abstracted configuration
type RegistryConfig struct {
	Storage StorageConfig `yaml:"storage,omitempty"`
	Logging LoggingConfig `yaml:"log,omitempty"`
	HTTP    HTTPConfig    `yaml:"http,omitempty"`
}

// StorageConfig defines storage backend configuration
type StorageConfig struct {
	Type       string                 `yaml:"type"`
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
}

// LoggingConfig defines logging configuration
type LoggingConfig struct {
	Level  string                 `yaml:"level,omitempty"`
	Fields map[string]interface{} `yaml:"fields,omitempty"`
}

// HTTPConfig defines HTTP server configuration
type HTTPConfig struct {
	Secret string `yaml:"secret,omitempty"`
}

// RegistryFactory creates registry instances
type RegistryFactory interface {
	NewRegistry(ctx context.Context, config *RegistryConfig) (Registry, error)
}
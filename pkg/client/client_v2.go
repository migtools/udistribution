package client

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"

	_ "github.com/distribution/distribution/v3/registry/storage/driver/azure"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/base"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/factory"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/gcs"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/middleware"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/oss"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/s3-aws"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/swift"

	uconfiguration "github.com/migtools/udistribution/pkg/distribution/configuration"
	def "github.com/migtools/udistribution/pkg/client/default"
	ureg "github.com/migtools/udistribution/pkg/registry"
	"github.com/migtools/udistribution/pkg/registry/distribution"
)

// ClientV2 uses the new registry abstraction layer
type ClientV2 struct {
	registry ureg.Registry
	config   *ureg.RegistryConfig
	factory  ureg.RegistryFactory
}

// NewClientV2 creates a new client using the registry abstraction
func NewClientV2(configString string, envs []string) (*ClientV2, error) {
	if configString == "" {
		configString = def.Config
	}

	// Parse config into udistribution's config format
	config, err := parseConfigV2(configString, envs)
	if err != nil {
		return nil, ureg.ConvertDistributionError(err)
	}

	// Use factory to create registry
	factory := &distribution.DistributionFactory{}
	ctx := context.Background()

	registry, err := factory.NewRegistry(ctx, config)
	if err != nil {
		return nil, ureg.ConvertDistributionError(err)
	}

	return &ClientV2{
		registry: registry,
		config:   config,
		factory:  factory,
	}, nil
}

func (c *ClientV2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.registry.ServeHTTP(w, r)
}

func (c *ClientV2) Health(ctx context.Context) error {
	return c.registry.Health(ctx)
}

func (c *ClientV2) Shutdown(ctx context.Context) error {
	return c.registry.Shutdown(ctx)
}

func (c *ClientV2) GetConfig() *ureg.RegistryConfig {
	return c.config
}

// parseConfigV2 parses environment variables and config string into udistribution config format
func parseConfigV2(configString string, envs []string) (*ureg.RegistryConfig, error) {
	// First parse using the existing distribution parser
	distConfig, err := uconfiguration.ParseEnvironment(configString, envs)
	if err != nil {
		return nil, err
	}

	// Generate secret if needed
	if distConfig.HTTP.Secret == "" {
		secret, err := generateSecret()
		if err != nil {
			return nil, err
		}
		distConfig.HTTP.Secret = secret
	}

	// Convert to udistribution config format
	config := &ureg.RegistryConfig{
		Storage: ureg.StorageConfig{
			Type:       distConfig.Storage.Type(),
			Parameters: convertParameters(distConfig.Storage.Parameters()),
		},
		HTTP: ureg.HTTPConfig{
			Secret: distConfig.HTTP.Secret,
		},
	}

	// Set logging config if present
	if distConfig.Log.Level != "" {
		config.Logging = ureg.LoggingConfig{
			Level:  string(distConfig.Log.Level),
			Fields: distConfig.Log.Fields,
		}
	}

	return config, nil
}

func convertParameters(params map[string]interface{}) map[string]interface{} {
	if params == nil {
		return make(map[string]interface{})
	}
	
	// Deep copy to avoid modifying original
	result := make(map[string]interface{})
	for k, v := range params {
		result[k] = v
	}
	return result
}

func generateSecret() (string, error) {
	const randomSecretSize = 32
	var secretBytes [randomSecretSize]byte
	if _, err := rand.Read(secretBytes[:]); err != nil {
		return "", fmt.Errorf("could not generate random bytes for HTTP secret: %v", err)
	}
	return string(secretBytes[:]), nil
}
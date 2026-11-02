package distribution

import (
	"context"
	"net/http"

	"github.com/distribution/distribution/v3/configuration"
	dcontext "github.com/distribution/distribution/v3/context"
	"github.com/distribution/distribution/v3/registry/handlers"
	"github.com/distribution/distribution/v3/version"
	ureg "github.com/migtools/udistribution/pkg/registry"
	"github.com/migtools/udistribution/pkg/distribution/registry"
)

// DistributionRegistry adapts distribution/distribution to our Registry interface
type DistributionRegistry struct {
	app    *handlers.App
	config *configuration.Configuration
	ctx    context.Context
}

func (d *DistributionRegistry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.app.ServeHTTP(w, r)
}

func (d *DistributionRegistry) Health(ctx context.Context) error {
	// Distribution doesn't provide a direct health check API
	// We assume if the app was created successfully, it's healthy
	return nil
}

func (d *DistributionRegistry) Shutdown(ctx context.Context) error {
	// Distribution doesn't provide a direct shutdown API
	// The registry cleanup is handled by garbage collection
	return nil
}

// DistributionFactory creates distribution-based registries
type DistributionFactory struct{}

func (f *DistributionFactory) NewRegistry(ctx context.Context, config *ureg.RegistryConfig) (ureg.Registry, error) {
	// Convert udistribution config to distribution config
	distConfig, err := convertConfig(config)
	if err != nil {
		return nil, err
	}

	// Set up context similar to distribution's registry setup
	registryCtx := dcontext.WithVersion(dcontext.Background(), version.Version)
	
	// Configure logging
	registryCtx, err = registry.ConfigureLogging(registryCtx, distConfig)
	if err != nil {
		return nil, ureg.ConvertDistributionError(err)
	}

	// Configure bugsnag
	registry.ConfigureBugsnag(distConfig)

	// Create handlers.App
	app := handlers.NewApp(registryCtx, distConfig)

	return &DistributionRegistry{
		app:    app,
		config: distConfig,
		ctx:    registryCtx,
	}, nil
}

// convertConfig converts udistribution config to distribution config
func convertConfig(uconfig *ureg.RegistryConfig) (*configuration.Configuration, error) {
	config := &configuration.Configuration{
		Version: configuration.MajorMinorVersion(0, 1),
		Storage: configuration.Storage{
			uconfig.Storage.Type: configuration.Parameters(uconfig.Storage.Parameters),
		},
	}

	// Set HTTP config
	config.HTTP.Secret = uconfig.HTTP.Secret

	// Set default log level if not specified
	if uconfig.Logging.Level != "" {
		config.Log.Level = configuration.Loglevel(uconfig.Logging.Level)
		if uconfig.Logging.Fields != nil {
			config.Log.Fields = uconfig.Logging.Fields
		}
	}

	return config, nil
}
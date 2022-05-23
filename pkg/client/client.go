package client

import (
	"context"
	"crypto/rand"
	"fmt"

	// "github.com/distribution/distribution/v3/registry/storage/driver/factory"
	uconfiguration "github.com/kaovilai/udistribution/pkg/distribution/configuration"
	"github.com/kaovilai/udistribution/pkg/distribution/registry"

	"github.com/distribution/distribution/v3/configuration"
	dcontext "github.com/distribution/distribution/v3/context"
	"github.com/distribution/distribution/v3/registry/handlers"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/azure"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/base"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/factory"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/filesystem"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/gcs"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/middleware"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/oss"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/s3-aws"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/swift"
	_ "github.com/distribution/distribution/v3/registry/storage/driver/testdriver"

	"github.com/distribution/distribution/v3/uuid"
	"github.com/distribution/distribution/v3/version"
	def "github.com/kaovilai/udistribution/pkg/client/default"
)

type Client struct {
	config *configuration.Configuration
	app    *handlers.App
}

// NewClient creates a new client from the provided configuration.
func NewClient(configString string, envs []string) (client *Client, err error) {
	if configString == "" {
		configString = def.Config
	}
	// resolve configuration using parameters
	config, err := uconfiguration.ParseEnvironment(configString, envs)
	if err != nil {
		return nil, err
	}
	configureSecret(config)
	ctx, err := GetContext(config)
	if err != nil {
		return nil, err
	}
	// configure bugsnag like https://github.com/distribution/distribution/blob/4363fb1ef4676df2b9d99e3630e1b568141597c4/registry/registry.go#L148
	registry.ConfigureBugsnag(config)
	// inject a logger into the uuid library. warns us if there is a problem
	// with uuid generation under low entropy.
	uuid.Loggerf = dcontext.GetLogger(ctx).Warnf
	client = &Client{
		config: config,
		app:    handlers.NewApp(ctx, config),
	}
	return client, err
	// // initialize driver factory like https://github.com/distribution/distribution/blob/1d33874951b749df7e070b1c702ea418bbc57ed1/registry/root.go#L55
	// storageParams := config.Storage.Parameters()
	// if storageParams == nil {
	// 	storageParams = make(configuration.Parameters)
	// }
	// storageParams["useragent"] = fmt.Sprintf("docker-distribution/%s %s", version.Version, runtime.Version())

	// driver, err := factory.Create(config.Storage.Type(), storageParams)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "failed to construct %s driver: %v", config.Storage.Type(), err)
	// }
}

// randomSecretSize is the number of random bytes to generate if no secret
// was specified.
const randomSecretSize = 32

// configureSecret creates a random secret if a secret wasn't included in the
// configuration.
func configureSecret(configuration *configuration.Configuration) {
	if configuration.HTTP.Secret == "" {
		var secretBytes [randomSecretSize]byte
		if _, err := rand.Read(secretBytes[:]); err != nil {
			panic(fmt.Sprintf("could not generate random bytes for HTTP secret: %v", err))
		}
		configuration.HTTP.Secret = string(secretBytes[:])
	}
}

func (c *Client) GetApp() *handlers.App {
	return c.app
}

func GetContext(config *configuration.Configuration) (context.Context, error) {
	// setup context like https://github.com/distribution/distribution/blob/4363fb1ef4676df2b9d99e3630e1b568141597c4/registry/registry.go#L94
	ctx := dcontext.WithVersion(dcontext.Background(), version.Version)
	// configure logging like https://github.com/distribution/distribution/blob/4363fb1ef4676df2b9d99e3630e1b568141597c4/registry/registry.go#L143
	ctx, err := registry.ConfigureLogging(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error configuring logger: %v", err)
	}
	return ctx, nil
}

package client

import (
	"errors"
	"fmt"

	uconfiguration "github.com/kaovilai/udistribution/pkg/configuration"

	"runtime"

	distribution "github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/configuration"
	dcontext "github.com/distribution/distribution/v3/context"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/storage"
	"github.com/distribution/distribution/v3/registry/storage/driver"
	"github.com/distribution/distribution/v3/registry/storage/driver/factory"
	"github.com/distribution/distribution/v3/version"
)

type Client struct {
	config *configuration.Configuration
	storage driver.StorageDriver
	registry distribution.Namespace
}

// NewClient creates a new client from the provided configuration.
func NewClient(config *configuration.Configuration, env []string) (client *Client, err error) {
	if config == nil {
		config = getDefaultConfig()
	}
	c, err := uconfiguration.ParseEnvironment(config, env)
	if err != nil {
		return nil, err
	}
	client = &Client{
		config: c,
	}
	client.initStorage()
	ctx := dcontext.WithVersion(dcontext.Background(), version.Version)
	client.registry, err = storage.NewRegistry(ctx, client.storage, )
	return client, err
}

func (c *Client) initStorage() (err error) {
	if c.config == nil {
		return errors.New("configuration is nil")
	}
	// override the storage driver's UA string for registry outbound HTTP requests
	storageParams := c.config.Storage.Parameters()
	if storageParams == nil {
		storageParams = make(configuration.Parameters)
	}
	storageParams["useragent"] = fmt.Sprintf("docker-distribution/%s %s", version.Version, runtime.Version())

	c.storage, err = factory.Create(c.config.Storage.Type(), storageParams)
	if err != nil {
		return err
	}
	// TODO: Add more bits from https://github.com/distribution/distribution/blob/f637481c67241151dc6d6fe2b12852e2ad8d70c2/registry/handlers/app.go#L155
	return nil
}

func (c *Client) Registry() distribution.Namespace {
	return c.registry
}

func testNamespace(){
	c, err := NewClient(nil, nil)
	if err != nil {
		panic(err)
	}
	r := c.Registry()
	ref, err := reference.ParseNamed("test/test")
	if err != nil {
		panic(err)
	}
	r.Repository(dcontext.Background(), ref)
}
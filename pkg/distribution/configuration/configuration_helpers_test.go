package configuration

import (
	"net/http"
	"time"

	"github.com/distribution/distribution/v3/configuration"
)

type wantConfigOption func(*configuration.Configuration)

func getWantConfig(opts ...wantConfigOption) *configuration.Configuration {
	config := defaultWantConfig
	for _, o := range opts {
		o(&config)
	}
	return &config
}

func withStorage(storage configuration.Storage) wantConfigOption {
	return func(config *configuration.Configuration) {
		config.Storage = storage
	}
}

func withHeaders(headers http.Header) wantConfigOption {
	return func(config *configuration.Configuration) {
		config.HTTP.Headers = headers
	}
}

func withHTTP(http HTTP) wantConfigOption {
	return func(config *configuration.Configuration) {
		config.HTTP.Addr = http.Addr
		config.HTTP.Net = http.Net
		config.HTTP.Host = http.Host
		config.HTTP.Headers = http.Headers
		// TODO: #1 Add more fields from struct
	}
}

type HTTP struct {
	// Addr specifies the bind address for the registry instance.
	Addr string `yaml:"addr,omitempty"`

	// Net specifies the net portion of the bind address. A default empty value means tcp.
	Net string `yaml:"net,omitempty"`

	// Host specifies an externally-reachable address for the registry, as a fully
	// qualified URL.
	Host string `yaml:"host,omitempty"`

	Prefix string `yaml:"prefix,omitempty"`

	// Secret specifies the secret key which HMAC tokens are created with.
	Secret string `yaml:"secret,omitempty"`

	// RelativeURLs specifies that relative URLs should be returned in
	// Location headers
	RelativeURLs bool `yaml:"relativeurls,omitempty"`

	// Amount of time to wait for connection to drain before shutting down when registry
	// receives a stop signal
	DrainTimeout time.Duration `yaml:"draintimeout,omitempty"`

	// TLS instructs the http server to listen with a TLS configuration.
	// This only support simple tls configuration with a cert and key.
	// Mostly, this is useful for testing situations or simple deployments
	// that require tls. If more complex configurations are required, use
	// a proxy or make a proposal to add support here.
	TLS struct {
		// Certificate specifies the path to an x509 certificate file to
		// be used for TLS.
		Certificate string `yaml:"certificate,omitempty"`

		// Key specifies the path to the x509 key file, which should
		// contain the private portion for the file specified in
		// Certificate.
		Key string `yaml:"key,omitempty"`

		// Specifies the CA certs for client authentication
		// A file may contain multiple CA certificates encoded as PEM
		ClientCAs []string `yaml:"clientcas,omitempty"`

		// Specifies the lowest TLS version allowed
		MinimumTLS string `yaml:"minimumtls,omitempty"`

		// Specifies a list of cipher suites allowed
		CipherSuites []string `yaml:"ciphersuites,omitempty"`

		// LetsEncrypt is used to configuration setting up TLS through
		// Let's Encrypt instead of manually specifying certificate and
		// key. If a TLS certificate is specified, the Let's Encrypt
		// section will not be used.
		LetsEncrypt struct {
			// CacheFile specifies cache file to use for lets encrypt
			// certificates and keys.
			CacheFile string `yaml:"cachefile,omitempty"`

			// Email is the email to use during Let's Encrypt registration
			Email string `yaml:"email,omitempty"`

			// Hosts specifies the hosts which are allowed to obtain Let's
			// Encrypt certificates.
			Hosts []string `yaml:"hosts,omitempty"`
		} `yaml:"letsencrypt,omitempty"`
	} `yaml:"tls,omitempty"`

	// Headers is a set of headers to include in HTTP responses. A common
	// use case for this would be security headers such as
	// Strict-Transport-Security. The map keys are the header names, and
	// the values are the associated header payloads.
	Headers http.Header `yaml:"headers,omitempty"`

	// Debug configures the http debug interface, if specified. This can
	// include services such as pprof, expvar and other data that should
	// not be exposed externally. Left disabled by default.
	Debug struct {
		// Addr specifies the bind address for the debug server.
		Addr string `yaml:"addr,omitempty"`
		// Prometheus configures the Prometheus telemetry endpoint.
		Prometheus struct {
			Enabled bool   `yaml:"enabled,omitempty"`
			Path    string `yaml:"path,omitempty"`
		} `yaml:"prometheus,omitempty"`
	} `yaml:"debug,omitempty"`

	// HTTP2 configuration options
	HTTP2 struct {
		// Specifies whether the registry should disallow clients attempting
		// to connect via http2. If set to true, only http/1.1 is supported.
		Disabled bool `yaml:"disabled,omitempty"`
	} `yaml:"http2,omitempty"`
}

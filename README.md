# udistribution [![Go](https://github.com/kaovilai/udistribution/actions/workflows/go.yml/badge.svg)](https://github.com/kaovilai/udistribution/actions/workflows/go.yml)[![codecov](https://codecov.io/gh/kaovilai/udistribution/branch/main/graph/badge.svg?token=tmGT4hOtQb)](https://codecov.io/gh/kaovilai/udistribution)[![Total alerts](https://img.shields.io/lgtm/alerts/g/kaovilai/udistribution.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/kaovilai/udistribution/alerts/)[![Go Report Card](https://goreportcard.com/badge/github.com/kaovilai/udistribution)](https://goreportcard.com/report/github.com/kaovilai/udistribution)[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
Go library providing a client to interface with storage drivers of [distribution/distribution](https://github.com/distribution/distribution) without a listening HTTP server.

[Develop in gitpod](https://gitpod.io/#https://github.com/kaovilai/udistribution.git)

## Goal:
- Given a config and/or environment variables conforming to [available configurations](https://docs.docker.com/registry/configuration/)
  - a client interface can be initialized to access functions enabling access to methods in [distribution api spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api) without a listening registry HTTP Server by exposing ServeHTTP method.

Making it easier for Go programs to consume APIs on a needed basis without a listening server. This approach maybe more secure in an environment where it is not practical to obtain TLS certificates from a trusted certificate authorities, such as an unpredictable hostname/ip address.

Current functionality:
- [x] Modifies distribution/distribution to
  - [x] Initialize client with config string and/or environment variables
  - [x] ServeHTTP method can be accessed after initialization
- [x] Implement a function to register new transport type to "github.com/containers/image/v5/transports"
  - [x] Consumes distribution/distribution using exposed ServeHTTP method
  - [x] [End to end test](pkg/e2e_test.go) which demonstrates how the transport type is registered and can be used to access distribution/distribution

## Getting Started
Usage example as [seen in test](https://github.com/kaovilai/udistribution/blob/dd4070c5d75f4601e62d5a7b495a7ebd96b053f9/pkg/e2e_test.go#L45-L72)
```go
	ut, err := udistribution.NewTransportFromNewConfig("", os.Environ())
	defer ut.Deregister()
	if err != nil {
		t.Errorf("failed to create transport with environment variables: %v", err)
	}
	srcRef, err := docker.ParseReference("//alpine")
	if err != nil {
		t.Errorf("failed to parse reference: %v", err)
	}
	destRef, err := ut.ParseReference("//alpine")
	if err != nil {
		t.Errorf("failed to parse reference: %v", err)
	}
	pc, err := getPolicyContext()
	if err != nil {
		t.Errorf("failed to get policy context: %v", err)
	}
	ctx, err := getDefaultContext()
	if err != nil {
		t.Errorf("failed to get default context: %v", err)
	}
	_, err = copy.Image(context.Background(), pc, destRef, srcRef, &copy.Options{
		SourceCtx:      ctx,
		DestinationCtx: ctx,
	})
	if err != nil {
		t.Errorf("%v", errors.Wrapf(err, "failed to copy image"))
	}
```
First you call `NewClient` with a config string and environment variables.
Then you call the client's `ServeHTTP` method with a desired HTTP request.

You can use `httptest.NewRecorder` to record the response.

Alternatively, you may use alltransports.ParseImageName(ref) when transport name `ut.Name()://` is in the reference instead of using `ut.ParseReference`

## Known issues:
Prometheus metrics config must be disabled.

## NOTICE:
- This library contains some parts from [distribution/distribution](https://github.com/distribution/distribution) which is licensed under the Apache License 2.0.
  - Some parts has been modified to accommodate usage in this library.
  - A copy of the original distribution/distribution license is included in the repository at [LICENSE](LICENSE)

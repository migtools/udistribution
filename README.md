# udistribution [![Go](https://github.com/kaovilai/udistribution/actions/workflows/go.yml/badge.svg)](https://github.com/kaovilai/udistribution/actions/workflows/go.yml)[![codecov](https://codecov.io/gh/kaovilai/udistribution/branch/main/graph/badge.svg?token=tmGT4hOtQb)](https://codecov.io/gh/kaovilai/udistribution)[![Total alerts](https://img.shields.io/lgtm/alerts/g/kaovilai/udistribution.svg?logo=lgtm&logoWidth=18)](https://lgtm.com/projects/g/kaovilai/udistribution/alerts/)[![Go Report Card](https://goreportcard.com/badge/github.com/kaovilai/udistribution)](https://goreportcard.com/report/github.com/kaovilai/udistribution)[![License](https://img.shields.io/:license-apache-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
Go library providing a client to interface with storage drivers of [distribution/distribution](https://github.com/distribution/distribution) without a serving HTTP server.

Goal:
- Given a config and/or environment variables conforming to [available configurations](https://docs.docker.com/registry/configuration/)
  - a client interface can be initialized to access functions enabling access to methods in [distribution api spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api) without a running registry HTTP Server by communicating directly with [supported storage drivers](https://docs.docker.com/registry/storage-drivers/).

Initial priority:
s3, gcs, azure storage drivers

Known issues:
Prometheus metrics config must be disabled.

NOTICE:
- This library contains some parts from [distribution/distribution](https://github.com/distribution/distribution) which is licensed under the Apache License 2.0.
  - Some parts has been modified to accommodate usage in this library.
  - A copy of the original distribution/distribution license is included in the repository at [LICENSE](LICENSE)
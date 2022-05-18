# udistribution
Go library providing functions and interfaces for accessing distribution/distribution without a serving HTTP server.

Goal:
- A config conforming to https://github.com/distribution/distribution/blob/main/cmd/registry/config-example.yml and/or environment variables, a client interface can be initialized to access functions enabling access to methods in https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api without use of HTTP Server by communicating directly with [supported storage drivers](https://docs.docker.com/registry/storage-drivers/).

Initial priority:
s3, gcs, azure storage drivers

# udistribution
Go library providing functions and interfaces for accessing distribution/distribution without a serving HTTP server.

Goal:
- Given a config and/or environment variables conforming to [available configurations](https://docs.docker.com/registry/configuration/)
  - a client interface can be initialized to access functions enabling access to methods in [distribution api spec](https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api) without use of HTTP Server by communicating directly with [supported storage drivers](https://docs.docker.com/registry/storage-drivers/).

Initial priority:
s3, gcs, azure storage drivers

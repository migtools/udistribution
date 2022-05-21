Bits from container/image/docker modified for use with udistribution.

Modifications:
- Transport name renamed from docker to udistribution-docker
- Modified docker_client.go/makeRequestToResolvedURLOnce to use response recorder against distribution/distribution Registry App's ServeHTTP method instead of sending requests to a listening HTTP server.
# Target Go files
GO_TARGET ?= ./...
BUILD_TAGS ?= "include_gcs include_oss"
.PHONY: fmt
fmt:
	go fmt $(GO_TARGET)
.PHONY: test
test:
	@which pkg-config >/dev/null 2>&1 || (echo "Error: pkg-config is required for tests. Install with: brew install pkg-config (macOS), apt-get install pkg-config (Ubuntu), or dnf install pkgconfig (Fedora/RHEL)" && exit 1)
	@pkg-config --exists gpgme || (echo "Error: gpgme library is required for tests. Install with: brew install gpgme (macOS), apt-get install libgpgme-dev (Ubuntu), or dnf install gpgme-devel (Fedora/RHEL)" && exit 1)
	@# Only packages with test files can be covered; including no-test packages in
	@# -coverprofile triggers 'go: no such tool "covdata"' on newer Go toolchains.
	@# A go list failure (even one that still prints a partial package list) must
	@# fail the target rather than silently test a subset.
	@set -e; \
	TEST_PKGS=$$(go list -tags $(BUILD_TAGS) -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' $(GO_TARGET)) || { echo "Error: go list failed to load packages for $(GO_TARGET)"; exit 1; }; \
	if [ -n "$$TEST_PKGS" ]; then \
		go test -tags $(BUILD_TAGS) -v $$TEST_PKGS -coverprofile cover.out; \
	else \
		rm -f cover.out; \
		go test -tags $(BUILD_TAGS) -v $(GO_TARGET); \
	fi

.PHONY: build
build:
	go build -tags $(BUILD_TAGS) $(GO_TARGET)

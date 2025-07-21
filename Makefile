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
	go test -tags $(BUILD_TAGS) -v $(GO_TARGET) -coverprofile cover.out

.PHONY: build
build:
	go build -tags $(BUILD_TAGS) $(GO_TARGET)

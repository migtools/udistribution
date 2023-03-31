# Target Go files
GO_TARGET ?= ./...
BUILD_TAGS ?= "include_gcs include_oss"
.PHONY: fmt
fmt:
	go fmt $(GO_TARGET)
.PHONY: test
test:
	go test -tags $(BUILD_TAGS) -v $(GO_TARGET) -coverprofile cover.out

.PHONY: build
build:
	go build -tags $(BUILD_TAGS) $(GO_TARGET)

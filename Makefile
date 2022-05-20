# Target Go files
GO_TARGET ?= ./...

fmt:
	go fmt $(GO_TARGET)
test:
	go test $(GO_TARGET) -coverprofile cover.out
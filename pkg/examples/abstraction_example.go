package examples

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/migtools/udistribution/pkg/client"
)

// ExampleUsingAbstraction demonstrates how to use the new registry abstraction
func ExampleUsingAbstraction() {
	// Create client using the new abstraction layer
	clientV2, err := client.NewClientV2("", []string{
		"REGISTRY_STORAGE=filesystem",
		"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/tmp/registry-example",
		"REGISTRY_STORAGE_DELETE_ENABLED=true",
		"REGISTRY_LOG_LEVEL=info",
	})
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Health check
	ctx := context.Background()
	err = clientV2.Health(ctx)
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return
	}

	fmt.Println("Registry is healthy")

	// Example HTTP request
	req, _ := http.NewRequest("GET", "/v2/", nil)
	rr := httptest.NewRecorder()

	clientV2.ServeHTTP(rr, req)
	fmt.Printf("Registry responded with status: %d\n", rr.Code)

	// Shutdown when done
	err = clientV2.Shutdown(ctx)
	if err != nil {
		fmt.Printf("Shutdown failed: %v\n", err)
	}

	fmt.Println("Registry shutdown complete")
}

// ExampleBackwardCompatibility shows that the old client still works
func ExampleBackwardCompatibility() {
	// Old client still works
	oldClient, err := client.NewClient("", []string{
		"REGISTRY_STORAGE=filesystem",
		"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/tmp/registry-old",
	})
	if err != nil {
		fmt.Printf("Failed to create old client: %v\n", err)
		return
	}

	// Old client methods still work
	app := oldClient.GetApp()
	if app != nil {
		fmt.Println("Old client created successfully")
	}
}
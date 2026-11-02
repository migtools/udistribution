package client

import (
	"context"
	"testing"
)

func TestNewClientV2(t *testing.T) {
	// Test with minimal configuration - just create the client for now
	client, err := NewClientV2("", []string{
		"REGISTRY_STORAGE=filesystem",
		"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/tmp/registry",
		"REGISTRY_STORAGE_DELETE_ENABLED=true",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if client.registry == nil {
		t.Fatal("registry should not be nil")
	}

	if client.config == nil {
		t.Fatal("config should not be nil")
	}

	// For now, just test that the client was created successfully
	// TODO: Test ServeHTTP once we resolve the nil pointer issue
}

func TestClientV2_Health(t *testing.T) {
	client, err := NewClientV2("", []string{
		"REGISTRY_STORAGE=filesystem",
		"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/tmp/registry",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Health(ctx)
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

func TestClientV2_Shutdown(t *testing.T) {
	client, err := NewClientV2("", []string{
		"REGISTRY_STORAGE=filesystem",
		"REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY=/tmp/registry",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Shutdown(ctx)
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func TestParseConfigV2(t *testing.T) {
	configYaml := `
version: 0.1
storage:
  filesystem:
    rootdirectory: /tmp/registry
`
	config, err := parseConfigV2(configYaml, []string{
		"REGISTRY_STORAGE=s3",
		"REGISTRY_STORAGE_S3_BUCKET=test-bucket",
		"REGISTRY_STORAGE_S3_REGION=us-east-1",
		"REGISTRY_LOG_LEVEL=info",
	})
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if config.Storage.Type != "s3" {
		t.Errorf("expected storage type 's3', got %s", config.Storage.Type)
	}

	if config.Storage.Parameters["bucket"] != "test-bucket" {
		t.Errorf("expected bucket 'test-bucket', got %v", config.Storage.Parameters["bucket"])
	}

	if config.Storage.Parameters["region"] != "us-east-1" {
		t.Errorf("expected region 'us-east-1', got %v", config.Storage.Parameters["region"])
	}

	if config.Logging.Level != "info" {
		t.Errorf("expected log level 'info', got %s", config.Logging.Level)
	}

	if config.HTTP.Secret == "" {
		t.Error("expected secret to be generated")
	}
}
package pkg

import (
	"context"
	"os"
	"testing"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/types"
	"github.com/kaovilai/udistribution/pkg/image/udistribution"
	"github.com/pkg/errors"
)

/*
To test e2e, fill s3 credentials below and run:
export UDISTRIBUTION_TEST_E2E_ENABLE=true
export REGISTRY_STORAGE=s3
export REGISTRY_STORAGE_S3_BUCKET=<your-bucket>
export REGISTRY_STORAGE_S3_ACCESSKEY=<your-access-key>
export REGISTRY_STORAGE_S3_SECRETKEY=<your-secret-key>
export REGISTRY_STORAGE_S3_REGION=<your-region>
go test -v ./pkg/e2e_test.go

Note: This test will likely fail on macos due to lack of darwin container images.
udistribution/pkg/e2e_test.go:63: failed to copy image: choosing an image from manifest list docker://alpine:latest: no image found in manifest list for architecture amd64, variant "", OS darwin
*/
func TestE2e(t *testing.T) {
	t.Logf("TestE2e called")
	// Set test environment variables when running in IDE.
	// os.Setenv("UDISTRIBUTION_TEST_E2E_ENABLE", "true")
	// os.Setenv("REGISTRY_STORAGE", "s3")
	// os.Setenv("REGISTRY_STORAGE_S3_BUCKET", "")
	// os.Setenv("REGISTRY_STORAGE_S3_ACCESSKEY", "")
	// os.Setenv("REGISTRY_STORAGE_S3_SECRETKEY", "")
	// os.Setenv("REGISTRY_STORAGE_S3_REGION", "us-east-1")
	// only test if found key in env
	if os.Getenv("UDISTRIBUTION_TEST_E2E_ENABLE") == "" {
		t.Skip("UDISTRIBUTION_TEST_E2E_ENABLE not set, skipping e2e test")
	}
	if os.Getenv("REGISTRY_STORAGE") == "" {
		t.Skip("REGISTRY_STORAGE not set, skipping e2e test")
	}
	ut, err := udistribution.NewTransportFromNewConfig("", os.Environ())
	defer ut.Deregister()
	if err != nil {
		t.Errorf("failed to create transport with environment variables: %v", err)
	}
	srcRef, err := docker.ParseReference("//alpine")
	if err != nil {
		t.Errorf("failed to parse reference: %v", err)
	}
	destRef, err := ut.ParseReference("//alpine")
	if err != nil {
		t.Errorf("failed to parse reference: %v", err)
	}
	pc, err := getPolicyContext()
	if err != nil {
		t.Errorf("failed to get policy context: %v", err)
	}
	ctx, err := getDefaultContext()
	if err != nil {
		t.Errorf("failed to get default context: %v", err)
	}
	_, err = copy.Image(context.Background(), pc, destRef, srcRef, &copy.Options{
		SourceCtx:      ctx,
		DestinationCtx: ctx,
	})
	if err != nil {
		t.Errorf("%v", errors.Wrapf(err, "failed to copy image"))
	}
}

func getPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{Default: []signature.PolicyRequirement{signature.NewPRInsecureAcceptAnything()}}
	return signature.NewPolicyContext(policy)
}

func getDefaultContext() (*types.SystemContext, error) {
	ctx := &types.SystemContext{
		DockerDaemonInsecureSkipTLSVerify: true,
		DockerInsecureSkipTLSVerify:       types.OptionalBoolTrue,
		DockerDisableDestSchema1MIMETypes: true,
		OSChoice:                          "linux",
	}
	return ctx, nil
}

package pkg

import (
	"context"
	"fmt"
	"os"
	"testing"

	azstorage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports"
	"github.com/containers/image/v5/types"
	"github.com/distribution/distribution/v3/uuid"
	"github.com/kaovilai/udistribution/pkg/image/udistribution"
	"github.com/kaovilai/udistribution/pkg/internal/image/imagesource"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

/*
To test e2e, fill s3 credentials below and run:
export UDISTRIBUTION_TEST_E2E_ENABLE=true
export REGISTRY_STORAGE=s3
export REGISTRY_STORAGE_S3_BUCKET=<your-bucket>
export REGISTRY_STORAGE_S3_ACCESSKEY=<your-access-key>
export REGISTRY_STORAGE_S3_SECRETKEY=<your-secret-key>
export REGISTRY_STORAGE_S3_REGION=<your-region>
export REGISTRY_STORAGE_DELETE_ENABLED=true
go test -v ./pkg/e2e_test.go

Note: This test will likely fail on macos due to lack of darwin container images.
udistribution/pkg/e2e_test.go:63: failed to copy image: choosing an image from manifest list docker://alpine:latest: no image found in manifest list for architecture amd64, variant "", OS darwin
*/
func TestE2e(t *testing.T) {
	t.Logf("TestE2e called")
	// Set test environment variables when running in IDE.
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
	srcRef, err := docker.ParseReference("//quay.io/konveyor/openshift-velero-plugin:latest")
	if err != nil {
		t.Errorf("failed to parse reference: %v", err)
	}
	randomRepoName := "udistribution-e2e-test" + uuid.Generate().String()
	destRef, err := ut.ParseReference(fmt.Sprintf("//%s/openshift-velero-plugin:latest", randomRepoName))
	if err != nil {
		t.Errorf("failed to parse reference: %v", err)
	}
	// TODO: uncomment
	// Remove existing if any
	// storageParam := ut.GetApp().Config.Storage.Parameters()
	// err = destRef.DeleteImage(context.Background(), nil)
	// if err != nil {
	// 	// ignore unable to delete before copy.
	// 	if errors.Cause(err) != getUnableToDeleteError(udistribution.GetRef(destRef)) {
	// 		log.Printf("error isn't due to unable to delete: %v", getUnableToDeleteError(udistribution.GetRef(destRef)))
	// 		if storageParam["delete"] != nil {
	// 			deleteParam := storageParam["delete"].(map[string]bool)
	// 			if deleteParam["enabled"] == true {
	// 				t.Errorf("failed to delete image: %v", err)
	// 			} else {
	// 				t.Logf("delete disabled, skipping delete test")
	// 			}
	// 		}
	// 	}
	// }
	pc, err := getPolicyContext()
	if err != nil {
		t.Errorf("failed to get policy context: %v", err)
	}
	ctx, err := getDefaultContext()
	if err != nil {
		t.Errorf("failed to get default context: %v", err)
	}
	options := copy.Options{
		SourceCtx:      ctx,
		DestinationCtx: ctx,
	}
	_, err = copy.Image(context.Background(), pc, destRef, srcRef, &options)
	if err != nil {
		t.Errorf("%v", errors.Wrapf(err, "failed to copy image"))
	}
	// t.Errorf("fail here")
	// Cleanup
	// err = destRef.DeleteImage(context.Background(), nil)
	// if err != nil {
	// 	t.Errorf("failed to delete image: %v", err)
	// }

	// test that udistributionReference when used as source can call ConfigBlob() which is used during velero restore
	publicRawSource, err := destRef.NewImageSource(context.Background(), options.SourceCtx)
	if err != nil {
		t.Error(errors.Wrapf(err, "initializing source %s", transports.ImageName(destRef)))
	}
	rawSource := imagesource.FromPublic(publicRawSource)
	defer func() {
		if err := rawSource.Close(); err != nil {
			t.Errorf(" (src: %v)", err)
		}
	}()
	unparsedToplevel := image.UnparsedInstance(rawSource, nil)
	// targetInstance := unparsedToplevel // inside copyOneImage(ctx, policyContext, options, unparsedToplevel, unparsedToplevel, nil)
	unparsedImage := unparsedToplevel
	src, err := image.FromUnparsedImage(context.Background(), options.SourceCtx, unparsedImage)
	if err != nil {
		t.Error(errors.Wrapf(err, "initializing image from source %s", transports.ImageName(rawSource.Reference())))
	}
	// instanceDigest := targetInstance // inside copyUpdatedConfigAndManifest, and copyUpdatedConfigAndManifests
	// pendingImage := src // inside copyConfig(ctx, pendingImage)
	// pendingImage is now src again
	srcInfo := src.ConfigInfo()
	if srcInfo.Digest != "" {
		maxParallelDownloads := uint(6) // from var
		max := options.MaxParallelDownloads
		if max == 0 {
			max = maxParallelDownloads
		}
		concurrentBlobCopiesSemaphore := semaphore.NewWeighted(int64(max))
		if err := concurrentBlobCopiesSemaphore.Acquire(context.Background(), 1); err != nil {
			// This can only fail with ctx.Err(), so no need to blame acquiring the semaphore.
			_ = fmt.Errorf("copying config: %w", err)
		}
		defer concurrentBlobCopiesSemaphore.Release(1)

		func() { // A scope for defer
			// we don't care about progress bar
			// progressPool := c.newProgressPool()
			// defer progressPool.Wait()
			// bar := c.createProgressBar(progressPool, false, srcInfo, "config", "done")
			// defer bar.Abort(false)

			configBlob, err := src.ConfigBlob(context.Background())
			if err != nil {
				t.Error(errors.Wrapf(err, "reading config blob %s", srcInfo.Digest))
			}
			fmt.Printf("configBlob: %v", configBlob)
			// destInfo, err := c.copyBlobFromStream(ctx, bytes.NewReader(configBlob), srcInfo, nil, false, true, false, bar, -1, false)
			// if err != nil {
			// 	return types.BlobInfo{}, err
			// }

			// bar.mark100PercentComplete()
			// return destInfo, nil
		}()
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

// TODO:
// // https://github.com/containers/image/blob/857a813795f6a5dc8116912a86ba7956a315cd81/docker/docker_image_src.go#L638
// func getUnableToDeleteError(ref reference.Named) error {
// 	return errors.Errorf("Unable to delete %v. Image may not exist or is not stored with a v2 Schema in a v2 registry", ref)
// }

func TestS3Store(t *testing.T) {
	// test s3 driver
	validRegions := []string{
		"us-east-1",
		"eu-north-1",
		"af-south-1",
	}
	for _, region := range validRegions {
		ut, err := udistribution.NewTransportFromNewConfig("", []string{
			"REGISTRY_STORAGE=s3",
			"REGISTRY_STORAGE_S3_BUCKET=udistribution-test-e2e",
			"REGISTRY_STORAGE_S3_ACCESSKEY=<your-access-key>",
			"REGISTRY_STORAGE_S3_SECRETKEY=<your-secret-key>",
			"REGISTRY_STORAGE_S3_REGION=" + region,
		})
		if err != nil {
			t.Fatal(err)
		}
		ut.Deregister()
	}
}
func TestAzureStore(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("TestE2e recovered from panic: %v", r)
			if r.(azstorage.AzureStorageServiceError).Code == "AuthenticationFailed" {
				t.Log("azure storage driver initialized and authentication failed as expected")
			} else {
				t.Fatalf("TestE2e unexpected error: %v", r)
			}
		}
	}()
	ut, err := udistribution.NewTransportFromNewConfig("", []string{
		"REGISTRY_STORAGE=azure",
		"REGISTRY_STORAGE_AZURE_ACCOUNTNAME=accountname",
		"REGISTRY_STORAGE_AZURE_ACCOUNTKEY=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==", //sample key
		"REGISTRY_STORAGE_AZURE_CONTAINER=udistribution-test-e2e",
	})
	if err != nil {
		t.Fatal(err)
	}
	ut.Deregister()
}

func TestGCSStore(t *testing.T) {
	t.Skip("GCS is not supported yet")
	// TODO: test gcs driver
}

package pkg

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/image"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports"
	"github.com/containers/image/v5/types"
	"github.com/distribution/distribution/v3/uuid"
	"github.com/kaovilai/udistribution/pkg/image/udistribution"
	"github.com/kaovilai/udistribution/pkg/internal/image/imagesource"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	_, err = copy.Image(context.Background(), pc, destRef, srcRef, &copy.Options{
		SourceCtx:      ctx,
		DestinationCtx: ctx,
	})
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
	publicRawSource, err := destRef.NewImageSource(context.Background(), ctx)
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
	// Parse the list and get a copy of the original value after it's re-encoded.
	manifestList, manifestType, err := unparsedToplevel.Manifest(context.Background())
	if err != nil {
		t.Error(errors.Wrapf(err, "reading manifest list"))
	}
	originalList, err := manifest.ListFromBlob(manifestList, manifestType)
	if err != nil {
		t.Error(errors.Wrapf(err, "parsing manifest list %q", string(manifestList)))
	}
	options := copy.Options{
		SourceCtx:      ctx,
		DestinationCtx: ctx,
	}
	updatedList := originalList.Clone()
	instanceDigests := updatedList.Instances()
	imagesToCopy := len(instanceDigests)
	if options.ImageListSelection == copy.CopySpecificImages {
		imagesToCopy = len(options.Instances)
	}
	updates := make([]manifest.ListUpdate, len(instanceDigests))
	instancesCopied := 0
	for i, instanceDigest := range instanceDigests {
		if options.ImageListSelection == copy.CopySpecificImages {
			skip := true
			for _, instance := range options.Instances {
				if instance == instanceDigest {
					skip = false
					break
				}
			}
			if skip {
				update, err := updatedList.Instance(instanceDigest)
				if err != nil {
					t.Error(err)
				}
				logrus.Debugf("Skipping instance %s (%d/%d)", instanceDigest, i+1, len(instanceDigests))
				// Record the digest/size/type of the manifest that we didn't copy.
				updates[i] = update
				continue
			}
		}
		logrus.Debugf("Copying instance %s (%d/%d)", instanceDigest, i+1, len(instanceDigests))
		logrus.Printf("Copying image %s (%d/%d)\n", instanceDigest, instancesCopied+1, imagesToCopy)
		unparsedInstance := image.UnparsedInstance(rawSource, &instanceDigest)
		// updatedManifest, updatedManifestType, updatedManifestDigest, err := c.copyOneImage(ctx, policyContext, options, unparsedToplevel, unparsedInstance, &instanceDigest)
		// if err != nil {
		// 	return nil, err
		// }
		instancesCopied++
		// // Record the result of a possible conversion here.
		// update := manifest.ListUpdate{
		// 	Digest:    updatedManifestDigest,
		// 	Size:      int64(len(updatedManifest)),
		// 	MediaType: updatedManifestType,
		// }
		// updates[i] = update
		// targetInstance := instanceDigest
		// Now reset the digest/size/types of the manifests in the list to account for any conversions that we made.
		// if err = updatedList.UpdateInstances(updates); err != nil {
		// 	t.Error(errors.Wrapf(err, "updating manifest list"))
		// }
		// manifestBytes, retManifestDigest, err := ic.copyUpdatedConfigAndManifest(ctx, targetInstance)
		src, err := image.FromUnparsedImage(context.Background(), options.SourceCtx, unparsedInstance)
		if err != nil {
			t.Error(errors.Wrapf(err, "creating source image"))
		}
		pendingImage := src
		pendingImage.ConfigInfo()
		configBlob, err := src.ConfigBlob(context.Background())
		if err != nil {
			t.Error(errors.Wrapf(err, "getting config blob"))
		}
		logrus.Infof("config blob: %v", configBlob)
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

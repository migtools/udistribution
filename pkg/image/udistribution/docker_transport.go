package udistribution

import (
	"context"
	"fmt"
	"strings"

	"github.com/containers/image/v5/docker/policyconfiguration"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/transports"
	"github.com/distribution/distribution/v3/uuid"

	// "github.com/containers/image/v5/transports"
	"github.com/containers/image/v5/types"
	"github.com/migtools/udistribution/pkg/client"
	"github.com/migtools/udistribution/pkg/constants"
	"github.com/pkg/errors"
)

// User will initialize by calling InitializeTransport()
// func init() {
// 	transports.Register(Transport)
// }

// Transport is an ImageTransport for container registry-hosted images.
// var Transport = UdistributionTransport{}

type UdistributionTransport struct {
	*client.Client
	name string
	uuid string
}

// Create new transport and register.
// When you are done with this transport, use Deregister() to unregister it from available transports.
func NewTransport(client *client.Client, name string) *UdistributionTransport {
	t := UdistributionTransport{
		Client: client,
		name:   name,
		uuid:   uuid.Generate().String(),
	}
	if transports.Get(t.Name()) == nil {
		transports.Register(t)
	}
	return &t
}

// Create new transport with client params and register.
// When you are done with this transport, use Deregister() to unregister it from available transports.
func NewTransportFromNewConfig(config string, env []string) (*UdistributionTransport, error) {
	c, err := client.NewClient(config, env)
	if err != nil {
		return nil, err
	}
	t := UdistributionTransport{
		Client: c,
		name:   c.GetApp().Config.Storage.Type(),
		uuid:   uuid.Generate().String(),
	}
	if transports.Get(t.Name()) == nil {
		transports.Register(t)
	}
	return &t, nil
}

func (u UdistributionTransport) Deregister() {
	transports.Delete(u.Name())
}

func (t UdistributionTransport) Name() string {
	return constants.TransportPrefix + t.name + "-" + t.uuid
}

// ParseReference converts a string, which should not start with the ImageTransport.Name prefix, into an ImageReference.
func (t UdistributionTransport) ParseReference(reference string) (types.ImageReference, error) {
	return ParseReference(reference, &t)
}

// ValidatePolicyConfigurationScope checks that scope is a valid name for a signature.PolicyTransportScopes keys
// (i.e. a valid PolicyConfigurationIdentity() or PolicyConfigurationNamespaces() return value).
// It is acceptable to allow an invalid value which will never be matched, it can "only" cause user confusion.
// scope passed to this function will not be "", that value is always allowed.
func (t UdistributionTransport) ValidatePolicyConfigurationScope(scope string) error {
	// FIXME? We could be verifying the various character set and length restrictions
	// from docker/distribution/reference.regexp.go, but other than that there
	// are few semantically invalid strings.
	return nil
}

// udistributionReference is an ImageReference for Docker images.
type udistributionReference struct {
	ref reference.Named // By construction we know that !reference.IsNameOnly(ref)
	*UdistributionTransport
}

// for test
func GetRef(ir types.ImageReference) reference.Named {
	return ir.(udistributionReference).ref
}

// ParseReference converts a string, which should not start with the ImageTransport.Name prefix, into an Docker ImageReference.
func ParseReference(refString string, ut *UdistributionTransport) (types.ImageReference, error) {
	if !strings.HasPrefix(refString, "//") {
		return nil, errors.Errorf("docker: image reference %s does not start with //", refString)
	}
	ref, err := reference.ParseNormalizedNamed(strings.TrimPrefix(refString, "//"))
	if err != nil {
		return nil, err
	}
	ref = reference.TagNameOnly(ref)
	return NewReference(ref, ut)
}

// NewReference returns a Docker reference for a named reference. The reference must satisfy !reference.IsNameOnly().
func NewReference(ref reference.Named, ut *UdistributionTransport) (types.ImageReference, error) {
	return newReference(ref, ut)
}

// newReference returns a udistributionReference for a named reference.
func newReference(ref reference.Named, ut *UdistributionTransport) (udistributionReference, error) {
	if ut == nil {
		return udistributionReference{}, errors.Errorf("image reference is invalid: missing transport")
	}
	if reference.IsNameOnly(ref) {
		return udistributionReference{}, errors.Errorf("Docker reference %s has neither a tag nor a digest", reference.FamiliarString(ref))
	}
	// A github.com/distribution/reference value can have a tag and a digest at the same time!
	// The docker/distribution API does not really support that (we can’t ask for an image with a specific
	// tag and digest), so fail.  This MAY be accepted in the future.
	// (Even if it were supported, the semantics of policy namespaces are unclear - should we drop
	// the tag or the digest first?)
	_, isTagged := ref.(reference.NamedTagged)
	_, isDigested := ref.(reference.Canonical)
	if isTagged && isDigested {
		return udistributionReference{}, errors.Errorf("Docker references with both a tag and digest are currently not supported")
	}
	return udistributionReference{
		ref:                    ref,
		UdistributionTransport: ut,
	}, nil
}

func (ref udistributionReference) Transport() types.ImageTransport {
	return ref.UdistributionTransport
}

// StringWithinTransport returns a string representation of the reference, which MUST be such that
// reference.Transport().ParseReference(reference.StringWithinTransport()) returns an equivalent reference.
// NOTE: The returned string is not promised to be equal to the original input to ParseReference;
// e.g. default attribute values omitted by the user may be filled in in the return value, or vice versa.
// WARNING: Do not use the return value in the UI to describe an image, it does not contain the Transport().Name() prefix.
func (ref udistributionReference) StringWithinTransport() string {
	return "//" + reference.FamiliarString(ref.ref)
}

// DockerReference returns a Docker reference associated with this reference
// (fully explicit, i.e. !reference.IsNameOnly, but reflecting user intent,
// not e.g. after redirect or alias processing), or nil if unknown/not applicable.
func (ref udistributionReference) DockerReference() reference.Named {
	return ref.ref
}

// PolicyConfigurationIdentity returns a string representation of the reference, suitable for policy lookup.
// This MUST reflect user intent, not e.g. after processing of third-party redirects or aliases;
// The value SHOULD be fully explicit about its semantics, with no hidden defaults, AND canonical
// (i.e. various references with exactly the same semantics should return the same configuration identity)
// It is fine for the return value to be equal to StringWithinTransport(), and it is desirable but
// not required/guaranteed that it will be a valid input to Transport().ParseReference().
// Returns "" if configuration identities for these references are not supported.
func (ref udistributionReference) PolicyConfigurationIdentity() string {
	res, err := policyconfiguration.DockerReferenceIdentity(ref.ref)
	if res == "" || err != nil { // Coverage: Should never happen, NewReference above should refuse values which could cause a failure.
		panic(fmt.Sprintf("Internal inconsistency: policyconfiguration.DockerReferenceIdentity returned %#v, %v", res, err))
	}
	return res
}

// PolicyConfigurationNamespaces returns a list of other policy configuration namespaces to search
// for if explicit configuration for PolicyConfigurationIdentity() is not set.  The list will be processed
// in order, terminating on first match, and an implicit "" is always checked at the end.
// It is STRONGLY recommended for the first element, if any, to be a prefix of PolicyConfigurationIdentity(),
// and each following element to be a prefix of the element preceding it.
func (ref udistributionReference) PolicyConfigurationNamespaces() []string {
	return policyconfiguration.DockerReferenceNamespaces(ref.ref)
}

// NewImage returns a types.ImageCloser for this reference, possibly specialized for this ImageTransport.
// The caller must call .Close() on the returned ImageCloser.
// NOTE: If any kind of signature verification should happen, build an UnparsedImage from the value returned by NewImageSource,
// verify that UnparsedImage, and convert it into a real Image via image.FromUnparsedImage.
// WARNING: This may not do the right thing for a manifest list, see image.FromSource for details.
func (ref udistributionReference) NewImage(ctx context.Context, sys *types.SystemContext) (types.ImageCloser, error) {
	return newImage(ctx, sys, ref)
}

// NewImageSource returns a types.ImageSource for this reference.
// The caller must call .Close() on the returned ImageSource.
func (ref udistributionReference) NewImageSource(ctx context.Context, sys *types.SystemContext) (types.ImageSource, error) {
	return newImageSource(ctx, sys, ref)
}

// NewImageDestination returns a types.ImageDestination for this reference.
// The caller must call .Close() on the returned ImageDestination.
func (ref udistributionReference) NewImageDestination(ctx context.Context, sys *types.SystemContext) (types.ImageDestination, error) {
	return newImageDestination(sys, ref)
}

// DeleteImage deletes the named image from the registry, if supported.
func (ref udistributionReference) DeleteImage(ctx context.Context, sys *types.SystemContext) error {
	return deleteImage(ctx, sys, ref)
}

// tagOrDigest returns a tag or digest from the reference.
func (ref udistributionReference) tagOrDigest() (string, error) {
	if ref, ok := ref.ref.(reference.Canonical); ok {
		return ref.Digest().String(), nil
	}
	if ref, ok := ref.ref.(reference.NamedTagged); ok {
		return ref.Tag(), nil
	}
	// This should not happen, NewReference above refuses reference.IsNameOnly values.
	return "", errors.Errorf("Internal inconsistency: Reference %s unexpectedly has neither a digest nor a tag", reference.FamiliarString(ref.ref))
}

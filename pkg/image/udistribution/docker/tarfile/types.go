package tarfile

import (
	internal "github.com/kaovilai/udistribution/pkg/image/udistribution/docker/internal/tarfile"
)

// ManifestItem is an element of the array stored in the top-level manifest.json file.
type ManifestItem = internal.ManifestItem // All public members from the internal package remain accessible.

// This file implements file-center upload strategy planning: choose channel
// (direct|proxy) and encoding (single|multipart) from size, runtime config, and
// backend capabilities so clients can auto-select chunked transfer paths.

package file

import (
	"context"

	storagesvc "lina-core/internal/service/storage"
)

// Upload strategy channel and encoding constants (neutral, not vendor-specific).
const (
	// UploadChannelDirect means client transfers bytes to the object backend.
	UploadChannelDirect = "direct"
	// UploadChannelProxy means client transfers bytes through the host.
	UploadChannelProxy = "proxy"
	// UploadEncodingSingle means one object put (presigned or form whole-file).
	UploadEncodingSingle = "single"
	// UploadEncodingMultipart means object multipart / host-chunked upload.
	UploadEncodingMultipart = "multipart"
	// bytesPerMegabyte is the MiB unit used by upload size configuration.
	bytesPerMegabyte = int64(1024 * 1024)
)

// UploadStrategy describes the planned transfer path for one file upload.
type UploadStrategy struct {
	// Channel is direct or proxy.
	Channel string
	// Encoding is single or multipart.
	Encoding string
}

// UploadMultipartPlan carries part execution parameters for multipart encoding.
type UploadMultipartPlan struct {
	// PartSize is the preferred part size in bytes.
	PartSize int64
	// MinPartSize is the minimum intermediate part size in bytes.
	MinPartSize int64
	// MaxParts is a soft upper bound for part count (0 means unspecified).
	MaxParts int
	// MaxConcurrency is the suggested client parallel part count.
	MaxConcurrency int
}

// planUploadStrategy selects channel and encoding for one upload.
// proxy multipart is always available (host can assemble parts when cloud MP is not).
func planUploadStrategy(
	sizeBytes int64,
	multipartEnabled bool,
	thresholdMB int64,
	supportsDirect bool,
	supportsCloudMultipart bool,
) UploadStrategy {
	useMultipart := multipartEnabled &&
		thresholdMB > 0 &&
		sizeBytes >= thresholdMB*bytesPerMegabyte

	if !useMultipart {
		if supportsDirect {
			return UploadStrategy{Channel: UploadChannelDirect, Encoding: UploadEncodingSingle}
		}
		return UploadStrategy{Channel: UploadChannelProxy, Encoding: UploadEncodingSingle}
	}

	// Above threshold: prefer direct multipart when cloud multipart is available.
	if supportsDirect && supportsCloudMultipart {
		return UploadStrategy{Channel: UploadChannelDirect, Encoding: UploadEncodingMultipart}
	}
	// Otherwise host-mediated chunked upload (works for local and non-MP clouds).
	return UploadStrategy{Channel: UploadChannelProxy, Encoding: UploadEncodingMultipart}
}

// buildMultipartPlan returns part parameters for multipart encoding.
func buildMultipartPlan(partSizeMB int64, maxConcurrency int64) *UploadMultipartPlan {
	if partSizeMB < 5 {
		partSizeMB = 5
	}
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}
	return &UploadMultipartPlan{
		PartSize:       partSizeMB * bytesPerMegabyte,
		MinPartSize:    5 * bytesPerMegabyte,
		MaxParts:       10000,
		MaxConcurrency: int(maxConcurrency),
	}
}

// resolveCloudMultipartSupport probes NamespaceFiles multipart capability.
func (s *serviceImpl) resolveCloudMultipartSupport(ctx context.Context) (bool, error) {
	if s == nil || s.storage == nil {
		return false, nil
	}
	return s.storage.SupportsMultipart(ctx, storagesvc.NamespaceFiles)
}

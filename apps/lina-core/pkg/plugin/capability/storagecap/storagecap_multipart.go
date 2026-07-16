// This file defines optional object multipart-upload types for storage
// providers. Multipart lets large objects be written as ordered parts without
// requiring a single PutObject-sized transfer. Providers that do not implement
// MultipartUploadProvider report unsupported via SupportsMultipart helpers.

package storagecap

import (
	"context"
	"io"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
)

// MultipartUploadProvider is optionally implemented by object backends that can
// run cloud-style multipart upload sessions for scoped keys.
type MultipartUploadProvider interface {
	// SupportsMultipart reports whether the backend can run multipart uploads
	// under the current configuration.
	SupportsMultipart(ctx context.Context) bool
	// CreateMultipart starts one multipart upload for a scoped object key.
	CreateMultipart(ctx context.Context, in ProviderMultipartCreateInput) (*ProviderMultipartSession, error)
	// UploadPart writes one part of an in-flight multipart upload.
	UploadPart(ctx context.Context, in ProviderMultipartPartInput) (*ProviderMultipartPartResult, error)
	// CompleteMultipart assembles uploaded parts into the final object.
	CompleteMultipart(ctx context.Context, in ProviderMultipartCompleteInput) (*ProviderObject, error)
	// AbortMultipart aborts one multipart upload and releases provider resources.
	AbortMultipart(ctx context.Context, in ProviderMultipartAbortInput) error
	// CreateMultipartPartAccess issues short-lived client access for one part.
	// Backends that only support host-mediated UploadPart may return an error
	// with CodeStorageMultipartUnsupported.
	CreateMultipartPartAccess(ctx context.Context, in ProviderMultipartPartAccessInput) (*DirectAccess, error)
}

// ProviderMultipartCreateInput starts one provider-level multipart upload.
type ProviderMultipartCreateInput struct {
	// Key is the scoped provider object key.
	Key string
	// ContentType is the optional MIME type for the final object.
	ContentType string
	// Overwrite controls whether an existing object may be replaced when the
	// provider can encode that constraint at create time.
	Overwrite bool
}

// ProviderMultipartSession identifies one in-flight provider multipart upload.
type ProviderMultipartSession struct {
	// UploadID is the provider-issued multipart upload identifier.
	UploadID string
	// Key is the scoped provider object key.
	Key string
	// ProviderID identifies the backend that owns the session for diagnostics.
	ProviderID string
}

// ProviderMultipartPartInput writes one part.
type ProviderMultipartPartInput struct {
	// Key is the scoped provider object key (must match CreateMultipart).
	Key string
	// UploadID identifies the multipart session.
	UploadID string
	// PartNumber is a 1-based part index.
	PartNumber int32
	// Body carries the part payload.
	Body io.Reader
	// Size is the part size when known. Negative means unknown.
	Size int64
}

// ProviderMultipartPartResult acknowledges one uploaded part.
type ProviderMultipartPartResult struct {
	// PartNumber is the 1-based part index.
	PartNumber int32
	// ETag is the provider part etag required for CompleteMultipart.
	ETag string
}

// ProviderMultipartCompleteInput finishes one multipart upload.
type ProviderMultipartCompleteInput struct {
	// Key is the scoped provider object key.
	Key string
	// UploadID identifies the multipart session.
	UploadID string
	// Parts lists uploaded parts in ascending PartNumber order.
	Parts []ProviderMultipartCompletedPart
}

// ProviderMultipartCompletedPart is one part entry for CompleteMultipart.
type ProviderMultipartCompletedPart struct {
	// PartNumber is the 1-based part index.
	PartNumber int32
	// ETag is the part etag returned by UploadPart or the client.
	ETag string
}

// ProviderMultipartAbortInput aborts one multipart upload.
type ProviderMultipartAbortInput struct {
	// Key is the scoped provider object key.
	Key string
	// UploadID identifies the multipart session.
	UploadID string
}

// ProviderMultipartPartAccessInput issues client access for one part.
type ProviderMultipartPartAccessInput struct {
	// Key is the scoped provider object key.
	Key string
	// UploadID identifies the multipart session.
	UploadID string
	// PartNumber is the 1-based part index.
	PartNumber int32
	// Size is the expected part size when known. Negative means unknown.
	Size int64
	// ContentType is an optional MIME constraint when the provider encodes it.
	ContentType string
	// TTL bounds issued access lifetime. Zero uses provider default.
	TTL time.Duration
}

// MultipartCreateInput starts one plugin-visible multipart upload.
type MultipartCreateInput struct {
	// Path is the plugin-local logical object path.
	Path string
	// ContentType is the optional MIME type.
	ContentType string
	// Overwrite controls whether an existing object may be replaced.
	Overwrite bool
}

// MultipartCreateOutput identifies one plugin-visible multipart session.
type MultipartCreateOutput struct {
	// UploadID is an opaque session identifier for subsequent part calls.
	UploadID string
	// Path is the normalized plugin-local logical path.
	Path string
	// ProviderID identifies the active backend for diagnostics.
	ProviderID string
}

// MultipartPartInput writes one plugin-visible part through the host.
type MultipartPartInput struct {
	// Path is the plugin-local logical object path.
	Path string
	// UploadID identifies the multipart session.
	UploadID string
	// PartNumber is a 1-based part index.
	PartNumber int32
	// Body carries the part payload.
	Body io.Reader
	// Size is the part size when known. Negative means unknown.
	Size int64
}

// MultipartPartOutput acknowledges one uploaded part.
type MultipartPartOutput struct {
	// PartNumber is the 1-based part index.
	PartNumber int32
	// ETag is the part etag required for CompleteMultipart.
	ETag string
}

// MultipartCompleteInput finishes one plugin-visible multipart upload.
type MultipartCompleteInput struct {
	// Path is the plugin-local logical object path.
	Path string
	// UploadID identifies the multipart session.
	UploadID string
	// Parts lists uploaded parts in ascending PartNumber order.
	Parts []MultipartCompletedPart
}

// MultipartCompletedPart is one plugin-visible complete-part entry.
type MultipartCompletedPart struct {
	// PartNumber is the 1-based part index.
	PartNumber int32
	// ETag is the part etag.
	ETag string
}

// MultipartCompleteOutput returns metadata after a successful multipart complete.
type MultipartCompleteOutput struct {
	// Object is plugin-visible object metadata.
	Object *Object
}

// MultipartAbortInput aborts one plugin-visible multipart upload.
type MultipartAbortInput struct {
	// Path is the plugin-local logical object path.
	Path string
	// UploadID identifies the multipart session.
	UploadID string
}

// MultipartPartAccessInput issues client access for one plugin-visible part.
type MultipartPartAccessInput struct {
	// Path is the plugin-local logical object path.
	Path string
	// UploadID identifies the multipart session.
	UploadID string
	// PartNumber is the 1-based part index.
	PartNumber int32
	// Size is the expected part size when known. Negative means unknown.
	Size int64
	// ContentType is optional MIME type.
	ContentType string
	// TTL optionally bounds issued access lifetime.
	TTL time.Duration
}

// MultipartPartAccessOutput returns neutral client transfer access for one part.
type MultipartPartAccessOutput struct {
	// Access is the neutral client transfer description.
	Access *DirectAccess
	// Path is the normalized plugin-local logical path.
	Path string
}

// AsMultipartUploadProvider returns the optional multipart capability.
func AsMultipartUploadProvider(provider Provider) (MultipartUploadProvider, bool) {
	if provider == nil {
		return nil, false
	}
	multipart, ok := provider.(MultipartUploadProvider)
	return multipart, ok
}

// SupportsMultipart reports whether provider can run multipart uploads.
// Providers that do not implement MultipartUploadProvider return false.
func SupportsMultipart(ctx context.Context, provider Provider) bool {
	multipart, ok := AsMultipartUploadProvider(provider)
	if !ok || multipart == nil {
		return false
	}
	return multipart.SupportsMultipart(ctx)
}

// NewMultipartUnsupportedError returns a stable unsupported-multipart error.
func NewMultipartUnsupportedError() error {
	return bizerr.NewCode(CodeStorageMultipartUnsupported)
}

// NewMultipartSessionInvalidError returns a stable invalid-session error.
func NewMultipartSessionInvalidError() error {
	return bizerr.NewCode(CodeStorageMultipartSessionInvalid)
}

// NewMultipartPartInvalidError returns a stable invalid-part error.
func NewMultipartPartInvalidError() error {
	return bizerr.NewCode(CodeStorageMultipartPartInvalid)
}

// NormalizeMultipartUploadID trims a multipart upload identifier.
func NormalizeMultipartUploadID(uploadID string) string {
	return strings.TrimSpace(uploadID)
}

// ValidateMultipartPartNumber reports whether partNumber is a valid 1-based index.
func ValidateMultipartPartNumber(partNumber int32) bool {
	return partNumber >= 1
}

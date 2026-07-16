// This file implements plugin-visible multipart upload methods on the storage
// adapter. Multipart is an optional provider capability; unsupported backends
// return a stable unsupported error so callers can fall back to single Put.

package capabilityhost

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/storagecap"
)

// SupportsMultipart reports whether the active backend supports multipart.
func (s *storageAdapter) SupportsMultipart(ctx context.Context) (bool, error) {
	if err := s.validateServiceScope(); err != nil {
		return false, err
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return false, err
	}
	return storagecap.SupportsMultipart(ctx, provider), nil
}

// CreateMultipart starts one plugin-scoped multipart upload.
func (s *storageAdapter) CreateMultipart(ctx context.Context, in storagecap.MultipartCreateInput) (*storagecap.MultipartCreateOutput, error) {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	session, err := multipart.CreateMultipart(ctx, storagecap.ProviderMultipartCreateInput{
		Key:         s.objectKey(ctx, objectPath),
		ContentType: strings.TrimSpace(in.ContentType),
		Overwrite:   in.Overwrite,
	})
	if err != nil {
		return nil, err
	}
	if session == nil || strings.TrimSpace(session.UploadID) == "" {
		return nil, storagecap.NewMultipartSessionInvalidError()
	}
	if strings.TrimSpace(session.ProviderID) == "" {
		session.ProviderID = providerID
	}
	return &storagecap.MultipartCreateOutput{
		UploadID:   session.UploadID,
		Path:       objectPath,
		ProviderID: session.ProviderID,
	}, nil
}

// UploadPart writes one part through the active provider.
func (s *storageAdapter) UploadPart(ctx context.Context, in storagecap.MultipartPartInput) (*storagecap.MultipartPartOutput, error) {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return nil, err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" || !storagecap.ValidateMultipartPartNumber(in.PartNumber) {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	result, err := multipart.UploadPart(ctx, storagecap.ProviderMultipartPartInput{
		Key:        s.objectKey(ctx, objectPath),
		UploadID:   uploadID,
		PartNumber: in.PartNumber,
		Body:       in.Body,
		Size:       in.Size,
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	return &storagecap.MultipartPartOutput{
		PartNumber: result.PartNumber,
		ETag:       result.ETag,
	}, nil
}

// CompleteMultipart assembles parts into the final plugin object.
func (s *storageAdapter) CompleteMultipart(ctx context.Context, in storagecap.MultipartCompleteInput) (*storagecap.MultipartCompleteOutput, error) {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return nil, err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" || len(in.Parts) == 0 {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	parts := make([]storagecap.ProviderMultipartCompletedPart, 0, len(in.Parts))
	for _, part := range in.Parts {
		if !storagecap.ValidateMultipartPartNumber(part.PartNumber) || strings.TrimSpace(part.ETag) == "" {
			return nil, storagecap.NewMultipartPartInvalidError()
		}
		parts = append(parts, storagecap.ProviderMultipartCompletedPart{
			PartNumber: part.PartNumber,
			ETag:       strings.TrimSpace(part.ETag),
		})
	}
	object, err := multipart.CompleteMultipart(ctx, storagecap.ProviderMultipartCompleteInput{
		Key:      s.objectKey(ctx, objectPath),
		UploadID: uploadID,
		Parts:    parts,
	})
	if err != nil {
		return nil, err
	}
	if object == nil {
		return nil, bizerr.NewCode(storagecap.CodeStorageMultipartCompleteFailed)
	}
	return &storagecap.MultipartCompleteOutput{
		Object: s.providerObject(objectPath, providerID, object),
	}, nil
}

// AbortMultipart aborts one plugin-scoped multipart upload.
func (s *storageAdapter) AbortMultipart(ctx context.Context, in storagecap.MultipartAbortInput) error {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" {
		return storagecap.NewMultipartSessionInvalidError()
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return err
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return storagecap.NewMultipartUnsupportedError()
	}
	return multipart.AbortMultipart(ctx, storagecap.ProviderMultipartAbortInput{
		Key:      s.objectKey(ctx, objectPath),
		UploadID: uploadID,
	})
}

// CreateMultipartPartAccess issues client access for one multipart part.
func (s *storageAdapter) CreateMultipartPartAccess(ctx context.Context, in storagecap.MultipartPartAccessInput) (*storagecap.MultipartPartAccessOutput, error) {
	objectPath, err := s.normalizeObjectPath(in.Path)
	if err != nil {
		return nil, err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" || !storagecap.ValidateMultipartPartNumber(in.PartNumber) {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, err
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	access, err := multipart.CreateMultipartPartAccess(ctx, storagecap.ProviderMultipartPartAccessInput{
		Key:         s.objectKey(ctx, objectPath),
		UploadID:    uploadID,
		PartNumber:  in.PartNumber,
		Size:        in.Size,
		ContentType: strings.TrimSpace(in.ContentType),
		TTL:         in.TTL,
	})
	if err != nil {
		return nil, err
	}
	if access == nil {
		return nil, storagecap.NewDirectAccessIssueFailedError()
	}
	if strings.TrimSpace(access.ProviderID) == "" {
		access.ProviderID = providerID
	}
	return &storagecap.MultipartPartAccessOutput{Access: access, Path: objectPath}, nil
}

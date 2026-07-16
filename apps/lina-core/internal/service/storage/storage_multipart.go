// This file implements host-internal multipart helpers on the local filesystem
// backend (always unsupported) and the resolving Service (NamespaceFiles only).

package storage

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/storagecap"
)

// SupportsMultipart always returns false for the local filesystem backend.
func (s *serviceImpl) SupportsMultipart(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// CreateMultipart is unsupported on the local filesystem backend.
func (s *serviceImpl) CreateMultipart(_ context.Context, _ MultipartCreateInput) (*MultipartCreateOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}

// UploadPart is unsupported on the local filesystem backend.
func (s *serviceImpl) UploadPart(_ context.Context, _ MultipartPartInput) (*MultipartPartOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}

// CompleteMultipart is unsupported on the local filesystem backend.
func (s *serviceImpl) CompleteMultipart(_ context.Context, _ MultipartCompleteInput) (*MultipartCompleteOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}

// AbortMultipart is unsupported on the local filesystem backend.
func (s *serviceImpl) AbortMultipart(_ context.Context, _ MultipartAbortInput) error {
	return storagecap.NewMultipartUnsupportedError()
}

// CreateMultipartPartAccess is unsupported on the local filesystem backend.
func (s *serviceImpl) CreateMultipartPartAccess(_ context.Context, _ MultipartPartAccessInput) (*MultipartPartAccessOutput, error) {
	return nil, storagecap.NewMultipartUnsupportedError()
}

// SupportsMultipart reports whether NamespaceFiles can use cloud multipart.
func (s *resolvingService) SupportsMultipart(ctx context.Context, namespace string) (bool, error) {
	if !isFilesNamespace(namespace) {
		return false, nil
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return false, mapResolveError(err)
	}
	return storagecap.SupportsMultipart(ctx, provider), nil
}

// CreateMultipart starts one NamespaceFiles multipart upload on the active provider.
func (s *resolvingService) CreateMultipart(ctx context.Context, in MultipartCreateInput) (*MultipartCreateOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return nil, err
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	session, err := multipart.CreateMultipart(ctx, storagecap.ProviderMultipartCreateInput{
		Key:         providerKey,
		ContentType: strings.TrimSpace(in.ContentType),
		Overwrite:   in.Overwrite,
	})
	if err != nil {
		return nil, mapResolveError(err)
	}
	if session == nil || strings.TrimSpace(session.UploadID) == "" {
		return nil, storagecap.NewMultipartSessionInvalidError()
	}
	if strings.TrimSpace(session.ProviderID) == "" {
		session.ProviderID = providerID
	}
	return &MultipartCreateOutput{
		UploadID:    session.UploadID,
		ProviderID:  session.ProviderID,
		ProviderKey: providerKey,
	}, nil
}

// UploadPart writes one NamespaceFiles part.
func (s *resolvingService) UploadPart(ctx context.Context, in MultipartPartInput) (*MultipartPartOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return nil, err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" || !storagecap.ValidateMultipartPartNumber(in.PartNumber) {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	result, err := multipart.UploadPart(ctx, storagecap.ProviderMultipartPartInput{
		Key:        providerKey,
		UploadID:   uploadID,
		PartNumber: in.PartNumber,
		Body:       in.Body,
		Size:       in.Size,
	})
	if err != nil {
		return nil, mapResolveError(err)
	}
	if result == nil {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	return &MultipartPartOutput{PartNumber: result.PartNumber, ETag: result.ETag}, nil
}

// CompleteMultipart finishes one NamespaceFiles multipart upload.
func (s *resolvingService) CompleteMultipart(ctx context.Context, in MultipartCompleteInput) (*MultipartCompleteOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return nil, err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" || len(in.Parts) == 0 {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
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
		Key:      providerKey,
		UploadID: uploadID,
		Parts:    parts,
	})
	if err != nil {
		return nil, mapResolveError(err)
	}
	return &MultipartCompleteOutput{Object: objectFromProvider(object, in.Key)}, nil
}

// AbortMultipart aborts one NamespaceFiles multipart upload.
func (s *resolvingService) AbortMultipart(ctx context.Context, in MultipartAbortInput) error {
	if !isFilesNamespace(in.Namespace) {
		return storagecap.NewMultipartUnsupportedError()
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" {
		return storagecap.NewMultipartSessionInvalidError()
	}
	_, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return mapResolveError(err)
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return storagecap.NewMultipartUnsupportedError()
	}
	return mapResolveError(multipart.AbortMultipart(ctx, storagecap.ProviderMultipartAbortInput{
		Key:      providerKey,
		UploadID: uploadID,
	}))
}

// CreateMultipartPartAccess issues client access for one NamespaceFiles part.
func (s *resolvingService) CreateMultipartPartAccess(ctx context.Context, in MultipartPartAccessInput) (*MultipartPartAccessOutput, error) {
	if !isFilesNamespace(in.Namespace) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	providerKey, err := filesProviderKey(in.Key)
	if err != nil {
		return nil, err
	}
	uploadID := storagecap.NormalizeMultipartUploadID(in.UploadID)
	if uploadID == "" || !storagecap.ValidateMultipartPartNumber(in.PartNumber) {
		return nil, storagecap.NewMultipartPartInvalidError()
	}
	providerID, provider, err := storagecap.ResolveProvider(ctx, s.runtime, s.localProvider)
	if err != nil {
		return nil, mapResolveError(err)
	}
	multipart, ok := storagecap.AsMultipartUploadProvider(provider)
	if !ok || multipart == nil || !multipart.SupportsMultipart(ctx) {
		return nil, storagecap.NewMultipartUnsupportedError()
	}
	access, err := multipart.CreateMultipartPartAccess(ctx, storagecap.ProviderMultipartPartAccessInput{
		Key:         providerKey,
		UploadID:    uploadID,
		PartNumber:  in.PartNumber,
		Size:        in.Size,
		ContentType: strings.TrimSpace(in.ContentType),
		TTL:         in.TTL,
	})
	if err != nil {
		return nil, mapResolveError(err)
	}
	if access == nil {
		return nil, storagecap.NewDirectAccessIssueFailedError()
	}
	if strings.TrimSpace(access.ProviderID) == "" {
		access.ProviderID = providerID
	}
	return &MultipartPartAccessOutput{
		Access:      access,
		ProviderID:  providerID,
		ProviderKey: providerKey,
	}, nil
}

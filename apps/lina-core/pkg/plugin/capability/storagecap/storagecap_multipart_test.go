// This file tests optional multipart provider probing helpers.

package storagecap

import (
	"context"
	"io"
	"testing"
)

type multipartTestProvider struct {
	support bool
}

func (p *multipartTestProvider) Put(context.Context, ProviderPutInput) (*ProviderObject, error) {
	return nil, nil
}
func (p *multipartTestProvider) Get(context.Context, ProviderGetInput) (*ProviderGetOutput, error) {
	return nil, nil
}
func (p *multipartTestProvider) Delete(context.Context, ProviderDeleteInput) error { return nil }
func (p *multipartTestProvider) DeleteMany(context.Context, ProviderDeleteManyInput) error {
	return nil
}
func (p *multipartTestProvider) List(context.Context, ProviderListInput) (*ProviderListOutput, error) {
	return nil, nil
}
func (p *multipartTestProvider) ListCursor(context.Context, ProviderListCursorInput) (*ProviderListCursorOutput, error) {
	return nil, nil
}
func (p *multipartTestProvider) Stat(context.Context, ProviderStatInput) (*ProviderStatOutput, error) {
	return nil, nil
}
func (p *multipartTestProvider) BatchStat(context.Context, ProviderBatchStatInput) (*ProviderBatchStatOutput, error) {
	return nil, nil
}
func (p *multipartTestProvider) SupportsMultipart(context.Context) bool { return p.support }
func (p *multipartTestProvider) CreateMultipart(context.Context, ProviderMultipartCreateInput) (*ProviderMultipartSession, error) {
	return &ProviderMultipartSession{UploadID: "u1", Key: "k"}, nil
}
func (p *multipartTestProvider) UploadPart(context.Context, ProviderMultipartPartInput) (*ProviderMultipartPartResult, error) {
	return &ProviderMultipartPartResult{PartNumber: 1, ETag: "e1"}, nil
}
func (p *multipartTestProvider) CompleteMultipart(context.Context, ProviderMultipartCompleteInput) (*ProviderObject, error) {
	return &ProviderObject{Key: "k", Size: 1}, nil
}
func (p *multipartTestProvider) AbortMultipart(context.Context, ProviderMultipartAbortInput) error {
	return nil
}
func (p *multipartTestProvider) CreateMultipartPartAccess(context.Context, ProviderMultipartPartAccessInput) (*DirectAccess, error) {
	return &DirectAccess{Mode: DirectAccessModePresignedURL, Method: "PUT", URL: "https://example.test/part"}, nil
}

func TestSupportsMultipartPlainProvider(t *testing.T) {
	t.Parallel()
	if SupportsMultipart(context.Background(), plainProvider{}) {
		t.Fatal("plain provider must not support multipart")
	}
}

func TestSupportsMultipartOptionalProvider(t *testing.T) {
	t.Parallel()
	if SupportsMultipart(context.Background(), &multipartTestProvider{support: false}) {
		t.Fatal("expected false when SupportsMultipart returns false")
	}
	if !SupportsMultipart(context.Background(), &multipartTestProvider{support: true}) {
		t.Fatal("expected true when SupportsMultipart returns true")
	}
}

func TestValidateMultipartPartNumber(t *testing.T) {
	t.Parallel()
	if ValidateMultipartPartNumber(0) || ValidateMultipartPartNumber(-1) {
		t.Fatal("part numbers must be 1-based")
	}
	if !ValidateMultipartPartNumber(1) {
		t.Fatal("part number 1 must be valid")
	}
}

// Ensure Body io.Reader is referenced for compile stability of test helpers.
var _ io.Reader

// This file tests client direct-access helpers and optional provider probing.

package storagecap

import (
	"context"
	"io"
	"testing"
	"time"
)

type directAccessTestProvider struct {
	support bool
	access  *DirectAccess
	err     error
	lastIn  ProviderDirectAccessInput
}

func (p *directAccessTestProvider) Put(context.Context, ProviderPutInput) (*ProviderObject, error) {
	return nil, nil
}
func (p *directAccessTestProvider) Get(context.Context, ProviderGetInput) (*ProviderGetOutput, error) {
	return nil, nil
}
func (p *directAccessTestProvider) Delete(context.Context, ProviderDeleteInput) error { return nil }
func (p *directAccessTestProvider) DeleteMany(context.Context, ProviderDeleteManyInput) error {
	return nil
}
func (p *directAccessTestProvider) List(context.Context, ProviderListInput) (*ProviderListOutput, error) {
	return nil, nil
}
func (p *directAccessTestProvider) ListCursor(context.Context, ProviderListCursorInput) (*ProviderListCursorOutput, error) {
	return nil, nil
}
func (p *directAccessTestProvider) Stat(context.Context, ProviderStatInput) (*ProviderStatOutput, error) {
	return nil, nil
}
func (p *directAccessTestProvider) BatchStat(context.Context, ProviderBatchStatInput) (*ProviderBatchStatOutput, error) {
	return nil, nil
}
func (p *directAccessTestProvider) SupportsDirectAccess(context.Context, DirectAccessOperation) bool {
	return p.support
}
func (p *directAccessTestProvider) CreateDirectAccess(_ context.Context, in ProviderDirectAccessInput) (*DirectAccess, error) {
	p.lastIn = in
	if p.err != nil {
		return nil, p.err
	}
	return p.access, nil
}

type plainProvider struct{}

func (plainProvider) Put(context.Context, ProviderPutInput) (*ProviderObject, error) { return nil, nil }
func (plainProvider) Get(context.Context, ProviderGetInput) (*ProviderGetOutput, error) {
	return nil, nil
}
func (plainProvider) Delete(context.Context, ProviderDeleteInput) error     { return nil }
func (plainProvider) DeleteMany(context.Context, ProviderDeleteManyInput) error {
	return nil
}
func (plainProvider) List(context.Context, ProviderListInput) (*ProviderListOutput, error) {
	return nil, nil
}
func (plainProvider) ListCursor(context.Context, ProviderListCursorInput) (*ProviderListCursorOutput, error) {
	return nil, nil
}
func (plainProvider) Stat(context.Context, ProviderStatInput) (*ProviderStatOutput, error) {
	return nil, nil
}
func (plainProvider) BatchStat(context.Context, ProviderBatchStatInput) (*ProviderBatchStatOutput, error) {
	return nil, nil
}

func TestSupportsDirectAccessPlainProvider(t *testing.T) {
	t.Parallel()
	if SupportsDirectAccess(context.Background(), plainProvider{}, DirectAccessOpPut) {
		t.Fatal("plain provider must not support direct access")
	}
}

func TestCreateDirectAccessProxyWhenUnsupported(t *testing.T) {
	t.Parallel()
	access, err := CreateDirectAccess(context.Background(), "local", plainProvider{}, ProviderDirectAccessInput{
		Key:       "files/a",
		Operation: DirectAccessOpPut,
	})
	if err != nil {
		t.Fatalf("CreateDirectAccess: %v", err)
	}
	if !IsProxyDirectAccess(access) {
		t.Fatalf("mode=%q want proxy", access.Mode)
	}
}

func TestCreateDirectAccessPresigned(t *testing.T) {
	t.Parallel()
	expires := time.Now().Add(10 * time.Minute).UTC()
	provider := &directAccessTestProvider{
		support: true,
		access: &DirectAccess{
			Mode:      DirectAccessModePresignedURL,
			Method:    "PUT",
			URL:       "https://example.test/object",
			ExpiresAt: expires,
		},
	}
	access, err := CreateDirectAccess(context.Background(), "linapro-storage-s3", provider, ProviderDirectAccessInput{
		Key:         "files/1/a.bin",
		Operation:   DirectAccessOpPut,
		Size:        12,
		ContentType: "application/octet-stream",
		TTL:         5 * time.Minute,
		Overwrite:   true,
	})
	if err != nil {
		t.Fatalf("CreateDirectAccess: %v", err)
	}
	if access.Mode != DirectAccessModePresignedURL || access.URL == "" {
		t.Fatalf("unexpected access: %+v", access)
	}
	if access.ProviderID != "linapro-storage-s3" {
		t.Fatalf("provider id=%q", access.ProviderID)
	}
	if provider.lastIn.Key != "files/1/a.bin" || provider.lastIn.Size != 12 {
		t.Fatalf("input not forwarded: %+v", provider.lastIn)
	}
}

func TestCreateDirectAccessInvalidOperation(t *testing.T) {
	t.Parallel()
	_, err := CreateDirectAccess(context.Background(), "local", plainProvider{}, ProviderDirectAccessInput{
		Key:       "k",
		Operation: "delete",
	})
	if err == nil {
		t.Fatal("expected invalid operation error")
	}
}

// Ensure test provider still satisfies the plain Provider contract.
var (
	_ Provider             = (*directAccessTestProvider)(nil)
	_ DirectAccessProvider = (*directAccessTestProvider)(nil)
	_ Provider             = plainProvider{}
	_ io.Reader            = nil
)

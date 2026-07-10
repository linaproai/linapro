// This file verifies generic plugin-owned capability descriptor registration,
// owner/method indexing, deterministic snapshots, and duplicate rejection.

package capregistry

import (
	"context"
	"strings"
	"testing"
)

func TestRegistryRegistersDescriptorAndIndexesMethods(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	descriptor := testDescriptor("linapro-ai-core", "ai", "v1", "text.generate", "text.method_status.get")
	if err := registry.Register(descriptor); err != nil {
		t.Fatalf("register descriptor: %v", err)
	}

	stored, ok := registry.LookupDescriptor("linapro-ai-core", "ai", "v1")
	if !ok {
		t.Fatal("expected descriptor to be indexed")
	}
	if stored.OwnerPluginID != "linapro-ai-core" || stored.Service != "ai" || stored.Version != "v1" {
		t.Fatalf("unexpected descriptor key: %#v", stored)
	}
	if len(stored.Methods) != 2 {
		t.Fatalf("expected two methods, got %#v", stored.Methods)
	}

	method, ok := registry.LookupMethod("linapro-ai-core", "ai", "v1", "text.generate")
	if !ok {
		t.Fatal("expected method to be indexed")
	}
	if method.OwnerPluginID != "linapro-ai-core" || method.Descriptor.Method != "text.generate" {
		t.Fatalf("unexpected indexed method: %#v", method)
	}
	if method.Descriptor.Capability != "host:ai:text.generate" {
		t.Fatalf("unexpected method capability: %#v", method.Descriptor)
	}
}

func TestRegistryIndexesMethodInvoker(t *testing.T) {
	t.Parallel()

	invoker := testInvoker{}
	descriptor := testDescriptor("linapro-ai-core", "ai", "v1", "text.generate")
	descriptor.Invoker = invoker
	registry := NewRegistry()
	if err := registry.Register(descriptor); err != nil {
		t.Fatalf("register descriptor: %v", err)
	}

	method, ok := registry.LookupMethod("linapro-ai-core", "ai", "v1", "text.generate")
	if !ok {
		t.Fatal("expected method to be indexed")
	}
	if method.Invoker == nil {
		t.Fatal("expected method invoker to be retained")
	}
	result, err := method.Invoker.Invoke(context.Background(), Invocation{})
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
	if string(result.Payload) != "ok" {
		t.Fatalf("unexpected invoker payload: %#v", result)
	}
}

func TestRegistrySnapshotsAreSortedCopies(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	for _, descriptor := range []Descriptor{
		testDescriptor("owner-b", "workflow", "v1", "run.execute"),
		testDescriptor("owner-a", "ai", "v1", "text.generate"),
	} {
		if err := registry.Register(descriptor); err != nil {
			t.Fatalf("register descriptor: %v", err)
		}
	}

	descriptors := registry.Descriptors()
	if len(descriptors) != 2 {
		t.Fatalf("expected two descriptors, got %#v", descriptors)
	}
	if descriptors[0].OwnerPluginID != "owner-a" || descriptors[1].OwnerPluginID != "owner-b" {
		t.Fatalf("expected sorted descriptors, got %#v", descriptors)
	}
	descriptors[0].Methods[0].Method = "mutated"
	if method, ok := registry.LookupMethod("owner-a", "ai", "v1", "text.generate"); !ok || method.Descriptor.Method != "text.generate" {
		t.Fatalf("expected registry state to be immutable from snapshot mutation, got %#v ok=%v", method, ok)
	}

	methods := registry.Methods()
	if len(methods) != 2 {
		t.Fatalf("expected two methods, got %#v", methods)
	}
	if methods[0].OwnerPluginID != "owner-a" || methods[1].OwnerPluginID != "owner-b" {
		t.Fatalf("expected sorted methods, got %#v", methods)
	}
}

func TestRegistryRejectsDuplicateDescriptorAndMethods(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	descriptor := testDescriptor("linapro-ai-core", "ai", "v1", "text.generate")
	if err := registry.Register(descriptor); err != nil {
		t.Fatalf("register descriptor: %v", err)
	}
	if err := registry.Register(descriptor); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("expected duplicate descriptor rejection, got %v", err)
	}

	err := NewRegistry().Register(testDescriptor("linapro-ai-core", "ai", "v1", "text.generate", "text.generate"))
	if err == nil || !strings.Contains(err.Error(), "duplicate method") {
		t.Fatalf("expected duplicate method rejection, got %v", err)
	}
}

func TestRegistryRejectsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	valid := testDescriptor("linapro-ai-core", "ai", "v1", "text.generate")
	tests := []struct {
		name   string
		mutate func(*Descriptor)
		want   string
	}{
		{
			name:   "missing owner",
			mutate: func(descriptor *Descriptor) { descriptor.OwnerPluginID = " " },
			want:   "owner plugin id is required",
		},
		{
			name:   "missing service",
			mutate: func(descriptor *Descriptor) { descriptor.Service = " " },
			want:   "service is required",
		},
		{
			name:   "missing version",
			mutate: func(descriptor *Descriptor) { descriptor.Version = " " },
			want:   "version is required",
		},
		{
			name:   "missing methods",
			mutate: func(descriptor *Descriptor) { descriptor.Methods = nil },
			want:   "methods are required",
		},
		{
			name:   "missing method name",
			mutate: func(descriptor *Descriptor) { descriptor.Methods[0].Method = " " },
			want:   "method is required",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			descriptor := valid
			descriptor.Methods = append([]MethodDescriptor(nil), valid.Methods...)
			tc.mutate(&descriptor)
			err := NewRegistry().Register(descriptor)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestNilRegistryRejectsRegistrationAndReadsEmpty(t *testing.T) {
	t.Parallel()

	var registry *Registry
	if err := registry.Register(testDescriptor("linapro-ai-core", "ai", "v1", "text.generate")); err == nil {
		t.Fatal("expected nil registry registration to fail")
	}
	if _, ok := registry.LookupDescriptor("linapro-ai-core", "ai", "v1"); ok {
		t.Fatal("expected nil registry descriptor lookup to miss")
	}
	if _, ok := registry.LookupMethod("linapro-ai-core", "ai", "v1", "text.generate"); ok {
		t.Fatal("expected nil registry method lookup to miss")
	}
	if descriptors := registry.Descriptors(); descriptors != nil {
		t.Fatalf("expected nil registry descriptors to be nil, got %#v", descriptors)
	}
	if methods := registry.Methods(); methods != nil {
		t.Fatalf("expected nil registry methods to be nil, got %#v", methods)
	}
}

func testDescriptor(owner string, service string, version string, methods ...string) Descriptor {
	methodDescriptors := make([]MethodDescriptor, 0, len(methods))
	for _, method := range methods {
		methodDescriptors = append(methodDescriptors, MethodDescriptor{
			Method:          method,
			Capability:      "host:" + service + ":" + method,
			Risk:            RiskLevelExecute,
			ResourceKind:    ResourceKindNone,
			RequestPayload:  "HostServiceJSONRequest",
			ResponsePayload: "HostServiceJSONResponse",
		})
	}
	return Descriptor{
		OwnerPluginID:  owner,
		Service:        service,
		Version:        version,
		SourceContract: "lina-plugin-" + owner + "/backend/cap/" + service + "cap",
		DynamicContract: "lina-plugin-" + owner + "/backend/cap/" +
			service + "cap/bridge",
		Methods: methodDescriptors,
	}
}

type testInvoker struct{}

func (testInvoker) Invoke(context.Context, Invocation) (*InvocationResult, error) {
	return &InvocationResult{Payload: []byte("ok")}, nil
}

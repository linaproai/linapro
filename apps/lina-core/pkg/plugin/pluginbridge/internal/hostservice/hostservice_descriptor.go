// This file derives private host-service governance lookup tables from the
// public protocol/hostservices catalog. The catalog is the single metadata
// source for service, method, capability, resource, payload, guest, and
// dispatcher publication state.

package hostservice

import "lina-core/pkg/plugin/pluginbridge/protocol/hostservices"

// HostServiceResourceKind describes which authorization resource shape a host
// service declaration uses in plugin manifests.
type HostServiceResourceKind = hostservices.ResourceKind

// Host-service resource kinds used by manifest validation and governance tests.
const (
	HostServiceResourceNone     = hostservices.ResourceKindNone
	HostServiceResourcePath     = hostservices.ResourceKindPath
	HostServiceResourceTable    = hostservices.ResourceKindTable
	HostServiceResourceKey      = hostservices.ResourceKindKey
	HostServiceResourceRef      = hostservices.ResourceKindRef
	HostServiceResourceReserved = hostservices.ResourceKindReserved
)

// HostServiceDescriptor describes one logical host service family.
type HostServiceDescriptor = hostservices.ServiceDescriptor

// HostServiceMethodDescriptor describes one governed host service method.
type HostServiceMethodDescriptor = hostservices.MethodDescriptor

// HostServiceDescriptors returns a copy of the governed host service descriptor table.
func HostServiceDescriptors() []HostServiceDescriptor {
	return hostservices.Catalog()
}

// HostServiceMethodDescriptors returns all governed host-service method descriptors.
func HostServiceMethodDescriptors() []HostServiceMethodDescriptor {
	return append([]HostServiceMethodDescriptor(nil), hostServiceMethodDescriptors()...)
}

func hostServiceMethodDescriptors() []HostServiceMethodDescriptor {
	return hostservices.Methods()
}

func buildHostServiceMethodCapabilityMap() map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, descriptor := range hostServiceMethodDescriptors() {
		if descriptor.Service == "" || descriptor.Method == "" || descriptor.Capability == "" {
			continue
		}
		if result[descriptor.Service] == nil {
			result[descriptor.Service] = make(map[string]string)
		}
		result[descriptor.Service][descriptor.Method] = descriptor.Capability
	}
	return result
}

func buildHostServiceMethodResourceMap() map[string]map[string]HostServiceResourceKind {
	result := make(map[string]map[string]HostServiceResourceKind)
	for _, descriptor := range hostServiceMethodDescriptors() {
		if descriptor.Service == "" || descriptor.Method == "" {
			continue
		}
		if result[descriptor.Service] == nil {
			result[descriptor.Service] = make(map[string]HostServiceResourceKind)
		}
		result[descriptor.Service][descriptor.Method] = descriptor.ResourceKind
	}
	return result
}

func buildHostServiceCapabilitySet() map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range hostServiceMethodDescriptors() {
		if descriptor.Capability != "" {
			result[descriptor.Capability] = struct{}{}
		}
	}
	return result
}

func buildHostServiceResourceKindSet(kind HostServiceResourceKind) map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range HostServiceDescriptors() {
		if descriptor.ResourceKind == kind {
			result[descriptor.Service] = struct{}{}
		}
	}
	return result
}

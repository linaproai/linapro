// This file derives private host-service governance lookup tables from the
// public protocol/hostservices catalog. The catalog is the single metadata
// source for service, method, capability, resource, payload, guest, and
// dispatcher publication state.

package hostservice

import (
	"lina-core/pkg/plugin/capability/capregistry"
	"lina-core/pkg/plugin/pluginbridge/protocol/hostservices"
)

// hostServiceResourceKind describes which authorization resource shape a host
// service declaration uses in plugin manifests.
type hostServiceResourceKind = hostservices.ResourceKind

// Host-service resource kinds used by manifest validation and governance tests.
const (
	hostServiceResourceNone  = hostservices.ResourceKindNone
	hostServiceResourcePath  = hostservices.ResourceKindPath
	hostServiceResourceTable = hostservices.ResourceKindTable
	hostServiceResourceKey   = hostservices.ResourceKindKey
	hostServiceResourceRef   = hostservices.ResourceKindRef
)

// hostServiceDescriptor describes one logical host service family.
type hostServiceDescriptor = hostservices.ServiceDescriptor

// hostServiceMethodDescriptor describes one governed host service method.
type hostServiceMethodDescriptor = hostservices.MethodDescriptor

// hostServiceDescriptors returns a copy of the governed host service descriptor table.
func hostServiceDescriptors() []hostServiceDescriptor {
	return hostservices.Catalog()
}

// HostServiceCatalogFromCapabilityDescriptors returns the core-owned static
// catalog merged with plugin-owned owner capability descriptor projections.
func HostServiceCatalogFromCapabilityDescriptors(descriptors []capregistry.Descriptor) ([]hostservices.ServiceDescriptor, error) {
	return hostservices.CatalogWithDescriptors(descriptors)
}

// hostServiceMethodDescriptors returns all governed host-service method descriptors.
func hostServiceMethodDescriptors() []hostServiceMethodDescriptor {
	return append([]hostServiceMethodDescriptor(nil), rawHostServiceMethodDescriptors()...)
}

// HostServiceMethodsFromCapabilityDescriptors returns all core-owned and
// plugin-owned method projections from the merged catalog.
func HostServiceMethodsFromCapabilityDescriptors(descriptors []capregistry.Descriptor) ([]hostservices.MethodDescriptor, error) {
	return hostservices.MethodsWithDescriptors(descriptors)
}

func rawHostServiceMethodDescriptors() []hostServiceMethodDescriptor {
	return hostservices.Methods()
}

func publishedHostServiceMethodDescriptors() []hostServiceMethodDescriptor {
	descriptors := hostServiceMethodDescriptors()
	result := make([]hostServiceMethodDescriptor, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.Published {
			result = append(result, descriptor)
		}
	}
	return result
}

func buildHostServiceMethodCapabilityMap() map[string]map[string]string {
	result := make(map[string]map[string]string)
	for _, descriptor := range publishedHostServiceMethodDescriptors() {
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

func buildHostServiceMethodResourceMap() map[string]map[string]hostServiceResourceKind {
	result := make(map[string]map[string]hostServiceResourceKind)
	for _, descriptor := range publishedHostServiceMethodDescriptors() {
		if descriptor.Service == "" || descriptor.Method == "" {
			continue
		}
		if result[descriptor.Service] == nil {
			result[descriptor.Service] = make(map[string]hostServiceResourceKind)
		}
		result[descriptor.Service][descriptor.Method] = descriptor.ResourceKind
	}
	return result
}

func buildHostServiceCapabilitySet() map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range publishedHostServiceMethodDescriptors() {
		if descriptor.Capability != "" {
			result[descriptor.Capability] = struct{}{}
		}
	}
	return result
}

func buildHostServiceResourceKindSet(kind hostServiceResourceKind) map[string]struct{} {
	result := make(map[string]struct{})
	for _, descriptor := range hostServiceDescriptors() {
		if descriptor.ResourceKind == kind {
			result[descriptor.Service] = struct{}{}
		}
	}
	return result
}

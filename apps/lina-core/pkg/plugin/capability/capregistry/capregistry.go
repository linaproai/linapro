// Package capregistry defines the generic descriptor registry used by
// plugin-owned domain capabilities. The registry stores owner/service/version
// metadata and method projections only; owner plugins keep typed contracts,
// provider factories, codecs, and business handlers in their own cap packages.
package capregistry

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability"
)

// ResourceKind describes which authorization resource shape a capability
// method uses when it is projected as a dynamic host service.
type ResourceKind string

// Capability descriptor resource kinds.
const (
	ResourceKindNone  ResourceKind = "none"
	ResourceKindPath  ResourceKind = "path"
	ResourceKindTable ResourceKind = "table"
	ResourceKindKey   ResourceKind = "key"
	ResourceKindRef   ResourceKind = "resource"
)

// RiskLevel classifies one owner method for authorization and upgrade previews.
type RiskLevel string

// Capability descriptor risk levels.
const (
	RiskLevelRead    RiskLevel = "read"
	RiskLevelWrite   RiskLevel = "write"
	RiskLevelExecute RiskLevel = "execute"
)

// Descriptor describes one plugin-owned capability service version published by
// an owner plugin.
type Descriptor struct {
	// OwnerPluginID is the plugin ID that owns this capability contract.
	OwnerPluginID string
	// Service is the dynamic service key, such as `ai`.
	Service string
	// Version is the owner capability protocol version, such as `v1`.
	Version string
	// SourceContract names the public Go contract package for source plugins.
	SourceContract string
	// DynamicContract names the public bridge SDK package for dynamic plugins.
	DynamicContract string
	// Invoker routes authorized dynamic host-service calls to the owner plugin.
	// It is runtime-only metadata and is never projected into catalogs.
	Invoker Invoker
	// Methods contains all methods published by this descriptor.
	Methods []MethodDescriptor
}

// Invocation carries one authorized plugin-owned dynamic host-service call.
type Invocation struct {
	// CallerPluginID identifies the dynamic plugin that initiated the call.
	CallerPluginID string
	// OwnerPluginID identifies the plugin that owns the target capability.
	OwnerPluginID string
	// Services exposes the owner-scoped host capability directory that owner
	// handlers can use to construct typed provider environments without
	// reaching into host internals.
	Services capability.Services
	// Service is the owner capability service key.
	Service string
	// Version is the owner capability protocol version.
	Version string
	// Method is the target owner capability method.
	Method string
	// ResourceRef is an optional governed resource reference.
	ResourceRef string
	// Table is reserved for core-owned data host services and should remain
	// empty for plugin-owned capabilities.
	Table string
	// Payload carries owner-encoded request bytes.
	Payload []byte
}

// InvocationResult carries an owner handler response using host-call status
// code semantics without importing pluginbridge and forming a package cycle.
type InvocationResult struct {
	// Status is the host-call status code. Zero is success.
	Status uint32
	// Payload carries success payload bytes or a pre-encoded error payload.
	Payload []byte
}

// Invoker dispatches authorized dynamic calls to an owner plugin.
type Invoker interface {
	// Invoke executes one authorized owner capability method.
	Invoke(ctx context.Context, invocation Invocation) (*InvocationResult, error)
}

// MethodDescriptor describes one method published by a plugin-owned capability
// descriptor.
type MethodDescriptor struct {
	// Method is the wire method name, such as `text.generate`.
	Method string
	// Capability is the authorization capability implied by this method.
	Capability string
	// Risk classifies the method for authorization and upgrade previews.
	Risk RiskLevel
	// ResourceKind describes the method resource declaration shape.
	ResourceKind ResourceKind
	// RequestPayload names the public request payload type when one exists.
	RequestPayload string
	// ResponsePayload names the public response payload type when one exists.
	ResponsePayload string
}

// RegisteredMethod is one indexed method projection with its owner key.
type RegisteredMethod struct {
	// OwnerPluginID is the plugin ID that owns this capability contract.
	OwnerPluginID string
	// Service is the dynamic service key.
	Service string
	// Version is the owner capability protocol version.
	Version string
	// Descriptor contains method-level metadata.
	Descriptor MethodDescriptor
	// Invoker routes authorized calls to the owner plugin for this method.
	Invoker Invoker
}

// Registry stores generic plugin-owned capability descriptors and method indexes.
type Registry struct {
	mu sync.RWMutex

	descriptors map[descriptorKey]Descriptor
	methods     map[methodKey]RegisteredMethod
}

type descriptorKey struct {
	owner   string
	service string
	version string
}

type methodKey struct {
	owner   string
	service string
	version string
	method  string
}

// NewRegistry creates an empty capability descriptor registry.
func NewRegistry() *Registry {
	return &Registry{
		descriptors: make(map[descriptorKey]Descriptor),
		methods:     make(map[methodKey]RegisteredMethod),
	}
}

// Register validates and stores one owner capability descriptor.
func (r *Registry) Register(descriptor Descriptor) error {
	if r == nil {
		return gerror.New("capability descriptor registry is nil")
	}
	normalized, err := normalizeDescriptor(descriptor)
	if err != nil {
		return err
	}
	key := descriptorKey{
		owner:   normalized.OwnerPluginID,
		service: normalized.Service,
		version: normalized.Version,
	}
	methods := make([]RegisteredMethod, 0, len(normalized.Methods))
	for _, method := range normalized.Methods {
		methods = append(methods, RegisteredMethod{
			OwnerPluginID: normalized.OwnerPluginID,
			Service:       normalized.Service,
			Version:       normalized.Version,
			Descriptor:    method,
			Invoker:       normalized.Invoker,
		})
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.descriptors[key]; exists {
		return gerror.Newf(
			"capability descriptor already registered: owner=%s service=%s version=%s",
			key.owner,
			key.service,
			key.version,
		)
	}
	for _, method := range methods {
		key := method.key()
		if _, exists := r.methods[key]; exists {
			return gerror.Newf(
				"capability method already registered: owner=%s service=%s version=%s method=%s",
				key.owner,
				key.service,
				key.version,
				key.method,
			)
		}
	}
	r.descriptors[key] = cloneDescriptor(normalized)
	for _, method := range methods {
		r.methods[method.key()] = cloneRegisteredMethod(method)
	}
	return nil
}

// LookupDescriptor returns one registered owner descriptor by owner, service,
// and version.
func (r *Registry) LookupDescriptor(ownerPluginID string, service string, version string) (Descriptor, bool) {
	if r == nil {
		return Descriptor{}, false
	}
	key := descriptorKey{
		owner:   strings.TrimSpace(ownerPluginID),
		service: strings.TrimSpace(service),
		version: strings.TrimSpace(version),
	}
	r.mu.RLock()
	descriptor, ok := r.descriptors[key]
	r.mu.RUnlock()
	if !ok {
		return Descriptor{}, false
	}
	return cloneDescriptor(descriptor), true
}

// LookupMethod returns one indexed owner method by owner, service, version, and
// method.
func (r *Registry) LookupMethod(ownerPluginID string, service string, version string, method string) (RegisteredMethod, bool) {
	if r == nil {
		return RegisteredMethod{}, false
	}
	key := methodKey{
		owner:   strings.TrimSpace(ownerPluginID),
		service: strings.TrimSpace(service),
		version: strings.TrimSpace(version),
		method:  strings.TrimSpace(method),
	}
	r.mu.RLock()
	registered, ok := r.methods[key]
	r.mu.RUnlock()
	if !ok {
		return RegisteredMethod{}, false
	}
	return cloneRegisteredMethod(registered), true
}

// Descriptors returns all registered descriptors in deterministic owner,
// service, and version order.
func (r *Registry) Descriptors() []Descriptor {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	items := make([]Descriptor, 0, len(r.descriptors))
	for _, descriptor := range r.descriptors {
		items = append(items, cloneDescriptor(descriptor))
	}
	r.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return descriptorSortKey(items[i]) < descriptorSortKey(items[j])
	})
	return items
}

// Methods returns all registered method projections in deterministic owner,
// service, version, and method order.
func (r *Registry) Methods() []RegisteredMethod {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	items := make([]RegisteredMethod, 0, len(r.methods))
	for _, method := range r.methods {
		items = append(items, cloneRegisteredMethod(method))
	}
	r.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		return methodSortKey(items[i]) < methodSortKey(items[j])
	})
	return items
}

func (m RegisteredMethod) key() methodKey {
	return methodKey{
		owner:   m.OwnerPluginID,
		service: m.Service,
		version: m.Version,
		method:  m.Descriptor.Method,
	}
}

func normalizeDescriptor(descriptor Descriptor) (Descriptor, error) {
	descriptor.OwnerPluginID = strings.TrimSpace(descriptor.OwnerPluginID)
	descriptor.Service = strings.TrimSpace(descriptor.Service)
	descriptor.Version = strings.TrimSpace(descriptor.Version)
	descriptor.SourceContract = strings.TrimSpace(descriptor.SourceContract)
	descriptor.DynamicContract = strings.TrimSpace(descriptor.DynamicContract)
	if descriptor.OwnerPluginID == "" {
		return Descriptor{}, gerror.New("capability descriptor owner plugin id is required")
	}
	if descriptor.Service == "" {
		return Descriptor{}, gerror.Newf("capability descriptor service is required: owner=%s", descriptor.OwnerPluginID)
	}
	if descriptor.Version == "" {
		return Descriptor{}, gerror.Newf(
			"capability descriptor version is required: owner=%s service=%s",
			descriptor.OwnerPluginID,
			descriptor.Service,
		)
	}
	if len(descriptor.Methods) == 0 {
		return Descriptor{}, gerror.Newf(
			"capability descriptor methods are required: owner=%s service=%s version=%s",
			descriptor.OwnerPluginID,
			descriptor.Service,
			descriptor.Version,
		)
	}
	seenMethods := make(map[string]struct{}, len(descriptor.Methods))
	for i := range descriptor.Methods {
		method := &descriptor.Methods[i]
		method.Method = strings.TrimSpace(method.Method)
		method.Capability = strings.TrimSpace(method.Capability)
		method.RequestPayload = strings.TrimSpace(method.RequestPayload)
		method.ResponsePayload = strings.TrimSpace(method.ResponsePayload)
		if method.ResourceKind == "" {
			method.ResourceKind = ResourceKindNone
		}
		if method.Method == "" {
			return Descriptor{}, gerror.Newf(
				"capability descriptor method is required: owner=%s service=%s version=%s",
				descriptor.OwnerPluginID,
				descriptor.Service,
				descriptor.Version,
			)
		}
		if _, exists := seenMethods[method.Method]; exists {
			return Descriptor{}, gerror.Newf(
				"capability descriptor contains duplicate method: owner=%s service=%s version=%s method=%s",
				descriptor.OwnerPluginID,
				descriptor.Service,
				descriptor.Version,
				method.Method,
			)
		}
		seenMethods[method.Method] = struct{}{}
	}
	return descriptor, nil
}

func cloneDescriptor(descriptor Descriptor) Descriptor {
	descriptor.Methods = append([]MethodDescriptor(nil), descriptor.Methods...)
	return descriptor
}

func cloneRegisteredMethod(method RegisteredMethod) RegisteredMethod {
	return RegisteredMethod{
		OwnerPluginID: method.OwnerPluginID,
		Service:       method.Service,
		Version:       method.Version,
		Descriptor:    method.Descriptor,
		Invoker:       method.Invoker,
	}
}

func descriptorSortKey(descriptor Descriptor) string {
	return descriptor.OwnerPluginID + "\x00" + descriptor.Service + "\x00" + descriptor.Version
}

func methodSortKey(method RegisteredMethod) string {
	return method.OwnerPluginID + "\x00" + method.Service + "\x00" + method.Version + "\x00" + method.Descriptor.Method
}

// This file discovers and executes dynamic-plugin Jobs declarations through the
// shared Wasm bridge so plugin-owned scheduled jobs reuse the unified host task
// pipeline.

package runtime

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgecodec "lina-core/pkg/plugin/pluginbridge/protocol"
)

// jobDiscoveryCollector stores dynamic-plugin Jobs declarations discovered
// from one guest-side RegisterJobs execution.
type jobDiscoveryCollector struct {
	pluginID string
	seen     map[string]struct{}
	items    []*bridgecontract.JobContract
}

var _ wasm.JobRegistrationCollector = (*jobDiscoveryCollector)(nil)

// newJobDiscoveryCollector creates one in-memory collector for a single plugin
// discovery execution.
func newJobDiscoveryCollector(pluginID string) *jobDiscoveryCollector {
	return &jobDiscoveryCollector{
		pluginID: strings.TrimSpace(pluginID),
		seen:     make(map[string]struct{}),
		items:    make([]*bridgecontract.JobContract, 0),
	}
}

// Register validates and stores one discovered job contract.
func (c *jobDiscoveryCollector) Register(contract *bridgecontract.JobContract) error {
	if contract == nil {
		return gerror.New("dynamic plugin job declaration cannot be nil")
	}

	contractSnapshot := *contract
	if err := bridgecontract.ValidateJobContracts(c.pluginID, []*bridgecontract.JobContract{&contractSnapshot}); err != nil {
		return err
	}
	if _, exists := c.seen[contractSnapshot.Name]; exists {
		return gerror.Newf("dynamic plugin %s job name is duplicated: %s", c.pluginID, contractSnapshot.Name)
	}
	c.seen[contractSnapshot.Name] = struct{}{}
	c.items = append(c.items, &contractSnapshot)
	return nil
}

// Items returns a detached copy of the discovered job contract list.
func (c *jobDiscoveryCollector) Items() []*bridgecontract.JobContract {
	if c == nil || len(c.items) == 0 {
		return []*bridgecontract.JobContract{}
	}
	items := make([]*bridgecontract.JobContract, 0, len(c.items))
	for _, item := range c.items {
		if item == nil {
			continue
		}
		itemSnapshot := *item
		items = append(items, &itemSnapshot)
	}
	return items
}

// DiscoverJobContracts runs the reserved guest-side Jobs registration entry
// point and collects all declared dynamic-plugin job contracts.
func (s *serviceImpl) DiscoverJobContracts(
	ctx context.Context,
	manifest *catalog.Manifest,
) ([]*bridgecontract.JobContract, error) {
	if manifest == nil {
		return nil, gerror.New("dynamic plugin manifest cannot be nil")
	}
	if manifest.RuntimeArtifact == nil || strings.TrimSpace(manifest.RuntimeArtifact.Path) == "" {
		return nil, gerror.Newf("dynamic plugin %s is missing executable runtime artifact", manifest.ID)
	}
	if manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		return nil, gerror.Newf("dynamic plugin %s does not declare an executable Wasm bridge", manifest.ID)
	}

	collector := newJobDiscoveryCollector(manifest.ID)
	request := &bridgecontract.BridgeRequestEnvelopeV1{
		PluginID: strings.TrimSpace(manifest.ID),
		Route: &bridgecontract.RouteMatchSnapshotV1{
			RoutePath:    bridgecontract.DeclaredJobRegistrationInternalPath,
			InternalPath: bridgecontract.DeclaredJobRegistrationInternalPath,
			RequestType:  bridgecontract.DeclaredJobRegistrationRequestType,
		},
		RequestID: guid.S(),
	}
	requestContent, err := bridgecodec.EncodeRequestEnvelope(request)
	if err != nil {
		return nil, err
	}

	if s.wasmRuntime == nil {
		return nil, gerror.New("dynamic wasm runtime is not configured")
	}
	response, err := s.wasmRuntime.ExecuteBridge(ctx, wasm.ExecutionInput{
		PluginID:                  manifest.ID,
		ArtifactPath:              manifest.RuntimeArtifact.Path,
		BridgeSpec:                manifest.BridgeSpec,
		Capabilities:              manifest.HostCapabilities,
		HostServices:              manifest.HostServices,
		ArtifactDefaultConfig:     buildArtifactDefaultConfig(manifest),
		ArtifactManifestResources: buildArtifactManifestResources(manifest),
		ExecutionSource:           bridgecontract.ExecutionSourceJobsDiscovery,
		RoutePath:                 bridgecontract.DeclaredJobRegistrationInternalPath,
		RequestID:                 request.RequestID,
		JobCollector:              collector,
	}, requestContent)
	if err != nil {
		return nil, err
	}
	return normalizeDiscoveredJobContracts(manifest.ID, response, collector)
}

// ExecuteDeclaredJob runs one declared dynamic-plugin job through the active
// runtime bridge.
func (s *serviceImpl) ExecuteDeclaredJob(
	ctx context.Context,
	manifest *catalog.Manifest,
	contract *bridgecontract.JobContract,
) error {
	if manifest == nil {
		return gerror.New("dynamic plugin manifest cannot be nil")
	}
	if contract == nil {
		return gerror.New("dynamic plugin job contract cannot be nil")
	}
	if manifest.RuntimeArtifact == nil || strings.TrimSpace(manifest.RuntimeArtifact.Path) == "" {
		return gerror.Newf("dynamic plugin %s is missing executable runtime artifact", manifest.ID)
	}
	if manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		return gerror.Newf("dynamic plugin %s does not declare an executable Wasm bridge", manifest.ID)
	}

	routePath := bridgecontract.BuildDeclaredJobRoutePath(contract)
	request := &bridgecontract.BridgeRequestEnvelopeV1{
		PluginID: strings.TrimSpace(manifest.ID),
		Route: &bridgecontract.RouteMatchSnapshotV1{
			RoutePath:    routePath,
			InternalPath: routePath,
			RequestType:  strings.TrimSpace(contract.RequestType),
		},
		RequestID: guid.S(),
	}
	requestContent, err := bridgecodec.EncodeRequestEnvelope(request)
	if err != nil {
		return err
	}

	if s.wasmRuntime == nil {
		return gerror.New("dynamic wasm runtime is not configured")
	}
	response, err := s.wasmRuntime.ExecuteBridge(ctx, wasm.ExecutionInput{
		PluginID:                  manifest.ID,
		ArtifactPath:              manifest.RuntimeArtifact.Path,
		BridgeSpec:                manifest.BridgeSpec,
		Capabilities:              manifest.HostCapabilities,
		HostServices:              manifest.HostServices,
		ArtifactDefaultConfig:     buildArtifactDefaultConfig(manifest),
		ArtifactManifestResources: buildArtifactManifestResources(manifest),
		ExecutionSource:           bridgecontract.ExecutionSourceJobs,
		RoutePath:                 routePath,
		RequestID:                 request.RequestID,
	}, requestContent)
	if err != nil {
		return err
	}
	return normalizeDeclaredJobResponse(contract, response)
}

// normalizeDiscoveredJobContracts converts one bridge response into the
// discovered job contract list expected by the integration layer.
func normalizeDiscoveredJobContracts(
	pluginID string,
	response *bridgecontract.BridgeResponseEnvelopeV1,
	collector *jobDiscoveryCollector,
) ([]*bridgecontract.JobContract, error) {
	if response == nil {
		return nil, gerror.New("dynamic plugin Jobs registration returned no execution result")
	}
	if response.StatusCode == http.StatusNotFound {
		return []*bridgecontract.JobContract{}, nil
	}
	if response.Failure != nil {
		return nil, gerror.New(strings.TrimSpace(response.Failure.Message))
	}
	if response.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(response.Body))
		if message == "" {
			message = http.StatusText(int(response.StatusCode))
		}
		return nil, gerror.Newf("dynamic plugin Jobs discovery failed (%s): %s", strings.TrimSpace(pluginID), message)
	}

	contracts := collector.Items()
	if err := bridgecontract.ValidateJobContracts(pluginID, contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// normalizeDeclaredJobResponse converts one bridge response into the shared
// job handler error contract expected by scheduled-job execution.
func normalizeDeclaredJobResponse(
	contract *bridgecontract.JobContract,
	response *bridgecontract.BridgeResponseEnvelopeV1,
) error {
	if response == nil {
		return gerror.New("dynamic plugin job returned no execution result")
	}
	if response.Failure != nil {
		return gerror.New(strings.TrimSpace(response.Failure.Message))
	}
	if response.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(response.Body))
		if message == "" {
			message = http.StatusText(int(response.StatusCode))
		}
		return gerror.Newf("dynamic plugin job execution failed (%s): %s", strings.TrimSpace(contract.Name), message)
	}
	return nil
}

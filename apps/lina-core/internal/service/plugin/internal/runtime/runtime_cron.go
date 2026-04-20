// This file discovers and executes dynamic-plugin cron declarations through
// the shared Wasm bridge so plugin-owned scheduled jobs reuse the unified host
// task pipeline.

package runtime

import (
	"context"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	"lina-core/pkg/pluginbridge"
)

// cronDiscoveryCollector stores dynamic-plugin cron declarations discovered
// from one guest-side RegisterCrons execution.
type cronDiscoveryCollector struct {
	pluginID string
	seen     map[string]struct{}
	items    []*pluginbridge.CronContract
}

// Ensure cronDiscoveryCollector satisfies the shared Wasm discovery contract.
var _ wasm.CronRegistrationCollector = (*cronDiscoveryCollector)(nil)

// newCronDiscoveryCollector creates one in-memory collector for a single
// plugin discovery execution.
func newCronDiscoveryCollector(pluginID string) *cronDiscoveryCollector {
	return &cronDiscoveryCollector{
		pluginID: strings.TrimSpace(pluginID),
		seen:     make(map[string]struct{}),
		items:    make([]*pluginbridge.CronContract, 0),
	}
}

// Register validates and stores one discovered cron contract.
func (c *cronDiscoveryCollector) Register(contract *pluginbridge.CronContract) error {
	if contract == nil {
		return gerror.New("动态插件定时任务声明不能为空")
	}

	contractSnapshot := *contract
	if err := pluginbridge.ValidateCronContracts(c.pluginID, []*pluginbridge.CronContract{&contractSnapshot}); err != nil {
		return err
	}
	if _, exists := c.seen[contractSnapshot.Name]; exists {
		return gerror.Newf("动态插件 %s 的定时任务 name 不能重复: %s", c.pluginID, contractSnapshot.Name)
	}
	c.seen[contractSnapshot.Name] = struct{}{}
	c.items = append(c.items, &contractSnapshot)
	return nil
}

// Items returns a detached copy of the discovered cron contract list.
func (c *cronDiscoveryCollector) Items() []*pluginbridge.CronContract {
	if c == nil || len(c.items) == 0 {
		return []*pluginbridge.CronContract{}
	}
	items := make([]*pluginbridge.CronContract, 0, len(c.items))
	for _, item := range c.items {
		if item == nil {
			continue
		}
		itemSnapshot := *item
		items = append(items, &itemSnapshot)
	}
	return items
}

// DiscoverCronContracts runs the reserved guest-side cron registration entry
// point and collects all declared dynamic-plugin cron contracts.
func (s *serviceImpl) DiscoverCronContracts(
	ctx context.Context,
	manifest *catalog.Manifest,
) ([]*pluginbridge.CronContract, error) {
	if manifest == nil {
		return nil, gerror.New("动态插件清单不能为空")
	}
	if manifest.RuntimeArtifact == nil || strings.TrimSpace(manifest.RuntimeArtifact.Path) == "" {
		return nil, gerror.Newf("动态插件 %s 缺少可执行运行时产物", manifest.ID)
	}
	if manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		return nil, gerror.Newf("动态插件 %s 未声明可执行 Wasm bridge", manifest.ID)
	}

	collector := newCronDiscoveryCollector(manifest.ID)
	request := &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID: strings.TrimSpace(manifest.ID),
		Route: &pluginbridge.RouteMatchSnapshotV1{
			RoutePath:    pluginbridge.DeclaredCronRegistrationInternalPath,
			InternalPath: pluginbridge.DeclaredCronRegistrationInternalPath,
			RequestType:  pluginbridge.DeclaredCronRegistrationRequestType,
		},
		RequestID: guid.S(),
	}
	requestContent, err := pluginbridge.EncodeRequestEnvelope(request)
	if err != nil {
		return nil, err
	}

	response, err := wasm.ExecuteBridge(ctx, wasm.ExecutionInput{
		PluginID:        manifest.ID,
		ArtifactPath:    manifest.RuntimeArtifact.Path,
		BridgeSpec:      manifest.BridgeSpec,
		Capabilities:    manifest.HostCapabilities,
		HostServices:    manifest.HostServices,
		ExecutionSource: pluginbridge.ExecutionSourceCronDiscovery,
		RoutePath:       pluginbridge.DeclaredCronRegistrationInternalPath,
		RequestID:       request.RequestID,
		CronCollector:   collector,
	}, requestContent)
	if err != nil {
		return nil, err
	}
	return normalizeDiscoveredCronContracts(manifest.ID, response, collector)
}

// ExecuteDeclaredCronJob runs one declared dynamic-plugin cron job through the
// active runtime bridge.
func (s *serviceImpl) ExecuteDeclaredCronJob(
	ctx context.Context,
	manifest *catalog.Manifest,
	contract *pluginbridge.CronContract,
) error {
	if manifest == nil {
		return gerror.New("动态插件清单不能为空")
	}
	if contract == nil {
		return gerror.New("动态插件定时任务契约不能为空")
	}
	if manifest.RuntimeArtifact == nil || strings.TrimSpace(manifest.RuntimeArtifact.Path) == "" {
		return gerror.Newf("动态插件 %s 缺少可执行运行时产物", manifest.ID)
	}
	if manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		return gerror.Newf("动态插件 %s 未声明可执行 Wasm bridge", manifest.ID)
	}

	request := &pluginbridge.BridgeRequestEnvelopeV1{
		PluginID: strings.TrimSpace(manifest.ID),
		Route: &pluginbridge.RouteMatchSnapshotV1{
			RoutePath:    pluginbridge.BuildDeclaredCronRoutePath(contract),
			InternalPath: strings.TrimSpace(contract.InternalPath),
			RequestType:  strings.TrimSpace(contract.RequestType),
		},
		RequestID: guid.S(),
	}
	requestContent, err := pluginbridge.EncodeRequestEnvelope(request)
	if err != nil {
		return err
	}

	response, err := wasm.ExecuteBridge(ctx, wasm.ExecutionInput{
		PluginID:        manifest.ID,
		ArtifactPath:    manifest.RuntimeArtifact.Path,
		BridgeSpec:      manifest.BridgeSpec,
		Capabilities:    manifest.HostCapabilities,
		HostServices:    manifest.HostServices,
		ExecutionSource: pluginbridge.ExecutionSourceCron,
		RoutePath:       pluginbridge.BuildDeclaredCronRoutePath(contract),
		RequestID:       request.RequestID,
	}, requestContent)
	if err != nil {
		return err
	}
	return normalizeDeclaredCronResponse(contract, response)
}

// normalizeDiscoveredCronContracts converts one bridge response into the
// discovered cron contract list expected by the integration layer.
func normalizeDiscoveredCronContracts(
	pluginID string,
	response *pluginbridge.BridgeResponseEnvelopeV1,
	collector *cronDiscoveryCollector,
) ([]*pluginbridge.CronContract, error) {
	if response == nil {
		return nil, gerror.New("动态插件定时任务注册未返回执行结果")
	}
	if response.StatusCode == http.StatusNotFound {
		return []*pluginbridge.CronContract{}, nil
	}
	if response.Failure != nil {
		return nil, gerror.New(strings.TrimSpace(response.Failure.Message))
	}
	if response.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(response.Body))
		if message == "" {
			message = http.StatusText(int(response.StatusCode))
		}
		return nil, gerror.Newf("动态插件定时任务发现失败(%s): %s", strings.TrimSpace(pluginID), message)
	}

	contracts := collector.Items()
	if err := pluginbridge.ValidateCronContracts(pluginID, contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

// normalizeDeclaredCronResponse converts one bridge response into the shared
// cron handler error contract expected by scheduled-job execution.
func normalizeDeclaredCronResponse(
	contract *pluginbridge.CronContract,
	response *pluginbridge.BridgeResponseEnvelopeV1,
) error {
	if response == nil {
		return gerror.New("动态插件定时任务未返回执行结果")
	}
	if response.Failure != nil {
		return gerror.New(strings.TrimSpace(response.Failure.Message))
	}
	if response.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(response.Body))
		if message == "" {
			message = http.StatusText(int(response.StatusCode))
		}
		return gerror.Newf("动态插件定时任务执行失败(%s): %s", strings.TrimSpace(contract.Name), message)
	}
	return nil
}

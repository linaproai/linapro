// This file executes dynamic-plugin lifecycle callbacks through the WASM
// bridge before or after host-owned lifecycle side effects run.

package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/wasm"
	bridgecodec "lina-core/pkg/pluginbridge/codec"
	bridgecontract "lina-core/pkg/pluginbridge/contract"
	"lina-core/pkg/pluginhost"
)

// DynamicLifecycleInput carries host-owned fields published to one dynamic
// lifecycle callback handler.
type DynamicLifecycleInput struct {
	// PluginID is the lifecycle operation target plugin.
	PluginID string
	// Operation is the source-compatible Before* or After* lifecycle operation name.
	Operation pluginhost.LifecycleHook
	// FromVersion is the effective version before upgrade when applicable.
	FromVersion string
	// ToVersion is the target version for upgrade when applicable.
	ToVersion string
	// TenantID is the tenant affected by tenant-scoped lifecycle operations.
	TenantID int
	// FromMode is the previous install mode for install-mode changes.
	FromMode string
	// ToMode is the target install mode for install-mode changes.
	ToMode string
}

// DynamicLifecycleDecision records one dynamic lifecycle handler result.
type DynamicLifecycleDecision struct {
	// PluginID is the callback owner.
	PluginID string
	// Operation is the invoked lifecycle operation.
	Operation pluginhost.LifecycleHook
	// OK reports whether this handler allowed the lifecycle action.
	OK bool
	// Reason is the stable veto reason or execution diagnostic.
	Reason string
	// Err records bridge execution or response decoding errors.
	Err error
}

// RunDynamicLifecycleCallback executes the dynamic handler for the given
// operation when the manifest declares one.
func (s *serviceImpl) RunDynamicLifecycleCallback(
	ctx context.Context,
	manifest *catalog.Manifest,
	input DynamicLifecycleInput,
) (*DynamicLifecycleDecision, error) {
	if manifest == nil {
		return nil, nil
	}
	contract := findDynamicLifecycleContract(manifest, input.Operation)
	if contract == nil {
		return nil, nil
	}
	decision := &DynamicLifecycleDecision{
		PluginID:  manifest.ID,
		Operation: input.Operation,
		OK:        true,
	}

	handlerCtx := ctx
	cancel := func() {}
	if timeout := dynamicLifecycleTimeout(contract); timeout > 0 {
		handlerCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	requestEnvelope, err := buildDynamicLifecycleRequest(manifest, contract, input)
	if err != nil {
		decision.OK = false
		decision.Err = err
		decision.Reason = dynamicLifecycleErrorReason(manifest.ID, input.Operation)
		return decision, err
	}
	requestContent, err := bridgecodec.EncodeRequestEnvelope(requestEnvelope)
	if err != nil {
		decision.OK = false
		decision.Err = err
		decision.Reason = dynamicLifecycleErrorReason(manifest.ID, input.Operation)
		return decision, err
	}

	response, err := wasm.ExecuteBridge(handlerCtx, wasm.ExecutionInput{
		PluginID:        manifest.ID,
		ArtifactPath:    manifest.RuntimeArtifact.Path,
		BridgeSpec:      manifest.BridgeSpec,
		Capabilities:    manifest.HostCapabilities,
		HostServices:    manifest.HostServices,
		ExecutionSource: bridgecontract.ExecutionSourceLifecycle,
		RoutePath:       contract.InternalPath,
		RequestID:       requestEnvelope.RequestID,
	}, requestContent)
	if err != nil {
		decision.OK = false
		decision.Err = err
		decision.Reason = dynamicLifecycleErrorReason(manifest.ID, input.Operation)
		return decision, err
	}
	if err = applyDynamicLifecycleResponse(decision, response); err != nil {
		return decision, err
	}
	return decision, nil
}

// RunDynamicLifecyclePrecondition executes one dynamic Before* handler when declared.
func (s *serviceImpl) RunDynamicLifecyclePrecondition(
	ctx context.Context,
	manifest *catalog.Manifest,
	input DynamicLifecycleInput,
) (*DynamicLifecycleDecision, error) {
	return s.RunDynamicLifecycleCallback(ctx, manifest, input)
}

// findDynamicLifecycleContract returns the handler contract for one operation.
func findDynamicLifecycleContract(
	manifest *catalog.Manifest,
	operation pluginhost.LifecycleHook,
) *bridgecontract.LifecycleContract {
	if manifest == nil {
		return nil
	}
	for _, contract := range manifest.LifecycleHandlers {
		if contract == nil {
			continue
		}
		if contract.Operation.String() == operation.String() {
			return contract
		}
	}
	return nil
}

// dynamicLifecycleTimeout returns the per-handler timeout.
func dynamicLifecycleTimeout(contract *bridgecontract.LifecycleContract) time.Duration {
	if contract == nil || contract.TimeoutMs <= 0 {
		return pluginhost.DefaultLifecycleHookTimeout
	}
	return time.Duration(contract.TimeoutMs) * time.Millisecond
}

// buildDynamicLifecycleRequest creates one bridge request envelope for a
// lifecycle callback invocation.
func buildDynamicLifecycleRequest(
	manifest *catalog.Manifest,
	contract *bridgecontract.LifecycleContract,
	input DynamicLifecycleInput,
) (*bridgecontract.BridgeRequestEnvelopeV1, error) {
	if manifest == nil || manifest.RuntimeArtifact == nil {
		return nil, gerror.New("dynamic lifecycle manifest or artifact is missing")
	}
	if manifest.BridgeSpec == nil || !manifest.BridgeSpec.RouteExecution {
		return nil, gerror.New("dynamic lifecycle requires executable Wasm bridge metadata")
	}

	pluginID := strings.TrimSpace(input.PluginID)
	if pluginID == "" {
		pluginID = strings.TrimSpace(manifest.ID)
	}
	body, err := json.Marshal(bridgecontract.LifecycleRequest{
		PluginID:    pluginID,
		Operation:   input.Operation.String(),
		FromVersion: strings.TrimSpace(input.FromVersion),
		ToVersion:   strings.TrimSpace(input.ToVersion),
		TenantID:    input.TenantID,
		FromMode:    strings.TrimSpace(input.FromMode),
		ToMode:      strings.TrimSpace(input.ToMode),
	})
	if err != nil {
		return nil, err
	}

	return &bridgecontract.BridgeRequestEnvelopeV1{
		PluginID: pluginID,
		Route: &bridgecontract.RouteMatchSnapshotV1{
			Method:       http.MethodPost,
			PublicPath:   contract.InternalPath,
			InternalPath: contract.InternalPath,
			RoutePath:    contract.InternalPath,
			RequestType:  contract.RequestType,
		},
		Request: &bridgecontract.HTTPRequestSnapshotV1{
			Method:       http.MethodPost,
			PublicPath:   contract.InternalPath,
			InternalPath: contract.InternalPath,
			RawPath:      contract.InternalPath,
			ContentType:  "application/json",
			Body:         body,
		},
		RequestID: guid.S(),
	}, nil
}

// applyDynamicLifecycleResponse converts one bridge response into a lifecycle
// decision, failing closed on bridge failures and non-success statuses.
func applyDynamicLifecycleResponse(
	decision *DynamicLifecycleDecision,
	response *bridgecontract.BridgeResponseEnvelopeV1,
) error {
	if decision == nil {
		return nil
	}
	if response == nil {
		err := gerror.New("dynamic lifecycle handler returned no execution result")
		decision.OK = false
		decision.Err = err
		decision.Reason = dynamicLifecycleErrorReason(decision.PluginID, decision.Operation)
		return err
	}
	if response.Failure != nil {
		err := gerror.New(strings.TrimSpace(response.Failure.Message))
		decision.OK = false
		decision.Err = err
		decision.Reason = dynamicLifecycleErrorReason(decision.PluginID, decision.Operation)
		return err
	}
	if response.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(response.Body))
		if message == "" {
			message = http.StatusText(int(response.StatusCode))
		}
		err := gerror.New(message)
		decision.OK = false
		decision.Err = err
		decision.Reason = dynamicLifecycleErrorReason(decision.PluginID, decision.Operation)
		return err
	}
	if len(response.Body) == 0 {
		decision.OK = true
		return nil
	}

	payload := &bridgecontract.LifecycleDecision{}
	if err := json.Unmarshal(response.Body, payload); err != nil {
		decision.OK = false
		decision.Err = err
		decision.Reason = dynamicLifecycleErrorReason(decision.PluginID, decision.Operation)
		return err
	}
	decision.OK = payload.OK
	decision.Reason = strings.TrimSpace(payload.Reason)
	if !decision.OK && decision.Reason == "" {
		decision.Reason = dynamicLifecycleVetoReason(decision.PluginID, decision.Operation)
	}
	return nil
}

// dynamicLifecycleVetoReason returns the default veto reason key.
func dynamicLifecycleVetoReason(pluginID string, operation pluginhost.LifecycleHook) string {
	return "plugin." + strings.TrimSpace(pluginID) + ".lifecycle." + operation.String() + ".vetoed"
}

// dynamicLifecycleErrorReason returns the default bridge error reason key.
func dynamicLifecycleErrorReason(pluginID string, operation pluginhost.LifecycleHook) string {
	return "plugin." + strings.TrimSpace(pluginID) + ".lifecycle." + operation.String() + ".failed"
}

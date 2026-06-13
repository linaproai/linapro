// This file tests Jobs-domain host service dispatch for dynamic plugin job
// declaration discovery.

package wasm

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// trackingJobCollector records discovered Jobs declarations.
type trackingJobCollector struct {
	items []*protocol.JobContract
}

// Register stores one discovered job declaration.
func (c *trackingJobCollector) Register(contract *protocol.JobContract) error {
	if contract == nil {
		return nil
	}
	snapshot := *contract
	c.items = append(c.items, &snapshot)
	return nil
}

// TestHandleHostServiceInvokeJobsRegisterAcceptsDiscoverySource verifies
// jobs.register is accepted during dynamic Jobs discovery and normalizes the
// submitted declaration before it reaches the collector.
func TestHandleHostServiceInvokeJobsRegisterAcceptsDiscoverySource(t *testing.T) {
	collector := &trackingJobCollector{}
	hcc := newJobsRegisterHostCallContext(protocol.ExecutionSourceJobsDiscovery, collector)

	response := invokeJobsRegisterHostService(t, hcc, &protocol.JobContract{
		Name:        " heartbeat ",
		Pattern:     "# */10 * * * *",
		RequestType: "JobHeartbeatReq",
	})
	if response.Status != protocol.HostCallStatusSuccess {
		t.Fatalf("expected jobs.register discovery to succeed, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if len(collector.items) != 1 {
		t.Fatalf("expected one collected job declaration, got %#v", collector.items)
	}
	collected := collector.items[0]
	if collected.Name != "heartbeat" {
		t.Fatalf("expected normalized name heartbeat, got %q", collected.Name)
	}
	if collected.Timezone != protocol.DefaultJobContractTimezone {
		t.Fatalf("expected default timezone %q, got %q", protocol.DefaultJobContractTimezone, collected.Timezone)
	}
	if collected.Scope != protocol.JobScopeAllNode {
		t.Fatalf("expected default all-node scope, got %q", collected.Scope)
	}
	if collected.Concurrency != protocol.JobConcurrencySingleton {
		t.Fatalf("expected default singleton concurrency, got %q", collected.Concurrency)
	}
	if collected.MaxConcurrency != 1 {
		t.Fatalf("expected default max concurrency 1, got %d", collected.MaxConcurrency)
	}
	if collected.TimeoutSeconds != protocol.DefaultJobContractTimeoutSeconds {
		t.Fatalf("expected default timeout %d, got %d", protocol.DefaultJobContractTimeoutSeconds, collected.TimeoutSeconds)
	}
}

// TestHandleHostServiceInvokeJobsRegisterRejectsRuntimeSource verifies
// ordinary route and task executions cannot mutate dynamic Jobs declarations.
func TestHandleHostServiceInvokeJobsRegisterRejectsRuntimeSource(t *testing.T) {
	collector := &trackingJobCollector{}
	hcc := newJobsRegisterHostCallContext(protocol.ExecutionSourceRoute, collector)

	response := invokeJobsRegisterHostService(t, hcc, &protocol.JobContract{
		Name:        "heartbeat",
		Pattern:     "# */10 * * * *",
		RequestType: "JobHeartbeatReq",
	})
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected runtime jobs.register to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if len(collector.items) != 0 {
		t.Fatalf("expected no collected declarations after runtime denial, got %#v", collector.items)
	}
}

// TestHandleHostServiceInvokeJobsRegisterRejectsMissingAuthorization verifies
// capability grants alone do not authorize the discovery declaration method.
func TestHandleHostServiceInvokeJobsRegisterRejectsMissingAuthorization(t *testing.T) {
	hcc := newJobsRegisterHostCallContext(protocol.ExecutionSourceJobsDiscovery, &trackingJobCollector{})
	hcc.hostServices = []*protocol.HostServiceSpec{{
		Service: protocol.HostServiceJobs,
		Methods: []string{
			protocol.HostServiceMethodJobsBatchGet,
		},
	}}

	response := invokeJobsRegisterHostService(t, hcc, &protocol.JobContract{
		Name:        "heartbeat",
		Pattern:     "# */10 * * * *",
		RequestType: "JobHeartbeatReq",
	})
	if response.Status != protocol.HostCallStatusCapabilityDenied {
		t.Fatalf("expected unauthorized jobs.register to be denied, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// TestHandleHostServiceInvokeJobsRegisterRejectsInvalidContract verifies
// invalid declarations are rejected before reaching the collector.
func TestHandleHostServiceInvokeJobsRegisterRejectsInvalidContract(t *testing.T) {
	collector := &trackingJobCollector{}
	hcc := newJobsRegisterHostCallContext(protocol.ExecutionSourceJobsDiscovery, collector)

	response := invokeJobsRegisterHostService(t, hcc, &protocol.JobContract{
		Name:        "heartbeat",
		RequestType: "JobHeartbeatReq",
	})
	if response.Status != protocol.HostCallStatusInvalidRequest {
		t.Fatalf("expected invalid declaration to be rejected, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if len(collector.items) != 0 {
		t.Fatalf("expected invalid declaration not to reach collector, got %#v", collector.items)
	}
}

// TestHandleHostServiceInvokeRejectsStandaloneCronService verifies the old cron
// host service is not part of the dynamic runtime or discovery dispatcher.
func TestHandleHostServiceInvokeRejectsStandaloneCronService(t *testing.T) {
	response := handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, newJobsRegisterHostCallContext(protocol.ExecutionSourceJobsDiscovery, &trackingJobCollector{})),
		protocol.MarshalHostServiceRequestEnvelope(&protocol.HostServiceRequestEnvelope{
			Service: "cron",
			Method:  "cron.register",
			Payload: protocol.MarshalHostServiceJobsRegisterRequest(&protocol.HostServiceJobsRegisterRequest{
				Contract: &protocol.JobContract{
					Name:        "heartbeat",
					Pattern:     "# */10 * * * *",
					RequestType: "JobHeartbeatReq",
				},
			}),
		}),
	)
	if response.Status != protocol.HostCallStatusNotFound {
		t.Fatalf("expected old cron service to be not found, got status=%d payload=%s", response.Status, string(response.Payload))
	}
}

// newJobsRegisterHostCallContext constructs a Jobs-capable host call context
// for one dynamic discovery execution.
func newJobsRegisterHostCallContext(
	source protocol.ExecutionSource,
	collector JobRegistrationCollector,
) *hostCallContext {
	return &hostCallContext{
		pluginID: "test-plugin-jobs-register",
		capabilities: map[string]struct{}{
			protocol.CapabilityJobs: {},
		},
		hostServices: []*protocol.HostServiceSpec{{
			Service: protocol.HostServiceJobs,
			Methods: []string{
				protocol.HostServiceMethodJobsRegister,
			},
		}},
		executionSource: source,
		jobCollector:    collector,
	}
}

// invokeJobsRegisterHostService dispatches one jobs.register request.
func invokeJobsRegisterHostService(
	t *testing.T,
	hcc *hostCallContext,
	contract *protocol.JobContract,
) *protocol.HostCallResponseEnvelope {
	t.Helper()

	return handleHostServiceInvoke(
		context.Background(),
		withTestHostCallRuntime(t, hcc),
		protocol.MarshalHostServiceRequestEnvelope(&protocol.HostServiceRequestEnvelope{
			Service: protocol.HostServiceJobs,
			Method:  protocol.HostServiceMethodJobsRegister,
			Payload: protocol.MarshalHostServiceJobsRegisterRequest(&protocol.HostServiceJobsRegisterRequest{
				Contract: contract,
			}),
		}),
	)
}

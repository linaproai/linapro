// This file tests cron host service discovery registration flows.

package wasm

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// testCronRegistrationCollector stores discovered contracts for assertions.
type testCronRegistrationCollector struct {
	items []*pluginbridge.CronContract
	err   error
}

// Register stores one received contract or returns the configured error.
func (c *testCronRegistrationCollector) Register(contract *pluginbridge.CronContract) error {
	if c.err != nil {
		return c.err
	}
	if contract == nil {
		return gerror.New("nil cron contract")
	}
	contractSnapshot := *contract
	c.items = append(c.items, &contractSnapshot)
	return nil
}

// TestHandleHostServiceInvokeCronRegister verifies cron discovery executions
// can register one normalized cron contract through the host service.
func TestHandleHostServiceInvokeCronRegister(t *testing.T) {
	collector := &testCronRegistrationCollector{}
	hcc := &hostCallContext{
		pluginID: "plugin-demo-dynamic",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityCron: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceCron,
			Methods: []string{pluginbridge.HostServiceMethodCronRegister},
		}},
		executionSource: pluginbridge.ExecutionSourceCronDiscovery,
		cronCollector:   collector,
	}

	response := invokeCronHostService(
		t,
		hcc,
		pluginbridge.MarshalHostServiceCronRegisterRequest(&pluginbridge.HostServiceCronRegisterRequest{
			Contract: &pluginbridge.CronContract{
				Name:         "heartbeat",
				Pattern:      "# */10 * * * *",
				InternalPath: "cron-heartbeat",
			},
		}),
	)
	if response.Status != pluginbridge.HostCallStatusSuccess {
		t.Fatalf("expected cron.register success, got status=%d payload=%s", response.Status, string(response.Payload))
	}
	if len(collector.items) != 1 {
		t.Fatalf("expected one registered cron contract, got %#v", collector.items)
	}
	if collector.items[0].InternalPath != "/cron-heartbeat" || collector.items[0].Scope != pluginbridge.CronScopeAllNode {
		t.Fatalf("expected registered cron contract to be normalized, got %#v", collector.items[0])
	}
}

// TestHandleHostServiceInvokeCronRegisterRejectsWrongExecutionSource verifies
// cron registration is only available during discovery executions.
func TestHandleHostServiceInvokeCronRegisterRejectsWrongExecutionSource(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "plugin-demo-dynamic",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityCron: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceCron,
			Methods: []string{pluginbridge.HostServiceMethodCronRegister},
		}},
		executionSource: pluginbridge.ExecutionSourceCron,
		cronCollector:   &testCronRegistrationCollector{},
	}

	response := invokeCronHostService(
		t,
		hcc,
		pluginbridge.MarshalHostServiceCronRegisterRequest(&pluginbridge.HostServiceCronRegisterRequest{
			Contract: &pluginbridge.CronContract{
				Name:         "heartbeat",
				Pattern:      "# */10 * * * *",
				InternalPath: "/cron-heartbeat",
			},
		}),
	)
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Fatalf("expected cron.register to reject non-discovery execution, got status=%d", response.Status)
	}
}

// TestHandleHostServiceInvokeCronRegisterRejectsCollectorErrors verifies
// collector validation failures are surfaced as invalid requests.
func TestHandleHostServiceInvokeCronRegisterRejectsCollectorErrors(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "plugin-demo-dynamic",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityCron: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceCron,
			Methods: []string{pluginbridge.HostServiceMethodCronRegister},
		}},
		executionSource: pluginbridge.ExecutionSourceCronDiscovery,
		cronCollector: &testCronRegistrationCollector{
			err: gerror.New("duplicate cron name"),
		},
	}

	response := invokeCronHostService(
		t,
		hcc,
		pluginbridge.MarshalHostServiceCronRegisterRequest(&pluginbridge.HostServiceCronRegisterRequest{
			Contract: &pluginbridge.CronContract{
				Name:         "heartbeat",
				Pattern:      "# */10 * * * *",
				InternalPath: "/cron-heartbeat",
			},
		}),
	)
	if response.Status != pluginbridge.HostCallStatusInvalidRequest {
		t.Fatalf("expected cron.register collector error to map to invalid request, got status=%d", response.Status)
	}
}

// TestHandleHostServiceInvokeCronDiscoveryBlocksNonCronServices verifies the
// reserved discovery execution cannot call other host-service families.
func TestHandleHostServiceInvokeCronDiscoveryBlocksNonCronServices(t *testing.T) {
	hcc := &hostCallContext{
		pluginID: "plugin-demo-dynamic",
		capabilities: map[string]struct{}{
			pluginbridge.CapabilityRuntime: {},
		},
		hostServices: []*pluginbridge.HostServiceSpec{{
			Service: pluginbridge.HostServiceRuntime,
			Methods: []string{pluginbridge.HostServiceMethodRuntimeInfoNow},
		}},
		executionSource: pluginbridge.ExecutionSourceCronDiscovery,
	}

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceRuntime,
		Method:  pluginbridge.HostServiceMethodRuntimeInfoNow,
	}
	response := handleHostServiceInvoke(context.Background(), hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
	if response.Status != pluginbridge.HostCallStatusCapabilityDenied {
		t.Fatalf("expected non-cron host service to be blocked during cron discovery, got status=%d", response.Status)
	}
}

// invokeCronHostService dispatches one cron host-service request and returns
// the raw response envelope for assertions.
func invokeCronHostService(
	t *testing.T,
	hcc *hostCallContext,
	payload []byte,
) *pluginbridge.HostCallResponseEnvelope {
	t.Helper()

	request := &pluginbridge.HostServiceRequestEnvelope{
		Service: pluginbridge.HostServiceCron,
		Method:  pluginbridge.HostServiceMethodCronRegister,
		Payload: payload,
	}
	return handleHostServiceInvoke(context.Background(), hcc, pluginbridge.MarshalHostServiceRequestEnvelope(request))
}

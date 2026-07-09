// This file adapts scheduled-job host-service calls to the shared Jobs
// capability service and dynamic-plugin Jobs discovery collector.

package wasm

import (
	"context"
	"strings"

	jobv1 "lina-core/api/job/v1"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/jobcap"
	bridgecontract "lina-core/pkg/plugin/pluginbridge/contract"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// JobRegistrationCollector captures dynamic-plugin job declarations during one
// host-driven discovery execution.
type JobRegistrationCollector interface {
	// Register validates and stores one discovered job contract.
	Register(contract *bridgecontract.JobContract) error
}

// dispatchJobsHostService routes scheduled-job domain host-service calls.
func dispatchJobsHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if method == bridgehostservice.HostServiceMethodJobsRegister {
		return handleHostJobsRegister(hcc, payload)
	}
	service := jobsServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("jobs")
	}
	switch method {
	case bridgehostservice.HostServiceMethodJobsBatchGet:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.BatchGet(ctx, jobIDs(request.IDs))
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodJobsList:
		var request jobsListRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.List(ctx, jobcap.ListInput{
			Keyword: strings.TrimSpace(request.Keyword),
			Group:   strings.TrimSpace(request.Group),
			Status:  jobv1.Status(strings.TrimSpace(request.Status)),
			Page: capmodel.PageRequest{
				PageNum:  request.PageNum,
				PageSize: request.PageSize,
			},
		})
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodJobsEnsureVisible:
		var request idsRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.EnsureVisible(ctx, jobIDs(request.IDs))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodJobsCreate:
		var request jobcap.SaveInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.Create(ctx, request)
		return domainCapabilityResult(result, err)
	case bridgehostservice.HostServiceMethodJobsUpdate:
		var request jobcap.UpdateInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Update(ctx, request)
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodJobsDelete:
		var request jobIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Delete(ctx, jobcap.JobID(strings.TrimSpace(request.JobID)))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodJobsRun:
		var request jobIDRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.Run(ctx, jobcap.JobID(strings.TrimSpace(request.JobID)))
		return domainCapabilityResult(true, err)
	case bridgehostservice.HostServiceMethodJobsSetStatus:
		var request jobsSetStatusRequest
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		err := service.SetStatus(ctx, jobcap.JobID(strings.TrimSpace(request.JobID)), jobv1.Status(strings.TrimSpace(request.Status)))
		return domainCapabilityResult(true, err)
	default:
		return domainMethodNotFound("jobs", method)
	}
}

// handleHostJobsRegister validates one dynamic-plugin Jobs declaration request
// and forwards it to the current discovery collector.
func handleHostJobsRegister(
	hcc *hostCallContext,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	if hcc == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInternalError, "host call context not available")
	}
	if bridgecontract.NormalizeExecutionSource(hcc.executionSource) != bridgecontract.ExecutionSourceJobsDiscovery {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusCapabilityDenied,
			"jobs.register only supports Jobs discovery executions",
		)
	}
	if hcc.jobCollector == nil {
		return bridgehostcall.NewHostCallErrorResponse(
			bridgehostcall.HostCallStatusInternalError,
			"Jobs discovery collector not configured",
		)
	}

	request, err := bridgehostservice.UnmarshalHostServiceJobsRegisterRequest(payload)
	if err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if request == nil || request.Contract == nil {
		return bridgehostcall.NewHostCallErrorResponse(bridgehostcall.HostCallStatusInvalidRequest, "Jobs registration contract is required")
	}
	contractSnapshot := *request.Contract
	if err = bridgecontract.ValidateJobContracts(hcc.pluginID, []*bridgecontract.JobContract{&contractSnapshot}); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	if err = hcc.jobCollector.Register(&contractSnapshot); err != nil {
		return hostCallErrorFromError(bridgehostcall.HostCallStatusInvalidRequest, err)
	}
	return bridgehostcall.NewHostCallSuccessResponse(nil)
}

// jobsServiceForHostCall resolves the scheduled-job service for one host call.
func jobsServiceForHostCall(hcc *hostCallContext) jobcap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Jobs()
}

// jobIDs converts transport string identifiers into typed job IDs.
func jobIDs(ids []string) []jobcap.JobID {
	out := make([]jobcap.JobID, 0, len(ids))
	for _, id := range ids {
		if value := strings.TrimSpace(id); value != "" {
			out = append(out, jobcap.JobID(value))
		}
	}
	return out
}

// jobsListRequest carries governed scheduled-job search parameters.
type jobsListRequest struct {
	Keyword  string `json:"keyword,omitempty"`
	Group    string `json:"group,omitempty"`
	Status   string `json:"status,omitempty"`
	PageNum  int    `json:"pageNum,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
}

type jobIDRequest struct {
	JobID string `json:"jobId"`
}

type jobsSetStatusRequest struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

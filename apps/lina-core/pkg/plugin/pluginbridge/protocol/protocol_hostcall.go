// protocol_hostcall.go exposes low-level WASM host-call constants, payloads, and codecs through the public protocol facade.
// Keep host-call aliases separate from higher-level host service payloads so ABI maintenance remains localized.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/hostcall"

type (
	HostCallLogRequest         = hostcall.HostCallLogRequest
	HostCallErrorPayload       = hostcall.HostCallErrorPayload
	HostCallResponseEnvelope   = hostcall.HostCallResponseEnvelope
	HostCallStateDeleteRequest = hostcall.HostCallStateDeleteRequest
	HostCallStateGetRequest    = hostcall.HostCallStateGetRequest
	HostCallStateGetResponse   = hostcall.HostCallStateGetResponse
	HostCallStateSetRequest    = hostcall.HostCallStateSetRequest
)

const (
	HostModuleName                  = hostcall.HostModuleName
	HostCallFunctionName            = hostcall.HostCallFunctionName
	DefaultGuestHostCallAllocExport = hostcall.DefaultGuestHostCallAllocExport
	HostCallStatusSuccess           = hostcall.HostCallStatusSuccess
	HostCallStatusCapabilityDenied  = hostcall.HostCallStatusCapabilityDenied
	HostCallStatusNotFound          = hostcall.HostCallStatusNotFound
	HostCallStatusInvalidRequest    = hostcall.HostCallStatusInvalidRequest
	HostCallStatusInternalError     = hostcall.HostCallStatusInternalError
	OpcodeServiceInvoke             = hostcall.OpcodeServiceInvoke
	LogLevelDebug                   = hostcall.LogLevelDebug
	LogLevelInfo                    = hostcall.LogLevelInfo
	LogLevelWarning                 = hostcall.LogLevelWarning
	LogLevelError                   = hostcall.LogLevelError
)

var (
	MarshalHostCallResponse             = hostcall.MarshalHostCallResponse
	UnmarshalHostCallResponse           = hostcall.UnmarshalHostCallResponse
	NewHostCallSuccessResponse          = hostcall.NewHostCallSuccessResponse
	NewHostCallEmptySuccessResponse     = hostcall.NewHostCallEmptySuccessResponse
	NewHostCallErrorResponse            = hostcall.NewHostCallErrorResponse
	NewHostCallErrorResponseFromError   = hostcall.NewHostCallErrorResponseFromError
	NewHostCallErrorPayloadResponse     = hostcall.NewHostCallErrorPayloadResponse
	UnmarshalHostCallErrorPayload       = hostcall.UnmarshalHostCallErrorPayload
	MarshalHostCallLogRequest           = hostcall.MarshalHostCallLogRequest
	UnmarshalHostCallLogRequest         = hostcall.UnmarshalHostCallLogRequest
	MarshalHostCallStateGetRequest      = hostcall.MarshalHostCallStateGetRequest
	UnmarshalHostCallStateGetRequest    = hostcall.UnmarshalHostCallStateGetRequest
	MarshalHostCallStateGetResponse     = hostcall.MarshalHostCallStateGetResponse
	UnmarshalHostCallStateGetResponse   = hostcall.UnmarshalHostCallStateGetResponse
	MarshalHostCallStateSetRequest      = hostcall.MarshalHostCallStateSetRequest
	UnmarshalHostCallStateSetRequest    = hostcall.UnmarshalHostCallStateSetRequest
	MarshalHostCallStateDeleteRequest   = hostcall.MarshalHostCallStateDeleteRequest
	UnmarshalHostCallStateDeleteRequest = hostcall.UnmarshalHostCallStateDeleteRequest
)

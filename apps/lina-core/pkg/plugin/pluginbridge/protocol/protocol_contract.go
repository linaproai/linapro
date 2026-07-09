// protocol_contract.go exposes bridge contract aliases, constants, and validators through the public protocol facade.
// Keep these declarations as direct one-to-one aliases so protocol callers do not depend on internal bridge subpackages.

package protocol

import "lina-core/pkg/plugin/pluginbridge/contract"

type (
	BridgeFailureV1          = contract.BridgeFailureV1
	BridgeRequestEnvelopeV1  = contract.BridgeRequestEnvelopeV1
	BridgeResponseEnvelopeV1 = contract.BridgeResponseEnvelopeV1
	BridgeSpec               = contract.BridgeSpec
	ExecutionSource          = contract.ExecutionSource
	HTTPRequestSnapshotV1    = contract.HTTPRequestSnapshotV1
	IdentitySnapshotV1       = contract.IdentitySnapshotV1
	JobConcurrency           = contract.JobConcurrency
	JobContract              = contract.JobContract
	JobScope                 = contract.JobScope
	LifecycleContract        = contract.LifecycleContract
	LifecycleDecision        = contract.LifecycleDecision
	LifecycleOperation       = contract.LifecycleOperation
	LifecycleRequest         = contract.LifecycleRequest
	ManifestSnapshotV1       = contract.ManifestSnapshotV1
	RouteContract            = contract.RouteContract
	RouteMatchSnapshotV1     = contract.RouteMatchSnapshotV1
)

const (
	CodecProtobuf                 = contract.CodecProtobuf
	AccessPublic                  = contract.AccessPublic
	AccessLogin                   = contract.AccessLogin
	RuntimeKindWasm               = contract.RuntimeKindWasm
	ABIVersionV1                  = contract.ABIVersionV1
	SupportedABIVersion           = contract.SupportedABIVersion
	DefaultGuestAllocExport       = contract.DefaultGuestAllocExport
	DefaultGuestExecuteExport     = contract.DefaultGuestExecuteExport
	BridgeFailureCodeUnauthorized = contract.BridgeFailureCodeUnauthorized
	BridgeFailureCodeForbidden    = contract.BridgeFailureCodeForbidden
	BridgeFailureCodeBadRequest   = contract.BridgeFailureCodeBadRequest
	BridgeFailureCodeNotFound     = contract.BridgeFailureCodeNotFound
	BridgeFailureCodeInternal     = contract.BridgeFailureCodeInternal

	ExecutionSourceRoute         = contract.ExecutionSourceRoute
	ExecutionSourceHook          = contract.ExecutionSourceHook
	ExecutionSourceJobs          = contract.ExecutionSourceJobs
	ExecutionSourceJobsDiscovery = contract.ExecutionSourceJobsDiscovery
	ExecutionSourceLifecycle     = contract.ExecutionSourceLifecycle

	DefaultJobContractTimezone          = contract.DefaultJobContractTimezone
	DefaultJobContractTimeoutSeconds    = contract.DefaultJobContractTimeoutSeconds
	DeclaredJobRouteBasePath            = contract.DeclaredJobRouteBasePath
	DeclaredJobRegistrationInternalPath = contract.DeclaredJobRegistrationInternalPath
	DeclaredJobRegistrationRequestType  = contract.DeclaredJobRegistrationRequestType
	JobScopeMasterOnly                  = contract.JobScopeMasterOnly
	JobScopeAllNode                     = contract.JobScopeAllNode
	JobConcurrencySingleton             = contract.JobConcurrencySingleton
	JobConcurrencyParallel              = contract.JobConcurrencyParallel

	LifecycleOperationBeforeInstall           = contract.LifecycleOperationBeforeInstall
	LifecycleOperationAfterInstall            = contract.LifecycleOperationAfterInstall
	LifecycleOperationBeforeUpgrade           = contract.LifecycleOperationBeforeUpgrade
	LifecycleOperationUpgrade                 = contract.LifecycleOperationUpgrade
	LifecycleOperationAfterUpgrade            = contract.LifecycleOperationAfterUpgrade
	LifecycleOperationBeforeDisable           = contract.LifecycleOperationBeforeDisable
	LifecycleOperationAfterDisable            = contract.LifecycleOperationAfterDisable
	LifecycleOperationBeforeUninstall         = contract.LifecycleOperationBeforeUninstall
	LifecycleOperationUninstall               = contract.LifecycleOperationUninstall
	LifecycleOperationAfterUninstall          = contract.LifecycleOperationAfterUninstall
	LifecycleOperationBeforeTenantDisable     = contract.LifecycleOperationBeforeTenantDisable
	LifecycleOperationAfterTenantDisable      = contract.LifecycleOperationAfterTenantDisable
	LifecycleOperationBeforeTenantDelete      = contract.LifecycleOperationBeforeTenantDelete
	LifecycleOperationAfterTenantDelete       = contract.LifecycleOperationAfterTenantDelete
	LifecycleOperationBeforeInstallModeChange = contract.LifecycleOperationBeforeInstallModeChange
	LifecycleOperationAfterInstallModeChange  = contract.LifecycleOperationAfterInstallModeChange
)

var (
	ValidateRouteContracts        = contract.ValidateRouteContracts
	NormalizeBridgeSpec           = contract.NormalizeBridgeSpec
	ValidateBridgeSpec            = contract.ValidateBridgeSpec
	NormalizeLifecycleContract    = contract.NormalizeLifecycleContract
	ValidateLifecycleContracts    = contract.ValidateLifecycleContracts
	IsSupportedLifecycleOperation = contract.IsSupportedLifecycleOperation
	NormalizeJobScope             = contract.NormalizeJobScope
	NormalizeJobConcurrency       = contract.NormalizeJobConcurrency
	NormalizeJobContract          = contract.NormalizeJobContract
	ValidateJobContracts          = contract.ValidateJobContracts
	BuildPluginJobHandlerRef      = contract.BuildPluginJobHandlerRef
	BuildDeclaredJobRoutePath     = contract.BuildDeclaredJobRoutePath
	NormalizeExecutionSource      = contract.NormalizeExecutionSource
)

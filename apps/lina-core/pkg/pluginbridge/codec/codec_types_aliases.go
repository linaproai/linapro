// This file aliases shared bridge contract types for codec implementation.

package codec

import "lina-core/pkg/pluginbridge/contract"

type (
	BridgeSpec              = contract.BridgeSpec
	RouteContract           = contract.RouteContract
	BridgeRequestEnvelopeV1 = contract.BridgeRequestEnvelopeV1
	RouteMatchSnapshotV1    = contract.RouteMatchSnapshotV1
	HTTPRequestSnapshotV1   = contract.HTTPRequestSnapshotV1
	IdentitySnapshotV1      = contract.IdentitySnapshotV1
	BridgeResponseEnvelopeV1 = contract.BridgeResponseEnvelopeV1
	BridgeFailureV1         = contract.BridgeFailureV1
)

const (
	CodecProtobuf             = contract.CodecProtobuf
	AccessPublic              = contract.AccessPublic
	AccessLogin               = contract.AccessLogin
	RuntimeKindWasm           = contract.RuntimeKindWasm
	ABIVersionV1              = contract.ABIVersionV1
	SupportedABIVersion       = contract.SupportedABIVersion
	DefaultGuestAllocExport    = contract.DefaultGuestAllocExport
	DefaultGuestExecuteExport  = contract.DefaultGuestExecuteExport
	bridgeFailureCodeUnauthorized = contract.BridgeFailureCodeUnauthorized
	bridgeFailureCodeForbidden    = contract.BridgeFailureCodeForbidden
	bridgeFailureCodeBadRequest   = contract.BridgeFailureCodeBadRequest
	bridgeFailureCodeNotFound     = contract.BridgeFailureCodeNotFound
	bridgeFailureCodeInternal     = contract.BridgeFailureCodeInternal
)

var (
	ValidateRouteContracts = contract.ValidateRouteContracts
	ValidateBridgeSpec     = contract.ValidateBridgeSpec
)

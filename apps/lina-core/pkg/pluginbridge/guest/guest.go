// Package guest provides the guest-side runtime, controller dispatcher, and
// host-service clients used by Lina dynamic plugin modules.
package guest

// guestRequestBuffer, guestResponseBuffer, and guestHostCallResponseBuffer are
// package-level globals reused across bridge invocations. They are safe ONLY
// within a single-threaded Wasm guest module. The host MUST NOT invoke the same
// module instance concurrently; each concurrent request should use a separate
// wazero module instantiation.
//
// guestHostCallResponseBuffer is separate from guestResponseBuffer to avoid
// conflicts when host functions are called during execute() processing, since
// the host writes host call responses into this buffer via re-entrant alloc.
//
// These globals intentionally remain package-level so the guest exports can
// exchange stable pointers with the host runtime.
var (
	guestRequestBuffer          []byte
	guestResponseBuffer         []byte
	guestHostCallResponseBuffer []byte
)

// GuestHandler defines the guest-side dynamic route handler interface.
type GuestHandler func(*BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)

// GuestRuntime exposes the guest-side request buffer and execution contract
// published to dynamic plugin entrypoints.
type GuestRuntime interface {
	// HandleEncodedRequest decodes one host request, executes the guest handler,
	// and returns encoded response bytes. Decode errors become bad-request bridge
	// responses; handler errors become internal-error bridge responses.
	HandleEncodedRequest(content []byte) ([]byte, error)
	// Alloc reserves guest memory for the next incoming request and returns the
	// guest pointer exposed to the host. The buffer is process-local to the Wasm
	// instance and is not safe for concurrent host calls into the same instance.
	Alloc(size uint32) uint32
	// RequestBuffer returns the mutable request buffer currently exposed to the host.
	// Callers must treat the slice as the current invocation buffer only.
	RequestBuffer() []byte
	// Execute handles the currently written request buffer and exposes the encoded
	// response buffer. It returns pointer and length values for the host, or an
	// error when length exceeds the allocated request buffer.
	Execute(length uint32) (uint32, uint32, error)
	// ResponseBuffer returns the current encoded response buffer produced by the
	// most recent Execute or ExposeResponseBuffer call.
	ResponseBuffer() []byte
	// ExposeResponseBuffer publishes one encoded response payload through the
	// shared guest response buffer and returns its pointer-length pair.
	ExposeResponseBuffer(content []byte) (uint32, uint32, error)
	// HostCallAlloc reserves guest memory for an incoming host call response.
	// It uses a separate buffer from Alloc so re-entrant host-service calls do not
	// overwrite the active request payload.
	HostCallAlloc(size uint32) uint32
	// HostCallResponseBuffer returns the current host call response buffer written
	// by the host-service bridge.
	HostCallResponseBuffer() []byte
}

// guestRuntime hosts one guest-side request dispatcher.
type guestRuntime struct {
	handler GuestHandler
}

// NewGuestRuntime creates one guest runtime wrapper around a business handler.
func NewGuestRuntime(handler GuestHandler) GuestRuntime {
	return &guestRuntime{handler: handler}
}

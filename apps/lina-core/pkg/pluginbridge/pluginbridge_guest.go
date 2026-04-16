// This file provides the guest runtime buffers and encoded request execution
// helpers used by plugin bridge guest modules.

package pluginbridge

import (
	"unsafe"

	"github.com/gogf/gf/v2/errors/gerror"
)

// guestRequestBuffer, guestResponseBuffer, and guestHostCallResponseBuffer are
// package-level globals reused across bridge invocations. They are safe ONLY
// within a single-threaded Wasm guest module. The host MUST NOT invoke the same
// module instance concurrently; each concurrent request should use a separate
// wazero module instantiation.
//
// guestHostCallResponseBuffer is separate from guestResponseBuffer to avoid
// conflicts when host functions are called during execute() processing, since
// the host writes host call responses into this buffer via re-entrant alloc.
var guestRequestBuffer []byte
var guestResponseBuffer []byte
var guestHostCallResponseBuffer []byte

// GuestHandler defines the guest-side dynamic route handler interface.
type GuestHandler func(*BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)

// GuestRuntime hosts one guest-side request dispatcher.
type GuestRuntime struct {
	handler GuestHandler
}

// NewGuestRuntime creates one guest runtime wrapper around a business handler.
func NewGuestRuntime(handler GuestHandler) *GuestRuntime {
	return &GuestRuntime{handler: handler}
}

// HandleEncodedRequest decodes one host request, executes the guest handler, and returns encoded response bytes.
func (r *GuestRuntime) HandleEncodedRequest(content []byte) ([]byte, error) {
	if r == nil || r.handler == nil {
		return EncodeResponseEnvelope(NewInternalErrorResponse("Dynamic guest runtime is not initialized"))
	}

	request, err := DecodeRequestEnvelope(content)
	if err != nil {
		return EncodeResponseEnvelope(NewBadRequestResponse(err.Error()))
	}
	response, err := r.handler(request)
	if err != nil {
		return EncodeResponseEnvelope(NewInternalErrorResponse(err.Error()))
	}
	if response == nil {
		response = NewInternalErrorResponse("Dynamic guest runtime returned nil response")
	}
	return EncodeResponseEnvelope(response)
}

// Alloc reserves guest memory for the next incoming request.
func (*GuestRuntime) Alloc(size uint32) uint32 {
	if cap(guestRequestBuffer) < int(size) {
		guestRequestBuffer = make([]byte, size)
	} else {
		guestRequestBuffer = guestRequestBuffer[:size]
	}
	if len(guestRequestBuffer) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&guestRequestBuffer[0])))
}

// RequestBuffer returns the mutable request buffer currently exposed to the host.
func (*GuestRuntime) RequestBuffer() []byte {
	return guestRequestBuffer
}

// Execute handles the currently written request buffer and exposes the encoded response buffer.
func (r *GuestRuntime) Execute(length uint32) (uint32, uint32, error) {
	if int(length) > len(guestRequestBuffer) {
		return 0, 0, gerror.New("guest request length exceeds allocated buffer")
	}
	response, err := r.HandleEncodedRequest(guestRequestBuffer[:length])
	if err != nil {
		return 0, 0, err
	}
	return r.ExposeResponseBuffer(response)
}

// ResponseBuffer returns the current encoded response buffer.
func (*GuestRuntime) ResponseBuffer() []byte {
	return guestResponseBuffer
}

// ExposeResponseBuffer publishes one encoded response payload through the
// shared guest response buffer and returns the stable pointer-length pair.
func (*GuestRuntime) ExposeResponseBuffer(content []byte) (uint32, uint32, error) {
	guestResponseBuffer = append(guestResponseBuffer[:0], content...)
	if len(guestResponseBuffer) == 0 {
		return 0, 0, nil
	}
	return uint32(uintptr(unsafe.Pointer(&guestResponseBuffer[0]))), uint32(len(guestResponseBuffer)), nil
}

// HostCallAlloc reserves guest memory for an incoming host call response.
// This uses a separate buffer from Alloc to avoid overwriting the in-flight
// request data during re-entrant host function calls.
func (*GuestRuntime) HostCallAlloc(size uint32) uint32 {
	if cap(guestHostCallResponseBuffer) < int(size) {
		guestHostCallResponseBuffer = make([]byte, size)
	} else {
		guestHostCallResponseBuffer = guestHostCallResponseBuffer[:size]
	}
	if len(guestHostCallResponseBuffer) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&guestHostCallResponseBuffer[0])))
}

// HostCallResponseBuffer returns the current host call response buffer.
func (*GuestRuntime) HostCallResponseBuffer() []byte {
	return guestHostCallResponseBuffer
}

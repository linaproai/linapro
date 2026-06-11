// This file preserves historical JSON capability envelope names while ordinary
// domain host-service payloads are carried by the generic JSON envelope.

package protocol

// HostServiceCapabilityJSONRequest preserves the historical name for the
// generic JSON host-service request envelope.
type HostServiceCapabilityJSONRequest = HostServiceJSONRequest

// HostServiceCapabilityJSONResponse preserves the historical name for the
// generic JSON host-service response envelope.
type HostServiceCapabilityJSONResponse = HostServiceJSONResponse

// MarshalHostServiceCapabilityJSONRequest encodes one JSON value request.
func MarshalHostServiceCapabilityJSONRequest(req *HostServiceCapabilityJSONRequest) []byte {
	return MarshalHostServiceJSONRequest(req)
}

// UnmarshalHostServiceCapabilityJSONRequest decodes one JSON value request.
func UnmarshalHostServiceCapabilityJSONRequest(data []byte) (*HostServiceCapabilityJSONRequest, error) {
	return UnmarshalHostServiceJSONRequest(data)
}

// MarshalHostServiceCapabilityJSONResponse encodes one JSON value response.
func MarshalHostServiceCapabilityJSONResponse(resp *HostServiceCapabilityJSONResponse) []byte {
	return MarshalHostServiceJSONResponse(resp)
}

// UnmarshalHostServiceCapabilityJSONResponse decodes one JSON value response.
func UnmarshalHostServiceCapabilityJSONResponse(data []byte) (*HostServiceCapabilityJSONResponse, error) {
	return UnmarshalHostServiceJSONResponse(data)
}

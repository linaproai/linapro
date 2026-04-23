## MODIFIED Requirements

### Requirement: Dynamic plugins reuse public bridge components to reduce writing complexity

The system SHALL abstracts the dynamic plugin bridge envelope, binary codec, guest-side processor adaptation, error response assistance, and typed guest controller adaptation into `apps/lina-core/pkg` public components, preventing plugin authors from repeatedly writing the underlying ABI, codec templates, and manual conversion logic from envelope to API DTO in each dynamic plugin.

#### Scenario: Dynamic plugin controller directly reuses API request and response DTO

- **WHEN** Developer implements guest controller for a dynamic plugin route that has declared DTO in `backend/api/.../v1`
- **THEN** The controller can declare methods using the form `func(ctx context.Context, req *v1.XxxReq) (res *v1.XxxRes, err error)`
- **AND** guest route dispatcher matches runtime `RequestType` based on request DTO type name
- **AND** The dynamic routing contract built by the host continues to reuse the same API DTO metadata

#### Scenario: typed guest controller can still access the bridge context and write custom responses

- **WHEN** typed guest controller needs to read `pluginId`, `requestId`, identity snapshot, routing metadata, or return download stream / additional response header / custom status code
- **THEN** `pkg/pluginbridge` MUST provide helper methods for reading the bridge envelope from `context.Context`
- **AND** The component MUST provide helper methods for writing response headers, raw response bodies, or custom status codes
- **AND** Plugin authors do not need to fall back to directly declaring `func(*BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)` to complete these scenarios

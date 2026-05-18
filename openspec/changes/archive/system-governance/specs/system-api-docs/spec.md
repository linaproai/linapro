## ADDED Requirements

### Requirement: System API Documentation Page Display
The system SHALL provide a "System API Docs" page that displays backend API documentation through an iframe embedding of a Stoplight Elements rendered static HTML document. The document data comes from the backend GoFrame auto-generated `/api.json` OpenAPI v3 specification file.

#### Scenario: Normal Loading of API Documentation
- **WHEN** user clicks "System Info > System API Docs" menu
- **THEN** the page loads the static HTML file via iframe, displaying the Stoplight Elements rendered API documentation, including all backend endpoint paths, parameters, and response definitions

### Requirement: Online API Testing
Stoplight Elements SHALL support users testing endpoints directly on the documentation page without navigating to third-party tools.

#### Scenario: Online Testing of Endpoints
- **WHEN** user selects an endpoint on the documentation page and attempts to send a request
- **THEN** the page displays a request parameter input area; after filling in parameters, the user can send a request and view the response. The documentation uses the current logged-in user's JWT Token information for convenient direct endpoint testing

### Requirement: API Documentation Address Configurable
The OpenAPI specification file address loaded by the API documentation SHALL be specified through frontend configuration, not hardcoded in components.

#### Scenario: Modify API Documentation Address
- **WHEN** developer modifies the OpenAPI specification file address in frontend configuration
- **THEN** the documentation page loads the API documentation from the new address

### Requirement: API Documentation Style Isolation
The API documentation SHALL achieve complete style isolation through iframe embedding, preventing the document component's CSS from polluting global page styles.

#### Scenario: Style Isolation
- **WHEN** user accesses the System API Docs page
- **THEN** the API documentation's styles do not affect the system's other pages' borders, fonts, colors, or other style presentations

### Requirement: API Documentation Request Address Prefix Must Be Dynamically Generated from Access Entry
The system SHALL generate `servers[0].url` based on the current request's scheme, host, and port when returning the host `/api.json` OpenAPI document, ensuring that the request address prefix displayed for each endpoint in the API documentation and the Try It request target match the current deployment entry. This runtime address MUST NOT rely on a fixed hardcoded `localhost` address.

#### Scenario: Access API Documentation Through Frontend Dev Service Proxy
- **WHEN** admin opens the System API Docs page through the frontend dev service, and `/api.json` is proxied from the frontend service to the backend service
- **THEN** the `servers[0].url` returned by `/api.json` uses the scheme, host, and port that the backend service received with the request
- **AND** the request address prefix displayed and used for Try It in the API documentation points to the backend service address

#### Scenario: Access API Documentation Through Backend Direct Connection
- **WHEN** admin opens the System API Docs page by directly accessing the backend service exposed address, e.g., container maps `8080` to host `8088`
- **THEN** the `servers[0].url` returned by `/api.json` matches the scheme, host, and port used when the browser accesses the backend service
- **AND** the API documentation no longer displays the fixed `http://localhost:8080` prefix

#### Scenario: Access API Documentation Through HTTPS Proxy
- **WHEN** admin requests `/api.json` through a proxy that sets `X-Forwarded-Proto: https`
- **THEN** the `servers[0].url` returned by `/api.json` uses the `https` scheme
- **AND** the host and port retain the values provided by the request entry

// This file exposes OpenAPI document metadata sourced from the embedded
// metadata asset.

package config

import "context"

// OpenApiConfig holds OpenAPI documentation configuration.
type OpenApiConfig struct {
	Title             string `json:"title"`             // Title is the API document title.
	Description       string `json:"description"`       // Description is the API document summary.
	Version           string `json:"version"`           // Version is the API document version string.
	ServerUrl         string `json:"serverUrl"`         // ServerUrl is the backend endpoint shown by OpenAPI viewers.
	ServerDescription string `json:"serverDescription"` // ServerDescription is the display name for ServerUrl.
}

func defaultOpenApiConfig() OpenApiConfig {
	return OpenApiConfig{
		Title:             "Lina Framework API",
		Description:       "Lina core host service API reference.",
		Version:           "v1.0.0",
		ServerDescription: "Core Host API Server",
	}
}

// GetOpenApi reads OpenAPI config from embedded metadata.
func (s *serviceImpl) GetOpenApi(ctx context.Context) *OpenApiConfig {
	cfg := s.GetMetadata(ctx).OpenApi
	return &cfg
}

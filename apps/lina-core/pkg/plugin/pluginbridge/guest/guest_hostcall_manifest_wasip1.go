//go:build wasip1

// This file provides guest-side helpers for plugin manifest resources.

package guest

import (
	"strings"

	"lina-core/pkg/plugin/pluginbridge/protocol"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"
)

// manifestHostService is the default guest-side manifest client.
type manifestHostService struct{}

// defaultManifestHostService stores the singleton manifest client.
var defaultManifestHostService ManifestHostService = &manifestHostService{}

// Manifest returns the plugin manifest-resource guest client.
func Manifest() ManifestHostService {
	return defaultManifestHostService
}

// Get reads one manifest resource as bytes.
func (*manifestHostService) Get(path string) ([]byte, bool, error) {
	request := &protocol.HostServiceManifestGetRequest{Path: path}
	payload, err := invokeGuestHostService(
		protocol.HostServiceManifest,
		protocol.HostServiceMethodManifestGet,
		path,
		"",
		protocol.MarshalHostServiceManifestGetRequest(request),
	)
	if err != nil {
		return nil, false, err
	}
	response, err := protocol.UnmarshalHostServiceManifestGetResponse(payload)
	if err != nil {
		return nil, false, err
	}
	if response == nil || !response.Found {
		return nil, false, nil
	}
	return response.Body, true, nil
}

// GetText reads one manifest resource as UTF-8 text.
func (s *manifestHostService) GetText(path string) (string, bool, error) {
	body, found, err := s.Get(path)
	if err != nil || !found {
		return "", found, err
	}
	return string(body), true, nil
}

// Scan decodes a YAML manifest resource or nested key into target.
func (s *manifestHostService) Scan(path string, key string, target any) (bool, error) {
	if target == nil {
		return false, gerror.New("manifest scan target cannot be nil")
	}
	body, found, err := s.Get(path)
	if err != nil || !found {
		return found, err
	}
	if strings.TrimSpace(key) == "" {
		if err = yaml.Unmarshal(body, target); err != nil {
			return true, gerror.Wrapf(err, "scan manifest resource failed path=%s", path)
		}
		return true, nil
	}
	jsonDoc, err := gjson.LoadYaml(body)
	if err != nil {
		return true, gerror.Wrapf(err, "parse manifest resource failed path=%s", path)
	}
	if err = jsonDoc.Get(strings.TrimSpace(key)).Scan(target); err != nil {
		return true, gerror.Wrapf(err, "scan manifest resource failed path=%s key=%s", path, key)
	}
	return true, nil
}

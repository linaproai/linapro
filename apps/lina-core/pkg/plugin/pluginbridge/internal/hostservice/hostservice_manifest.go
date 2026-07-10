// This file customizes JSON/YAML serialization for host service declarations
// so plugin manifests can keep one stable `resources` envelope while allowing
// data-specific table requests under `resources.tables`.

package hostservice

import (
	"bytes"
	"encoding/json"
	"lina-core/pkg/plugin/pluginbridge/protocol/hostservices"

	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"
)

// hostServiceSpecWire is the JSON decoding shape for host service declarations.
type hostServiceSpecWire struct {
	Owner     string          `json:"owner,omitempty" yaml:"owner,omitempty"`
	Service   string          `json:"service" yaml:"service"`
	Version   string          `json:"version,omitempty" yaml:"version,omitempty"`
	Methods   []string        `json:"methods" yaml:"methods"`
	Resources json.RawMessage `json:"resources,omitempty" yaml:"-"`
}

// hostServiceStorageResourcesWire mirrors the manifest representation for
// storage path declarations nested under `resources`.
type hostServiceStorageResourcesWire struct {
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

// hostServiceDataResourcesWire mirrors the manifest representation for data
// table declarations nested under `resources`.
type hostServiceDataResourcesWire struct {
	Tables []string `json:"tables,omitempty" yaml:"tables,omitempty"`
}

// hostServiceKeyResourcesWire mirrors the manifest representation for host
// runtime key declarations nested under `resources`.
type hostServiceKeyResourcesWire struct {
	Keys []string `json:"keys,omitempty" yaml:"keys,omitempty"`
}

// hostServiceNetworkResourceWire is the manifest-facing network resource item.
type hostServiceNetworkResourceWire struct {
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
}

// MarshalJSON serializes host service declarations using the manifest-facing
// `resources` envelope. Storage services emit `resources.paths`, data services
// emit `resources.tables`, while network continues to emit URL resources.
func (spec HostServiceSpec) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"service": spec.Service,
		"methods": spec.Methods,
	}
	if spec.Owner != "" {
		payload["owner"] = spec.Owner
	}
	if spec.Version != "" {
		payload["version"] = spec.Version
	}
	if len(spec.Paths) > 0 {
		payload["resources"] = &hostServiceStorageResourcesWire{Paths: spec.Paths}
	} else if len(spec.Tables) > 0 {
		payload["resources"] = &hostServiceDataResourcesWire{Tables: spec.Tables}
	} else if len(spec.Keys) > 0 {
		payload["resources"] = &hostServiceKeyResourcesWire{Keys: spec.Keys}
	} else if spec.Service == hostservices.HostServiceNetwork && len(spec.Resources) > 0 {
		payload["resources"] = marshalNetworkResources(spec.Resources)
	} else if len(spec.Resources) > 0 {
		payload["resources"] = spec.Resources
	}
	return json.Marshal(payload)
}

// UnmarshalJSON restores one host service declaration from the current manifest shape.
func (spec *HostServiceSpec) UnmarshalJSON(data []byte) error {
	var wire hostServiceSpecWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	spec.Service = wire.Service
	spec.Owner = wire.Owner
	spec.Version = wire.Version
	spec.Methods = append([]string(nil), wire.Methods...)
	spec.Paths = nil
	spec.Tables = nil
	spec.Keys = nil
	spec.Resources = nil

	if len(bytes.TrimSpace(wire.Resources)) == 0 {
		return nil
	}

	trimmed := bytes.TrimSpace(wire.Resources)
	switch trimmed[0] {
	case '[':
		resources, err := unmarshalJSONResourcesByService(spec.Service, trimmed)
		if err != nil {
			return err
		}
		spec.Resources = resources
	case '{':
		switch normalizeHostServiceName(spec.Service) {
		case hostservices.HostServiceStorage, hostservices.HostServiceManifest:
			var storageResources hostServiceStorageResourcesWire
			if err := json.Unmarshal(trimmed, &storageResources); err != nil {
				return err
			}
			spec.Paths = append([]string(nil), storageResources.Paths...)
			return nil
		case hostservices.HostServiceHostConfig:
			var keyResources hostServiceKeyResourcesWire
			if err := json.Unmarshal(trimmed, &keyResources); err != nil {
				return err
			}
			spec.Keys = append([]string(nil), keyResources.Keys...)
			return nil
		}
		var dataResources hostServiceDataResourcesWire
		if err := json.Unmarshal(trimmed, &dataResources); err != nil {
			return err
		}
		spec.Tables = append([]string(nil), dataResources.Tables...)
	default:
		return gerror.New("host service resources must be an array or object")
	}
	return nil
}

// MarshalYAML serializes host service declarations using the same manifest
// shape as JSON so plugin.yaml and embedded artifact snapshots stay aligned.
func (spec HostServiceSpec) MarshalYAML() (interface{}, error) {
	payload := map[string]interface{}{
		"service": spec.Service,
		"methods": spec.Methods,
	}
	if spec.Owner != "" {
		payload["owner"] = spec.Owner
	}
	if spec.Version != "" {
		payload["version"] = spec.Version
	}
	if len(spec.Paths) > 0 {
		payload["resources"] = &hostServiceStorageResourcesWire{Paths: spec.Paths}
	} else if len(spec.Tables) > 0 {
		payload["resources"] = &hostServiceDataResourcesWire{Tables: spec.Tables}
	} else if len(spec.Keys) > 0 {
		payload["resources"] = &hostServiceKeyResourcesWire{Keys: spec.Keys}
	} else if spec.Service == hostservices.HostServiceNetwork && len(spec.Resources) > 0 {
		payload["resources"] = marshalNetworkResources(spec.Resources)
	} else if len(spec.Resources) > 0 {
		payload["resources"] = spec.Resources
	}
	return payload, nil
}

// UnmarshalYAML restores one host service declaration from plugin.yaml using
// the unified `resources` envelope.
func (spec *HostServiceSpec) UnmarshalYAML(node *yaml.Node) error {
	type hostServiceSpecYAMLWire struct {
		Owner     string    `yaml:"owner,omitempty"`
		Service   string    `yaml:"service"`
		Version   string    `yaml:"version,omitempty"`
		Methods   []string  `yaml:"methods"`
		Resources yaml.Node `yaml:"resources,omitempty"`
	}

	var wire hostServiceSpecYAMLWire
	if err := node.Decode(&wire); err != nil {
		return err
	}

	spec.Service = wire.Service
	spec.Owner = wire.Owner
	spec.Version = wire.Version
	spec.Methods = append([]string(nil), wire.Methods...)
	spec.Paths = nil
	spec.Tables = nil
	spec.Keys = nil
	spec.Resources = nil

	if wire.Resources.Kind == 0 {
		return nil
	}

	switch wire.Resources.Kind {
	case yaml.SequenceNode:
		resources, err := unmarshalYAMLResourcesByService(spec.Service, &wire.Resources)
		if err != nil {
			return err
		}
		spec.Resources = resources
	case yaml.MappingNode:
		switch normalizeHostServiceName(spec.Service) {
		case hostservices.HostServiceStorage, hostservices.HostServiceManifest:
			var storageResources hostServiceStorageResourcesWire
			if err := wire.Resources.Decode(&storageResources); err != nil {
				return err
			}
			spec.Paths = append([]string(nil), storageResources.Paths...)
			return nil
		case hostservices.HostServiceHostConfig:
			var keyResources hostServiceKeyResourcesWire
			if err := wire.Resources.Decode(&keyResources); err != nil {
				return err
			}
			spec.Keys = append([]string(nil), keyResources.Keys...)
			return nil
		}
		var dataResources hostServiceDataResourcesWire
		if err := wire.Resources.Decode(&dataResources); err != nil {
			return err
		}
		spec.Tables = append([]string(nil), dataResources.Tables...)
	default:
		return &yaml.TypeError{Errors: []string{"host service resources must be a sequence or mapping"}}
	}
	return nil
}

// marshalNetworkResources converts normalized network resources back into the
// manifest wire shape that exposes URL entries.
func marshalNetworkResources(resources []*HostServiceResourceSpec) []*hostServiceNetworkResourceWire {
	if len(resources) == 0 {
		return nil
	}
	items := make([]*hostServiceNetworkResourceWire, 0, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		items = append(items, &hostServiceNetworkResourceWire{
			URL: resource.Ref,
		})
	}
	return items
}

// unmarshalJSONResourcesByService decodes service-specific JSON `resources`
// payloads into normalized host resource declarations.
func unmarshalJSONResourcesByService(service string, payload []byte) ([]*HostServiceResourceSpec, error) {
	if normalizeHostServiceName(service) == hostservices.HostServiceNetwork {
		var resources []*hostServiceNetworkResourceWire
		if err := json.Unmarshal(payload, &resources); err != nil {
			return nil, err
		}
		return normalizeNetworkWireResources(resources), nil
	}

	var resources []*HostServiceResourceSpec
	if err := json.Unmarshal(payload, &resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// unmarshalYAMLResourcesByService decodes service-specific YAML `resources`
// payloads into normalized host resource declarations.
func unmarshalYAMLResourcesByService(service string, node *yaml.Node) ([]*HostServiceResourceSpec, error) {
	if normalizeHostServiceName(service) == hostservices.HostServiceNetwork {
		var resources []*hostServiceNetworkResourceWire
		if err := node.Decode(&resources); err != nil {
			return nil, err
		}
		return normalizeNetworkWireResources(resources), nil
	}

	var resources []*HostServiceResourceSpec
	if err := node.Decode(&resources); err != nil {
		return nil, err
	}
	return resources, nil
}

// normalizeNetworkWireResources converts manifest network URL entries into the
// normalized resource spec shape used by validation and runtime code.
func normalizeNetworkWireResources(resources []*hostServiceNetworkResourceWire) []*HostServiceResourceSpec {
	if len(resources) == 0 {
		return nil
	}
	items := make([]*HostServiceResourceSpec, 0, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		target := resource.URL
		items = append(items, &HostServiceResourceSpec{
			Ref: target,
		})
	}
	return items
}

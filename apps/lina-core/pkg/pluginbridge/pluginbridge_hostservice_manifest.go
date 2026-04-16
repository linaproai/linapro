// This file customizes JSON/YAML serialization for host service declarations
// so plugin manifests can keep one stable `resources` envelope while allowing
// data-specific table requests under `resources.tables`.

package pluginbridge

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"
)

type hostServiceSpecWire struct {
	Service   string          `json:"service" yaml:"service"`
	Methods   []string        `json:"methods" yaml:"methods"`
	Paths     []string        `json:"paths,omitempty" yaml:"paths,omitempty"`
	Tables    []string        `json:"tables,omitempty" yaml:"tables,omitempty"`
	Resources json.RawMessage `json:"resources,omitempty" yaml:"-"`
}

type hostServiceStorageResourcesWire struct {
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

type hostServiceDataResourcesWire struct {
	Tables []string `json:"tables,omitempty" yaml:"tables,omitempty"`
}

type hostServiceNetworkResourceWire struct {
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
	Ref string `json:"ref,omitempty" yaml:"ref,omitempty"`
}

// MarshalJSON serializes host service declarations using the manifest-facing
// `resources` envelope. Storage services emit `resources.paths`, data services
// emit `resources.tables`, while network continues to emit URL resources.
func (spec HostServiceSpec) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"service": spec.Service,
		"methods": spec.Methods,
	}
	if len(spec.Paths) > 0 {
		payload["resources"] = &hostServiceStorageResourcesWire{Paths: spec.Paths}
	} else if len(spec.Tables) > 0 {
		payload["resources"] = &hostServiceDataResourcesWire{Tables: spec.Tables}
	} else if spec.Service == HostServiceNetwork && len(spec.Resources) > 0 {
		payload["resources"] = marshalNetworkResources(spec.Resources)
	} else if len(spec.Resources) > 0 {
		payload["resources"] = spec.Resources
	}
	return json.Marshal(payload)
}

// UnmarshalJSON restores one host service declaration from either the current
// manifest shape or a lenient legacy top-level `tables` form.
func (spec *HostServiceSpec) UnmarshalJSON(data []byte) error {
	var wire hostServiceSpecWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	spec.Service = wire.Service
	spec.Methods = append([]string(nil), wire.Methods...)
	spec.Paths = append([]string(nil), wire.Paths...)
	spec.Tables = append([]string(nil), wire.Tables...)
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
		// Mapping-shaped resources are service-specific: storage maps to
		// `resources.paths`, while data maps to `resources.tables`.
		if normalizeHostServiceName(spec.Service) == HostServiceStorage {
			var storageResources hostServiceStorageResourcesWire
			if err := json.Unmarshal(trimmed, &storageResources); err != nil {
				return err
			}
			spec.Paths = append([]string(nil), storageResources.Paths...)
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
	if len(spec.Paths) > 0 {
		payload["resources"] = &hostServiceStorageResourcesWire{Paths: spec.Paths}
	} else if len(spec.Tables) > 0 {
		payload["resources"] = &hostServiceDataResourcesWire{Tables: spec.Tables}
	} else if spec.Service == HostServiceNetwork && len(spec.Resources) > 0 {
		payload["resources"] = marshalNetworkResources(spec.Resources)
	} else if len(spec.Resources) > 0 {
		payload["resources"] = spec.Resources
	}
	return payload, nil
}

// UnmarshalYAML restores one host service declaration from plugin.yaml using
// the unified `resources` envelope, with lenient fallback for legacy `tables`.
func (spec *HostServiceSpec) UnmarshalYAML(node *yaml.Node) error {
	type hostServiceSpecYAMLWire struct {
		Service   string    `yaml:"service"`
		Methods   []string  `yaml:"methods"`
		Paths     []string  `yaml:"paths,omitempty"`
		Tables    []string  `yaml:"tables,omitempty"`
		Resources yaml.Node `yaml:"resources,omitempty"`
	}

	var wire hostServiceSpecYAMLWire
	if err := node.Decode(&wire); err != nil {
		return err
	}

	spec.Service = wire.Service
	spec.Methods = append([]string(nil), wire.Methods...)
	spec.Paths = append([]string(nil), wire.Paths...)
	spec.Tables = append([]string(nil), wire.Tables...)
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
		// Mapping-shaped resources are service-specific: storage maps to
		// `resources.paths`, while data maps to `resources.tables`.
		if normalizeHostServiceName(spec.Service) == HostServiceStorage {
			var storageResources hostServiceStorageResourcesWire
			if err := wire.Resources.Decode(&storageResources); err != nil {
				return err
			}
			spec.Paths = append([]string(nil), storageResources.Paths...)
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

func unmarshalJSONResourcesByService(service string, payload []byte) ([]*HostServiceResourceSpec, error) {
	if normalizeHostServiceName(service) == HostServiceNetwork {
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

func unmarshalYAMLResourcesByService(service string, node *yaml.Node) ([]*HostServiceResourceSpec, error) {
	if normalizeHostServiceName(service) == HostServiceNetwork {
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
		if strings.TrimSpace(target) == "" {
			target = resource.Ref
		}
		items = append(items, &HostServiceResourceSpec{
			Ref: target,
		})
	}
	return items
}

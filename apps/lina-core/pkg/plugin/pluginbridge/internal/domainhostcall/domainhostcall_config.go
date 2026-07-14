// This file implements plugin-scoped config, host config, and manifest guest
// clients through injected raw host-service transport.

package domainhostcall

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/hostconfigcap"
	"lina-core/pkg/plugin/capability/manifestcap"
	"lina-core/pkg/plugin/capability/plugincap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// pluginConfigService adapts plugins.config.get host calls to plugincap.ConfigService.
type pluginConfigService struct{ baseService }

// hostConfigClient adapts hostConfig.get transport calls to simple helper methods.
type hostConfigClient struct{ baseService }

// hostConfigCapabilityService adapts a hostConfigClient to hostconfigcap.Service.
type hostConfigCapabilityService struct {
	client *hostConfigClient
}

// hostConfigSysConfigService adapts governed sys_config methods to host services.
type hostConfigSysConfigService struct {
	client *hostConfigClient
}

// manifestClient adapts manifest.get transport calls to simple helper methods.
type manifestClient struct{ baseService }

// manifestCapabilityService adapts a manifestClient to manifestcap.Service.
type manifestCapabilityService struct {
	client *manifestClient
}

// HostConfig creates the host configuration guest client.
func HostConfig(invoker HostServiceInvoker) *hostConfigClient {
	return &hostConfigClient{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// HostConfigCapability creates the read-only host configuration capability client.
func HostConfigCapability(invoker HostServiceInvoker) hostconfigcap.Service {
	return &hostConfigCapabilityService{client: HostConfig(invoker)}
}

// PluginConfig creates the plugin-scoped config capability guest client.
func PluginConfig(invoker HostServiceInvoker) plugincap.ConfigService {
	return &pluginConfigService{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// Manifest creates the manifest-resource guest client.
func Manifest(invoker HostServiceInvoker) *manifestClient {
	return &manifestClient{baseService: newBaseServiceWithHostService(nil, invoker)}
}

// ManifestCapability creates the manifest-resource capability client.
func ManifestCapability(invoker HostServiceInvoker) manifestcap.Service {
	return &manifestCapabilityService{client: Manifest(invoker)}
}

// Get reads one authorized host config value as JSON.
func (s *hostConfigClient) Get(key string) (string, bool, error) {
	return s.configValue(protocol.HostServiceHostConfig, protocol.HostServiceMethodHostConfigGet, key, key)
}

// String reads one authorized host config value as a string.
func (s *hostConfigClient) String(key string) (string, bool, error) {
	value, found, err := s.Get(key)
	if err != nil || !found {
		return "", found, err
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	return strings.Trim(value, `"`), true, nil
}

// Bool reads one authorized host config value as a bool.
func (s *hostConfigClient) Bool(key string) (bool, bool, error) {
	value, found, err := s.Get(key)
	if err != nil || !found {
		return false, found, err
	}
	var decoded bool
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, true, gerror.Wrapf(err, "parse host config %s bool failed", key)
	}
	return parsed, true, nil
}

// Int reads one authorized host config value as an int.
func (s *hostConfigClient) Int(key string) (int, bool, error) {
	value, found, err := s.Get(key)
	if err != nil || !found {
		return 0, found, err
	}
	var decoded int
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, true, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse host config %s int failed", key)
	}
	return parsed, true, nil
}

// Duration reads one authorized host config value as a duration.
func (s *hostConfigClient) Duration(key string) (time.Duration, bool, error) {
	value, found, err := s.Get(key)
	if err != nil || !found {
		return 0, found, err
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		value = decoded
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, true, gerror.Wrapf(err, "parse host config %s duration failed", key)
	}
	return parsed, true, nil
}

// Get returns the raw host config value for the requested key.
func (s *hostConfigCapabilityService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value, found, err := s.client.Get(key)
	if err != nil {
		return nil, err
	}
	if !found {
		if defaultValue != nil {
			return gvar.New(defaultValue), nil
		}
		return nil, nil
	}
	result := gvarFromJSONValue(value)
	if result == nil || result.IsNil() {
		if defaultValue != nil {
			return gvar.New(defaultValue), nil
		}
		return nil, nil
	}
	return result, nil
}

// Exists reports whether a host config key is available.
func (s *hostConfigCapabilityService) Exists(_ context.Context, key string) (bool, error) {
	_, found, err := s.client.Get(key)
	return found, err
}

// StringValue reads a host config string value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) String(_ context.Context, key string, defaultValue string) (string, error) {
	value, found, err := s.client.String(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}

// BoolValue reads a host config bool value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	value, found, err := s.client.Bool(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}

// IntValue reads a host config int value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) Int(_ context.Context, key string, defaultValue int) (int, error) {
	value, found, err := s.client.Int(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}

// DurationValue reads a host config duration value or returns defaultValue when absent.
func (s *hostConfigCapabilityService) Duration(_ context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, found, err := s.client.Duration(key)
	if err != nil || !found {
		return defaultValue, err
	}
	return value, nil
}

// SysConfig returns the dynamic single-key sys_config subresource adapter.
func (s *hostConfigCapabilityService) SysConfig() hostconfigcap.SysConfigService {
	return hostConfigSysConfigService{client: s.client}
}

// Get reads one governed sys_config projection.
func (s hostConfigSysConfigService) Get(_ context.Context, key hostconfigcap.SysConfigKey) (*hostconfigcap.SysConfigInfo, error) {
	var out *hostconfigcap.SysConfigInfo
	err := s.client.callHostServiceJSONRequest(
		protocol.HostServiceHostConfig,
		protocol.HostServiceMethodHostConfigSysConfigGet,
		string(key),
		"",
		hostConfigSysConfigKeyRequest{Key: string(key)},
		&out,
	)
	return out, err
}

// BatchGet reads governed sys_config projections.
func (s hostConfigSysConfigService) BatchGet(ctx context.Context, keys []hostconfigcap.SysConfigKey) (*capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey], error) {
	out := &capmodel.BatchResult[*hostconfigcap.SysConfigInfo, hostconfigcap.SysConfigKey]{
		Items:      map[hostconfigcap.SysConfigKey]*hostconfigcap.SysConfigInfo{},
		MissingIDs: []hostconfigcap.SysConfigKey{},
	}
	for _, key := range keys {
		item, err := s.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		if item == nil {
			out.MissingIDs = append(out.MissingIDs, key)
			continue
		}
		out.Items[key] = item
	}
	return out, nil
}

// List is not published as a dynamic hostconfig host-service method because
// list authorization cannot be represented by one `resources.keys` resourceRef.
func (hostConfigSysConfigService) List(_ context.Context, _ hostconfigcap.ListSysConfigInput) (*capmodel.PageResult[*hostconfigcap.SysConfigInfo], error) {
	return nil, unsupportedDynamicMethodError("hostconfig.sys_config.list")
}

// SetValue writes one governed sys_config value with optional management flag.
func (s hostConfigSysConfigService) SetValue(
	_ context.Context,
	key hostconfigcap.SysConfigKey,
	value string,
	options *hostconfigcap.SetSysConfigValueOptions,
) error {
	req := hostConfigSysConfigSetValueRequest{
		Key:   string(key),
		Value: value,
	}
	if options != nil {
		req.SystemManageable = options.SystemManageable
	}
	return s.client.callHostServiceJSONRequest(
		protocol.HostServiceHostConfig,
		protocol.HostServiceMethodHostConfigSysConfigSetValue,
		string(key),
		"",
		req,
		nil,
	)
}

// BatchSetValue writes multiple governed sys_config values. Dynamic host
// services authorize one key per call, so the guest adapter fans out to
// SetValue while preserving item order and shared options.
func (s hostConfigSysConfigService) BatchSetValue(
	ctx context.Context,
	items []hostconfigcap.SetSysConfigValueItem,
	options *hostconfigcap.SetSysConfigValueOptions,
) error {
	for _, item := range items {
		if err := s.SetValue(ctx, item.Key, item.Value, options); err != nil {
			return err
		}
	}
	return nil
}

// Reset resets one governed sys_config value.
func (s hostConfigSysConfigService) Reset(_ context.Context, key hostconfigcap.SysConfigKey) error {
	return s.client.callHostServiceJSONRequest(
		protocol.HostServiceHostConfig,
		protocol.HostServiceMethodHostConfigSysConfigReset,
		string(key),
		"",
		hostConfigSysConfigKeyRequest{Key: string(key)},
		nil,
	)
}

// EnsureVisible rejects when any sys_config key is outside caller scope.
func (s hostConfigSysConfigService) EnsureVisible(ctx context.Context, keys []hostconfigcap.SysConfigKey) error {
	for _, key := range keys {
		if _, err := s.Get(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

type hostConfigSysConfigKeyRequest struct {
	Key string `json:"key"`
}

type hostConfigSysConfigSetValueRequest struct {
	Key              string `json:"key"`
	Value            string `json:"value"`
	SystemManageable *bool  `json:"systemManageable,omitempty"`
}

// Get returns the raw plugin configuration value for the given key.
func (s *pluginConfigService) Get(_ context.Context, key string, defaultValue any) (*gvar.Var, error) {
	value, found, err := s.configValue(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsConfigGet, "", key)
	if err != nil {
		return nil, err
	}
	if !found {
		if defaultValue != nil {
			return gvar.New(defaultValue), nil
		}
		return nil, nil
	}
	return gvarFromJSONValue(value), nil
}

// Exists reports whether the given configuration key exists.
func (s *pluginConfigService) Exists(_ context.Context, key string) (bool, error) {
	_, found, err := s.configValue(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsConfigGet, "", key)
	return found, err
}

// Scan scans the configuration section into target.
func (s *pluginConfigService) Scan(ctx context.Context, key string, target any) error {
	if target == nil {
		return gerror.New("plugin config scan target cannot be nil")
	}
	value, err := s.Get(ctx, key, nil)
	if err != nil || value == nil || value.IsNil() {
		return err
	}
	return value.Scan(target)
}

// String reads a string value or returns defaultValue when the key is absent or blank.
func (s *pluginConfigService) String(_ context.Context, key string, defaultValue string) (string, error) {
	value, found, err := s.configValue(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsConfigGet, "", key)
	if err != nil {
		return "", err
	}
	if !found || strings.TrimSpace(value) == "" {
		return defaultValue, nil
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, nil
	}
	return strings.Trim(value, `"`), nil
}

// Bool reads a bool value or returns defaultValue when the key is absent.
func (s *pluginConfigService) Bool(_ context.Context, key string, defaultValue bool) (bool, error) {
	value, found, err := s.configValue(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsConfigGet, "", key)
	if err != nil {
		return false, err
	}
	if !found {
		return defaultValue, nil
	}
	var decoded bool
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, nil
	}
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false, gerror.Wrapf(err, "parse config %s bool failed", key)
	}
	return parsed, nil
}

// Int reads an int value or returns defaultValue when the key is absent.
func (s *pluginConfigService) Int(_ context.Context, key string, defaultValue int) (int, error) {
	value, found, err := s.configValue(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsConfigGet, "", key)
	if err != nil {
		return 0, err
	}
	if !found {
		return defaultValue, nil
	}
	var decoded int
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded, nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, gerror.Wrapf(err, "parse config %s int failed", key)
	}
	return parsed, nil
}

// Duration reads a time.Duration value or returns defaultValue when the key is absent or blank.
func (s *pluginConfigService) Duration(_ context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, found, err := s.configValue(protocol.HostServicePlugins, protocol.HostServiceMethodPluginsConfigGet, "", key)
	if err != nil {
		return 0, err
	}
	if !found {
		return defaultValue, nil
	}
	var decoded string
	if err = json.Unmarshal([]byte(value), &decoded); err == nil {
		value = decoded
	}
	parsed, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, gerror.Wrapf(err, "parse config %s duration failed", key)
	}
	return parsed, nil
}

// Get reads one manifest resource as bytes.
func (s *manifestClient) Get(path string) ([]byte, bool, error) {
	request := &protocol.HostServiceManifestGetRequest{Path: path}
	payload, err := s.callHostService(
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

// GetMany reads authorized manifest resources by explicit path.
func (s *manifestClient) GetMany(paths []string) (*manifestcap.GetManyOutput, error) {
	response := &manifestcap.GetManyOutput{}
	err := s.callHostServiceJSONRequest(
		protocol.HostServiceManifest,
		protocol.HostServiceMethodManifestGetMany,
		manifestBatchResourceRef(paths),
		"",
		manifestcap.GetManyInput{Paths: paths},
		response,
	)
	if err != nil {
		return nil, err
	}
	if response.Resources == nil {
		response.Resources = []*manifestcap.ResourceContent{}
	}
	return response, nil
}

// List lists authorized manifest resource metadata under one prefix.
func (s *manifestClient) List(prefix string, limit int) (*manifestcap.ListOutput, error) {
	response := &manifestcap.ListOutput{}
	err := s.callHostServiceJSONRequest(
		protocol.HostServiceManifest,
		protocol.HostServiceMethodManifestList,
		manifestListResourceRef(prefix),
		"",
		manifestcap.ListInput{Prefix: prefix, Limit: limit},
		response,
	)
	if err != nil {
		return nil, err
	}
	if response.Resources == nil {
		response.Resources = []*manifestcap.Resource{}
	}
	return response, nil
}

// GetText reads one manifest resource as UTF-8 text.
func (s *manifestClient) GetText(path string) (string, bool, error) {
	body, found, err := s.Get(path)
	if err != nil || !found {
		return "", found, err
	}
	return string(body), true, nil
}

// Scan decodes a YAML manifest resource or nested key into target.
func (s *manifestClient) Scan(path string, key string, target any) (bool, error) {
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

// GetResource returns one raw resource under the current plugin manifest root.
func (s *manifestCapabilityService) Get(_ context.Context, path string) ([]byte, error) {
	content, _, err := s.client.Get(path)
	return content, err
}

// GetMany returns raw resources for explicit authorized manifest-relative paths.
func (s *manifestCapabilityService) GetMany(_ context.Context, input manifestcap.GetManyInput) (*manifestcap.GetManyOutput, error) {
	return s.client.GetMany(input.Paths)
}

// List returns authorized manifest resource metadata under one bounded prefix.
func (s *manifestCapabilityService) List(_ context.Context, input manifestcap.ListInput) (*manifestcap.ListOutput, error) {
	return s.client.List(input.Prefix, input.Limit)
}

// Exists reports whether one allowed manifest resource exists.
func (s *manifestCapabilityService) Exists(_ context.Context, path string) (bool, error) {
	_, found, err := s.client.Get(path)
	return found, err
}

// ScanResource unmarshals the selected YAML resource, or the nested key inside it, into target.
func (s *manifestCapabilityService) Scan(_ context.Context, path string, key string, target any) error {
	_, err := s.client.Scan(path, key, target)
	return err
}

func (s baseService) configValue(service string, method string, resourceRef string, key string) (string, bool, error) {
	payload, err := s.callHostService(
		service,
		method,
		resourceRef,
		"",
		protocol.MarshalHostServiceConfigKeyRequest(&protocol.HostServiceConfigKeyRequest{Key: key}),
	)
	if err != nil {
		return "", false, err
	}
	if len(payload) == 0 {
		return "", false, nil
	}
	response, err := protocol.UnmarshalHostServiceConfigValueResponse(payload)
	if err != nil {
		return "", false, err
	}
	return response.Value, response.Found, nil
}

func gvarFromJSONValue(value string) *gvar.Var {
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err == nil {
		return gvar.New(decoded)
	}
	return gvar.New(value)
}

func manifestBatchResourceRef(paths []string) string {
	return batchResourceRef(paths)
}

func manifestListResourceRef(prefix string) string {
	trimmed := strings.Trim(strings.ReplaceAll(strings.TrimSpace(prefix), "\\", "/"), "/")
	if trimmed == "" {
		return ".manifest-list-probe"
	}
	return trimmed + "/.manifest-list-probe"
}

func batchResourceRef(paths []string) string {
	first := ""
	for _, rawPath := range paths {
		trimmed := strings.ReplaceAll(strings.TrimSpace(rawPath), "\\", "/")
		if trimmed == "" {
			continue
		}
		if first == "" {
			first = trimmed
			continue
		}
		return commonDirectoryResourceRef(paths)
	}
	return first
}

func commonDirectoryResourceRef(paths []string) string {
	prefix := ""
	for _, rawPath := range paths {
		trimmed := strings.Trim(strings.ReplaceAll(strings.TrimSpace(rawPath), "\\", "/"), "/")
		if trimmed == "" {
			continue
		}
		dir := pathDirectoryPrefix(trimmed)
		if prefix == "" {
			prefix = dir
			continue
		}
		for prefix != "" && !strings.HasPrefix(trimmed, prefix) {
			prefix = parentDirectoryPrefix(strings.TrimSuffix(prefix, "/"))
		}
	}
	return prefix
}

func pathDirectoryPrefix(path string) string {
	if strings.HasSuffix(path, "/") {
		return path
	}
	index := strings.LastIndex(path, "/")
	if index < 0 {
		return ""
	}
	return path[:index+1]
}

func parentDirectoryPrefix(path string) string {
	index := strings.LastIndex(path, "/")
	if index < 0 {
		return ""
	}
	return path[:index+1]
}

var (
	_ plugincap.ConfigService = (*pluginConfigService)(nil)
	_ hostconfigcap.Service   = (*hostConfigCapabilityService)(nil)
	_ manifestcap.Service     = (*manifestCapabilityService)(nil)
)

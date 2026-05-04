// This file tests host service declaration validation, normalization, and
// capability authorization rules for the structured core host services.

package pluginbridge

import (
	"encoding/json"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestValidateHostServiceSpecsNormalizesStoragePaths verifies validation trims,
// sorts, and normalizes storage-style host service declarations.
func TestValidateHostServiceSpecsNormalizesStoragePaths(t *testing.T) {
	specs := []*HostServiceSpec{
		{
			Service: " STORAGE ",
			Methods: []string{"Get", "put"},
			Paths:   []string{" reports/ ", "exports/daily.json"},
		},
		{
			Service: "runtime",
			Methods: []string{"info.uuid", "log.write"},
		},
	}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected host service specs to validate, got error: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("expected 2 normalized specs, got %d", len(specs))
	}
	if specs[0].Service != HostServiceRuntime {
		t.Fatalf("expected runtime spec to sort first, got %s", specs[0].Service)
	}
	if specs[1].Service != HostServiceStorage {
		t.Fatalf("expected storage spec to be normalized, got %s", specs[1].Service)
	}
	if len(specs[1].Methods) != 2 || specs[1].Methods[0] != HostServiceMethodStorageGet || specs[1].Methods[1] != HostServiceMethodStoragePut {
		t.Fatalf("expected normalized storage methods [get put], got %#v", specs[1].Methods)
	}
	if len(specs[1].Paths) != 2 || specs[1].Paths[0] != "exports/daily.json" || specs[1].Paths[1] != "reports/" {
		t.Fatalf("expected normalized storage paths, got %#v", specs[1].Paths)
	}
}

// TestValidateHostServiceSpecsRejectsRuntimeResources verifies runtime service
// declarations cannot carry resource entries.
func TestValidateHostServiceSpecsRejectsRuntimeResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceRuntime,
		Methods: []string{HostServiceMethodRuntimeInfoUUID},
		Resources: []*HostServiceResourceSpec{{
			Ref: "unexpected",
		}},
	}})
	if err == nil {
		t.Fatal("expected runtime host service resources to be rejected")
	}
}

// TestNormalizeHostServiceSpecsReturnsError verifies dynamic declarations use
// explicit errors instead of panicking on invalid host service input.
func TestNormalizeHostServiceSpecsReturnsError(t *testing.T) {
	normalized, err := NormalizeHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet},
	}})
	if err == nil {
		t.Fatal("expected invalid host service declaration to return an error")
	}
	if len(normalized) != 0 {
		t.Fatalf("expected invalid host service declaration to return no normalized entries, got %#v", normalized)
	}
}

// TestMustNormalizeHostServiceSpecsPanics verifies the Must helper remains
// fail-fast for compile-time-only declarations.
func TestMustNormalizeHostServiceSpecsPanics(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected MustNormalizeHostServiceSpecs to panic for invalid declarations")
		}
	}()

	MustNormalizeHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet},
	}})
}

// TestValidateHostServiceSpecsAcceptsCronWithoutResources verifies cron
// registration service declarations use the same resource-less shape as
// runtime host services.
func TestValidateHostServiceSpecsAcceptsCronWithoutResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceCron,
		Methods: []string{HostServiceMethodCronRegister},
	}})
	if err != nil {
		t.Fatalf("expected cron host service without resources to validate, got %v", err)
	}
}

// TestValidateHostServiceSpecsAcceptsConfigWithoutResources verifies config
// read access is authorized at the service/method level for full config reads.
func TestValidateHostServiceSpecsAcceptsConfigWithoutResources(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServiceConfig,
		Methods: []string{
			HostServiceMethodConfigGet,
			HostServiceMethodConfigExists,
			HostServiceMethodConfigDuration,
		},
	}}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected config host service without resources to validate, got %v", err)
	}

	capabilities := CapabilityMapFromHostServices(specs)
	if _, ok := capabilities[CapabilityConfig]; !ok {
		t.Fatalf("expected config declaration to derive %s capability", CapabilityConfig)
	}
}

// TestValidateHostServiceSpecsDefaultsConfigMethods verifies omitted config
// methods grant the complete read-only config permission surface.
func TestValidateHostServiceSpecsDefaultsConfigMethods(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServiceConfig,
	}}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected config host service without explicit methods to validate, got %v", err)
	}
	expectedMethods := []string{
		HostServiceMethodConfigBool,
		HostServiceMethodConfigDuration,
		HostServiceMethodConfigExists,
		HostServiceMethodConfigGet,
		HostServiceMethodConfigInt,
		HostServiceMethodConfigString,
	}
	if !reflect.DeepEqual(specs[0].Methods, expectedMethods) {
		t.Fatalf("expected omitted config methods to default to all read-only methods, got %#v", specs[0].Methods)
	}

	capabilities := CapabilitiesFromHostServices([]*HostServiceSpec{{Service: HostServiceConfig}})
	if len(capabilities) != 1 || capabilities[0] != CapabilityConfig {
		t.Fatalf("expected omitted config methods to derive config capability, got %#v", capabilities)
	}
}

// TestValidateHostServiceSpecsRejectsConfigUnsupportedMethods verifies config
// declarations only accept known read-only config methods.
func TestValidateHostServiceSpecsRejectsConfigUnsupportedMethods(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceConfig,
		Methods: []string{HostServiceMethodConfigGet, "set"},
	}})
	if err == nil {
		t.Fatal("expected unsupported config host service methods to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsConfigResources verifies config service
// declarations do not accept key-level resource restrictions in this model.
func TestValidateHostServiceSpecsRejectsConfigResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceConfig,
		Methods: []string{HostServiceMethodConfigGet},
		Resources: []*HostServiceResourceSpec{{
			Ref: "monitor.*",
		}},
	}})
	if err == nil {
		t.Fatal("expected config host service resources to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsCronResources verifies cron registration
// declarations do not accept resource refs.
func TestValidateHostServiceSpecsRejectsCronResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceCron,
		Methods: []string{HostServiceMethodCronRegister},
		Resources: []*HostServiceResourceSpec{{
			Ref: "unexpected",
		}},
	}})
	if err == nil {
		t.Fatal("expected cron host service resources to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsStorageLegacyResources verifies storage
// services no longer accept legacy `resources.ref` declarations.
func TestValidateHostServiceSpecsRejectsStorageLegacyResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet},
		Resources: []*HostServiceResourceSpec{{
			Ref: "plugin-private-files",
		}},
	}})
	if err == nil {
		t.Fatal("expected storage legacy resource refs to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsCoreServiceWithoutResource verifies
// resource-bearing services fail validation when required scopes are absent.
func TestValidateHostServiceSpecsRejectsCoreServiceWithoutResource(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet},
	}})
	if err == nil {
		t.Fatal("expected storage host service without paths to be rejected")
	}
}

// TestValidateHostServiceSpecsAcceptsDataTables verifies data service
// declarations normalize and accept authorized table lists.
func TestValidateHostServiceSpecsAcceptsDataTables(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList, HostServiceMethodDataUpdate},
		Tables:  []string{" sys_plugin_node_state ", "sys_user"},
	}})
	if err != nil {
		t.Fatalf("expected data host service tables to validate, got %v", err)
	}
}

// TestValidateHostServiceSpecsRejectsDataResources verifies data services must
// use table authorization instead of generic resources.
func TestValidateHostServiceSpecsRejectsDataResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList},
		Resources: []*HostServiceResourceSpec{{
			Ref: "unexpected",
		}},
	}})
	if err == nil {
		t.Fatal("expected data host service resources to be rejected")
	}
}

// TestValidateHostServiceSpecsAcceptsNetworkURLPatterns verifies network
// services accept normalized URL-pattern resources.
func TestValidateHostServiceSpecsAcceptsNetworkURLPatterns(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceNetwork,
		Methods: []string{HostServiceMethodNetworkRequest},
		Resources: []*HostServiceResourceSpec{{
			Ref: " https://*.example.com/api ",
		}},
	}})
	if err != nil {
		t.Fatalf("expected network url patterns to validate, got %v", err)
	}
}

// TestValidateHostServiceSpecsAcceptsCacheLockNotifyResources verifies generic
// resource-based services normalize their declared refs.
func TestValidateHostServiceSpecsAcceptsCacheLockNotifyResources(t *testing.T) {
	specs := []*HostServiceSpec{
		{
			Service: HostServiceCache,
			Methods: []string{HostServiceMethodCacheGet, HostServiceMethodCacheSet},
			Resources: []*HostServiceResourceSpec{
				{Ref: " order-sync-cache "},
			},
		},
		{
			Service: HostServiceLock,
			Methods: []string{HostServiceMethodLockAcquire, HostServiceMethodLockRelease},
			Resources: []*HostServiceResourceSpec{
				{Ref: " order-sync-lock "},
			},
		},
		{
			Service: HostServiceNotify,
			Methods: []string{HostServiceMethodNotifySend},
			Resources: []*HostServiceResourceSpec{
				{Ref: " inbox "},
			},
		},
	}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected cache/lock/notify host service specs to validate, got %v", err)
	}
	if specs[0].Resources[0].Ref != "order-sync-cache" {
		t.Fatalf("expected normalized cache resource ref, got %#v", specs[0].Resources[0])
	}
	if specs[1].Resources[0].Ref != "order-sync-lock" {
		t.Fatalf("expected normalized lock resource ref, got %#v", specs[1].Resources[0])
	}
	if specs[2].Resources[0].Ref != "inbox" {
		t.Fatalf("expected normalized notify resource ref, got %#v", specs[2].Resources[0])
	}
}

// TestValidateHostServiceSpecsRejectsLegacyNetworkGovernanceFields verifies
// network resources reject deprecated per-resource governance knobs.
func TestValidateHostServiceSpecsRejectsLegacyNetworkGovernanceFields(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceNetwork,
		Methods: []string{HostServiceMethodNetworkRequest},
		Resources: []*HostServiceResourceSpec{{
			Ref:          "https://api.example.com",
			AllowMethods: []string{"GET"},
		}},
	}})
	if err == nil {
		t.Fatal("expected legacy network governance fields to be rejected")
	}
}

// TestHostServiceSpecJSONUsesResourcePathsForStorage verifies storage services
// marshal and unmarshal through `resources.paths`.
func TestHostServiceSpecJSONUsesResourcePathsForStorage(t *testing.T) {
	spec := &HostServiceSpec{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet, HostServiceMethodStoragePut},
		Paths:   []string{"reports/", "exports/daily.json"},
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("expected storage host service json marshal to succeed, got %v", err)
	}

	var encoded map[string]interface{}
	if err = json.Unmarshal(payload, &encoded); err != nil {
		t.Fatalf("expected marshaled storage host service json to decode, got %v", err)
	}
	resources, ok := encoded["resources"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected storage host service json resources object, got %#v", encoded["resources"])
	}
	paths, ok := resources["paths"].([]interface{})
	if !ok || len(paths) != 2 {
		t.Fatalf("expected storage host service json resources.paths, got %#v", resources["paths"])
	}

	decoded := &HostServiceSpec{}
	if err = json.Unmarshal(payload, decoded); err != nil {
		t.Fatalf("expected storage host service json unmarshal to succeed, got %v", err)
	}
	if decoded.Service != HostServiceStorage || len(decoded.Paths) != 2 {
		t.Fatalf("unexpected decoded storage host service: %#v", decoded)
	}
	if len(decoded.Resources) != 0 {
		t.Fatalf("expected storage host service to decode without resource refs, got %#v", decoded.Resources)
	}
}

// TestHostServiceSpecJSONUsesResourceTablesForData verifies data services
// marshal and unmarshal through `resources.tables`.
func TestHostServiceSpecJSONUsesResourceTablesForData(t *testing.T) {
	spec := &HostServiceSpec{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList, HostServiceMethodDataGet},
		Tables:  []string{"sys_plugin_node_state"},
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("expected data host service json marshal to succeed, got %v", err)
	}

	var encoded map[string]interface{}
	if err = json.Unmarshal(payload, &encoded); err != nil {
		t.Fatalf("expected marshaled data host service json to decode, got %v", err)
	}
	resources, ok := encoded["resources"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data host service json resources object, got %#v", encoded["resources"])
	}
	tables, ok := resources["tables"].([]interface{})
	if !ok || len(tables) != 1 || tables[0] != "sys_plugin_node_state" {
		t.Fatalf("expected data host service json resources.tables, got %#v", resources["tables"])
	}

	decoded := &HostServiceSpec{}
	if err = json.Unmarshal(payload, decoded); err != nil {
		t.Fatalf("expected data host service json unmarshal to succeed, got %v", err)
	}
	if decoded.Service != HostServiceData || len(decoded.Tables) != 1 || decoded.Tables[0] != "sys_plugin_node_state" {
		t.Fatalf("unexpected decoded data host service: %#v", decoded)
	}
	if len(decoded.Resources) != 0 {
		t.Fatalf("expected data host service to decode without ref resources, got %#v", decoded.Resources)
	}
}

// TestHostServiceSpecYAMLUsesResourceTablesForData verifies YAML uses the same
// `resources.tables` shape for data service declarations.
func TestHostServiceSpecYAMLUsesResourceTablesForData(t *testing.T) {
	spec := &HostServiceSpec{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList, HostServiceMethodDataGet},
		Tables:  []string{"sys_plugin_node_state"},
	}

	payload, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("expected data host service yaml marshal to succeed, got %v", err)
	}

	var encoded map[string]interface{}
	if err = yaml.Unmarshal(payload, &encoded); err != nil {
		t.Fatalf("expected marshaled data host service yaml to decode, got %v", err)
	}
	resources, ok := encoded["resources"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data host service yaml resources object, got %#v", encoded["resources"])
	}
	tables, ok := resources["tables"].([]interface{})
	if !ok || len(tables) != 1 || tables[0] != "sys_plugin_node_state" {
		t.Fatalf("expected data host service yaml resources.tables, got %#v", resources["tables"])
	}

	decoded := &HostServiceSpec{}
	if err = yaml.Unmarshal(payload, decoded); err != nil {
		t.Fatalf("expected data host service yaml unmarshal to succeed, got %v", err)
	}
	if decoded.Service != HostServiceData || len(decoded.Tables) != 1 || decoded.Tables[0] != "sys_plugin_node_state" {
		t.Fatalf("unexpected decoded data host service: %#v", decoded)
	}
	if len(decoded.Resources) != 0 {
		t.Fatalf("expected data host service to decode without ref resources, got %#v", decoded.Resources)
	}
}

// TestHostServiceSpecYAMLUsesURLForNetworkResources verifies network services
// marshal and unmarshal YAML resource URLs through the shared resources array.
func TestHostServiceSpecYAMLUsesURLForNetworkResources(t *testing.T) {
	spec := &HostServiceSpec{
		Service: HostServiceNetwork,
		Methods: []string{HostServiceMethodNetworkRequest},
		Resources: []*HostServiceResourceSpec{{
			Ref: "https://*.example.com/api",
		}},
	}

	payload, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("expected network host service yaml marshal to succeed, got %v", err)
	}

	var encoded map[string]interface{}
	if err = yaml.Unmarshal(payload, &encoded); err != nil {
		t.Fatalf("expected marshaled network host service yaml to decode, got %v", err)
	}
	resources, ok := encoded["resources"].([]interface{})
	if !ok || len(resources) != 1 {
		t.Fatalf("expected network host service yaml resources array, got %#v", encoded["resources"])
	}
	item, ok := resources[0].(map[string]interface{})
	if !ok || item["url"] != "https://*.example.com/api" {
		t.Fatalf("expected network host service yaml url field, got %#v", resources[0])
	}

	decoded := &HostServiceSpec{}
	if err = yaml.Unmarshal(payload, decoded); err != nil {
		t.Fatalf("expected network host service yaml unmarshal to succeed, got %v", err)
	}
	if decoded.Service != HostServiceNetwork || len(decoded.Resources) != 1 || decoded.Resources[0].Ref != "https://*.example.com/api" {
		t.Fatalf("unexpected decoded network host service: %#v", decoded)
	}
}

// TestValidateHostServiceSpecsRejectsDuplicateMethods verifies normalized
// method duplicates are rejected during validation.
func TestValidateHostServiceSpecsRejectsDuplicateMethods(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet, "GET"},
		Paths:   []string{"reports/"},
	}})
	if err == nil {
		t.Fatal("expected duplicate storage methods to be rejected")
	}
}

// TestCapabilitiesFromHostServicesDerivesCapabilitySet verifies capability
// derivation expands service declarations into the expected sorted set.
func TestCapabilitiesFromHostServicesDerivesCapabilitySet(t *testing.T) {
	capabilities := CapabilitiesFromHostServices([]*HostServiceSpec{
		{
			Service: HostServiceRuntime,
			Methods: []string{HostServiceMethodRuntimeInfoUUID},
		},
		{
			Service: HostServiceData,
			Methods: []string{HostServiceMethodDataList, HostServiceMethodDataCreate},
			Tables:  []string{"sys_plugin_node_state"},
		},
	})
	if len(capabilities) != 3 {
		t.Fatalf("expected 3 derived capabilities, got %#v", capabilities)
	}
	if capabilities[0] != CapabilityDataMutate || capabilities[1] != CapabilityDataRead || capabilities[2] != CapabilityRuntime {
		t.Fatalf("unexpected derived capabilities ordering: %#v", capabilities)
	}
}

// TestCapabilitiesFromHostServicesDerivesLowPriorityCapabilitySet verifies
// derived capability ordering remains stable for cache, lock, and notify
// services.
func TestCapabilitiesFromHostServicesDerivesLowPriorityCapabilitySet(t *testing.T) {
	capabilities := CapabilitiesFromHostServices([]*HostServiceSpec{
		{
			Service: HostServiceCache,
			Methods: []string{HostServiceMethodCacheGet, HostServiceMethodCacheSet},
			Resources: []*HostServiceResourceSpec{
				{Ref: "order-sync-cache"},
			},
		},
		{
			Service: HostServiceLock,
			Methods: []string{HostServiceMethodLockAcquire},
			Resources: []*HostServiceResourceSpec{
				{Ref: "order-sync-lock"},
			},
		},
		{
			Service: HostServiceNotify,
			Methods: []string{HostServiceMethodNotifySend},
			Resources: []*HostServiceResourceSpec{
				{Ref: "inbox"},
			},
		},
	})

	if len(capabilities) != 3 {
		t.Fatalf("expected 3 derived capabilities, got %#v", capabilities)
	}
	if capabilities[0] != CapabilityCache || capabilities[1] != CapabilityLock || capabilities[2] != CapabilityNotify {
		t.Fatalf("unexpected derived low priority capabilities ordering: %#v", capabilities)
	}
}

// This file tests host service declaration validation, normalization, and
// capability authorization rules for the structured core host services.

package pluginbridge

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

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

func TestValidateHostServiceSpecsRejectsCoreServiceWithoutResource(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet},
	}})
	if err == nil {
		t.Fatal("expected storage host service without paths to be rejected")
	}
}

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

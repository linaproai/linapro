// This file tests host service declaration validation, normalization, and
// capability authorization rules for the structured core host services.

package hostservice

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

// TestValidateHostServiceSpecsRejectsCron verifies cron is not a plugin.yaml
// host service declaration.
func TestValidateHostServiceSpecsRejectsCron(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: "cron",
		Methods: []string{"register"},
	}})
	if err == nil {
		t.Fatal("expected cron host service declarations to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsI18n verifies runtime translation is not
// published as a dynamic-plugin host service.
func TestValidateHostServiceSpecsRejectsI18n(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: "i18n",
		Methods: []string{"locale.get"},
	}})
	if err == nil {
		t.Fatal("expected i18n host service declarations to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsMissingMethods verifies host service
// declarations must grant concrete methods explicitly.
func TestValidateHostServiceSpecsRejectsMissingMethods(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServiceHostConfig,
		Keys:    []string{"workspace.basePath"},
	}}

	if err := ValidateHostServiceSpecs(specs); err == nil {
		t.Fatal("expected host service without methods to be rejected")
	}
	if capabilities := CapabilityMapFromHostServices(specs); len(capabilities) != 0 {
		t.Fatalf("expected host service without methods to derive no capabilities, got %#v", capabilities)
	}
}

// TestValidateHostServiceSpecsAcceptsPluginsConfigWithoutResources verifies
// plugin config read access is authorized as a plugins domain method.
func TestValidateHostServiceSpecsAcceptsPluginsConfigWithoutResources(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServicePlugins,
		Methods: []string{HostServiceMethodPluginsConfigGet},
	}}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected plugins config method without resources to validate, got %v", err)
	}

	capabilities := CapabilityMapFromHostServices(specs)
	if _, ok := capabilities[CapabilityPlugins]; !ok {
		t.Fatalf("expected plugins config declaration to derive %s capability", CapabilityPlugins)
	}
}

// TestValidateHostServiceSpecsAcceptsOrgTenantWithoutResources verifies
// org and tenant host-service calls are authorized at the service/method level.
func TestValidateHostServiceSpecsAcceptsOrgTenantWithoutResources(t *testing.T) {
	specs := []*HostServiceSpec{
		{
			Service: HostServiceOrg,
			Methods: []string{
				HostServiceMethodOrgStatus,
				HostServiceMethodOrgListUserDeptAssignments,
				HostServiceMethodOrgGetUserDeptIDs,
			},
		},
		{
			Service: HostServiceTenant,
			Methods: []string{
				HostServiceMethodTenantStatus,
				HostServiceMethodTenantListUserTenants,
				HostServiceMethodTenantValidateSwitch,
			},
		},
	}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected org and tenant host services without resources to validate, got %v", err)
	}

	capabilities := CapabilityMapFromHostServices(specs)
	if _, ok := capabilities[CapabilityOrg]; !ok {
		t.Fatalf("expected org declaration to derive %s capability", CapabilityOrg)
	}
	if _, ok := capabilities[CapabilityTenant]; !ok {
		t.Fatalf("expected tenant declaration to derive %s capability", CapabilityTenant)
	}
}

// TestValidateHostServiceSpecsAcceptsDomainServicesWithoutResources verifies
// ordinary domain host services are authorized by service and method only.
func TestValidateHostServiceSpecsAcceptsDomainServicesWithoutResources(t *testing.T) {
	specs := []*HostServiceSpec{
		{Service: HostServiceAuthz, Methods: []string{HostServiceMethodAuthzBatchGetPermissions, HostServiceMethodAuthzBatchHasPermissions}},
		{Service: HostServiceDict, Methods: []string{HostServiceMethodDictResolveLabels, HostServiceMethodDictListValues, HostServiceMethodDictEnsureValuesVisible}},
		{Service: HostServiceFiles, Methods: []string{HostServiceMethodFilesBatchGet, HostServiceMethodFilesSearch}},
		{Service: HostServiceSessions, Methods: []string{HostServiceMethodSessionsCurrent, HostServiceMethodSessionsSearch, HostServiceMethodSessionsBatchGetUserOnlineStatus, HostServiceMethodSessionsEnsureVisible}},
		{Service: HostServiceJobs, Methods: []string{HostServiceMethodJobsBatchGet, HostServiceMethodJobsSearch, HostServiceMethodJobsEnsureVisible, HostServiceMethodJobsRegister}},
		{Service: HostServiceInfra, Methods: []string{HostServiceMethodInfraBatchGetStatus}},
		{Service: HostServiceAPIDoc, Methods: []string{HostServiceMethodAPIDocFindRouteTitleOperationKeys}},
		{Service: HostServiceBizCtx, Methods: []string{HostServiceMethodBizCtxCurrent}},
		{Service: HostServiceRoute, Methods: []string{HostServiceMethodRouteMetadataGet}},
		{Service: HostServiceNotifications, Methods: []string{HostServiceMethodNotificationsBatchGetMessages}},
		{Service: HostServicePlugins, Methods: []string{HostServiceMethodPluginsIsEnabled}},
	}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected ordinary domain host services without resources to validate, got %v", err)
	}

	capabilities := CapabilityMapFromHostServices(specs)
	for _, capability := range []string{
		CapabilityAuthz,
		CapabilityDict,
		CapabilityFiles,
		CapabilitySessions,
		CapabilityJobs,
		CapabilityInfra,
		CapabilityAPIDoc,
		CapabilityBizCtx,
		CapabilityRoute,
		CapabilityNotifications,
		CapabilityPlugins,
	} {
		if _, ok := capabilities[capability]; !ok {
			t.Fatalf("expected domain declaration to derive %s capability, got %#v", capability, capabilities)
		}
	}
}

// TestValidateHostServiceSpecsAcceptsPluginLifecycleMethods verifies dynamic
// plugin-domain lifecycle governance access is explicit method-level
// authorization under the plugins service.
func TestValidateHostServiceSpecsAcceptsPluginLifecycleMethods(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServicePlugins,
		Methods: []string{
			HostServiceMethodPluginsLifecycleEnsureTenantPluginDisable,
			HostServiceMethodPluginsLifecycleNotifyTenantPluginDisabled,
			HostServiceMethodPluginsLifecycleEnsureTenantDelete,
			HostServiceMethodPluginsLifecycleNotifyTenantDeleted,
		},
	}}
	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected plugin lifecycle host service methods to validate, got %v", err)
	}
	capabilities := CapabilityMapFromHostServices(specs)
	if _, ok := capabilities[CapabilityPlugins]; !ok {
		t.Fatalf("expected lifecycle declarations to derive %s capability, got %#v", CapabilityPlugins, capabilities)
	}
}

// TestValidateHostServiceSpecsRejectsStandaloneConfigService verifies the old
// standalone config service is no longer published.
func TestValidateHostServiceSpecsRejectsStandaloneConfigService(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: "config",
		Methods: []string{"get"},
	}})
	if err == nil {
		t.Fatal("expected standalone config service to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsPluginsConfigTypedMethods verifies config
// helper names are not authorization methods under plugins.
func TestValidateHostServiceSpecsRejectsPluginsConfigTypedMethods(t *testing.T) {
	for _, method := range []string{
		"config.exists",
		"config.string",
		"config.bool",
		"config.int",
		"config.duration",
	} {
		method := method
		t.Run(method, func(t *testing.T) {
			err := ValidateHostServiceSpecs([]*HostServiceSpec{{
				Service: HostServicePlugins,
				Methods: []string{HostServiceMethodPluginsConfigGet, method},
			}})
			if err == nil {
				t.Fatalf("expected plugins config typed helper method %s to be rejected", method)
			}
		})
	}
}

// TestValidateHostServiceSpecsRejectsPluginsConfigUnsupportedMethods verifies
// plugin config declarations only accept config.get.
func TestValidateHostServiceSpecsRejectsPluginsConfigUnsupportedMethods(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServicePlugins,
		Methods: []string{HostServiceMethodPluginsConfigGet, "config.set"},
	}})
	if err == nil {
		t.Fatal("expected unsupported plugins config methods to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsPluginsConfigResources verifies plugins
// config declarations do not accept resource restrictions in this model.
func TestValidateHostServiceSpecsRejectsPluginsConfigResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServicePlugins,
		Methods: []string{HostServiceMethodPluginsConfigGet},
		Resources: []*HostServiceResourceSpec{{
			Ref: "monitor.*",
		}},
	}})
	if err == nil {
		t.Fatal("expected plugins config resources to be rejected")
	}
}

// TestValidateHostServiceSpecsAcceptsHostConfigKeys verifies hostConfig
// declarations use resources.keys as their resource boundary.
func TestValidateHostServiceSpecsAcceptsHostConfigKeys(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServiceHostConfig,
		Methods: []string{HostServiceMethodHostConfigGet},
		Keys:    []string{" i18n.default ", "workspace.basePath"},
	}}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected hostConfig keys to validate, got %v", err)
	}
	if len(specs[0].Keys) != 2 || specs[0].Keys[0] != "i18n.default" || specs[0].Keys[1] != "workspace.basePath" {
		t.Fatalf("expected normalized hostConfig keys, got %#v", specs[0].Keys)
	}
	capabilities := CapabilityMapFromHostServices(specs)
	if _, ok := capabilities[CapabilityHostConfig]; !ok {
		t.Fatalf("expected hostConfig declaration to derive %s capability", CapabilityHostConfig)
	}
}

// TestValidateHostServiceSpecsRejectsLegacyHostRuntimeName verifies fresh
// declarations must use the current hostConfig service name. Legacy persisted
// release snapshots are migrated by the plugin catalog before validation.
func TestValidateHostServiceSpecsRejectsLegacyHostRuntimeName(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: "hostRuntime",
		Methods: []string{HostServiceMethodHostConfigGet},
		Keys:    []string{"workspace.basePath"},
	}}

	if err := ValidateHostServiceSpecs(specs); err == nil {
		t.Fatal("expected legacy hostRuntime declaration to be rejected for fresh host-service specs")
	}
}

// TestValidateHostServiceSpecsRejectsHostConfigWithoutKeys verifies key-scoped
// runtime declarations must explicitly request authorized host config keys.
func TestValidateHostServiceSpecsRejectsHostConfigWithoutKeys(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceHostConfig,
		Methods: []string{HostServiceMethodHostConfigGet},
	}})
	if err == nil {
		t.Fatal("expected hostConfig without resources.keys to be rejected")
	}
}

// TestValidateHostServiceSpecsAcceptsManifestPaths verifies manifest
// declarations use resources.paths as their resource boundary.
func TestValidateHostServiceSpecsAcceptsManifestPaths(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServiceManifest,
		Methods: []string{HostServiceMethodManifestGet},
		Paths: []string{
			" metadata.yaml ",
			"config/config.example.yaml",
			"i18n/zh-CN/plugin.json",
			"resources/*.yaml",
			"sql/001-schema.sql",
		},
	}}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected manifest paths to validate, got %v", err)
	}
	expectedPaths := []string{
		"config/config.example.yaml",
		"i18n/zh-CN/plugin.json",
		"metadata.yaml",
		"resources/*.yaml",
		"sql/001-schema.sql",
	}
	if !reflect.DeepEqual(specs[0].Paths, expectedPaths) {
		t.Fatalf("expected normalized manifest paths, got %#v", specs[0].Paths)
	}
	capabilities := CapabilityMapFromHostServices(specs)
	if _, ok := capabilities[CapabilityManifest]; !ok {
		t.Fatalf("expected manifest declaration to derive %s capability", CapabilityManifest)
	}
}

// TestValidateHostServiceSpecsRejectsUnsafeManifestPaths verifies manifest
// declarations reject paths that could escape the manifest root.
func TestValidateHostServiceSpecsRejectsUnsafeManifestPaths(t *testing.T) {
	for _, manifestPath := range []string{
		"",
		"../metadata.yaml",
		"/etc/passwd",
		`C:\secret.yaml`,
		"http://example.com/metadata.yaml",
		"manifest/metadata.yaml",
	} {
		manifestPath := manifestPath
		t.Run(manifestPath, func(t *testing.T) {
			err := ValidateHostServiceSpecs([]*HostServiceSpec{{
				Service: HostServiceManifest,
				Methods: []string{HostServiceMethodManifestGet},
				Paths:   []string{manifestPath},
			}})
			if err == nil {
				t.Fatalf("expected unsafe manifest path %q to be rejected", manifestPath)
			}
		})
	}
}

// TestValidateHostServiceSpecsRejectsCronResources verifies cron cannot be
// declared as a resource-scoped runtime host service.
func TestValidateHostServiceSpecsRejectsCronResources(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: "cron",
		Methods: []string{"register"},
		Resources: []*HostServiceResourceSpec{{
			Ref: "unexpected",
		}},
	}})
	if err == nil {
		t.Fatal("expected cron host service resources to be rejected")
	}
}

// TestValidateHostServiceSpecsRejectsStorageResourceRefs verifies storage
// services require path declarations instead of generic resource refs.
func TestValidateHostServiceSpecsRejectsStorageResourceRefs(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceStorage,
		Methods: []string{HostServiceMethodStorageGet},
		Resources: []*HostServiceResourceSpec{{
			Ref: "plugin-private-files",
		}},
	}})
	if err == nil {
		t.Fatal("expected storage resource refs to be rejected")
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

// TestValidateHostServiceSpecsRejectsDataTablesWithoutPlugin verifies data
// service declarations require plugin-aware table ownership validation.
func TestValidateHostServiceSpecsRejectsDataTablesWithoutPlugin(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList, HostServiceMethodDataUpdate},
		Tables:  []string{" sys_plugin_node_state ", "sys_user"},
	}})
	if err == nil {
		t.Fatal("expected data host service tables without plugin ID to be rejected")
	}
}

// TestValidateHostServiceSpecsForPluginAcceptsOwnedDataTables verifies
// production validation allows only data tables in the current plugin namespace.
func TestValidateHostServiceSpecsForPluginAcceptsOwnedDataTables(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList, HostServiceMethodDataUpdate},
		Tables: []string{
			" plugin_linapro_demo_dynamic_record ",
			"plugin_linapro_demo_dynamic",
		},
	}}

	if err := ValidateHostServiceSpecsForPlugin("linapro-demo-dynamic", specs); err != nil {
		t.Fatalf("expected plugin-owned data tables to validate, got %v", err)
	}
	if len(specs[0].Tables) != 2 || specs[0].Tables[0] != "plugin_linapro_demo_dynamic" || specs[0].Tables[1] != "plugin_linapro_demo_dynamic_record" {
		t.Fatalf("expected normalized plugin-owned tables, got %#v", specs[0].Tables)
	}
}

// TestValidateHostServiceSpecsForPluginRejectsCoreDataTables verifies dynamic
// data service declarations cannot authorize host sys_* core tables.
func TestValidateHostServiceSpecsForPluginRejectsCoreDataTables(t *testing.T) {
	err := ValidateHostServiceSpecsForPlugin("linapro-demo-dynamic", []*HostServiceSpec{{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList},
		Tables:  []string{"sys_plugin_node_state"},
	}})
	if err == nil {
		t.Fatal("expected host core data table to be rejected")
	}
}

// TestNormalizeHostServiceSpecsForPluginRejectsOtherPluginDataTables verifies
// official capability plugin tables stay inaccessible through another plugin's
// generic data host service authorization.
func TestNormalizeHostServiceSpecsForPluginRejectsOtherPluginDataTables(t *testing.T) {
	normalized, err := NormalizeHostServiceSpecsForPlugin("linapro-demo-dynamic", []*HostServiceSpec{{
		Service: HostServiceData,
		Methods: []string{HostServiceMethodDataList},
		Tables:  []string{"plugin_linapro_org_dept"},
	}})
	if err == nil {
		t.Fatal("expected other plugin data table to be rejected")
	}
	if len(normalized) != 0 {
		t.Fatalf("expected rejected declaration to return no normalized entries, got %#v", normalized)
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
			Service: HostServiceNotifications,
			Methods: []string{HostServiceMethodNotificationsSend},
			Resources: []*HostServiceResourceSpec{
				{Ref: " inbox "},
			},
		},
	}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected cache/lock/notifications host service specs to validate, got %v", err)
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

// TestValidateHostServiceSpecsAcceptsAITextMethodsWithoutResources verifies AI
// text host service declarations derive host:ai:text without resource refs.
func TestValidateHostServiceSpecsAcceptsAITextMethodsWithoutResources(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: " AI ",
		Methods: []string{" Text.Generate "},
	}}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected ai text host service specs to validate, got %v", err)
	}
	if len(specs[0].Resources) != 0 {
		t.Fatalf("expected ai host service to stay resource-less, got %#v", specs[0].Resources)
	}
	if specs[0].Service != HostServiceAI {
		t.Fatalf("expected normalized ai service, got %s", specs[0].Service)
	}
	if len(specs[0].Methods) != 1 || specs[0].Methods[0] != HostServiceMethodAITextGenerate {
		t.Fatalf("expected normalized ai text method, got %#v", specs[0].Methods)
	}
	capabilities := CapabilityMapFromHostServices(specs)
	if _, ok := capabilities[CapabilityAIText]; !ok {
		t.Fatalf("expected ai declaration to derive %s capability", CapabilityAIText)
	}
}

// TestValidateHostServiceSpecsAcceptsAIMultimodalMethodsWithoutResources
// verifies multimodal AI declarations derive method-specific capabilities
// without resource refs.
func TestValidateHostServiceSpecsAcceptsAIMultimodalMethodsWithoutResources(t *testing.T) {
	specs := []*HostServiceSpec{{
		Service: HostServiceAI,
		Methods: []string{
			HostServiceMethodAIImageGenerate,
			HostServiceMethodAIEmbeddingCreate,
			HostServiceMethodAIAudioTranscribe,
			HostServiceMethodAIVisionAnalyze,
			HostServiceMethodAIDocumentAnalyze,
			HostServiceMethodAISafetyModerate,
			HostServiceMethodAIVideoGenerate,
			HostServiceMethodAIVideoOperationGet,
		},
	}}

	if err := ValidateHostServiceSpecs(specs); err != nil {
		t.Fatalf("expected multimodal ai host service specs to validate, got %v", err)
	}
	capabilities := CapabilityMapFromHostServices(specs)
	for _, capability := range []string{
		CapabilityAIImage,
		CapabilityAIEmbedding,
		CapabilityAIAudio,
		CapabilityAIVision,
		CapabilityAIDocument,
		CapabilityAISafety,
		CapabilityAIVideo,
	} {
		if _, ok := capabilities[capability]; !ok {
			t.Fatalf("expected ai declaration to derive %s capability", capability)
		}
	}
}

// TestValidateHostServiceSpecsRejectsAIUnsupportedMethods verifies unsupported
// AI methods are rejected before runtime.
func TestValidateHostServiceSpecsRejectsAIUnsupportedMethods(t *testing.T) {
	for _, testCase := range []HostServiceSpec{
		{
			Service: HostServiceAI,
			Methods: []string{"computer.act"},
		},
		{
			Service: HostServiceAI,
			Methods: []string{"ui.operate"},
		},
	} {
		spec := testCase
		if err := ValidateHostServiceSpecs([]*HostServiceSpec{&spec}); err == nil {
			t.Fatalf("expected invalid ai host service method to be rejected: %#v", testCase)
		}
	}
}

// TestValidateHostServiceSpecsRejectsAIResources verifies AI declarations use
// method authorization only and reject every resource declaration shape.
func TestValidateHostServiceSpecsRejectsAIResources(t *testing.T) {
	for _, testCase := range []HostServiceSpec{
		{
			Service: HostServiceAI,
			Methods: []string{HostServiceMethodAITextGenerate},
			Resources: []*HostServiceResourceSpec{{
				Ref: "purpose:content.summary",
			}},
		},
		{
			Service: HostServiceAI,
			Methods: []string{HostServiceMethodAITextGenerate},
			Paths:   []string{"reports/"},
		},
		{
			Service: HostServiceAI,
			Methods: []string{HostServiceMethodAITextGenerate},
			Tables:  []string{"plugin_demo_reports"},
		},
		{
			Service: HostServiceAI,
			Methods: []string{HostServiceMethodAITextGenerate},
			Keys:    []string{"ai.default"},
		},
	} {
		spec := testCase
		if err := ValidateHostServiceSpecs([]*HostServiceSpec{&spec}); err == nil {
			t.Fatalf("expected ai host service resources to be rejected: %#v", testCase)
		}
	}
}

// TestValidateHostServiceSpecsRejectsNetworkResourceGovernanceFields verifies
// network resources only declare URL patterns.
func TestValidateHostServiceSpecsRejectsNetworkResourceGovernanceFields(t *testing.T) {
	err := ValidateHostServiceSpecs([]*HostServiceSpec{{
		Service: HostServiceNetwork,
		Methods: []string{HostServiceMethodNetworkRequest},
		Resources: []*HostServiceResourceSpec{{
			Ref:          "https://api.example.com",
			AllowMethods: []string{"GET"},
		}},
	}})
	if err == nil {
		t.Fatal("expected network resource governance fields to be rejected")
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

// TestHostServiceSpecJSONUsesResourceKeysForHostConfig verifies hostConfig
// services marshal and unmarshal through `resources.keys`.
func TestHostServiceSpecJSONUsesResourceKeysForHostConfig(t *testing.T) {
	spec := &HostServiceSpec{
		Service: HostServiceHostConfig,
		Methods: []string{HostServiceMethodHostConfigGet},
		Keys:    []string{"workspace.basePath", "i18n.default"},
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("expected hostConfig host service json marshal to succeed, got %v", err)
	}

	var encoded map[string]interface{}
	if err = json.Unmarshal(payload, &encoded); err != nil {
		t.Fatalf("expected marshaled hostConfig host service json to decode, got %v", err)
	}
	resources, ok := encoded["resources"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected hostConfig host service json resources object, got %#v", encoded["resources"])
	}
	keys, ok := resources["keys"].([]interface{})
	if !ok || len(keys) != 2 {
		t.Fatalf("expected hostConfig host service json resources.keys, got %#v", resources["keys"])
	}

	decoded := &HostServiceSpec{}
	if err = json.Unmarshal(payload, decoded); err != nil {
		t.Fatalf("expected hostConfig host service json unmarshal to succeed, got %v", err)
	}
	if decoded.Service != HostServiceHostConfig || len(decoded.Keys) != 2 {
		t.Fatalf("unexpected decoded hostConfig host service: %#v", decoded)
	}
	if len(decoded.Resources) != 0 {
		t.Fatalf("expected hostConfig host service to decode without resource refs, got %#v", decoded.Resources)
	}
}

// TestHostServiceSpecYAMLUsesResourcePathsForManifest verifies manifest
// services marshal and unmarshal through `resources.paths`.
func TestHostServiceSpecYAMLUsesResourcePathsForManifest(t *testing.T) {
	spec := &HostServiceSpec{
		Service: HostServiceManifest,
		Methods: []string{HostServiceMethodManifestGet},
		Paths:   []string{"metadata.yaml", "resources/*.yaml"},
	}

	payload, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("expected manifest host service yaml marshal to succeed, got %v", err)
	}

	var encoded map[string]interface{}
	if err = yaml.Unmarshal(payload, &encoded); err != nil {
		t.Fatalf("expected marshaled manifest host service yaml to decode, got %v", err)
	}
	resources, ok := encoded["resources"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected manifest host service yaml resources object, got %#v", encoded["resources"])
	}
	paths, ok := resources["paths"].([]interface{})
	if !ok || len(paths) != 2 {
		t.Fatalf("expected manifest host service yaml resources.paths, got %#v", resources["paths"])
	}

	decoded := &HostServiceSpec{}
	if err = yaml.Unmarshal(payload, decoded); err != nil {
		t.Fatalf("expected manifest host service yaml unmarshal to succeed, got %v", err)
	}
	if decoded.Service != HostServiceManifest || len(decoded.Paths) != 2 {
		t.Fatalf("unexpected decoded manifest host service: %#v", decoded)
	}
	if len(decoded.Resources) != 0 {
		t.Fatalf("expected manifest host service to decode without resource refs, got %#v", decoded.Resources)
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
			Service: HostServiceNotifications,
			Methods: []string{HostServiceMethodNotificationsSend},
			Resources: []*HostServiceResourceSpec{
				{Ref: "inbox"},
			},
		},
	})

	if len(capabilities) != 3 {
		t.Fatalf("expected 3 derived capabilities, got %#v", capabilities)
	}
	if capabilities[0] != CapabilityCache || capabilities[1] != CapabilityLock || capabilities[2] != CapabilityNotifications {
		t.Fatalf("unexpected derived low priority capabilities ordering: %#v", capabilities)
	}
}

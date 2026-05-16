// This file contains source-plugin upgrade lifecycle callback planning and
// snapshot conversion helpers.

package sourceupgrade

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginhost"
)

// sourceUpgradePlan keeps the validated input and persisted snapshots used by
// source-plugin upgrade callbacks and failure recording.
type sourceUpgradePlan struct {
	// fromSnapshot is the effective manifest snapshot before upgrade.
	fromSnapshot *catalog.ManifestSnapshot
	// toSnapshot is the discovered target manifest snapshot.
	toSnapshot *catalog.ManifestSnapshot
	// callbackInput is the published upgrade callback input.
	callbackInput pluginhost.SourcePluginUpgradeInput
}

// buildSourceUpgradePlan creates the published callback input from effective
// and target release manifest snapshots.
func (s *serviceImpl) buildSourceUpgradePlan(
	currentRelease *entity.SysPluginRelease,
	targetRelease *entity.SysPluginRelease,
	fromVersion string,
	toVersion string,
) (*sourceUpgradePlan, error) {
	if currentRelease == nil {
		return nil, bizerr.NewCode(CodePluginSourceUpgradeRegistryRequired)
	}
	if targetRelease == nil {
		return nil, bizerr.NewCode(CodePluginSourceUpgradeTargetReleaseRequired)
	}

	fromSnapshot, err := s.catalogSvc.ParseManifestSnapshot(currentRelease.ManifestSnapshot)
	if err != nil {
		return nil, err
	}
	toSnapshot, err := s.catalogSvc.ParseManifestSnapshot(targetRelease.ManifestSnapshot)
	if err != nil {
		return nil, err
	}
	if fromSnapshot == nil || toSnapshot == nil {
		return nil, bizerr.NewCode(
			CodePluginSourceUpgradeTargetReleaseRequired,
			bizerr.P("pluginId", targetRelease.PluginId),
		)
	}

	input := pluginhost.NewSourcePluginUpgradeInput(
		targetRelease.PluginId,
		strings.TrimSpace(fromVersion),
		strings.TrimSpace(toVersion),
		sourceManifestSnapshotView(fromSnapshot),
		sourceManifestSnapshotView(toSnapshot),
	)
	return &sourceUpgradePlan{
		fromSnapshot:  fromSnapshot,
		toSnapshot:    toSnapshot,
		callbackInput: input,
	}, nil
}

// executeBeforeSourceUpgrade runs unified lifecycle pre-upgrade callbacks.
func (s *serviceImpl) executeBeforeSourceUpgrade(
	ctx context.Context,
	manifest *catalog.Manifest,
	plan *sourceUpgradePlan,
) error {
	if manifest == nil || manifest.SourcePlugin == nil || plan == nil {
		return nil
	}
	participants := []pluginhost.LifecycleParticipant{
		{
			PluginID: manifest.ID,
			Callback: pluginhost.NewSourcePluginLifecycleCallbackAdapter(manifest.SourcePlugin),
		},
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook:         pluginhost.LifecycleHookBeforeUpgrade,
		UpgradeInput: plan.callbackInput,
		Participants: participants,
	})
	if result.OK {
		return nil
	}
	return bizerr.NewCode(
		CodePluginSourceUpgradeLifecycleVetoed,
		bizerr.P("pluginId", manifest.ID),
		bizerr.P("operation", pluginhost.LifecycleHookBeforeUpgrade.String()),
		bizerr.P("reasons", summarizeSourceUpgradeVetoReasons(result.Decisions)),
	)
}

// executeSourceUpgradeCallback invokes the target-version plugin's custom
// upgrade callback when the plugin registered one.
func (s *serviceImpl) executeSourceUpgradeCallback(
	ctx context.Context,
	manifest *catalog.Manifest,
	plan *sourceUpgradePlan,
) error {
	if manifest == nil || manifest.SourcePlugin == nil || plan == nil {
		return nil
	}
	handler := manifest.SourcePlugin.GetUpgradeHandler()
	if handler == nil {
		return nil
	}
	return handler(ctx, plan.callbackInput)
}

// executeAfterSourceUpgrade invokes the optional post-upgrade lifecycle
// callback after host state has become effective.
func (s *serviceImpl) executeAfterSourceUpgrade(
	ctx context.Context,
	manifest *catalog.Manifest,
	plan *sourceUpgradePlan,
) error {
	if manifest == nil || manifest.SourcePlugin == nil || plan == nil {
		return nil
	}
	participants := []pluginhost.LifecycleParticipant{
		{
			PluginID: manifest.ID,
			Callback: pluginhost.NewSourcePluginLifecycleCallbackAdapter(manifest.SourcePlugin),
		},
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook:         pluginhost.LifecycleHookAfterUpgrade,
		UpgradeInput: plan.callbackInput,
		Participants: participants,
	})
	if result.OK {
		return nil
	}
	return bizerr.NewCode(
		CodePluginSourceUpgradeLifecycleVetoed,
		bizerr.P("pluginId", manifest.ID),
		bizerr.P("operation", pluginhost.LifecycleHookAfterUpgrade.String()),
		bizerr.P("reasons", summarizeSourceUpgradeVetoReasons(result.Decisions)),
	)
}

// sourceManifestSnapshotView converts a catalog snapshot into the stable
// pluginhost manifest snapshot view published to source plugins.
func sourceManifestSnapshotView(snapshot *catalog.ManifestSnapshot) pluginhost.ManifestSnapshot {
	if snapshot == nil {
		return nil
	}
	return pluginhost.NewManifestSnapshotFromValues(pluginhost.ManifestSnapshotValues{
		ID:                    snapshot.ID,
		Name:                  snapshot.Name,
		Version:               snapshot.Version,
		Type:                  snapshot.Type,
		ScopeNature:           snapshot.ScopeNature,
		SupportsMultiTenant:   snapshot.SupportsMultiTenant,
		DefaultInstallMode:    snapshot.DefaultInstallMode,
		Description:           snapshot.Description,
		InstallSQLCount:       snapshot.InstallSQLCount,
		UninstallSQLCount:     snapshot.UninstallSQLCount,
		MockSQLCount:          snapshot.MockSQLCount,
		MenuCount:             snapshot.MenuCount,
		BackendHookCount:      snapshot.BackendHookCount,
		ResourceSpecCount:     snapshot.ResourceSpecCount,
		HostServiceAuthNeeded: snapshot.HostServiceAuthRequired,
	})
}

// summarizeSourceUpgradeVetoReasons builds one deterministic veto summary for
// source-upgrade bizerr params and audit logs.
func summarizeSourceUpgradeVetoReasons(decisions []pluginhost.LifecycleDecision) string {
	items := make([]string, 0, len(decisions))
	for _, decision := range decisions {
		if decision.OK {
			continue
		}
		reason := strings.TrimSpace(decision.Reason)
		if reason == "" && decision.Err != nil {
			reason = decision.Err.Error()
		}
		if reason == "" {
			reason = "plugin." + strings.TrimSpace(decision.PluginID) + ".lifecycle.vetoed"
		}
		items = append(items, strings.TrimSpace(decision.PluginID)+":"+reason)
	}
	if len(items) == 0 {
		return "unknown"
	}
	return strings.Join(items, ";")
}

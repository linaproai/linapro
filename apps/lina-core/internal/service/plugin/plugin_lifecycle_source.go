// This file keeps source-plugin lifecycle helpers that have not yet migrated
// from the root facade during the C-stage lifecycle refactor.

package plugin

import (
	"context"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/pluginhost"
)

// sourceLifecyclePolicy carries host-side action options into source-plugin
// generic lifecycle callbacks.
type sourceLifecyclePolicy struct {
	startupAutoEnable bool
	purgeStorageData  bool
}

// executeSourcePluginBeforeLifecycle invokes lifecycle facade precondition
// callbacks registered by one source plugin before host side effects run.
func (s *serviceImpl) executeSourcePluginBeforeLifecycle(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
	policy sourceLifecyclePolicy,
) error {
	if manifest == nil || manifest.SourcePlugin == nil {
		return nil
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook: hook,
		PluginInput: pluginhost.NewSourcePluginLifecycleInputWithPolicy(
			manifest.ID,
			hook.String(),
			pluginhost.SourcePluginLifecyclePolicy{
				StartupAutoEnable: policy.startupAutoEnable,
				PurgeStorageData:  policy.purgeStorageData,
			},
		),
		Participants: []pluginhost.LifecycleParticipant{
			{
				PluginID: manifest.ID,
				Callback: pluginhost.NewSourcePluginLifecycleCallbackAdapter(manifest.SourcePlugin),
			},
		},
	})
	if result.OK {
		return nil
	}
	reasons := s.summarizeLocalizedLifecycleVetoReasons(ctx, result.Decisions)
	return bizerr.NewCode(
		CodePluginLifecyclePreconditionVetoed,
		bizerr.P("operation", hook.String()),
		bizerr.P("pluginId", manifest.ID),
		bizerr.P("reasons", reasons),
	)
}

// executeSourcePluginAfterLifecycle invokes non-blocking lifecycle callbacks
// registered by one source plugin after host side effects have succeeded.
func (s *serviceImpl) executeSourcePluginAfterLifecycle(
	ctx context.Context,
	manifest *catalog.Manifest,
	hook pluginhost.LifecycleHook,
	policy sourceLifecyclePolicy,
) {
	if manifest == nil || manifest.SourcePlugin == nil {
		return
	}
	result := pluginhost.RunLifecycleCallbacks(ctx, pluginhost.LifecycleRequest{
		Hook: hook,
		PluginInput: pluginhost.NewSourcePluginLifecycleInputWithPolicy(
			manifest.ID,
			hook.String(),
			pluginhost.SourcePluginLifecyclePolicy{
				StartupAutoEnable: policy.startupAutoEnable,
				PurgeStorageData:  policy.purgeStorageData,
			},
		),
		Participants: []pluginhost.LifecycleParticipant{
			{
				PluginID: manifest.ID,
				Callback: pluginhost.NewSourcePluginLifecycleCallbackAdapter(manifest.SourcePlugin),
			},
		},
	})
	if result.OK {
		return
	}
	logger.Warningf(
		ctx,
		"source plugin after lifecycle callback failed operation=%s plugin=%s reasons=%s",
		hook,
		manifest.ID,
		summarizeLifecycleVetoReasons(result.Decisions),
	)
}

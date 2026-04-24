// Package sourceupgrade implements source-plugin upgrade discovery, explicit
// upgrade execution, and startup fail-fast validation for the host plugin domain.
package sourceupgrade

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/runtime"
	sourceupgradecontract "lina-core/pkg/sourceupgrade/contract"
)

type (
	// SourceUpgradeStatus aliases the stable source-plugin upgrade status contract.
	SourceUpgradeStatus = sourceupgradecontract.SourcePluginStatus

	// SourceUpgradeResult aliases the stable explicit source-plugin upgrade result contract.
	SourceUpgradeResult = sourceupgradecontract.SourcePluginUpgradeResult
)

// Service defines the host-side source-plugin upgrade governance contract.
type Service interface {
	// ListSourceUpgradeStatuses scans source manifests and returns one
	// effective-versus-discovered upgrade-status item per source plugin.
	ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error)
	// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
	// current effective version to the newer discovered source version.
	UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error)
	// ValidateSourcePluginUpgradeReadiness fails fast when any installed source
	// plugin still has a newer discovered source version waiting to be upgraded.
	ValidateSourcePluginUpgradeReadiness(ctx context.Context) error
}

// Ensure serviceImpl satisfies Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest discovery, registry, and release governance.
	catalogSvc catalog.Service
	// lifecycleSvc provides install/uninstall lifecycle orchestration.
	lifecycleSvc lifecycle.Service
	// runtimeSvc provides dynamic plugin reconciliation and route dispatch.
	runtimeSvc runtime.Service
	// integrationSvc provides host extension, menu, hook, and resource integration.
	integrationSvc integration.Service
}

// New creates and returns a new source-plugin upgrade governance service.
func New(
	catalogSvc catalog.Service,
	lifecycleSvc lifecycle.Service,
	runtimeSvc runtime.Service,
	integrationSvc integration.Service,
) Service {
	return &serviceImpl{
		catalogSvc:     catalogSvc,
		lifecycleSvc:   lifecycleSvc,
		runtimeSvc:     runtimeSvc,
		integrationSvc: integrationSvc,
	}
}

// sourceUpgradeCandidate keeps the discovered manifest, current registry row,
// and flattened upgrade status together during one planning/execution cycle.
type sourceUpgradeCandidate struct {
	// manifest is the discovered source-plugin manifest.
	manifest *catalog.Manifest
	// registry is the synchronized registry row for the plugin.
	registry *entity.SysPlugin
	// status is the flattened upgrade status projection used by callers.
	status *SourceUpgradeStatus
}

// ListSourceUpgradeStatuses scans source manifests and returns one
// effective-versus-discovered upgrade-status item per source plugin.
func (s *serviceImpl) ListSourceUpgradeStatuses(ctx context.Context) ([]*SourceUpgradeStatus, error) {
	candidates, err := s.listSourceUpgradeCandidates(ctx, false)
	if err != nil {
		return nil, err
	}

	items := make([]*SourceUpgradeStatus, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == nil || candidate.status == nil {
			continue
		}
		items = append(items, candidate.status)
	}
	return items, nil
}

// UpgradeSourcePlugin applies one explicit source-plugin upgrade from the
// current effective version to the newer discovered source version.
func (s *serviceImpl) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*SourceUpgradeResult, error) {
	candidate, err := s.findSourceUpgradeCandidate(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if candidate == nil || candidate.manifest == nil || candidate.status == nil {
		return nil, gerror.Newf("源码插件升级候选不存在: %s", pluginID)
	}

	result := &SourceUpgradeResult{
		PluginID:    candidate.status.PluginID,
		Name:        candidate.status.Name,
		FromVersion: candidate.status.EffectiveVersion,
		ToVersion:   candidate.status.DiscoveredVersion,
	}
	if candidate.status.Installed != catalog.InstalledYes {
		result.Message = "源码插件未安装，跳过升级。"
		return result, nil
	}

	registry, err := s.catalogSvc.SyncManifest(ctx, candidate.manifest)
	if err != nil {
		return nil, err
	}
	candidate.registry = registry
	candidate.status, err = buildSourceUpgradeStatus(candidate.manifest, registry)
	if err != nil {
		return nil, err
	}
	result.FromVersion = candidate.status.EffectiveVersion
	result.ToVersion = candidate.status.DiscoveredVersion

	versionCompare, err := compareSourceUpgradeVersions(
		candidate.status.EffectiveVersion,
		candidate.status.DiscoveredVersion,
	)
	if err != nil {
		return nil, err
	}
	if versionCompare == 0 {
		result.Message = "当前源码插件已是最新版本，无需升级。"
		return result, nil
	}
	if versionCompare > 0 {
		return nil, gerror.Newf(
			"源码插件 %s 当前生效版本 %s 高于源码发现版本 %s，当前不支持降级或回滚",
			candidate.status.PluginID,
			candidate.status.EffectiveVersion,
			candidate.status.DiscoveredVersion,
		)
	}

	targetRelease, err := s.catalogSvc.GetRelease(
		ctx,
		candidate.manifest.ID,
		candidate.manifest.Version,
	)
	if err != nil {
		return nil, err
	}
	if targetRelease == nil {
		return nil, gerror.Newf(
			"源码插件升级缺少目标发布记录: %s@%s",
			candidate.manifest.ID,
			candidate.manifest.Version,
		)
	}

	currentRelease, err := s.catalogSvc.GetRegistryRelease(ctx, candidate.registry)
	if err != nil {
		return nil, err
	}

	if err = s.lifecycleSvc.ExecuteManifestSQLFiles(
		ctx,
		candidate.manifest,
		catalog.MigrationDirectionUpgrade,
	); err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}
	if err = s.integrationSvc.SyncPluginMenusAndPermissions(ctx, candidate.manifest); err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}
	if err = s.integrationSvc.SyncPluginResourceReferences(ctx, candidate.manifest); err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}
	if err = s.applySourcePluginUpgradedRelease(ctx, candidate.registry, candidate.manifest, targetRelease); err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}

	updatedRegistry, err := s.catalogSvc.GetRegistry(ctx, candidate.manifest.ID)
	if err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}
	if updatedRegistry == nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, gerror.Newf("源码插件升级后注册表不存在: %s", candidate.manifest.ID)
	}

	if currentRelease != nil && currentRelease.Id > 0 && currentRelease.Id != targetRelease.Id {
		if err = s.catalogSvc.UpdateReleaseState(
			ctx,
			currentRelease.Id,
			catalog.ReleaseStatusInstalled,
			"",
		); err != nil {
			s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
			return nil, err
		}
	}
	if err = s.catalogSvc.UpdateReleaseState(
		ctx,
		targetRelease.Id,
		catalog.BuildReleaseStatus(updatedRegistry.Installed, updatedRegistry.Status),
		s.catalogSvc.BuildPackagePath(candidate.manifest),
	); err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}
	if err = s.runtimeSvc.SyncPluginNodeState(
		ctx,
		updatedRegistry.PluginId,
		updatedRegistry.Version,
		updatedRegistry.Installed,
		updatedRegistry.Status,
		"Source plugin upgraded by development upgrade tool.",
	); err != nil {
		s.markSourcePluginReleaseFailed(ctx, candidate.manifest, targetRelease)
		return nil, err
	}

	result.Executed = true
	result.Message = fmt.Sprintf(
		"源码插件已从 %s 升级到 %s。",
		candidate.status.EffectiveVersion,
		candidate.status.DiscoveredVersion,
	)
	return result, nil
}

// ValidateSourcePluginUpgradeReadiness fails fast when any installed source
// plugin still has a newer discovered source version waiting to be upgraded.
func (s *serviceImpl) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	statuses, err := s.ListSourceUpgradeStatuses(ctx)
	if err != nil {
		return err
	}

	pending := make([]*SourceUpgradeStatus, 0)
	for _, item := range statuses {
		if item == nil || !item.NeedsUpgrade {
			continue
		}
		pending = append(pending, item)
	}
	if len(pending) == 0 {
		return nil
	}
	return buildSourcePluginUpgradePendingError(pending)
}

// listSourceUpgradeCandidates scans all discovered source manifests and returns
// their upgrade-governance view in stable plugin-ID order.
func (s *serviceImpl) listSourceUpgradeCandidates(
	ctx context.Context,
	synchronize bool,
) ([]*sourceUpgradeCandidate, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}

	items := make([]*sourceUpgradeCandidate, 0)
	for _, manifest := range manifests {
		if manifest == nil || catalog.NormalizeType(manifest.Type) != catalog.TypeSource {
			continue
		}

		var registry *entity.SysPlugin
		if synchronize {
			registry, err = s.catalogSvc.SyncManifest(ctx, manifest)
			if err != nil {
				return nil, err
			}
		} else {
			registry, err = s.catalogSvc.GetRegistry(ctx, manifest.ID)
			if err != nil {
				return nil, err
			}
		}
		status, err := buildSourceUpgradeStatus(manifest, registry)
		if err != nil {
			return nil, err
		}
		items = append(items, &sourceUpgradeCandidate{
			manifest: manifest,
			registry: registry,
			status:   status,
		})
	}
	return items, nil
}

// findSourceUpgradeCandidate returns the synchronized upgrade candidate for the
// requested source plugin identifier.
func (s *serviceImpl) findSourceUpgradeCandidate(ctx context.Context, pluginID string) (*sourceUpgradeCandidate, error) {
	normalizedID := strings.TrimSpace(pluginID)
	if normalizedID == "" {
		return nil, gerror.New("源码插件 ID 不能为空")
	}

	candidates, err := s.listSourceUpgradeCandidates(ctx, false)
	if err != nil {
		return nil, err
	}
	for _, candidate := range candidates {
		if candidate == nil || candidate.status == nil {
			continue
		}
		if candidate.status.PluginID == normalizedID {
			return candidate, nil
		}
	}
	return nil, gerror.Newf("未找到源码插件: %s", normalizedID)
}

// buildSourceUpgradeStatus flattens the manifest and registry state of one
// source plugin into an operator-facing upgrade-status projection.
func buildSourceUpgradeStatus(
	manifest *catalog.Manifest,
	registry *entity.SysPlugin,
) (*SourceUpgradeStatus, error) {
	if manifest == nil {
		return nil, gerror.New("源码插件清单不能为空")
	}

	status := &SourceUpgradeStatus{
		PluginID:          strings.TrimSpace(manifest.ID),
		Name:              strings.TrimSpace(manifest.Name),
		DiscoveredVersion: strings.TrimSpace(manifest.Version),
	}
	if registry != nil {
		if strings.TrimSpace(registry.PluginId) != "" {
			status.PluginID = strings.TrimSpace(registry.PluginId)
		}
		if strings.TrimSpace(registry.Name) != "" {
			status.Name = strings.TrimSpace(registry.Name)
		}
		status.EffectiveVersion = strings.TrimSpace(registry.Version)
		status.Installed = registry.Installed
		status.Enabled = registry.Status
	}
	if status.Installed != catalog.InstalledYes {
		return status, nil
	}

	versionCompare, err := compareSourceUpgradeVersions(status.EffectiveVersion, status.DiscoveredVersion)
	if err != nil {
		return nil, err
	}
	status.NeedsUpgrade = versionCompare < 0
	status.DowngradeDetected = versionCompare > 0
	return status, nil
}

// compareSourceUpgradeVersions compares the current effective version and the
// currently discovered source version for one source plugin.
func compareSourceUpgradeVersions(effectiveVersion string, discoveredVersion string) (int, error) {
	effective := strings.TrimSpace(effectiveVersion)
	discovered := strings.TrimSpace(discoveredVersion)
	if effective == "" || discovered == "" {
		return 0, nil
	}
	return catalog.CompareSemanticVersions(effective, discovered)
}

// applySourcePluginUpgradedRelease promotes the discovered release to the
// current effective registry version without changing installed/enabled flags.
func (s *serviceImpl) applySourcePluginUpgradedRelease(
	ctx context.Context,
	registry *entity.SysPlugin,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
) error {
	if registry == nil {
		return gerror.New("源码插件注册表不能为空")
	}
	if manifest == nil {
		return gerror.New("源码插件清单不能为空")
	}
	if release == nil {
		return gerror.New("源码插件目标发布记录不能为空")
	}

	stableState := catalog.DeriveHostState(registry.Installed, registry.Status)
	data := do.SysPlugin{
		Name:         manifest.Name,
		Version:      manifest.Version,
		Type:         manifest.Type,
		ReleaseId:    release.Id,
		Installed:    registry.Installed,
		Status:       registry.Status,
		DesiredState: stableState,
		CurrentState: stableState,
		ManifestPath: manifest.ManifestPath,
		Checksum:     s.catalogSvc.BuildRegistryChecksum(manifest),
		Remark:       manifest.Description,
	}
	if registry.Generation <= 0 {
		data.Generation = int64(1)
	}

	_, err := dao.SysPlugin.Ctx(ctx).
		Where(do.SysPlugin{PluginId: registry.PluginId}).
		Data(data).
		Update()
	return err
}

// markSourcePluginReleaseFailed best-effort records that one explicit
// source-plugin upgrade stopped after the target release had been prepared.
func (s *serviceImpl) markSourcePluginReleaseFailed(
	ctx context.Context,
	manifest *catalog.Manifest,
	release *entity.SysPluginRelease,
) {
	if manifest == nil || release == nil {
		return
	}
	if err := s.catalogSvc.UpdateReleaseState(
		ctx,
		release.Id,
		catalog.ReleaseStatusFailed,
		s.catalogSvc.BuildPackagePath(manifest),
	); err != nil {
		return
	}
}

// buildSourcePluginUpgradePendingError formats one startup fail-fast message
// listing all installed source plugins that still need explicit upgrade.
func buildSourcePluginUpgradePendingError(pending []*SourceUpgradeStatus) error {
	if len(pending) == 0 {
		return nil
	}

	lines := make([]string, 0, len(pending)+2)
	lines = append(lines, "检测到已安装源码插件存在待完成升级，宿主启动前必须先执行开发态升级命令。")
	for _, item := range pending {
		if item == nil {
			continue
		}
		lines = append(
			lines,
			fmt.Sprintf(
				"- plugin=%s current=%s discovered=%s command=make upgrade confirm=upgrade scope=source-plugin plugin=%s",
				item.PluginID,
				item.EffectiveVersion,
				item.DiscoveredVersion,
				item.PluginID,
			),
		)
	}
	if len(pending) > 1 {
		lines = append(lines, "如需一次性处理全部待升级源码插件，可执行：make upgrade confirm=upgrade scope=source-plugin plugin=all")
	}
	return gerror.New(strings.Join(lines, "\n"))
}

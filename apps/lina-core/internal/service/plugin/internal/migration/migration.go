// Package migration executes plugin SQL migrations and records abstract
// migration history entries for later review and lifecycle reconciliation.
package migration

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/datahost"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/internal/service/plugin/internal/store"
	"lina-core/pkg/dialect"
)

// Service defines the plugin SQL migration executor contract.
type Service interface {
	// ExecuteManifestSQLFiles executes plugin manifest SQL files and records every attempt
	// in sys_plugin_migration. The mock phase is intentionally excluded from this entry
	// point because mock data must be loaded transactionally via
	// ExecuteManifestMockSQLFilesInTx.
	ExecuteManifestSQLFiles(
		ctx context.Context,
		manifest *catalog.Manifest,
		direction plugintypes.MigrationDirection,
	) error
	// ExecuteManifestMockSQLFilesInTx executes a plugin's mock-data SQL files inside the
	// caller-supplied transaction and records each step in sys_plugin_migration.
	ExecuteManifestMockSQLFilesInTx(
		ctx context.Context,
		manifest *catalog.Manifest,
	) MockSQLExecutionResult
	// ExecuteManifestMockSQLFiles executes a plugin's optional mock-data SQL in
	// one transaction and returns MockDataLoadError when the load rolls back.
	ExecuteManifestMockSQLFiles(ctx context.Context, manifest *catalog.Manifest) error
	// ResolveSQLAssets extracts lifecycle SQL either from embedded runtime artifact sections
	// or from source-style directory conventions, while preserving execution order.
	ResolveSQLAssets(
		manifest *catalog.Manifest,
		direction plugintypes.MigrationDirection,
	) ([]*SQLAsset, error)
	// ResolvePluginSQLAssets resolves SQL assets from the manifest and returns them as catalog.ArtifactSQLAsset
	// slices for callers that expect the catalog asset type rather than migration.SQLAsset.
	ResolvePluginSQLAssets(manifest *catalog.Manifest, direction plugintypes.MigrationDirection) ([]*catalog.ArtifactSQLAsset, error)
}

// Ensure serviceImpl satisfies the migration executor contract.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	// catalogSvc provides manifest asset path discovery and content reads.
	catalogSvc catalog.Service
	// storeSvc provides release records used for migration ledger ownership.
	storeSvc storeReader
}

// storeReader defines the narrow store capability needed by migration ledger writes.
type storeReader interface {
	// GetRelease returns the governance release row for one plugin version.
	GetRelease(ctx context.Context, pluginID string, version string) (*store.ReleaseRecord, error)
}

// New creates a plugin SQL migration executor.
func New(catalogSvc catalog.Service, storeSvc storeReader) Service {
	return &serviceImpl{
		catalogSvc: catalogSvc,
		storeSvc:   storeSvc,
	}
}

// SQLAsset describes one install/uninstall SQL step after host extraction.
type SQLAsset struct {
	// Key is the canonical identifier for this SQL step.
	Key string
	// Content is the raw SQL text to execute.
	Content string
}

// ExecuteManifestSQLFiles executes plugin manifest SQL files and records every attempt
// in sys_plugin_migration. The mock phase is intentionally excluded from this entry
// point because mock data must be loaded transactionally via
// ExecuteManifestMockSQLFilesInTx; callers that want to load mock data MUST go
// through that method instead.
func (s *serviceImpl) ExecuteManifestSQLFiles(
	ctx context.Context,
	manifest *catalog.Manifest,
	direction plugintypes.MigrationDirection,
) error {
	if manifest == nil {
		return gerror.New("plugin manifest cannot be nil")
	}
	if direction == plugintypes.MigrationDirectionMock {
		return gerror.New("mock SQL files must be executed via ExecuteManifestMockSQLFilesInTx")
	}

	sqlAssets, err := s.ResolveSQLAssets(manifest, direction)
	if err != nil {
		return err
	}

	release, err := s.storeSvc.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return err
	}
	if release == nil {
		return gerror.Newf("plugin release record does not exist: %s@%s", manifest.ID, manifest.Version)
	}

	err = dao.SysPluginMigration.Transaction(ctx, func(txCtx context.Context, _ gdb.TX) error {
		for index, asset := range sqlAssets {
			if asset == nil {
				return gerror.New("plugin SQL asset cannot be nil")
			}
			var (
				checksum     = fmt.Sprintf("%x", sha256.Sum256([]byte(asset.Content)))
				migrationKey = buildMigrationKey(direction, index+1)
				executedAt   = time.Now()
				execErr      = s.executeSQLAsset(txCtx, manifest.ID, direction, asset)
			)
			if execErr != nil {
				return gerror.Wrapf(execErr, "execute plugin SQL failed: %s", asset.Key)
			}
			if recordErr := s.recordMigration(txCtx, manifest.ID, release.Id, direction, migrationKey, index+1, checksum, &executedAt, nil); recordErr != nil {
				return recordErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	datahost.InvalidateTableContractCache(ctx, manifest.ID, "plugin_sql_"+direction.String())
	return nil
}

// ResolveSQLAssets extracts lifecycle SQL either from embedded runtime artifact sections
// or from source-style directory conventions, while preserving execution order.
func (s *serviceImpl) ResolveSQLAssets(
	manifest *catalog.Manifest,
	direction plugintypes.MigrationDirection,
) ([]*SQLAsset, error) {
	if manifest == nil {
		return []*SQLAsset{}, nil
	}

	if manifest.RuntimeArtifact != nil {
		var embeddedAssets []*catalog.ArtifactSQLAsset
		switch direction {
		case plugintypes.MigrationDirectionUninstall:
			embeddedAssets = manifest.RuntimeArtifact.UninstallSQLAssets
		case plugintypes.MigrationDirectionMock:
			embeddedAssets = manifest.RuntimeArtifact.MockSQLAssets
		default:
			embeddedAssets = manifest.RuntimeArtifact.InstallSQLAssets
		}
		if len(embeddedAssets) > 0 {
			items := make([]*SQLAsset, 0, len(embeddedAssets))
			for _, asset := range embeddedAssets {
				if asset == nil {
					continue
				}
				items = append(items, &SQLAsset{
					Key:     asset.Key,
					Content: asset.Content,
				})
			}
			return items, nil
		}
	}

	var relativePaths []string
	switch direction {
	case plugintypes.MigrationDirectionUninstall:
		relativePaths = s.catalogSvc.ListUninstallSQLPaths(manifest)
	case plugintypes.MigrationDirectionMock:
		relativePaths = s.catalogSvc.ListMockSQLPaths(manifest)
	default:
		relativePaths = s.catalogSvc.ListInstallSQLPaths(manifest)
	}
	items := make([]*SQLAsset, 0, len(relativePaths))
	for _, relativePath := range relativePaths {
		sqlContent, err := s.catalogSvc.ReadSourcePluginAssetContent(manifest, relativePath)
		if err != nil {
			return nil, err
		}
		if sqlContent == "" {
			return nil, gerror.Newf("plugin SQL file is empty: %s", relativePath)
		}
		items = append(items, &SQLAsset{
			Key:     filepath.Base(relativePath),
			Content: sqlContent,
		})
	}
	return items, nil
}

// executeSQLAsset translates and executes one plugin SQL asset statement by
// statement so execution never depends on driver-level multi-statement support.
func (s *serviceImpl) executeSQLAsset(
	ctx context.Context,
	pluginID string,
	direction plugintypes.MigrationDirection,
	asset *SQLAsset,
) error {
	if asset == nil {
		return gerror.New("plugin SQL asset cannot be nil")
	}
	dbDialect, err := currentDialect(ctx)
	if err != nil {
		return err
	}
	sourceName := pluginSQLSourceName(pluginID, direction, asset.Key)
	content, err := dbDialect.TranslateDDL(ctx, sourceName, asset.Content)
	if err != nil {
		return gerror.Wrapf(err, "translate plugin SQL failed: %s", sourceName)
	}
	for index, statement := range dialect.SplitSQLStatements(content) {
		if _, err = g.DB().Exec(ctx, statement); err != nil {
			return gerror.Wrapf(err, "execute plugin SQL statement %d failed: %s", index+1, sourceName)
		}
	}
	return nil
}

// currentDialect resolves the active database dialect from database.default.link.
func currentDialect(ctx context.Context) (dialect.Dialect, error) {
	linkVar, err := g.Cfg().Get(ctx, "database.default.link")
	if err != nil {
		return nil, gerror.Wrap(err, "read database connection configuration failed")
	}
	if linkVar == nil {
		return nil, gerror.New("database connection configuration database.default.link must not be empty")
	}
	link := strings.TrimSpace(linkVar.String())
	if link == "" {
		return nil, gerror.New("database connection configuration database.default.link must not be empty")
	}
	return dialect.From(link)
}

// pluginSQLSourceName builds the diagnostic source name passed to the dialect.
func pluginSQLSourceName(pluginID string, direction plugintypes.MigrationDirection, key string) string {
	return fmt.Sprintf("plugin=%s phase=%s file=%s", pluginID, direction.String(), key)
}

// ResolvePluginSQLAssets resolves SQL assets from the manifest and returns them as catalog.ArtifactSQLAsset
// slices for callers that expect the catalog asset type rather than migration.SQLAsset.
func (s *serviceImpl) ResolvePluginSQLAssets(manifest *catalog.Manifest, direction plugintypes.MigrationDirection) ([]*catalog.ArtifactSQLAsset, error) {
	assets, err := s.ResolveSQLAssets(manifest, direction)
	if err != nil {
		return nil, err
	}
	result := make([]*catalog.ArtifactSQLAsset, 0, len(assets))
	for _, a := range assets {
		if a == nil {
			continue
		}
		result = append(result, &catalog.ArtifactSQLAsset{Key: a.Key, Content: a.Content})
	}
	return result, nil
}

// buildMigrationKey returns the canonical migration key for a given phase and sequence number.
func buildMigrationKey(direction plugintypes.MigrationDirection, sequenceNo int) string {
	normalizedDirection := strings.TrimSpace(strings.ToLower(direction.String()))
	if normalizedDirection == "" {
		normalizedDirection = plugintypes.MigrationDirectionInstall.String()
	}
	if sequenceNo <= 0 {
		sequenceNo = 1
	}
	return fmt.Sprintf("%s-step-%03d", normalizedDirection, sequenceNo)
}

// getMigration loads one previously recorded migration attempt for the given
// release and migration key.
func (s *serviceImpl) getMigration(
	ctx context.Context,
	pluginID string,
	releaseID int,
	phase plugintypes.MigrationDirection,
	migrationKey string,
) (*entity.SysPluginMigration, error) {
	var migration *entity.SysPluginMigration
	err := dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{
			PluginId:     pluginID,
			ReleaseId:    releaseID,
			Phase:        phase.String(),
			MigrationKey: migrationKey,
		}).
		Scan(&migration)
	return migration, err
}

// recordMigration upserts the execution result for one SQL migration step so
// install and uninstall attempts remain auditable and re-runnable.
func (s *serviceImpl) recordMigration(
	ctx context.Context,
	pluginID string,
	releaseID int,
	phase plugintypes.MigrationDirection,
	migrationKey string,
	sequenceNo int,
	checksum string,
	executedAt *time.Time,
	execErr error,
) error {
	status := plugintypes.MigrationExecutionStatusSucceeded
	message := ""
	if execErr != nil {
		status = plugintypes.MigrationExecutionStatusFailed
		message = execErr.Error()
	}

	existing, err := s.getMigration(ctx, pluginID, releaseID, phase, migrationKey)
	if err != nil {
		return err
	}

	data := do.SysPluginMigration{
		PluginId:       pluginID,
		ReleaseId:      releaseID,
		Phase:          phase.String(),
		MigrationKey:   migrationKey,
		ExecutionOrder: sequenceNo,
		Checksum:       checksum,
		Status:         status.String(),
		ErrorMessage:   message,
		ExecutedAt:     executedAt,
	}

	if existing == nil {
		_, err = dao.SysPluginMigration.Ctx(ctx).Data(data).Insert()
		return err
	}

	_, err = dao.SysPluginMigration.Ctx(ctx).
		Where(do.SysPluginMigration{Id: existing.Id}).
		Data(data).
		Update()
	return err
}

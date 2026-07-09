// This file implements the host database upgrade command by replaying all
// idempotent host SQL assets against the configured database.

package cmd

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/cmd/internal/sqlassets"
	"lina-core/pkg/logger"
)

// UpgradeInput defines the command-line options for the database upgrade
// command.
type UpgradeInput struct {
	g.Meta    `name:"upgrade" brief:"replay host SQL assets for database upgrade, requires --confirm=upgrade"`
	Confirm   string `name:"confirm" brief:"explicit confirmation value, must be 'upgrade'"`
	SQLSource string `name:"sql-source" brief:"SQL asset source: embedded or local; defaults to embedded"`
}

// UpgradeOutput carries the command result placeholder.
type UpgradeOutput struct{}

// Upgrade replays all host SQL resources after an explicit safety confirmation
// is provided. Host SQL files are required to be idempotent, so replaying them
// upgrades an existing database to the latest delivered schema and seed state.
func (m *Main) Upgrade(ctx context.Context, in UpgradeInput) (out *UpgradeOutput, err error) {
	if err = requireCommandConfirmation(upgradeCommandName, in.Confirm); err != nil {
		return nil, err
	}
	source, err := sqlassets.ResolveSource(in.SQLSource)
	if err != nil {
		return nil, err
	}
	assets, err := sqlassets.ScanInit(ctx, source)
	if err != nil {
		return nil, gerror.Wrap(err, "scan upgrade SQL files failed")
	}
	if len(assets) == 0 {
		logger.Warning(ctx, "no SQL files found for upgrade")
		return
	}
	if err = sqlassets.Execute(ctx, assets); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Database upgrade completed.")
	return
}

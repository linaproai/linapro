// This file implements the host database initialization command with explicit
// SQL asset source selection for development and runtime execution.

package cmd

import (
	"context"
	"strings"

	"lina-core/internal/cmd/internal/dbconfig"
	"lina-core/pkg/dialect"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/cmd/internal/sqlassets"
	"lina-core/pkg/logger"
)

// InitInput defines the command-line options for the sensitive database
// initialization command.
type InitInput struct {
	g.Meta    `name:"init" brief:"initialize database schema and seed data (DDL + seed DML), requires --confirm=init"`
	Confirm   string `name:"confirm" brief:"explicit confirmation value, must be 'init'"`
	SQLSource string `name:"sql-source" brief:"SQL asset source: embedded or local; defaults to embedded"`
	Rebuild   string `name:"rebuild" brief:"whether to drop and recreate the configured database before initialization: true or false"`
}

// InitOutput carries the command result placeholder.
type InitOutput struct{}

// Init initializes host SQL resources after an explicit safety confirmation is
// provided.
func (m *Main) Init(ctx context.Context, in InitInput) (out *InitOutput, err error) {
	if err = requireCommandConfirmation(initCommandName, in.Confirm); err != nil {
		return nil, err
	}
	source, err := sqlassets.ResolveSource(in.SQLSource)
	if err != nil {
		return nil, err
	}
	rebuild, err := parseInitRebuildFlag(in.Rebuild)
	if err != nil {
		return nil, err
	}
	if err = prepareInitDatabase(ctx, rebuild); err != nil {
		return nil, err
	}
	assets, err := sqlassets.ScanInit(ctx, source)
	if err != nil {
		return nil, gerror.Wrap(err, "scan initialization SQL files failed")
	}
	if len(assets) == 0 {
		logger.Warning(ctx, "no SQL files found for initialization")
		return
	}
	if err = sqlassets.Execute(ctx, assets); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Database initialization completed.")
	return
}

// parseInitRebuildFlag parses the optional init rebuild flag while treating an
// omitted value as a non-destructive initialization.
func parseInitRebuildFlag(value string) (bool, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "false", "0", "no", "n", "off":
		return false, nil
	case "true", "1", "yes", "y", "on":
		return true, nil
	default:
		return false, gerror.Newf("unsupported rebuild value: %s; available values are true or false", value)
	}
}

// prepareInitDatabase creates the canonical database before executing SQL
// assets and drops it first when rebuild is explicitly enabled.
func prepareInitDatabase(ctx context.Context, rebuild bool) (err error) {
	link, err := dbconfig.CurrentDatabaseLink(ctx)
	if err != nil {
		return err
	}
	dbDialect, err := dialect.From(link)
	if err != nil {
		return err
	}
	return dbDialect.PrepareDatabase(ctx, link, rebuild)
}

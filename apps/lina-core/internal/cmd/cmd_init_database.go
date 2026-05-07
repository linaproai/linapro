// This file prepares the configured database before host SQL assets are
// executed by the init command.

package cmd

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/dialect"
)

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
	link, err := currentDatabaseLink(ctx)
	if err != nil {
		return err
	}
	dbDialect, err := dialect.From(link)
	if err != nil {
		return err
	}
	return dbDialect.PrepareDatabase(ctx, link, rebuild)
}

// currentDatabaseLink returns the configured database.default.link value.
func currentDatabaseLink(ctx context.Context) (string, error) {
	linkVar, err := g.Cfg().Get(ctx, "database.default.link")
	if err != nil {
		return "", gerror.Wrap(err, "read database connection configuration failed")
	}
	if linkVar == nil {
		return "", gerror.New("database connection configuration database.default.link must not be empty")
	}
	link := strings.TrimSpace(linkVar.String())
	if link == "" {
		return "", gerror.New("database connection configuration database.default.link must not be empty")
	}
	return link, nil
}

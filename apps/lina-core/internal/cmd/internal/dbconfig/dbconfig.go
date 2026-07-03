// Package dbconfig reads command-time database configuration shared by startup
// and maintenance command internals.
package dbconfig

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// CurrentDatabaseLink returns the configured database.default.link value.
func CurrentDatabaseLink(ctx context.Context) (string, error) {
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

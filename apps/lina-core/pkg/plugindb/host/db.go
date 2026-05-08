// This file implements the reusable host-side plugindb Driver / DB wrapper and
// DoCommit governance interception.

package host

import (
	"context"
	"fmt"
	"strings"
	"sync"

	mysqlDriver "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/logger"
	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// pluginDataDriverTypePrefix prefixes governed driver types registered for
// host-side plugin data access wrappers.
const pluginDataDriverTypePrefix = "plugin-data-"

// pluginDataDriver wraps one base SQL driver so new DB handles return the
// governed pluginDataDB implementation.
type pluginDataDriver struct {
	baseType string
}

// pluginDataDB decorates one GoFrame DB handle with DoCommit authorization and
// audit enforcement for plugin-owned data access.
type pluginDataDB struct {
	gdb.DB
}

// Governed host-side plugin data DB registry and cache state.
var (
	pluginDataDriverRegisterOnce sync.Once
	pluginDataDBCacheMu          sync.Mutex
	pluginDataDBCache            = make(map[string]gdb.DB)
)

// DB returns one governed host-side plugindb connection wrapper.
func DB() (gdb.DB, error) {
	registerPluginDataDrivers()

	baseDB := g.DB()
	if baseDB == nil {
		return nil, gerror.New("plugin data service database is not configured")
	}
	baseConfig := baseDB.GetConfig()
	if baseConfig == nil {
		return nil, gerror.New("plugin data service database config is missing")
	}

	configNode := *baseConfig
	driverType, err := pluginDataDriverType(configNode.Type)
	if err != nil {
		return nil, err
	}
	configNode.Type = driverType
	configNode.Link = ""

	cacheKey := buildPluginDataDBCacheKey(&configNode)
	pluginDataDBCacheMu.Lock()
	defer pluginDataDBCacheMu.Unlock()
	if db, ok := pluginDataDBCache[cacheKey]; ok {
		return db, nil
	}

	db, err := gdb.New(configNode)
	if err != nil {
		return nil, err
	}
	db.SetDebug(baseDB.GetDebug())
	pluginDataDBCache[cacheKey] = db
	return db, nil
}

// registerPluginDataDrivers installs the governed DB drivers once per process.
func registerPluginDataDrivers() {
	pluginDataDriverRegisterOnce.Do(func() {
		for _, baseType := range []string{"mysql", "mariadb", "tidb"} {
			if err := gdb.Register(pluginDataDriverTypePrefix+baseType, &pluginDataDriver{baseType: baseType}); err != nil {
				panic(gerror.Wrapf(err, "register plugin data driver failed baseType=%s", baseType))
			}
		}
	})
}

// pluginDataDriverType normalizes one base driver type into the governed
// wrapper driver type understood by DB.
func pluginDataDriverType(baseType string) (string, error) {
	normalizedBaseType := strings.ToLower(strings.TrimSpace(baseType))
	switch normalizedBaseType {
	case "mysql", "mariadb", "tidb":
		return pluginDataDriverTypePrefix + normalizedBaseType, nil
	default:
		return "", gerror.Newf("plugin data service does not support database type: %s", baseType)
	}
}

// buildPluginDataDBCacheKey builds the in-process cache key for one governed
// DB handle derived from the effective config node.
func buildPluginDataDBCacheKey(config *gdb.ConfigNode) string {
	if config == nil {
		return ""
	}
	return fmt.Sprintf(
		"%s|%s|%s|%s|%s|%s|%s",
		config.Type,
		config.Link,
		config.Host,
		config.Port,
		config.User,
		config.Name,
		config.Namespace,
	)
}

// New creates one governed DB wrapper around the base SQL driver.
func (driver *pluginDataDriver) New(core *gdb.Core, node *gdb.ConfigNode) (gdb.DB, error) {
	baseDB, err := mysqlDriver.New().New(core, node)
	if err != nil {
		return nil, err
	}
	return &pluginDataDB{DB: baseDB}, nil
}

// DoCommit validates governed SQL access before delegating to the wrapped DB
// and records audit logs for success and failure paths.
func (db *pluginDataDB) DoCommit(ctx context.Context, in gdb.DoCommitInput) (out gdb.DoCommitOutput, err error) {
	metadata := AuditFromContext(ctx)
	if metadata != nil {
		if validateErr := validatePluginDataCommit(metadata, in); validateErr != nil {
			logger.Warningf(
				ctx,
				"plugin data service commit blocked plugin=%s table=%s method=%s type=%s transaction=%t err=%v",
				metadata.PluginID,
				metadata.Table,
				metadata.Method,
				in.Type,
				metadata.Transaction,
				validateErr,
			)
			return out, validateErr
		}
	}

	out, err = db.DB.DoCommit(ctx, in)
	if metadata != nil {
		if err != nil {
			logger.Warningf(
				ctx,
				"plugin data service commit failed plugin=%s table=%s method=%s type=%s transaction=%t err=%v",
				metadata.PluginID,
				metadata.Table,
				metadata.Method,
				in.Type,
				metadata.Transaction,
				err,
			)
		} else {
			logger.Infof(
				ctx,
				"plugin data service commit plugin=%s table=%s method=%s type=%s transaction=%t source=%s userId=%d",
				metadata.PluginID,
				metadata.Table,
				metadata.Method,
				in.Type,
				metadata.Transaction,
				metadata.ExecutionSource,
				metadata.UserID,
			)
		}
	}
	return out, err
}

// validatePluginDataCommit validates one SQL commit request against the current
// audit metadata and allowed host-service method set.
func validatePluginDataCommit(metadata *AuditMetadata, in gdb.DoCommitInput) error {
	if metadata == nil {
		return nil
	}
	if metadata.ResourceTable == "" {
		return gerror.New("plugin data service audit context is missing resourceTable")
	}

	switch in.Type {
	case gdb.SqlTypeBegin, gdb.SqlTypeTXCommit, gdb.SqlTypeTXRollback:
		if !metadata.Transaction {
			return gerror.Newf("plugin data service non-transaction method cannot execute transaction commit type: %s", in.Type)
		}
		return nil
	case gdb.SqlTypeQueryContext, gdb.SqlTypeStmtQueryContext, gdb.SqlTypeStmtQueryRowContext:
		return validatePluginDataCommitTable(metadata, in)
	case gdb.SqlTypeExecContext, gdb.SqlTypeStmtExecContext, gdb.SqlTypePrepareContext:
		if metadata.Method != bridgehostservice.HostServiceMethodDataCreate &&
			metadata.Method != bridgehostservice.HostServiceMethodDataUpdate &&
			metadata.Method != bridgehostservice.HostServiceMethodDataDelete &&
			metadata.Method != bridgehostservice.HostServiceMethodDataTransaction {
			return gerror.Newf("plugin data service method %s cannot execute mutation commit type %s", metadata.Method, in.Type)
		}
	}
	return validatePluginDataCommitTable(metadata, in)
}

// validatePluginDataCommitTable verifies that the SQL statement references the
// authorized host table recorded in the audit metadata.
func validatePluginDataCommitTable(metadata *AuditMetadata, in gdb.DoCommitInput) error {
	normalizedSQL := strings.ToLower(strings.TrimSpace(in.Sql))
	normalizedTable := strings.ToLower(strings.TrimSpace(metadata.ResourceTable))
	if normalizedSQL == "" || normalizedTable == "" {
		return nil
	}
	if !strings.Contains(normalizedSQL, normalizedTable) {
		return gerror.Newf("plugin data service SQL does not reference authorized table %s", metadata.ResourceTable)
	}
	return nil
}

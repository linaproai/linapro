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
	"lina-core/pkg/pluginbridge"
)

const pluginDataDriverTypePrefix = "plugin-data-"

type pluginDataDriver struct {
	baseType string
}

type pluginDataDB struct {
	gdb.DB
}

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

func registerPluginDataDrivers() {
	pluginDataDriverRegisterOnce.Do(func() {
		for _, baseType := range []string{"mysql", "mariadb", "tidb"} {
			if err := gdb.Register(pluginDataDriverTypePrefix+baseType, &pluginDataDriver{baseType: baseType}); err != nil {
				panic(gerror.Wrapf(err, "register plugin data driver failed baseType=%s", baseType))
			}
		}
	})
}

func pluginDataDriverType(baseType string) (string, error) {
	normalizedBaseType := strings.ToLower(strings.TrimSpace(baseType))
	switch normalizedBaseType {
	case "mysql", "mariadb", "tidb":
		return pluginDataDriverTypePrefix + normalizedBaseType, nil
	default:
		return "", gerror.Newf("plugin data service 暂不支持数据库类型: %s", baseType)
	}
}

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

func (driver *pluginDataDriver) New(core *gdb.Core, node *gdb.ConfigNode) (gdb.DB, error) {
	baseDB, err := mysqlDriver.New().New(core, node)
	if err != nil {
		return nil, err
	}
	return &pluginDataDB{DB: baseDB}, nil
}

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

func validatePluginDataCommit(metadata *AuditMetadata, in gdb.DoCommitInput) error {
	if metadata == nil {
		return nil
	}
	if metadata.ResourceTable == "" {
		return gerror.New("plugin data service 审计上下文缺少 resourceTable")
	}

	switch in.Type {
	case gdb.SqlTypeBegin, gdb.SqlTypeTXCommit, gdb.SqlTypeTXRollback:
		if !metadata.Transaction {
			return gerror.Newf("plugin data service 非事务方法不允许执行事务提交类型: %s", in.Type)
		}
		return nil
	case gdb.SqlTypeQueryContext, gdb.SqlTypeStmtQueryContext, gdb.SqlTypeStmtQueryRowContext:
		return validatePluginDataCommitTable(metadata, in)
	case gdb.SqlTypeExecContext, gdb.SqlTypeStmtExecContext, gdb.SqlTypePrepareContext:
		if metadata.Method != pluginbridge.HostServiceMethodDataCreate &&
			metadata.Method != pluginbridge.HostServiceMethodDataUpdate &&
			metadata.Method != pluginbridge.HostServiceMethodDataDelete &&
			metadata.Method != pluginbridge.HostServiceMethodDataTransaction {
			return gerror.Newf("plugin data service 方法 %s 不允许执行变更提交类型 %s", metadata.Method, in.Type)
		}
	}
	return validatePluginDataCommitTable(metadata, in)
}

func validatePluginDataCommitTable(metadata *AuditMetadata, in gdb.DoCommitInput) error {
	normalizedSQL := strings.ToLower(strings.TrimSpace(in.Sql))
	normalizedTable := strings.ToLower(strings.TrimSpace(metadata.ResourceTable))
	if normalizedSQL == "" || normalizedTable == "" {
		return nil
	}
	if !strings.Contains(normalizedSQL, normalizedTable) {
		return gerror.Newf("plugin data service SQL 未命中授权表 %s", metadata.ResourceTable)
	}
	return nil
}

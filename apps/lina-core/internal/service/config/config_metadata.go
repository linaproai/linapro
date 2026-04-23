// This file loads embedded delivery metadata that feeds OpenAPI documentation
// and the version-information page.

package config

import (
	"context"
	"io/fs"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/packed"
)

// metadataConfigPath is the embedded metadata asset consumed by version and
// OpenAPI metadata readers.
const metadataConfigPath = "manifest/config/metadata.yaml"

// MetadataConfig holds embedded delivery metadata rendered by host pages.
type MetadataConfig struct {
	Framework MetadataFrameworkInfo   `json:"framework"` // Framework contains top-level framework metadata used by host pages.
	OpenApi   OpenApiConfig           `json:"openapi"`   // OpenApi contains OpenAPI document metadata.
	Backend   []MetadataComponentInfo `json:"backend"`   // Backend contains backend component cards for the version page.
	Frontend  []MetadataComponentInfo `json:"frontend"`  // Frontend contains frontend component cards for the version page.
}

// MetadataFrameworkInfo holds framework-level metadata from metadata.yaml.
type MetadataFrameworkInfo struct {
	Name          string `json:"name" yaml:"name"`                   // Name is the framework display name.
	Version       string `json:"version" yaml:"version"`             // Version is the current framework version.
	Description   string `json:"description" yaml:"description"`     // Description is the framework summary.
	Homepage      string `json:"homepage" yaml:"homepage"`           // Homepage is the project website URL.
	RepositoryURL string `json:"repositoryUrl" yaml:"repositoryUrl"` // RepositoryURL is the framework source repository URL.
	License       string `json:"license" yaml:"license"`             // License is the open-source license label.
}

// MetadataComponentInfo holds one component entry from metadata.yaml.
type MetadataComponentInfo struct {
	Name        string `json:"name"`        // Name is the display name of the component.
	Version     string `json:"version"`     // Version is the configured display version or the auto placeholder.
	Url         string `json:"url"`         // Url is the component homepage.
	Description string `json:"description"` // Description is the short component summary.
}

// GetMetadata reads embedded delivery metadata from the packaged resource file.
func (s *serviceImpl) GetMetadata(ctx context.Context) *MetadataConfig {
	return cloneMetadataConfig(processStaticConfigCaches.metadata.load(func() *MetadataConfig {
		content, err := fs.ReadFile(packed.Files, metadataConfigPath)
		if err != nil {
			panic(gerror.Wrapf(err, "读取嵌入元数据配置 %s 失败", metadataConfigPath))
		}

		adapter, err := gcfg.NewAdapterContent(string(content))
		if err != nil {
			panic(gerror.Wrap(err, "解析嵌入元数据配置失败"))
		}

		cfg := &MetadataConfig{
			OpenApi: defaultOpenApiConfig(),
		}
		mustScanMetadataConfig(ctx, adapter, "framework", &cfg.Framework)
		mustScanMetadataConfig(ctx, adapter, "openapi", &cfg.OpenApi)
		mustScanMetadataConfig(ctx, adapter, "backend", &cfg.Backend)
		mustScanMetadataConfig(ctx, adapter, "frontend", &cfg.Frontend)
		return cfg
	}))
}

// mustScanMetadataConfig scans one embedded metadata section into the target
// object and panics on malformed metadata.
func mustScanMetadataConfig(ctx context.Context, adapter *gcfg.AdapterContent, key string, target any) {
	if target == nil {
		panic(gerror.Newf("元数据配置 %s 的扫描目标不能为空", key))
	}

	value, err := adapter.Get(ctx, key)
	if err != nil {
		panic(gerror.Wrapf(err, "读取嵌入元数据配置 %s 失败", key))
	}
	if value == nil {
		return
	}
	if err = gconv.Scan(value, target); err != nil {
		panic(gerror.Wrapf(err, "读取嵌入元数据配置 %s 失败", key))
	}
}

## 1. Manifest and Asset Discovery

- [x] 1.1 为源码插件清单新增`consumer.frontend`结构，支持`mount_path`、`index`和`spa_fallback`。
- [x] 1.2 校验消费端前端挂载路径，禁止根路径、非法路径和宿主保留前缀。
- [x] 1.3 扩展源码目录和嵌入式文件系统扫描，发现`frontend/consumer/`下的消费端前端资产。
- [x] 1.4 在插件清单快照中记录消费端前端资源数量。

## 2. Frontend Asset Hosting

- [x] 2.1 新增按插件 ID、版本和相对路径读取消费端前端资产的服务方法。
- [x] 2.2 新增稳定挂载路径解析，支持入口文件、静态资源和显式开启的`SPA`回退。
- [x] 2.3 为消费端`HTML`入口注入`base`，保证挂载路径下的相对资源加载正确。
- [x] 2.4 为插件前端资产响应统一设置`Content-Type`、`Cache-Control`和`ETag`，并支持`If-None-Match`返回`304`。

## 3. Cache and Lifecycle Governance

- [x] 3.1 新增消费端前端挂载索引缓存，缓存权威来源为清单、资产列表、插件版本和启用状态。
- [x] 3.2 在源码插件安装、卸载、启用、禁用和升级成功后失效消费端前端挂载索引。
- [x] 3.3 在插件运行时共享修订号刷新时联动清空消费端前端挂载索引，覆盖集群环境下的跨实例一致性。
- [x] 3.4 确保缓存读取返回克隆对象，避免调用方修改进程内缓存。

## 4. Governance Snapshot

- [x] 4.1 新增消费端前端治理快照，投影消费端前端挂载与插件治理元数据。
- [x] 4.2 快照包含插件 ID、版本、启用状态、租户治理声明、默认安装模式、挂载路径、入口、`SPA`回退和资产数量。

## 5. Verification

- [x] 5.1 运行`gofmt`覆盖所有新增和修改的`Go`文件。
- [x] 5.2 运行`cd apps/lina-core && $env:GF_GCFG_PATH=(Resolve-Path manifest/config).Path; $env:GF_GCFG_FILE='config.template.yaml'; go test ./internal/service/plugin/internal/catalog -run "TestDiscoverPluginVuePathsUseDirectoryConvention|TestBuildPluginManifestSnapshotIncludesDirectoryDiscoveredAssets|TestNormalizeConsumerSpecValidatesFrontendMountContract|TestScanEmbeddedSourcePluginManifestsUsesPluginEmbeddedFiles" -count=1`，结果通过。
- [x] 5.3 运行`cd apps/lina-core && $env:GF_GCFG_PATH=(Resolve-Path manifest/config).Path; $env:GF_GCFG_FILE='config.template.yaml'; go test ./internal/cmd -run "TestApplyPluginFrontendAssetHeadersEmitsValidators|TestRequestETagMatches|TestParseSourceConsumerPluginAssetRequestPath|TestParsePluginAssetRequestPath" -count=1`，结果通过。
- [x] 5.4 运行`cd apps/lina-core && $env:GF_GCFG_PATH=(Resolve-Path manifest/config).Path; $env:GF_GCFG_FILE='config.template.yaml'; go test ./internal/service/plugin -run "TestRewriteSourceConsumerHTMLBase|TestApplySourceConsumerMountAssetPolicyForcesRevalidation|TestCloneFrontendAssetOutputProtectsCachedBytes|TestLoadSourceConsumerFrontendMountEntriesCachesIndex|TestNormalizeSourceConsumerFrontendAssetPath|TestSourceConsumerFrontendAssetDeclared|TestMatchSourceConsumerFrontendMountPath|TestIsSourceConsumerFrontendMount|TestSourceConsumerFrontendResourceIndex|TestFindSourceConsumerFrontendOverlappingMount|TestSourceConsumerSPAFallbackEnabled|TestActiveSourceConsumerFrontendSpec|TestLooksLikeSourceConsumerStaticAsset|TestBuildConsumerSurfaceSnapshot" -count=1`，结果通过。
- [x] 5.5 静态扫描确认实现范围集中在消费端前端资产、挂载索引和治理快照。
- [x] 5.6 运行`git diff --check`，结果通过。
- [x] 5.7 运行`openspec validate plugin-consumer-frontend --strict`，结果通过。

## Review Notes

- i18n 影响：不新增或修改用户可见运行时文案、前端语言包、插件`manifest/i18n`资源或`apidoc i18n JSON`。
- 数据权限影响：不新增业务数据读取、写入、导出、详情或聚合接口；消费端前端资产托管不改变角色数据权限边界。
- 缓存一致性影响：新增进程内消费端前端挂载索引；单机模式通过生命周期本地失效，集群模式复用插件运行时共享修订号刷新进行跨实例观察和本地失效。
- 开发工具与脚本影响：不新增或修改默认开发工具、构建脚本、测试脚本或跨平台入口。

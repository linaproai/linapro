// This file defines media plugin business error codes.

package media

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeMediaTableCheckFailed reports that media table inspection failed.
	CodeMediaTableCheckFailed = bizerr.MustDefine("MEDIA_TABLE_CHECK_FAILED", "检测媒体数据表失败", gcode.CodeInternalError)
	// CodeMediaTableNotInstalled reports that plugin SQL has not been installed.
	CodeMediaTableNotInstalled = bizerr.MustDefine("MEDIA_TABLE_NOT_INSTALLED", "媒体数据表不存在，请先安装插件", gcode.CodeNotFound)
	// CodeMediaSwitchValueInvalid reports that a numeric switch value is invalid.
	CodeMediaSwitchValueInvalid = bizerr.MustDefine("MEDIA_SWITCH_VALUE_INVALID", "开关值只能是1或2", gcode.CodeInvalidParameter)
	// CodeMediaBinaryValueInvalid reports that a numeric binary value is invalid.
	CodeMediaBinaryValueInvalid = bizerr.MustDefine("MEDIA_BINARY_VALUE_INVALID", "是否标记只能是0或1", gcode.CodeInvalidParameter)
	// CodeMediaStrategyNameRequired reports that strategy name is missing.
	CodeMediaStrategyNameRequired = bizerr.MustDefine("MEDIA_STRATEGY_NAME_REQUIRED", "策略名称不能为空", gcode.CodeInvalidParameter)
	// CodeMediaStrategyContentRequired reports that strategy content is missing.
	CodeMediaStrategyContentRequired = bizerr.MustDefine("MEDIA_STRATEGY_CONTENT_REQUIRED", "策略内容不能为空", gcode.CodeInvalidParameter)
	// CodeMediaStrategyNotFound reports that a strategy does not exist.
	CodeMediaStrategyNotFound = bizerr.MustDefine("MEDIA_STRATEGY_NOT_FOUND", "媒体策略不存在", gcode.CodeNotFound)
	// CodeMediaStrategyReferenced reports that a strategy is referenced by bindings.
	CodeMediaStrategyReferenced = bizerr.MustDefine("MEDIA_STRATEGY_REFERENCED", "该媒体策略仍被绑定引用，不能删除", gcode.CodeInvalidOperation)
	// CodeMediaStrategyCountQueryFailed reports that strategy count query failed.
	CodeMediaStrategyCountQueryFailed = bizerr.MustDefine("MEDIA_STRATEGY_COUNT_QUERY_FAILED", "查询媒体策略总数失败", gcode.CodeInternalError)
	// CodeMediaStrategyListQueryFailed reports that strategy list query failed.
	CodeMediaStrategyListQueryFailed = bizerr.MustDefine("MEDIA_STRATEGY_LIST_QUERY_FAILED", "查询媒体策略列表失败", gcode.CodeInternalError)
	// CodeMediaStrategyDetailQueryFailed reports that strategy detail query failed.
	CodeMediaStrategyDetailQueryFailed = bizerr.MustDefine("MEDIA_STRATEGY_DETAIL_QUERY_FAILED", "查询媒体策略详情失败", gcode.CodeInternalError)
	// CodeMediaStrategyCreateFailed reports that strategy creation failed.
	CodeMediaStrategyCreateFailed = bizerr.MustDefine("MEDIA_STRATEGY_CREATE_FAILED", "创建媒体策略失败", gcode.CodeInternalError)
	// CodeMediaStrategyUpdateFailed reports that strategy update failed.
	CodeMediaStrategyUpdateFailed = bizerr.MustDefine("MEDIA_STRATEGY_UPDATE_FAILED", "更新媒体策略失败", gcode.CodeInternalError)
	// CodeMediaStrategyDeleteFailed reports that strategy deletion failed.
	CodeMediaStrategyDeleteFailed = bizerr.MustDefine("MEDIA_STRATEGY_DELETE_FAILED", "删除媒体策略失败", gcode.CodeInternalError)
	// CodeMediaBindingDeviceRequired reports that device ID is missing.
	CodeMediaBindingDeviceRequired = bizerr.MustDefine("MEDIA_BINDING_DEVICE_REQUIRED", "设备国标ID不能为空", gcode.CodeInvalidParameter)
	// CodeMediaBindingTenantRequired reports that tenant ID is missing.
	CodeMediaBindingTenantRequired = bizerr.MustDefine("MEDIA_BINDING_TENANT_REQUIRED", "租户ID不能为空", gcode.CodeInvalidParameter)
	// CodeMediaBindingCountQueryFailed reports that binding count query failed.
	CodeMediaBindingCountQueryFailed = bizerr.MustDefine("MEDIA_BINDING_COUNT_QUERY_FAILED", "查询媒体策略绑定总数失败", gcode.CodeInternalError)
	// CodeMediaBindingListQueryFailed reports that binding list query failed.
	CodeMediaBindingListQueryFailed = bizerr.MustDefine("MEDIA_BINDING_LIST_QUERY_FAILED", "查询媒体策略绑定列表失败", gcode.CodeInternalError)
	// CodeMediaBindingSaveFailed reports that binding save failed.
	CodeMediaBindingSaveFailed = bizerr.MustDefine("MEDIA_BINDING_SAVE_FAILED", "保存媒体策略绑定失败", gcode.CodeInternalError)
	// CodeMediaBindingDeleteFailed reports that binding deletion failed.
	CodeMediaBindingDeleteFailed = bizerr.MustDefine("MEDIA_BINDING_DELETE_FAILED", "删除媒体策略绑定失败", gcode.CodeInternalError)
	// CodeMediaTietaTokenRequired reports that token-based media authorization is missing a Tieta token.
	CodeMediaTietaTokenRequired = bizerr.MustDefine("MEDIA_TIETA_TOKEN_REQUIRED", "铁塔 token 不能为空", gcode.CodeNotAuthorized)
	// CodeMediaTietaTokenInvalid reports that Tieta rejected the provided token.
	CodeMediaTietaTokenInvalid = bizerr.MustDefine("MEDIA_TIETA_TOKEN_INVALID", "铁塔 token 无效：{message}", gcode.CodeNotAuthorized)
	// CodeMediaTietaBaseURLMissing reports that the Tieta base URL has not been configured.
	CodeMediaTietaBaseURLMissing = bizerr.MustDefine("MEDIA_TIETA_BASE_URL_MISSING", "铁塔系统 baseUrl 未配置", gcode.CodeInternalError)
	// CodeMediaTietaUserInfoFailed reports that the Tieta user-info request failed.
	CodeMediaTietaUserInfoFailed = bizerr.MustDefine("MEDIA_TIETA_USER_INFO_FAILED", "请求铁塔用户信息失败", gcode.CodeInternalError)
	// CodeMediaTietaUserInfoInvalid reports that the Tieta user-info response cannot be parsed.
	CodeMediaTietaUserInfoInvalid = bizerr.MustDefine("MEDIA_TIETA_USER_INFO_INVALID", "解析铁塔用户信息失败", gcode.CodeInternalError)
	// CodeMediaTietaTenantMismatch reports that the requested tenant does not match the Tieta token tenant.
	CodeMediaTietaTenantMismatch = bizerr.MustDefine("MEDIA_TIETA_TENANT_MISMATCH", "请求租户与铁塔 token 租户不一致", gcode.CodeNotAuthorized)
	// CodeMediaTietaTenantMissing reports that the Tieta token did not return a tenant ID.
	CodeMediaTietaTenantMissing = bizerr.MustDefine("MEDIA_TIETA_TENANT_MISSING", "铁塔 token 未返回租户ID", gcode.CodeNotAuthorized)
	// CodeMediaTietaDevicePermissionFailed reports that the Tieta device-permission request failed.
	CodeMediaTietaDevicePermissionFailed = bizerr.MustDefine("MEDIA_TIETA_DEVICE_PERMISSION_FAILED", "请求铁塔设备权限失败", gcode.CodeInternalError)
	// CodeMediaTietaDevicePermissionInvalid reports that the Tieta device-permission response cannot be parsed.
	CodeMediaTietaDevicePermissionInvalid = bizerr.MustDefine("MEDIA_TIETA_DEVICE_PERMISSION_INVALID", "解析铁塔设备权限失败", gcode.CodeInternalError)
	// CodeMediaTietaDevicePermissionDenied reports that Tieta explicitly denied device access.
	CodeMediaTietaDevicePermissionDenied = bizerr.MustDefine("MEDIA_TIETA_DEVICE_PERMISSION_DENIED", "铁塔设备权限校验失败：{message}", gcode.CodeNotAuthorized)
	// CodeMediaAliasRequired reports that stream alias is missing.
	CodeMediaAliasRequired = bizerr.MustDefine("MEDIA_ALIAS_REQUIRED", "流别名不能为空", gcode.CodeInvalidParameter)
	// CodeMediaStreamPathRequired reports that stream path is missing.
	CodeMediaStreamPathRequired = bizerr.MustDefine("MEDIA_STREAM_PATH_REQUIRED", "真实流路径不能为空", gcode.CodeInvalidParameter)
	// CodeMediaAliasNotFound reports that a stream alias does not exist.
	CodeMediaAliasNotFound = bizerr.MustDefine("MEDIA_ALIAS_NOT_FOUND", "流别名不存在", gcode.CodeNotFound)
	// CodeMediaAliasCountQueryFailed reports that alias count query failed.
	CodeMediaAliasCountQueryFailed = bizerr.MustDefine("MEDIA_ALIAS_COUNT_QUERY_FAILED", "查询流别名总数失败", gcode.CodeInternalError)
	// CodeMediaAliasListQueryFailed reports that alias list query failed.
	CodeMediaAliasListQueryFailed = bizerr.MustDefine("MEDIA_ALIAS_LIST_QUERY_FAILED", "查询流别名列表失败", gcode.CodeInternalError)
	// CodeMediaAliasDetailQueryFailed reports that alias detail query failed.
	CodeMediaAliasDetailQueryFailed = bizerr.MustDefine("MEDIA_ALIAS_DETAIL_QUERY_FAILED", "查询流别名详情失败", gcode.CodeInternalError)
	// CodeMediaAliasCreateFailed reports that alias creation failed.
	CodeMediaAliasCreateFailed = bizerr.MustDefine("MEDIA_ALIAS_CREATE_FAILED", "创建流别名失败", gcode.CodeInternalError)
	// CodeMediaAliasUpdateFailed reports that alias update failed.
	CodeMediaAliasUpdateFailed = bizerr.MustDefine("MEDIA_ALIAS_UPDATE_FAILED", "更新流别名失败", gcode.CodeInternalError)
	// CodeMediaAliasDeleteFailed reports that alias deletion failed.
	CodeMediaAliasDeleteFailed = bizerr.MustDefine("MEDIA_ALIAS_DELETE_FAILED", "删除流别名失败", gcode.CodeInternalError)
	// CodeMediaTenantWhiteTenantRequired reports that tenant whitelist tenant ID is missing.
	CodeMediaTenantWhiteTenantRequired = bizerr.MustDefine("MEDIA_TENANT_WHITE_TENANT_REQUIRED", "租户ID不能为空", gcode.CodeInvalidParameter)
	// CodeMediaTenantWhiteIPRequired reports that tenant whitelist IP is missing.
	CodeMediaTenantWhiteIPRequired = bizerr.MustDefine("MEDIA_TENANT_WHITE_IP_REQUIRED", "白名单地址不能为空", gcode.CodeInvalidParameter)
	// CodeMediaTenantWhiteIPInvalid reports that tenant whitelist IP is not a valid IPv4 or IPv6 address.
	CodeMediaTenantWhiteIPInvalid = bizerr.MustDefine("MEDIA_TENANT_WHITE_IP_INVALID", "白名单地址必须是有效的 IPv4 或 IPv6 地址", gcode.CodeInvalidParameter)
	// CodeMediaTenantWhiteDescriptionTooLong reports that tenant whitelist description is too long.
	CodeMediaTenantWhiteDescriptionTooLong = bizerr.MustDefine("MEDIA_TENANT_WHITE_DESCRIPTION_TOO_LONG", "白名单描述长度不能超过32个字符", gcode.CodeInvalidParameter)
	// CodeMediaTenantWhiteEnableInvalid reports that tenant whitelist enable value is invalid.
	CodeMediaTenantWhiteEnableInvalid = bizerr.MustDefine("MEDIA_TENANT_WHITE_ENABLE_INVALID", "租户白名单启用状态只能是0或1", gcode.CodeInvalidParameter)
	// CodeMediaTenantWhiteNotFound reports that a tenant whitelist entry does not exist.
	CodeMediaTenantWhiteNotFound = bizerr.MustDefine("MEDIA_TENANT_WHITE_NOT_FOUND", "租户白名单不存在", gcode.CodeNotFound)
	// CodeMediaTenantWhiteDuplicate reports that a tenant whitelist natural key already exists.
	CodeMediaTenantWhiteDuplicate = bizerr.MustDefine("MEDIA_TENANT_WHITE_DUPLICATE", "租户白名单已存在", gcode.CodeInvalidParameter)
	// CodeMediaTenantWhiteCountQueryFailed reports that tenant whitelist count query failed.
	CodeMediaTenantWhiteCountQueryFailed = bizerr.MustDefine("MEDIA_TENANT_WHITE_COUNT_QUERY_FAILED", "查询租户白名单总数失败", gcode.CodeInternalError)
	// CodeMediaTenantWhiteListQueryFailed reports that tenant whitelist list query failed.
	CodeMediaTenantWhiteListQueryFailed = bizerr.MustDefine("MEDIA_TENANT_WHITE_LIST_QUERY_FAILED", "查询租户白名单列表失败", gcode.CodeInternalError)
	// CodeMediaTenantWhiteDetailQueryFailed reports that tenant whitelist detail query failed.
	CodeMediaTenantWhiteDetailQueryFailed = bizerr.MustDefine("MEDIA_TENANT_WHITE_DETAIL_QUERY_FAILED", "查询租户白名单详情失败", gcode.CodeInternalError)
	// CodeMediaTenantWhiteCreateFailed reports that tenant whitelist creation failed.
	CodeMediaTenantWhiteCreateFailed = bizerr.MustDefine("MEDIA_TENANT_WHITE_CREATE_FAILED", "创建租户白名单失败", gcode.CodeInternalError)
	// CodeMediaTenantWhiteUpdateFailed reports that tenant whitelist update failed.
	CodeMediaTenantWhiteUpdateFailed = bizerr.MustDefine("MEDIA_TENANT_WHITE_UPDATE_FAILED", "更新租户白名单失败", gcode.CodeInternalError)
	// CodeMediaTenantWhiteDeleteFailed reports that tenant whitelist deletion failed.
	CodeMediaTenantWhiteDeleteFailed = bizerr.MustDefine("MEDIA_TENANT_WHITE_DELETE_FAILED", "删除租户白名单失败", gcode.CodeInternalError)
	// CodeMediaNodeNumInvalid reports that a media node number is invalid.
	CodeMediaNodeNumInvalid = bizerr.MustDefine("MEDIA_NODE_NUM_INVALID", "节点编号必须在0到255之间", gcode.CodeInvalidParameter)
	// CodeMediaNodeNameRequired reports that media node name is missing.
	CodeMediaNodeNameRequired = bizerr.MustDefine("MEDIA_NODE_NAME_REQUIRED", "节点名称不能为空", gcode.CodeInvalidParameter)
	// CodeMediaNodeNameTooLong reports that media node name is too long.
	CodeMediaNodeNameTooLong = bizerr.MustDefine("MEDIA_NODE_NAME_TOO_LONG", "节点名称长度不能超过32个字符", gcode.CodeInvalidParameter)
	// CodeMediaNodeURLRequired reports that media node gateway URL is missing.
	CodeMediaNodeURLRequired = bizerr.MustDefine("MEDIA_NODE_URL_REQUIRED", "节点网关地址不能为空", gcode.CodeInvalidParameter)
	// CodeMediaNodeURLTooLong reports that media node gateway URL is too long.
	CodeMediaNodeURLTooLong = bizerr.MustDefine("MEDIA_NODE_URL_TOO_LONG", "节点网关地址长度不能超过255个字符", gcode.CodeInvalidParameter)
	// CodeMediaNodeNotFound reports that a media node does not exist.
	CodeMediaNodeNotFound = bizerr.MustDefine("MEDIA_NODE_NOT_FOUND", "媒体节点不存在", gcode.CodeNotFound)
	// CodeMediaNodeDuplicate reports that a media node number already exists.
	CodeMediaNodeDuplicate = bizerr.MustDefine("MEDIA_NODE_DUPLICATE", "媒体节点编号已存在", gcode.CodeInvalidParameter)
	// CodeMediaNodeReferenced reports that a media node is referenced by config rows.
	CodeMediaNodeReferenced = bizerr.MustDefine("MEDIA_NODE_REFERENCED", "该媒体节点仍被设备节点或租户流配置引用，不能删除", gcode.CodeInvalidOperation)
	// CodeMediaNodeCountQueryFailed reports that media node count query failed.
	CodeMediaNodeCountQueryFailed = bizerr.MustDefine("MEDIA_NODE_COUNT_QUERY_FAILED", "查询媒体节点总数失败", gcode.CodeInternalError)
	// CodeMediaNodeListQueryFailed reports that media node list query failed.
	CodeMediaNodeListQueryFailed = bizerr.MustDefine("MEDIA_NODE_LIST_QUERY_FAILED", "查询媒体节点列表失败", gcode.CodeInternalError)
	// CodeMediaNodeDetailQueryFailed reports that media node detail query failed.
	CodeMediaNodeDetailQueryFailed = bizerr.MustDefine("MEDIA_NODE_DETAIL_QUERY_FAILED", "查询媒体节点详情失败", gcode.CodeInternalError)
	// CodeMediaNodeCreateFailed reports that media node creation failed.
	CodeMediaNodeCreateFailed = bizerr.MustDefine("MEDIA_NODE_CREATE_FAILED", "创建媒体节点失败", gcode.CodeInternalError)
	// CodeMediaNodeUpdateFailed reports that media node update failed.
	CodeMediaNodeUpdateFailed = bizerr.MustDefine("MEDIA_NODE_UPDATE_FAILED", "更新媒体节点失败", gcode.CodeInternalError)
	// CodeMediaNodeDeleteFailed reports that media node deletion failed.
	CodeMediaNodeDeleteFailed = bizerr.MustDefine("MEDIA_NODE_DELETE_FAILED", "删除媒体节点失败", gcode.CodeInternalError)
	// CodeMediaDeviceNodeDeviceRequired reports that device-node device ID is missing.
	CodeMediaDeviceNodeDeviceRequired = bizerr.MustDefine("MEDIA_DEVICE_NODE_DEVICE_REQUIRED", "设备国标ID不能为空", gcode.CodeInvalidParameter)
	// CodeMediaDeviceNodeDeviceTooLong reports that device-node device ID is too long.
	CodeMediaDeviceNodeDeviceTooLong = bizerr.MustDefine("MEDIA_DEVICE_NODE_DEVICE_TOO_LONG", "设备国标ID长度不能超过64个字符", gcode.CodeInvalidParameter)
	// CodeMediaDeviceNodeNotFound reports that a device-node mapping does not exist.
	CodeMediaDeviceNodeNotFound = bizerr.MustDefine("MEDIA_DEVICE_NODE_NOT_FOUND", "设备节点不存在", gcode.CodeNotFound)
	// CodeMediaDeviceNodeDuplicate reports that a device-node mapping already exists.
	CodeMediaDeviceNodeDuplicate = bizerr.MustDefine("MEDIA_DEVICE_NODE_DUPLICATE", "设备节点已存在", gcode.CodeInvalidParameter)
	// CodeMediaDeviceNodeCountQueryFailed reports that device-node count query failed.
	CodeMediaDeviceNodeCountQueryFailed = bizerr.MustDefine("MEDIA_DEVICE_NODE_COUNT_QUERY_FAILED", "查询设备节点总数失败", gcode.CodeInternalError)
	// CodeMediaDeviceNodeListQueryFailed reports that device-node list query failed.
	CodeMediaDeviceNodeListQueryFailed = bizerr.MustDefine("MEDIA_DEVICE_NODE_LIST_QUERY_FAILED", "查询设备节点列表失败", gcode.CodeInternalError)
	// CodeMediaDeviceNodeDetailQueryFailed reports that device-node detail query failed.
	CodeMediaDeviceNodeDetailQueryFailed = bizerr.MustDefine("MEDIA_DEVICE_NODE_DETAIL_QUERY_FAILED", "查询设备节点详情失败", gcode.CodeInternalError)
	// CodeMediaDeviceNodeCreateFailed reports that device-node creation failed.
	CodeMediaDeviceNodeCreateFailed = bizerr.MustDefine("MEDIA_DEVICE_NODE_CREATE_FAILED", "创建设备节点失败", gcode.CodeInternalError)
	// CodeMediaDeviceNodeUpdateFailed reports that device-node update failed.
	CodeMediaDeviceNodeUpdateFailed = bizerr.MustDefine("MEDIA_DEVICE_NODE_UPDATE_FAILED", "更新设备节点失败", gcode.CodeInternalError)
	// CodeMediaDeviceNodeDeleteFailed reports that device-node deletion failed.
	CodeMediaDeviceNodeDeleteFailed = bizerr.MustDefine("MEDIA_DEVICE_NODE_DELETE_FAILED", "删除设备节点失败", gcode.CodeInternalError)
	// CodeMediaTenantStreamTenantRequired reports that tenant stream config tenant ID is missing.
	CodeMediaTenantStreamTenantRequired = bizerr.MustDefine("MEDIA_TENANT_STREAM_TENANT_REQUIRED", "租户ID不能为空", gcode.CodeInvalidParameter)
	// CodeMediaTenantStreamTenantTooLong reports that tenant stream config tenant ID is too long.
	CodeMediaTenantStreamTenantTooLong = bizerr.MustDefine("MEDIA_TENANT_STREAM_TENANT_TOO_LONG", "租户ID长度不能超过64个字符", gcode.CodeInvalidParameter)
	// CodeMediaTenantStreamMaxConcurrentInvalid reports that tenant stream max concurrency is invalid.
	CodeMediaTenantStreamMaxConcurrentInvalid = bizerr.MustDefine("MEDIA_TENANT_STREAM_MAX_CONCURRENT_INVALID", "最大并发数不能小于0", gcode.CodeInvalidParameter)
	// CodeMediaTenantStreamEnableInvalid reports that tenant stream enable value is invalid.
	CodeMediaTenantStreamEnableInvalid = bizerr.MustDefine("MEDIA_TENANT_STREAM_ENABLE_INVALID", "租户流配置启用状态只能是0或1", gcode.CodeInvalidParameter)
	// CodeMediaTenantStreamNotFound reports that a tenant stream config does not exist.
	CodeMediaTenantStreamNotFound = bizerr.MustDefine("MEDIA_TENANT_STREAM_NOT_FOUND", "租户流配置不存在", gcode.CodeNotFound)
	// CodeMediaTenantStreamDuplicate reports that a tenant stream config already exists.
	CodeMediaTenantStreamDuplicate = bizerr.MustDefine("MEDIA_TENANT_STREAM_DUPLICATE", "租户流配置已存在", gcode.CodeInvalidParameter)
	// CodeMediaTenantStreamCountQueryFailed reports that tenant stream config count query failed.
	CodeMediaTenantStreamCountQueryFailed = bizerr.MustDefine("MEDIA_TENANT_STREAM_COUNT_QUERY_FAILED", "查询租户流配置总数失败", gcode.CodeInternalError)
	// CodeMediaTenantStreamListQueryFailed reports that tenant stream config list query failed.
	CodeMediaTenantStreamListQueryFailed = bizerr.MustDefine("MEDIA_TENANT_STREAM_LIST_QUERY_FAILED", "查询租户流配置列表失败", gcode.CodeInternalError)
	// CodeMediaTenantStreamDetailQueryFailed reports that tenant stream config detail query failed.
	CodeMediaTenantStreamDetailQueryFailed = bizerr.MustDefine("MEDIA_TENANT_STREAM_DETAIL_QUERY_FAILED", "查询租户流配置详情失败", gcode.CodeInternalError)
	// CodeMediaTenantStreamCreateFailed reports that tenant stream config creation failed.
	CodeMediaTenantStreamCreateFailed = bizerr.MustDefine("MEDIA_TENANT_STREAM_CREATE_FAILED", "创建租户流配置失败", gcode.CodeInternalError)
	// CodeMediaTenantStreamUpdateFailed reports that tenant stream config update failed.
	CodeMediaTenantStreamUpdateFailed = bizerr.MustDefine("MEDIA_TENANT_STREAM_UPDATE_FAILED", "更新租户流配置失败", gcode.CodeInternalError)
	// CodeMediaTenantStreamDeleteFailed reports that tenant stream config deletion failed.
	CodeMediaTenantStreamDeleteFailed = bizerr.MustDefine("MEDIA_TENANT_STREAM_DELETE_FAILED", "删除租户流配置失败", gcode.CodeInternalError)
)

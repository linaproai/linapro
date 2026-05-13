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
)

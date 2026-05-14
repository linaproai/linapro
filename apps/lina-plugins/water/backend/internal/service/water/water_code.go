// This file defines water plugin business error codes.

package water

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeWaterMediaTableCheckFailed reports that media table inspection failed.
	CodeWaterMediaTableCheckFailed = bizerr.MustDefine("WATER_MEDIA_TABLE_CHECK_FAILED", "检测媒体策略表失败", gcode.CodeInternalError)
	// CodeWaterMediaTableNotInstalled reports that required media tables are missing.
	CodeWaterMediaTableNotInstalled = bizerr.MustDefine("WATER_MEDIA_TABLE_NOT_INSTALLED", "媒体策略表不存在，请先安装媒体插件", gcode.CodeNotFound)
	// CodeWaterTenantRequired reports that the media tenant ID is missing.
	CodeWaterTenantRequired = bizerr.MustDefine("WATER_TENANT_REQUIRED", "媒体租户ID不能为空", gcode.CodeInvalidParameter)
	// CodeWaterImageRequired reports that the input image is missing.
	CodeWaterImageRequired = bizerr.MustDefine("WATER_IMAGE_REQUIRED", "图片不能为空", gcode.CodeInvalidParameter)
	// CodeWaterTaskNotFound reports that a task ID was not found in recent status storage.
	CodeWaterTaskNotFound = bizerr.MustDefine("WATER_TASK_NOT_FOUND", "水印任务不存在或已过期", gcode.CodeNotFound)
	// CodeWaterTaskQueueFull reports that the in-memory task queue is full.
	CodeWaterTaskQueueFull = bizerr.MustDefine("WATER_TASK_QUEUE_FULL", "水印任务队列已满，请稍后再试", gcode.CodeInvalidOperation)
	// CodeWaterStrategyQueryFailed reports that media strategy query failed.
	CodeWaterStrategyQueryFailed = bizerr.MustDefine("WATER_STRATEGY_QUERY_FAILED", "查询媒体策略失败", gcode.CodeInternalError)
	// CodeWaterStrategyParseFailed reports that the matched strategy YAML cannot be parsed.
	CodeWaterStrategyParseFailed = bizerr.MustDefine("WATER_STRATEGY_PARSE_FAILED", "解析水印策略失败", gcode.CodeInvalidParameter)
	// CodeWaterImageBase64Invalid reports that image base64 text is invalid.
	CodeWaterImageBase64Invalid = bizerr.MustDefine("WATER_IMAGE_BASE64_INVALID", "图片base64格式无效", gcode.CodeInvalidParameter)
	// CodeWaterImageDecodeFailed reports that the image bytes cannot be decoded.
	CodeWaterImageDecodeFailed = bizerr.MustDefine("WATER_IMAGE_DECODE_FAILED", "解码图片失败", gcode.CodeInvalidParameter)
	// CodeWaterImageEncodeFailed reports that the output image cannot be encoded.
	CodeWaterImageEncodeFailed = bizerr.MustDefine("WATER_IMAGE_ENCODE_FAILED", "编码水印图片失败", gcode.CodeInternalError)
	// CodeWaterImageDrawFailed reports that watermark drawing failed.
	CodeWaterImageDrawFailed = bizerr.MustDefine("WATER_IMAGE_DRAW_FAILED", "绘制水印失败", gcode.CodeInternalError)
	// CodeWaterCallbackURLInvalid reports that the callback URL is invalid.
	CodeWaterCallbackURLInvalid = bizerr.MustDefine("WATER_CALLBACK_URL_INVALID", "回调地址无效", gcode.CodeInvalidParameter)
	// CodeWaterCallbackFailed reports that callback delivery failed.
	CodeWaterCallbackFailed = bizerr.MustDefine("WATER_CALLBACK_FAILED", "发送水印回调失败", gcode.CodeInternalError)
)

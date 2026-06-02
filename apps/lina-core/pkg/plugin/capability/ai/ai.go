// Package ai defines the host AI capability namespace exposed through the
// plugin capability directory. The package only aggregates typed AI sub
// capabilities; each sub capability keeps its own DTOs, status, fallback, and
// provider contract.
package ai

import "lina-core/pkg/plugin/capability/ai/aitext"

// Service aggregates typed AI sub capabilities under one stable namespace.
//
// Service 聚合宿主发布的 AI 子能力，适用于源码插件、动态插件和宿主模块通过统一入口访问文本等能力，同时避免在根能力目录继续追加 AI 子能力方法。
type Service interface {
	// Text returns the text AI capability service.
	//
	// Text 返回文本 AI 子能力服务；未配置 provider 时也必须返回可降级服务，由子能力自身返回结构化不可用错误。
	Text() aitext.Service
}

// serviceImpl stores typed AI sub capability services.
type serviceImpl struct {
	text aitext.Service
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// New creates an AI namespace service from explicit sub capability services.
func New(text aitext.Service) Service {
	if text == nil {
		text = aitext.New(nil)
	}
	return &serviceImpl{text: text}
}

// Text returns the text AI capability service.
func (s *serviceImpl) Text() aitext.Service {
	if s == nil || s.text == nil {
		return aitext.New(nil)
	}
	return s.text
}

// ForPlugin returns a plugin-scoped AI namespace service while preserving the
// runtime-owned AI sub capability implementations. The scoped namespace binds
// pluginID to downstream AI provider requests through each sub capability, so
// source plugins and dynamic plugins can consume host AI services without
// manually supplying or spoofing the caller identity.
//
// This host-injected source identity is important for AI invocation audit,
// usage attribution, troubleshooting, and future plugin-level governance such
// as quota, rate limit, tier access, or purpose policy decisions. When service
// is nil, the returned namespace still exposes fallback sub capabilities so
// callers receive structured unavailable errors instead of nil services.
//
// ForPlugin 返回绑定插件身份的 AI 命名空间服务，同时保留宿主运行期持有的 AI 子能力实现。
// 该方法会把 pluginID 通过各个子能力注入到后续 provider 请求中，使源码插件和动态插件可以消费宿主
// AI 能力，而不需要也不能由调用方手动填写或伪造来源身份。
//
// 宿主可信注入的来源身份用于 AI 调用审计、用量归因、问题定位，以及后续插件级配额、限流、
// 档位访问和 purpose 策略等治理能力。service 为空时仍返回带 fallback 子能力的命名空间，
// 确保调用方获得结构化不可用错误，而不是 nil service。
func ForPlugin(service Service, pluginID string) Service {
	if service == nil {
		return New(aitext.ForPlugin(nil, pluginID))
	}
	return New(aitext.ForPlugin(service.Text(), pluginID))
}

package demo

import (
	"context"
)

const summaryMessage = "这是一条来自 plugin-demo-source 接口的简要介绍，用于验证源码插件菜单页可读取插件后端数据。"

// SummaryOutput defines one concise plugin summary payload.
type SummaryOutput struct {
	// Message is the concise page introduction returned from the plugin API.
	Message string
}

// Summary returns one concise plugin summary payload.
func (s *Service) Summary(ctx context.Context) (out *SummaryOutput, err error) {
	return &SummaryOutput{
		Message: summaryMessage,
	}, nil
}

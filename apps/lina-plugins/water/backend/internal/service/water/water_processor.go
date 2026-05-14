// This file implements snapshot processing orchestration.

package water

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// processSnapshot resolves strategy and returns one processed image result.
func processSnapshot(ctx context.Context, in SubmitSnapInput) (*ProcessOutput, error) {
	start := time.Now()
	deviceID := strings.TrimSpace(in.DeviceId)
	if deviceID == "" {
		deviceID = strings.TrimSpace(in.DeviceCode)
	}

	strategy, err := ResolveStrategy(ctx, in.Tenant, deviceID)
	if err != nil {
		return nil, err
	}
	inputImage, err := decodeImageDataURL(in.Image)
	if err != nil {
		return nil, err
	}
	if strategy == nil || !strategy.Matched {
		return skippedProcessOutput(inputImage, StrategySourceNone, strategySourceLabel(StrategySourceNone), start)
	}

	cfg, err := parseWatermarkStrategy(strategy.Strategy)
	if err != nil {
		return nil, err
	}
	if cfg == nil || !cfg.Enabled {
		return skippedProcessOutputWithStrategy(inputImage, strategy, start)
	}

	output, err := drawWatermark(inputImage, *cfg)
	if err != nil {
		return nil, err
	}
	return &ProcessOutput{
		Success:      true,
		Status:       TaskStatusSuccess,
		Message:      "处理完成",
		Image:        encodePNGDataURL(output),
		StrategyId:   strategy.StrategyId,
		StrategyName: strategy.StrategyName,
		Source:       strategy.Source,
		SourceLabel:  strategy.SourceLabel,
		DurationMs:   time.Since(start).Milliseconds(),
	}, nil
}

// skippedProcessOutput returns the original image when no strategy matched.
func skippedProcessOutput(input []byte, source StrategySource, sourceLabel string, start time.Time) (*ProcessOutput, error) {
	dataURL, err := ensurePNGDataURL(input)
	if err != nil {
		return nil, err
	}
	return &ProcessOutput{
		Success:     true,
		Status:      TaskStatusSkipped,
		Message:     "未匹配到启用水印策略，已跳过",
		Image:       dataURL,
		Source:      source,
		SourceLabel: sourceLabel,
		DurationMs:  time.Since(start).Milliseconds(),
	}, nil
}

// skippedProcessOutputWithStrategy returns the original image when a strategy has no enabled watermark config.
func skippedProcessOutputWithStrategy(input []byte, strategy *resolvedStrategy, start time.Time) (*ProcessOutput, error) {
	dataURL, err := ensurePNGDataURL(input)
	if err != nil {
		return nil, err
	}
	return &ProcessOutput{
		Success:      true,
		Status:       TaskStatusSkipped,
		Message:      "策略未启用水印，已跳过",
		Image:        dataURL,
		StrategyId:   strategy.StrategyId,
		StrategyName: strategy.StrategyName,
		Source:       strategy.Source,
		SourceLabel:  strategy.SourceLabel,
		DurationMs:   time.Since(start).Milliseconds(),
	}, nil
}

// buildCallbackPayload builds the hotgo-compatible callback payload.
func buildCallbackPayload(in SubmitSnapInput, image string) snapPayload {
	return snapPayload{
		ErrorCode:   in.ErrorCode,
		DeviceCode:  in.DeviceCode,
		ChannelCode: in.ChannelCode,
		DeviceIdx:   in.DeviceIdx,
		Image:       image,
		ImageName:   in.ImageName,
		ImagePath:   in.ImagePath,
		AccessNode:  in.AccessNode,
		AcceptNode:  in.AcceptNode,
		UploadUrl:   in.UploadUrl,
	}
}

// normalizedCallbackURL returns the non-empty callback URL, accepting hotgo's legacy url field.
func normalizedCallbackURL(in SubmitSnapInput) string {
	if strings.TrimSpace(in.CallbackUrl) != "" {
		return strings.TrimSpace(in.CallbackUrl)
	}
	return strings.TrimSpace(in.Url)
}

// sendResultToURL sends one callback payload as JSON.
func sendResultToURL(ctx context.Context, callbackURL string, payload snapPayload) error {
	parsedURL, err := url.ParseRequestURI(callbackURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return bizerr.WrapCode(err, CodeWaterCallbackURLInvalid)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return bizerr.WrapCode(err, CodeWaterCallbackFailed)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(body))
	if err != nil {
		return bizerr.WrapCode(err, CodeWaterCallbackFailed)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "linapro-water/1.0")

	client := &http.Client{Timeout: defaultCallbackTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return bizerr.WrapCode(err, CodeWaterCallbackFailed)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warningf(ctx, "关闭水印回调响应体失败: %v", closeErr)
		}
	}()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return bizerr.NewCode(CodeWaterCallbackFailed)
	}
	return nil
}

// This file implements the public water service methods.

package water

import (
	"context"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
)

// SubmitSnap submits one asynchronous watermark snapshot task.
func (s *serviceImpl) SubmitSnap(ctx context.Context, in SubmitSnapInput) (*SubmitSnapOutput, error) {
	normalized, err := normalizeSubmitSnapInput(in)
	if err != nil {
		return nil, err
	}
	taskID := generateTaskID()
	s.store.create(taskID, normalized)
	err = s.queue.submit(ctx, &watermarkTask{
		id:      taskID,
		ctx:     ctx,
		request: normalized,
	})
	if err != nil {
		s.store.update(taskID, func(record *taskRecord) {
			record.Status = TaskStatusFailed
			record.Success = false
			record.Message = "提交失败"
			record.Error = err.Error()
		})
		return nil, err
	}
	return &SubmitSnapOutput{
		Success: true,
		TaskId:  taskID,
		Status:  TaskStatusQueued,
	}, nil
}

// Preview synchronously renders one watermark preview.
func (s *serviceImpl) Preview(ctx context.Context, in PreviewInput) (*ProcessOutput, error) {
	normalized, err := normalizePreviewInput(in)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	output, err := processSnapshot(ctx, SubmitSnapInput{
		Tenant:      normalized.Tenant,
		DeviceId:    normalized.DeviceId,
		DeviceCode:  normalized.DeviceCode,
		ChannelCode: normalized.ChannelCode,
		Image:       normalized.Image,
	})
	if err != nil {
		return nil, err
	}
	output.DurationMs = time.Since(start).Milliseconds()
	return output, nil
}

// GetTask returns one recent watermark task snapshot.
func (s *serviceImpl) GetTask(ctx context.Context, taskID string) (*TaskSnapshot, error) {
	normalized := strings.TrimSpace(taskID)
	if normalized == "" {
		return nil, bizerr.NewCode(CodeWaterTaskNotFound)
	}
	return s.store.get(ctx, normalized)
}

// normalizeSubmitSnapInput validates and normalizes asynchronous task input.
func normalizeSubmitSnapInput(in SubmitSnapInput) (SubmitSnapInput, error) {
	in.Tenant = strings.TrimSpace(in.Tenant)
	in.DeviceId = strings.TrimSpace(in.DeviceId)
	in.DeviceType = strings.TrimSpace(in.DeviceType)
	in.DeviceCode = strings.TrimSpace(in.DeviceCode)
	in.ChannelCode = strings.TrimSpace(in.ChannelCode)
	in.Image = strings.TrimSpace(in.Image)
	in.CallbackUrl = strings.TrimSpace(in.CallbackUrl)
	in.Url = strings.TrimSpace(in.Url)
	if in.DeviceCode == "" {
		in.DeviceCode = in.DeviceId
	}
	if in.Tenant == "" {
		return in, bizerr.NewCode(CodeWaterTenantRequired)
	}
	if in.Image == "" {
		return in, bizerr.NewCode(CodeWaterImageRequired)
	}
	return in, nil
}

// normalizePreviewInput validates and normalizes synchronous preview input.
func normalizePreviewInput(in PreviewInput) (PreviewInput, error) {
	in.Tenant = strings.TrimSpace(in.Tenant)
	in.DeviceId = strings.TrimSpace(in.DeviceId)
	in.DeviceCode = strings.TrimSpace(in.DeviceCode)
	in.ChannelCode = strings.TrimSpace(in.ChannelCode)
	in.Image = strings.TrimSpace(in.Image)
	if in.DeviceCode == "" {
		in.DeviceCode = in.DeviceId
	}
	if in.Tenant == "" {
		return in, bizerr.NewCode(CodeWaterTenantRequired)
	}
	if in.Image == "" {
		return in, bizerr.NewCode(CodeWaterImageRequired)
	}
	return in, nil
}

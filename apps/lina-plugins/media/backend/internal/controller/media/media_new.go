// This file constructs the media plugin controller.

package media

import (
	"lina-plugin-media/backend/api/media"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ControllerV1 is the media plugin controller.
type ControllerV1 struct {
	mediaSvc mediasvc.Service // mediaSvc handles plugin business operations.
}

// NewV1 creates and returns a new media controller instance.
func NewV1() media.IMediaV1 {
	return &ControllerV1{mediaSvc: mediasvc.New()}
}

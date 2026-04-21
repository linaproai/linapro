// Package notice exposes the host notice controller constructor to source
// plugins without granting direct access to internal packages.
package notice

import (
	noticeapi "lina-core/api/notice"
	internalnotice "lina-core/internal/controller/notice"
)

// NewV1 creates and returns the host notice controller for source-plugin route binding.
func NewV1() noticeapi.INoticeV1 {
	return internalnotice.NewV1()
}

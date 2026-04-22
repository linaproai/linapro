// This file exposes the plugin-local generated notice entity through the
// service package so existing controller code can keep a stable type name.

package notice

import entitymodel "lina-plugin-content-notice/backend/internal/model/entity"

// NoticeEntity mirrors the generated plugin_content_notice entity owned by this plugin.
type NoticeEntity = entitymodel.Notice

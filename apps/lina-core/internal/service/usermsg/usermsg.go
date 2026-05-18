// Package usermsg exposes current-user inbox message read and mutation
// contracts backed by the unified notification service.
package usermsg

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/bizctx"
	notifysvc "lina-core/internal/service/notify"
)

// Stable i18n key convention used to localize inbox category labels and tag
// colors. The host does not enumerate specific category codes here; senders
// (host services or source plugins) register translations at
// `notify.category.{code}.label` and `notify.category.{code}.color` in their
// own manifest/i18n bundles, and the host i18n resource aggregator merges
// them at runtime. This keeps the inbox UI category-agnostic.
const (
	// usermsgCategoryI18nNamespace is the parent i18n namespace shared by all category labels and colors.
	usermsgCategoryI18nNamespace = "notify.category."
	// usermsgCategoryLabelI18nSuffix is the i18n key suffix that resolves the category display label.
	usermsgCategoryLabelI18nSuffix = ".label"
	// usermsgCategoryColorI18nSuffix is the i18n key suffix that resolves the category tag color.
	usermsgCategoryColorI18nSuffix = ".color"
	// usermsgCategoryFallbackCode is the canonical fallback category used when a message has no declared category code.
	usermsgCategoryFallbackCode = "other"
	// usermsgCategoryDefaultColor is the safety color rendered when no category color resource is configured.
	usermsgCategoryDefaultColor = "default"
	// usermsgCategoryDefaultLabel is the safety label rendered when no category label resource is configured.
	usermsgCategoryDefaultLabel = "Notification"
)

// Service defines the usermsg service contract.
type Service interface {
	// Get returns one current-user message detail for preview consumption.
	// The lookup is constrained to the authenticated user in context and
	// delegates ownership checks to the notification inbox service.
	Get(ctx context.Context, id int64) (*MessageDetail, error)
	// UnreadCount returns unread message count for the authenticated user.
	// Missing authentication returns a usermsg business error; notification
	// backend errors are propagated.
	UnreadCount(ctx context.Context) (int, error)
	// List queries messages for the authenticated user with pagination. Category
	// labels and colors are resolved through runtime i18n resources with stable
	// fallbacks; notification backend errors are propagated.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// MarkRead marks a single message as read for the authenticated user only.
	// Ownership and visibility are enforced by the notification inbox service.
	MarkRead(ctx context.Context, id int64) error
	// MarkReadAll marks all messages as read for the authenticated user only.
	// The operation is idempotent for already-read messages.
	MarkReadAll(ctx context.Context) error
	// Delete deletes a single message for the authenticated user only. Ownership
	// and visibility are enforced before mutation by the notification service.
	Delete(ctx context.Context, id int64) error
	// Clear deletes all messages for the authenticated user only. Missing
	// authentication returns a usermsg business error.
	Clear(ctx context.Context) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service
	notifySvc notifysvc.Service     // Unified notify service
	i18nSvc   usermsgI18nTranslator // Host i18n service for category label/color localization
}

// usermsgI18nTranslator defines the narrow translation capability usermsg needs.
type usermsgI18nTranslator interface {
	// Translate returns one runtime translation key with caller-provided fallback text.
	Translate(ctx context.Context, key string, fallback string) string
}

// New creates a usermsg service from explicit runtime-owned dependencies.
func New(bizCtxSvc bizctx.Service, notifySvc notifysvc.Service, i18nSvc usermsgI18nTranslator) Service {
	return &serviceImpl{
		bizCtxSvc: bizCtxSvc,
		notifySvc: notifySvc,
		i18nSvc:   i18nSvc,
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum  int // Page number, starting from 1
	PageSize int // Page size
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*MessageItem // Message list
	Total int            // Total count
}

// MessageItem defines one user message facade item.
type MessageItem struct {
	Id           int64       // Message ID
	UserId       int64       // Recipient user ID
	Title        string      // Message title
	CategoryCode string      // Opaque sender-declared category code identifying the notification kind
	TypeLabel    string      // Localized category label resolved at the host
	TypeColor    string      // Localized category tag color resolved at the host
	SourceType   string      // Message source type
	SourceId     int64       // Message source ID
	IsRead       int         // Whether the message has been read
	ReadAt       *gtime.Time // Read time
	CreatedAt    *gtime.Time // Creation time
}

// MessageDetail defines one current-user message detail payload used by the
// inbox preview dialog.
type MessageDetail struct {
	Id            int64       // Message ID
	Title         string      // Message title
	CategoryCode  string      // Opaque sender-declared category code identifying the notification kind
	TypeLabel     string      // Localized category label resolved at the host
	TypeColor     string      // Localized category tag color resolved at the host
	SourceType    string      // Message source type
	SourceId      int64       // Message source ID
	Content       string      // Renderable message body content
	CreatedByName string      // Sender display name
	CreatedAt     *gtime.Time // Message creation time
}

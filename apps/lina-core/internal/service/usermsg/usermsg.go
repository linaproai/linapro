package usermsg

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/bizctx"
	i18nsvc "lina-core/internal/service/i18n"
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
	Get(ctx context.Context, id int64) (*MessageDetail, error)
	// UnreadCount returns unread message count for current user.
	UnreadCount(ctx context.Context) (int, error)
	// List queries message list for current user with pagination.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// MarkRead marks a single message as read.
	MarkRead(ctx context.Context, id int64) error
	// MarkReadAll marks all messages as read for current user.
	MarkReadAll(ctx context.Context) error
	// Delete deletes a single message for current user.
	Delete(ctx context.Context, id int64) error
	// Clear deletes all messages for current user.
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

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{
		bizCtxSvc: bizctx.New(),
		notifySvc: notifysvc.New(),
		i18nSvc:   i18nsvc.New(),
	}
}

// resolveCategoryCode normalizes an inbox message category code, falling back
// to the canonical "other" bucket when the sender did not declare one.
func resolveCategoryCode(categoryCode string) string {
	if categoryCode == "" {
		return usermsgCategoryFallbackCode
	}
	return categoryCode
}

// localizeCategoryLabel resolves the localized category display label for the
// given category code. Translation is looked up at
// `notify.category.{code}.label`. If the requested code has no translation
// resource, it falls back to the canonical "other" bucket and finally to a
// safety literal so the inbox never renders an empty category cell.
func (s *serviceImpl) localizeCategoryLabel(ctx context.Context, categoryCode string) string {
	if s == nil || s.i18nSvc == nil {
		return usermsgCategoryDefaultLabel
	}
	code := resolveCategoryCode(categoryCode)
	key := usermsgCategoryI18nNamespace + code + usermsgCategoryLabelI18nSuffix
	if label := s.i18nSvc.Translate(ctx, key, ""); label != "" {
		return label
	}
	if code != usermsgCategoryFallbackCode {
		fallbackKey := usermsgCategoryI18nNamespace + usermsgCategoryFallbackCode + usermsgCategoryLabelI18nSuffix
		if label := s.i18nSvc.Translate(ctx, fallbackKey, ""); label != "" {
			return label
		}
	}
	return usermsgCategoryDefaultLabel
}

// localizeCategoryColor resolves the localized category tag color for the
// given category code. Color is treated as a localizable display attribute so
// senders can override their preferred palette per locale if needed; the
// resolution path mirrors localizeCategoryLabel and falls back to a generic
// neutral color.
func (s *serviceImpl) localizeCategoryColor(ctx context.Context, categoryCode string) string {
	if s == nil || s.i18nSvc == nil {
		return usermsgCategoryDefaultColor
	}
	code := resolveCategoryCode(categoryCode)
	key := usermsgCategoryI18nNamespace + code + usermsgCategoryColorI18nSuffix
	if color := s.i18nSvc.Translate(ctx, key, ""); color != "" {
		return color
	}
	if code != usermsgCategoryFallbackCode {
		fallbackKey := usermsgCategoryI18nNamespace + usermsgCategoryFallbackCode + usermsgCategoryColorI18nSuffix
		if color := s.i18nSvc.Translate(ctx, fallbackKey, ""); color != "" {
			return color
		}
	}
	return usermsgCategoryDefaultColor
}

// getCurrentUserId extracts current user ID from context.
func (s *serviceImpl) getCurrentUserId(ctx context.Context) (int64, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil || bizCtx.UserId == 0 {
		return 0, gerror.New("未登录")
	}
	return int64(bizCtx.UserId), nil
}

// UnreadCount returns unread message count for current user.
func (s *serviceImpl) UnreadCount(ctx context.Context) (int, error) {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	return s.notifySvc.InboxUnreadCount(ctx, userId)
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

// List queries message list for current user with pagination.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return nil, err
	}

	out, err := s.notifySvc.InboxList(ctx, notifysvc.InboxListInput{
		UserID:   userId,
		PageNum:  in.PageNum,
		PageSize: in.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*MessageItem, 0, len(out.List))
	for _, item := range out.List {
		if item == nil {
			continue
		}
		categoryCode := resolveCategoryCode(item.CategoryCode)
		items = append(items, &MessageItem{
			Id:           item.Id,
			UserId:       item.UserID,
			Title:        item.Title,
			CategoryCode: categoryCode,
			TypeLabel:    s.localizeCategoryLabel(ctx, categoryCode),
			TypeColor:    s.localizeCategoryColor(ctx, categoryCode),
			SourceType:   item.SourceType,
			SourceId:     item.SourceID,
			IsRead:       item.IsRead,
			ReadAt:       item.ReadAt,
			CreatedAt:    item.CreatedAt,
		})
	}

	return &ListOutput{List: items, Total: out.Total}, nil
}

// MarkRead marks a single message as read.
func (s *serviceImpl) MarkRead(ctx context.Context, id int64) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxMarkRead(ctx, userId, id)
}

// MarkReadAll marks all messages as read for current user.
func (s *serviceImpl) MarkReadAll(ctx context.Context) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxMarkAllRead(ctx, userId)
}

// Delete deletes a single message for current user.
func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxDelete(ctx, userId, id)
}

// Clear deletes all messages for current user.
func (s *serviceImpl) Clear(ctx context.Context) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxClear(ctx, userId)
}

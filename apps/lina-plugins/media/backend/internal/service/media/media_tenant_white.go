// This file implements tenant whitelist CRUD for the media plugin.

package media

import (
	"context"
	"net"
	"strings"
	"unicode/utf8"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
	entitymodel "lina-plugin-media/backend/internal/model/entity"
)

// ListTenantWhitesInput defines tenant whitelist list filters.
type ListTenantWhitesInput struct {
	PageNum  int    // PageNum is the requested page number.
	PageSize int    // PageSize is the requested page size.
	Keyword  string // Keyword fuzzy-matches tenant ID, IP, or description.
	Enable   *int   // Enable filters whitelist status when set.
}

// ListTenantWhitesOutput defines paged tenant whitelist entries.
type ListTenantWhitesOutput struct {
	List  []*TenantWhiteOutput // List contains the current page.
	Total int                  // Total is the total matched row count.
}

// TenantWhiteOutput defines one tenant whitelist response.
type TenantWhiteOutput struct {
	TenantId    string // TenantId is the media tenant ID.
	Ip          string // Ip is the whitelist address.
	Description string // Description is the whitelist description.
	Enable      int    // Enable marks whether the whitelist entry is active.
	CreatorId   int    // CreatorId is the creator user ID.
	CreateTime  string // CreateTime is the formatted creation time.
	UpdaterId   int    // UpdaterId is the last updater user ID.
	UpdateTime  string // UpdateTime is the formatted update time.
}

// TenantWhiteMutationInput defines tenant whitelist create/update input.
type TenantWhiteMutationInput struct {
	TenantId    string // TenantId is the media tenant ID.
	Ip          string // Ip is the whitelist address.
	Description string // Description is the whitelist description.
	Enable      int    // Enable marks whether the whitelist entry is active.
}

// TenantWhiteMutationOutput defines tenant whitelist mutation result.
type TenantWhiteMutationOutput struct {
	TenantId string // TenantId is the media tenant ID.
	Ip       string // Ip is the whitelist address.
}

// tenantWhiteEntity reuses the plugin-local generated tenant whitelist entity.
type tenantWhiteEntity = entitymodel.MediaTenantWhite

// ListTenantWhites returns paged tenant whitelist entries.
func (s *serviceImpl) ListTenantWhites(ctx context.Context, in ListTenantWhitesInput) (*ListTenantWhitesOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.MediaTenantWhite.Columns()
	model := dao.MediaTenantWhite.Ctx(ctx)

	keyword := strings.TrimSpace(in.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.TenantId+" LIKE ? OR "+columns.Ip+" LIKE ? OR "+columns.Description+" LIKE ?)",
			likeKeyword,
			likeKeyword,
			likeKeyword,
		)
	}
	if in.Enable != nil {
		enable, err := normalizeWhiteEnableValue(*in.Enable, WhiteEnabled)
		if err != nil {
			return nil, err
		}
		model = model.Where(columns.Enable, enable)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantWhiteCountQueryFailed)
	}

	items := make([]*tenantWhiteEntity, 0)
	err = model.
		OrderAsc(columns.TenantId).
		OrderAsc(columns.Ip).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantWhiteListQueryFailed)
	}

	list := make([]*TenantWhiteOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildTenantWhiteOutput(item))
	}
	return &ListTenantWhitesOutput{List: list, Total: total}, nil
}

// GetTenantWhite returns one tenant whitelist entry by natural key.
func (s *serviceImpl) GetTenantWhite(ctx context.Context, tenantID string, ip string) (*TenantWhiteOutput, error) {
	record, err := s.getTenantWhiteEntity(ctx, tenantID, ip)
	if err != nil {
		return nil, err
	}
	return buildTenantWhiteOutput(record), nil
}

// CreateTenantWhite creates one tenant whitelist entry.
func (s *serviceImpl) CreateTenantWhite(ctx context.Context, in TenantWhiteMutationInput) (*TenantWhiteMutationOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalized, err := normalizeTenantWhiteMutationInput(in)
	if err != nil {
		return nil, err
	}
	exists, err := s.tenantWhiteExists(ctx, normalized.TenantId, normalized.Ip)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, bizerr.NewCode(CodeMediaTenantWhiteDuplicate)
	}

	actorID := int(s.currentActorID(ctx))
	now := gtime.Now()
	_, err = dao.MediaTenantWhite.Ctx(ctx).Data(do.MediaTenantWhite{
		TenantId:    normalized.TenantId,
		Ip:          normalized.Ip,
		Description: normalized.Description,
		Enable:      normalized.Enable,
		CreatorId:   actorID,
		CreateTime:  now,
		UpdaterId:   actorID,
		UpdateTime:  now,
	}).Insert()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantWhiteCreateFailed)
	}
	return &TenantWhiteMutationOutput{TenantId: normalized.TenantId, Ip: normalized.Ip}, nil
}

// UpdateTenantWhite updates one tenant whitelist entry.
func (s *serviceImpl) UpdateTenantWhite(ctx context.Context, tenantID string, ip string, in TenantWhiteMutationInput) (*TenantWhiteMutationOutput, error) {
	normalizedTenantID, normalizedIP, err := normalizeTenantWhiteKey(tenantID, ip)
	if err != nil {
		return nil, err
	}
	if _, err = s.getTenantWhiteEntity(ctx, normalizedTenantID, normalizedIP); err != nil {
		return nil, err
	}
	normalized, err := normalizeTenantWhiteMutationInput(in)
	if err != nil {
		return nil, err
	}
	if normalized.TenantId != normalizedTenantID || normalized.Ip != normalizedIP {
		exists, err := s.tenantWhiteExists(ctx, normalized.TenantId, normalized.Ip)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, bizerr.NewCode(CodeMediaTenantWhiteDuplicate)
		}
	}

	_, err = dao.MediaTenantWhite.Ctx(ctx).
		Where(do.MediaTenantWhite{TenantId: normalizedTenantID, Ip: normalizedIP}).
		Data(do.MediaTenantWhite{
			TenantId:    normalized.TenantId,
			Ip:          normalized.Ip,
			Description: normalized.Description,
			Enable:      normalized.Enable,
			UpdaterId:   int(s.currentActorID(ctx)),
			UpdateTime:  gtime.Now(),
		}).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantWhiteUpdateFailed)
	}
	return &TenantWhiteMutationOutput{TenantId: normalized.TenantId, Ip: normalized.Ip}, nil
}

// DeleteTenantWhite deletes one tenant whitelist entry.
func (s *serviceImpl) DeleteTenantWhite(ctx context.Context, tenantID string, ip string) (*TenantWhiteMutationOutput, error) {
	normalizedTenantID, normalizedIP, err := normalizeTenantWhiteKey(tenantID, ip)
	if err != nil {
		return nil, err
	}
	if _, err = s.getTenantWhiteEntity(ctx, normalizedTenantID, normalizedIP); err != nil {
		return nil, err
	}

	_, err = dao.MediaTenantWhite.Ctx(ctx).
		Where(do.MediaTenantWhite{TenantId: normalizedTenantID, Ip: normalizedIP}).
		Delete()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantWhiteDeleteFailed)
	}
	return &TenantWhiteMutationOutput{TenantId: normalizedTenantID, Ip: normalizedIP}, nil
}

// tenantWhiteExists reports whether one tenant whitelist natural key exists.
func (s *serviceImpl) tenantWhiteExists(ctx context.Context, tenantID string, ip string) (bool, error) {
	count, err := dao.MediaTenantWhite.Ctx(ctx).
		Where(do.MediaTenantWhite{TenantId: tenantID, Ip: ip}).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaTenantWhiteDetailQueryFailed)
	}
	return count > 0, nil
}

// getTenantWhiteEntity returns one tenant whitelist entity by natural key.
func (s *serviceImpl) getTenantWhiteEntity(ctx context.Context, tenantID string, ip string) (*tenantWhiteEntity, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalizedTenantID, normalizedIP, err := normalizeTenantWhiteKey(tenantID, ip)
	if err != nil {
		return nil, err
	}

	var record *tenantWhiteEntity
	err = dao.MediaTenantWhite.Ctx(ctx).
		Where(do.MediaTenantWhite{TenantId: normalizedTenantID, Ip: normalizedIP}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTenantWhiteDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeMediaTenantWhiteNotFound)
	}
	return record, nil
}

// normalizeTenantWhiteMutationInput validates tenant whitelist mutation input.
func normalizeTenantWhiteMutationInput(in TenantWhiteMutationInput) (TenantWhiteMutationInput, error) {
	tenantID, ip, err := normalizeTenantWhiteKey(in.TenantId, in.Ip)
	if err != nil {
		return TenantWhiteMutationInput{}, err
	}
	description := strings.TrimSpace(in.Description)
	if utf8.RuneCountInString(description) > 32 {
		return TenantWhiteMutationInput{}, bizerr.NewCode(CodeMediaTenantWhiteDescriptionTooLong)
	}
	enable, err := normalizeWhiteEnableValue(in.Enable, WhiteEnabled)
	if err != nil {
		return TenantWhiteMutationInput{}, err
	}
	return TenantWhiteMutationInput{
		TenantId:    tenantID,
		Ip:          ip,
		Description: description,
		Enable:      enable,
	}, nil
}

// normalizeTenantWhiteKey validates and trims the tenant whitelist natural key.
func normalizeTenantWhiteKey(tenantID string, ip string) (string, string, error) {
	normalizedTenantID := strings.TrimSpace(tenantID)
	if normalizedTenantID == "" {
		return "", "", bizerr.NewCode(CodeMediaTenantWhiteTenantRequired)
	}
	normalizedIP := strings.TrimSpace(ip)
	if normalizedIP == "" {
		return "", "", bizerr.NewCode(CodeMediaTenantWhiteIPRequired)
	}
	if net.ParseIP(normalizedIP) == nil {
		return "", "", bizerr.NewCode(CodeMediaTenantWhiteIPInvalid)
	}
	return normalizedTenantID, normalizedIP, nil
}

// normalizeWhiteEnableValue validates and normalizes tenant whitelist enable value.
func normalizeWhiteEnableValue(value int, defaultValue WhiteEnableValue) (int, error) {
	if value == 0 && defaultValue == WhiteEnabled {
		return int(WhiteDisabled), nil
	}
	switch WhiteEnableValue(value) {
	case WhiteEnabled, WhiteDisabled:
		return value, nil
	default:
		return 0, bizerr.NewCode(CodeMediaTenantWhiteEnableInvalid)
	}
}

// buildTenantWhiteOutput converts one generated tenant whitelist entity into service output.
func buildTenantWhiteOutput(item *tenantWhiteEntity) *TenantWhiteOutput {
	if item == nil {
		return &TenantWhiteOutput{}
	}
	return &TenantWhiteOutput{
		TenantId:    item.TenantId,
		Ip:          item.Ip,
		Description: item.Description,
		Enable:      item.Enable,
		CreatorId:   item.CreatorId,
		CreateTime:  formatTime(item.CreateTime),
		UpdaterId:   item.UpdaterId,
		UpdateTime:  formatTime(item.UpdateTime),
	}
}

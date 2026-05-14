// This file implements media stream alias CRUD.

package media

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
	"lina-plugin-media/backend/internal/model/do"
	entitymodel "lina-plugin-media/backend/internal/model/entity"
)

// ListAliasesInput defines stream alias list filters.
type ListAliasesInput struct {
	PageNum  int    // PageNum is the requested page number.
	PageSize int    // PageSize is the requested page size.
	Keyword  string // Keyword fuzzy-matches alias or stream path.
}

// ListAliasesOutput defines paged stream aliases.
type ListAliasesOutput struct {
	List  []*AliasOutput // List contains the current page.
	Total int            // Total is the total matched row count.
}

// AliasOutput defines one stream alias response.
type AliasOutput struct {
	Id         int64  // Id is the alias primary key.
	Alias      string // Alias is the stream alias.
	AutoRemove int    // AutoRemove marks whether the alias should be automatically removed.
	StreamPath string // StreamPath is the real stream path.
	CreateTime string // CreateTime is the formatted creation time.
}

// AliasMutationInput defines stream alias create/update input.
type AliasMutationInput struct {
	Alias      string // Alias is the stream alias.
	AutoRemove int    // AutoRemove marks whether the alias should be automatically removed.
	StreamPath string // StreamPath is the real stream path.
}

// aliasEntity reuses the plugin-local generated stream alias entity.
type aliasEntity = entitymodel.MediaStreamAlias

// ListAliases returns paged stream aliases.
func (s *serviceImpl) ListAliases(ctx context.Context, in ListAliasesInput) (*ListAliasesOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	columns := dao.MediaStreamAlias.Columns()
	model := dao.MediaStreamAlias.Ctx(ctx)

	keyword := strings.TrimSpace(in.Keyword)
	if keyword != "" {
		likeKeyword := "%" + keyword + "%"
		model = model.Where(
			"("+columns.Alias+" LIKE ? OR "+columns.StreamPath+" LIKE ?)",
			likeKeyword,
			likeKeyword,
		)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaAliasCountQueryFailed)
	}

	items := make([]*aliasEntity, 0)
	err = model.
		OrderDesc(columns.CreateTime).
		OrderDesc(columns.Id).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaAliasListQueryFailed)
	}

	list := make([]*AliasOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildAliasOutput(item))
	}
	return &ListAliasesOutput{List: list, Total: total}, nil
}

// GetAlias returns one stream alias by ID.
func (s *serviceImpl) GetAlias(ctx context.Context, id int64) (*AliasOutput, error) {
	record, err := s.getAliasEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	return buildAliasOutput(record), nil
}

// CreateAlias creates one stream alias.
func (s *serviceImpl) CreateAlias(ctx context.Context, in AliasMutationInput) (int64, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return 0, err
	}
	normalized, err := normalizeAliasMutationInput(in)
	if err != nil {
		return 0, err
	}

	id, err := dao.MediaStreamAlias.Ctx(ctx).Data(do.MediaStreamAlias{
		Alias:      normalized.Alias,
		AutoRemove: normalized.AutoRemove,
		StreamPath: normalized.StreamPath,
	}).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeMediaAliasCreateFailed)
	}
	return id, nil
}

// UpdateAlias updates one stream alias.
func (s *serviceImpl) UpdateAlias(ctx context.Context, id int64, in AliasMutationInput) error {
	if _, err := s.getAliasEntity(ctx, id); err != nil {
		return err
	}
	normalized, err := normalizeAliasMutationInput(in)
	if err != nil {
		return err
	}

	_, err = dao.MediaStreamAlias.Ctx(ctx).
		Where(do.MediaStreamAlias{Id: id}).
		Data(do.MediaStreamAlias{
			Alias:      normalized.Alias,
			AutoRemove: normalized.AutoRemove,
			StreamPath: normalized.StreamPath,
		}).
		Update()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaAliasUpdateFailed)
	}
	return nil
}

// DeleteAlias deletes one stream alias.
func (s *serviceImpl) DeleteAlias(ctx context.Context, id int64) error {
	if _, err := s.getAliasEntity(ctx, id); err != nil {
		return err
	}
	_, err := dao.MediaStreamAlias.Ctx(ctx).
		Where(do.MediaStreamAlias{Id: id}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaAliasDeleteFailed)
	}
	return nil
}

// getAliasEntity returns one stream alias entity by ID.
func (s *serviceImpl) getAliasEntity(ctx context.Context, id int64) (*aliasEntity, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	var record *aliasEntity
	err := dao.MediaStreamAlias.Ctx(ctx).
		Where(do.MediaStreamAlias{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaAliasDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeMediaAliasNotFound)
	}
	return record, nil
}

// normalizeAliasMutationInput validates stream alias mutation input.
func normalizeAliasMutationInput(in AliasMutationInput) (AliasMutationInput, error) {
	alias := strings.TrimSpace(in.Alias)
	if alias == "" {
		return AliasMutationInput{}, bizerr.NewCode(CodeMediaAliasRequired)
	}
	streamPath := strings.TrimSpace(in.StreamPath)
	if streamPath == "" {
		return AliasMutationInput{}, bizerr.NewCode(CodeMediaStreamPathRequired)
	}
	autoRemove, err := normalizeBinaryValue(in.AutoRemove, BinaryNo)
	if err != nil {
		return AliasMutationInput{}, err
	}
	return AliasMutationInput{
		Alias:      alias,
		AutoRemove: autoRemove,
		StreamPath: streamPath,
	}, nil
}

// buildAliasOutput converts one generated alias entity into service output.
func buildAliasOutput(item *aliasEntity) *AliasOutput {
	if item == nil {
		return &AliasOutput{}
	}
	return &AliasOutput{
		Id:         item.Id,
		Alias:      item.Alias,
		AutoRemove: item.AutoRemove,
		StreamPath: item.StreamPath,
		CreateTime: formatTime(item.CreateTime),
	}
}

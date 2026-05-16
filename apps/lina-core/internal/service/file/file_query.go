// This file contains file metadata list, detail, dictionary lookup, and
// data-scope-aware mutation operations.

package file

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	dictsvc "lina-core/internal/service/dict"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/gdbutil"
	"lina-core/pkg/logger"
)

// List returns paginated file records.
func (s *serviceImpl) List(ctx context.Context, in *ListInput) (*ListOutput, error) {
	m := dao.SysFile.Ctx(ctx)
	if in.Name != "" {
		m = m.WhereLike(dao.SysFile.Columns().Name, fmt.Sprintf("%%%s%%", in.Name))
	}
	if in.Original != "" {
		m = m.WhereLike(dao.SysFile.Columns().Original, fmt.Sprintf("%%%s%%", in.Original))
	}
	if in.Suffix != "" {
		m = m.Where(dao.SysFile.Columns().Suffix, in.Suffix)
	}
	if in.BeginTime != "" {
		m = m.WhereGTE(dao.SysFile.Columns().CreatedAt, in.BeginTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.SysFile.Columns().CreatedAt, in.EndTime)
	}
	if in.Scene != "" {
		m = m.Where(dao.SysFile.Columns().Scene, in.Scene)
	}
	var err error
	m, err = s.applyFileDataScope(ctx, m)
	if err != nil {
		return nil, err
	}
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	cols := dao.SysFile.Columns()
	var (
		orderBy           = cols.Id
		allowedSortFields = map[string]string{
			"size":      cols.Size,
			"createdAt": cols.CreatedAt,
		}
		direction = gdbutil.NormalizeOrderDirectionOrDefault(in.OrderDirection, gdbutil.OrderDirectionDESC)
	)
	if in.OrderBy != "" {
		if field, ok := allowedSortFields[in.OrderBy]; ok {
			orderBy = field
		}
	}

	var files []*entity.SysFile
	err = gdbutil.ApplyModelOrder(m.Page(in.PageNum, in.PageSize), orderBy, direction).Scan(&files)
	if err != nil {
		return nil, err
	}

	userNameMap := s.loadCreatorNames(ctx, files)
	baseUrl := s.getBaseUrl(ctx)
	items := make([]*ListOutputItem, len(files))
	for i, f := range files {
		fileCopy := *f
		if fileCopy.Url != "" && baseUrl != "" {
			fileCopy.Url = baseUrl + fileCopy.Url
		}
		items[i] = &ListOutputItem{
			SysFile:       &fileCopy,
			CreatedByName: userNameMap[f.CreatedBy],
		}
	}
	return &ListOutput{List: items, Total: total}, nil
}

// loadCreatorNames resolves uploader usernames for list output enrichment.
func (s *serviceImpl) loadCreatorNames(ctx context.Context, files []*entity.SysFile) map[int64]string {
	userIdMap := make(map[int64]bool)
	for _, f := range files {
		if f.CreatedBy > 0 {
			userIdMap[f.CreatedBy] = true
		}
	}
	userNameMap := make(map[int64]string)
	if len(userIdMap) == 0 {
		return userNameMap
	}
	userIds := make([]int64, 0, len(userIdMap))
	for uid := range userIdMap {
		userIds = append(userIds, uid)
	}
	var users []*entity.SysUser
	err := dao.SysUser.Ctx(ctx).WhereIn(dao.SysUser.Columns().Id, userIds).Scan(&users)
	if err == nil {
		for _, u := range users {
			userNameMap[int64(u.Id)] = u.Username
		}
	}
	return userNameMap
}

// Info returns file info by ID.
func (s *serviceImpl) Info(ctx context.Context, id int64) (*entity.SysFile, error) {
	if err := s.ensureFilesVisible(ctx, []int64{id}); err != nil {
		return nil, err
	}
	var file *entity.SysFile
	err := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Id, id).Scan(&file)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileRecordQueryFailed)
	}
	if file == nil {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	return file, nil
}

// InfoByIds returns file info by multiple IDs.
func (s *serviceImpl) InfoByIds(ctx context.Context, ids []int64) ([]*entity.SysFile, error) {
	if err := s.ensureFilesVisible(ctx, ids); err != nil {
		return nil, err
	}
	var files []*entity.SysFile
	model := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, ids)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&files)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeFileRecordQueryFailed)
	}
	baseUrl := s.getBaseUrl(ctx)
	if baseUrl != "" {
		for _, f := range files {
			if f.Url != "" {
				f.Url = baseUrl + f.Url
			}
		}
	}
	return files, nil
}

// Delete removes files by IDs (soft delete in DB, also removes physical files).
func (s *serviceImpl) Delete(ctx context.Context, idsStr string) error {
	ids := gstr.SplitAndTrim(idsStr, ",")
	if len(ids) == 0 {
		return bizerr.NewCode(CodeFileDeleteRequired)
	}
	idList := make([]int64, 0, len(ids))
	for _, idStr := range ids {
		idList = append(idList, gconv.Int64(idStr))
	}
	if err := s.ensureFilesVisible(ctx, idList); err != nil {
		return err
	}
	var files []*entity.SysFile
	model := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, idList)
	model = datascope.ApplyTenantScope(ctx, model, datascope.TenantColumn)
	err := model.Scan(&files)
	if err != nil {
		return err
	}
	deleteModel := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, idList)
	deleteModel = datascope.ApplyTenantScope(ctx, deleteModel, datascope.TenantColumn)
	if _, err = deleteModel.Delete(); err != nil {
		return err
	}
	for _, f := range files {
		if deleteErr := s.storage.Delete(ctx, f.Path); deleteErr != nil {
			logger.Warningf(ctx, "delete storage file failed path=%s err=%v", f.Path, deleteErr)
		}
	}
	return nil
}

// getBaseUrl returns the base URL (scheme + host) from the current HTTP request context.
func (s *serviceImpl) getBaseUrl(ctx context.Context) string {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return ""
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// UsageScenes returns all usage scenes from dictionary.
func (s *serviceImpl) UsageScenes(ctx context.Context) ([]*UsageScenesOutput, error) {
	list, err := s.dictSvc.DataByType(ctx, DictTypeFileScene)
	if err != nil {
		return nil, err
	}
	items := make([]*UsageScenesOutput, 0, len(list))
	for _, item := range list {
		items = append(items, &UsageScenesOutput{
			Value: item.Value,
			Label: item.Label,
		})
	}
	return items, nil
}

// Suffixes returns distinct file suffixes from the database.
func (s *serviceImpl) Suffixes(ctx context.Context) ([]*SuffixesOutput, error) {
	model, err := s.applyFileDataScope(ctx, dao.SysFile.Ctx(ctx))
	if err != nil {
		return nil, err
	}
	result, err := model.Fields(dao.SysFile.Columns().Suffix).
		Group(dao.SysFile.Columns().Suffix).
		OrderAsc(dao.SysFile.Columns().Suffix).
		Array()
	if err != nil {
		return nil, err
	}
	items := make([]*SuffixesOutput, 0, len(result))
	for _, v := range result {
		suffix := v.String()
		if suffix == "" {
			continue
		}
		items = append(items, &SuffixesOutput{
			Value: suffix,
			Label: suffix,
		})
	}
	return items, nil
}

// Detail returns file info with scene label.
func (s *serviceImpl) Detail(ctx context.Context, id int64) (*DetailOutput, error) {
	if err := s.ensureFilesVisible(ctx, []int64{id}); err != nil {
		return nil, err
	}
	var file *entity.SysFile
	err := dao.SysFile.Ctx(ctx).Where(dao.SysFile.Columns().Id, id).Scan(&file)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, bizerr.NewCode(CodeFileNotFound)
	}
	baseUrl := s.getBaseUrl(ctx)
	if baseUrl != "" && file.Url != "" {
		file.Url = baseUrl + file.Url
	}

	var createdByName string
	if file.CreatedBy > 0 {
		var user *entity.SysUser
		err = dao.SysUser.Ctx(ctx).Where(dao.SysUser.Columns().Id, file.CreatedBy).Scan(&user)
		if err == nil && user != nil {
			createdByName = user.Username
		}
	}
	sceneLabel := s.dictSvc.GetLabelByValue(ctx, dictsvc.GetLabelByValueInput{
		DictType: DictTypeFileScene,
		Value:    file.Scene,
	})
	return &DetailOutput{
		SysFile:       file,
		CreatedByName: createdByName,
		SceneLabel:    sceneLabel,
	}, nil
}

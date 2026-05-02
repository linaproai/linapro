// This file implements scheduled-job group CRUD operations.

package jobmgmt

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/gdbutil"
)

// ListGroups returns scheduled-job groups with pagination and job counts.
func (s *serviceImpl) ListGroups(ctx context.Context, in ListGroupsInput) (*ListGroupsOutput, error) {
	model := dao.SysJobGroup.Ctx(ctx)
	cols := dao.SysJobGroup.Columns()
	if keyword := strings.TrimSpace(in.Code); keyword != "" {
		model = model.WhereLike(cols.Code, "%"+keyword+"%")
	}
	if keyword := strings.TrimSpace(in.Name); keyword != "" {
		model = model.WhereLike(cols.Name, "%"+keyword+"%")
		if s.defaultGroupMatchesKeyword(ctx, keyword) {
			model = model.WhereOr(cols.Code, defaultBuiltinGroupCode)
		}
	}

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	var groups []*entity.SysJobGroup
	err = applySingleOrder(
		model,
		in.OrderBy,
		in.OrderDirection,
		map[orderField]string{
			orderFieldID:        cols.Id,
			orderFieldSortOrder: cols.SortOrder,
			orderFieldCode:      cols.Code,
			orderFieldName:      cols.Name,
			orderFieldCreatedAt: cols.CreatedAt,
			orderFieldUpdatedAt: cols.UpdatedAt,
		},
		cols.SortOrder,
		gdbutil.OrderDirectionASC,
	).Page(in.PageNum, in.PageSize).Scan(&groups)
	if err != nil {
		return nil, err
	}

	jobCounts, err := jobCountMapByGroupIDs(ctx, groups)
	if err != nil {
		return nil, err
	}

	items := make([]*GroupListItem, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		s.localizeGroupForDisplay(ctx, group)
		items = append(items, &GroupListItem{
			SysJobGroup: group,
			JobCount:    jobCounts[group.Id],
		})
	}
	return &ListGroupsOutput{List: items, Total: total}, nil
}

// groupJobCountRow stores one aggregated job count grouped by job-group ID.
type groupJobCountRow struct {
	GroupID uint64 `orm:"group_id"`  // GroupID is the owning scheduled-job group ID.
	Count   int64  `orm:"job_count"` // Count is the number of jobs in the group.
}

// jobCountMapByGroupIDs loads job counts for all listed groups in one grouped query.
func jobCountMapByGroupIDs(ctx context.Context, groups []*entity.SysJobGroup) (map[uint64]int64, error) {
	groupIDs := make([]uint64, 0, len(groups))
	for _, group := range groups {
		if group == nil || group.Id == 0 {
			continue
		}
		groupIDs = append(groupIDs, group.Id)
	}
	if len(groupIDs) == 0 {
		return map[uint64]int64{}, nil
	}

	var rows []*groupJobCountRow
	jobCols := dao.SysJob.Columns()
	err := dao.SysJob.Ctx(ctx).
		Fields(jobCols.GroupId, "COUNT(1) AS job_count").
		WhereIn(jobCols.GroupId, groupIDs).
		Group(jobCols.GroupId).
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	counts := make(map[uint64]int64, len(groupIDs))
	for _, row := range rows {
		if row == nil {
			continue
		}
		counts[row.GroupID] = row.Count
	}
	return counts, nil
}

// CreateGroup persists one new scheduled-job group.
func (s *serviceImpl) CreateGroup(ctx context.Context, in SaveGroupInput) (uint64, error) {
	code := strings.TrimSpace(in.Code)
	name := strings.TrimSpace(in.Name)
	if code == "" {
		return 0, bizerr.NewCode(CodeJobGroupCodeRequired)
	}
	if name == "" {
		return 0, bizerr.NewCode(CodeJobGroupNameRequired)
	}

	count, err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{Code: code}).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, bizerr.NewCode(CodeJobGroupCodeExists)
	}

	insertID, err := dao.SysJobGroup.Ctx(ctx).Data(do.SysJobGroup{
		Code:      code,
		Name:      name,
		Remark:    strings.TrimSpace(in.Remark),
		SortOrder: in.SortOrder,
		IsDefault: 0,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return gconv.Uint64(insertID), nil
}

// UpdateGroup updates one existing scheduled-job group.
func (s *serviceImpl) UpdateGroup(ctx context.Context, in UpdateGroupInput) error {
	group, err := s.groupByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if group == nil {
		return bizerr.NewCode(CodeJobGroupNotFound)
	}

	code := strings.TrimSpace(in.Code)
	name := strings.TrimSpace(in.Name)
	if code == "" {
		return bizerr.NewCode(CodeJobGroupCodeRequired)
	}
	if name == "" {
		return bizerr.NewCode(CodeJobGroupNameRequired)
	}

	count, err := dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{Code: code}).
		WhereNot(dao.SysJobGroup.Columns().Id, in.ID).
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeJobGroupCodeExists)
	}

	_, err = dao.SysJobGroup.Ctx(ctx).
		Where(do.SysJobGroup{Id: in.ID}).
		Data(do.SysJobGroup{
			Code:      code,
			Name:      name,
			Remark:    strings.TrimSpace(in.Remark),
			SortOrder: in.SortOrder,
		}).
		Update()
	return err
}

// DeleteGroups removes one or more groups and migrates their jobs to the default group.
func (s *serviceImpl) DeleteGroups(ctx context.Context, ids string) error {
	groupIDs := parseUint64IDs(ids)
	if len(groupIDs) == 0 {
		return bizerr.NewCode(CodeJobGroupDeleteRequired)
	}

	defaultGroup, err := s.defaultGroup(ctx)
	if err != nil {
		return err
	}

	validIDs := make([]uint64, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		group, groupErr := s.groupByID(ctx, groupID)
		if groupErr != nil {
			return groupErr
		}
		if group == nil {
			continue
		}
		if group.IsDefault == 1 || group.Id == defaultGroup.Id {
			return bizerr.NewCode(CodeJobGroupDefaultDeleteDenied)
		}
		validIDs = append(validIDs, groupID)
	}
	if len(validIDs) == 0 {
		return bizerr.NewCode(CodeJobGroupDeleteEmpty)
	}

	return dao.SysJobGroup.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		jobCols := dao.SysJob.Columns()
		if _, txErr := dao.SysJob.Ctx(ctx).
			WhereIn(jobCols.GroupId, validIDs).
			Data(do.SysJob{GroupId: defaultGroup.Id}).
			Update(); txErr != nil {
			return txErr
		}
		_, txErr := dao.SysJobGroup.Ctx(ctx).
			WhereIn(dao.SysJobGroup.Columns().Id, validIDs).
			Delete()
		return txErr
	})
}

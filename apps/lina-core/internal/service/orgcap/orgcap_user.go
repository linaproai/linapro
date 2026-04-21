// This file implements user-facing organization capability queries and
// association persistence behind the orgcap service contract.

package orgcap

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	deptsvc "lina-core/internal/service/dept"
)

// ListUserDeptAssignments returns user -> department projections for the provided users.
func (s *serviceImpl) ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*UserDeptAssignment, error) {
	assignments := make(map[int]*UserDeptAssignment)
	if !s.Enabled(ctx) || len(userIDs) == 0 {
		return assignments, nil
	}

	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		WhereIn(dao.SysUserDept.Columns().UserId, userIDs).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	deptIDs := make([]int, 0, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		assignments[item.UserId] = &UserDeptAssignment{DeptID: item.DeptId}
		deptIDs = append(deptIDs, item.DeptId)
	}
	if len(deptIDs) == 0 {
		return assignments, nil
	}

	var deptList []*entity.SysDept
	err = dao.SysDept.Ctx(ctx).
		WhereIn(dao.SysDept.Columns().Id, deptIDs).
		Scan(&deptList)
	if err != nil {
		return nil, err
	}
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		for userID, assignment := range assignments {
			if assignment != nil && assignment.DeptID == deptItem.Id {
				assignments[userID] = &UserDeptAssignment{DeptID: deptItem.Id, DeptName: deptItem.Name}
			}
		}
	}
	return assignments, nil
}

// GetUserIDsByDept returns user IDs associated with the given department subtree.
func (s *serviceImpl) GetUserIDsByDept(ctx context.Context, deptID int) ([]int, error) {
	if !s.Enabled(ctx) {
		return []int{}, nil
	}
	deptIDs, err := s.deptSvc.GetDeptAndDescendantIds(ctx, deptID)
	if err != nil {
		return nil, err
	}

	var userDepts []*entity.SysUserDept
	err = dao.SysUserDept.Ctx(ctx).
		WhereIn(dao.SysUserDept.Columns().DeptId, deptIDs).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	seen := make(map[int]struct{}, len(userDepts))
	ids := make([]int, 0, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		if _, ok := seen[item.UserId]; ok {
			continue
		}
		seen[item.UserId] = struct{}{}
		ids = append(ids, item.UserId)
	}
	return ids, nil
}

// GetAllAssignedUserIDs returns all user IDs that currently hold department assignments.
func (s *serviceImpl) GetAllAssignedUserIDs(ctx context.Context) ([]int, error) {
	if !s.Enabled(ctx) {
		return []int{}, nil
	}

	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Fields(dao.SysUserDept.Columns().UserId).
		Distinct().
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		ids = append(ids, item.UserId)
	}
	return ids, nil
}

// GetUserDeptInfo returns one user's department projection.
func (s *serviceImpl) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	if !s.Enabled(ctx) {
		return 0, "", nil
	}

	var userDept *entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		Where(dao.SysUserDept.Columns().UserId, userID).
		Scan(&userDept)
	if err != nil || userDept == nil {
		return 0, "", err
	}

	var deptItem *entity.SysDept
	err = dao.SysDept.Ctx(ctx).
		Where(dao.SysDept.Columns().Id, userDept.DeptId).
		Scan(&deptItem)
	if err != nil || deptItem == nil {
		return 0, "", err
	}
	return deptItem.Id, deptItem.Name, nil
}

// GetUserDeptName returns one user's department name for online-session projection.
func (s *serviceImpl) GetUserDeptName(ctx context.Context, userID int) (string, error) {
	_, deptName, err := s.GetUserDeptInfo(ctx, userID)
	return deptName, err
}

// GetUserPostIDs returns one user's post association list.
func (s *serviceImpl) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	if !s.Enabled(ctx) {
		return []int{}, nil
	}

	var userPosts []*entity.SysUserPost
	err := dao.SysUserPost.Ctx(ctx).
		Where(dao.SysUserPost.Columns().UserId, userID).
		Scan(&userPosts)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(userPosts))
	for _, item := range userPosts {
		if item == nil {
			continue
		}
		ids = append(ids, item.PostId)
	}
	return ids, nil
}

// ReplaceUserAssignments rewrites one user's department and post associations.
func (s *serviceImpl) ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	if !s.Enabled(ctx) {
		return nil
	}
	if err := s.CleanupUserAssignments(ctx, userID); err != nil {
		return err
	}
	if deptID != nil && *deptID > 0 {
		if _, err := dao.SysUserDept.Ctx(ctx).Data(do.SysUserDept{UserId: userID, DeptId: *deptID}).Insert(); err != nil {
			return err
		}
	}
	for _, postID := range postIDs {
		if _, err := dao.SysUserPost.Ctx(ctx).Data(do.SysUserPost{UserId: userID, PostId: postID}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

// CleanupUserAssignments deletes one user's optional organization associations.
func (s *serviceImpl) CleanupUserAssignments(ctx context.Context, userID int) error {
	if !s.storageInstalled(ctx) {
		return nil
	}
	if _, err := dao.SysUserDept.Ctx(ctx).Where(dao.SysUserDept.Columns().UserId, userID).Delete(); err != nil {
		return err
	}
	if _, err := dao.SysUserPost.Ctx(ctx).Where(dao.SysUserPost.Columns().UserId, userID).Delete(); err != nil {
		return err
	}
	return nil
}

// UserDeptTree returns the optional department tree used by host user management.
func (s *serviceImpl) UserDeptTree(ctx context.Context) ([]*deptsvc.TreeNode, error) {
	if !s.Enabled(ctx) {
		return []*deptsvc.TreeNode{}, nil
	}
	return s.deptSvc.UserDeptTree(ctx)
}

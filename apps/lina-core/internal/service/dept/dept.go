// Package dept implements department tree queries, mutation flows, and
// descendant traversal helpers for the Lina core host service.
package dept

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// Service defines the dept service contract.
type Service interface {
	// List queries dept list with filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Create creates a new dept.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetById retrieves dept by ID.
	GetById(ctx context.Context, id int) (*entity.SysDept, error)
	// Update updates dept information with transaction support.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes a dept.
	Delete(ctx context.Context, id int) error
	// Tree builds dept tree structure.
	Tree(ctx context.Context) ([]*TreeNode, error)
	// Exclude returns dept list excluding specified dept and its descendants.
	Exclude(ctx context.Context, in ExcludeInput) ([]*entity.SysDept, error)
	// Users gets users for leader selection.
	// When deptId=0, returns all users. When deptId>0, returns users in the dept and all its sub-depts.
	// Supports keyword search on username/nickname and result limit.
	Users(ctx context.Context, deptId int, keyword string, limit int) ([]*DeptUser, error)
	// UserDeptTree builds dept tree with user count per node, plus an "未分配部门" virtual node.
	UserDeptTree(ctx context.Context) ([]*TreeNode, error)
	// GetDeptAndDescendantIds returns the given deptId plus all descendant dept IDs (cross-database compatible).
	// This is a shared utility method for traversing department hierarchies.
	GetDeptAndDescendantIds(ctx context.Context, deptId int) ([]int, error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

func New() Service {
	return &serviceImpl{}
}

type TreeNode struct {
	Id        int         `json:"id"`
	Label     string      `json:"label"`
	UserCount int         `json:"userCount"`
	Children  []*TreeNode `json:"children"`
}

type DeptUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"` // Username
	Nickname string `json:"nickname"` // Nickname
}

// ListInput defines input for List function.
type ListInput struct {
	Name   string // Department name, supports fuzzy search
	Status *int   // Status: 1=Normal 0=Disabled
}

// ListOutput defines output for List function.
type ListOutput struct {
	List []*entity.SysDept // Department list
}

// List queries dept list with filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysDept.Columns()
		m    = dao.SysDept.Ctx(ctx)
	)

	// Apply filters
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}

	// Query all, ordered by order_num ASC
	var list []*entity.SysDept
	err := m.OrderAsc(cols.OrderNum).Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List: list,
	}, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	ParentId int    // Parent department ID, 0 means top-level
	Name     string // Department name
	Code     string // Department code
	OrderNum int    // Display order
	Leader   int    // Leader user ID
	Phone    string // Contact phone
	Email    string // Email
	Status   int    // Status: 1=Normal 0=Disabled
	Remark   string // Remark
}

// Create creates a new dept.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check code uniqueness
	if in.Code != "" {
		if err := s.checkCodeUnique(ctx, in.Code, 0); err != nil {
			return 0, err
		}
	}

	// Calculate ancestors
	var ancestors string
	if in.ParentId == 0 {
		ancestors = "0"
	} else {
		parent, err := s.GetById(ctx, in.ParentId)
		if err != nil {
			return 0, err
		}
		ancestors = fmt.Sprintf("%s,%d", parent.Ancestors, in.ParentId)
	}

	// Insert dept (GoFrame auto-fills created_at and updated_at)
	id, err := dao.SysDept.Ctx(ctx).Data(do.SysDept{
		ParentId:  in.ParentId,
		Ancestors: ancestors,
		Name:      in.Name,
		Code:      in.Code,
		OrderNum:  in.OrderNum,
		Leader:    in.Leader,
		Phone:     in.Phone,
		Email:     in.Email,
		Status:    in.Status,
		Remark:    in.Remark,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetById retrieves dept by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysDept, error) {
	var dept *entity.SysDept
	err := dao.SysDept.Ctx(ctx).
		Where(do.SysDept{Id: id}).
		Scan(&dept)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, gerror.New("部门不存在")
	}
	return dept, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id       int     // Department ID
	ParentId *int    // Parent department ID
	Name     *string // Department name
	Code     *string // Department code
	OrderNum *int    // Display order
	Leader   *int    // Leader user ID
	Phone    *string // Contact phone
	Email    *string // Email
	Status   *int    // Status: 1=Normal 0=Disabled
	Remark   *string // Remark
}

// Update updates dept information with transaction support.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	// Check dept exists
	dept, err := s.GetById(ctx, in.Id)
	if err != nil {
		return err
	}

	data := do.SysDept{}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Code != nil {
		if *in.Code != "" {
			if err := s.checkCodeUnique(ctx, *in.Code, in.Id); err != nil {
				return err
			}
		}
		data.Code = *in.Code
	}
	if in.OrderNum != nil {
		data.OrderNum = *in.OrderNum
	}
	if in.Leader != nil {
		data.Leader = *in.Leader
	}
	if in.Phone != nil {
		data.Phone = *in.Phone
	}
	if in.Email != nil {
		data.Email = *in.Email
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	// Handle parent change: recalculate ancestors
	if in.ParentId != nil && *in.ParentId != dept.ParentId {
		newParentId := *in.ParentId
		var newAncestors string
		if newParentId == 0 {
			newAncestors = "0"
		} else {
			parent, err := s.GetById(ctx, newParentId)
			if err != nil {
				return err
			}
			newAncestors = fmt.Sprintf("%s,%d", parent.Ancestors, newParentId)
		}

		oldAncestors := dept.Ancestors
		data.ParentId = newParentId
		data.Ancestors = newAncestors

		// Update children's ancestors within transaction
		oldPrefix := fmt.Sprintf("%s,%d", oldAncestors, in.Id)
		newPrefix := fmt.Sprintf("%s,%d", newAncestors, in.Id)

		cols := dao.SysDept.Columns()
		var children []*entity.SysDept
		err = dao.SysDept.Ctx(ctx).
			Where(
				dao.SysDept.Ctx(ctx).Builder().
					WhereLike(cols.Ancestors, oldPrefix+",%").
					WhereOr(cols.ParentId, in.Id),
			).
			Scan(&children)
		if err != nil {
			return err
		}

		// Use transaction for updating children
		if len(children) > 0 {
			err = dao.SysDept.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
				for _, child := range children {
					childNewAncestors := gstr.Replace(child.Ancestors, oldPrefix, newPrefix, 1)
					_, err = dao.SysDept.Ctx(ctx).
						Where(do.SysDept{Id: child.Id}).
						Data(do.SysDept{
							Ancestors: childNewAncestors,
						}).
						Update()
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
	}

	_, err = dao.SysDept.Ctx(ctx).Where(do.SysDept{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes a dept.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	cols := dao.SysDept.Columns()

	// Check no children
	childCount, err := dao.SysDept.Ctx(ctx).
		Where(cols.ParentId, id).
		Count()
	if err != nil {
		return err
	}
	if childCount > 0 {
		return gerror.New("存在子部门，不允许删除")
	}

	// Check no users in dept
	userCount, err := dao.SysUserDept.Ctx(ctx).
		Where(do.SysUserDept{DeptId: id}).
		Count()
	if err != nil {
		return err
	}
	if userCount > 0 {
		return gerror.New("部门存在用户，不允许删除")
	}

	// Soft delete using GoFrame's auto soft-delete feature
	_, err = dao.SysDept.Ctx(ctx).
		Where(do.SysDept{Id: id}).
		Delete()
	return err
}

// Tree builds dept tree structure.
func (s *serviceImpl) Tree(ctx context.Context) ([]*TreeNode, error) {
	cols := dao.SysDept.Columns()

	var depts []*entity.SysDept
	err := dao.SysDept.Ctx(ctx).
		OrderAsc(cols.OrderNum).
		Scan(&depts)
	if err != nil {
		return nil, err
	}

	// Build tree from flat list
	nodeMap := make(map[int]*TreeNode)
	for _, d := range depts {
		nodeMap[d.Id] = &TreeNode{
			Id:       d.Id,
			Label:    d.Name,
			Children: make([]*TreeNode, 0),
		}
	}

	var roots []*TreeNode
	for _, d := range depts {
		node := nodeMap[d.Id]
		if parent, ok := nodeMap[d.ParentId]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	return roots, nil
}

// ExcludeInput defines input for Exclude function.
type ExcludeInput struct {
	Id int // Department ID to exclude
}

// Exclude returns dept list excluding specified dept and its descendants.
func (s *serviceImpl) Exclude(ctx context.Context, in ExcludeInput) ([]*entity.SysDept, error) {
	// Get the target dept
	dept, err := s.GetById(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	cols := dao.SysDept.Columns()
	prefix := fmt.Sprintf("%s,%d", dept.Ancestors, in.Id)

	// Get all depts excluding the target and its descendants
	var list []*entity.SysDept
	err = dao.SysDept.Ctx(ctx).
		WhereNot(cols.Id, in.Id).
		WhereNotLike(cols.Ancestors, prefix+",%").
		WhereNotLike(cols.Ancestors, prefix).
		OrderAsc(cols.OrderNum).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// Users gets users for leader selection.
// When deptId=0, returns all users. When deptId>0, returns users in the dept and all its sub-depts.
// Supports keyword search on username/nickname and result limit.
func (s *serviceImpl) Users(ctx context.Context, deptId int, keyword string, limit int) ([]*DeptUser, error) {
	uCols := dao.SysUser.Columns()

	if deptId == 0 {
		// Return all users (for new dept creation)
		q := dao.SysUser.Ctx(ctx).
			Fields(uCols.Id, uCols.Username, uCols.Nickname)
		if keyword != "" {
			q = q.Where(
				fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", uCols.Username, uCols.Nickname),
				"%"+keyword+"%", "%"+keyword+"%",
			)
		}
		if limit > 0 {
			q = q.Limit(limit)
		}
		var users []*entity.SysUser
		if err := q.Scan(&users); err != nil {
			return nil, err
		}
		result := make([]*DeptUser, 0, len(users))
		for _, u := range users {
			result = append(result, &DeptUser{
				Id:       u.Id,
				Username: u.Username,
				Nickname: u.Nickname,
			})
		}
		return result, nil
	}

	// Collect the selected dept and all its descendant depts via parent_id (cross-database compatible).
	var (
		deptCols  = dao.SysDept.Columns()
		deptIds   = []int{deptId}
		parentIds = []int{deptId}
	)
	for len(parentIds) > 0 {
		childValues, err := dao.SysDept.Ctx(ctx).
			WhereIn(deptCols.ParentId, parentIds).
			Fields(deptCols.Id).
			Array()
		if err != nil {
			return nil, err
		}
		var childIds = gconv.Ints(childValues)
		deptIds = append(deptIds, childIds...)
		parentIds = childIds
	}

	// Query sys_user_dept for user_ids in the subtree
	var userDepts []*entity.SysUserDept
	err := dao.SysUserDept.Ctx(ctx).
		WhereIn(dao.SysUserDept.Columns().DeptId, deptIds).
		Scan(&userDepts)
	if err != nil {
		return nil, err
	}

	if len(userDepts) == 0 {
		return make([]*DeptUser, 0), nil
	}

	// Deduplicate user IDs
	seen := make(map[int]struct{})
	userIds := make([]int, 0, len(userDepts))
	for _, ud := range userDepts {
		if _, ok := seen[ud.UserId]; !ok {
			seen[ud.UserId] = struct{}{}
			userIds = append(userIds, ud.UserId)
		}
	}

	// Query sys_user for those IDs
	q := dao.SysUser.Ctx(ctx).
		Fields(uCols.Id, uCols.Username, uCols.Nickname).
		WhereIn(uCols.Id, userIds)
	if keyword != "" {
		q = q.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", uCols.Username, uCols.Nickname),
			"%"+keyword+"%", "%"+keyword+"%",
		)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	var users []*entity.SysUser
	if err := q.Scan(&users); err != nil {
		return nil, err
	}

	// Convert to DeptUser
	result := make([]*DeptUser, 0, len(users))
	for _, u := range users {
		result = append(result, &DeptUser{
			Id:       u.Id,
			Username: u.Username,
			Nickname: u.Nickname,
		})
	}

	return result, nil
}

// UserDeptTree builds dept tree with user count per node, plus an "未分配部门" virtual node.
func (s *serviceImpl) UserDeptTree(ctx context.Context) ([]*TreeNode, error) {
	// Get base tree
	nodes, err := s.Tree(ctx)
	if err != nil {
		return nil, err
	}

	// Get user count per dept via sys_user_dept (only count non-deleted users)
	type DeptCount struct {
		DeptId int `json:"dept_id"`
		Cnt    int `json:"cnt"`
	}
	var counts []DeptCount
	err = dao.SysUserDept.Ctx(ctx).
		Fields("dept_id, COUNT(*) as cnt").
		InnerJoin(
			dao.SysUser.Table(),
			fmt.Sprintf(
				"%s.%s = %s.%s",
				dao.SysUserDept.Table(), dao.SysUserDept.Columns().UserId,
				dao.SysUser.Table(), dao.SysUser.Columns().Id,
			),
		).
		Group("dept_id").
		Scan(&counts)
	if err != nil {
		return nil, err
	}
	countMap := make(map[int]int)
	for _, c := range counts {
		countMap[c.DeptId] = c.Cnt
	}

	// Apply user counts to tree nodes (parent = self + all descendants)
	var applyCount func(nodes []*TreeNode)
	applyCount = func(nodes []*TreeNode) {
		for _, n := range nodes {
			applyCount(n.Children)
			n.UserCount = countMap[n.Id]
			for _, child := range n.Children {
				n.UserCount += child.UserCount
			}
			n.Label = fmt.Sprintf("%s(%d)", n.Label, n.UserCount)
		}
	}
	applyCount(nodes)

	// Count unassigned users (users not in sys_user_dept)
	totalUsers, err := dao.SysUser.Ctx(ctx).Count()
	if err != nil {
		return nil, err
	}
	assignedUsers := 0
	for _, c := range countMap {
		assignedUsers += c
	}
	unassignedCount := totalUsers - assignedUsers

	// Append "未分配部门" virtual node at the end
	unassignedNode := &TreeNode{
		Id:        0,
		Label:     fmt.Sprintf("未分配部门(%d)", unassignedCount),
		UserCount: unassignedCount,
		Children:  make([]*TreeNode, 0),
	}
	result := make([]*TreeNode, 0, len(nodes)+1)
	result = append(result, nodes...)
	result = append(result, unassignedNode)

	return result, nil
}

// checkCodeUnique checks if the dept code is unique (excluding the given dept ID for updates).
func (s *serviceImpl) checkCodeUnique(ctx context.Context, code string, excludeId int) error {
	cols := dao.SysDept.Columns()
	m := dao.SysDept.Ctx(ctx).
		Where(cols.Code, code)
	if excludeId > 0 {
		m = m.WhereNot(cols.Id, excludeId)
	}
	count, err := m.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return gerror.New("部门编码已存在")
	}
	return nil
}

// GetDeptAndDescendantIds returns the given deptId plus all descendant dept IDs (cross-database compatible).
// This is a shared utility method for traversing department hierarchies.
func (s *serviceImpl) GetDeptAndDescendantIds(ctx context.Context, deptId int) ([]int, error) {
	var (
		deptCols  = dao.SysDept.Columns()
		deptIds   = []int{deptId}
		parentIds = []int{deptId}
	)
	for len(parentIds) > 0 {
		childValues, err := dao.SysDept.Ctx(ctx).
			WhereIn(deptCols.ParentId, parentIds).
			Fields(deptCols.Id).
			Array()
		if err != nil {
			return nil, err
		}
		childIds := gconv.Ints(childValues)
		deptIds = append(deptIds, childIds...)
		parentIds = childIds
	}
	return deptIds, nil
}

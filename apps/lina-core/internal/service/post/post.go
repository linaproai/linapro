// Package post implements post management, option queries, and export services
// for the Lina core host service.
package post

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/dept"
)

// Service defines the post service contract.
type Service interface {
	// List queries post list with pagination and filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Create creates a new post.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetById retrieves post by ID.
	GetById(ctx context.Context, id int) (*entity.SysPost, error)
	// Update updates post information.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes posts by comma-separated IDs.
	Delete(ctx context.Context, ids string) error
	// DeptTree returns department tree structure with "未分配部门" virtual node.
	DeptTree(ctx context.Context) ([]*DeptTreeNode, error)
	// OptionSelect returns post options for select dropdown.
	OptionSelect(ctx context.Context, in OptionSelectInput) ([]PostOption, error)
	// Export generates an Excel file with post data based on filters.
	Export(ctx context.Context, in ExportInput) (data []byte, err error)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	deptSvc dept.Service
}

func New() Service {
	return &serviceImpl{
		deptSvc: dept.New(),
	}
}

type ListInput struct {
	PageNum  int
	PageSize int
	DeptId   *int
	Code     string
	Name     string // Post name, supports fuzzy search
	Status   *int   // Status: 1=Normal 0=Disabled
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*entity.SysPost // Post list
	Total int               // Total count
}

// List queries post list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		cols = dao.SysPost.Columns()
		m    = dao.SysPost.Ctx(ctx)
	)

	// Apply filters
	if in.DeptId != nil {
		if *in.DeptId == 0 {
			// Unassigned: posts with dept_id = 0
			m = m.Where(cols.DeptId, 0)
		} else {
			// Include selected dept and all descendant depts
			deptIds, err := s.getDeptAndDescendantIds(ctx, *in.DeptId)
			if err != nil {
				return nil, err
			}
			m = m.WhereIn(cols.DeptId, deptIds)
		}
	}
	if in.Code != "" {
		m = m.WhereLike(cols.Code, "%"+in.Code+"%")
	}
	if in.Name != "" {
		m = m.WhereLike(cols.Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(cols.Status, *in.Status)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	var list []*entity.SysPost
	err = m.Page(in.PageNum, in.PageSize).
		OrderAsc(cols.Sort).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	return &ListOutput{
		List:  list,
		Total: total,
	}, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	DeptId int    // Department ID, 0 means unassigned
	Code   string // Post code
	Name   string // Post name
	Sort   int    // Display order
	Status int    // Status: 1=Normal 0=Disabled
	Remark string // Remark
}

// Create creates a new post.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	// Check code uniqueness
	count, err := dao.SysPost.Ctx(ctx).
		Where(do.SysPost{Code: in.Code}).
		Count()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, gerror.New("岗位编码已存在")
	}

	// Insert post (GoFrame auto-fills created_at and updated_at)
	id, err := dao.SysPost.Ctx(ctx).Data(do.SysPost{
		DeptId: in.DeptId,
		Code:   in.Code,
		Name:   in.Name,
		Sort:   in.Sort,
		Status: in.Status,
		Remark: in.Remark,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetById retrieves post by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int) (*entity.SysPost, error) {
	var post *entity.SysPost
	err := dao.SysPost.Ctx(ctx).
		Where(do.SysPost{Id: id}).
		Scan(&post)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, gerror.New("岗位不存在")
	}
	return post, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id     int     // Post ID
	DeptId *int    // Department ID
	Code   *string // Post code
	Name   *string // Post name
	Sort   *int    // Display order
	Status *int    // Status: 1=Normal 0=Disabled
	Remark *string // Remark
}

// Update updates post information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	// Check post exists
	if _, err := s.GetById(ctx, in.Id); err != nil {
		return err
	}

	data := do.SysPost{}
	if in.DeptId != nil {
		data.DeptId = *in.DeptId
	}
	if in.Code != nil {
		data.Code = *in.Code
	}
	if in.Name != nil {
		data.Name = *in.Name
	}
	if in.Sort != nil {
		data.Sort = *in.Sort
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err := dao.SysPost.Ctx(ctx).Where(do.SysPost{Id: in.Id}).Data(data).Update()
	return err
}

// Delete soft-deletes posts by comma-separated IDs.
func (s *serviceImpl) Delete(ctx context.Context, ids string) error {
	idList := gstr.SplitAndTrim(ids, ",")
	if len(idList) == 0 {
		return gerror.New("请选择要删除的岗位")
	}

	cols := dao.SysUserPost.Columns()
	var validIds []int
	for _, idStr := range idList {
		id := gconv.Int(idStr)
		if id == 0 {
			continue
		}

		// Check if post is assigned to users
		count, err := dao.SysUserPost.Ctx(ctx).
			Where(cols.PostId, id).
			Count()
		if err != nil {
			return err
		}
		if count > 0 {
			return gerror.Newf("岗位ID %d 已分配给用户，不能删除", id)
		}
		validIds = append(validIds, id)
	}

	if len(validIds) == 0 {
		return gerror.New("没有有效的岗位ID")
	}

	// Soft delete using GoFrame's auto soft-delete feature
	_, err := dao.SysPost.Ctx(ctx).
		WhereIn(dao.SysPost.Columns().Id, validIds).
		Delete()
	return err
}

// DeptTreeNode defines a department tree node.
type DeptTreeNode struct {
	Id        int             `json:"id"`        // Department ID
	Label     string          `json:"label"`     // Department name (with post count)
	PostCount int             `json:"postCount"` // Post count
	Children  []*DeptTreeNode `json:"children"`  // Child departments
}

// DeptTree returns department tree structure with "未分配部门" virtual node.
func (s *serviceImpl) DeptTree(ctx context.Context) ([]*DeptTreeNode, error) {
	cols := dao.SysDept.Columns()
	var depts []*entity.SysDept
	err := dao.SysDept.Ctx(ctx).
		OrderAsc(cols.OrderNum).
		Scan(&depts)
	if err != nil {
		return nil, err
	}

	// Build tree
	nodeMap := make(map[int]*DeptTreeNode)
	for _, d := range depts {
		nodeMap[d.Id] = &DeptTreeNode{
			Id:       d.Id,
			Label:    d.Name,
			Children: make([]*DeptTreeNode, 0),
		}
	}

	var roots []*DeptTreeNode
	for _, d := range depts {
		node := nodeMap[d.Id]
		if parent, ok := nodeMap[d.ParentId]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	// Append "未分配部门" virtual node
	unassignedNode := &DeptTreeNode{
		Id:       0,
		Label:    "未分配部门",
		Children: make([]*DeptTreeNode, 0),
	}
	roots = append(roots, unassignedNode)

	// Count posts per dept
	type DeptCount struct {
		DeptId int `json:"dept_id"` // Department ID
		Cnt    int `json:"cnt"`     // Post count
	}
	var counts []DeptCount
	err = dao.SysPost.Ctx(ctx).
		Fields("dept_id, COUNT(*) as cnt").
		Group("dept_id").
		Scan(&counts)
	if err != nil {
		return nil, err
	}
	countMap := make(map[int]int)
	for _, c := range counts {
		countMap[c.DeptId] = c.Cnt
	}

	// Apply post counts (parent = self + all descendants), process children first
	var applyCount func(nodes []*DeptTreeNode)
	applyCount = func(nodes []*DeptTreeNode) {
		for _, n := range nodes {
			applyCount(n.Children)
			n.PostCount = countMap[n.Id]
			for _, child := range n.Children {
				n.PostCount += child.PostCount
			}
			n.Label = fmt.Sprintf("%s(%d)", n.Label, n.PostCount)
		}
	}
	// Apply to real dept nodes only (not the virtual unassigned node)
	applyCount(roots[:len(roots)-1])

	// Handle unassigned node separately
	unassignedNode.PostCount = countMap[0]
	unassignedNode.Label = fmt.Sprintf("未分配部门(%d)", unassignedNode.PostCount)

	return roots, nil
}

// PostOption defines a post option for select dropdown.
type PostOption struct {
	PostId   int    `json:"postId"`   // Post ID
	PostName string `json:"postName"` // Post name
}

// OptionSelectInput defines input for OptionSelect function.
type OptionSelectInput struct {
	DeptId *int // Department ID
}

// OptionSelect returns post options for select dropdown.
func (s *serviceImpl) OptionSelect(ctx context.Context, in OptionSelectInput) ([]PostOption, error) {
	cols := dao.SysPost.Columns()
	m := dao.SysPost.Ctx(ctx).
		Where(cols.Status, 1)

	if in.DeptId != nil {
		deptIds, err := s.getDeptAndDescendantIds(ctx, *in.DeptId)
		if err != nil {
			return nil, err
		}
		m = m.WhereIn(cols.DeptId, deptIds)
	}

	var list []*entity.SysPost
	err := m.OrderAsc(cols.Sort).Scan(&list)
	if err != nil {
		return nil, err
	}

	options := make([]PostOption, 0, len(list))
	for _, p := range list {
		options = append(options, PostOption{
			PostId:   p.Id,
			PostName: p.Name,
		})
	}

	return options, nil
}

// getDeptAndDescendantIds returns the given deptId plus all descendant dept IDs using shared dept service method.
func (s *serviceImpl) getDeptAndDescendantIds(ctx context.Context, deptId int) ([]int, error) {
	return s.deptSvc.GetDeptAndDescendantIds(ctx, deptId)
}

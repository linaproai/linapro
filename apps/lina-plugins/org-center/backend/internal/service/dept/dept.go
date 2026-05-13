// Package dept implements department management for the org-center source
// plugin. It owns plugin_org_center_dept and related organization queries through plugin-local
// services instead of depending on host-internal dept packages.
package dept

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-org-center/backend/internal/dao"
	"lina-plugin-org-center/backend/internal/model/do"
	entitymodel "lina-plugin-org-center/backend/internal/model/entity"
)

// Table and column constants for organization storage.
const (
	colDeptID        = "id"
	colDeptParentID  = "parent_id"
	colDeptAncestors = "ancestors"
	colDeptName      = "name"
	colDeptCode      = "code"
	colDeptOrderNum  = "order_num"
	colDeptLeader    = "leader"
	colDeptPhone     = "phone"
	colDeptEmail     = "email"
	colDeptStatus    = "status"
	colDeptRemark    = "remark"

	colUserID       = "id"
	colUserUsername = "username"
	colUserNickname = "nickname"
	colUserDeptID   = "dept_id"
	colUserUserID   = "user_id"
)

// Service defines the department service contract.
type Service interface {
	// List queries the department list with optional filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// Create creates one department record.
	Create(ctx context.Context, in CreateInput) (int, error)
	// GetByID retrieves one department detail by primary key.
	GetByID(ctx context.Context, id int) (*DeptEntity, error)
	// Update updates one department record.
	Update(ctx context.Context, in UpdateInput) error
	// Delete deletes one department when no child or user binding blocks it.
	Delete(ctx context.Context, id int) error
	// Tree returns the plain department tree.
	Tree(ctx context.Context) ([]*TreeNode, error)
	// Exclude returns department candidates excluding one subtree.
	Exclude(ctx context.Context, in ExcludeInput) ([]*DeptEntity, error)
	// Users returns selectable users for one department subtree.
	Users(ctx context.Context, deptID int, keyword string, limit int) ([]*DeptUser, error)
	// DescendantDeptIDs returns the given department plus all descendants.
	DescendantDeptIDs(ctx context.Context, deptID int) ([]int, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains plugin-owned department rows.
}

// New creates and returns a new department service instance.
func New(tenantFilter plugincontract.TenantFilterService) Service {
	return &serviceImpl{tenantFilter: tenantFilter}
}

// DeptEntity mirrors the plugin-local generated plugin_org_center_dept entity.
type DeptEntity = entitymodel.Dept

// ListInput defines filters for department list queries.
type ListInput struct {
	Name   string
	Status *int
}

// ListOutput defines the department list result.
type ListOutput struct {
	List []*DeptEntity
}

// CreateInput defines the create-department input.
type CreateInput struct {
	ParentId int
	Name     string
	Code     string
	OrderNum int
	Leader   int
	Phone    string
	Email    string
	Status   int
	Remark   string
}

// UpdateInput defines the update-department input.
type UpdateInput struct {
	Id       int
	ParentId *int
	Name     *string
	Code     *string
	OrderNum *int
	Leader   *int
	Phone    *string
	Email    *string
	Status   *int
	Remark   *string
}

// ExcludeInput defines the exclude-subtree input.
type ExcludeInput struct {
	Id int
}

// TreeNode represents one department tree node.
type TreeNode struct {
	Id       int         `json:"id"`
	Label    string      `json:"label"`
	Children []*TreeNode `json:"children"`
}

// DeptUser represents one selectable user row.
type DeptUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// userRow reuses the plugin-local generated sys_user entity.
type userRow = entitymodel.SysUser

// List queries the department list with optional filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	model := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx))
	if in.Name != "" {
		model = model.WhereLike(colDeptName, "%"+in.Name+"%")
	}
	if in.Status != nil {
		model = model.Where(colDeptStatus, *in.Status)
	}

	list := make([]*DeptEntity, 0)
	if err := model.OrderAsc(colDeptOrderNum).Scan(&list); err != nil {
		return nil, err
	}
	return &ListOutput{List: list}, nil
}

// Create creates one department record.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int, error) {
	if in.Code != "" {
		if err := s.checkCodeUnique(ctx, in.Code, 0); err != nil {
			return 0, err
		}
	}

	ancestors := "0"
	if in.ParentId != 0 {
		parent, err := s.GetByID(ctx, in.ParentId)
		if err != nil {
			return 0, err
		}
		ancestors = fmt.Sprintf("%s,%d", parent.Ancestors, in.ParentId)
	}

	id, err := dao.Dept.Ctx(ctx).Data(do.Dept{
		TenantId:  s.tenantFilter.Current(ctx),
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

// GetByID retrieves one department detail by primary key.
func (s *serviceImpl) GetByID(ctx context.Context, id int) (*DeptEntity, error) {
	var dept *DeptEntity
	err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
		Where(colDeptID, id).
		Scan(&dept)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, bizerr.NewCode(CodeDeptNotFound)
	}
	return dept, nil
}

// Update updates one department record.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	dept, err := s.GetByID(ctx, in.Id)
	if err != nil {
		return err
	}

	buildData := func(includeParent bool, newParentID int, newAncestors string) (do.Dept, error) {
		data := do.Dept{}
		if includeParent {
			data.ParentId = newParentID
			data.Ancestors = newAncestors
		}
		if in.Name != nil {
			data.Name = *in.Name
		}
		if in.Code != nil {
			if *in.Code != "" {
				if err := s.checkCodeUnique(ctx, *in.Code, in.Id); err != nil {
					return data, err
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
		return data, nil
	}

	if in.ParentId != nil && *in.ParentId != dept.ParentId {
		newParentID := *in.ParentId
		newAncestors := "0"
		if newParentID != 0 {
			parent, err := s.GetByID(ctx, newParentID)
			if err != nil {
				return err
			}
			newAncestors = fmt.Sprintf("%s,%d", parent.Ancestors, newParentID)
		}

		oldPrefix := fmt.Sprintf("%s,%d", dept.Ancestors, in.Id)
		newPrefix := fmt.Sprintf("%s,%d", newAncestors, in.Id)
		children := make([]*DeptEntity, 0)
		err = s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
			WhereLike(colDeptAncestors, oldPrefix+",%").
			WhereOr(colDeptParentID, in.Id).
			Scan(&children)
		if err != nil {
			return err
		}

		err = dao.Dept.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			for _, child := range children {
				if child == nil {
					continue
				}
				childAncestors := gstr.Replace(child.Ancestors, oldPrefix, newPrefix, 1)
				_, err = tx.Model(dao.Dept.Table()).Safe().Ctx(ctx).
					OmitNilData().
					Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
					Where(colDeptID, child.Id).
					Data(do.Dept{Ancestors: childAncestors}).
					Update()
				if err != nil {
					return err
				}
			}
			data, buildErr := buildData(true, newParentID, newAncestors)
			if buildErr != nil {
				return buildErr
			}
			_, err = tx.Model(dao.Dept.Table()).Safe().Ctx(ctx).
				OmitNilData().
				Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
				Where(colDeptID, in.Id).
				Data(data).
				Update()
			return err
		})
		return err
	}

	data, err := buildData(false, 0, "")
	if err != nil {
		return err
	}
	_, err = dao.Dept.Ctx(ctx).
		OmitNilData().
		Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
		Where(colDeptID, in.Id).
		Data(data).
		Update()
	return err
}

// Delete deletes one department when no child or user binding blocks it.
func (s *serviceImpl) Delete(ctx context.Context, id int) error {
	childCount, err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
		Where(colDeptParentID, id).
		Count()
	if err != nil {
		return err
	}
	if childCount > 0 {
		return bizerr.NewCode(CodeDeptHasChildrenDeleteDenied)
	}

	userCount, err := s.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		Where(colUserDeptID, id).
		Count()
	if err != nil {
		return err
	}
	if userCount > 0 {
		return bizerr.NewCode(CodeDeptHasUsersDeleteDenied)
	}

	_, err = s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
		Where(colDeptID, id).
		Delete()
	return err
}

// Tree returns the plain department tree.
func (s *serviceImpl) Tree(ctx context.Context) ([]*TreeNode, error) {
	deptList := make([]*DeptEntity, 0)
	err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
		OrderAsc(colDeptOrderNum).
		Scan(&deptList)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[int]*TreeNode, len(deptList))
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		nodeMap[deptItem.Id] = &TreeNode{Id: deptItem.Id, Label: deptItem.Name, Children: make([]*TreeNode, 0)}
	}

	roots := make([]*TreeNode, 0)
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		node := nodeMap[deptItem.Id]
		if parent, ok := nodeMap[deptItem.ParentId]; ok {
			parent.Children = append(parent.Children, node)
			continue
		}
		roots = append(roots, node)
	}
	return roots, nil
}

// Exclude returns department candidates excluding one subtree.
func (s *serviceImpl) Exclude(ctx context.Context, in ExcludeInput) ([]*DeptEntity, error) {
	dept, err := s.GetByID(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("%s,%d", dept.Ancestors, in.Id)
	list := make([]*DeptEntity, 0)
	err = s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
		WhereNot(colDeptID, in.Id).
		WhereNotLike(colDeptAncestors, prefix+",%").
		WhereNotLike(colDeptAncestors, prefix).
		OrderAsc(colDeptOrderNum).
		Scan(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Users returns selectable users for one department subtree.
func (s *serviceImpl) Users(ctx context.Context, deptID int, keyword string, limit int) ([]*DeptUser, error) {
	if deptID == 0 {
		model := s.tenantFilter.Apply(ctx, dao.SysUser.Ctx(ctx)).Fields(colUserID, colUserUsername, colUserNickname)
		if keyword != "" {
			model = model.Where(
				fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", colUserUsername, colUserNickname),
				"%"+keyword+"%", "%"+keyword+"%",
			)
		}
		if limit > 0 {
			model = model.Limit(limit)
		}
		rows := make([]*userRow, 0)
		if err := model.Scan(&rows); err != nil {
			return nil, err
		}
		return toDeptUsers(rows), nil
	}

	deptIDs, err := s.DescendantDeptIDs(ctx, deptID)
	if err != nil {
		return nil, err
	}

	userDeptRows := make([]*entitymodel.UserDept, 0)
	err = s.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		WhereIn(colUserDeptID, deptIDs).
		Scan(&userDeptRows)
	if err != nil {
		return nil, err
	}
	if len(userDeptRows) == 0 {
		return make([]*DeptUser, 0), nil
	}

	seen := make(map[int]struct{}, len(userDeptRows))
	userIDs := make([]int, 0, len(userDeptRows))
	for _, row := range userDeptRows {
		if row == nil {
			continue
		}
		if _, ok := seen[row.UserId]; ok {
			continue
		}
		seen[row.UserId] = struct{}{}
		userIDs = append(userIDs, row.UserId)
	}

	model := dao.SysUser.Ctx(ctx).
		Fields(colUserID, colUserUsername, colUserNickname).
		Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
		WhereIn(colUserID, userIDs)
	if keyword != "" {
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", colUserUsername, colUserNickname),
			"%"+keyword+"%", "%"+keyword+"%",
		)
	}
	if limit > 0 {
		model = model.Limit(limit)
	}
	rows := make([]*userRow, 0)
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	return toDeptUsers(rows), nil
}

// DescendantDeptIDs returns the given department plus all descendants.
func (s *serviceImpl) DescendantDeptIDs(ctx context.Context, deptID int) ([]int, error) {
	deptIDs := []int{deptID}
	parentIDs := []int{deptID}
	for len(parentIDs) > 0 {
		childValues, err := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
			WhereIn(colDeptParentID, parentIDs).
			Fields(colDeptID).
			Array()
		if err != nil {
			return nil, err
		}
		childIDs := gconv.Ints(childValues)
		deptIDs = append(deptIDs, childIDs...)
		parentIDs = childIDs
	}
	return deptIDs, nil
}

// checkCodeUnique checks whether one department code already exists.
func (s *serviceImpl) checkCodeUnique(ctx context.Context, code string, excludeID int) error {
	model := s.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).Where(colDeptCode, code)
	if excludeID > 0 {
		model = model.WhereNot(colDeptID, excludeID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeDeptCodeExists)
	}
	return nil
}

// toDeptUsers converts narrow user projections into selectable user rows.
func toDeptUsers(rows []*userRow) []*DeptUser {
	result := make([]*DeptUser, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		result = append(result, &DeptUser{Id: row.Id, Username: row.Username, Nickname: row.Nickname})
	}
	return result
}

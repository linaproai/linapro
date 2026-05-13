// Package orgcapadapter provides the org-center implementation of the
// host organization-capability provider contract.
package orgcapadapter

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/orgcap"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-org-center/backend/internal/dao"
	"lina-plugin-org-center/backend/internal/model/do"
	entitymodel "lina-plugin-org-center/backend/internal/model/entity"
	deptsvc "lina-plugin-org-center/backend/internal/service/dept"
)

const (
	// postStatusEnabled is the enabled status value used by org-center posts.
	postStatusEnabled = 1
	// orgCapUnassignedDeptLabelKey is the runtime i18n key for the synthetic
	// Unassigned node exposed through the host orgcap contract.
	orgCapUnassignedDeptLabelKey = "plugin.org-center.post.tree.unassignedDept"
)

// Provider implements the stable host organization-capability contract.
type Provider struct {
	deptSvc      deptsvc.Service                    // deptSvc resolves department tree relationships.
	tenantFilter plugincontract.TenantFilterService // tenantFilter constrains organization provider queries.
}

// Ensure Provider implements the published organization-capability provider.
var _ orgcap.Provider = (*Provider)(nil)

// New creates and returns a new provider instance.
func New(tenantFilter plugincontract.TenantFilterService) *Provider {
	return &Provider{
		deptSvc:      deptsvc.New(tenantFilter),
		tenantFilter: tenantFilter,
	}
}

// deptCountRow is the grouped user-count projection keyed by department.
type deptCountRow struct {
	DeptID int `json:"deptId"`
	Cnt    int `json:"cnt"`
}

// ListUserDeptAssignments returns user -> department projections for the provided users.
func (p *Provider) ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment)
	if len(userIDs) == 0 {
		return assignments, nil
	}

	var userDepts []*entitymodel.UserDept
	if err := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		WhereIn(dao.UserDept.Columns().UserId, userIDs).
		Scan(&userDepts); err != nil {
		return nil, err
	}

	deptIDs := make([]int, 0, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		assignments[item.UserId] = &orgcap.UserDeptAssignment{DeptID: item.DeptId}
		deptIDs = append(deptIDs, item.DeptId)
	}
	if len(deptIDs) == 0 {
		return assignments, nil
	}

	var deptList []*entitymodel.Dept
	if err := p.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
		WhereIn(dao.Dept.Columns().Id, deptIDs).
		Scan(&deptList); err != nil {
		return nil, err
	}
	for _, deptItem := range deptList {
		if deptItem == nil {
			continue
		}
		for userID, assignment := range assignments {
			if assignment != nil && assignment.DeptID == deptItem.Id {
				assignments[userID] = &orgcap.UserDeptAssignment{
					DeptID:   deptItem.Id,
					DeptName: deptItem.Name,
				}
			}
		}
	}
	return assignments, nil
}

// GetUserIDsByDept returns user IDs associated with the given department subtree.
func (p *Provider) GetUserIDsByDept(ctx context.Context, deptID int) ([]int, error) {
	deptIDs, err := p.deptSvc.DescendantDeptIDs(ctx, deptID)
	if err != nil {
		return nil, err
	}

	var userDepts []*entitymodel.UserDept
	if err = p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		WhereIn(dao.UserDept.Columns().DeptId, deptIDs).
		Scan(&userDepts); err != nil {
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
func (p *Provider) GetAllAssignedUserIDs(ctx context.Context) ([]int, error) {
	var userDepts []*entitymodel.UserDept
	if err := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		Fields(dao.UserDept.Columns().UserId).
		Distinct().
		Scan(&userDepts); err != nil {
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
func (p *Provider) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	var userDept *entitymodel.UserDept
	if err := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		Where(dao.UserDept.Columns().UserId, userID).
		Scan(&userDept); err != nil || userDept == nil {
		return 0, "", err
	}

	var deptItem *entitymodel.Dept
	if err := p.tenantFilter.Apply(ctx, dao.Dept.Ctx(ctx)).
		Where(dao.Dept.Columns().Id, userDept.DeptId).
		Scan(&deptItem); err != nil || deptItem == nil {
		return 0, "", err
	}
	return deptItem.Id, deptItem.Name, nil
}

// GetUserDeptIDs returns one user's department identifier list.
func (p *Provider) GetUserDeptIDs(ctx context.Context, userID int) ([]int, error) {
	var userDepts []*entitymodel.UserDept
	if err := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		Where(dao.UserDept.Columns().UserId, userID).
		Scan(&userDepts); err != nil {
		return nil, err
	}

	deptIDs := make([]int, 0, len(userDepts))
	seen := make(map[int]struct{}, len(userDepts))
	for _, item := range userDepts {
		if item == nil {
			continue
		}
		if _, ok := seen[item.DeptId]; ok {
			continue
		}
		seen[item.DeptId] = struct{}{}
		deptIDs = append(deptIDs, item.DeptId)
	}
	return deptIDs, nil
}

// ApplyUserDeptScope injects an EXISTS-based department membership constraint
// into a host-owned query without materializing all visible user IDs in memory.
func (p *Provider) ApplyUserDeptScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	subQuery, empty, err := p.BuildUserDeptScopeExists(ctx, userIDColumn, currentUserID)
	if err != nil || empty {
		return model, empty, err
	}
	return model.Where("EXISTS ?", subQuery), false, nil
}

// BuildUserDeptScopeExists builds an EXISTS subquery for department membership
// without applying it immediately, allowing host callers to compose it with
// additional OR branches.
func (p *Provider) BuildUserDeptScopeExists(
	ctx context.Context,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	deptIDs, err := p.currentVisibleDeptIDs(ctx, currentUserID)
	if err != nil {
		return nil, false, err
	}
	if len(deptIDs) == 0 {
		return nil, true, nil
	}

	cols := dao.UserDept.Columns()
	subQuery := p.tenantFilter.Apply(ctx, dao.UserDept.Ctx(ctx)).
		Fields(cols.UserId).
		Where(fmt.Sprintf("%s = %s", qualifiedUserDeptColumn(cols.UserId), userIDColumn)).
		WhereIn(cols.DeptId, deptIDs)
	return subQuery, false, nil
}

// currentVisibleDeptIDs returns the current user's department IDs plus all
// descendant department IDs with duplicates removed.
func (p *Provider) currentVisibleDeptIDs(ctx context.Context, currentUserID int) ([]int, error) {
	deptIDs, err := p.GetUserDeptIDs(ctx, currentUserID)
	if err != nil {
		return nil, err
	}
	if len(deptIDs) == 0 {
		return []int{}, nil
	}

	seen := make(map[int]struct{})
	visibleDeptIDs := make([]int, 0, len(deptIDs))
	for _, deptID := range deptIDs {
		descendantIDs, resolveErr := p.deptSvc.DescendantDeptIDs(ctx, deptID)
		if resolveErr != nil {
			return nil, resolveErr
		}
		for _, descendantID := range descendantIDs {
			if _, ok := seen[descendantID]; ok {
				continue
			}
			seen[descendantID] = struct{}{}
			visibleDeptIDs = append(visibleDeptIDs, descendantID)
		}
	}
	return visibleDeptIDs, nil
}

// qualifiedUserDeptColumn returns one fully qualified user-department column
// name for correlated subqueries.
func qualifiedUserDeptColumn(column string) string {
	return fmt.Sprintf("%s.%s", dao.UserDept.Table(), column)
}

// GetUserPostIDs returns one user's post association list.
func (p *Provider) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	var userPosts []*entitymodel.UserPost
	if err := p.tenantFilter.Apply(ctx, dao.UserPost.Ctx(ctx)).
		Where(dao.UserPost.Columns().UserId, userID).
		Scan(&userPosts); err != nil {
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
func (p *Provider) ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	return dao.UserDept.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(dao.UserDept.Table()).
			Ctx(ctx).
			Where(plugincontract.TenantFilterColumn, p.tenantFilter.Current(ctx)).
			Where(dao.UserDept.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		if _, err := tx.Model(dao.UserPost.Table()).
			Ctx(ctx).
			Where(plugincontract.TenantFilterColumn, p.tenantFilter.Current(ctx)).
			Where(dao.UserPost.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}

		if deptID != nil && *deptID > 0 {
			if _, err := tx.Model(dao.UserDept.Table()).
				Ctx(ctx).
				Data(do.UserDept{TenantId: p.tenantFilter.Current(ctx), UserId: userID, DeptId: *deptID}).
				Insert(); err != nil {
				return err
			}
		}
		for _, postID := range postIDs {
			if _, err := tx.Model(dao.UserPost.Table()).
				Ctx(ctx).
				Data(do.UserPost{TenantId: p.tenantFilter.Current(ctx), UserId: userID, PostId: postID}).
				Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

// CleanupUserAssignments deletes one user's optional organization associations.
func (p *Provider) CleanupUserAssignments(ctx context.Context, userID int) error {
	return dao.UserDept.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(dao.UserDept.Table()).
			Ctx(ctx).
			Where(plugincontract.TenantFilterColumn, p.tenantFilter.Current(ctx)).
			Where(dao.UserDept.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		if _, err := tx.Model(dao.UserPost.Table()).
			Ctx(ctx).
			Where(plugincontract.TenantFilterColumn, p.tenantFilter.Current(ctx)).
			Where(dao.UserPost.Columns().UserId, userID).
			Delete(); err != nil {
			return err
		}
		return nil
	})
}

// UserDeptTree returns the optional department tree used by host user management.
func (p *Provider) UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error) {
	plainTree, err := p.deptSvc.Tree(ctx)
	if err != nil {
		return nil, err
	}

	counts := make([]deptCountRow, 0)
	if err = p.tenantFilter.ApplyColumn(ctx, dao.UserDept.Ctx(ctx), qualifiedUserDeptColumn(plugincontract.TenantFilterColumn)).
		Fields("dept_id, COUNT(*) AS cnt").
		InnerJoin(
			dao.SysUser.Table(),
			fmt.Sprintf(
				"%s.%s = %s.%s",
				dao.UserDept.Table(), dao.UserDept.Columns().UserId,
				dao.SysUser.Table(), dao.SysUser.Columns().Id,
			),
		).
		Where(fmt.Sprintf("%s.%s", dao.SysUser.Table(), plugincontract.TenantFilterColumn), p.tenantFilter.Current(ctx)).
		Group("dept_id").
		Scan(&counts); err != nil {
		return nil, err
	}

	countMap := make(map[int]int, len(counts))
	for _, item := range counts {
		countMap[item.DeptID] = item.Cnt
	}

	nodes := convertDeptTreeNodes(plainTree)
	applyDeptUserCount(nodes, countMap)

	totalUsers, err := p.tenantFilter.Apply(ctx, dao.SysUser.Ctx(ctx)).Count()
	if err != nil {
		return nil, err
	}

	assignedUsers := 0
	for _, item := range countMap {
		assignedUsers += item
	}

	return append(nodes, newUnassignedDeptNode(totalUsers, assignedUsers)), nil
}

// ListPostOptions returns selectable post options for one department subtree.
func (p *Provider) ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error) {
	model := p.tenantFilter.Apply(ctx, dao.Post.Ctx(ctx)).Where(dao.Post.Columns().Status, postStatusEnabled)
	if deptID != nil {
		deptIDs, err := p.deptSvc.DescendantDeptIDs(ctx, *deptID)
		if err != nil {
			return nil, err
		}
		model = model.WhereIn(dao.Post.Columns().DeptId, deptIDs)
	}

	var posts []*entitymodel.Post
	if err := model.OrderAsc(dao.Post.Columns().Sort).Scan(&posts); err != nil {
		return nil, err
	}

	options := make([]*orgcap.PostOption, 0, len(posts))
	for _, postItem := range posts {
		if postItem == nil {
			continue
		}
		options = append(options, &orgcap.PostOption{
			PostID:   postItem.Id,
			PostName: postItem.Name,
		})
	}
	return options, nil
}

// convertDeptTreeNodes converts plugin-local tree nodes into the shared host contract.
func convertDeptTreeNodes(nodes []*deptsvc.TreeNode) []*orgcap.DeptTreeNode {
	result := make([]*orgcap.DeptTreeNode, 0, len(nodes))
	for _, node := range nodes {
		if node == nil {
			continue
		}
		result = append(result, &orgcap.DeptTreeNode{
			Id:       node.Id,
			Label:    node.Label,
			Children: convertDeptTreeNodes(node.Children),
		})
	}
	return result
}

// applyDeptUserCount rolls grouped department user counts up the tree and appends the count to labels.
func applyDeptUserCount(nodes []*orgcap.DeptTreeNode, countMap map[int]int) {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		applyDeptUserCount(node.Children, countMap)
		node.UserCount = countMap[node.Id]
		for _, child := range node.Children {
			if child == nil {
				continue
			}
			node.UserCount += child.UserCount
		}
		node.Label = fmt.Sprintf("%s(%d)", node.Label, node.UserCount)
	}
}

// newUnassignedDeptNode creates the synthetic Unassigned projection
// with a stable label key so host controllers can localize the label.
func newUnassignedDeptNode(totalUsers int, assignedUsers int) *orgcap.DeptTreeNode {
	unassignedUsers := totalUsers - assignedUsers
	return &orgcap.DeptTreeNode{
		Id:        0,
		Label:     fmt.Sprintf("Unassigned (%d)", unassignedUsers),
		LabelKey:  orgCapUnassignedDeptLabelKey,
		UserCount: unassignedUsers,
		Children:  make([]*orgcap.DeptTreeNode, 0),
	}
}

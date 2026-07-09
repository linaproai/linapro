// orgcap_impl.go implements optional organization-capability delegation. It
// checks source-plugin enablement before forwarding department, post, and
// data-scope operations, returning neutral values when the provider is absent
// so host services can degrade without hard dependencies on organization data.

package orgspi

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
)

// IsProviderEnabled always returns false.
func (noopEnablementReader) IsProviderEnabled(_ context.Context, _ string) bool {
	return false
}

// Available reports whether an active organization provider is available.
func (s *serviceImpl) Available(ctx context.Context) bool {
	if s == nil {
		return false
	}
	return s.manager.registry.StatusWithProvider(ctx, orgcap.CapabilityOrgV1, s.enablement, s.providerEnv).Available
}

// Status returns the current organization capability activation state.
func (s *serviceImpl) Status(ctx context.Context) capmodel.CapabilityStatus {
	if s == nil {
		return convertCapabilityStatus(NewManager().registry.Status(ctx, orgcap.CapabilityOrgV1, nil))
	}
	return convertCapabilityStatus(s.manager.registry.StatusWithProvider(ctx, orgcap.CapabilityOrgV1, s.enablement, s.providerEnv))
}

// Department returns department resource operations.
func (s *serviceImpl) Department() orgcap.DepartmentService {
	return departmentService{root: s}
}

// Post returns post resource operations.
func (s *serviceImpl) Post() orgcap.PostService {
	return postService{root: s}
}

// Assignment returns user organization assignment operations.
func (s *serviceImpl) Assignment() orgcap.AssignmentService {
	return s
}

// Scope returns host-internal organization data-scope operations.
func (s *serviceImpl) Scope() ScopeService {
	return s
}

// Workspace returns user-management workspace projections.
func (s *serviceImpl) Workspace() WorkspaceViewService {
	return s
}

// departmentService delegates department subresource operations to the root
// organization runtime service.
type departmentService struct {
	root *serviceImpl
}

// postService delegates post subresource operations to the root organization
// runtime service.
type postService struct {
	root *serviceImpl
}

// currentProvider returns the currently registered organization-capability provider.
func (s *serviceImpl) currentProvider(ctx context.Context) (Provider, error) {
	if s == nil {
		return nil, nil
	}
	provider, err := s.manager.registry.ActiveProviderWithError(ctx, orgcap.CapabilityOrgV1, s.enablement, s.providerEnv)
	if err != nil || provider == nil {
		return nil, err
	}
	typedProvider, ok := provider.(Provider)
	if !ok {
		return nil, nil
	}
	return typedProvider, nil
}

// Get returns one visible department projection.
func (s departmentService) Get(ctx context.Context, deptID int) (*orgcap.DeptInfo, error) {
	if deptID <= 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	result, err := s.BatchGet(ctx, []int{deptID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[deptID] == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return result.Items[deptID], nil
}

// BatchGet returns visible department projections and opaque missing IDs.
func (s departmentService) BatchGet(
	ctx context.Context,
	deptIDs []int,
) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.DeptInfo, int]{
		Items:      make(map[int]*orgcap.DeptInfo),
		MissingIDs: make([]int, 0),
	}
	if len(deptIDs) == 0 {
		return result, nil
	}
	if len(deptIDs) > orgcap.MaxVisibilityCheckSize {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", orgcap.MaxVisibilityCheckSize))
	}
	normalized := normalizePositiveIDs(deptIDs, orgcap.MaxVisibilityCheckSize)
	if len(normalized) == 0 {
		return result, nil
	}
	provider, err := s.root.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		result.MissingIDs = append(result.MissingIDs, normalized...)
		return result, nil
	}
	return provider.BatchGetDepartments(ctx, normalized)
}

// providerEnv builds lazy construction inputs for one organization provider.
func (s *serviceImpl) providerEnv(ctx context.Context, pluginID string) ProviderEnv {
	env := defaultProviderEnv(ctx, pluginID)
	if s != nil && s.envFactory != nil {
		env = s.envFactory(ctx, pluginID)
	}
	if env.PluginID == "" {
		env.PluginID = pluginID
	}
	return env
}

// ListUserDeptAssignments returns user -> department projections for the provided users.
func (s *serviceImpl) ListUserDeptAssignments(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	assignments := make(map[int]*orgcap.UserDeptAssignment)
	if len(userIDs) == 0 {
		return assignments, nil
	}
	result, err := s.BatchGetUserOrgProfiles(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return assignments, nil
	}
	for userID, profile := range result.Items {
		if profile == nil || profile.DeptID <= 0 {
			continue
		}
		assignments[userID] = &orgcap.UserDeptAssignment{
			DeptID:   profile.DeptID,
			DeptName: profile.DeptName,
		}
	}
	return assignments, nil
}

// GetUserDeptInfo returns one user's department projection.
func (s *serviceImpl) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	profile, err := s.ListByUser(ctx, userID)
	if err != nil {
		return 0, "", err
	}
	if profile == nil {
		return 0, "", nil
	}
	return profile.DeptID, profile.DeptName, nil
}

// GetUserDeptName returns one user's department name for online-session projection.
func (s *serviceImpl) GetUserDeptName(ctx context.Context, userID int) (string, error) {
	_, deptName, err := s.GetUserDeptInfo(ctx, userID)
	return deptName, err
}

// GetUserDeptIDs returns one user's department identifier list.
func (s *serviceImpl) GetUserDeptIDs(ctx context.Context, userID int) ([]int, error) {
	profile, err := s.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil || profile.DeptID <= 0 {
		return []int{}, nil
	}
	return []int{profile.DeptID}, nil
}

// ApplyUserDeptScope injects a database-side department-scope constraint for
// rows owned by the supplied user ID column.
func (s *serviceImpl) ApplyUserDeptScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	subQuery, empty, err := s.BuildUserDeptScopeExists(ctx, userIDColumn, currentUserID)
	if err != nil || empty || subQuery == nil || model == nil {
		return model, empty, err
	}
	return model.Where("EXISTS ?", subQuery), false, nil
}

// BuildUserDeptScopeExists builds the database-side department-scope EXISTS
// subquery for callers that need to compose it with additional OR branches.
func (s *serviceImpl) BuildUserDeptScopeExists(
	ctx context.Context,
	userIDColumn string,
	currentUserID int,
) (*gdb.Model, bool, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil {
		return nil, true, nil
	}
	return provider.BuildUserDeptScopeExists(ctx, userIDColumn, currentUserID)
}

// ApplyUserDeptFilter constrains user rows to a requested department subtree without materializing user IDs.
func (s *serviceImpl) ApplyUserDeptFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	deptID int,
) (*gdb.Model, bool, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil || model == nil {
		return model, false, nil
	}
	return provider.ApplyUserDeptFilter(ctx, model, userIDColumn, deptID)
}

// ApplyUserDeptUnassignedFilter constrains user rows to records without department assignments.
func (s *serviceImpl) ApplyUserDeptUnassignedFilter(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, false, err
	}
	if provider == nil || model == nil {
		return model, false, nil
	}
	return provider.ApplyUserDeptUnassignedFilter(ctx, model, userIDColumn)
}

// GetUserPostIDs returns one user's post association list.
func (s *serviceImpl) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	profile, err := s.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil || len(profile.PostIDs) == 0 {
		return []int{}, nil
	}
	return append([]int(nil), profile.PostIDs...), nil
}

// BatchGetUserOrgProfiles returns stable organization profiles for provided users.
func (s *serviceImpl) BatchGetUserOrgProfiles(
	ctx context.Context,
	userIDs []int,
) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	result := &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{
		Items:      make(map[int]*orgcap.UserOrgProfile),
		MissingIDs: make([]int, 0),
	}
	normalized := normalizePositiveIDs(userIDs, orgcap.MaxUserOrgProfileBatchSize)
	if len(normalized) == 0 {
		return result, nil
	}
	if len(userIDs) > orgcap.MaxUserOrgProfileBatchSize {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", orgcap.MaxUserOrgProfileBatchSize))
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		for _, userID := range normalized {
			result.Items[userID] = &orgcap.UserOrgProfile{UserID: userID}
		}
		return result, nil
	}
	return provider.BatchGetUserOrgProfiles(ctx, normalized)
}

// ListDeptTree returns a bounded ordinary department tree projection.
func (s *serviceImpl) ListDeptTree(ctx context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	maxNodes := input.MaxNodes
	if maxNodes <= 0 || maxNodes > orgcap.MaxDeptTreeNodes {
		maxNodes = orgcap.MaxDeptTreeNodes
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}, nil
	}
	return provider.ListDeptTree(ctx, orgcap.DeptTreeInput{MaxNodes: maxNodes})
}

// SearchDepartments returns bounded department candidates.
func (s *serviceImpl) SearchDepartments(
	ctx context.Context,
	input orgcap.DeptListInput,
) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}, nil
	}
	input.Page = normalizePage(input.Page, orgcap.MaxDeptSearchPageSize)
	return provider.SearchDepartments(ctx, input)
}

// List returns bounded department candidates.
func (s departmentService) List(
	ctx context.Context,
	input orgcap.DeptListInput,
) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return s.root.SearchDepartments(ctx, input)
}

// ListTree returns a bounded department tree projection.
func (s departmentService) ListTree(ctx context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	return s.root.ListDeptTree(ctx, input)
}

// ListOptions returns bounded department option projections.
func (s departmentService) ListOptions(
	ctx context.Context,
	input orgcap.DeptOptionsInput,
) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return s.root.SearchDepartments(ctx, orgcap.DeptListInput{
		Keyword: input.Keyword,
		Status:  input.Status,
		Page:    input.Page,
	})
}

// ListPostOptionsPage returns bounded post candidates.
func (s *serviceImpl) ListPostOptionsPage(
	ctx context.Context,
	input orgcap.PostOptionsInput,
) (*capmodel.PageResult[*orgcap.PostOption], error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}, nil
	}
	input.Page = normalizePage(input.Page, orgcap.MaxPostOptionsPageSize)
	status := input.Status
	if status == nil {
		enabledStatus := 1
		status = &enabledStatus
	}
	result, err := provider.ListPosts(ctx, orgcap.PostListInput{
		DeptID:  input.DeptID,
		Keyword: input.Keyword,
		Status:  status,
		Page:    input.Page,
	})
	if err != nil {
		return nil, err
	}
	options := make([]*orgcap.PostOption, 0)
	if result != nil {
		options = make([]*orgcap.PostOption, 0, len(result.Items))
		for _, item := range result.Items {
			if item == nil {
				continue
			}
			options = append(options, &orgcap.PostOption{PostID: item.PostID, PostName: item.PostName})
		}
		return &capmodel.PageResult[*orgcap.PostOption]{Items: options, Total: result.Total}, nil
	}
	return &capmodel.PageResult[*orgcap.PostOption]{Items: options}, nil
}

// Get returns one visible post projection.
func (s postService) Get(ctx context.Context, postID int) (*orgcap.PostInfo, error) {
	if postID <= 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	result, err := s.BatchGet(ctx, []int{postID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[postID] == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return result.Items[postID], nil
}

// BatchGet returns visible post projections and opaque missing IDs.
func (s postService) BatchGet(
	ctx context.Context,
	postIDs []int,
) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	result := &capmodel.BatchResult[*orgcap.PostInfo, int]{
		Items:      make(map[int]*orgcap.PostInfo),
		MissingIDs: make([]int, 0),
	}
	if len(postIDs) == 0 {
		return result, nil
	}
	if len(postIDs) > orgcap.MaxVisibilityCheckSize {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", orgcap.MaxVisibilityCheckSize))
	}
	provider, err := s.root.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		result.MissingIDs = append(result.MissingIDs, normalizePositiveIDs(postIDs, orgcap.MaxVisibilityCheckSize)...)
		return result, nil
	}
	return provider.BatchGetPosts(ctx, normalizePositiveIDs(postIDs, orgcap.MaxVisibilityCheckSize))
}

// List returns bounded post candidates.
func (s postService) List(
	ctx context.Context,
	input orgcap.PostListInput,
) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	provider, err := s.root.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return &capmodel.PageResult[*orgcap.PostInfo]{Items: []*orgcap.PostInfo{}}, nil
	}
	input.Page = normalizePage(input.Page, orgcap.MaxPostOptionsPageSize)
	return provider.ListPosts(ctx, input)
}

// ListOptions returns bounded post options.
func (s postService) ListOptions(
	ctx context.Context,
	input orgcap.PostOptionsInput,
) (*capmodel.PageResult[*orgcap.PostOption], error) {
	return s.root.ListPostOptionsPage(ctx, input)
}

// EnsureDepartmentsVisible verifies all department identifiers are visible.
func (s *serviceImpl) EnsureDepartmentsVisible(ctx context.Context, deptIDs []int) error {
	normalized := normalizePositiveIDs(deptIDs, orgcap.MaxVisibilityCheckSize)
	if len(deptIDs) > orgcap.MaxVisibilityCheckSize {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", orgcap.MaxVisibilityCheckSize))
	}
	if len(normalized) == 0 {
		return nil
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	result, err := provider.BatchGetDepartments(ctx, normalized)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	for _, deptID := range normalized {
		if result.Items[deptID] == nil {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
	}
	return nil
}

// EnsureVisible verifies department identifiers are visible.
func (s departmentService) EnsureVisible(ctx context.Context, deptIDs []int) error {
	return s.root.EnsureDepartmentsVisible(ctx, deptIDs)
}

// EnsurePostsVisible verifies all post identifiers are visible.
func (s *serviceImpl) EnsurePostsVisible(ctx context.Context, postIDs []int) error {
	normalized := normalizePositiveIDs(postIDs, orgcap.MaxVisibilityCheckSize)
	if len(postIDs) > orgcap.MaxVisibilityCheckSize {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", orgcap.MaxVisibilityCheckSize))
	}
	if len(normalized) == 0 {
		return nil
	}
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	result, err := provider.BatchGetPosts(ctx, normalized)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	for _, postID := range normalized {
		if result.Items[postID] == nil {
			return bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
	}
	return nil
}

// EnsureVisible verifies post identifiers are visible.
func (s postService) EnsureVisible(ctx context.Context, postIDs []int) error {
	return s.root.EnsurePostsVisible(ctx, postIDs)
}

// Create creates one department.
func (s departmentService) Create(ctx context.Context, input orgcap.DeptCreateInput) (int, error) {
	return s.root.CreateDepartment(ctx, input)
}

// Update updates one department.
func (s departmentService) Update(ctx context.Context, input orgcap.DeptUpdateInput) error {
	return s.root.UpdateDepartment(ctx, input)
}

// Delete deletes one department.
func (s departmentService) Delete(ctx context.Context, deptID int) error {
	return s.root.DeleteDepartment(ctx, deptID)
}

// CreateDepartment creates one department through the provider owner.
func (s *serviceImpl) CreateDepartment(ctx context.Context, input orgcap.DeptCreateInput) (int, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return 0, err
	}
	if provider == nil {
		return 0, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	return provider.CreateDepartment(ctx, input)
}

// UpdateDepartment updates one department through the provider owner.
func (s *serviceImpl) UpdateDepartment(ctx context.Context, input orgcap.DeptUpdateInput) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	return provider.UpdateDepartment(ctx, input)
}

// DeleteDepartment deletes one department through the provider owner.
func (s *serviceImpl) DeleteDepartment(ctx context.Context, deptID int) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	return provider.DeleteDepartment(ctx, deptID)
}

// CreatePost creates one post through the provider owner.
func (s *serviceImpl) CreatePost(ctx context.Context, input orgcap.PostCreateInput) (int, error) {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return 0, err
	}
	if provider == nil {
		return 0, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	return provider.CreatePost(ctx, input)
}

// UpdatePost updates one post through the provider owner.
func (s *serviceImpl) UpdatePost(ctx context.Context, input orgcap.PostUpdateInput) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	return provider.UpdatePost(ctx, input)
}

// DeletePost deletes one post through the provider owner.
func (s *serviceImpl) DeletePost(ctx context.Context, postID int) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", orgcap.CapabilityOrgV1))
	}
	return provider.DeletePost(ctx, postID)
}

// Create creates one post.
func (s postService) Create(ctx context.Context, input orgcap.PostCreateInput) (int, error) {
	return s.root.CreatePost(ctx, input)
}

// Update updates one post.
func (s postService) Update(ctx context.Context, input orgcap.PostUpdateInput) error {
	return s.root.UpdatePost(ctx, input)
}

// Delete deletes one post.
func (s postService) Delete(ctx context.Context, postID int) error {
	return s.root.DeletePost(ctx, postID)
}

// ReplaceUserAssignments rewrites one user's department and post associations.
func (s *serviceImpl) ReplaceUserAssignments(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	return provider.ReplaceUserAssignments(ctx, userID, deptID, postIDs)
}

// BatchGetUserProfiles returns stable organization profiles for visible users.
func (s *serviceImpl) BatchGetUserProfiles(
	ctx context.Context,
	userIDs []int,
) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	return s.BatchGetUserOrgProfiles(ctx, userIDs)
}

// ListByUser returns one user's organization profile.
func (s *serviceImpl) ListByUser(ctx context.Context, userID int) (*orgcap.UserOrgProfile, error) {
	result, err := s.BatchGetUserOrgProfiles(ctx, []int{userID})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Items[userID] == nil {
		return &orgcap.UserOrgProfile{UserID: userID}, nil
	}
	return result.Items[userID], nil
}

// BatchListByUsers returns user-to-department projections for the provided users.
func (s *serviceImpl) BatchListByUsers(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	return s.ListUserDeptAssignments(ctx, userIDs)
}

// ReplaceByUser rewrites one user's department and post associations.
func (s *serviceImpl) ReplaceByUser(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	return s.ReplaceUserAssignments(ctx, userID, deptID, postIDs)
}

// CleanupByUser deletes one user's optional organization associations.
func (s *serviceImpl) CleanupByUser(ctx context.Context, userID int) error {
	return s.CleanupUserAssignments(ctx, userID)
}

// CleanupUserAssignments deletes one user's optional organization associations.
func (s *serviceImpl) CleanupUserAssignments(ctx context.Context, userID int) error {
	provider, err := s.currentProvider(ctx)
	if err != nil {
		return err
	}
	if provider == nil {
		return nil
	}
	return provider.CleanupUserAssignments(ctx, userID)
}

// UserDeptTree returns the optional department tree used by host user management.
func (s *serviceImpl) UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error) {
	result, err := s.ListDeptTree(ctx, orgcap.DeptTreeInput{MaxNodes: orgcap.MaxDeptTreeNodes})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []*orgcap.DeptTreeNode{}, nil
	}
	return result.Items, nil
}

// ListPostOptions returns selectable post options for one department subtree.
func (s *serviceImpl) ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error) {
	result, err := s.ListPostOptionsPage(ctx, orgcap.PostOptionsInput{
		DeptID: deptID,
		Page:   capmodel.PageRequest{PageNum: 1, PageSize: orgcap.MaxPostOptionsPageSize},
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return []*orgcap.PostOption{}, nil
	}
	return result.Items, nil
}

// normalizePage applies conservative defaults with a method-specific maximum.
func normalizePage(page capmodel.PageRequest, maxPageSize int) capmodel.PageRequest {
	if page.PageNum <= 0 {
		page.PageNum = 1
	}
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = page.Limit
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if maxPageSize > 0 && pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	page.PageSize = pageSize
	return page
}

// normalizePositiveIDs removes invalid and duplicate positive integer identifiers.
func normalizePositiveIDs(ids []int, limit int) []int {
	if len(ids) == 0 {
		return nil
	}
	result := make([]int, 0, len(ids))
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

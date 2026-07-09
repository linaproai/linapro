// This file keeps user service tests aligned with explicit dependency
// injection when individual fakes are replaced after construction.

package user

import (
	"context"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	hostconfig "lina-core/internal/service/config"
	"lina-core/internal/service/datascope"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/kvcache"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/orgcap"
	"lina-core/pkg/plugin/capability/orgcap/orgspi"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
	"lina-core/pkg/plugin/pluginhost"
)

// newUserTestService constructs user service tests through explicit dependencies.
func newUserTestService(tenantManagersAndRuntimes ...any) Service {
	var (
		bizCtxSvc     = bizctx.New()
		configSvc     = hostconfig.New()
		clusterSvc    = cluster.New(configSvc.GetCluster(context.Background()))
		cacheCoordSvc = cachecoord.Default(clusterSvc)
		i18nSvc       = i18nsvc.New(bizCtxSvc, configSvc, cacheCoordSvc)
		sessionStore  = session.NewDBStore()
		pluginRuntime = userTestPluginRuntime{}
		orgCapSvc     = orgspi.New(nil, pluginRuntime, pluginRuntime.OrgProviderEnv)
		tenantSvc     = tenantspi.New(nil, nil, nil, nil)
	)
	if len(tenantManagersAndRuntimes) > 0 {
		var (
			manager    *tenantspi.Manager
			enablement interface {
				IsProviderEnabled(context.Context, string) bool
			}
		)
		if value, ok := tenantManagersAndRuntimes[0].(*tenantspi.Manager); ok {
			manager = value
			if len(tenantManagersAndRuntimes) > 1 {
				enablement, _ = tenantManagersAndRuntimes[1].(interface {
					IsProviderEnabled(context.Context, string) bool
				})
			}
		} else {
			enablement, _ = tenantManagersAndRuntimes[0].(interface {
				IsProviderEnabled(context.Context, string) bool
			})
		}
		tenantSvc = tenantspi.New(manager, enablement, nil, nil)
	}
	roleSvc := role.New(pluginRuntime, bizCtxSvc, configSvc, i18nSvc, orgCapSvc, tenantSvc)
	scopeSvc := datascope.New(bizCtxSvc, roleSvc, orgCapSvc.Scope())
	roleSvc.SetDataScopeService(scopeSvc)
	var (
		kvCacheSvc = kvcache.New()
		authSvc    = auth.New(configSvc, pluginRuntime, orgCapSvc, roleSvc, tenantSvc, sessionStore, kvCacheSvc)
		userSvc    = New(authSvc, bizCtxSvc, i18nSvc, orgCapSvc, roleSvc, scopeSvc, tenantSvc)
	)
	return userSvc.(*serviceImpl)
}

// userTestPluginRuntime supplies the narrow plugin-facing seams needed by user
// service tests without importing the plugin service package back into user.
type userTestPluginRuntime struct{}

// DispatchHookEvent ignores plugin hooks in user tests.
func (userTestPluginRuntime) DispatchHookEvent(_ context.Context, _ pluginhost.ExtensionPoint, _ map[string]interface{}) error {
	return nil
}

// FilterPermissionMenus leaves permission menus unchanged in user tests.
func (userTestPluginRuntime) FilterPermissionMenus(_ context.Context, menus []*entity.SysMenu) []*entity.SysMenu {
	return menus
}

// IsProviderEnabled reports that no plugin provider is active in user tests.
func (userTestPluginRuntime) IsProviderEnabled(_ context.Context, _ string) bool {
	return false
}

// OrgProviderEnv returns a neutral organization provider environment.
func (userTestPluginRuntime) OrgProviderEnv(_ context.Context, pluginID string) orgspi.ProviderEnv {
	return orgspi.ProviderEnv{PluginID: pluginID}
}

// setUserTestBizCtx replaces the business context dependency and refreshes
// the derived data-scope service used by user-management tests.
func setUserTestBizCtx(svc *serviceImpl, bizCtxSvc bizctx.Service) {
	svc.bizCtxSvc = bizCtxSvc
	refreshUserTestScope(svc)
}

// setUserTestOrgCap replaces the organization capability dependency and
// refreshes the derived data-scope service used by user-management tests.
func setUserTestOrgCap(svc *serviceImpl, orgCapSvc any) {
	if service, ok := orgCapSvc.(orgspi.Service); ok {
		svc.orgCapSvc = service
	} else {
		var orgScope orgspi.ScopeService
		if scope, ok := orgCapSvc.(orgspi.ScopeService); ok {
			orgScope = scope
		}
		svc.orgCapSvc = userTestOrgService{
			userTestOrgCapDirectory: userTestOrgCapDirectory{legacy: orgCapSvc},
			scope:                   orgScope,
		}
	}
	refreshUserTestScope(svc)
}

// userTestOrgService adapts legacy flat organization fakes to the host-owned
// orgspi.Service contract used by the user service.
type userTestOrgService struct {
	userTestOrgCapDirectory
	scope orgspi.ScopeService
}

// Scope returns the optional organization data-scope fake.
func (s userTestOrgService) Scope() orgspi.ScopeService {
	return s.scope
}

// Workspace returns optional user-management workspace projections.
func (s userTestOrgService) Workspace() orgspi.WorkspaceViewService {
	return userTestOrgWorkspace{legacy: s.legacy}
}

// userTestOrgCapDirectory adapts older flat organization fakes to the current
// orgcap.Service subresource directory used by user service internals.
type userTestOrgCapDirectory struct {
	legacy any
}

// userTestOrgWorkspace adapts optional legacy workspace projection fakes.
type userTestOrgWorkspace struct {
	legacy any
}

// UserDeptTree returns a fake department tree for user-management tests.
func (w userTestOrgWorkspace) UserDeptTree(ctx context.Context) ([]*orgcap.DeptTreeNode, error) {
	if provider, ok := w.legacy.(interface {
		UserDeptTree(context.Context) ([]*orgcap.DeptTreeNode, error)
	}); ok {
		return provider.UserDeptTree(ctx)
	}
	return []*orgcap.DeptTreeNode{}, nil
}

// ListPostOptions returns fake post options for user-management tests.
func (w userTestOrgWorkspace) ListPostOptions(ctx context.Context, deptID *int) ([]*orgcap.PostOption, error) {
	if provider, ok := w.legacy.(interface {
		ListPostOptions(context.Context, *int) ([]*orgcap.PostOption, error)
	}); ok {
		return provider.ListPostOptions(ctx, deptID)
	}
	return []*orgcap.PostOption{}, nil
}

// Available reports whether the legacy fake organization capability is active.
func (d userTestOrgCapDirectory) Available(ctx context.Context) bool {
	if provider, ok := d.legacy.(interface {
		Available(context.Context) bool
	}); ok {
		return provider.Available(ctx)
	}
	return false
}

// Status returns the legacy fake organization status when available.
func (d userTestOrgCapDirectory) Status(ctx context.Context) capmodel.CapabilityStatus {
	if provider, ok := d.legacy.(interface {
		Status(context.Context) capmodel.CapabilityStatus
	}); ok {
		return provider.Status(ctx)
	}
	return capmodel.CapabilityStatus{}
}

// Department returns the department subresource adapter.
func (d userTestOrgCapDirectory) Department() orgcap.DepartmentService {
	return userTestOrgDepartment{legacy: d.legacy}
}

// Post returns the post subresource adapter.
func (d userTestOrgCapDirectory) Post() orgcap.PostService {
	return userTestOrgPost{legacy: d.legacy}
}

// Assignment returns the user assignment subresource adapter.
func (d userTestOrgCapDirectory) Assignment() orgcap.AssignmentService {
	return userTestOrgAssignment{legacy: d.legacy}
}

// userTestOrgDepartment adapts flat department fake methods.
type userTestOrgDepartment struct {
	legacy any
}

// Get returns no department projection in user tests.
func (d userTestOrgDepartment) Get(context.Context, int) (*orgcap.DeptInfo, error) {
	return nil, nil
}

// BatchGet returns all requested departments as missing in user tests.
func (d userTestOrgDepartment) BatchGet(_ context.Context, deptIDs []int) (*capmodel.BatchResult[*orgcap.DeptInfo, int], error) {
	return &capmodel.BatchResult[*orgcap.DeptInfo, int]{
		Items:      map[int]*orgcap.DeptInfo{},
		MissingIDs: append([]int(nil), deptIDs...),
	}, nil
}

// List returns bounded fake departments.
func (d userTestOrgDepartment) List(ctx context.Context, input orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	if provider, ok := d.legacy.(interface {
		SearchDepartments(context.Context, orgcap.DeptListInput) (*capmodel.PageResult[*orgcap.DeptInfo], error)
	}); ok {
		return provider.SearchDepartments(ctx, input)
	}
	return &capmodel.PageResult[*orgcap.DeptInfo]{Items: []*orgcap.DeptInfo{}}, nil
}

// ListTree returns a fake department tree.
func (d userTestOrgDepartment) ListTree(ctx context.Context, input orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error) {
	if provider, ok := d.legacy.(interface {
		ListDeptTree(context.Context, orgcap.DeptTreeInput) (*orgcap.DeptTreeResult, error)
	}); ok {
		return provider.ListDeptTree(ctx, input)
	}
	return &orgcap.DeptTreeResult{Items: []*orgcap.DeptTreeNode{}}, nil
}

// ListOptions reuses the list adapter for user tests.
func (d userTestOrgDepartment) ListOptions(ctx context.Context, input orgcap.DeptOptionsInput) (*capmodel.PageResult[*orgcap.DeptInfo], error) {
	return d.List(ctx, orgcap.DeptListInput{
		Keyword: input.Keyword,
		Page:    input.Page,
	})
}

// EnsureVisible verifies department identifiers through the legacy fake.
func (d userTestOrgDepartment) EnsureVisible(ctx context.Context, deptIDs []int) error {
	if provider, ok := d.legacy.(interface {
		EnsureDepartmentsVisible(context.Context, []int) error
	}); ok {
		return provider.EnsureDepartmentsVisible(ctx, deptIDs)
	}
	return nil
}

// Create accepts department creation in user tests.
func (d userTestOrgDepartment) Create(context.Context, orgcap.DeptCreateInput) (int, error) {
	return 0, nil
}

// Update accepts department updates in user tests.
func (d userTestOrgDepartment) Update(context.Context, orgcap.DeptUpdateInput) error {
	return nil
}

// Delete accepts department deletion in user tests.
func (d userTestOrgDepartment) Delete(context.Context, int) error {
	return nil
}

// userTestOrgPost adapts flat post fake methods.
type userTestOrgPost struct {
	legacy any
}

// Get returns no post projection in user tests.
func (p userTestOrgPost) Get(context.Context, int) (*orgcap.PostInfo, error) {
	return nil, nil
}

// BatchGet returns all requested posts as missing in user tests.
func (p userTestOrgPost) BatchGet(_ context.Context, postIDs []int) (*capmodel.BatchResult[*orgcap.PostInfo, int], error) {
	return &capmodel.BatchResult[*orgcap.PostInfo, int]{
		Items:      map[int]*orgcap.PostInfo{},
		MissingIDs: append([]int(nil), postIDs...),
	}, nil
}

// List returns empty post projections in user tests.
func (p userTestOrgPost) List(context.Context, orgcap.PostListInput) (*capmodel.PageResult[*orgcap.PostInfo], error) {
	return &capmodel.PageResult[*orgcap.PostInfo]{Items: []*orgcap.PostInfo{}}, nil
}

// ListOptions returns fake post options.
func (p userTestOrgPost) ListOptions(ctx context.Context, input orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error) {
	if provider, ok := p.legacy.(interface {
		ListPostOptionsPage(context.Context, orgcap.PostOptionsInput) (*capmodel.PageResult[*orgcap.PostOption], error)
	}); ok {
		return provider.ListPostOptionsPage(ctx, input)
	}
	return &capmodel.PageResult[*orgcap.PostOption]{Items: []*orgcap.PostOption{}}, nil
}

// EnsureVisible verifies post identifiers through the legacy fake.
func (p userTestOrgPost) EnsureVisible(ctx context.Context, postIDs []int) error {
	if provider, ok := p.legacy.(interface {
		EnsurePostsVisible(context.Context, []int) error
	}); ok {
		return provider.EnsurePostsVisible(ctx, postIDs)
	}
	return nil
}

// Create accepts post creation in user tests.
func (p userTestOrgPost) Create(context.Context, orgcap.PostCreateInput) (int, error) {
	return 0, nil
}

// Update accepts post updates in user tests.
func (p userTestOrgPost) Update(context.Context, orgcap.PostUpdateInput) error {
	return nil
}

// Delete accepts post deletion in user tests.
func (p userTestOrgPost) Delete(context.Context, int) error {
	return nil
}

// userTestOrgAssignment adapts flat assignment fake methods.
type userTestOrgAssignment struct {
	legacy any
}

// BatchGetUserProfiles returns fake user organization profiles.
func (a userTestOrgAssignment) BatchGetUserProfiles(ctx context.Context, userIDs []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error) {
	if provider, ok := a.legacy.(interface {
		BatchGetUserOrgProfiles(context.Context, []int) (*capmodel.BatchResult[*orgcap.UserOrgProfile, int], error)
	}); ok {
		return provider.BatchGetUserOrgProfiles(ctx, userIDs)
	}
	return &capmodel.BatchResult[*orgcap.UserOrgProfile, int]{
		Items:      map[int]*orgcap.UserOrgProfile{},
		MissingIDs: append([]int(nil), userIDs...),
	}, nil
}

// ListByUser returns one fake user organization profile.
func (a userTestOrgAssignment) ListByUser(ctx context.Context, userID int) (*orgcap.UserOrgProfile, error) {
	result, err := a.BatchGetUserProfiles(ctx, []int{userID})
	if err != nil || result == nil {
		return nil, err
	}
	return result.Items[userID], nil
}

// BatchListByUsers returns fake department assignments.
func (a userTestOrgAssignment) BatchListByUsers(ctx context.Context, userIDs []int) (map[int]*orgcap.UserDeptAssignment, error) {
	if provider, ok := a.legacy.(interface {
		ListUserDeptAssignments(context.Context, []int) (map[int]*orgcap.UserDeptAssignment, error)
	}); ok {
		return provider.ListUserDeptAssignments(ctx, userIDs)
	}
	return map[int]*orgcap.UserDeptAssignment{}, nil
}

// GetUserDeptInfo returns one fake user department.
func (a userTestOrgAssignment) GetUserDeptInfo(ctx context.Context, userID int) (int, string, error) {
	if provider, ok := a.legacy.(interface {
		GetUserDeptInfo(context.Context, int) (int, string, error)
	}); ok {
		return provider.GetUserDeptInfo(ctx, userID)
	}
	return 0, "", nil
}

// GetUserDeptIDs returns fake user department IDs.
func (a userTestOrgAssignment) GetUserDeptIDs(ctx context.Context, userID int) ([]int, error) {
	if provider, ok := a.legacy.(interface {
		GetUserDeptIDs(context.Context, int) ([]int, error)
	}); ok {
		return provider.GetUserDeptIDs(ctx, userID)
	}
	return []int{}, nil
}

// GetUserPostIDs returns fake user post IDs.
func (a userTestOrgAssignment) GetUserPostIDs(ctx context.Context, userID int) ([]int, error) {
	if provider, ok := a.legacy.(interface {
		GetUserPostIDs(context.Context, int) ([]int, error)
	}); ok {
		return provider.GetUserPostIDs(ctx, userID)
	}
	return []int{}, nil
}

// ReplaceByUser delegates fake assignment replacement.
func (a userTestOrgAssignment) ReplaceByUser(ctx context.Context, userID int, deptID *int, postIDs []int) error {
	if provider, ok := a.legacy.(interface {
		ReplaceUserAssignments(context.Context, int, *int, []int) error
	}); ok {
		return provider.ReplaceUserAssignments(ctx, userID, deptID, postIDs)
	}
	return nil
}

// CleanupByUser delegates fake assignment cleanup.
func (a userTestOrgAssignment) CleanupByUser(ctx context.Context, userID int) error {
	if provider, ok := a.legacy.(interface {
		CleanupUserAssignments(context.Context, int) error
	}); ok {
		return provider.CleanupUserAssignments(ctx, userID)
	}
	return nil
}

// refreshUserTestScope rebuilds the stateless data-scope helper from the
// current explicit fake dependencies.
func refreshUserTestScope(svc *serviceImpl) {
	var orgScope orgspi.ScopeService
	if svc.orgCapSvc != nil {
		orgScope = svc.orgCapSvc.Scope()
	}
	svc.scopeSvc = datascope.New(svc.bizCtxSvc, svc.roleSvc, orgScope)
}

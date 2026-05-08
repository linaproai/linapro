## 1. Database Design and Migration

- [x] 1.1 Create SQL migration file `008-menu-role-management.sql`
- [x] 1.2 Create `sys_menu` table
- [x] 1.3 Create `sys_role` table
- [x] 1.4 Create `sys_role_menu` table
- [x] 1.5 Create `sys_user_role` table
- [x] 1.6 Add menu-related dictionary types (menu status, display status, menu type)
- [x] 1.7 Add initial role data (admin, user roles)
- [x] 1.8 Add initial menu data (system management menus and sub-menus)
- [x] 1.9 Assign admin role to default administrator user
- [x] 1.10 Add `sys.auth.loginPanelLayout` and `sys.auth.pageDesc` to `sys_config` seed data with default values

## 2. Backend - Menu Management Module

- [x] 2.1 Run `make dao` to generate menu-related DAO/DO/Entity
- [x] 2.2 Create menu API definition files `api/menu/v1/`
- [x] 2.3 Define menu list query endpoint `GET /menu`
- [x] 2.4 Define menu detail endpoint `GET /menu/:id`
- [x] 2.5 Define menu creation endpoint `POST /menu`
- [x] 2.6 Define menu update endpoint `PUT /menu/:id`
- [x] 2.7 Define menu deletion endpoint `DELETE /menu/:id`
- [x] 2.8 Define menu dropdown tree endpoint `GET /menu/treeselect`
- [x] 2.9 Define role menu tree endpoint `GET /menu/role/:roleId`
- [x] 2.10 Run `make ctrl` to generate menu controller skeleton
- [x] 2.11 Implement menu service layer `internal/service/menu/menu.go`
- [x] 2.12 Implement menu list query (tree structure)
- [x] 2.13 Implement menu detail query
- [x] 2.14 Implement menu creation (name uniqueness validation)
- [x] 2.15 Implement menu update
- [x] 2.16 Implement menu deletion (cascade child menus, clean role associations)
- [x] 2.17 Implement menu dropdown tree endpoint
- [x] 2.18 Implement role menu tree endpoint

## 3. Backend - Role Management Module

- [x] 3.1 Run `make dao` to generate role-related DAO/DO/Entity
- [x] 3.2 Create role API definition files `api/role/v1/`
- [x] 3.3 Define role list query endpoint `GET /role`
- [x] 3.4 Define role detail endpoint `GET /role/:id`
- [x] 3.5 Define role creation endpoint `POST /role`
- [x] 3.6 Define role update endpoint `PUT /role/:id`
- [x] 3.7 Define role deletion endpoint `DELETE /role/:id`
- [x] 3.8 Define role status toggle endpoint `PUT /role/:id/status`
- [x] 3.9 Define role dropdown options endpoint `GET /role/options`
- [x] 3.10 Define role user list endpoint `GET /role/:id/users`
- [x] 3.11 Define role assign users endpoint `POST /role/:id/users`
- [x] 3.12 Define remove user authorization endpoint `DELETE /role/:id/users/:userId`
- [x] 3.13 Run `make ctrl` to generate role controller skeleton
- [x] 3.14 Implement role service layer `internal/service/role/role.go`
- [x] 3.15 Implement role list query (paginated)
- [x] 3.16 Implement role detail query (including menu ID list)
- [x] 3.17 Implement role creation (name and key uniqueness validation)
- [x] 3.18 Implement role update (including menu association update)
- [x] 3.19 Implement role deletion (clean menu and user associations)
- [x] 3.20 Implement role status toggle
- [x] 3.21 Implement role dropdown options endpoint
- [x] 3.22 Implement role user list query
- [x] 3.23 Implement role assign users
- [x] 3.24 Implement remove user authorization

## 4. Backend - User Management Extension

- [x] 4.1 Extend user list API to return roleIds and roleNames
- [x] 4.2 Extend user detail API to return roleIds
- [x] 4.3 Extend user creation to support roleIds parameter
- [x] 4.4 Extend user update to support roleIds parameter
- [x] 4.5 Extend user deletion to clean sys_user_role associations

## 5. Backend - Login Authentication Extension

- [x] 5.1 Extend `/user/info` endpoint response structure
- [x] 5.2 Implement user role query logic
- [x] 5.3 Implement user menu tree construction logic
- [x] 5.4 Implement user permission identifier aggregation logic
- [x] 5.5 Handle super admin special logic (return all menus)
- [x] 5.6 Handle user with no roles (empty menu logic)

## 6. Backend - Login-Page Config Parameters

- [x] 6.1 Add protected public-frontend metadata, default values, domain validation, and whitelist projection for `sys.auth.loginPanelLayout`
- [x] 6.2 Add protected public-frontend metadata, default values, and length validation for `sys.auth.pageDesc`
- [x] 6.3 Update built-in system-parameter seed list in host initialization SQL
- [x] 6.4 Add backend config unit tests covering valid values, invalid value rejection, and public-frontend output

## 7. Frontend - Menu Management Page

- [x] 7.1 Create menu management API file `src/api/system/menu/`
- [x] 7.2 Create menu management route `src/router/routes/modules/system.ts`
- [x] 7.3 Create menu management page `src/views/system/menu/index.vue`
- [x] 7.4 Create menu form drawer `src/views/system/menu/menu-drawer.vue`
- [x] 7.5 Create menu data definitions `src/views/system/menu/data.ts`
- [x] 7.6 Implement menu tree table display
- [x] 7.7 Implement menu search functionality
- [x] 7.8 Implement menu creation functionality
- [x] 7.9 Implement menu editing functionality
- [x] 7.10 Implement menu deletion functionality (with cascade)
- [x] 7.11 Implement menu status toggle
- [x] 7.12 Implement menu icon selector

## 8. Frontend - Role Management Page

- [x] 8.1 Create role management API file `src/api/system/role/`
- [x] 8.2 Create role management route
- [x] 8.3 Create role management page `src/views/system/role/index.vue`
- [x] 8.4 Create role form drawer `src/views/system/role/role-drawer.vue`
- [x] 8.5 Create role data definitions `src/views/system/role/data.ts`
- [x] 8.6 Create role menu selection component `src/components/tree/MenuSelectTable.vue`
- [x] 8.7 Create role user assignment page `src/views/system/role/authUser.vue`
- [x] 8.8 Implement role list display
- [x] 8.9 Implement role search functionality
- [x] 8.10 Implement role creation (including menu selection)
- [x] 8.11 Implement role editing (including menu selection)
- [x] 8.12 Implement role deletion
- [x] 8.13 Implement role status toggle
- [x] 8.14 Implement role user assignment
- [x] 8.15 Implement remove user authorization

## 9. Frontend - User Management Extension

- [x] 9.1 Extend user list to add role column
- [x] 9.2 Extend user form to add role selector
- [x] 9.3 Extend user creation to support role association
- [x] 9.4 Extend user editing to support role association

## 10. Frontend - Login Authentication Extension

- [x] 10.1 Update user info type definitions, add menus and permissions
- [x] 10.2 Update useUserStore to store menu and permission info
- [x] 10.3 Configure frontend dynamic route generation logic
- [x] 10.4 Implement menu-tree-based route registration
- [x] 10.5 Implement button-level permission directive v-access

## 11. Frontend - Login Page Presentation

- [x] 11.1 Adjust frontend default preferences and public-frontend runtime sync so login page defaults to `panel-right`
- [x] 11.2 Adjust login-page assembly so forgot password, registration, mobile login, QR-code login, and third-party login entry points are explicitly hidden
- [x] 11.3 Adjust auth-route registration so unfinished auth subpages redirect to `/auth/login`
- [x] 11.4 Update login-page description to match product positioning
- [x] 11.5 Update frontend runtime tests for public-frontend config sync

## 12. Backend - Server Monitor Optimization

- [x] 12.1 Create SQL migration `009-server-monitor-updated-at.sql` adding `updated_at` field
- [x] 12.2 Run `make dao` to regenerate DAO/DO/Entity
- [x] 12.3 Modify `CollectAndStore` method to use framework auto-handled time fields
- [x] 12.4 Modify `GetLatest` method to sort by `updated_at`
- [x] 12.5 Update frontend "collection time" label to "data update time"
- [x] 12.6 Add `CleanupStale` method to `servermon` module
- [x] 12.7 Extract cron job registration logic to `cron/cron_servermon_cleanup.go`

## 13. E2E Tests

- [x] 13.1 Create menu management test `TC0060-menu-crud.ts`
- [x] 13.2 Test menu creation
- [x] 13.3 Test menu editing
- [x] 13.4 Test menu deletion
- [x] 13.5 Test menu tree display
- [x] 13.6 Create role management test `TC0061-role-crud.ts`
- [x] 13.7 Test role creation
- [x] 13.8 Test role editing (including menu selection)
- [x] 13.9 Test role deletion
- [x] 13.10 Test role user assignment
- [x] 13.11 Create user-role association test `TC0062-user-role.ts`
- [x] 13.12 Test user creation with role selection
- [x] 13.13 Test user editing with role modification
- [x] 13.14 Test user list displays roles
- [x] 13.15 Create login menu test `TC0063-auth-menu.ts`
- [x] 13.16 Test menu displays correctly after login
- [x] 13.17 Test different role users see different menus
- [x] 13.18 Create role form defaults test `TC0064-role-form-defaults.ts`
- [x] 13.19 Create login page presentation test `TC0102-login-page-presentation.ts`
- [x] 13.20 Test unfinished entry points stay hidden
- [x] 13.21 Test login page defaults to right-aligned layout
- [x] 13.22 Test system parameter switches layout

## 14. Integration and Verification

- [x] 14.1 Run backend unit tests
- [x] 14.2 Run frontend type checks
- [x] 14.3 Run full E2E test suite
- [x] 14.4 Manually verify super admin login menu
- [x] 14.5 Manually verify regular user login menu
- [x] 14.6 Manually verify button-level permission control
- [x] 14.7 Manually verify login page presentation and system-parameter behavior

## Feedback

- [x] **FB-1**: Menu page data display failure due to query form dictionary option loading method error causing VXE-Grid initialization failure
- [x] **FB-2**: Menu status search dropdown showing incorrect options
- [x] **FB-3**: Empty column on right side of table
- [x] **FB-4**: Status and display column label styles inconsistent with reference project
- [x] **FB-5**: Form row spacing and remark input style inconsistent with reference project
- [x] **FB-6**: Menu management page query form dictionary dropdowns empty; switched from `getDictOptionsSync()` to `getDictOptions()` for async loading
- [x] **FB-9**: Menu status search dropdown showing extra options due to test data pollution from TC0013; fixed with independent test dictionary type and SQL cleanup
- [x] **FB-10**: Edit menu drawer not showing edited menu content; root cause: `handleEdit` passes `isEdit` but drawer checks `update`
- [x] **FB-11**: Parent menu dropdown tree not showing child menus in new/edit drawer; root cause: backend returns tree but frontend treated as flat list
- [x] **FB-12**: Edit menu parent selector should disable current menu and descendants to prevent circular reference
- [x] **FB-13**: Remark input in new/edit menu drawer too small; replaced with Textarea component
- [x] **FB-14**: Clicking "add" on specific menu should show that menu as parent in the add panel; root cause: `resetForm()` called after `setFieldValue()`
- [x] **FB-16**: Role sort InputNumber default value correctly configured
- [x] **FB-17**: Data permission field has `rules: 'required'` for mandatory selection
- [x] **FB-18**: Role menu permission tree shows type icons and permission checkboxes; added Type and Icon fields to MenuTreeNode
- [x] **FB-19**: Role sort InputNumber default not displaying; changed `getDrawerSchema()` from async to sync
- [x] **FB-20**: Data permission field defaults to "all data permissions"; confirmed `defaultValue: 1` correct after FB-19 fix
- [x] **FB-21**: Role status Select dictionary loading timing issue; changed to RadioGroup with hardcoded options
- [x] **FB-22**: Server monitor data storage optimization; added `updated_at` field and `CleanupStale` method
- [x] **FB-24**: Improve cron job service layer encapsulation; added `CleanupStale` to servermon, extracted cron registration
- [x] **FB-login-1**: Expand `sys.auth.pageDesc` length limit to 500 characters with validation and regression coverage
- [x] **FB-login-2**: Update default login-page description and standardize default login-panel layout to right side

## Menu Structure Summary
```
Dashboard (sort:0)
  Analysis
  Workspace
System Management (sort:1)
  User Management + 7 buttons
  Role Management + 4 buttons
  Menu Management + 4 buttons
  Department Management + 4 buttons
  Post Management + 5 buttons
  Dictionary Management + 5 buttons
  Notice Announcement + 4 buttons
  Parameter Settings + 5 buttons
  File Management + 4 buttons
  Message List (hidden)
  Role Auth Users (hidden)
System Monitor (sort:2)
  Online Users + 2 buttons
  Server Monitor
  Operation Log + 4 buttons
  Login Log + 4 buttons
System Info (sort:3)
  System API
  Version Info
Personal Center (hidden, sort:99)
```

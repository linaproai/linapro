// Package descriptor implements host-owned access-control descriptor views used
// behind the public pluginhost callback contracts.
package descriptor

// MenuInput carries one host menu projection for plugin access-control callbacks.
type MenuInput struct {
	// ID is the menu identifier.
	ID int
	// ParentID is the parent menu identifier.
	ParentID int
	// Name is the menu display name.
	Name string
	// Path is the menu route path.
	Path string
	// Component is the menu component binding.
	Component string
	// Permission is the permission key bound to the menu.
	Permission string
	// MenuKey is the stable menu business key.
	MenuKey string
	// MenuType is the menu type code.
	MenuType string
	// Visible is the menu visibility status.
	Visible int
	// Status is the menu enabled status.
	Status int
}

// Menu is the immutable descriptor implementation returned through pluginhost.
type Menu struct {
	id         int
	parentID   int
	name       string
	path       string
	component  string
	permission string
	menuKey    string
	menuType   string
	visible    int
	status     int
}

// Permission is the immutable permission descriptor implementation returned
// through pluginhost.
type Permission struct {
	menuKey    string
	menuName   string
	permission string
}

// NewMenu creates one host-owned menu descriptor view.
func NewMenu(input MenuInput) *Menu {
	return &Menu{
		id:         input.ID,
		parentID:   input.ParentID,
		name:       input.Name,
		path:       input.Path,
		component:  input.Component,
		permission: input.Permission,
		menuKey:    input.MenuKey,
		menuType:   input.MenuType,
		visible:    input.Visible,
		status:     input.Status,
	}
}

// NewPermission creates one host-owned permission descriptor view.
func NewPermission(menuKey string, menuName string, permission string) *Permission {
	return &Permission{
		menuKey:    menuKey,
		menuName:   menuName,
		permission: permission,
	}
}

// ID returns the menu identifier.
func (d *Menu) ID() int {
	if d == nil {
		return 0
	}
	return d.id
}

// ParentID returns the parent menu identifier.
func (d *Menu) ParentID() int {
	if d == nil {
		return 0
	}
	return d.parentID
}

// Name returns the menu display name.
func (d *Menu) Name() string {
	if d == nil {
		return ""
	}
	return d.name
}

// Path returns the menu route path.
func (d *Menu) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

// Component returns the menu component binding.
func (d *Menu) Component() string {
	if d == nil {
		return ""
	}
	return d.component
}

// Permissions returns the menu permission string.
func (d *Menu) Permissions() string {
	if d == nil {
		return ""
	}
	return d.permission
}

// MenuKey returns the stable menu business key.
func (d *Menu) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

// Type returns the menu type code.
func (d *Menu) Type() string {
	if d == nil {
		return ""
	}
	return d.menuType
}

// Visible returns the menu visibility status.
func (d *Menu) Visible() int {
	if d == nil {
		return 0
	}
	return d.visible
}

// Status returns the menu enabled status.
func (d *Menu) Status() int {
	if d == nil {
		return 0
	}
	return d.status
}

// MenuKey returns the stable business key of the menu owning the permission.
func (d *Permission) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

// MenuName returns the display name of the menu owning the permission.
func (d *Permission) MenuName() string {
	if d == nil {
		return ""
	}
	return d.menuName
}

// Permission returns the concrete permission string.
func (d *Permission) Permission() string {
	if d == nil {
		return ""
	}
	return d.permission
}

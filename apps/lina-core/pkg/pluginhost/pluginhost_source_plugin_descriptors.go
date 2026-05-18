// This file defines descriptor wrappers published to source-plugin governance
// callbacks for menu and permission filtering.

package pluginhost

// MenuDescriptor exposes one published menu descriptor for plugin menu filtering.
type MenuDescriptor interface {
	// ID returns the menu id.
	ID() int
	// ParentID returns the parent menu id.
	ParentID() int
	// Name returns the menu display name.
	Name() string
	// Path returns the menu path.
	Path() string
	// Component returns the routed component name.
	Component() string
	// Permissions returns the permission key bound to the menu.
	Permissions() string
	// MenuKey returns the stable menu business key.
	MenuKey() string
	// Type returns the menu type.
	Type() string
	// Visible returns the visible status.
	Visible() int
	// Status returns the enabled status.
	Status() int
}

// menuDescriptor is the host-owned implementation of MenuDescriptor.
type menuDescriptor struct {
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

// PermissionDescriptor exposes one published permission descriptor for plugin permission filtering.
type PermissionDescriptor interface {
	// MenuKey returns the stable menu business key.
	MenuKey() string
	// MenuName returns the display name of the menu that owns the permission.
	MenuName() string
	// Permission returns the permission string.
	Permission() string
}

// permissionDescriptor is the host-owned implementation of PermissionDescriptor.
type permissionDescriptor struct {
	menuKey    string
	menuName   string
	permission string
}

// NewMenuDescriptor creates one published menu descriptor wrapper for plugins.
func NewMenuDescriptor(
	id int,
	parentID int,
	name string,
	path string,
	component string,
	permission string,
	menuKey string,
	menuType string,
	visible int,
	status int,
) MenuDescriptor {
	return &menuDescriptor{
		id:         id,
		parentID:   parentID,
		name:       name,
		path:       path,
		component:  component,
		permission: permission,
		menuKey:    menuKey,
		menuType:   menuType,
		visible:    visible,
		status:     status,
	}
}

// NewPermissionDescriptor creates one published permission descriptor wrapper for plugins.
func NewPermissionDescriptor(menuKey string, menuName string, permission string) PermissionDescriptor {
	return &permissionDescriptor{
		menuKey:    menuKey,
		menuName:   menuName,
		permission: permission,
	}
}

// ID returns the menu identifier.
func (d *menuDescriptor) ID() int {
	if d == nil {
		return 0
	}
	return d.id
}

// ParentID returns the parent menu identifier.
func (d *menuDescriptor) ParentID() int {
	if d == nil {
		return 0
	}
	return d.parentID
}

// Name returns the menu display name.
func (d *menuDescriptor) Name() string {
	if d == nil {
		return ""
	}
	return d.name
}

// Path returns the menu route path.
func (d *menuDescriptor) Path() string {
	if d == nil {
		return ""
	}
	return d.path
}

// Component returns the menu component binding.
func (d *menuDescriptor) Component() string {
	if d == nil {
		return ""
	}
	return d.component
}

// Permissions returns the menu permission string.
func (d *menuDescriptor) Permissions() string {
	if d == nil {
		return ""
	}
	return d.permission
}

// MenuKey returns the stable menu business key.
func (d *menuDescriptor) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

// Type returns the menu type code.
func (d *menuDescriptor) Type() string {
	if d == nil {
		return ""
	}
	return d.menuType
}

// Visible returns the menu visibility status.
func (d *menuDescriptor) Visible() int {
	if d == nil {
		return 0
	}
	return d.visible
}

// Status returns the menu enabled status.
func (d *menuDescriptor) Status() int {
	if d == nil {
		return 0
	}
	return d.status
}

// MenuKey returns the stable business key of the menu owning the permission.
func (d *permissionDescriptor) MenuKey() string {
	if d == nil {
		return ""
	}
	return d.menuKey
}

// MenuName returns the display name of the menu owning the permission.
func (d *permissionDescriptor) MenuName() string {
	if d == nil {
		return ""
	}
	return d.menuName
}

// Permission returns the concrete permission string.
func (d *permissionDescriptor) Permission() string {
	if d == nil {
		return ""
	}
	return d.permission
}

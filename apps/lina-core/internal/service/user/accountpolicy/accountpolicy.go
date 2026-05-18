// Package accountpolicy defines built-in user-account policy helpers owned by
// the user service component.
package accountpolicy

const (
	// DefaultAdminUsername is the built-in administrator username.
	DefaultAdminUsername = "admin"
)

// IsBuiltInAdminUsername reports whether the supplied username belongs to the
// built-in administrator account.
func IsBuiltInAdminUsername(username string) bool {
	return username == DefaultAdminUsername
}

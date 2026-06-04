//go:build !linux && !darwin

package process

// List conservatively returns no process metadata on platforms where linactl
// has not implemented shell-free process enumeration.
func List() ([]Info, error) {
	return nil, nil
}

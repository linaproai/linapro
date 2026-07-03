// This file defines typed runtime i18n values shared by locale metadata.

package i18n

// LocaleDirection represents the document text direction for one locale.
type LocaleDirection string

const (
	// LocaleDirectionLTR means left-to-right text layout.
	LocaleDirectionLTR LocaleDirection = "ltr"
)

// String returns the persisted string value for the locale direction.
func (direction LocaleDirection) String() string {
	return string(direction)
}

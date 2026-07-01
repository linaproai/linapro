// This file keeps tenant ID parsing coupled to the tenant capability contract
// instead of exposing a separate parsing component.

package tenantcap

import (
	"strconv"
	"strings"

	"lina-core/pkg/plugin/capability/capmodel"
)

// ParseTenantID decodes one domain ID into the host integer tenant key.
func ParseTenantID(value capmodel.DomainID) (int, error) {
	tenantID, err := strconv.Atoi(strings.TrimSpace(string(value)))
	if err != nil {
		return 0, err
	}
	return tenantID, nil
}

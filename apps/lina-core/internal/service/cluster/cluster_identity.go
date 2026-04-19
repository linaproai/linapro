// This file resolves a stable node identifier for single-node and clustered
// deployments.

package cluster

import (
	"os"

	"github.com/gogf/gf/v2/net/gipv4"
)

// generateNodeIdentifier prefers hostname, then intranet IP, and finally a
// local fallback to identify the current node.
func generateNodeIdentifier() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	if hostname == "" {
		hostname, err = gipv4.GetIntranetIp()
		if err != nil {
			hostname = ""
		}
	}
	if hostname == "" {
		hostname = "local-node"
	}
	return hostname
}

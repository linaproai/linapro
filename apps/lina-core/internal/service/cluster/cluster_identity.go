// This file resolves a stable node identifier for single-node and clustered
// deployments.

package cluster

import (
	"os"
	"strings"

	"github.com/gogf/gf/v2/net/gipv4"
)

// nodeIDEnvName names the optional process-level node identifier override used
// by clustered deployments and multi-process integration tests.
const nodeIDEnvName = "LINAPRO_NODE_ID"

// generateNodeIdentifier prefers hostname, then intranet IP, and finally a
// local fallback to identify the current node.
func generateNodeIdentifier() string {
	if nodeID := strings.TrimSpace(os.Getenv(nodeIDEnvName)); nodeID != "" {
		return nodeID
	}

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

package cluster

import (
	"os"

	"github.com/gogf/gf/v2/net/gipv4"
)

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

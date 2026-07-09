// protocol_hostservice_types.go exposes host service manifest declaration
// types through the public protocol facade. Payload DTOs and codecs are owned
// by the protocol package's codec files.

package protocol

import "lina-core/pkg/plugin/pluginbridge/internal/hostservice"

type (
	HostServiceResourceSpec = hostservice.HostServiceResourceSpec
	HostServiceSpec         = hostservice.HostServiceSpec
)

// Package cmd assembles the Lina core command tree used for host bootstrap and
// maintenance operations.
package cmd

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Main is the root command object for the Lina core CLI.
type Main struct {
	g.Meta `name:"main" root:"http"`
}

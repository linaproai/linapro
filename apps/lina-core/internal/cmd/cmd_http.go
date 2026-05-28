// This file contains the HTTP command entrypoint and delegates detailed
// startup responsibilities to focused HTTP helper files.

package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/cmd/internal/httpstartup"
)

// HttpInput defines CLI input for the HTTP startup command.
type HttpInput struct {
	g.Meta `name:"http" brief:"start http server"`
}

// HttpOutput is the CLI output placeholder for the HTTP startup command.
type HttpOutput struct{}

// Http bootstraps the host HTTP server, static API routes, plugin routes, and
// embedded frontend asset serving.
func (m *Main) Http(ctx context.Context, in HttpInput) (out *HttpOutput, err error) {
	if err = httpstartup.Run(ctx); err != nil {
		return nil, err
	}
	return
}

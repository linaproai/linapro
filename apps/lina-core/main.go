// This file boots the Lina core command tree and applies startup-wide runtime
// defaults.

package main

import (
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/cmd"
	"lina-core/internal/service/config"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
)

// main loads bootstrap configuration, configures logging, and runs the command tree.
func main() {
	var (
		ctx          = gctx.GetInitCtx()
		configSvc    = config.New()
		loggerConfig = configSvc.GetLogger(ctx)
	)
	logger.Configure(logger.RuntimeConfig{
		Structured:     loggerConfig.Extensions.Structured,
		TraceIDEnabled: loggerConfig.Extensions.TraceIDEnabled,
	})

	c, err := gcmd.NewFromObject(cmd.Main{})
	if err != nil {
		panic(err)
	}
	c.Run(ctx)
}

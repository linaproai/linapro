// This file boots the Lina core command tree and applies startup-wide runtime
// defaults.

package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/cmd"
	"lina-core/internal/service/config"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"

	_ "lina-plugins"
)

func main() {
	ctx := gctx.GetInitCtx()
	logger.Configure(config.New().GetLogger(ctx).Structured)

	c, err := gcmd.NewFromObject(cmd.Main{})
	if err != nil {
		panic(err)
	}
	c.Run(ctx)
}

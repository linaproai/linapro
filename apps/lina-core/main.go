package main

import (
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/cmd"

	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"

	_ "lina-plugins"
)

func main() {
	c, err := gcmd.NewFromObject(cmd.Main{})
	if err != nil {
		panic(err)
	}
	c.Run(gctx.GetInitCtx())
}

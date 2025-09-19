package main

import (
	_ "teamspeak-one-click-deploy/internal/packed"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"github.com/gogf/gf/v2/os/gctx"

	"teamspeak-one-click-deploy/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

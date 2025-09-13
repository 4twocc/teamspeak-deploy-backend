package main

import (
	_ "teamspeak-one-click-deploy/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"teamspeak-one-click-deploy/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}

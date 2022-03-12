package main

import (
	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/util"

	_ "github.com/tliron/kutil/logging/simple"
)

func main() {
	util.ExitOnSIGTERM()
	logging.Configure(1, nil)
	server := plugin.NewRunnableServer(new(Runnable))
	server.Start()
	util.Exit(0)
}
